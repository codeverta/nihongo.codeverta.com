import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/gamification/double_xp_button.dart';
import 'package:nihongo_master/core/gamification/xp_gain.dart';
import 'package:nihongo_master/core/unlock/unlock_flow.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

const _gain = XpGain(
  awarded: 50,
  total: 50,
  streak: 1,
  completionId: 'grammar:attempt-1',
  kind: 'lessons',
);

class _Unlocks extends UnlockRepository {
  _Unlocks() : super(Dio());
  final calls = <String>[];

  @override
  Future<UnlockIntent> createIntent(UnlockTarget target) async {
    calls.add('intent:${target.key}');
    return const UnlockIntent(alreadyUnlocked: false, nonce: 'n', userId: 'u');
  }

  @override
  Future<bool> claim(String nonce) async {
    calls.add('claim');
    return true;
  }

  @override
  Future<void> fallback(UnlockTarget target) async => calls.add('fallback');

  @override
  Future<XpGain?> claimFreeBoost(String completionId) async {
    calls.add('free:$completionId');
    return const XpGain(awarded: 50, total: 100, streak: 1);
  }
}

class _Shower implements RewardedAdShower {
  int shown = 0;
  @override
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  }) async {
    shown++;
    return RewardedAdOutcome.watched;
  }
}

class _Interstitials extends InterstitialController {
  _Interstitials(super.ref);
  int skips = 0;
  @override
  void skipNextCompletion() {
    skips++;
    super.skipNextCompletion();
  }
}

Future<
  ({_Unlocks unlocks, _Shower shower, _Interstitials Function() interstitials})
>
_pump(
  WidgetTester tester, {
  XpGain gain = _gain,
  bool lifetime = false,
  AdsConfig? placement,
  bool debugFallback = false,
}) async {
  final unlocks = _Unlocks();
  final shower = _Shower();
  late _Interstitials interstitials;
  await tester.pumpWidget(
    ProviderScope(
      key: UniqueKey(),
      overrides: [
        authControllerProvider.overrideWith(
          (_) => AuthController.test(
            AuthenticationState.authenticated(
              AuthenticatedUser(
                id: 'u',
                displayName: 'U',
                email: 'u@example.com',
                isLifetime: lifetime,
                isPremium: lifetime,
              ),
            ),
          ),
        ),
        xpBoostPlacementProvider.overrideWith((ref) async => placement),
        unlockRepositoryProvider.overrideWithValue(unlocks),
        unlockFlowProvider.overrideWithValue(
          UnlockFlow(
            repository: unlocks,
            shower: shower,
            useFallbackAfterTestAds: debugFallback,
          ),
        ),
        interstitialControllerProvider.overrideWith((ref) {
          interstitials = _Interstitials(ref);
          ref.onDispose(interstitials.dispose);
          return interstitials;
        }),
      ],
      child: MaterialApp(
        home: Scaffold(body: XpResultBadge(gain: gain)),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return (unlocks: unlocks, shower: shower, interstitials: () => interstitials);
}

const _adsOn = AdsConfig(
  enabled: true,
  xpBoostRewarded: AdPlacementConfig(
    enabled: true,
    unitId: testRewardedAdUnitId,
  ),
);

void main() {
  testWidgets('watching the ad doubles the XP and skips the full-screen ad', (
    tester,
  ) async {
    final env = await _pump(tester, placement: _adsOn);
    expect(find.text('Tonton iklan untuk 2× XP (+50 XP)'), findsOneWidget);
    await tester.tap(find.byKey(const Key('xp-boost-ad')));
    await tester.pumpAndSettle();
    expect(env.shower.shown, 1);
    expect(env.unlocks.calls, ['intent:xp-boost:grammar:attempt-1', 'claim']);
    expect(find.text('2× XP diterima (+50 XP)'), findsOneWidget);
    expect(env.interstitials().skips, 1);
  });

  testWidgets('Lifetime claims without an ad', (tester) async {
    final env = await _pump(tester, lifetime: true);
    expect(find.byKey(const Key('xp-boost-ad')), findsNothing);
    await tester.tap(find.byKey(const Key('xp-boost-free')));
    await tester.pumpAndSettle();
    expect(env.unlocks.calls, ['free:grammar:attempt-1']);
    expect(env.shower.shown, 0);
    expect(find.byKey(const Key('xp-boost-claimed')), findsOneWidget);
  });

  testWidgets('a debug build never fakes double XP through the fallback', (
    tester,
  ) async {
    final env = await _pump(tester, placement: _adsOn, debugFallback: true);
    await tester.tap(find.byKey(const Key('xp-boost-ad')));
    await tester.pumpAndSettle();
    expect(env.unlocks.calls, isNot(contains('fallback')));
  });

  testWidgets('not offered without XP, ads, or the switch for this kind', (
    tester,
  ) async {
    await _pump(
      tester,
      placement: _adsOn,
      gain: const XpGain(awarded: 0, total: 50, streak: 1, kind: 'lessons'),
    );
    expect(find.byKey(const Key('xp-boost-ad')), findsNothing);

    await _pump(tester);
    expect(find.byKey(const Key('xp-boost-ad')), findsNothing);

    await _pump(
      tester,
      placement: const AdsConfig(
        enabled: true,
        xpBoostRewarded: AdPlacementConfig(
          enabled: true,
          unitId: testRewardedAdUnitId,
        ),
        xpBoostKinds: CompletionKindsConfig(lessons: false),
      ),
    );
    expect(find.byKey(const Key('xp-boost-ad')), findsNothing);
  });
}
