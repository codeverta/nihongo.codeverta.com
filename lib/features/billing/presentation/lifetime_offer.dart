import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Where the skippable Lifetime screen was opened from.
enum LifetimeOfferSource {
  afterLogin('after_login'),
  periodic('periodic');

  const LifetimeOfferSource(this.wire);
  final String wire;

  static LifetimeOfferSource? parse(String? value) =>
      values.where((source) => source.wire == value).firstOrNull;
}

/// Admin settings from `GET /v1/pricing` (Harga Lifetime Access page).
@immutable
class LifetimeOfferConfig {
  const LifetimeOfferConfig({this.afterLogin = true, this.intervalDays = 10});

  factory LifetimeOfferConfig.fromJson(Object? json) {
    if (json is! Map) return const LifetimeOfferConfig();
    final days = json['offerIntervalDays'];
    return LifetimeOfferConfig(
      afterLogin: json['offerAfterLogin'] != false,
      intervalDays: days is num ? days.toInt().clamp(0, 90) : 10,
    );
  }

  /// Master switch: off means no screen after login and none periodically.
  final bool afterLogin;

  /// Show again on a normal app open after this many days. Zero: never.
  final int intervalDays;
}

/// The settings, with the defaults when the server cannot be reached.
final lifetimeOfferConfigProvider = FutureProvider<LifetimeOfferConfig>((
  ref,
) async {
  try {
    final response = await ref
        .watch(dioProvider)
        .get<Map<String, dynamic>>('/pricing')
        .timeout(const Duration(seconds: 5));
    return LifetimeOfferConfig.fromJson(response.data?['data']);
  } on Object catch (error) {
    debugPrint('Lifetime offer settings unavailable: $error');
    return const LifetimeOfferConfig();
  }
});

/// Whether [user] may be shown the Lifetime offer at all: signed in, without
/// Lifetime or premium, and allowed to buy (not the demo or reviewer account).
bool lifetimeOfferEligible(AuthenticatedUser? user) =>
    user != null &&
    !user.isLifetime &&
    !user.isPremium &&
    !user.isOfflinePremium &&
    user.canPurchase;

/// A login target that is safe to open after the offer.
String lifetimeOfferTarget(String? from) {
  if (from == null || !from.startsWith('/') || from.startsWith('//')) {
    return '/home';
  }
  final path = Uri.tryParse(from)?.path ?? '';
  if (path == '/premium' || path == '/login' || path == '/splash') {
    return '/home';
  }
  return from;
}

final lifetimeOfferProvider = Provider<LifetimeOffer>(LifetimeOffer.new);

/// Decides when the skippable Lifetime screen opens: after every login, and
/// again on a normal app open once the admin interval has passed since it
/// was last shown on this device for this account.
class LifetimeOffer {
  LifetimeOffer(this._ref);
  final Ref _ref;

  /// The periodic check runs at most once per app process, on its launch.
  bool _launchChecked = false;

  /// Waits this long after Beranda first appears, so a notification tap that
  /// opens another page wins.
  @visibleForTesting
  Duration launchDelay = const Duration(milliseconds: 1500);

  @visibleForTesting
  DateTime Function() now = DateTime.now;

  static String _key(String userId) => 'lifetime_offer_last_shown_v1_$userId';

  Future<LifetimeOfferConfig> _config() => _ref
      .read(lifetimeOfferConfigProvider.future)
      .timeout(
        const Duration(seconds: 3),
        onTimeout: () => const LifetimeOfferConfig(),
      );

  void _shown(String userId, LifetimeOfferSource source) {
    unawaited(
      _ref
          .read(sharedPreferencesProvider)
          .setInt(_key(userId), now().millisecondsSinceEpoch),
    );
    _ref.read(analyticsServiceProvider).logEvent('lifetime_offer_view', {
      'source': source.wire,
    });
  }

  /// Where to go after a successful login: the offer (which continues to
  /// [from] when skipped) or [from] itself.
  Future<String> routeAfterLogin(String? from) async {
    final target = lifetimeOfferTarget(from);
    final user = _ref.read(authControllerProvider).user;
    if (!lifetimeOfferEligible(user)) return target;
    final config = await _config();
    if (!config.afterLogin) return target;
    _shown(user!.id, LifetimeOfferSource.afterLogin);
    return Uri(
      path: '/premium',
      queryParameters: {
        'offer': LifetimeOfferSource.afterLogin.wire,
        'from': target,
      },
    ).toString();
  }

  /// Called once Beranda appears. On the app's launch only, opens the offer
  /// when the interval has passed and the user is still on Beranda.
  Future<void> maybeShowOnLaunch(BuildContext context) async {
    if (_launchChecked) return;
    _launchChecked = true;
    final router = GoRouter.maybeOf(context);
    if (router == null) return;
    await Future<void>.delayed(launchDelay);
    final user = _ref.read(authControllerProvider).user;
    if (!lifetimeOfferEligible(user)) return;
    final config = await _config();
    if (!config.afterLogin || config.intervalDays <= 0) return;
    final last = _ref.read(sharedPreferencesProvider).getInt(_key(user!.id));
    if (last != null &&
        now().difference(DateTime.fromMillisecondsSinceEpoch(last)) <
            Duration(days: config.intervalDays)) {
      return;
    }
    if (!context.mounted ||
        router.routerDelegate.currentConfiguration.uri.path != '/home') {
      return;
    }
    // Still the same account and still without Lifetime after the wait.
    final current = _ref.read(authControllerProvider).user;
    if (current?.id != user.id || !lifetimeOfferEligible(current)) return;
    _shown(user.id, LifetimeOfferSource.periodic);
    unawaited(
      router.push('/premium?offer=${LifetimeOfferSource.periodic.wire}'),
    );
  }
}

/// Runs [LifetimeOffer.maybeShowOnLaunch] when Beranda first appears.
class LifetimeOfferLaunchCheck extends ConsumerStatefulWidget {
  const LifetimeOfferLaunchCheck({required this.child, super.key});
  final Widget child;

  @override
  ConsumerState<LifetimeOfferLaunchCheck> createState() =>
      _LifetimeOfferLaunchCheckState();
}

class _LifetimeOfferLaunchCheckState
    extends ConsumerState<LifetimeOfferLaunchCheck> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        unawaited(ref.read(lifetimeOfferProvider).maybeShowOnLaunch(context));
      }
    });
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
