import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/data/vocabulary_repository.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_course_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_lesson_screen.dart';

void main() {
  testWidgets('course renders real progress and best score from provider', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          vocabularyLevelsProvider.overrideWith((ref) async => _levels),
        ],
        child: const MaterialApp(home: VocabularyCourseScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Data level dari backend'), findsOneWidget);
    expect(find.byKey(const Key('level-progress-N5')), findsOneWidget);
    expect(find.textContaining('Progres latihan 64%'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Dikuasai • 80%'),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Dikuasai • 80%'), findsOneWidget);
  });

  testWidgets(
      'lesson stays visible on a small phone and toggles example translation', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(360, 640);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          vocabularyRepositoryProvider.overrideWithValue(_LessonRepository()),
        ],
        child: MaterialApp(
          theme: ThemeData(
            outlinedButtonTheme: OutlinedButtonThemeData(
              style: OutlinedButton.styleFrom(
                // Match the app theme's full-width button constraint.
                minimumSize: const Size.fromHeight(52),
              ),
            ),
          ),
          home: const VocabularyLessonScreen(lessonId: 'n5-daily'),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('KATA 1 / 1'), findsOneWidget);
    expect(find.byKey(const Key('word-word-1')), findsOneWidget);
    expect(find.byKey(const Key('revealed-word-meaning')), findsOneWidget);
    expect(find.text('saya'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Saya adalah pelajar.'),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.ensureVisible(find.text('Terjemahan'));
    expect(find.text('Saya adalah pelajar.'), findsOneWidget);
    await tester.tap(find.text('Terjemahan'));
    await tester.pumpAndSettle();
    expect(find.text('Saya adalah pelajar.'), findsNothing);
    await tester.tap(find.text('Terjemahan'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Saya adalah pelajar.'),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Saya adalah pelajar.'), findsOneWidget);
    await tester.ensureVisible(find.text('私は学生です。'));
    expect(find.text('私は学生です。'), findsOneWidget);
    expect(tester.takeException(), isNull);

    await tester.scrollUntilVisible(
      find.text('Catatan pemakaian'),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.ensureVisible(find.text('Catatan pemakaian'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Catatan pemakaian'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Kata ganti netral.'));
    await tester.pumpAndSettle();
    expect(find.text('Kata ganti netral.').hitTestable(), findsOneWidget);
    expect(
      find.byKey(const Key('start-vocabulary-quiz')).hitTestable(),
      findsOneWidget,
    );
    expect(
      tester
          .widget<FilledButton>(find.byKey(const Key('start-vocabulary-quiz')))
          .onPressed,
      isNotNull,
    );
  });
}

class _LessonRepository extends VocabularyRepository {
  _LessonRepository() : super(Dio());

  @override
  Future<VocabularyLesson> getLesson(String lessonId) async =>
      const VocabularyLesson(
        id: 'n5-daily',
        level: 'N5',
        title: 'Kata dasar',
        description: 'Lesson pemula',
        bestScore: 0,
        completed: false,
        imageUrl: '',
        words: [
          VocabularyWord(
            id: 'word-1',
            term: '私',
            reading: 'わたし',
            meaning: 'saya',
            example: '私は学生です。',
            exampleMeaning: 'Saya adalah pelajar.',
            note: 'Kata ganti netral.',
          ),
        ],
      );
}

const _levels = [
  VocabularyLevel(
    level: 'N5',
    title: 'Fondasi API',
    description: 'Data level dari backend',
    requiredTier: 'free',
    progress: 64,
    lessons: [
      VocabularyLessonSummary(
        id: 'n5-daily',
        order: 1,
        title: 'Kata dasar API',
        description: 'Data lesson dari backend',
        wordCount: 10,
        bestScore: 80,
        completed: true,
      ),
    ],
  ),
];
