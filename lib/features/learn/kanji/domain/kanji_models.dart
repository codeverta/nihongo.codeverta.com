import 'package:flutter/foundation.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart'
    show AnswerFeedback, QuizOption;

@immutable
class KanjiLessonSummary {
  const KanjiLessonSummary({
    required this.id,
    required this.order,
    required this.chapter,
    required this.title,
    required this.characterCount,
    required this.questionCount,
    required this.bestScore,
    required this.completed,
  });
  factory KanjiLessonSummary.fromJson(Map<String, dynamic> j) =>
      KanjiLessonSummary(
        id: j['id'] as String,
        order: j['order'] as int,
        chapter: j['chapter'] as String,
        title: j['title'] as String,
        characterCount: j['characterCount'] as int,
        questionCount: j['questionCount'] as int,
        bestScore: j['bestScore'] as int,
        completed: j['completed'] as bool,
      );
  final String id;
  final String chapter;
  final String title;
  final int order;
  final int characterCount;
  final int questionCount;
  final int bestScore;
  final bool completed;
}

@immutable
class KanjiCourse {
  const KanjiCourse({
    required this.level,
    required this.title,
    required this.progress,
    required this.completedLessons,
    required this.lessons,
  });
  factory KanjiCourse.fromJson(Map<String, dynamic> j) => KanjiCourse(
    level: j['level'] as String,
    title: j['title'] as String,
    progress: j['progress'] as int,
    completedLessons: j['completedLessons'] as int,
    lessons: _items(j['lessons'], KanjiLessonSummary.fromJson),
  );
  final String level;
  final String title;
  final int progress;
  final int completedLessons;
  final List<KanjiLessonSummary> lessons;
}

@immutable
class KanjiExample {
  const KanjiExample({
    required this.japanese,
    required this.reading,
    required this.meaning,
  });
  factory KanjiExample.fromJson(Map<String, dynamic> j) => KanjiExample(
    japanese: j['japanese'] as String,
    reading: j['reading'] as String,
    meaning: j['meaning'] as String,
  );
  final String japanese;
  final String reading;
  final String meaning;
}

@immutable
class KanjiCharacter {
  const KanjiCharacter({
    required this.character,
    required this.meaning,
    required this.onReading,
    required this.kunReading,
    required this.strokes,
    required this.mnemonic,
    required this.note,
    required this.examples,
  });
  factory KanjiCharacter.fromJson(Map<String, dynamic> j) => KanjiCharacter(
    character: j['character'] as String,
    meaning: j['meaning'] as String,
    onReading: j['onReading'] as String,
    kunReading: j['kunReading'] as String,
    strokes: j['strokes'] as int,
    mnemonic: j['mnemonic'] as String,
    note: j['note'] as String,
    examples: _items(j['examples'], KanjiExample.fromJson),
  );
  final String character;
  final String meaning;
  final String onReading;
  final String kunReading;
  final int strokes;
  final String mnemonic;
  final String note;
  final List<KanjiExample> examples;
}

@immutable
class KanjiLesson {
  const KanjiLesson({required this.summary, required this.characters});
  factory KanjiLesson.fromJson(Map<String, dynamic> j) => KanjiLesson(
    summary: KanjiLessonSummary.fromJson(j),
    characters: _items(j['characters'], KanjiCharacter.fromJson),
  );
  final KanjiLessonSummary summary;
  final List<KanjiCharacter> characters;
}

@immutable
class KanjiQuestion {
  const KanjiQuestion({
    required this.id,
    required this.prompt,
    required this.options,
  });
  factory KanjiQuestion.fromJson(Map<String, dynamic> j) => KanjiQuestion(
    id: j['id'] as String,
    prompt: j['prompt'] as String,
    options: _items(j['options'], QuizOption.fromJson),
  );
  final String id;
  final String prompt;
  final List<QuizOption> options;
}

@immutable
class KanjiAttempt {
  const KanjiAttempt({required this.id, required this.questions});
  factory KanjiAttempt.fromJson(Map<String, dynamic> j) => KanjiAttempt(
    id: j['id'] as String,
    questions: _items(j['questions'], KanjiQuestion.fromJson),
  );
  final String id;
  final List<KanjiQuestion> questions;
}

@immutable
class KanjiResult {
  const KanjiResult({
    required this.score,
    required this.correct,
    required this.total,
    required this.passed,
    required this.bestScore,
    required this.progress,
    this.verified = true,
  });
  factory KanjiResult.fromJson(Map<String, dynamic> j) => KanjiResult(
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

typedef KanjiFeedback = AnswerFeedback;
List<T> _items<T>(Object? value, T Function(Map<String, dynamic>) parse) =>
    (value! as List<dynamic>)
        .map((e) => parse(e as Map<String, dynamic>))
        .toList(growable: false);
