import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/app_update/app_update_controller.dart';
import 'package:nihongo_master/core/app_update/app_update_gate.dart';
import 'package:nihongo_master/core/app_update/app_update_policy.dart';

const android = AppDeviceInfo(
  platform: 'android',
  buildNumber: 3,
  version: '1.0.2',
  osVersion: '12',
  androidApi: 31,
);
const oldAndroid = AppDeviceInfo(
  platform: 'android',
  buildNumber: 3,
  version: '1.0.2',
  osVersion: '8',
  androidApi: 26,
);

Map<String, dynamic> policy({
  int minimum = 4,
  int api = 28,
  String legacy = 'allow',
  bool enabled = true,
  String ios = '',
}) =>
    {
      'enabled': enabled,
      'minimumBuild': minimum,
      'latestBuild': 4,
      'latestVersion': '1.0.3',
      'minimumAndroidApi': api,
      'minimumIOSVersion': ios,
      'storeUrl':
          'https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster',
      'unsupportedDevicePolicy': legacy,
      'message': '',
    };

void main() {
  testWidgets('router back dispatch cannot navigate behind a mandatory update',
      (tester) async {
    var json = policy();
    final controller = AppUpdateController(
      deviceInfo: () async => android,
      fetchPolicy: () async => {'android': json},
    );
    final router = GoRouter(
      initialLocation: '/details',
      routes: [
        GoRoute(
          path: '/',
          builder: (_, state) => const Text('Beranda'),
          routes: [
            GoRoute(
              path: 'details',
              builder: (_, state) => const Text('Detail'),
            ),
          ],
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          appUpdateControllerProvider.overrideWith((ref) => controller),
        ],
        child: MaterialApp.router(
          routeInformationProvider: router.routeInformationProvider,
          routeInformationParser: router.routeInformationParser,
          routerDelegate: router.routerDelegate,
          backButtonDispatcher:
              AppUpdateBackDispatcher(() => controller.state.interceptBack),
          builder: (_, child) => AppUpdateGate(child: child!),
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();
    expect(router.routeInformationProvider.value.uri.path, '/details');
    expect(find.text('Update sekarang'), findsOneWidget);
    json = policy(minimum: 0);
    await controller.check();
    await tester.pumpAndSettle();
    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();
    expect(router.routeInformationProvider.value.uri.path, '/');
    await tester.pumpWidget(const SizedBox());
  });
  test(
      'supported legacy builds remain usable when only newer releases require a newer OS',
      () {
    final p = AppUpdatePolicy.fromJson(policy(minimum: 3), 'android');
    expect(p.evaluate(oldAndroid), AppUpdateRequirement.supported);
  });

  test(
      'incompatible devices use explicit allow/block policy rather than a store redirect',
      () {
    expect(
      AppUpdatePolicy.fromJson(policy(), 'android').evaluate(oldAndroid),
      AppUpdateRequirement.unsupportedAllowed,
    );
    expect(
      AppUpdatePolicy.fromJson(policy(legacy: 'block'), 'android')
          .evaluate(oldAndroid),
      AppUpdateRequirement.unsupportedBlocked,
    );
    expect(
      AppUpdatePolicy.fromJson(policy(), 'android').evaluate(android),
      AppUpdateRequirement.update,
    );
  });

  test(
      'iOS version comparison uses numeric components and disabled iOS never blocks',
      () {
    final json = policy(ios: '15.10')
      ..['storeUrl'] = 'https://apps.apple.com/app/id123456789';
    const ios = AppDeviceInfo(
      platform: 'ios',
      buildNumber: 3,
      version: '1.0.2',
      osVersion: '15.9',
    );
    expect(
      AppUpdatePolicy.fromJson(json, 'ios').evaluate(ios),
      AppUpdateRequirement.unsupportedAllowed,
    );
    json['minimumIOSVersion'] = '15.9.0';
    expect(
      AppUpdatePolicy.fromJson(json, 'ios').evaluate(ios),
      AppUpdateRequirement.update,
    );
    json['enabled'] = false;
    json['storeUrl'] = '';
    expect(
      AppUpdatePolicy.fromJson(json, 'ios').evaluate(ios),
      AppUpdateRequirement.supported,
    );
  });

  test('invalid floors and non-store links are rejected', () {
    expect(
      () => AppUpdatePolicy.fromJson(policy(minimum: 5), 'android'),
      throwsFormatException,
    );
    expect(
      () => AppUpdatePolicy.fromJson(
        policy()..['storeUrl'] = 'https://example.com',
        'android',
      ),
      throwsFormatException,
    );
  });

  test(
      'startup outage allows access; later confirmed requirements survive outages and can be rolled back',
      () async {
    var failed = true;
    var json = policy();
    final controller = AppUpdateController(
      deviceInfo: () async => android,
      fetchPolicy: () async {
        if (failed) throw StateError('offline');
        return {'android': json};
      },
    );
    addTearDown(controller.dispose);
    await controller.check();
    expect(controller.state.initialized, isTrue);
    expect(controller.state.showDialog, isFalse);
    failed = false;
    await controller.check();
    expect(controller.state.showDialog, isTrue);
    controller.continueOnLegacyDevice();
    expect(controller.state.showDialog, isTrue);
    failed = true;
    await controller.check();
    expect(controller.state.showDialog, isTrue);
    expect(controller.state.error, isNotNull);
    failed = false;
    json = policy(minimum: 0);
    await controller.check();
    expect(controller.state.showDialog, isFalse);
  });

  test(
      'legacy advisory is dismissed for this policy but a changed retirement policy blocks',
      () async {
    var json = policy();
    final controller = AppUpdateController(
      deviceInfo: () async => oldAndroid,
      fetchPolicy: () async => {'android': json},
    );
    addTearDown(controller.dispose);
    await controller.check();
    controller.continueOnLegacyDevice();
    await controller.check();
    expect(controller.state.showDialog, isFalse);
    json = policy(legacy: 'block');
    await controller.check();
    expect(controller.state.showDialog, isTrue);
    expect(
      controller.state.requirement,
      AppUpdateRequirement.unsupportedBlocked,
    );
  });

  testWidgets(
      'mandatory dialog blocks underlying taps and system back; policy rollback removes it',
      (tester) async {
    var json = policy();
    var taps = 0;
    final controller = AppUpdateController(
      deviceInfo: () async => android,
      fetchPolicy: () async => {'android': json},
    );
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          appUpdateControllerProvider.overrideWith((ref) => controller),
        ],
        child: MaterialApp(
          home: AppUpdateGate(
            child: TextButton(
              onPressed: () => taps++,
              child: const Text('Beranda'),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Update sekarang'), findsOneWidget);
    final dispatcher =
        AppUpdateBackDispatcher(() => controller.state.interceptBack);
    expect(await dispatcher.didPopRoute(), isTrue);
    await tester.tapAt(const Offset(10, 10));
    await tester.pump();
    expect(taps, 0);
    expect(find.text('Update sekarang'), findsOneWidget);
    json = policy(minimum: 0);
    await tester.tap(find.text('Periksa lagi'));
    await tester.pumpAndSettle();
    expect(find.text('Update sekarang'), findsNothing);
    await tester.tap(find.text('Beranda'));
    expect(taps, 1);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets(
      'old OS offers continuing supported legacy app, not an impossible store update',
      (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          appUpdateControllerProvider.overrideWith(
            (ref) => AppUpdateController(
              deviceInfo: () async => oldAndroid,
              fetchPolicy: () async => {'android': policy()},
            ),
          ),
        ],
        child: const MaterialApp(home: AppUpdateGate(child: Text('Beranda'))),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Update sekarang'), findsNothing);
    expect(find.text('Lanjutkan versi ini'), findsOneWidget);
    await tester.tap(find.text('Lanjutkan versi ini'));
    await tester.pumpAndSettle();
    expect(find.text('Lanjutkan versi ini'), findsNothing);
    await tester.pumpWidget(const SizedBox());
  });
}
