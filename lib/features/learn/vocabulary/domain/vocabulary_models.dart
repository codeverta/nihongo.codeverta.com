import 'package:flutter/foundation.dart';

@immutable
class VocabularyLevel {
  const VocabularyLevel({
    required this.level,
    required this.title,
    required this.description,
    required this.requiredTier,
    required this.progress,
    required this.lessons,
  });

  factory VocabularyLevel.fromJson(Map<String, dynamic> json) =>
      VocabularyLevel(
        level: json['level'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        requiredTier: json['requiredTier'] as String? ?? 'free',
        progress: json['progress'] as int? ?? 0,
        lessons: _list(json['lessons'], VocabularyLessonSummary.fromJson),
      );

  final String level;
  final String title;
  final String description;
  final String requiredTier;
  final int progress;
  final List<VocabularyLessonSummary> lessons;
}

@immutable
class VocabularyLessonSummary {
  const VocabularyLessonSummary({
    required this.id,
    required this.order,
    required this.title,
    required this.description,
    required this.wordCount,
    required this.bestScore,
    required this.completed,
    this.imageAsset = '',
    this.questionCount = 0,
  });

  factory VocabularyLessonSummary.fromJson(Map<String, dynamic> json) =>
      VocabularyLessonSummary(
        imageAsset: json['imageAsset'] as String? ?? '',
        questionCount: json['questionCount'] as int? ?? 0,
        id: json['id'] as String? ?? '',
        order: json['order'] as int? ?? 0,
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        wordCount: json['wordCount'] as int? ?? 0,
        bestScore: json['bestScore'] as int? ?? 0,
        completed: json['completed'] as bool? ?? false,
      );

  final String id;
  final String imageAsset;
  final int questionCount;
  final int order;
  final String title;
  final String description;
  final int wordCount;
  final int bestScore;
  final bool completed;
}

@immutable
class VocabularyLesson {
  const VocabularyLesson({
    required this.id,
    required this.level,
    required this.title,
    required this.description,
    required this.bestScore,
    required this.completed,
    required this.imageUrl,
    required this.words,
  });

  factory VocabularyLesson.fromJson(Map<String, dynamic> json) =>
      VocabularyLesson(
        id: json['id'] as String? ?? '',
        level: json['level'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        bestScore: json['bestScore'] as int? ?? 0,
        completed: json['completed'] as bool? ?? false,
        imageUrl: json['imageUrl'] as String? ?? '',
        words: _list(json['words'], VocabularyWord.fromJson),
      );

  final String id;
  final String level;
  final String title;
  final String description;
  final int bestScore;
  final bool completed;
  final String imageUrl;
  final List<VocabularyWord> words;
}

@immutable
class VocabularyWord {
  const VocabularyWord({
    required this.id,
    required this.term,
    required this.reading,
    required this.meaning,
    required this.example,
    required this.exampleMeaning,
    required this.note,
    this.examples = const [],
    this.imageAsset = '',
    this.imageUrl = '',
    this.imagePanel = 0,
    this.imageDescription = '',
  });

  factory VocabularyWord.fromJson(Map<String, dynamic> json) => VocabularyWord(
        examples: _list(json['examples'], VocabularyExample.fromJson),
        imageAsset: json['imageAsset'] as String? ?? '',
        imageUrl: json['imageUrl'] as String? ?? '',
        imagePanel: json['imagePanel'] as int? ?? 0,
        imageDescription: json['imageDescription'] as String? ?? '',
        id: json['id'] as String? ?? '',
        term: json['term'] as String? ?? '',
        reading: json['reading'] as String? ?? '',
        meaning: json['meaning'] as String? ?? '',
        example: json['example'] as String? ?? '',
        exampleMeaning: json['exampleMeaning'] as String? ?? '',
        note: json['note'] as String? ?? '',
      );

  final String id;
  final String term;
  final String reading;
  final String meaning;
  final String example;
  final String exampleMeaning;
  final String note;
  final List<VocabularyExample> examples;
  final String imageAsset;
  final String imageUrl;
  final int imagePanel;
  final String imageDescription;
}

@immutable
class VocabularyExample {
  const VocabularyExample({
    required this.japanese,
    required this.reading,
    required this.meaning,
  });
  factory VocabularyExample.fromJson(Map<String, dynamic> json) =>
      VocabularyExample(
        japanese: json['japanese'] as String? ?? '',
        reading: json['reading'] as String? ?? '',
        meaning: json['meaning'] as String? ?? '',
      );
  final String japanese;
  final String reading;
  final String meaning;
}

@immutable
class VocabularyAttempt {
  const VocabularyAttempt({
    required this.id,
    required this.lessonId,
    required this.questions,
    required this.imageUrl,
  });

