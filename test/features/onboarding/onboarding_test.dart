import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/onboarding/presentation/screens/onboarding_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets(
    'completed users can explicitly replay the real onboarding route',
    (tester) async {
      SharedPreferences.setMockInitialValues({'is_first_time': false});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(prefs),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
        ],
      );
      addTearDown(container.dispose);
      final router = container.read(appRouterProvider)
        ..go('/onboarding?preview=1');
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Kuasai Bahasa Jepang'), findsOneWidget);
      expect(find.byKey(const Key('onboarding-close')), findsOneWidget);
      expect(container.read(isFirstTimeProvider), false);
      expect(prefs.getBool('is_first_time'), false);
    },
  );

  testWidgets(
    'preview closes back to its origin without completing onboarding',
    (tester) async {
      var completions = 0;
      final router = GoRouter(
        initialLocation: '/origin',
        routes: [
          GoRoute(
            path: '/origin',
            builder: (_, _) => const Scaffold(body: Text('Origin')),
          ),
          GoRoute(
            path: '/onboarding',
            builder: (_, _) => OnboardingScreen(
              isPreview: true,
              onCompleted: () async {
                completions++;
              },
            ),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(MaterialApp.router(routerConfig: router));
      unawaited(router.push<void>('/onboarding'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('onboarding-page-2')));
      await tester.pumpAndSettle();
      expect(find.text('Kembali ke aplikasi'), findsOneWidget);
      await tester.tap(find.byKey(const Key('onboarding-next')));
      await tester.pumpAndSettle();
      expect(find.text('Origin'), findsOneWidget);
      expect(completions, 0);
    },
  );

  for (final scale in [1.0, 2.0]) {
    testWidgets('all three pages work at 320px with text scale $scale', (
      tester,
    ) async {
      tester.view.physicalSize = const Size(320, 568);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      var completions = 0;
      final router = GoRouter(
        initialLocation: '/onboarding',
        routes: [
          GoRoute(
            path: '/onboarding',
            builder: (_, _) => OnboardingScreen(
              onCompleted: () async {
                completions++;
              },
            ),
          ),
          GoRoute(
            path: '/home',
            builder: (_, _) => const Scaffold(body: Text('Home')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: router,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(
              context,
            ).copyWith(textScaler: TextScaler.linear(scale)),
            child: child!,
          ),
        ),
      );
      await tester.pumpAndSettle();
      for (var page = 0; page < 3; page++) {
        expect(find.byKey(Key('onboarding-art-$page')), findsOneWidget);
        expect(tester.takeException(), isNull);
        await tester.tap(find.byKey(const Key('onboarding-next')));
        await tester.pumpAndSettle();
      }
      expect(find.text('Home'), findsOneWidget);
      expect(completions, 1);
    });
  }

  testWidgets('onboarding art renders and exports a phone preview', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(390, 844);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final boundaryKey = GlobalKey();
    final fontDirectory = Platform.environment['ONBOARDING_PREVIEW_FONTS'];
    if (fontDirectory != null) {
      await tester.runAsync(() async {
        for (final font in {
          'Roboto': 'roboto-regular.ttf',
          'MaterialIcons': 'materialicons-regular.otf',
        }.entries) {
          final loader = FontLoader(font.key)
            ..addFont(
              File(
                '$fontDirectory/${font.value}',
              ).readAsBytes().then(ByteData.sublistView),
            );
          await loader.load();
        }
      });
    }
    await tester.pumpWidget(
      MaterialApp(
        theme: ThemeData(fontFamily: 'Roboto', useMaterial3: true),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: RepaintBoundary(
          key: boundaryKey,
          child: const OnboardingScreen(isPreview: true),
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.runAsync(() async {
      await Future<void>.delayed(const Duration(milliseconds: 200));
    });
    await tester.pumpAndSettle();
    final boundary =
        boundaryKey.currentContext!.findRenderObject()!
            as RenderRepaintBoundary;
    await tester.runAsync(() async {
      final image = await boundary.toImage();
      final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
      await File(
        '.dart_tool/onboarding-preview.png',
      ).writeAsBytes(bytes!.buffer.asUint8List());
      image.dispose();
    });
    expect(tester.takeException(), isNull);
  });
}
