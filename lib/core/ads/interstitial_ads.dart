import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';

/// Default minutes of foreground use between full-screen ads.
const defaultInterstitialMinutes = 10;

/// Counts foreground app time since the last full-screen ad. Time spent in
/// the background never counts.
class InterstitialPacer {
  InterstitialPacer({
    this.interval = const Duration(minutes: defaultInterstitialMinutes),
  });

  Duration interval;
  Duration _accumulated = Duration.zero;
  DateTime? _resumedAt;

  void resume(DateTime now) => _resumedAt ??= now;

  void pause(DateTime now) {
    final since = _resumedAt;
    if (since == null) return;
    _accumulated += now.difference(since);
    _resumedAt = null;
  }

  Duration elapsed(DateTime now) {
    final since = _resumedAt;
    return _accumulated +
        (since == null ? Duration.zero : now.difference(since));
  }

  bool due(DateTime now) => elapsed(now) >= interval;

  /// Start loading an ad a minute early so it is ready at the next break.
  bool nearlyDue(DateTime now) =>
      elapsed(now) >= interval - const Duration(minutes: 1);

  void reset(DateTime now) {
    _accumulated = Duration.zero;
    if (_resumedAt != null) _resumedAt = now;
  }
}

/// Loads and shows one interstitial. An interface so tests can fake it.
abstract class InterstitialAdGateway {
  bool get ready;
  Future<void> load(String unitId);

  /// Shows the loaded ad and completes when it is closed. Returns whether an
  /// ad was actually shown.
  Future<bool> show();
  void dispose();
}

class AdMobInterstitialGateway implements InterstitialAdGateway {
  InterstitialAd? _ad;
  bool _loading = false;

  @override
  bool get ready => _ad != null;

  @override
  Future<void> load(String unitId) async {
    if (_ad != null || _loading) return;
    _loading = true;
    final completer = Completer<void>();
    await InterstitialAd.load(
      adUnitId: unitId,
      request: const AdRequest(),
      adLoadCallback: InterstitialAdLoadCallback(
        onAdLoaded: (ad) {
          _ad = ad;
          _loading = false;
          if (!completer.isCompleted) completer.complete();
        },
        onAdFailedToLoad: (error) {
          debugPrint('Interstitial unavailable: $error');
          _loading = false;
          if (!completer.isCompleted) completer.complete();
        },
      ),
    );
    return completer.future;
  }

  @override
  Future<bool> show() async {
    final ad = _ad;
    if (ad == null) return false;
    _ad = null;
    final closed = Completer<bool>();
    ad.fullScreenContentCallback = FullScreenContentCallback(
      onAdDismissedFullScreenContent: (ad) {
        unawaited(ad.dispose());
        if (!closed.isCompleted) closed.complete(true);
      },
      onAdFailedToShowFullScreenContent: (ad, error) {
        unawaited(ad.dispose());
        if (!closed.isCompleted) closed.complete(false);
      },
    );
    try {
      await ad.show();
    } on Object catch (error) {
      debugPrint('Interstitial failed to show: $error');
      unawaited(ad.dispose());
      return false;
    }
    return closed.future;
  }

  @override
  void dispose() {
    unawaited(_ad?.dispose());
    _ad = null;
  }
}

final interstitialAdGatewayProvider = Provider<InterstitialAdGateway>((ref) {
  final gateway = AdMobInterstitialGateway();
  ref.onDispose(gateway.dispose);
  return gateway;
});

/// Shows an interstitial only when the user leaves a result screen, after
/// they have seen their score, and only once the configured minutes of use
/// have passed since the last full-screen ad. Never while studying, never
/// when opening content. Free users only: [adsReadyProvider] yields null for
/// Premium and Lifetime users.
class InterstitialController {
  InterstitialController(this._ref, {DateTime Function()? clock})
    : _clock = clock ?? DateTime.now {
    final state = WidgetsBinding.instance.lifecycleState;
    if (state == null || state == AppLifecycleState.resumed) {
      pacer.resume(_clock());
    }
    _lifecycle = AppLifecycleListener(
      onStateChange: (state) {
        if (state == AppLifecycleState.resumed) {
          pacer.resume(_clock());
        } else {
          pacer.pause(_clock());
        }
      },
    );
  }

  final Ref _ref;
  final DateTime Function() _clock;
  final pacer = InterstitialPacer();
  late final AppLifecycleListener _lifecycle;
  bool _showing = false;

  /// Set when the provider is disposed (sign-out, app restart, tests). A
  /// result screen torn down with it must not try to show an ad afterwards.
  bool _disposed = false;

  /// Completions the user already paid for with a rewarded ad (opening the
  /// content, or doubling its XP). Their result screen gets no full-screen ad.
  final Set<String> _paidWithAd = {};

  AdPlacementConfig? _placementFor(CompletionKind kind) {
    if (_disposed) return null;
    final config = _ref.read(adsReadyProvider).valueOrNull;
    final placement = config?.interstitial;
    if (config == null || placement == null || !placement.usable) return null;
    if (!config.interstitialKinds.allows(kind)) return null;
    final minutes = placement.every >= 5 && placement.every <= 60
        ? placement.every
        : defaultInterstitialMinutes;
    pacer.interval = Duration(minutes: minutes);
    return placement;
  }

  /// Call when a result screen appears. Starts loading an ad if one may be
  /// shown when the user leaves, so they never wait on a spinner.
  void prepareAfterCompletion(CompletionKind kind) {
    final placement = _placementFor(kind);
    if (placement == null || !pacer.nearlyDue(_clock())) return;
    final gateway = _ref.read(interstitialAdGatewayProvider);
    if (!gateway.ready) unawaited(gateway.load(placement.unitId));
  }

  /// Call when the user leaves a result screen. [key] names the completed
  /// item (for example `grammar:grammar-n3-01`) so a completion paid for
  /// with a rewarded ad is skipped.
  Future<void> showAfterCompletion(CompletionKind kind, {String? key}) async {
    if (_skipNext) {
      _skipNext = false;
      return;
    }
    if (key != null && _paidWithAd.remove(key)) return;
    final placement = _placementFor(kind);
    if (placement == null || _showing) return;
    final gateway = _ref.read(interstitialAdGatewayProvider);
    final now = _clock();
    if (!gateway.ready) {
      // Not loaded in time: load for the next completion instead of making
      // the user wait.
      if (pacer.nearlyDue(now)) unawaited(gateway.load(placement.unitId));
      return;
    }
    if (!pacer.due(now)) return;
    _showing = true;
    try {
      if (await gateway.show()) markFullScreenAdShown();
    } finally {
      _showing = false;
    }
  }

  bool _skipNext = false;

  /// The result screen on display was paid for with a rewarded ad (double
  /// XP): leaving it shows no full-screen ad.
  void skipNextCompletion() => _skipNext = true;

  /// Records that the user watched a rewarded ad for [key], so the result
  /// screen of that item shows no full-screen ad. Also restarts the count.
  void markPaidWithAd(String key) {
    _paidWithAd.add(key);
    markFullScreenAdShown();
  }

  /// Restarts the count after any full-screen ad, including rewarded ones,
  /// so users never see two in quick succession.
  void markFullScreenAdShown() => pacer.reset(_clock());

  void dispose() {
    _disposed = true;
    _lifecycle.dispose();
  }
}

final interstitialControllerProvider = Provider<InterstitialController>((ref) {
  final controller = InterstitialController(ref);
  ref.onDispose(controller.dispose);
  return controller;
});
