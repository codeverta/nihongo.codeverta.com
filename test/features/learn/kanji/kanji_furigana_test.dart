import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/kanji/data/kanji_repository.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/controllers/kanji_controller.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/screens/kanji_lesson_screen.dart';

void main() {
  testWidgets('kanji lesson shows furigana above the kanji by default', (
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
          kanjiRepositoryProvider.overrideWithValue(_KanjiRepository()),
        ],
        child: const MaterialApp(
          home: KanjiLessonScreen(lessonId: 'kanji-test'),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final toggle = find.byKey(const Key('kanji-furigana-toggle'));
    expect(find.text('おや'), findsOneWidget);
    await tester.scrollUntilVisible(
      toggle,
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.ensureVisible(toggle);
    expect(tester.widget<FilterChip>(toggle).selected, isTrue);

    await tester.tap(toggle);
    await tester.pumpAndSettle();
    expect(find.text('おや'), findsNothing);

    await tester.scrollUntilVisible(
      find.byKey(const Key('kanji-example-0')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('おや'), findsNothing);

    await tester.scrollUntilVisible(
      toggle,
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.ensureVisible(toggle);
    await tester.tap(toggle);
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byKey(const Key('kanji-example-0')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('おや'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}

class _KanjiRepository extends KanjiRepository {
  _KanjiRepository() : super(Dio());

  @override
  Future<KanjiLesson> getLesson(String id) async => const KanjiLesson(
    summary: KanjiLessonSummary(
      id: 'kanji-test',
      order: 1,
      chapter: 'Keluarga',
      title: 'Orang tua',
      characterCount: 1,
      questionCount: 0,
      bestScore: 0,
      completed: false,
    ),
    characters: [
      KanjiCharacter(
        character: '親',
        meaning: 'orang tua',
        onReading: 'シン',
        kunReading: 'おや',
        strokes: 16,
        mnemonic: 'Petunjuk.',
        note: 'Catatan.',
        examples: [
          KanjiExample(japanese: '親', reading: 'おや', meaning: 'orang tua'),
        ],
      ),
    ],
  );
}
