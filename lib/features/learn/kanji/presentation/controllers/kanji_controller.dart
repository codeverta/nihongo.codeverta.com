import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:nihongo_master/features/learn/kanji/data/kanji_repository.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

final kanjiRepositoryProvider = Provider<KanjiRepository>(
  (ref) => KanjiRepository(
    ref.watch(dioProvider),
    offlinePractice: ref.watch(offlinePracticeServiceProvider),
  ),
);
final kanjiCourseProvider = FutureProvider.autoDispose
    .family<KanjiCourse, String>((ref, level) {
      ref
        ..watch(authControllerProvider.select((s) => s.status))
        ..watch(authControllerProvider.select((s) => s.user?.id));
      return ref.watch(kanjiRepositoryProvider).getCourse(level: level);
    });

enum KanjiStage { study, quiz, result }

@immutable
class KanjiState {
  const KanjiState({
    required this.lesson,
    this.stage = KanjiStage.study,
    this.characterIndex = 0,
    this.questionIndex = 0,
    this.attempt,
    this.selection,
    this.feedback,
    this.result,
    this.busy = false,
    this.error,
  });
  final KanjiLesson lesson;
  final KanjiStage stage;
  final int characterIndex;
  final int questionIndex;
  final KanjiAttempt? attempt;
  final String? selection;
  final String? error;
  final KanjiFeedback? feedback;
  final KanjiResult? result;
  final bool busy;
  KanjiState copyWith({
    KanjiStage? stage,
    int? characterIndex,
    int? questionIndex,
    KanjiAttempt? attempt,
    String? selection,
    KanjiFeedback? feedback,
    KanjiResult? result,
    bool? busy,
    String? error,
    bool clearAnswer = false,
  }) => KanjiState(
    lesson: lesson,
    stage: stage ?? this.stage,
    characterIndex: characterIndex ?? this.characterIndex,
    questionIndex: questionIndex ?? this.questionIndex,
    attempt: attempt ?? this.attempt,
    selection: clearAnswer ? null : selection ?? this.selection,
    feedback: clearAnswer ? null : feedback ?? this.feedback,
    result: result ?? this.result,
    busy: busy ?? this.busy,
    error: error,
  );
}

final kanjiLessonProvider = StateNotifierProvider.autoDispose
    .family<KanjiController, AsyncValue<KanjiState>, String>((ref, id) {
      // Recreate the controller when the signed-in account changes.
      final userId = ref.watch(
        authControllerProvider.select((s) => s.user?.id),
      );
      final c = KanjiController(
        ref.watch(kanjiRepositoryProvider),
        id,
        history: userId == null
            ? null
            : ref.watch(learningHistoryProvider.notifier),
        onProgress: () {
          ref
            ..invalidate(kanjiCourseProvider)
            ..invalidate(homeDashboardProvider)
            ..invalidate(gamificationProvider);
        },
      );
      unawaited(c.load());
      return c;
    });

class KanjiController extends StateNotifier<AsyncValue<KanjiState>> {
  KanjiController(
    this.repository,
    this.lessonId, {
    required this.onProgress,
    this.history,
  }) : super(const AsyncValue.loading()) {
    _syncSubscription = repository.syncEvents.listen(_onPracticeSync);
  }
  final KanjiRepository repository;
  final String lessonId;
  final VoidCallback onProgress;
  final LearningHistoryController? history;
  StreamSubscription<OfflinePracticeSyncEvent>? _syncSubscription;

