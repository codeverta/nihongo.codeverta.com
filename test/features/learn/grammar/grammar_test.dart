import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/grammar/data/grammar_repository.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/controllers/grammar_controller.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/screens/grammar_course_screen.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/screens/grammar_lesson_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart'
    show QuizOption;

void main() {
  testWidgets('N2 opens its full catalog and searches Japanese patterns', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(360, 640);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final content =
        jsonDecode(File('backend/data/grammar/n2.json').readAsStringSync())
            as List<dynamic>;
    final lessons = content.indexed.map((entry) {
      final data = entry.$2 as Map<String, dynamic>;
      return GrammarLessonSummary(
        id: 'grammar-n2-${(entry.$1 + 1).toString().padLeft(2, '0')}',
        level: 'N2',
        order: entry.$1 + 1,
        chapter: data['chapter'] as String,
        title: data['title'] as String,
        pointCount: (data['points'] as List<dynamic>).length,
        questionCount: (data['questions'] as List<dynamic>).length,
        patterns: (data['points'] as List<dynamic>)
            .map((p) => (p as Map<String, dynamic>)['pattern'] as String)
            .toList(),
        bestScore: 0,
        completed: false,
      );
    }).toList();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          grammarCourseLevelProvider('N2').overrideWith(
            (ref) async => GrammarCourse(
              level: 'N2',
              title: 'Grammar JLPT N2',
              progress: 0,
              completedLessons: 0,
              lessons: lessons,
            ),
          ),
        ],
        child: const MaterialApp(home: GrammarCourseScreen(initialLevel: 'N2')),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Grammar JLPT N2'), findsOneWidget);
    expect(find.text('60 lesson • 120 pola • 300 soal'), findsOneWidget);
    expect(tester.takeException(), isNull);
    final search = find.byKey(const ValueKey('grammar-search-N2'));
    await tester.scrollUntilVisible(
      search,
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.enterText(search, 'に先立って');
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byKey(const Key('grammar-lesson-grammar-n2-59')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Seolah-olah dan sebelum tahap penting'), findsOneWidget);
    expect(find.byKey(const Key('grammar-lesson-grammar-n2-01')), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('course displays saved progress and searches lesson titles', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [grammarCourseProvider.overrideWith((ref) async => _course)],
        child: const MaterialApp(home: GrammarCourseScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('1/2 lesson selesai • 50%'), findsOneWidget);
    await tester.ensureVisible(find.byKey(const ValueKey('grammar-search-N5')));
    await tester.enterText(
      find.byKey(const ValueKey('grammar-search-N5')),
      'Waktu',
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('grammar-lesson-grammar-n5-01')), findsNothing);
    await tester.scrollUntilVisible(
      find.byKey(const Key('grammar-lesson-grammar-n5-02')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Waktu lampau'), findsOneWidget);
  });

  testWidgets(
    'small phone reads two patterns and completes five-question quiz',
    (tester) async {
      tester.view.physicalSize = const Size(360, 640);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = _Repository();
      await tester.pumpWidget(_app(repository));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      expect(find.text('Nです'), findsOneWidget);
      final furiganaToggle = find.byKey(const Key('grammar-furigana-toggle'));
      expect(tester.widget<FilterChip>(furiganaToggle).selected, isTrue);
      final example = find.byKey(const Key('grammar-example-0-0'));
      await tester.scrollUntilVisible(
        example,
        120,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('わたし'), findsOneWidget);
      await tester.ensureVisible(furiganaToggle);
      await tester.tap(furiganaToggle);
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        example,
        120,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('わたし'), findsNothing);
      await tester.ensureVisible(furiganaToggle);
      await tester.tap(furiganaToggle);
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        example,
        120,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('わたし'), findsOneWidget);
      await tester.ensureVisible(find.text('Hindari kesalahan ini'));
      await tester.scrollUntilVisible(find.text('Hindari kesalahan ini'), 200);
      expect(find.text('Hindari kesalahan ini').hitTestable(), findsOneWidget);
      expect(
        find.byKey(const Key('grammar-next-point')).hitTestable(),
        findsOneWidget,
      );
      await tester.tap(find.byKey(const Key('grammar-next-point')));
      await tester.pumpAndSettle();
      expect(find.text('Nではありません'), findsOneWidget);
      await tester.tap(find.byKey(const Key('grammar-start-quiz')));
      await tester.pumpAndSettle();
      for (var i = 1; i <= 5; i++) {
        await tester.tap(find.byKey(Key('grammar-option-q$i-a')));
        await tester.pumpAndSettle();
        await tester.tap(find.byKey(const Key('grammar-submit')));
        await tester.pumpAndSettle();
        expect(find.byKey(const Key('grammar-feedback')), findsOneWidget);
        await tester.tap(find.byKey(const Key('grammar-submit')));
        await tester.pumpAndSettle();
      }
      expect(find.text('Lesson dikuasai!'), findsOneWidget);
      expect(find.text('100%'), findsOneWidget);
      expect(repository.completed, 1);
      expect(tester.takeException(), isNull);
      await tester.ensureVisible(find.byKey(const Key('grammar-retry')));
      await tester.tap(find.byKey(const Key('grammar-retry')));
      await tester.pumpAndSettle();
      expect(find.text('Pertanyaan 1/5'), findsOneWidget);
      expect(repository.started, 2);
    },
  );

  test('network errors preserve answer selection and can be retried', () async {
    final repository = _Repository()..failAnswer = true;
    final controller = GrammarController(
      repository,
      'grammar-n5-01',
      onProgress: () {},
    );
    addTearDown(controller.dispose);
    await controller.load();
    await controller.startQuiz();
    controller.select('q1-a');
    await controller.submit();
    expect(controller.state.value!.selection, 'q1-a');
    expect(controller.state.value!.stage, GrammarStage.quiz);
    expect(controller.state.value!.busy, isFalse);
    expect(controller.state.value!.error, isNotNull);
    await controller.submit();
    expect(controller.state.value!.feedback!.correct, isTrue);
    expect(controller.state.value!.error, isNull);
    await controller.nextQuestion();
    expect(controller.state.value!.questionIndex, 1);
    expect(controller.state.value!.selection, isNull);
    expect(controller.state.value!.feedback, isNull);
  });

  test('late lesson response after disposal does not write state', () async {
    final repository = _DelayedRepository();
    final controller = GrammarController(
      repository,
      'grammar-n5-01',
      onProgress: () {},
    );
    final loading = controller.load();
    controller.dispose();
    repository.response.complete(_lesson);
    await expectLater(loading, completes);
  });
}

Widget _app(_Repository repository) => ProviderScope(
  overrides: [
    grammarRepositoryProvider.overrideWithValue(repository),
    authControllerProvider.overrideWith(
      (ref) => AuthController.test(
        const AuthenticationState.authenticated(
          AuthenticatedUser(
            id: 'user',
            displayName: 'Learner',
            email: 'demo@example.com',
          ),
        ),
      ),
    ),
  ],
  child: MaterialApp(
    theme: ThemeData(
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(minimumSize: const Size.fromHeight(52)),
      ),
    ),
    home: const GrammarLessonScreen(lessonId: 'grammar-n5-01'),
  ),
);

