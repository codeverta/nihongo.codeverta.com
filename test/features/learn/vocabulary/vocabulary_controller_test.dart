import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/learn/vocabulary/data/vocabulary_repository.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';

void main() {
  test('lesson flow studies, validates, completes, saves and retries',
      () async {
    final repository = _FakeVocabularyRepository();
    var progressRefreshes = 0;
    final controller = VocabularyLessonController(
      repository,
      'n5-daily',
      onProgressChanged: () => progressRefreshes++,
    );

    await controller.load();
    expect(controller.state.valueOrNull!.stage, VocabularyStage.study);
    expect(controller.state.valueOrNull!.lesson.words.single.term, '私');

    await controller.startQuiz();
    expect(controller.state.valueOrNull!.stage, VocabularyStage.quiz);

    controller.selectOption('wrong-option');
    await controller.submitAnswer();
    expect(controller.state.valueOrNull!.feedback?.correct, isFalse);
    expect(
      controller.state.valueOrNull!.feedback?.correctOptionId,
      'correct-option',
    );

    await controller.continueQuiz();
    expect(controller.state.valueOrNull!.stage, VocabularyStage.result);
    expect(controller.state.valueOrNull!.result?.score, 0);
    expect(progressRefreshes, 1);

    await controller.retry();
    expect(controller.state.valueOrNull!.stage, VocabularyStage.quiz);
    expect(controller.state.valueOrNull!.feedback, isNull);
    expect(repository.startedAttempts, 2);
  });
}

class _FakeVocabularyRepository extends VocabularyRepository {
  _FakeVocabularyRepository() : super(Dio());

  int startedAttempts = 0;

  @override
  Future<VocabularyLesson> getLesson(String lessonId) async => _lesson;

  @override
  Future<VocabularyAttempt> startAttempt(String lessonId) async {
    startedAttempts++;
    return _attempt;
  }

  @override
  Future<AnswerFeedback> answer({
    required String attemptId,
    required String questionId,
    required String optionId,
  }) async {
    return const AnswerFeedback(
      questionId: 'q1',
      correct: false,
      correctOptionId: 'correct-option',
      explanation: '私 berarti saya.',
    );
  }

  @override
  Future<VocabularyResult> complete(String attemptId) async => _result;
}

const _lesson = VocabularyLesson(
  id: 'n5-daily',
  level: 'N5',
  title: 'Kata dasar',
  description: 'Lesson pemula',
  bestScore: 0,
  completed: false,
  words: [
    VocabularyWord(
      id: 'word-1',
      term: '私',
      reading: 'わたし',
      meaning: 'saya',
      example: '私は学生です。',
      exampleMeaning: 'Saya pelajar.',
      note: 'Kata ganti netral.',
    ),
  ],
);

const _attempt = VocabularyAttempt(
  id: 'attempt-1',
  lessonId: 'n5-daily',
  questions: [
    VocabularyQuestion(
      id: 'q1',
      prompt: 'Apa arti 私?',
      options: [
        QuizOption(id: 'correct-option', text: 'saya'),
        QuizOption(id: 'wrong-option', text: 'air'),
      ],
    ),
  ],
);

const _result = VocabularyResult(
  attemptId: 'attempt-1',
  lessonId: 'n5-daily',
  score: 0,
  correctAnswers: 0,
  totalQuestions: 1,
  passed: false,
  bestScore: 0,
  progress: 0,
  canRetry: true,
);
