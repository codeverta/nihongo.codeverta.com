import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
import 'package:nihongo_master/features/map/presentation/japan_map_screen.dart';
import 'package:nihongo_master/features/map/presentation/spot_detail_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

final geometry = JapanGeometry.fromJson(
  jsonDecode(File('assets/map/japan_prefectures.json').readAsStringSync())
      as Map<String, dynamic>,
);

Map<String, dynamic> spotJson({
  String id = 'fushimi',
  String summaryId = '',
  String meaningEn = '',
}) => {
  'id': id,
  'category': 'shrine',
  'prefecture': '26',
  'lat': 34.9671,
  'lng': 135.7727,
  'name': {'id': 'Fushimi Inari-taisha', 'en': 'Fushimi Inari-taisha'},
  'nameJa': '伏見稲荷大社',
  'reading': 'ふしみいなりたいしゃ',
  'summary': {
    'id': summaryId.isEmpty
        ? {
            'text': 'Head shrine of Inari.',
            'sourceLang': 'en',
            'manual': false,
          }
        : {'text': summaryId, 'sourceLang': 'id', 'manual': true},
    'en': {
      'text': 'Head shrine of Inari.',
      'sourceLang': 'en',
      'manual': false,
    },
  },
  'tips': {'id': {'text': ''}, 'en': {'text': ''}},
  'phrases': [
    {
      'japanese': '鳥居',
      'reading': 'とりい',
      'meaningId': 'gerbang kuil',
      'meaningEn': meaningEn,
    },
  ],
  'image': null,
  'sources': [
    {
      'kind': 'wikipedia',
      'lang': 'en',
      'url': 'https://en.wikipedia.org/wiki/Fushimi_Inari-taisha',
      'license': 'CC BY-SA 4.0',
    },
  ],
};

class _Adapter implements HttpClientAdapter {
  _Adapter(this.handler);
  final ResponseBody Function(RequestOptions options) handler;
  final requests = <RequestOptions>[];
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}

Widget app(
  Widget child, {
  Locale locale = const Locale('id'),
  List<Override> overrides = const [],
}) => ProviderScope(
      overrides: overrides,
      child: MaterialApp(
        locale: locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: child,
      ),
    );

