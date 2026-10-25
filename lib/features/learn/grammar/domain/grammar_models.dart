import 'package:flutter/foundation.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart'
    show AnswerFeedback, QuizOption;

@immutable
class GrammarLessonSummary {
  const GrammarLessonSummary({
    required this.id,
    required this.order,
    required this.chapter,
    required this.title,
    required this.pointCount,
    required this.questionCount,
    required this.bestScore,
    required this.completed,
    this.level = 'N5',
    this.patterns = const [],
  });
  factory GrammarLessonSummary.fromJson(Map<String, dynamic> j) =>
      GrammarLessonSummary(
        id: j['id'] as String,
        order: j['order'] as int,
        chapter: j['chapter'] as String,
        title: j['title'] as String,
        pointCount: j['pointCount'] as int,
        questionCount: j['questionCount'] as int,
        bestScore: j['bestScore'] as int,
        completed: j['completed'] as bool,
        level: j['level'] as String? ?? 'N5',
        patterns: (j['patterns'] as List<dynamic>? ?? const [])
            .map((p) => p as String)
            .toList(),
      );
  final String id;
  final String level;
  final List<String> patterns;
  final String chapter;
  final String title;
  final int order;
  final int pointCount;
  final int questionCount;
  final int bestScore;
  final bool completed;
}

@immutable
class GrammarCourse {
  const GrammarCourse({
    required this.level,
    required this.title,
    required this.progress,
    required this.completedLessons,
    required this.lessons,
  });
  factory GrammarCourse.fromJson(Map<String, dynamic> j) => GrammarCourse(
    level: j['level'] as String,
    title: j['title'] as String,
    progress: j['progress'] as int,
    completedLessons: j['completedLessons'] as int,
    lessons: _items(j['lessons'], GrammarLessonSummary.fromJson),
  );
  final String level;
  final String title;
  final int progress;
  final int completedLessons;
  final List<GrammarLessonSummary> lessons;
}

@immutable
class GrammarExample {
  const GrammarExample({
    required this.japanese,
    required this.reading,
    required this.meaning,
  });
  factory GrammarExample.fromJson(Map<String, dynamic> j) => GrammarExample(
    japanese: j['japanese'] as String,
    reading: j['reading'] as String,
    meaning: j['meaning'] as String,
  );
  final String japanese;
  final String reading;
  final String meaning;
}

@immutable
class GrammarPoint {
  const GrammarPoint({
    required this.pattern,
    required this.meaning,
    required this.formation,
    required this.explanation,
    required this.commonMistake,
    required this.examples,
  });
  factory GrammarPoint.fromJson(Map<String, dynamic> j) => GrammarPoint(
    pattern: j['pattern'] as String,
    meaning: j['meaning'] as String,
    formation: j['formation'] as String,
    explanation: j['explanation'] as String,
    commonMistake: j['commonMistake'] as String,
    examples: _items(j['examples'], GrammarExample.fromJson),
  );
  final String pattern;
  final String meaning;
  final String formation;
  final String explanation;
  final String commonMistake;
  final List<GrammarExample> examples;
}

@immutable
class GrammarLesson {
  const GrammarLesson({required this.summary, required this.points});
  factory GrammarLesson.fromJson(Map<String, dynamic> j) => GrammarLesson(
    summary: GrammarLessonSummary.fromJson(j),
    points: _items(j['points'], GrammarPoint.fromJson),
  );
  final GrammarLessonSummary summary;
  final List<GrammarPoint> points;
}

@immutable
class GrammarQuestion {
  const GrammarQuestion({
    required this.id,
    required this.prompt,
    required this.options,
  });
  factory GrammarQuestion.fromJson(Map<String, dynamic> j) => GrammarQuestion(
    id: j['id'] as String,
    prompt: j['prompt'] as String,
    options: _items(j['options'], QuizOption.fromJson),
  );
  final String id;
  final String prompt;
  final List<QuizOption> options;
}

@immutable
class GrammarAttempt {
  const GrammarAttempt({required this.id, required this.questions});
  factory GrammarAttempt.fromJson(Map<String, dynamic> j) => GrammarAttempt(
    id: j['id'] as String,
    questions: _items(j['questions'], GrammarQuestion.fromJson),
  );
  final String id;
  final List<GrammarQuestion> questions;
}

@immutable
class GrammarResult {
  const GrammarResult({
    required this.score,
    required this.correct,
    required this.total,
    required this.passed,
    required this.bestScore,
    required this.progress,
    this.verified = true,
  });
  factory GrammarResult.fromJson(Map<String, dynamic> j) => GrammarResult(
    score: j['score'] as int,
    correct: j['correctAnswers'] as int,
    total: j['totalQuestions'] as int,
    passed: j['passed'] as bool,
    bestScore: j['bestScore'] as int,
    progress: j['progress'] as int,
    verified: j['verified'] != false,
  );
  final int score;
  final int correct;
  final int total;
  final int bestScore;
  final int progress;
  final bool passed;
  final bool verified;
  bool get pendingSync => !verified;
}

typedef GrammarFeedback = AnswerFeedback;
List<T> _items<T>(Object? value, T Function(Map<String, dynamic>) parse) =>
    (value! as List<dynamic>)
        .map((e) => parse(e as Map<String, dynamic>))
        .toList(growable: false);
