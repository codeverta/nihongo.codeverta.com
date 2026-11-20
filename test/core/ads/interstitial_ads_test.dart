import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';

class _FakeGateway implements InterstitialAdGateway {
  bool loaded = false;
  int loads = 0;
  int shows = 0;

  @override
  bool get ready => loaded;

  @override
  Future<void> load(String unitId) async {
    loads++;
    loaded = true;
  }

  @override
  Future<bool> show() async {
    if (!loaded) return false;
    loaded = false;
    shows++;
    return true;
  }

  @override
  void dispose() {}
}

const _config = AdsConfig(
  enabled: true,
  interstitial: AdPlacementConfig(
    enabled: true,
    unitId: testInterstitialAdUnitId,
    every: 10,
  ),
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('InterstitialPacer', () {
    final t0 = DateTime(2026, 11, 18, 10);

    test('only foreground time counts toward the interval', () {
      final pacer = InterstitialPacer()..resume(t0);
      pacer.pause(t0.add(const Duration(minutes: 6)));
      // An hour in the background adds nothing.
      pacer.resume(t0.add(const Duration(minutes: 66)));
      expect(pacer.due(t0.add(const Duration(minutes: 69))), isFalse);
      expect(pacer.nearlyDue(t0.add(const Duration(minutes: 69))), isTrue);
      expect(pacer.due(t0.add(const Duration(minutes: 70))), isTrue);
    });

    test('reset restarts the count from now', () {
      final pacer = InterstitialPacer()..resume(t0);
      pacer.reset(t0.add(const Duration(minutes: 12)));
      expect(
        pacer.elapsed(t0.add(const Duration(minutes: 15))),
        const Duration(minutes: 3),
      );
    });
  });

  group('InterstitialController', () {
    late DateTime now;
    late _FakeGateway gateway;

    Future<(ProviderContainer, InterstitialController)> setUp(
      AdsConfig? config,
    ) async {
      now = DateTime(2026, 11, 18, 10);
      gateway = _FakeGateway();
      final provider = Provider<InterstitialController>((ref) {
        final controller = InterstitialController(ref, clock: () => now);
        ref.onDispose(controller.dispose);
        return controller;
      });
      final container = ProviderContainer(
        overrides: [
          adsReadyProvider.overrideWith((ref) async => config),
          interstitialAdGatewayProvider.overrideWithValue(gateway),
        ],
      );
      addTearDown(container.dispose);
      await container.read(adsReadyProvider.future);
      return (container, container.read(provider));
    }

    test(
      'loads on the result screen and shows when the user leaves it',
      () async {
        final (_, controller) = await setUp(_config);
        now = now.add(const Duration(minutes: 5));
        controller.prepareAfterCompletion(CompletionKind.lessons);
        expect(gateway.loads, 0, reason: 'too early to even load');

        now = now.add(const Duration(minutes: 4, seconds: 30));
        controller.prepareAfterCompletion(CompletionKind.lessons);
        expect(gateway.loads, 1, reason: 'preloads a minute early');

        now = now.add(const Duration(minutes: 1));
        await controller.showAfterCompletion(CompletionKind.lessons);
        expect(gateway.shows, 1);

        // The count restarts after the ad.
        now = now.add(const Duration(minutes: 3));
        gateway.loaded = true;
        await controller.showAfterCompletion(CompletionKind.exams);
        expect(gateway.shows, 1);
      },
    );

    test(
      'an ad that did not load in time is loaded for the next completion',
      () async {
        final (_, controller) = await setUp(_config);
        now = now.add(const Duration(minutes: 11));
        await controller.showAfterCompletion(CompletionKind.anki);
        expect(gateway.shows, 0, reason: 'never makes the user wait');
        expect(gateway.loads, 1);
        await controller.showAfterCompletion(CompletionKind.anki);
        expect(gateway.shows, 1);
      },
    );

    test(
      'a completion paid for with a rewarded ad gets no full-screen ad',
      () async {
        final (_, controller) = await setUp(_config);
        controller.markPaidWithAd('grammar:grammar-n3-01');
        gateway.loaded = true;
        now = now.add(const Duration(minutes: 30));
        await controller.showAfterCompletion(
          CompletionKind.lessons,
          key: 'grammar:grammar-n3-01',
        );
        expect(gateway.shows, 0);
        // Only that one completion is skipped.
        await controller.showAfterCompletion(
          CompletionKind.lessons,
          key: 'grammar:grammar-n3-01',
        );
        expect(gateway.shows, 1);
      },
    );

    test('a rewarded ad restarts the count', () async {
      final (_, controller) = await setUp(_config);
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 11));
      controller.markFullScreenAdShown();
      await controller.showAfterCompletion(CompletionKind.lessons);
      expect(gateway.shows, 0);
    });

    test('each completion kind has its own switch', () async {
      final (_, controller) = await setUp(
        const AdsConfig(
          enabled: true,
          interstitial: AdPlacementConfig(
            enabled: true,
            unitId: testInterstitialAdUnitId,
            every: 10,
          ),
          interstitialKinds: CompletionKindsConfig(anki: false),
        ),
      );
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 30));
      await controller.showAfterCompletion(CompletionKind.anki);
      expect(gateway.shows, 0);
      await controller.showAfterCompletion(CompletionKind.practice);
      expect(gateway.shows, 1);
    });

    test('Premium and Lifetime users (no ads config) never see one', () async {
      final (_, controller) = await setUp(null);
      gateway.loaded = true;
      now = now.add(const Duration(hours: 2));
      controller.prepareAfterCompletion(CompletionKind.lessons);
      await controller.showAfterCompletion(CompletionKind.lessons);
      expect(gateway.shows, 0);
      expect(gateway.loads, 0);
    });

    test('a switched-off placement shows nothing', () async {
      final (_, controller) = await setUp(
        const AdsConfig(
          enabled: true,
          interstitial: AdPlacementConfig(
            unitId: testInterstitialAdUnitId,
            every: 10,
          ),
        ),
      );
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 30));
      await controller.showAfterCompletion(CompletionKind.lessons);
      expect(gateway.shows, 0);
    });

    test('the server interval is respected', () async {
      final (_, controller) = await setUp(
        const AdsConfig(
          enabled: true,
          interstitial: AdPlacementConfig(
            enabled: true,
            unitId: testInterstitialAdUnitId,
            every: 20,
          ),
        ),
      );
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 15));
      await controller.showAfterCompletion(CompletionKind.exams);
      expect(gateway.shows, 0);
      now = now.add(const Duration(minutes: 5));
      await controller.showAfterCompletion(CompletionKind.exams);
      expect(gateway.shows, 1);
    });

    test('background time is excluded via the app lifecycle', () async {
      final (_, controller) = await setUp(_config);
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 5));
      controller.pacer.pause(now);
      now = now.add(const Duration(hours: 1));
      controller.pacer.resume(now);
      await controller.showAfterCompletion(CompletionKind.lessons);
      expect(gateway.shows, 0);
      expect(WidgetsBinding.instance, isNotNull);
    });
  });

  test(
    'a result screen torn down with the app never shows an ad afterwards',
    () async {
      final gateway = _FakeGateway()..loaded = true;
      final container = ProviderContainer(
        overrides: [
          adsReadyProvider.overrideWith((ref) async => _config),
          interstitialAdGatewayProvider.overrideWithValue(gateway),
        ],
      );
      await container.read(adsReadyProvider.future);
      final controller = container.read(interstitialControllerProvider);
      container.dispose();
      await controller.showAfterCompletion(CompletionKind.lessons);
      controller.prepareAfterCompletion(CompletionKind.lessons);
      expect(gateway.shows, 0);
    },
  );

  test(
    'completion switches, new placements and the Anki threshold are parsed',
    () {
      final config = AdsConfig.fromJson(const {
        'enabled': true,
        'interstitialKinds': {'lessons': true, 'exams': false},
        'examTabBanner': {
          'enabled': true,
          'unitId': 'ca-app-pub-1234567890123456/1234567893',
        },
        'practiceResultBanner': {
          'enabled': true,
          'unitId': 'ca-app-pub-1234567890123456/1234567894',
        },
        'newsArticleNative': {
          'enabled': true,
          'unitId': 'ca-app-pub-1234567890123456/1234567895',
        },
        'ankiSessionMinCards': 15,
      });
      expect(config.interstitialKinds.allows(CompletionKind.lessons), isTrue);
      expect(config.interstitialKinds.allows(CompletionKind.exams), isFalse);
      expect(
        config.interstitialKinds.allows(CompletionKind.anki),
        isTrue,
        reason: 'missing means on',
      );
      expect(config.examTabBanner.usable, isTrue);
      expect(config.ankiSessionMinCards, 15);
      final debug = config.withTestUnits();
      expect(debug.examTabBanner.unitId, testBannerAdUnitId);
      expect(debug.practiceResultBanner.unitId, testBannerAdUnitId);
      expect(debug.newsArticleNative.unitId, testNativeAdUnitId);
      expect(
        AdsConfig.fromJson(const {
          'ankiSessionMinCards': 0,
        }).ankiSessionMinCards,
        defaultAnkiSessionMinCards,
      );
    },
  );

  test('new placements are parsed and swapped for test units', () {
    final config = AdsConfig.fromJson(const {
      'enabled': true,
      'interstitial': {
        'enabled': true,
        'unitId': 'ca-app-pub-1234567890123456/1234567890',
        'every': 10,
      },
      'dictionaryNative': {
        'enabled': true,
        'unitId': 'ca-app-pub-1234567890123456/1234567891',
        'every': 8,
      },
      'quizResultBanner': {
        'enabled': true,
        'unitId': 'ca-app-pub-1234567890123456/1234567892',
      },
    });
    expect(config.interstitial.every, 10);
    expect(config.dictionaryNative.every, 8);
    expect(config.quizResultBanner.usable, isTrue);
    final debug = config.withTestUnits();
    expect(debug.interstitial.unitId, testInterstitialAdUnitId);
    expect(debug.dictionaryNative.unitId, testNativeAdUnitId);
    expect(debug.quizResultBanner.unitId, testBannerAdUnitId);
  });
}
