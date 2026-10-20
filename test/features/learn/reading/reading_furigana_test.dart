import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/reading/data/reading_repository.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';
import 'package:nihongo_master/features/learn/reading/presentation/controllers/reading_controller.dart';
import 'package:nihongo_master/features/learn/reading/presentation/screens/reading_lesson_screen.dart';

void main() {
  testWidgets(
    'reading lesson shows ruby over kanji by default and toggles it',
    (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(const AuthenticationState.guest()),
            ),
            readingRepositoryProvider.overrideWithValue(_ReadingRepository()),
          ],
          child: const MaterialApp(
            home: ReadingLessonScreen(lessonId: 'reading-test'),
          ),
        ),
      );
      await tester.pumpAndSettle();

      final toggle = find.byKey(const Key('reading-furigana-toggle'));
      await tester.scrollUntilVisible(
        toggle,
        100,
        scrollable: find.byType(Scrollable).first,
      );
      expect(toggle, findsOneWidget);
      expect(
        tester
            .widget<FilterChip>(
              find.byKey(const Key('reading-furigana-toggle')),
            )
            .selected,
        isTrue,
      );
      expect(find.text('きのう'), findsOneWidget);
      await tester.scrollUntilVisible(
        find.byKey(const Key('reading-passage')),
        100,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('きのう'), findsOneWidget);

      await tester.scrollUntilVisible(
        toggle,
        100,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(toggle);
      await tester.pumpAndSettle();
      expect(find.text('きのう'), findsNothing);
      expect(find.text('昨日は休みでした。'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
}

class _ReadingRepository extends ReadingRepository {
  _ReadingRepository() : super(Dio());

  @override
  Future<ReadingLesson> getLesson(String id) async => const ReadingLesson(
    summary: ReadingLessonSummary(
      id: 'reading-test',
      japaneseTitle: '休みの日',
      description: 'Contoh reading.',
      minutes: 1,
      imageAsset: '',
      imageUrl: '',
      order: 1,
      chapter: 'Keseharian',
      title: 'Hari libur',
      pageCount: 1,
      questionCount: 0,
      bestScore: 0,
      completed: false,
    ),
    pages: [
      ReadingPage(
        title: 'Hari libur',
        japanese: '昨日は休みでした。',
        reading: 'きのうはやすみでした。',
        meaning: 'Kemarin adalah hari libur.',
        imageAsset: '',
        imageUrl: '',
        imagePanel: 0,
        imageDescription: '',
        vocabulary: [
          ReadingWord(japanese: '昨日', reading: 'きのう', meaning: 'kemarin'),
        ],
      ),
    ],
  );
}
