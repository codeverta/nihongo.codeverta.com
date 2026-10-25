import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';
import 'package:nihongo_master/features/learn/reading/data/reading_repository.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

final readingRepositoryProvider = Provider<ReadingRepository>(
  (ref) => ReadingRepository(
    ref.watch(dioProvider),
    offlinePractice: ref.watch(offlinePracticeServiceProvider),
  ),
);
final readingCourseProvider = FutureProvider.autoDispose
    .family<ReadingCourse, String>((ref, level) {
      ref
        ..watch(authControllerProvider.select((s) => s.status))
        ..watch(authControllerProvider.select((s) => s.user?.id));
      return ref.watch(readingRepositoryProvider).getCourse(level: level);
    });

enum ReadingStage { study, quiz, result }

@immutable
class ReadingState {
  const ReadingState({
    required this.lesson,
    this.stage = ReadingStage.study,
    this.pageIndex = 0,
    this.questionIndex = 0,
    this.attempt,
    this.selection,
    this.feedback,
    this.result,
    this.busy = false,
    this.error,
  });
  final ReadingLesson lesson;
  final ReadingStage stage;
  final int pageIndex;
  final int questionIndex;
  final ReadingAttempt? attempt;
  final String? selection;
  final String? error;
  final ReadingFeedback? feedback;
  final ReadingResult? result;
  final bool busy;
  ReadingState copyWith({
    ReadingStage? stage,
    int? pageIndex,
    int? questionIndex,
    ReadingAttempt? attempt,
    String? selection,
    ReadingFeedback? feedback,
    ReadingResult? result,
    bool? busy,
    String? error,
    bool clearAnswer = false,
  }) => ReadingState(
    lesson: lesson,
    stage: stage ?? this.stage,
    pageIndex: pageIndex ?? this.pageIndex,
    questionIndex: questionIndex ?? this.questionIndex,
    attempt: attempt ?? this.attempt,
    selection: clearAnswer ? null : selection ?? this.selection,
    feedback: clearAnswer ? null : feedback ?? this.feedback,
    result: result ?? this.result,
    busy: busy ?? this.busy,
    error: error,
  );
}

final readingLessonProvider = StateNotifierProvider.autoDispose
    .family<ReadingController, AsyncValue<ReadingState>, String>((ref, id) {
      // Recreate the controller when the signed-in account changes.
      final userId = ref.watch(
        authControllerProvider.select((s) => s.user?.id),
      );
      final c = ReadingController(
        ref.watch(readingRepositoryProvider),
        id,
        history: userId == null
            ? null
            : ref.watch(learningHistoryProvider.notifier),
        onProgress: () {
          ref
            ..invalidate(readingCourseProvider)
            ..invalidate(homeDashboardProvider)
            ..invalidate(gamificationProvider);
        },
      );
      unawaited(c.load());
      return c;
    });

class ReadingController extends StateNotifier<AsyncValue<ReadingState>> {
  ReadingController(
    this.repository,
    this.lessonId, {
    required this.onProgress,
    this.history,
  }) : super(const AsyncValue.loading()) {
    _syncSubscription = repository.syncEvents.listen(_onPracticeSync);
  }
  final ReadingRepository repository;
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
    final result = ReadingResult.fromJson(event.result);
    if (result.passed) history?.complete('reading', lessonId);
    onProgress();
    state = AsyncValue.data(current!.copyWith(result: result));
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    final loaded = await AsyncValue.guard(() async {
      final lesson = await repository.getLesson(lessonId);
      if (lesson.pages.isEmpty) {
        throw const ReadingException('Lesson ini belum memiliki materi.');
      }
      final savedPosition = mounted
          ? await history?.open(
              kind: 'reading',
              isActive: () => mounted,
              lessonId: lesson.summary.id,
              title: lesson.summary.title,
              total: lesson.pages.length,
              completed: lesson.summary.completed,
            )
          : null;
      return ReadingState(lesson: lesson, pageIndex: savedPosition ?? 0);
    });
    if (mounted) state = loaded;
  }

  void movePage(int delta) {
    final s = state.valueOrNull;
    if (s == null || s.busy || s.stage != ReadingStage.study) return;
    final index = s.pageIndex + delta;
    if (index >= 0 && index < s.lesson.pages.length) {
      state = AsyncValue.data(s.copyWith(pageIndex: index));
      history?.position('reading', lessonId, index);
    }
  }

  void select(String id) {
    final s = state.valueOrNull;
    if (s == null ||
        s.busy ||
        s.feedback != null ||
        s.stage != ReadingStage.quiz) {
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
        throw const ReadingException('Kuis lesson belum tersedia.');
      }
      if (mounted) {
        state = AsyncValue.data(
          ReadingState(
            lesson: s.lesson,
            stage: ReadingStage.quiz,
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
        history?.complete('reading', lessonId);
      }
      if (mounted) {
        if (result.verified) onProgress();
        state = AsyncValue.data(
          s.copyWith(stage: ReadingStage.result, result: result, busy: false),
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
      state = AsyncValue.data(ReadingState(lesson: s.lesson));
      history?.position('reading', lessonId, 0);
    }
  }

  @override
  void dispose() {
    unawaited(_syncSubscription?.cancel());
    super.dispose();
  }
}
