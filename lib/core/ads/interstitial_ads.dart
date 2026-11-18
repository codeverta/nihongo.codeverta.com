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

/// Screens a full-screen ad must never cover: startup, sign-in and the
/// Lifetime purchase flow.
bool interstitialAllowedOn(String path) {
  if (path.isEmpty || path == '/') return false;
  for (final blocked in const ['/onboarding', '/login', '/premium']) {
    if (path == blocked ||
        path.startsWith('$blocked/') ||
        path.startsWith('$blocked?')) {
      return false;
    }
  }
  return true;
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

/// Shows an interstitial at natural breaks (screen or tab changes) once the
/// user has used the app for the configured minutes. Free users only;
/// Premium and Lifetime users never get one because [adsReadyProvider]
/// yields null for them.
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

  AdPlacementConfig? get _placement {
    final placement = _ref.read(adsReadyProvider).valueOrNull?.interstitial;
    return placement != null && placement.usable ? placement : null;
  }

  /// Call when the user moves to another screen or tab.
  Future<void> onNaturalBreak(String path) async {
    final placement = _placement;
    if (placement == null || _showing) return;
    final minutes = placement.every >= 5 && placement.every <= 60
        ? placement.every
        : defaultInterstitialMinutes;
    pacer.interval = Duration(minutes: minutes);
    final now = _clock();
    final gateway = _ref.read(interstitialAdGatewayProvider);
    if (!pacer.nearlyDue(now)) return;
    if (!gateway.ready) {
      // Load now and show at the following break rather than making the user
      // wait on a spinner.
      unawaited(gateway.load(placement.unitId));
      return;
    }
    if (!pacer.due(now) || !interstitialAllowedOn(path)) return;
    _showing = true;
    try {
      if (await gateway.show()) markFullScreenAdShown();
    } finally {
      _showing = false;
    }
  }

  /// Restarts the count after any full-screen ad, including rewarded ones,
  /// so users never see two in quick succession.
  void markFullScreenAdShown() => pacer.reset(_clock());

  void dispose() => _lifecycle.dispose();
}

final interstitialControllerProvider = Provider<InterstitialController>((ref) {
  final controller = InterstitialController(ref);
  ref.onDispose(controller.dispose);
  return controller;
});
