import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/learn/kana/data/kana_repository.dart';
import 'package:nihongo_master/features/learn/kana/domain/kana_models.dart';
import 'package:nihongo_master/features/learn/kana/domain/stroke_matcher.dart';
import 'package:nihongo_master/features/learn/kana/presentation/kana_chart_screen.dart';
import 'package:nihongo_master/features/learn/kana/presentation/kana_detail_screen.dart';
import 'package:nihongo_master/features/learn/kana/presentation/kana_practice_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

final KanaStrokeData strokeData = parseKanaStrokes(
  File(kanaStrokesAsset).readAsStringSync(),
);

void main() {
  group('kana chart data', () {
    test('has the full beginner set in both scripts', () {
      for (final script in KanaScript.values) {
        final chart = KanaChart.of(script);
        expect(chart.list(KanaSection.basic), hasLength(46));
        expect(chart.list(KanaSection.dakuten), hasLength(25));
        expect(chart.writable(KanaSection.combo), hasLength(4));
        expect(
          chart.list(KanaSection.combo).where((k) => k.isCombo),
          hasLength(33),
        );
      }
      expect(KanaChart.of(KanaScript.katakana).find('キャ')?.romaji, 'kya');
      expect(KanaChart.of(KanaScript.hiragana).find('し')?.romaji, 'shi');
    });

    test('every writable kana has KanjiVG strokes', () {
      for (final script in KanaScript.values) {
        final chart = KanaChart.of(script);
        for (final section in KanaSection.values) {
          for (final kana in chart.writable(section)) {
            expect(strokeData[kana.char], isNotEmpty, reason: kana.char);
          }
        }
      }
      expect(strokeData['あ'], hasLength(3));
      expect(strokeData['く'], hasLength(1));
    });

    test('look-alike tips are attached to the right script', () {
      expect(
        KanaChart.of(KanaScript.katakana).find('シ')!.tips.last.en,
        contains('ツ'),
      );
      expect(KanaChart.of(KanaScript.hiragana).find('し')!.tips, hasLength(1));
    });
  });

  group('stroke matcher', () {
    const matcher = StrokeMatcher();
    final a = strokeData['あ']!;

    test('accepts the reference stroke and a slightly shaky copy', () {
      expect(matcher.check(a[0], a, 0), StrokeVerdict.correct);
      final shaky = [
        for (final (i, p) in a[1].indexed)
          p + Offset(i.isEven ? 0.02 : -0.02, 0.015),
      ];
      expect(matcher.check(shaky, a, 1), StrokeVerdict.correct);
    });

    test('explains reversed, out-of-order and wrong strokes', () {
      expect(
        matcher.check(a[0].reversed.toList(), a, 0),
        StrokeVerdict.reversed,
      );
      expect(matcher.check(a[1], a, 0), StrokeVerdict.wrongOrder);
      expect(
        matcher.check(const [Offset(0.9, 0.9), Offset(0.95, 0.95)], a, 0),
        StrokeVerdict.wrongShape,
      );
    });

    test('a tap matches a dot-like stroke', () {
      final ga = strokeData['が']!;
      final dot = ga[3];
      expect(
        matcher.check([dot[dot.length ~/ 2]], ga, 3),
        StrokeVerdict.correct,
      );
    });
  });

  group('kana screens', () {
    late SharedPreferences preferences;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      preferences = await SharedPreferences.getInstance();
    });

    Widget app(
      String initial, {
      double textScale = 1,
      Locale locale = const Locale('id'),
    }) {
      final router = GoRouter(
        initialLocation: initial,
        routes: [
          GoRoute(
            path: AppRoutes.kanaPath,
            builder: (context, state) => const KanaChartScreen(),
          ),
          GoRoute(
            path: '/learn/kana/char/:char',
            builder: (context, state) =>
                KanaDetailScreen(char: state.pathParameters['char']!),
          ),
          GoRoute(
            path: '/learn/kana/char/:char/practice',
            builder: (context, state) =>
                KanaPracticeScreen(char: state.pathParameters['char']!),
          ),
        ],
      );
      return ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          kanaStrokesProvider.overrideWith((ref) async => strokeData),
        ],
        child: MaterialApp.router(
          routerConfig: router,
          locale: locale,
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(
              context,
            ).copyWith(textScaler: TextScaler.linear(textScale)),
            child: child!,
          ),
        ),
      );
    }

    void smallPhone(WidgetTester tester) {
      tester.view.physicalSize = const Size(320, 640);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.reset);
    }

    testWidgets('chart fits a 320 px phone at 1.5x text and opens a kana', (
      tester,
    ) async {
      smallPhone(tester);
      await tester.pumpWidget(app(AppRoutes.kanaPath, textScale: 1.5));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      expect(find.byKey(const Key('kana-cell-あ')), findsOneWidget);

      await tester.tap(find.byKey(const Key('kana-cell-あ')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('kana-detail-char')), findsOneWidget);
      expect(find.byKey(const Key('kana-stroke-player')), findsOneWidget);
      expect(find.text('Goresan 3 / 3'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('katakana tab and combo detail render on a small phone', (
      tester,
    ) async {
      smallPhone(tester);
      await tester.pumpWidget(app(AppRoutes.kanaPath, textScale: 1.5));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Katakana'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('kana-cell-ア')), findsOneWidget);

      await tester.pumpWidget(
        app(AppRoutes.kanaCharPath('きゃ'), textScale: 1.5),
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('kana-part-き')), findsOneWidget);
      expect(find.byKey(const Key('kana-part-ゃ')), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('step mode walks the strokes one at a time', (tester) async {
      await tester.pumpWidget(app(AppRoutes.kanaCharPath('あ')));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Goresan berikutnya'));
      await tester.pump();
      expect(find.text('Goresan 1 / 3'), findsOneWidget);
      await tester.tap(find.byTooltip('Goresan berikutnya'));
      await tester.pump();
      expect(find.text('Goresan 2 / 3'), findsOneWidget);
    });

    Future<void> drawStroke(WidgetTester tester, Stroke stroke) async {
      await tester.ensureVisible(find.byKey(const Key('kana-practice-pad')));
      await tester.pumpAndSettle();
      final pad = tester.getRect(find.byKey(const Key('kana-practice-pad')));
      Offset at(Offset unit) =>
          pad.topLeft + Offset(unit.dx * pad.width, unit.dy * pad.height);
      final gesture = await tester.startGesture(at(stroke.first));
      for (final point in stroke.skip(1)) {
        await gesture.moveTo(at(point));
      }
      await gesture.up();
      await tester.pump();
    }

    testWidgets(
      'writing from memory without help marks the kana as practised',
      (tester) async {
        smallPhone(tester);
        await tester.pumpWidget(
          app(AppRoutes.kanaPracticePath('い'), textScale: 1.5),
        );
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
        expect(find.text('Tahap 1: Jiplak'), findsOneWidget);

        await tester.tap(find.byKey(const Key('kana-stage-memory')));
        await tester.pump();
        final strokes = strokeData['い']!;

        // Wrong order first: the second stroke is rejected with a hint.
        await drawStroke(tester, strokes[1]);
        expect(
          find.text('Urutannya belum tepat. Goresan itu nanti dulu.'),
          findsOneWidget,
        );

        for (final stroke in strokes) {
          await drawStroke(tester, stroke);
        }
        expect(find.byKey(const Key('kana-practice-complete')), findsOneWidget);
        expect(
          find.text('Hebat! Kamu bisa menulis huruf ini dari ingatan.'),
          findsOneWidget,
        );
        expect(preferences.getStringList('kana_practiced_v1'), ['い']);
        expect(find.byKey(const Key('kana-next-char')), findsOneWidget);
        expect(tester.takeException(), isNull);
      },
    );

    testWidgets('help after three misses withholds the check mark', (
      tester,
    ) async {
      await tester.pumpWidget(app(AppRoutes.kanaPracticePath('く')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('kana-stage-memory')));
      await tester.pump();
      for (var i = 0; i < 3; i++) {
        await drawStroke(tester, const [Offset(0.9, 0.1), Offset(0.95, 0.12)]);
      }
      expect(
        find.text('Lihat contoh goresan berwarna, lalu tiru.'),
        findsOneWidget,
      );
      await drawStroke(tester, strokeData['く']!.single);
      expect(
        find.text(
          'Selesai dengan bantuan. Ulangi tanpa bantuan untuk mendapat centang.',
        ),
        findsOneWidget,
      );
      expect(preferences.getStringList('kana_practiced_v1'), isNull);
    });

    testWidgets('tracing stage leads to the next stage', (tester) async {
      await tester.pumpWidget(app(AppRoutes.kanaPracticePath('く')));
      await tester.pumpAndSettle();
      await drawStroke(tester, strokeData['く']!.single);
      expect(find.text('Tahap selesai!'), findsOneWidget);
      await tester.ensureVisible(find.byKey(const Key('kana-next-stage')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('kana-next-stage')));
      await tester.pump();
      expect(find.text('Tahap 2: Dengan petunjuk'), findsOneWidget);
    });

    testWidgets('English UI shows English copy and tips', (tester) async {
      await tester.pumpWidget(
        app(AppRoutes.kanaCharPath('し'), locale: const Locale('en')),
      );
      await tester.pumpAndSettle();
      expect(find.text('Writing practice'), findsOneWidget);
      expect(find.text('Said "shi", not "si".'), findsOneWidget);
    });
  });
}
