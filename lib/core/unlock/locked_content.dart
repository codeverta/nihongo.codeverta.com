import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';
import 'package:nihongo_master/core/unlock/unlock_sheet.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Content the server refused with 402 `premium_required`. [target] is the
/// item a rewarded ad can open, or null when only Lifetime Access can.
/// Feature repositories throw this instead of their own exception type so the
/// unlock details in the response body survive.
class LockedContentException implements Exception {
  const LockedContentException({this.target, this.message = ''});

  /// Reads a 402 response from [error], which may be a raw [DioException] or
  /// an already-mapped [LockedContentException]. Null for anything else.
  static LockedContentException? from(Object? error) {
    if (error is LockedContentException) return error;
    if (!UnlockTarget.isLocked(error)) return null;
    final data = (error! as DioException).response?.data;
    final body = data is Map ? data['error'] : null;
    return LockedContentException(
      target: UnlockTarget.fromError(error),
      message: body is Map ? '${body['message'] ?? ''}'.trim() : '',
    );
  }

  final UnlockTarget? target;
  final String message;

  @override
  String toString() =>
      message.isEmpty ? 'Konten ini memerlukan Lifetime Access.' : message;
}

/// The locked-content view for [error], or null when [error] is not a lock.
/// Lets an error builder write `lockedContentView(e, reload) ?? otherView`.
Widget? lockedContentView(Object? error, VoidCallback onUnlocked) {
  final locked = LockedContentException.from(error);
  return locked == null
      ? null
      : LockedContentView(locked: locked, onUnlocked: onUnlocked);
}

/// True when [error] is the server asking a guest to sign in first.
bool isLoginRequired(Object? error) {
  if (error is! DioException || error.response?.statusCode != 401) return false;
  final data = error.response?.data;
  final body = data is Map ? data['error'] : null;
  return body is Map && body['code'] == 'login_required';
}

/// Full-screen state for locked content: explains the lock and offers the
/// unlock sheet (ad or Lifetime). [onUnlocked] runs after the item opened,
/// typically to reload it.
class LockedContentView extends ConsumerWidget {
  const LockedContentView({
    required this.locked,
    required this.onUnlocked,
    this.title,
    super.key,
  });

  final LockedContentException locked;
  final VoidCallback onUnlocked;
  final String? title;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final target = locked.target;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.lock_outline_rounded,
              size: 44,
              color: theme.colorScheme.primary,
            ),
            const SizedBox(height: 12),
            Text(
              title ?? 'Konten ini terkunci'.localized(context),
              style: theme.textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              target == null
                  ? 'Konten ini memerlukan Lifetime Access.'.localized(context)
                  : 'Buka dengan menonton satu iklan, atau beli Lifetime untuk membuka semuanya.'
                        .localized(context),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              key: const Key('locked-content-open'),
              onPressed: () async {
                if (target == null) {
                  await _openLifetime(context);
                  return;
                }
                final opened = await showUnlockSheet(context, target);
                if (opened) onUnlocked();
              },
              icon: Icon(
                target == null
                    ? Icons.workspace_premium_rounded
                    : Icons.lock_open_rounded,
              ),
              label: Text(
                target == null
                    ? 'Lihat Lifetime'.localized(context)
                    : 'Buka konten ini'.localized(context),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

Future<void> _openLifetime(BuildContext context) async {
  String? from;
  try {
    from = GoRouterState.of(context).uri.toString();
  } on Object {
    from = null;
  }
  final query = from == null ? '' : '?from=${Uri.encodeQueryComponent(from)}';
  await context.push('/premium$query');
}

/// Whether the account holds Lifetime Access (or premium), which opens every
/// locked item without ads.
bool hasLifetimeAccess(WidgetRef ref) {
  final user = ref.watch(authControllerProvider.select((s) => s.user));
  return (user?.isLifetime ?? false) || (user?.isPremium ?? false);
}

/// Lock badge for a list row: a lock while [target] is locked for this
/// account right now, otherwise [unlocked]. Only drawn when the server says
/// the gate for that group is on, so open content never shows a lock.
class UnlockLockIcon extends ConsumerWidget {
  const UnlockLockIcon({
    required this.target,
    required this.lockable,
    this.unlocked = const Icon(Icons.chevron_right_rounded),
    super.key,
  });

  final UnlockTarget target;

  /// Whether this item can be locked at all (for lessons: level N3–N1).
  final bool lockable;
  final Widget unlocked;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (!lockable || hasLifetimeAccess(ref)) return unlocked;
    final status = ref.watch(unlockStatusProvider).valueOrNull;
    if (status == null ||
        status.gates[target.group] != true ||
        status.isOpen(target)) {
      return unlocked;
    }
    return Icon(
      Icons.lock_outline_rounded,
      key: Key('locked-${target.key}'),
      semanticLabel: 'Terkunci'.localized(context),
    );
  }
}

/// N3–N1 lessons and decks are the lockable study levels.
bool isLockableLevel(String level) =>
    const {'N3', 'N2', 'N1'}.contains(level.toUpperCase());
