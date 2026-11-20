import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';
import 'package:nihongo_master/core/gamification/xp_gain.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/unlock/locked_content.dart';
import 'package:nihongo_master/core/unlock/unlock_flow.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

/// The double XP placement when an ad can be offered right now (Android, ads
/// on, consent given, free plan); null otherwise.
final xpBoostPlacementProvider = FutureProvider<AdsConfig?>((ref) async {
  final config = await ref.watch(adsReadyProvider.future);
  return config != null && config.xpBoostRewarded.usable ? config : null;
});

CompletionKind? _completionKind(String kind) => switch (kind) {
  'lessons' => CompletionKind.lessons,
  'exams' => CompletionKind.exams,
  'practice' => CompletionKind.practice,
  _ => null,
};

/// "Watch an ad for 2× XP (+N XP)" under a result's XP, or "Claim 2× XP"
/// without an ad for Lifetime accounts (docs/ads-xp-spec.md, phase C). The
/// server checks everything (verified ad, once, within 30 minutes, daily
/// limit); this only decides whether to offer it.
class DoubleXpButton extends ConsumerStatefulWidget {
  const DoubleXpButton({required this.gain, super.key});
  final XpGain gain;

  @override
  ConsumerState<DoubleXpButton> createState() => _DoubleXpButtonState();
}

class _DoubleXpButtonState extends ConsumerState<DoubleXpButton> {
  bool _busy = false;
  int? _bonus;
  String _pendingNonce = '';

  void _message(String text) =>
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(text)));

  void _claimed(int bonus) {
    ref.read(interstitialControllerProvider)
      ..markFullScreenAdShown()
      // The user just paid for this result with an ad: leaving it shows none.
      ..skipNextCompletion();
    ref
      ..invalidate(gamificationProvider)
      ..invalidate(homeDashboardProvider);
    setState(() => _bonus = bonus);
  }

  Future<void> _run(Future<void> Function() action) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      await action();
    } on UnlockException catch (error) {
      if (mounted) {
        _message(
          error.message.isNotEmpty
              ? error.message
              : 'Belum dapat mengklaim 2× XP. Periksa koneksi lalu coba lagi.'
                    .localized(context),
        );
      }
    } on Object {
      if (mounted) {
        _message(
          'Belum dapat mengklaim 2× XP. Periksa koneksi lalu coba lagi.'
              .localized(context),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _claimFree() => _run(() async {
    final gain = await ref
        .read(unlockRepositoryProvider)
        .claimFreeBoost(widget.gain.completionId);
    if (mounted) _claimed(gain?.awarded ?? widget.gain.awarded);
  });

  Future<void> _watchAd(String unitId) => _run(() async {
    final flow = ref.read(unlockFlowProvider);
    final attempt = _pendingNonce.isNotEmpty
        ? await flow.claim(_pendingNonce)
        : await flow.watchAd(
            UnlockTarget('xp-boost', widget.gain.completionId),
            unitId,
          );
    if (!mounted) return;
    if (attempt.outcome != UnlockOutcome.adUnavailable) {
      ref.read(interstitialControllerProvider).markFullScreenAdShown();
    }
    switch (attempt.outcome) {
      case UnlockOutcome.unlocked:
        _pendingNonce = '';
        _claimed(widget.gain.awarded);
      case UnlockOutcome.cancelled:
        _message(
          'Iklan ditutup sebelum selesai. Tonton sampai habis untuk mendapat 2× XP.'
              .localized(context),
        );
      case UnlockOutcome.adUnavailable:
        _message(
          'Iklan belum tersedia. Coba lagi sebentar lagi.'.localized(context),
        );
      case UnlockOutcome.pending:
        setState(() => _pendingNonce = attempt.nonce);
        _message(
          'Iklan sudah ditonton, tetapi konfirmasi dari Google belum sampai. Coba periksa lagi.'
              .localized(context),
        );
    }
  });

  @override
  Widget build(BuildContext context) {
    final gain = widget.gain;
    final bonus = _bonus;
    if (bonus != null) {
      return Text(
        '${'2× XP diterima'.localized(context)} (+$bonus XP)',
        key: const Key('xp-boost-claimed'),
        textAlign: TextAlign.center,
        style: TextStyle(
          fontWeight: FontWeight.w800,
          color: Theme.of(context).colorScheme.primary,
        ),
      );
    }
    final kind = _completionKind(gain.kind);
    if (gain.awarded <= 0 || gain.completionId.isEmpty || kind == null) {
      return const SizedBox.shrink();
    }
    if (hasLifetimeAccess(ref)) {
      return FilledButton.tonalIcon(
        key: const Key('xp-boost-free'),
        onPressed: _busy ? null : _claimFree,
        icon: const Icon(Icons.bolt_rounded),
        label: Text('Klaim 2× XP'.localized(context)),
      );
    }
    final config = ref.watch(xpBoostPlacementProvider).valueOrNull;
    if (config == null || !config.xpBoostKinds.allows(kind)) {
      return const SizedBox.shrink();
    }
    return FilledButton.tonalIcon(
      key: const Key('xp-boost-ad'),
      onPressed: _busy ? null : () => _watchAd(config.xpBoostRewarded.unitId),
      icon: _busy
          ? const SizedBox.square(
              dimension: 18,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : Icon(
              _pendingNonce.isEmpty
                  ? Icons.play_circle_outline_rounded
                  : Icons.refresh_rounded,
            ),
      label: Text(
        _pendingNonce.isEmpty
            ? '${'Tonton iklan untuk 2× XP'.localized(context)} (+${gain.awarded} XP)'
            : 'Periksa lagi'.localized(context),
      ),
    );
  }
}
