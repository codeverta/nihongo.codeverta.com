import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/unlock/unlock_flow.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';

/// Offers to open [target] by watching one rewarded ad, or by buying Lifetime
/// Access. Returns true when the item is open afterwards. The ad option is
/// only shown where an ad can be shown (Android, ads on, kind switched on);
/// otherwise the sheet offers Lifetime Access only.
///
/// [returnTo] is where the Lifetime purchase screen returns afterwards; it
/// defaults to the current route.
Future<bool> showUnlockSheet(
  BuildContext context,
  UnlockTarget target, {
  String? title,
  String? returnTo,
}) async {
  var from = returnTo;
  if (from == null) {
    try {
      from = GoRouterState.of(context).uri.toString();
    } on Object {
      from = null;
    }
  }
  final result = await showModalBottomSheet<UnlockSheetResult>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    builder: (_) => UnlockSheet(target: target, title: title),
  );
  // The Lifetime screen is opened from the caller's context, after the sheet
  // is gone. Swiping the sheet away yields null and does nothing.
  if (result == UnlockSheetResult.lifetime && context.mounted) {
    final query = from == null ? '' : '?from=${Uri.encodeQueryComponent(from)}';
    // Not awaited: after a purchase the Lifetime screen returns to [from]
    // itself, which reloads the content there.
    unawaited(context.push('/premium$query'));
    return false;
  }
  return result == UnlockSheetResult.opened;
}

/// How the unlock sheet was closed.
enum UnlockSheetResult { opened, lifetime, closed }

class UnlockSheet extends ConsumerStatefulWidget {
  const UnlockSheet({required this.target, this.title, super.key});
  final UnlockTarget target;
  final String? title;

  @override
  ConsumerState<UnlockSheet> createState() => _UnlockSheetState();
}

class _UnlockSheetState extends ConsumerState<UnlockSheet> {
  bool _busy = false;
  String? _message;
  int _adFailures = 0;
  String _pendingNonce = '';

  /// After this many ads fail to load, the limited no-ad option appears.
  static const _failuresBeforeFallback = 2;

  Future<void> _run(Future<UnlockAttempt> Function() action) async {
    if (_busy) return;
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      final attempt = await action();
      if (!mounted) return;
      final interstitials = ref.read(interstitialControllerProvider);
      if (attempt.outcome != UnlockOutcome.adUnavailable) {
        interstitials.markFullScreenAdShown();
      }
      switch (attempt.outcome) {
        case UnlockOutcome.unlocked:
          interstitials.markPaidWithAd(widget.target.key);
          ref.invalidate(unlockStatusProvider);
          Navigator.of(context).pop(UnlockSheetResult.opened);
          return;
        case UnlockOutcome.cancelled:
          _message =
              'Iklan ditutup sebelum selesai. Tonton sampai habis untuk membuka.'
                  .localized(context);
        case UnlockOutcome.adUnavailable:
          _adFailures++;
          _message = 'Iklan belum tersedia. Coba lagi sebentar lagi.'.localized(
            context,
          );
        case UnlockOutcome.pending:
          _pendingNonce = attempt.nonce;
          _message =
              'Iklan sudah ditonton, tetapi konfirmasi dari Google belum sampai. Coba periksa lagi.'
                  .localized(context);
      }
    } on UnlockException catch (error) {
      if (!mounted) return;
      _message = error.message.isNotEmpty
          ? error.message
          : 'Belum dapat membuka. Periksa koneksi lalu coba lagi.'.localized(
              context,
            );
    } on Object {
      if (!mounted) return;
      _message = 'Belum dapat membuka. Periksa koneksi lalu coba lagi.'
          .localized(context);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final placement = ref.watch(unlockRewardedPlacementProvider).valueOrNull;
    final status =
        ref.watch(unlockStatusProvider).valueOrNull ?? UnlockStatus.unknown;
    final flow = ref.watch(unlockFlowProvider);
    final target = widget.target;
    final adOffered =
        placement != null &&
        placement.unlockKinds.allows(target.kind) &&
        (status.offered[target.group] ?? true);
    final fallbackOffered =
        adOffered &&
        _adFailures >= _failuresBeforeFallback &&
        status.fallbackRemaining > 0;
    final theme = Theme.of(context);

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 0, 24, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Icon(
              Icons.lock_open_rounded,
              size: 36,
              color: theme.colorScheme.primary,
            ),
            const SizedBox(height: 12),
            Text(
              widget.title ?? 'Buka konten ini'.localized(context),
              style: theme.textTheme.titleLarge,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              adOffered
                  ? 'Tonton satu iklan singkat untuk membuka konten ini, atau beli Lifetime untuk membuka semuanya tanpa iklan.'
                        .localized(context)
                  : 'Konten ini memerlukan Lifetime Access.'.localized(context),
              style: theme.textTheme.bodyMedium,
              textAlign: TextAlign.center,
            ),
            if (adOffered && status.dailyCap > 0) ...[
              const SizedBox(height: 6),
              Text(
                '${'Sisa buka gratis hari ini'.localized(context)}: ${status.remaining}',
                style: theme.textTheme.bodySmall,
                textAlign: TextAlign.center,
              ),
            ],
            if (_message != null) ...[
              const SizedBox(height: 12),
              Text(
                _message!,
                key: const Key('unlock-message'),
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.error,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            const SizedBox(height: 20),
            if (_pendingNonce.isNotEmpty)
              FilledButton.icon(
                key: const Key('unlock-check-again'),
                onPressed: _busy
                    ? null
                    : () => _run(() => flow.claim(_pendingNonce)),
                icon: const Icon(Icons.refresh_rounded),
                label: Text('Periksa lagi'.localized(context)),
              )
            else if (adOffered)
              FilledButton.icon(
                key: const Key('unlock-watch-ad'),
                onPressed: _busy
                    ? null
                    : () => _run(
                        () => flow.watchAd(
                          target,
                          placement.unlockRewarded.unitId,
                        ),
                      ),
                icon: _busy
                    ? const SizedBox.square(
                        dimension: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.play_circle_outline_rounded),
                label: Text('Tonton iklan (gratis)'.localized(context)),
              ),
            if (fallbackOffered) ...[
              const SizedBox(height: 8),
              OutlinedButton(
                key: const Key('unlock-fallback'),
                onPressed: _busy
                    ? null
                    : () => _run(() async {
                        await flow.fallback(target);
                        return const UnlockAttempt(UnlockOutcome.unlocked);
                      }),
                child: Text(
                  '${'Buka tanpa iklan'.localized(context)} (${status.fallbackRemaining})',
                ),
              ),
            ],
            const SizedBox(height: 8),
            OutlinedButton.icon(
              key: const Key('unlock-lifetime'),
              onPressed: _busy
                  ? null
                  : () => Navigator.of(context).pop(UnlockSheetResult.lifetime),
              icon: const Icon(Icons.workspace_premium_rounded),
              label: Text('Beli Lifetime'.localized(context)),
            ),
            TextButton(
              onPressed: _busy
                  ? null
                  : () => Navigator.of(context).pop(UnlockSheetResult.closed),
              child: Text('Batal'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}