void main() {
  group('models', () {
    test('spot text falls back to English and hides untranslated phrases', () {
      final spot = MapSpot.fromJson(spotJson());
      expect(spot.summaryFor('id').sourceLang, 'en');
      expect(spot.phrasesFor('id'), hasLength(1));
      expect(spot.phrasesFor('en'), isEmpty);
      final manual = MapSpot.fromJson(
        spotJson(summaryId: 'Kuil utama Inari.', meaningEn: 'shrine gate'),
      );
      expect(manual.summaryFor('id').text, 'Kuil utama Inari.');
      expect(manual.phrasesFor('en').single.meaningFor('en'), 'shrine gate');
    });

    test('geometry projects capitals into their own prefecture', () {
      const size = Size(1000, 1000);
      final paths = geometry.paths(size);
      expect(geometry.prefectures, hasLength(47));
      for (final prefecture in geometry.prefectures) {
        final point = geometry.toCanvas(
          prefecture.code,
          prefecture.capitalLat,
          prefecture.capitalLng,
          size,
        );
        expect(
          paths[prefecture.code]!.contains(point),
          isTrue,
          reason: '${prefecture.name} capital should be inside its outline',
        );
      }
      final naha = geometry.toCanvas('47', 26.212, 127.681, size);
      expect(geometry.insetFrame(size).contains(naha), isTrue);
    });
  });

  group('repository', () {
    late SharedPreferences preferences;
    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      preferences = await SharedPreferences.getInstance();
    });

    test('revalidates spots with ETag and serves the cache offline', () async {
      var offline = false;
      final body = jsonEncode({
        'data': {
          'spots': [spotJson()],
          'categories': mapCategories,
        },
      });
      final adapter = _Adapter((options) {
        if (offline) {
          throw DioException.connectionError(
            requestOptions: options,
            reason: 'offline',
          );
        }
        if (options.headers['If-None-Match'] == '"v1"') {
          return ResponseBody.fromString('', 304);
        }
        return ResponseBody.fromString(
          body,
          200,
          headers: {
            'etag': ['"v1"'],
            'content-type': ['application/json'],
          },
        );
      });
      final dio = Dio(BaseOptions(baseUrl: 'https://api.test/v1'))
        ..httpClientAdapter = adapter;
      final repository = MapRepository(dio, preferences);

      final first = await repository.spots();
      expect(first.spots.single.id, 'fushimi');
      expect(first.fromCache, isFalse);

      final second = await repository.spots();
      expect(second.spots.single.id, 'fushimi');
      expect(adapter.requests.last.headers['If-None-Match'], '"v1"');

      offline = true;
      final third = await repository.spots();
      expect(third.fromCache, isTrue);
      expect(third.spots, hasLength(1));
    });

    test('job counts fall back to the last saved counts', () async {
      var offline = false;
      final adapter = _Adapter((options) {
        if (offline) {
          throw DioException.connectionError(
            requestOptions: options,
            reason: 'offline',
          );
        }
        return ResponseBody.fromString(
          jsonEncode({
            'data': {
              'counts': {'13': 4, '27': 0},
            },
          }),
          200,
          headers: {
            'content-type': ['application/json'],
          },
        );
      });
      final dio = Dio(BaseOptions(baseUrl: 'https://api.test/v1'))
        ..httpClientAdapter = adapter;
      final repository = MapRepository(dio, preferences);
      expect((await repository.jobCounts())['13'], 4);
      offline = true;
      final cached = await repository.jobCounts();
      expect(cached.fromCache, isTrue);
      expect(cached['13'], 4);
      expect(cached['27'], 0);
    });
  });

  group('screens', () {
    List<Override> overrides({Map<String, int> counts = const {'26': 3}}) => [
      japanGeometryProvider.overrideWith((ref) async => geometry),
      mapSpotsProvider.overrideWith(
        (ref) async => MapData(spots: [MapSpot.fromJson(spotJson())]),
      ),
      mapJobCountsProvider.overrideWith(
        (ref) async => JobCounts(counts: counts),
      ),
      mapPrefectureJobsProvider.overrideWith(
        (ref, prefecture) async => JobPage.fromJson({'items': [], 'total': 0}),
      ),
    ];

    testWidgets('spot detail labels English fallback text', (tester) async {
      await tester.pumpWidget(
        app(const SpotDetailScreen(id: 'fushimi'), overrides: overrides()),
      );
      await tester.pumpAndSettle();
      expect(find.text('伏見稲荷大社'), findsOneWidget);
      expect(find.byKey(const Key('english-fallback-note')), findsOneWidget);
      expect(find.text('鳥居'), findsOneWidget);
      expect(find.byKey(const Key('open-google-maps')), findsOneWidget);
      expect(find.byKey(const Key('spot-prefecture-jobs')), findsOneWidget);
      expect(
        googleMapsUri(MapSpot.fromJson(spotJson())).toString(),
        'https://www.google.com/maps/search/?api=1&query=34.9671%2C135.7727',
      );
    });

    testWidgets('English mode hides phrases without English meaning', (
      tester,
    ) async {
      await tester.pumpWidget(
        app(
          const SpotDetailScreen(id: 'fushimi'),
          locale: const Locale('en'),
          overrides: overrides(counts: const {}),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('鳥居'), findsNothing);
      expect(find.byKey(const Key('english-fallback-note')), findsNothing);
      expect(find.byKey(const Key('spot-prefecture-jobs')), findsNothing);
    });

    testWidgets('job pin opens the prefecture sheet', (tester) async {
      await tester.pumpWidget(
        app(const JapanMapScreen(), overrides: overrides()),
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('japan-map')), findsOneWidget);
      await tester.tap(find.byKey(const Key('job-pin-26')));
      await tester.pumpAndSettle();
      expect(find.text('京都府'), findsOneWidget);
      expect(find.text('Fushimi Inari-taisha'), findsWidgets);
      expect(find.byKey(const Key('map-all-jobs')), findsOneWidget);
    });

    testWidgets('job layer can be hidden', (tester) async {
      await tester.pumpWidget(
        app(const JapanMapScreen(), overrides: overrides()),
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('job-pin-26')), findsOneWidget);
      await tester.tap(find.byKey(const Key('map-layer-jobs')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('job-pin-26')), findsNothing);
    });
  });
}
