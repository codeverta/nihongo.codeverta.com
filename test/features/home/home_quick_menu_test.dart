import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_quick_menu.dart';

void main() {
  Future<List<String>> pumpMenu(
    WidgetTester tester, {
    required bool lifetime,
    List<String>? opened,
  }) async {
    final categories = opened ?? <String>[];
    final router = GoRouter(
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => Scaffold(
            body: Padding(
              padding: const EdgeInsets.all(20),
              child: HomeQuickMenu(
                openExamCategory: categories.add,
                hasLifetimeAccess: lifetime,
              ),
            ),
          ),
        ),
        for (final path in [
          '/manga',
          '/anki',
          '/ssw',
          '/jobs',
          '/map',
          '/news',
        ])
          GoRoute(
            path: path,
            builder: (context, state) =>
                Scaffold(body: Text('destination $path')),
          ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.pumpAndSettle();
    return categories;
  }

  testWidgets('shows eight menu items in a 4x2 grid without scrolling', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(360, 760);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await pumpMenu(tester, lifetime: false);
    const ids = ['jft', 'nat', 'manga', 'anki', 'ssw', 'jobs', 'map', 'news'];
    for (final id in ids) {
      expect(find.byKey(Key('home-quick-$id')), findsOneWidget);
    }
    final firstRow = tester.getTopLeft(find.byKey(const Key('home-quick-jft')));
    final lastOfFirstRow =
        tester.getTopLeft(find.byKey(const Key('home-quick-anki')));
    final secondRow = tester.getTopLeft(find.byKey(const Key('home-quick-ssw')));
    expect(lastOfFirstRow.dy, firstRow.dy);
    expect(secondRow.dy, greaterThan(firstRow.dy));
    expect(find.byKey(const Key('home-quick-premium')), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('hides the Premium crown for Lifetime Access', (tester) async {
    await pumpMenu(tester, lifetime: true);
    expect(find.byKey(const Key('home-quick-premium')), findsNothing);
  });

  testWidgets('JFT and NAT open their exam categories', (tester) async {
    final opened = await pumpMenu(tester, lifetime: false);
    await tester.tap(find.byKey(const Key('home-quick-jft')));
    await tester.tap(find.byKey(const Key('home-quick-nat')));
    await tester.pumpAndSettle();
    expect(opened, ['JFT Test', 'NAT Test']);
  });

  for (final entry in {
    'manga': '/manga',
    'anki': '/anki',
    'ssw': '/ssw',
    'jobs': '/jobs',
    'map': '/map',
    'news': '/news',
  }.entries) {
    testWidgets('${entry.key} opens ${entry.value}', (tester) async {
      await pumpMenu(tester, lifetime: false);
      await tester.tap(find.byKey(Key('home-quick-${entry.key}')));
      await tester.pumpAndSettle();
      expect(find.text('destination ${entry.value}'), findsOneWidget);
    });
  }

  test('icons use theme colors and contain no unresolved placeholders', () {
    for (final id in [
      'jft',
      'nat',
      'manga',
      'anki',
      'ssw',
      'jobs',
      'map',
      'news',
    ]) {
      final svg = homeQuickMenuSvg(
        id,
        color: const Color(0xFF112233),
        accent: const Color(0xFFD7263D),
        surface: const Color(0xFFFFFFFF),
      );
      expect(svg, contains('#112233'));
      expect(svg, isNot(contains('{')));
      expect(svg.length, greaterThan(200));
    }
  });
}
