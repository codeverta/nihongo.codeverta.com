import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/prefecture_living.dart';
import 'package:nihongo_master/features/map/domain/prefecture_profile.dart';
import 'package:nihongo_master/features/map/presentation/prefecture_profile_screen.dart';

/// The backend's bundled data, so the app is tested against real profiles.
final _profiles = PrefectureProfiles.fromJson(
  jsonDecode(File('backend/internal/prefinfo/profiles.json').readAsStringSync())
      as Map<String, dynamic>,
);

Future<void> _pump(
  WidgetTester tester,
  String code, {
  Locale locale = const Locale('id'),
  PrefectureProfiles? profiles,
}) async {
  tester.view
    ..physicalSize = const Size(390, 844)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  final router = GoRouter(
    initialLocation: prefectureProfilePath(code),
    routes: [
      GoRoute(
        path: '/map/prefecture/:code',
        builder: (context, state) =>
            PrefectureProfileScreen(code: state.pathParameters['code']!),
      ),
    ],
  );
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        mapPrefectureProfilesProvider.overrideWith(
          (ref) async => profiles ?? _profiles,
        ),
        mapPrefectureLivingProvider.overrideWith(
          (ref) async => null as PrefectureLivingData?,
        ),
      ],
      child: MaterialApp.router(
        theme: AppTheme.lightTheme,
        locale: locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        routerConfig: router,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  test('bundled profiles cover every prefecture with text and facts', () {
    expect(_profiles.byCode, hasLength(47));
    for (final profile in _profiles.byCode.values) {
      expect(profile.textFor('id'), isNotNull, reason: profile.code);
      expect(profile.nameJa, isNotEmpty, reason: profile.code);
    }
    final tokyo = _profiles['13']!;
    expect(tokyo.nameJa, '東京都');
    expect(tokyo.density, greaterThan(1000));
    expect(regionOf('13'), 'Kanto');
    expect(regionOf('47'), 'Kyushu & Okinawa');
  });

  testWidgets('the profile page reads like a Wikipedia article', (
    tester,
  ) async {
    await _pump(tester, '13');
    expect(find.byKey(const Key('prefecture-profile-name-ja')), findsOneWidget);
    expect(find.byKey(const Key('prefecture-profile-infobox')), findsOneWidget);
    expect(find.text('Kanto'), findsOneWidget);
    expect(
      find.byKey(const Key('prefecture-profile-english-note')),
      findsNothing,
    );
    await tester.scrollUntilVisible(
      find.byKey(const Key('prefecture-profile-attribution')),
      300,
    );
    expect(
      find.byKey(const Key('prefecture-profile-wikipedia')),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('Hokkaido falls back to English text and says so', (
    tester,
  ) async {
    await _pump(tester, '01');
    expect(
      find.byKey(const Key('prefecture-profile-english-note')),
      findsOneWidget,
    );
  });

  testWidgets('a missing profile offers a retry instead of crashing', (
    tester,
  ) async {
    await _pump(tester, '13', profiles: const PrefectureProfiles(byCode: {}));
    expect(find.text('Profil prefektur belum dapat dimuat.'), findsOneWidget);
  });
}
