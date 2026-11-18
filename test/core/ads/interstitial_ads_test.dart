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

  test('never covers startup, sign-in or the purchase flow', () {
    expect(interstitialAllowedOn('/'), isFalse);
    expect(interstitialAllowedOn('/onboarding'), isFalse);
    expect(interstitialAllowedOn('/login'), isFalse);
    expect(interstitialAllowedOn('/premium'), isFalse);
    expect(interstitialAllowedOn('/premium?from=/jobs/1'), isFalse);
    expect(interstitialAllowedOn('/learn/grammar/lessons/:lessonId'), isTrue);
    expect(interstitialAllowedOn('/tab/dictionary'), isTrue);
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
      'shows at the first natural break after 10 minutes, then waits again',
      () async {
        final (_, controller) = await setUp(_config);
        now = now.add(const Duration(minutes: 5));
        await controller.onNaturalBreak('/learn');
        expect(gateway.loads, 0, reason: 'too early to even load');

        now = now.add(const Duration(minutes: 4, seconds: 30));
        await controller.onNaturalBreak('/learn');
        expect(gateway.loads, 1, reason: 'preloads a minute early');
        expect(gateway.shows, 0);

        now = now.add(const Duration(minutes: 1));
        await controller.onNaturalBreak('/tab/dictionary');
        expect(gateway.shows, 1);

        // The count restarts after the ad.
        now = now.add(const Duration(minutes: 3));
        await controller.onNaturalBreak('/learn');
        expect(gateway.shows, 1);
      },
    );

    test('waits for an allowed screen even when due', () async {
      final (_, controller) = await setUp(_config);
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 11));
      await controller.onNaturalBreak('/premium');
      await controller.onNaturalBreak('/login');
      expect(gateway.shows, 0);
      await controller.onNaturalBreak('/home');
      expect(gateway.shows, 1);
    });

    test('a rewarded ad restarts the count', () async {
      final (_, controller) = await setUp(_config);
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 11));
      controller.markFullScreenAdShown();
      await controller.onNaturalBreak('/home');
      expect(gateway.shows, 0);
    });

    test('Premium and Lifetime users (no ads config) never see one', () async {
      final (_, controller) = await setUp(null);
      gateway.loaded = true;
      now = now.add(const Duration(hours: 2));
      await controller.onNaturalBreak('/home');
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
      await controller.onNaturalBreak('/home');
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
      await controller.onNaturalBreak('/home');
      expect(gateway.shows, 0);
      now = now.add(const Duration(minutes: 5));
      await controller.onNaturalBreak('/home');
      expect(gateway.shows, 1);
    });

    test('background time is excluded via the app lifecycle', () async {
      final (_, controller) = await setUp(_config);
      gateway.loaded = true;
      now = now.add(const Duration(minutes: 5));
      controller.pacer.pause(now);
      now = now.add(const Duration(hours: 1));
      controller.pacer.resume(now);
      await controller.onNaturalBreak('/home');
      expect(gateway.shows, 0);
      expect(WidgetsBinding.instance, isNotNull);
    });
  });

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
