import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';

void main() {
  for (final exams in [false, true]) {
    testWidgets('SSW phone search opens field and preserves entry tab: $exams',
        (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final router = GoRouter(
        initialLocation: exams ? '/ssw?tab=test' : '/ssw',
        routes: [
          GoRoute(
            path: '/ssw',
            builder: (_, state) => SswHubScreen(
              exams: state.uri.queryParameters['tab'] == 'test',
            ),
          ),
          GoRoute(
            path: '/ssw/:sectorId',
            builder: (_, state) => SswSectorScreen(
              sectorId: state.pathParameters['sectorId']!,
              exams: state.uri.queryParameters['tab'] == 'test',
            ),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: router,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: const TextScaler.linear(1.4),
            ),
            child: child!,
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.text('19 bidang'),
        100,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('19 bidang'), findsOneWidget);
      await tester.enterText(find.byType(TextField), 'kehutanan');
      tester.testTextInput.hide();
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.text('Kehutanan'),
        100,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(find.text('Kehutanan'));
      await tester.pumpAndSettle();
      final expected = exams ? 'Latihan soal' : 'Materi bidang';
      await tester.scrollUntilVisible(
        find.text(expected),
        100,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text(expected), findsOneWidget);
      expect(find.text('Coming Soon'), findsWidgets);
      expect(find.byType(FilledButton), findsNothing);
      await tester.scrollUntilVisible(
        find.text(exams ? 'Materi' : 'Tes'),
        -100,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(find.text(exams ? 'Materi' : 'Tes'));
      await tester.pumpAndSettle();
      expect(
        find.text(exams ? 'Materi bidang' : 'Latihan soal'),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
    });
  }
}
