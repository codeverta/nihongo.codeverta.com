import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/notifications/notification_prompt.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_notification_card.dart';
import 'package:nihongo_master/features/onboarding/presentation/screens/onboarding_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

class _FakePush extends PushNotificationController {
  _FakePush(super.preferences, {this.result = PushToggleResult.on});
  final PushToggleResult result;
  int enableCalls = 0;

  @override
  Future<PushToggleResult> setEnabled({required bool enabled}) async {
    enableCalls++;
    return result;
  }

  @override
  Future<void> initialize() async {}
}

Future<
  ({
    ProviderContainer container,
    SharedPreferences prefs,
    _FakePush push,
    List<String> completed,
  })
>
_pumpOnboarding(
  WidgetTester tester, {
  PushToggleResult result = PushToggleResult.on,
  bool withOptIn = true,
  Locale locale = const Locale('id'),
  Size size = const Size(390, 844),
  double textScale = 1,
}) async {
  tester.view
    ..physicalSize = size
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  late _FakePush push;
  final container = ProviderContainer(
    overrides: [
      sharedPreferencesProvider.overrideWithValue(prefs),
      pushNotificationsProvider.overrideWith((ref) {
        return push = _FakePush(prefs, result: result);
      }),
    ],
  );
  addTearDown(container.dispose);
  push = container.read(pushNotificationsProvider.notifier) as _FakePush;
  final completed = <String>[];
  final router = GoRouter(
    initialLocation: '/onboarding',
    routes: [
      GoRoute(
        path: '/onboarding',
        builder: (_, _) => OnboardingScreen(
          onCompleted: () async => completed.add('done'),
          notificationOptIn: withOptIn
              ? OnboardingNotificationOptIn(
                  enable: () => enableNotificationsFromPrompt(
                    container.read,
                    NotificationPromptPlace.onboarding,
                  ),
                  skip: () => skipNotificationPrompt(
                    container.read,
                    NotificationPromptPlace.onboarding,
                  ),
                )
              : null,
        ),
      ),
      GoRoute(
        path: '/home',
        builder: (_, _) => const Scaffold(body: Text('Home')),
      ),
    ],
  );
  addTearDown(router.dispose);
  await tester.pumpWidget(
    UncontrolledProviderScope(
      container: container,
      child: MaterialApp.router(
        routerConfig: router,
        locale: locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(
            context,
          ).copyWith(textScaler: TextScaler.linear(textScale)),
          child: child!,
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return (container: container, prefs: prefs, push: push, completed: completed);
}

Future<void> _tapEnable(WidgetTester tester) async {
  final enable = find.byKey(const Key('onboarding-notification-enable'));
  await tester.ensureVisible(enable);
  await tester.pumpAndSettle();
  await tester.tap(enable);
  await tester.pumpAndSettle();
}

Future<void> _tapSkip(WidgetTester tester) async {
  final skip = find.byKey(const Key('onboarding-notification-skip'));
  await tester.ensureVisible(skip);
  await tester.pumpAndSettle();
  await tester.tap(skip);
  await tester.pumpAndSettle();
}

Future<void> _toNotificationPage(WidgetTester tester) async {
  for (var i = 0; i < 3; i++) {
    await tester.tap(find.byKey(const Key('onboarding-next')));
    await tester.pumpAndSettle();
  }
}

void main() {
  testWidgets('without the hook the guide stays three pages', (tester) async {
    await _pumpOnboarding(tester, withOptIn: false);
    expect(find.byKey(const Key('onboarding-page-3')), findsNothing);
    expect(find.text('01 / 03', findRichText: true), findsNothing);
    expect(find.textContaining('01 / 03'), findsOneWidget);
  });

  testWidgets('a fourth page explains first, no system dialog yet', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester);
    expect(find.byKey(const Key('onboarding-page-3')), findsOneWidget);
    expect(find.textContaining('01 / 04'), findsOneWidget);
    await _toNotificationPage(tester);
    expect(find.text('Jangan ketinggalan kabar penting'), findsOneWidget);
    expect(find.text('Aktifkan notifikasi'), findsOneWidget);
    expect(find.text('Nanti saja'), findsOneWidget);
    expect(h.push.enableCalls, 0);
    expect(h.container.read(notificationPromptAnsweredProvider), isFalse);
  });

  testWidgets('"Aktifkan notifikasi" asks, confirms and records the answer', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester);
    await _toNotificationPage(tester);
    await _tapEnable(tester);
    expect(h.push.enableCalls, 1);
    expect(find.text('Notifikasi aktif. Terima kasih!'), findsOneWidget);
    expect(
      find.byKey(const Key('onboarding-notification-enable')),
      findsNothing,
    );
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isTrue);
    await tester.tap(find.byKey(const Key('onboarding-next')));
    await tester.pumpAndSettle();
    expect(find.text('Home'), findsOneWidget);
    expect(h.completed, ['done']);
  });

  testWidgets('denied permission never blocks onboarding', (tester) async {
    final h = await _pumpOnboarding(tester, result: PushToggleResult.denied);
    await _toNotificationPage(tester);
    await _tapEnable(tester);
    expect(find.textContaining('Izin belum diberikan'), findsOneWidget);
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isTrue);
    await tester.tap(find.byKey(const Key('onboarding-next')));
    await tester.pumpAndSettle();
    expect(find.text('Home'), findsOneWidget);
  });

  testWidgets('a failed attempt is not an answer and can be retried', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester, result: PushToggleResult.failed);
    await _toNotificationPage(tester);
    await _tapEnable(tester);
    expect(find.textContaining('belum dapat diaktifkan'), findsOneWidget);
    expect(
      find.byKey(const Key('onboarding-notification-enable')),
      findsOneWidget,
    );
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isNull);
  });

  testWidgets('"Nanti saja" shows no dialog and is never asked again', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester);
    await _toNotificationPage(tester);
    await _tapSkip(tester);
    expect(h.push.enableCalls, 0);
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isTrue);
    expect(h.container.read(shouldAskNotificationsProvider), isFalse);
  });

  testWidgets('leaving the page without choosing counts as "Nanti saja"', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester);
    await _toNotificationPage(tester);
    await tester.tap(find.byKey(const Key('onboarding-next')));
    await tester.pumpAndSettle();
    expect(h.push.enableCalls, 0);
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isTrue);
    expect(find.text('Home'), findsOneWidget);
  });

  testWidgets('skipping the whole guide leaves the Beranda card to ask', (
    tester,
  ) async {
    final h = await _pumpOnboarding(tester);
    await tester.tap(find.byKey(const Key('onboarding-skip')));
    await tester.pumpAndSettle();
    expect(h.prefs.getBool('notification_prompt_answered_v1'), isNull);
    expect(find.text('Home'), findsOneWidget);
  });

  testWidgets('fits 320 px at 2x text in English on every state', (
    tester,
  ) async {
    await _pumpOnboarding(
      tester,
      locale: const Locale('en'),
      size: const Size(320, 568),
      textScale: 2,
    );
    await _toNotificationPage(tester);
    expect(find.text('Never miss important news'), findsOneWidget);
    expect(find.text('Turn on notifications'), findsOneWidget);
    expect(find.text('Maybe later'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await _tapEnable(tester);
    expect(find.text('Notifications are on. Thank you!'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  group('Beranda card', () {
    Future<(ProviderContainer, _FakePush, SharedPreferences)> pumpCard(
      WidgetTester tester, {
      PushToggleResult result = PushToggleResult.on,
      Map<String, Object> prefsValues = const {},
      bool available = true,
    }) async {
      SharedPreferences.setMockInitialValues(prefsValues);
      final prefs = await SharedPreferences.getInstance();
      late _FakePush push;
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(prefs),
          pushNotificationsProvider.overrideWith((ref) {
            return push = _FakePush(prefs, result: result);
          }),
          shouldAskNotificationsProvider.overrideWith(
            (ref) =>
                available &&
                !ref.watch(notificationPromptAnsweredProvider) &&
                !ref.watch(pushNotificationsProvider).enabled,
          ),
        ],
      );
      addTearDown(container.dispose);
      container.read(pushNotificationsProvider.notifier);
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: Scaffold(body: HomeNotificationCard()),
          ),
        ),
      );
      await tester.pumpAndSettle();
      return (container, push, prefs);
    }

    testWidgets('shows once for people who were never asked', (tester) async {
      await pumpCard(tester);
      expect(find.byKey(const Key('home-notification-card')), findsOneWidget);
    });

    testWidgets('hidden when already asked or unsupported', (tester) async {
      await pumpCard(
        tester,
        prefsValues: {'notification_prompt_answered_v1': true},
      );
      expect(find.byKey(const Key('home-notification-card')), findsNothing);
      await pumpCard(tester, available: false);
      expect(find.byKey(const Key('home-notification-card')), findsNothing);
    });

    testWidgets('"Aktifkan" asks and removes the card', (tester) async {
      final (_, push, prefs) = await pumpCard(tester);
      await tester.tap(find.byKey(const Key('home-notification-enable')));
      await tester.pumpAndSettle();
      expect(push.enableCalls, 1);
      expect(prefs.getBool('notification_prompt_answered_v1'), isTrue);
      expect(find.byKey(const Key('home-notification-card')), findsNothing);
    });

    testWidgets('"Nanti saja" removes the card without asking', (tester) async {
      final (_, push, prefs) = await pumpCard(tester);
      await tester.tap(find.byKey(const Key('home-notification-skip')));
      await tester.pumpAndSettle();
      expect(push.enableCalls, 0);
      expect(prefs.getBool('notification_prompt_answered_v1'), isTrue);
      expect(find.byKey(const Key('home-notification-card')), findsNothing);
    });

    testWidgets('denied permission explains and still closes the card', (
      tester,
    ) async {
      await pumpCard(tester, result: PushToggleResult.denied);
      await tester.tap(find.byKey(const Key('home-notification-enable')));
      await tester.pump();
      await tester.pump();
      expect(find.textContaining('Izin belum diberikan'), findsOneWidget);
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('home-notification-card')), findsNothing);
    });
  });
}