  factory VocabularyAttempt.fromJson(Map<String, dynamic> json) =>
      VocabularyAttempt(
        id: json['id'] as String? ?? '',
        lessonId: json['lessonId'] as String? ?? '',
        questions: _list(json['questions'], VocabularyQuestion.fromJson),
        imageUrl: json['imageUrl'] as String? ?? '',
      );

  final String id;
  final String lessonId;
  final List<VocabularyQuestion> questions;
  final String imageUrl;
}

@immutable
class VocabularyQuestion {
  const VocabularyQuestion({
    required this.id,
    required this.prompt,
    required this.options,
    this.kind = 'Arti kata',
    this.imageAsset = '',
    this.imageUrl = '',
    this.imagePanel = 0,
    this.imageDescription = '',
  });

  factory VocabularyQuestion.fromJson(Map<String, dynamic> json) =>
      VocabularyQuestion(
        kind: json['kind'] as String? ?? 'Arti kata',
        imageAsset: json['imageAsset'] as String? ?? '',
        imageUrl: json['imageUrl'] as String? ?? '',
        imagePanel: json['imagePanel'] as int? ?? 0,
        imageDescription: json['imageDescription'] as String? ?? '',
        id: json['id'] as String? ?? '',
        prompt: json['prompt'] as String? ?? '',
        options: _list(json['options'], QuizOption.fromJson),
      );

  final String id;
  final String prompt;
  final List<QuizOption> options;
  final String kind;
  final String imageAsset;
  final String imageUrl;
  final int imagePanel;
  final String imageDescription;
}

@immutable
class QuizOption {
  const QuizOption({required this.id, required this.text});

  factory QuizOption.fromJson(Map<String, dynamic> json) => QuizOption(
        id: json['id'] as String? ?? '',
        text: json['text'] as String? ?? '',
      );

  final String id;
  final String text;
}

@immutable
class AnswerFeedback {
  const AnswerFeedback({
    required this.questionId,
    required this.correct,
    required this.correctOptionId,
    required this.explanation,
  });

  factory AnswerFeedback.fromJson(Map<String, dynamic> json) => AnswerFeedback(
        questionId: json['questionId'] as String? ?? '',
        correct: json['correct'] as bool? ?? false,
        correctOptionId: json['correctOptionId'] as String? ?? '',
        explanation: json['explanation'] as String? ?? '',
      );

  final String questionId;
  final bool correct;
  final String correctOptionId;
  final String explanation;
}

@immutable
class VocabularyResult {
  const VocabularyResult({
    required this.attemptId,
    required this.lessonId,
    required this.score,
    required this.correctAnswers,
    required this.totalQuestions,
    required this.passed,
    required this.bestScore,
    required this.progress,
    required this.canRetry,
  });

  factory VocabularyResult.fromJson(Map<String, dynamic> json) =>
      VocabularyResult(
        attemptId: json['attemptId'] as String? ?? '',
        lessonId: json['lessonId'] as String? ?? '',
        score: json['score'] as int? ?? 0,
        correctAnswers: json['correctAnswers'] as int? ?? 0,
        totalQuestions: json['totalQuestions'] as int? ?? 0,
        passed: json['passed'] as bool? ?? false,
        bestScore: json['bestScore'] as int? ?? 0,
        progress: json['progress'] as int? ?? 0,
        canRetry: json['canRetry'] as bool? ?? true,
      );

  final String attemptId;
  final String lessonId;
  final int score;
  final int correctAnswers;
  final int totalQuestions;
  final bool passed;
  final int bestScore;
  final int progress;
  final bool canRetry;
}

List<T> _list<T>(Object? value, T Function(Map<String, dynamic>) fromJson) {
  return (value as List<dynamic>? ?? const [])
      .whereType<Map<String, dynamic>>()
      .map(fromJson)
      .toList(growable: false);
}
