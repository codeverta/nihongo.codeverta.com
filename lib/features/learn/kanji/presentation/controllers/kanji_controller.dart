import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/kanji/data/kanji_repository.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';

final kanjiRepositoryProvider = Provider<KanjiRepository>(
  (ref) => KanjiRepository(ref.watch(dioProvider)),
);
final kanjiCourseProvider = FutureProvider.autoDispose<KanjiCourse>((ref) {
  ref
    ..watch(authControllerProvider.select((s) => s.status))
    ..watch(authControllerProvider.select((s) => s.user?.id));
  return ref.watch(kanjiRepositoryProvider).getCourse();
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
  }) =>
      KanjiState(
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
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  final c = KanjiController(
    ref.watch(kanjiRepositoryProvider),
    id,
    onProgress: () {
      ref
        ..invalidate(kanjiCourseProvider)
        ..invalidate(homeDashboardProvider);
    },
  );
  unawaited(c.load());
  return c;
});

class KanjiController extends StateNotifier<AsyncValue<KanjiState>> {
  KanjiController(this.repository, this.lessonId, {required this.onProgress})
      : super(const AsyncValue.loading());
  final KanjiRepository repository;
  final String lessonId;
  final VoidCallback onProgress;
  Future<void> load() async {
    state = const AsyncValue.loading();
    final loaded = await AsyncValue.guard(() async {
      final lesson = await repository.getLesson(lessonId);
      if (lesson.characters.isEmpty) {
        throw const KanjiException('Lesson ini belum memiliki materi.');
      }
      return KanjiState(lesson: lesson);
    });
    if (mounted) state = loaded;
  }

  void moveCharacter(int delta) {
    final s = state.valueOrNull;
    if (s == null || s.busy || s.stage != KanjiStage.study) return;
    final index = s.characterIndex + delta;
    if (index >= 0 && index < s.lesson.characters.length) {
      state = AsyncValue.data(s.copyWith(characterIndex: index));
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
      if (mounted) {
        onProgress();
        state = AsyncValue.data(
          s.copyWith(
            stage: KanjiStage.result,
            result: result,
            busy: false,
          ),
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
    }
  }
}
