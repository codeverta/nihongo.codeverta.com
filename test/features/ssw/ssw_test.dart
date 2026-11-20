import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  late KaigoCurriculum forestry;
  setUpAll(() async {
    forestry = KaigoCurriculum.fromJson(
      jsonDecode(
            await rootBundle.loadString('assets/content/ssw/forestry.json'),
          )
          as Map<String, dynamic>,
    );
  });

  for (final exams in [false, true]) {
    testWidgets(
      'SSW phone search opens field and preserves entry tab: $exams',
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
        SharedPreferences.setMockInitialValues({});
        final preferences = await SharedPreferences.getInstance();
        await tester.pumpWidget(
          ProviderScope(
            overrides: [
              // The hub orders fields by recent access, kept in preferences.
              sharedPreferencesProvider.overrideWithValue(preferences),
              sswCurriculumProvider(
                'forestry',
              ).overrideWith((ref) async => forestry),
              authControllerProvider.overrideWith(
                (ref) => AuthController.test(
                  const AuthenticationState.authenticated(
                    AuthenticatedUser(
                      id: 'student',
                      displayName: 'Student',
                      email: 's@example.test',
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
        // Every SSW field now has material and tests (no Coming Soon preview).
        final material = find.text('Kurikulum Kehutanan');
        final test = find.text('Simulasi · ${sswExamTitle('SSW-FORESTRY')}');
        await tester.scrollUntilVisible(
          exams ? test : material,
          100,
          scrollable: find.byType(Scrollable).last,
        );
        expect(exams ? test : material, findsOneWidget);
        expect(exams ? material : test, findsNothing);
        expect(find.text('Coming Soon'), findsNothing);
        await tester.scrollUntilVisible(
          find.text(exams ? 'Materi' : 'Tes'),
          -100,
          scrollable: find.byType(Scrollable).last,
        );
        await tester.tap(find.text(exams ? 'Materi' : 'Tes'));
        await tester.pumpAndSettle();
        await tester.scrollUntilVisible(
          exams ? material : test,
          100,
          scrollable: find.byType(Scrollable).last,
        );
        expect(exams ? material : test, findsOneWidget);
        expect(tester.takeException(), isNull);
      },
    );
  }
}
