import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';

enum UnlockOutcome {
  /// The item is open now.
  unlocked,

  /// The user closed the ad before the reward.
  cancelled,

  /// No ad could be loaded or shown.
  adUnavailable,

  /// The ad was watched but Google's confirmation has not reached the server
  /// yet. [UnlockAttempt.nonce] can be claimed again later.
  pending,
}

@immutable
class UnlockAttempt {
  const UnlockAttempt(this.outcome, {this.nonce = ''});
  final UnlockOutcome outcome;
  final String nonce;
}

/// Runs one "watch an ad to open this" attempt. The server only opens the
/// item after Google's signed server-side verification callback, so the app
/// never decides on its own that an ad counted.
class UnlockFlow {
  UnlockFlow({
    required this.repository,
    required this.shower,
    bool? useFallbackAfterTestAds,
    this.claimAttempts = 3,
  }) : useFallbackAfterTestAds = useFallbackAfterTestAds ?? kDebugMode;

  final UnlockRepository repository;
  final RewardedAdShower shower;

  /// Debug builds show Google's test ads, which never call this backend, so
  /// a watched test ad is recorded through the (limited) fallback instead.
  final bool useFallbackAfterTestAds;

  /// How many times to ask the server, which itself waits a few seconds for
  /// Google's callback each time.
  final int claimAttempts;

  Future<UnlockAttempt> watchAd(UnlockTarget target, String unitId) async {
    final intent = await repository.createIntent(target);
    if (intent.alreadyUnlocked) {
      return const UnlockAttempt(UnlockOutcome.unlocked);
    }
    final outcome = await shower.show(
      unitId,
      userId: intent.userId,
      customData: intent.nonce,
    );
    switch (outcome) {
      case RewardedAdOutcome.dismissed:
        return const UnlockAttempt(UnlockOutcome.cancelled);
      case RewardedAdOutcome.unavailable:
        return const UnlockAttempt(UnlockOutcome.adUnavailable);
      case RewardedAdOutcome.watched:
        // Double XP has no fallback (it is a bonus, not access), so a test ad
        // in a debug build cannot earn it; the claim then stays pending.
        if (useFallbackAfterTestAds && target.kind != 'xp-boost') {
          await repository.fallback(target);
          return const UnlockAttempt(UnlockOutcome.unlocked);
        }
        return claim(intent.nonce);
    }
  }

  /// Asks the server whether Google has confirmed the ad for [nonce].
  Future<UnlockAttempt> claim(String nonce) async {
    for (var attempt = 0; attempt < claimAttempts; attempt++) {
      if (await repository.claim(nonce)) {
        return const UnlockAttempt(UnlockOutcome.unlocked);
      }
    }
    return UnlockAttempt(UnlockOutcome.pending, nonce: nonce);
  }

  /// Opens the item without an ad, for when no ad could be loaded.
  Future<void> fallback(UnlockTarget target) => repository.fallback(target);
}

final unlockFlowProvider = Provider<UnlockFlow>(
  (ref) => UnlockFlow(
    repository: ref.watch(unlockRepositoryProvider),
    shower: ref.watch(rewardedAdShowerProvider),
  ),
);
