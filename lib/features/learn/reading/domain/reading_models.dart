import 'package:flutter/foundation.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart'
    show AnswerFeedback, QuizOption;

@immutable
class ReadingLessonSummary {
  const ReadingLessonSummary({
    required this.id,
    required this.japaneseTitle,
    required this.description,
    required this.minutes,
    required this.imageAsset,
    required this.imageUrl,
    required this.order,
    required this.chapter,
    required this.title,
    required this.pageCount,
    required this.questionCount,
    required this.bestScore,
    required this.completed,
  });
  factory ReadingLessonSummary.fromJson(Map<String, dynamic> j) =>
      ReadingLessonSummary(
        id: j['id'] as String,
        japaneseTitle: j['japaneseTitle'] as String,
        description: j['description'] as String,
        minutes: j['minutes'] as int,
        imageAsset: j['imageAsset'] as String,
        imageUrl: j['imageUrl'] as String,
        order: j['order'] as int,
        chapter: j['chapter'] as String,
        title: j['title'] as String,
        pageCount: j['pageCount'] as int,
        questionCount: j['questionCount'] as int,
        bestScore: j['bestScore'] as int,
        completed: j['completed'] as bool,
      );
  final String japaneseTitle;
  final String description;
  final int minutes;
  final String imageAsset;
  final String imageUrl;
  final String id;
  final String chapter;
  final String title;
  final int order;
  final int pageCount;
  final int questionCount;
  final int bestScore;
  final bool completed;
}

@immutable
class ReadingCourse {
  const ReadingCourse({
    required this.level,
    required this.title,
    required this.progress,
    required this.completedLessons,
    required this.lessons,
  });
  factory ReadingCourse.fromJson(Map<String, dynamic> j) => ReadingCourse(
        level: j['level'] as String,
        title: j['title'] as String,
        progress: j['progress'] as int,
        completedLessons: j['completedLessons'] as int,
        lessons: _items(j['lessons'], ReadingLessonSummary.fromJson),
      );
  final String level;
  final String title;
  final int progress;
  final int completedLessons;
  final List<ReadingLessonSummary> lessons;
}

@immutable
class ReadingWord {
  const ReadingWord({
    required this.japanese,
    required this.reading,
    required this.meaning,
  });
  factory ReadingWord.fromJson(Map<String, dynamic> j) => ReadingWord(
        japanese: j['japanese'] as String,
        reading: j['reading'] as String,
        meaning: j['meaning'] as String,
      );
  final String japanese;
  final String reading;
  final String meaning;
}

@immutable
class ReadingPage {
  const ReadingPage({
    required this.title,
    required this.japanese,
    required this.reading,
    required this.meaning,
    required this.imageAsset,
    required this.imageUrl,
    required this.imagePanel,
    required this.imageDescription,
    required this.vocabulary,
  });
  factory ReadingPage.fromJson(Map<String, dynamic> j) => ReadingPage(
        title: j['title'] as String,
        japanese: j['japanese'] as String,
        reading: j['reading'] as String,
        meaning: j['meaning'] as String,
        imageAsset: j['imageAsset'] as String,
        imageUrl: j['imageUrl'] as String,
        imagePanel: j['imagePanel'] as int,
        imageDescription: j['imageDescription'] as String,
        vocabulary: _items(j['vocabulary'], ReadingWord.fromJson),
      );
  final String title;
  final String japanese;
  final String reading;
  final String meaning;
  final String imageAsset;
  final String imageUrl;
  final int imagePanel;
  final String imageDescription;
  final List<ReadingWord> vocabulary;
}

@immutable
class ReadingLesson {
  const ReadingLesson({required this.summary, required this.pages});
  factory ReadingLesson.fromJson(Map<String, dynamic> j) => ReadingLesson(
        summary: ReadingLessonSummary.fromJson(j),
        pages: _items(j['pages'], ReadingPage.fromJson),
      );
  final ReadingLessonSummary summary;
  final List<ReadingPage> pages;
}

@immutable
class ReadingQuestion {
  const ReadingQuestion({
    required this.id,
    required this.prompt,
    required this.options,
  });
  factory ReadingQuestion.fromJson(Map<String, dynamic> j) => ReadingQuestion(
        id: j['id'] as String,
        prompt: j['prompt'] as String,
        options: _items(j['options'], QuizOption.fromJson),
      );
  final String id;
  final String prompt;
  final List<QuizOption> options;
}

@immutable
class ReadingAttempt {
  const ReadingAttempt({required this.id, required this.questions});
  factory ReadingAttempt.fromJson(Map<String, dynamic> j) => ReadingAttempt(
        id: j['id'] as String,
        questions: _items(j['questions'], ReadingQuestion.fromJson),
      );
  final String id;
  final List<ReadingQuestion> questions;
}

@immutable
class ReadingResult {
  const ReadingResult({
    required this.score,
    required this.correct,
    required this.total,
    required this.passed,
    required this.bestScore,
    required this.progress,
  });
  factory ReadingResult.fromJson(Map<String, dynamic> j) => ReadingResult(
        score: j['score'] as int,
        correct: j['correctAnswers'] as int,
        total: j['totalQuestions'] as int,
        passed: j['passed'] as bool,
        bestScore: j['bestScore'] as int,
        progress: j['progress'] as int,
      );
  final int score;
  final int correct;
  final int total;
  final int bestScore;
  final int progress;
  final bool passed;
}

typedef ReadingFeedback = AnswerFeedback;
List<T> _items<T>(Object? value, T Function(Map<String, dynamic>) parse) =>
    (value! as List<dynamic>)
        .map((e) => parse(e as Map<String, dynamic>))
        .toList(growable: false);
