import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/grammar/data/grammar_repository.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';

final grammarRepositoryProvider = Provider<GrammarRepository>(
  (ref) => GrammarRepository(ref.watch(dioProvider)),
);
final grammarCourseProvider = FutureProvider.autoDispose<GrammarCourse>((ref) {
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  return ref.watch(grammarRepositoryProvider).getCourse();
});

final grammarCourseLevelProvider =
    FutureProvider.autoDispose.family<GrammarCourse, String>((ref, level) {
  if (level == 'N5') return ref.watch(grammarCourseProvider.future);
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  return ref.watch(grammarRepositoryProvider).getCourse(level: level);
});

enum GrammarStage { study, quiz, result }

@immutable
class GrammarState {
  const GrammarState({
    required this.lesson,
    this.stage = GrammarStage.study,
    this.pointIndex = 0,
    this.questionIndex = 0,
    this.attempt,
    this.selection,
    this.feedback,
    this.result,
    this.busy = false,
    this.error,
  });
  final GrammarLesson lesson;
  final GrammarStage stage;
  final int pointIndex;
  final int questionIndex;
  final GrammarAttempt? attempt;
  final String? selection;
  final String? error;
  final GrammarFeedback? feedback;
  final GrammarResult? result;
  final bool busy;
  GrammarState copyWith({
    GrammarStage? stage,
    int? pointIndex,
    int? questionIndex,
    GrammarAttempt? attempt,
    String? selection,
    GrammarFeedback? feedback,
    GrammarResult? result,
    bool? busy,
    String? error,
    bool clearAnswer = false,
  }) =>
      GrammarState(
        lesson: lesson,
        stage: stage ?? this.stage,
        pointIndex: pointIndex ?? this.pointIndex,
        questionIndex: questionIndex ?? this.questionIndex,
        attempt: attempt ?? this.attempt,
        selection: clearAnswer ? null : selection ?? this.selection,
        feedback: clearAnswer ? null : feedback ?? this.feedback,
        result: result ?? this.result,
        busy: busy ?? this.busy,
        error: error,
      );
}

final grammarLessonProvider = StateNotifierProvider.autoDispose
    .family<GrammarController, AsyncValue<GrammarState>, String>((ref, id) {
  // Recreate the controller when the signed-in account changes.
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  final c = GrammarController(
    ref.watch(grammarRepositoryProvider),
    id,
    onProgress: () {
      ref
        ..invalidate(grammarCourseProvider)
        ..invalidate(grammarCourseLevelProvider)
        ..invalidate(homeDashboardProvider);
    },
  );
  unawaited(c.load());
  return c;
});

class GrammarController extends StateNotifier<AsyncValue<GrammarState>> {
  GrammarController(this.repository, this.lessonId, {required this.onProgress})
      : super(const AsyncValue.loading());
  final GrammarRepository repository;
  final String lessonId;
  final VoidCallback onProgress;
  Future<void> load() async {
    state = const AsyncValue.loading();
    final loaded = await AsyncValue.guard(() async {
      final lesson = await repository.getLesson(lessonId);
      if (lesson.points.isEmpty) {
        throw const GrammarException('Lesson ini belum memiliki materi.');
      }
      return GrammarState(lesson: lesson);
    });
    if (mounted) state = loaded;
  }

  void movePoint(int delta) {
    final s = state.valueOrNull;
    if (s == null || s.busy || s.stage != GrammarStage.study) return;
    final index = s.pointIndex + delta;
    if (index >= 0 && index < s.lesson.points.length) {
      state = AsyncValue.data(s.copyWith(pointIndex: index));
    }
  }

  void select(String id) {
    final s = state.valueOrNull;
    if (s == null ||
        s.busy ||
        s.feedback != null ||
        s.stage != GrammarStage.quiz) {
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
        throw const GrammarException('Kuis lesson belum tersedia.');
      }
      if (mounted) {
        state = AsyncValue.data(
          GrammarState(
            lesson: s.lesson,
            stage: GrammarStage.quiz,
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
            stage: GrammarStage.result,
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
      state = AsyncValue.data(GrammarState(lesson: s.lesson));
    }
  }
}
