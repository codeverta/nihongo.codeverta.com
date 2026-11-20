import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/anki/presentation/course_anki_button.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

Future<void> _pump(WidgetTester tester, List<Widget> actions) async {
  final router = GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => Scaffold(
          appBar: AppBar(title: const Text('Course'), actions: actions),
          body: const SizedBox.shrink(),
        ),
      ),
      GoRoute(
        path: '/anki/decks/:deckId',
        builder: (context, state) =>
            Scaffold(body: Text('deck ${state.pathParameters['deckId']}')),
      ),
    ],
  );
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
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
        theme: AppTheme.lightTheme,
        routerConfig: router,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  test('courses map to the matching built-in Anki decks', () {
    expect(courseAnkiDecks('kanji', 'N5').map((d) => d.id), ['k_n5']);
    expect(courseAnkiDecks('kanji', 'N1').map((d) => d.id), ['k_n1']);
    expect(courseAnkiDecks('vocabulary', 'N5').map((d) => d.id), ['n5', 'n5m']);
    expect(courseAnkiDecks('vocabulary', 'N3').map((d) => d.id), ['n3']);
    expect(courseAnkiDecks('grammar', 'N5'), isEmpty);
    expect(courseAnkiDecks('reading', 'N5'), isEmpty);
    expect(courseAnkiDecks('kanji', 'JFT'), isEmpty);
  });

  testWidgets('kanji opens its single deck directly', (tester) async {
    await _pump(tester, const [CourseAnkiButton(type: 'kanji', level: 'N5')]);
    await tester.tap(find.byKey(const ValueKey('course-anki-button')));
    await tester.pumpAndSettle();
    expect(find.text('deck k_n5'), findsOneWidget);
  });

  testWidgets('vocabulary lets the learner choose between two decks', (
    tester,
  ) async {
    await _pump(tester, const [
      CourseAnkiButton(type: 'vocabulary', level: 'N4'),
    ]);
    await tester.tap(find.byKey(const ValueKey('course-anki-button')));
    await tester.pumpAndSettle();
    expect(find.text('N4 · Bacaan Kosakata'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('course-anki-deck-n4m')));
    await tester.pumpAndSettle();
    expect(find.text('deck n4m'), findsOneWidget);
  });

  testWidgets('grammar has no Anki action', (tester) async {
    await _pump(tester, const [CourseAnkiButton(type: 'grammar', level: 'N5')]);
    expect(find.byKey(const ValueKey('course-anki-button')), findsNothing);
  });
}