  void _onPracticeSync(OfflinePracticeSyncEvent event) {
    final current = state.valueOrNull;
    if (!mounted ||
        current?.attempt?.id != event.attemptId ||
        current?.result == null) {
      return;
    }
    final result = KanjiResult.fromJson(event.result);
    if (result.passed) history?.complete('kanji', lessonId);
    onProgress();
    state = AsyncValue.data(current!.copyWith(result: result));
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    final loaded = await AsyncValue.guard(() async {
      final lesson = await repository.getLesson(lessonId);
      if (lesson.characters.isEmpty) {
        throw const KanjiException('Lesson ini belum memiliki materi.');
      }
      final savedPosition = mounted
          ? await history?.open(
              kind: 'kanji',
              isActive: () => mounted,
              lessonId: lesson.summary.id,
              title: lesson.summary.title,
              total: lesson.characters.length,
              completed: lesson.summary.completed,
            )
          : null;
      return KanjiState(lesson: lesson, characterIndex: savedPosition ?? 0);
    });
    if (mounted) state = loaded;
  }

  void moveCharacter(int delta) {
    final s = state.valueOrNull;
    if (s == null || s.busy || s.stage != KanjiStage.study) return;
    final index = s.characterIndex + delta;
    if (index >= 0 && index < s.lesson.characters.length) {
      state = AsyncValue.data(s.copyWith(characterIndex: index));
      history?.position('kanji', lessonId, index);
    }
  }

  void select(String id) {
    final s = state.valueOrNull;
    if (s == null ||
        s.busy ||
        s.feedback != null ||
        s.stage != KanjiStage.quiz) {
      return;
    }
    state = AsyncValue.data(s.copyWith(selection: id));
  }

  Future<void> startQuiz() async {
    final s = state.valueOrNull;
    if (s == null || s.busy) return;
    state = AsyncValue.data(s.copyWith(busy: true));
    try {
      final attempt = await repository.startAttempt(lessonId);
      if (attempt.questions.isEmpty) {
        throw const KanjiException('Kuis lesson belum tersedia.');
      }
      if (mounted) {
        state = AsyncValue.data(
          KanjiState(
            lesson: s.lesson,
            stage: KanjiStage.quiz,
            attempt: attempt,
          ),
        );
      }
    } on Object catch (e) {
      if (mounted) {
        state = AsyncValue.data(s.copyWith(error: e.toString(), busy: false));
      }
    }
  }

  Future<void> submit() async {
    final s = state.valueOrNull;
    if (s == null ||
        s.busy ||
        s.selection == null ||
        s.attempt == null ||
        s.feedback != null) {
      return;
    }
    state = AsyncValue.data(s.copyWith(busy: true));
    try {
      final feedback = await repository.answer(
        s.attempt!.id,
        s.attempt!.questions[s.questionIndex].id,
        s.selection!,
      );
      if (mounted) {
        state = AsyncValue.data(s.copyWith(feedback: feedback, busy: false));
      }
    } on Object catch (e) {
      if (mounted) {
        state = AsyncValue.data(s.copyWith(error: e.toString(), busy: false));
      }
    }
  }

  Future<void> nextQuestion() async {
    final s = state.valueOrNull;
    if (s == null || s.busy || s.feedback == null || s.attempt == null) return;
    if (s.questionIndex < s.attempt!.questions.length - 1) {
      state = AsyncValue.data(
        s.copyWith(questionIndex: s.questionIndex + 1, clearAnswer: true),
      );
      return;
    }
    state = AsyncValue.data(s.copyWith(busy: true));
    try {
      final result = await repository.complete(s.attempt!.id);
      if (mounted && result.verified && result.passed) {
        history?.complete('kanji', lessonId);
      }
      if (mounted) {
        if (result.verified) onProgress();
        state = AsyncValue.data(
          s.copyWith(stage: KanjiStage.result, result: result, busy: false),
        );
      }
    } on Object catch (e) {
      if (mounted) {
        state = AsyncValue.data(s.copyWith(error: e.toString(), busy: false));
      }
    }
  }

  void studyAgain() {
    final s = state.valueOrNull;
    if (s != null && !s.busy) {
      state = AsyncValue.data(KanjiState(lesson: s.lesson));
      history?.position('kanji', lessonId, 0);
    }
  }

  @override
  void dispose() {
    unawaited(_syncSubscription?.cancel());
    super.dispose();
  }
}
