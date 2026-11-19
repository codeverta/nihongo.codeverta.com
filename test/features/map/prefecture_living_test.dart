import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
import 'package:nihongo_master/features/map/domain/prefecture_living.dart';
import 'package:nihongo_master/features/map/presentation/japan_map_screen.dart';
import 'package:nihongo_master/features/map/presentation/prefecture_living_card.dart';

final _geometry = JapanGeometry.fromJson(
  jsonDecode(File('assets/map/japan_prefectures.json').readAsStringSync())
      as Map<String, dynamic>,
);

/// The backend's bundled data, so the app is tested against real figures.
final _bundled = PrefectureLivingData.fromJson(
  jsonDecode(File('backend/internal/prefinfo/default.json').readAsStringSync())
      as Map<String, dynamic>,
);

Future<void> _pumpMap(
  WidgetTester tester, {
  PrefectureLivingData? living,
}) async {
  tester.view
    ..physicalSize = const Size(390, 780)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        japanGeometryProvider.overrideWith((ref) async => _geometry),
        mapSpotsProvider.overrideWith((ref) async => const MapData(spots: [])),
        mapJobCountsProvider.overrideWith(
          (ref) async => const JobCounts(counts: {'13': 4}),
        ),
        mapPrefectureJobsProvider.overrideWith(
          (ref, prefecture) async =>
              JobPage.fromJson({'items': <Object>[], 'total': 0}),
        ),
        mapPrefectureLivingProvider.overrideWith((ref) async => living),
      ],
      child: MaterialApp(
        theme: AppTheme.lightTheme,
        locale: const Locale('id'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: const JapanMapScreen(),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  test('bundled data derives monthly pay, rupiah and heatmap shares', () {
    final tokyo = _bundled['13']!;
    expect(tokyo.minimumWage, 1280);
    expect(_bundled.monthlyWage(tokyo), 1280 * 8 * 21);
    expect(_bundled.toRupiah(1000), (1000 * _bundled.meta.idrPerJpy).round());
    final (low, high) = _bundled.wageRange;
    expect(low, lessThan(high));
    expect(_bundled.wageShare('13'), 1.0);
    expect(_bundled.prefectures, hasLength(47));
  });

  test('rupiah is shortened for display', () {
    expect(formatRupiahShort(24620000), 'Rp 24,6 jt');
    expect(formatRupiahShort(950000), 'Rp 950 rb');
    expect(formatRupiahShort(24620000, english: true), 'Rp 24.6 M');
    expect(formatYen(1280), '¥1.280');
  });

  testWidgets('the prefecture sheet shows wage and living information', (
    tester,
  ) async {
    await _pumpMap(tester, living: _bundled);
    await tester.tap(find.byKey(const Key('job-pin-13')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('prefecture-living-card')), findsOneWidget);
    expect(find.textContaining('¥1.280'), findsOneWidget);
    expect(find.text('#1 / 47'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('the wage heatmap can be switched on and shows a legend', (
    tester,
  ) async {
    await _pumpMap(tester, living: _bundled);
    expect(find.byKey(const Key('wage-heat-legend')), findsNothing);
    await tester.ensureVisible(find.byKey(const Key('map-layer-wages')));
    await tester.tap(find.byKey(const Key('map-layer-wages')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('wage-heat-legend')), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('without data the map hides wage features', (tester) async {
    await _pumpMap(tester);
    expect(find.byKey(const Key('map-layer-wages')), findsNothing);
    await tester.tap(find.byKey(const Key('job-pin-13')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('prefecture-living-card')), findsNothing);
  });
}
