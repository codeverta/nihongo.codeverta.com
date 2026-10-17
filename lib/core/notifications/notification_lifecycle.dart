import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';
import 'package:nihongo_master/core/routing/app_router.dart';

class NotificationLifecycle extends ConsumerStatefulWidget {
  const NotificationLifecycle({required this.child, super.key});
  final Widget child;
  @override
  ConsumerState<NotificationLifecycle> createState() =>
      _NotificationLifecycleState();
}

class _NotificationLifecycleState extends ConsumerState<NotificationLifecycle>
    with WidgetsBindingObserver {
  bool _navigationScheduled = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      unawaited(ref.read(pushNotificationsProvider.notifier).refresh());
    }
  }

  void _navigate() {
    if (_navigationScheduled ||
        ref.read(pushNotificationsProvider).navigation == null ||
        ref.read(isFirstTimeProvider)) {
      return;
    }
    _navigationScheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _navigationScheduled = false;
      if (!mounted || ref.read(isFirstTimeProvider)) return;
      final route = ref.read(pushNotificationsProvider).navigation;
      if (route == null) return;
      ref.read(pushNotificationsProvider.notifier).consumeNavigation();
      ref.read(appRouterProvider).go(safeNotificationRoute(route));
    });
  }

  @override
  Widget build(BuildContext context) {
    ref
      ..listen(pushNotificationsProvider, (_, next) {
        if (next.navigation != null) _navigate();
      })
      ..listen(isFirstTimeProvider, (_, next) {
        if (!next) _navigate();
      });
    return widget.child;
  }
}
