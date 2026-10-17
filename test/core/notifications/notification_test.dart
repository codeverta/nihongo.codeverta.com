import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/notifications/notification_lifecycle.dart';
import 'package:nihongo_master/core/notifications/notification_settings_screen.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

class FakePush extends PushNotificationController {
  FakePush(super.preferences);
  @override
  Future<void> initialize() async {}
  void navigate(String route) {
    state = PushNotificationState(
      busy: false,
      available: true,
      navigation: route,
    );
  }
}

void main() {
  test(
    'notification navigation rejects external and unapproved destinations',
    () {
      for (final route in [
        'https://evil.example',
        '//evil.example',
        '/admin/',
        '/learn/grammar?level=N1',
        '/delete-account',
        null,
        42,
      ]) {
        expect(safeNotificationRoute(route), '/home');
      }
      expect(safeNotificationRoute('/dictionary'), '/dictionary');
      expect(safeNotificationRoute('/anki'), '/anki');
    },
  );

  testWidgets(
    'notification tap waits for onboarding and then opens destination',
    (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      final push = FakePush(prefs);
      final router = GoRouter(
        initialLocation: '/home',
        routes: [
          GoRoute(path: '/home', builder: (_, __) => const Text('home-view')),
          GoRoute(
            path: '/dictionary',
            builder: (_, __) => const Text('dictionary-view'),
          ),
        ],
      );
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(prefs),
          pushNotificationsProvider.overrideWith((ref) => push),
          appRouterProvider.overrideWithValue(router),
        ],
      );
      addTearDown(container.dispose);
      addTearDown(router.dispose);
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp.router(
            routerConfig: router,
            builder: (_, child) => NotificationLifecycle(child: child!),
          ),
        ),
      );
      push.navigate('/dictionary');
      await tester.pumpAndSettle();
      expect(find.text('home-view'), findsOneWidget);
      expect(push.state.navigation, '/dictionary');
      await container.read(isFirstTimeProvider.notifier).completeOnboarding();
      await tester.pumpAndSettle();
      expect(find.text('dictionary-view'), findsOneWidget);
      expect(push.state.navigation, isNull);
    },
  );

  testWidgets(
    'Apple push settings remain unavailable without requesting permission',
    (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [sharedPreferencesProvider.overrideWithValue(prefs)],
          child: const MaterialApp(home: NotificationSettingsScreen()),
        ),
      );
      await tester.pumpAndSettle();
      expect(
        find.textContaining('Dukungan iPhone/iPad sedang disiapkan'),
        findsOneWidget,
      );
      expect(
        tester.widget<SwitchListTile>(find.byType(SwitchListTile)).onChanged,
        isNull,
      );
    },
    variant: TargetPlatformVariant.only(TargetPlatform.iOS),
  );
}
