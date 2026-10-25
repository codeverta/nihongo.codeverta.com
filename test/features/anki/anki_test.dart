import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/anki/presentation/anki_controller.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

List<Map<String, dynamic>> _sourceDecks() =>
    ((jsonDecode(
                  File(
                    'backend/internal/anki/resources/catalog.json',
                  ).readAsStringSync(),
                )
                as Map<String, dynamic>)['decks']
            as List<dynamic>)
        .cast<Map<String, dynamic>>();

class StudyRepository extends AnkiRepository {
  StudyRepository({this.failFirst = false}) : super(Dio());
  final bool failFirst;
  final reviews = <Map<String, dynamic>>[];
  final catalog = _sourceDecks();
  @override
  Future<List<AnkiDeck>> decks() async =>
      catalog.map(AnkiDeck.fromJson).toList();
  @override
  Future<AnkiDeck> deck(String id) async => AnkiDeck.fromJson({
    ...catalog.firstWhere((d) => d['id'] == id),
    'stats': {'seen': reviews.isEmpty ? 0 : 1, 'due': 0, 'new': 100},
    'locked': false,
  });
  @override
  Future<AnkiSession> study(String id) async {
    final deck = catalog.firstWhere((d) => d['id'] == id);
    final card = (deck['preview'] as List).first as Map<String, dynamic>;
    return AnkiSession.fromJson({
      'title': deck['title'],
      'instructions': deck['instructions'],
      'cards': [
        {
          ...card,
          'version': 0,
          'intervals': {
            'again': '1 menit',
            'hard': '6 jam',
            'good': '1 hari',
            'easy': '4 hari',
          },
        },
      ],
      'stats': {'due': 0},
    });
  }

  @override
  Future<Map<String, dynamic>> review(
    String id,
    AnkiCard card,
    String rating,
    String requestId,
  ) async {
    reviews.add({
      'id': id,
      'index': card.index,
      'version': card.version,
      'rating': rating,
      'requestId': requestId,
    });
    await Future<void>.value();
    if (failFirst && reviews.length == 1) {
      throw const AnkiFailure('Review belum terkonfirmasi.', null);
    }
    return {'saved': true, 'version': 1, 'due': '2026-10-18T08:00:00Z'};
  }
}

void main() {
  testWidgets('Anki deck search and empty state use English labels', (
    tester,
  ) async {
    final repository = StudyRepository();
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
        child: const MaterialApp(
          locale: Locale('en'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: AnkiDeckListScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(
      find.text('A little today.\nRemember more tomorrow.'),
      findsOneWidget,
    );
    expect(find.text('Search decks'), findsOneWidget);
    await tester.enterText(
      find.byKey(const Key('anki-search')),
      'no such deck',
    );
    await tester.pumpAndSettle();
    expect(
      find.text('No decks found. Try another name or category.'),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });

  test(
    'flip required, double ratings ignored, uncertain retry preserves request identity',
    () async {
      final repository = StudyRepository(failFirst: true);
      final controller = AnkiStudyController(repository, 'wakayama');
      addTearDown(controller.dispose);
      await controller.load();
      await controller.rate('good');
      expect(repository.reviews, isEmpty);
      controller.reveal();
      await Future.wait([controller.rate('good'), controller.rate('easy')]);
      expect(repository.reviews.length, 1);
      expect(controller.state.position, 0);
      expect(controller.state.error, isNotNull);
      controller.reveal();
      await controller.rate('easy');
      expect(repository.reviews.length, 2);
      expect(repository.reviews[0], repository.reviews[1]);
      expect(
        RegExp(
          r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
        ).hasMatch(repository.reviews.first['requestId'] as String),
        isTrue,
      );
      expect(controller.state.position, 1);
      expect(controller.state.current, isNull);
    },
  );
  test(
    'repository sends schedule version and stable request ID to backend',
    () async {
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test/v1'));
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            expect(options.path, '/anki/decks/wakayama/review');
            expect(options.method, 'POST');
            expect(options.data, {
              'index': 0,
              'version': 3,
              'rating': 'hard',
              'requestId': 'request-id',
            });
            handler.resolve(
              Response<Map<String, dynamic>>(
                requestOptions: options,
                data: {
                  'data': {'saved': true, 'version': 4},
                },
              ),
            );
          },
        ),
      );
      final card = AnkiCard.fromJson({
        'index': 0,
        'version': 3,
        'front': '和歌山市',
        'answers': ['わかやまし'],
      });
      final result = await AnkiRepository(
        dio,
      ).review('wakayama', card, 'hard', 'request-id');
      expect(result['version'], 4);
    },
  );
  testWidgets(
    'phone searches the real catalog, explains Wakayama, then reveals and saves a flashcard',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = StudyRepository();
      final router = GoRouter(
        initialLocation: '/anki',
        routes: [
          GoRoute(
            path: '/anki',
            builder: (_, __) => const AnkiDeckListScreen(),
          ),
          GoRoute(
            path: '/anki/decks/:deckId',
            builder: (_, s) =>
                AnkiDeckDetailScreen(deckId: s.pathParameters['deckId']!),
          ),
          GoRoute(
            path: '/anki/decks/:deckId/study',
            builder: (_, s) =>
                AnkiStudyScreen(deckId: s.pathParameters['deckId']!),
          ),
        ],
      );
      addTearDown(router.dispose);
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
          child: MaterialApp.router(
            routerConfig: router,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.4)),
              child: child!,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      Future<void> show(Finder finder, {double delta = 250}) async {
        await tester.scrollUntilVisible(
          finder,
          delta,
          scrollable: find.byType(Scrollable).first,
          maxScrolls: 100,
        );
        await tester.pumpAndSettle();
      }

      await show(find.byKey(const Key('anki-search')));
      await tester.enterText(find.byKey(const Key('anki-search')), 'wakayama');
      tester.testTextInput.hide();
      await tester.pumpAndSettle();
      await show(find.byKey(const Key('anki-deck-wakayama')));
      await tester.tap(find.byKey(const Key('anki-deck-wakayama')));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('Wakayama adalah salah satu prefektur'),
        findsOneWidget,
      );
      await show(find.byKey(const Key('anki-start')));
      await tester.tap(find.byKey(const Key('anki-start')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('anki-card-back')), findsNothing);
      expect(find.byKey(const Key('anki-rate-good')), findsNothing);
      await show(find.byKey(const Key('anki-reveal')));
      await tester.tap(find.byKey(const Key('anki-reveal')));
      await tester.pumpAndSettle();
      await show(find.byKey(const Key('anki-rate-good')));
      await tester.tap(find.byKey(const Key('anki-rate-good')));
      await tester.pumpAndSettle();
      expect(repository.reviews.length, 1);
      expect(find.text('Sesi selesai!'), findsOneWidget);
      expect(tester.takeException(), isNull);
      router.go('/anki/decks/esyn_easy');
      await tester.pumpAndSettle();
      await show(find.text('Ketersediaan deck'));
      expect(find.byKey(const Key('anki-start')), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );
}
