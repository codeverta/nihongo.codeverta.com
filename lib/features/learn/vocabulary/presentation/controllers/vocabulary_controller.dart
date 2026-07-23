import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/data/vocabulary_repository.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';

final vocabularyRepositoryProvider = Provider<VocabularyRepository>((ref) {
  return VocabularyRepository(ref.watch(dioProvider));
});

final vocabularyLevelsProvider =
    FutureProvider.autoDispose<List<VocabularyLevel>>((ref) {
  ref.watch(authControllerProvider.select((state) => state.status));
  return ref.watch(vocabularyRepositoryProvider).getLevels();
});

enum VocabularyStage { study, quiz, result }

@immutable
class VocabularyLessonState {
  const VocabularyLessonState({
    required this.lesson,
    this.stage = VocabularyStage.study,
    this.wordIndex = 0,
    this.questionIndex = 0,
    this.attempt,
    this.selectedOptionId,
    this.feedback,
    this.result,
    this.submitting = false,
  });

  final VocabularyLesson lesson;
  final VocabularyStage stage;
  final int wordIndex;
  final int questionIndex;
  final VocabularyAttempt? attempt;
  final String? selectedOptionId;
  final AnswerFeedback? feedback;
  final VocabularyResult? result;
  final bool submitting;

  VocabularyLessonState copyWith({
    VocabularyStage? stage,
    int? wordIndex,
    int? questionIndex,
    VocabularyAttempt? attempt,
    String? selectedOptionId,
    bool clearSelection = false,
    AnswerFeedback? feedback,
    bool clearFeedback = false,
    VocabularyResult? result,
    bool clearResult = false,
    bool? submitting,
  }) {
    return VocabularyLessonState(
      lesson: lesson,
      stage: stage ?? this.stage,
      wordIndex: wordIndex ?? this.wordIndex,
      questionIndex: questionIndex ?? this.questionIndex,
      attempt: attempt ?? this.attempt,
      selectedOptionId:
          clearSelection ? null : selectedOptionId ?? this.selectedOptionId,
      feedback: clearFeedback ? null : feedback ?? this.feedback,
      result: clearResult ? null : result ?? this.result,
      submitting: submitting ?? this.submitting,
    );
  }
}

final vocabularyLessonControllerProvider = StateNotifierProvider.autoDispose
    .family<VocabularyLessonController, AsyncValue<VocabularyLessonState>,
        String>((ref, lessonId) {
  final controller = VocabularyLessonController(
    ref.watch(vocabularyRepositoryProvider),
    lessonId,
    onProgressChanged: () {
      ref
        ..invalidate(vocabularyLevelsProvider)
        ..invalidate(homeDashboardProvider);
    },
  );
  unawaited(controller.load());
  return controller;
});

class VocabularyLessonController
    extends StateNotifier<AsyncValue<VocabularyLessonState>> {
  VocabularyLessonController(
    this._repository,
    this._lessonId, {
    required this.onProgressChanged,
  }) : super(const AsyncValue.loading());

  final VocabularyRepository _repository;
  final String _lessonId;
  final VoidCallback onProgressChanged;

  Future<void> load() async {
    state = await AsyncValue.guard(() async {
      final lesson = await _repository.getLesson(_lessonId);
      if (lesson.words.isEmpty) {
        throw const VocabularyException('Lesson ini belum memiliki kosakata.');
      }
      return VocabularyLessonState(lesson: lesson);
    });
  }

  void previousWord() {
    final current = state.valueOrNull;
    if (current == null || current.wordIndex == 0) return;
    state = AsyncValue.data(current.copyWith(wordIndex: current.wordIndex - 1));
  }

  void nextWord() {
    final current = state.valueOrNull;
    if (current == null ||
        current.wordIndex >= current.lesson.words.length - 1) {
      return;
    }
    state = AsyncValue.data(current.copyWith(wordIndex: current.wordIndex + 1));
  }

  Future<void> startQuiz() async {
    final current = state.valueOrNull;
    if (current == null || current.submitting) return;
    state = AsyncValue.data(current.copyWith(submitting: true));
    try {
      final attempt = await _repository.startAttempt(_lessonId);
      state = AsyncValue.data(
        current.copyWith(
          stage: VocabularyStage.quiz,
          attempt: attempt,
          questionIndex: 0,
          clearSelection: true,
          clearFeedback: true,
          clearResult: true,
          submitting: false,
        ),
      );
    } on Object catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  void selectOption(String optionId) {
    final current = state.valueOrNull;
    if (current == null || current.feedback != null || current.submitting) {
      return;
    }
    state = AsyncValue.data(current.copyWith(selectedOptionId: optionId));
  }

  Future<void> submitAnswer() async {
    final current = state.valueOrNull;
    final attempt = current?.attempt;
    final option = current?.selectedOptionId;
    if (current == null ||
        attempt == null ||
        option == null ||
        current.submitting) {
      return;
    }
    final question = attempt.questions[current.questionIndex];
    state = AsyncValue.data(current.copyWith(submitting: true));
    try {
      final feedback = await _repository.answer(
        attemptId: attempt.id,
        questionId: question.id,
        optionId: option,
      );
      state = AsyncValue.data(
        current.copyWith(feedback: feedback, submitting: false),
      );
    } on Object catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> continueQuiz() async {
    final current = state.valueOrNull;
    final attempt = current?.attempt;
    if (current == null || attempt == null || current.feedback == null) return;
    if (current.questionIndex < attempt.questions.length - 1) {
      state = AsyncValue.data(
        current.copyWith(
          questionIndex: current.questionIndex + 1,
          clearSelection: true,
          clearFeedback: true,
        ),
      );
      return;
    }
    state = AsyncValue.data(current.copyWith(submitting: true));
    try {
      final result = await _repository.complete(attempt.id);
      onProgressChanged();
      state = AsyncValue.data(
        current.copyWith(
          stage: VocabularyStage.result,
          result: result,
          submitting: false,
        ),
      );
    } on Object catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> retry() => startQuiz();
}