class _Repository extends GrammarRepository {
  _Repository() : super(Dio());
  int started = 0;
  int completed = 0;
  bool failAnswer = false;
  @override
  Future<GrammarLesson> getLesson(String id) async => _lesson;
  @override
  Future<GrammarAttempt> startAttempt(String id) async {
    started++;
    return GrammarAttempt(
      id: 'attempt-$started',
      questions: [
        for (var i = 1; i <= 5; i++)
          GrammarQuestion(
            id: 'q$i',
            prompt: '私は学生（　）。',
            options: [
              QuizOption(id: 'q$i-a', text: 'です'),
              QuizOption(id: 'q$i-b', text: 'ます'),
            ],
          ),
      ],
    );
  }

  @override
  Future<GrammarFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async {
    if (failAnswer) {
      failAnswer = false;
      throw const GrammarException('Koneksi terputus. Coba lagi.');
    }
    return GrammarFeedback(
      questionId: question,
      correct: option.endsWith('-a'),
      correctOptionId: '$question-a',
      explanation: '学生 adalah kata benda, sehingga gunakan です.',
    );
  }

  @override
  Future<GrammarResult> complete(String id) async {
    completed++;
    return const GrammarResult(
      score: 100,
      correct: 5,
      total: 5,
      passed: true,
      bestScore: 100,
      progress: 50,
    );
  }
}

class _DelayedRepository extends _Repository {
  final response = Completer<GrammarLesson>();
  @override
  Future<GrammarLesson> getLesson(String id) => response.future;
}

const _summary = GrammarLessonSummary(
  id: 'grammar-n5-01',
  order: 1,
  chapter: 'Kalimat dasar',
  title: 'Identitas',
  pointCount: 2,
  questionCount: 5,
  bestScore: 100,
  completed: true,
);
const _course = GrammarCourse(
  level: 'N5',
  title: 'Grammar JLPT N5',
  progress: 50,
  completedLessons: 1,
  lessons: [
    _summary,
    GrammarLessonSummary(
      id: 'grammar-n5-02',
      order: 2,
      chapter: 'Kalimat dasar',
      title: 'Waktu lampau',
      pointCount: 2,
      questionCount: 5,
      bestScore: 0,
      completed: false,
    ),
  ],
);
const _lesson = GrammarLesson(
  summary: _summary,
  points: [
    GrammarPoint(
      pattern: 'Nです',
      meaning: 'adalah …',
      formation: 'Kata benda + です',
      explanation: 'です menutup pernyataan sopan tentang identitas.',
      commonMistake: 'Jangan menambahkan です setelah kata kerja ます.',
      examples: [
        GrammarExample(
          japanese: '私は学生です。',
          reading: 'わたしは がくせいです。',
          meaning: 'Saya pelajar.',
        ),
        GrammarExample(
          japanese: 'これは本です。',
          reading: 'これは ほんです。',
          meaning: 'Ini buku.',
        ),
      ],
    ),
    GrammarPoint(
      pattern: 'Nではありません',
      meaning: 'bukan …',
      formation: 'Kata benda + ではありません',
      explanation: 'Negatif sopan untuk kata benda.',
      commonMistake: 'Jangan memakai ない langsung setelah kata benda.',
      examples: [
        GrammarExample(
          japanese: '私は先生ではありません。',
          reading: 'わたしは せんせいではありません。',
          meaning: 'Saya bukan guru.',
        ),
        GrammarExample(
          japanese: 'これは本ではありません。',
          reading: 'これは ほんではありません。',
          meaning: 'Ini bukan buku.',
        ),
      ],
    ),
  ],
);
