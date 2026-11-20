import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

class _SwipeRepository extends AnkiRepository {
  _SwipeRepository({this.fail = false}) : super(Dio());
  final bool fail;
  final ratings = <String>[];

  @override
  Future<AnkiSession> study(String id) async => AnkiSession.fromJson({
    'title': 'Swipe deck',
    'cards': [
      {
        'index': 0,
        'front': '猫',
        'answers': ['ねこ'],
        'meaning': 'kucing',
      },
      {
        'index': 1,
        'front': '犬',
        'answers': ['いぬ'],
        'meaning': 'anjing',
      },
    ],
    'stats': {'due': 2},
  });

  @override
  Future<Map<String, dynamic>> review(
    String id,
    AnkiCard card,
    String rating,
    String requestId,
  ) async {
    ratings.add(rating);
    if (fail) throw const AnkiFailure('Penilaian belum tersimpan.', null);
    return {'saved': true, 'version': 1};
  }
}

const _card = Key('anki-swipe-card');
const _back = Key('anki-card-back');

Future<void> _pumpStudy(
  WidgetTester tester,
  AnkiRepository repository, {
  bool reduceMotion = false,
}) async {
  tester.view.physicalSize = const Size(400, 860);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        ankiRepositoryProvider.overrideWithValue(repository),
        authControllerProvider.overrideWith(
          (ref) => AuthController.test(
            const AuthenticationState.authenticated(
              AuthenticatedUser(
                id: 'learner',
                displayName: 'Learner',
                email: 'learner@example.test',
              ),
            ),
          ),
        ),
      ],
      child: MaterialApp(
        theme: AppTheme.lightTheme,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(
            context,
          ).copyWith(disableAnimations: reduceMotion),
          child: child!,
        ),
        home: const AnkiStudyScreen(deckId: 'swipe'),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

String _topFront(WidgetTester tester) =>
    tester.widget<Text>(find.byKey(const Key('anki-card-front'))).data!;

double _cardOffset(WidgetTester tester) => tester
    .widget<Transform>(
      find
          .descendant(of: find.byKey(_card), matching: find.byType(Transform))
          .first,
    )
    .transform
    .getTranslation()
    .x;

void main() {
  testWidgets('tapping the card flips it front to back and back again', (
    tester,
  ) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    expect(find.byKey(_back), findsNothing);

    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();
    expect(find.byKey(_back), findsOneWidget);
    expect(find.text('ねこ'), findsOneWidget);
    expect(find.byKey(const Key('anki-rate-good')), findsOneWidget);

    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();
    expect(find.byKey(_back), findsNothing);
    expect(repository.ratings, isEmpty);
    expect(tester.takeException(), isNull);
  });

  testWidgets('swiping is ignored until the answer is shown', (tester) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    await tester.drag(find.byKey(_card), const Offset(300, 0));
    await tester.pumpAndSettle();
    expect(repository.ratings, isEmpty);
    expect(_topFront(tester), '猫');
    expect(_cardOffset(tester), 0);
  });

  testWidgets('swipe right rates "good" and the next card moves up', (
    tester,
  ) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    // The next card peeks out from behind the top card.
    expect(find.text('犬'), findsOneWidget);
    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();

    final gesture = await tester.startGesture(
      tester.getCenter(find.byKey(_card)),
    );
    await gesture.moveBy(const Offset(20, 0));
    await gesture.moveBy(const Offset(200, 0));
    await tester.pump();
    expect(find.byKey(const Key('anki-swipe-right')), findsOneWidget);
    expect(find.byKey(const Key('anki-swipe-left')), findsNothing);
    await gesture.up();
    await tester.pumpAndSettle();

    expect(repository.ratings, ['good']);
    expect(_topFront(tester), '犬');
    expect(find.byKey(_back), findsNothing);
    expect(_cardOffset(tester), 0);
    expect(tester.takeException(), isNull);
  });

  testWidgets('swipe left rates "again"', (tester) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();
    await tester.drag(find.byKey(_card), const Offset(-260, 0));
    await tester.pumpAndSettle();
    expect(repository.ratings, ['again']);
    expect(_topFront(tester), '犬');
  });

  testWidgets('a fast fling past the slop counts as a swipe', (tester) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();
    await tester.fling(find.byKey(_card), const Offset(-80, 0), 2000);
    await tester.pumpAndSettle();
    expect(repository.ratings, ['again']);
  });

  testWidgets('a short slow swipe springs back without rating', (tester) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository);
    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();

    final gesture = await tester.startGesture(
      tester.getCenter(find.byKey(_card)),
    );
    for (var i = 0; i < 6; i++) {
      await gesture.moveBy(const Offset(10, 0));
      await tester.pump(const Duration(milliseconds: 100));
    }
    expect(_cardOffset(tester), greaterThan(0));
    await tester.pump(const Duration(milliseconds: 300));
    await gesture.up();
    await tester.pumpAndSettle();

    expect(repository.ratings, isEmpty);
    expect(_cardOffset(tester), closeTo(0, .5));
    expect(_topFront(tester), '猫');
    // Still revealed: the swipe did not flip or advance the card.
    expect(find.byKey(_back), findsOneWidget);
  });

  testWidgets('a failed review brings the card back with the retry button', (
    tester,
  ) async {
    final repository = _SwipeRepository(fail: true);
    await _pumpStudy(tester, repository);
    await tester.tap(find.byKey(_card));
    await tester.pumpAndSettle();
    await tester.drag(find.byKey(_card), const Offset(260, 0));
    await tester.pumpAndSettle();

    expect(repository.ratings, ['good']);
    expect(_topFront(tester), '猫');
    expect(_cardOffset(tester), closeTo(0, .5));
    expect(
      find.textContaining(RegExp('Coba simpan kembali|Try saving again')),
      findsOneWidget,
    );
    // Swiping the other way must not change the pending rating.
    await tester.drag(find.byKey(_card), const Offset(-260, 0));
    await tester.pumpAndSettle();
    expect(repository.ratings, ['good']);
  });

  testWidgets('reduced motion still swipes, without animations', (
    tester,
  ) async {
    final repository = _SwipeRepository();
    await _pumpStudy(tester, repository, reduceMotion: true);
    await tester.tap(find.byKey(_card));
    await tester.pump();
    await tester.drag(find.byKey(_card), const Offset(260, 0));
    // No fly-out animation: one frame later the review is already saved.
    await tester.pump();
    expect(repository.ratings, ['good']);
    await tester.pumpAndSettle();
    expect(_topFront(tester), '犬');
    expect(tester.takeException(), isNull);
  });
}
