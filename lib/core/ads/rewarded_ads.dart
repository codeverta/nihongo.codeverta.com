import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';

/// What happened when the user agreed to watch a rewarded ad.
enum RewardedAdOutcome {
  /// The ad finished and the reward was earned.
  watched,

  /// No ad could be loaded or shown (no fill, network, SDK error). Per the
  /// product decision the user is not blocked: this still grants a credit.
  unavailable,

  /// The user closed the ad before earning the reward.
  dismissed,
}

/// Loads and shows one full-screen rewarded ad. An interface so tests and
/// widget previews can substitute a fake.
// ignore: one_member_abstracts
abstract class RewardedAdShower {
  /// Shows one ad. [userId] and [customData] are passed to AdMob server-side
  /// verification, so Google's callback to the backend names the account and
  /// the unlock request; leave them out for ads the server does not verify.
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  });
}

class AdMobRewardedAdShower implements RewardedAdShower {
  static const _loadTimeout = Duration(seconds: 15);

  @override
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  }) async {
    final RewardedAd ad;
    try {
      ad = await _load(unitId).timeout(_loadTimeout);
      if (userId != null || customData != null) {
        await ad.setServerSideOptions(
          ServerSideVerificationOptions(userId: userId, customData: customData),
        );
      }
    } on Object catch (error) {
      debugPrint('Rewarded ad unavailable: $error');
      return RewardedAdOutcome.unavailable;
    }
    final result = Completer<RewardedAdOutcome>();
    var earned = false;
    ad.fullScreenContentCallback = FullScreenContentCallback(
      onAdDismissedFullScreenContent: (ad) {
        unawaited(ad.dispose());
        if (!result.isCompleted) {
          result.complete(
            earned ? RewardedAdOutcome.watched : RewardedAdOutcome.dismissed,
          );
        }
      },
      onAdFailedToShowFullScreenContent: (ad, error) {
        unawaited(ad.dispose());
        if (!result.isCompleted) result.complete(RewardedAdOutcome.unavailable);
      },
    );
    try {
      await ad.show(onUserEarnedReward: (_, _) => earned = true);
    } on Object catch (error) {
      debugPrint('Rewarded ad failed to show: $error');
      unawaited(ad.dispose());
      return RewardedAdOutcome.unavailable;
    }
    return result.future;
  }

  Future<RewardedAd> _load(String unitId) {
    final completer = Completer<RewardedAd>();
    unawaited(
      RewardedAd.load(
        adUnitId: unitId,
        request: const AdRequest(),
        rewardedAdLoadCallback: RewardedAdLoadCallback(
          onAdLoaded: completer.complete,
          onAdFailedToLoad: completer.completeError,
        ),
      ),
    );
    return completer.future;
  }
}

final rewardedAdShowerProvider = Provider<RewardedAdShower>(
  (ref) => AdMobRewardedAdShower(),
);

/// The Tanya AI rewarded placement when it can be offered right now: the
/// platform supports ads, the server switch is on, consent allows requests
/// and the user is on the free plan. Null means "send the user to Lifetime".
/// The rewarded placement that opens locked content, when an ad can be offered
/// right now (Android, server switch on, consent given, free plan). Null means
/// only Lifetime Access can open the content.
final unlockRewardedPlacementProvider = FutureProvider<AdsConfig?>((ref) async {
  final config = await ref.watch(adsReadyProvider.future);
  return config != null && config.unlockRewarded.usable ? config : null;
});

/// The Tanya AI rewarded placement for a Lifetime account that used its
/// daily messages (Android only; null on iOS or with ads off).
final aiChatLifetimeRewardedPlacementProvider =
    FutureProvider<AdPlacementConfig?>((ref) async {
      final config = await ref.watch(optInAdsReadyProvider.future);
      final placement = config?.aiChatRewarded;
      return placement != null && placement.usable ? placement : null;
    });

final aiChatRewardedPlacementProvider = FutureProvider<AdPlacementConfig?>((
  ref,
) async {
  final config = await ref.watch(adsReadyProvider.future);
  final placement = config?.aiChatRewarded;
  return placement != null && placement.usable ? placement : null;
});
