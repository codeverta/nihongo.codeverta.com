import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:nihongo_master/features/billing/presentation/lifetime_offer.dart';
import 'package:nihongo_master/features/billing/presentation/premium_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

AuthenticationState _user({
  bool lifetime = false,
  bool canPurchase = true,
  String id = 'u1',
}) => AuthenticationState.authenticated(
  AuthenticatedUser(
    id: id,
    displayName: 'U',
    email: 'u@example.com',
    isLifetime: lifetime,
    isPremium: lifetime,
    canPurchase: canPurchase,
  ),
);

class _Billing extends BillingController {
  _Billing()
    : super(
        repository: BillingRepository(Dio()),
        gateway: null,
        platform: '',
        onChanged: () async {},
      ) {
    state = const BillingState(ready: true);
  }

  void grant() => state = const BillingState(premium: true);
}

Future<ProviderContainer> _container({
  AuthenticationState? auth,
  LifetimeOfferConfig config = const LifetimeOfferConfig(),
  Map<String, Object> prefs = const {},
}) async {
  SharedPreferences.setMockInitialValues(prefs);
  final preferences = await SharedPreferences.getInstance();
  final container = ProviderContainer(
    overrides: [
      authControllerProvider.overrideWith(
        (_) => AuthController.test(auth ?? _user()),
      ),
      sharedPreferencesProvider.overrideWithValue(preferences),
      lifetimeOfferConfigProvider.overrideWith((_) async => config),
      analyticsServiceProvider.overrideWithValue(AnalyticsService(preferences)),
    ],
  );
  addTearDown(container.dispose);
  return container;
}

void main() {
  group('after login', () {
    test(
      'a free account sees the offer, which continues to the login target',
      () async {
        final container = await _container();
        final route = await container
            .read(lifetimeOfferProvider)
            .routeAfterLogin('/learn/grammar?level=N3');
        final uri = Uri.parse(route);
        expect(uri.path, '/premium');
        expect(uri.queryParameters['offer'], 'after_login');
        expect(uri.queryParameters['from'], '/learn/grammar?level=N3');
        expect(
          container
              .read(sharedPreferencesProvider)
              .getInt('lifetime_offer_last_shown_v1_u1'),
          isNotNull,
        );
      },
    );

    test(
      'Lifetime, demo/reviewer and a switched-off offer go straight on',
      () async {
        for (final container in [
          await _container(auth: _user(lifetime: true)),
          await _container(auth: _user(canPurchase: false)),
          await _container(
            config: const LifetimeOfferConfig(afterLogin: false),
          ),
        ]) {
          expect(
            await container
                .read(lifetimeOfferProvider)
                .routeAfterLogin('/news/1'),
            '/news/1',
          );
        }
      },
    );

    test('unsafe or looping targets fall back to Beranda', () {
      for (final from in [
        null,
        'https://evil.test',
        '//evil',
        '/premium',
        '/login?from=/x',
      ]) {
        expect(lifetimeOfferTarget(from), '/home');
      }
      expect(lifetimeOfferTarget('/exams/n3'), '/exams/n3');
    });
  });

  group('premium screen as the offer', () {
    Future<_Billing> pump(
      WidgetTester tester, {
      LifetimeOfferSource offer = LifetimeOfferSource.afterLogin,
      bool lifetime = false,
    }) async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final billing = _Billing();
      final router = GoRouter(
        initialLocation: offer == LifetimeOfferSource.periodic
            ? '/home'
            : '/premium',
        routes: [
          GoRoute(path: '/home', builder: (_, _) => const Text('Beranda')),
          GoRoute(path: '/learn/x', builder: (_, _) => const Text('Pelajaran')),
          GoRoute(
            path: '/premium',
            builder: (_, _) =>
                PremiumScreen(offer: offer, offerFrom: '/learn/x'),
          ),
        ],
      );
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (_) => AuthController.test(_user(lifetime: lifetime)),
            ),
            billingControllerProvider.overrideWith((_) => billing),
            lifetimeDisplayPriceProvider.overrideWith((_) async => 99000),
            analyticsServiceProvider.overrideWithValue(
              AnalyticsService(preferences),
            ),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      if (offer == LifetimeOfferSource.periodic) {
        router.push('/premium');
      }
      await tester.pumpAndSettle();
      return billing;
    }

    testWidgets('"Lanjut gratis" continues to the login target', (
      tester,
    ) async {
      await pump(tester);
      await tester.scrollUntilVisible(
        find.byKey(const Key('lifetime-offer-skip')),
        200,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(find.byKey(const Key('lifetime-offer-skip')));
      await tester.pumpAndSettle();
      expect(find.text('Pelajaran'), findsOneWidget);
    });

    testWidgets('the close button and Android back skip too', (tester) async {
      await pump(tester);
      await tester.tap(find.byKey(const Key('lifetime-offer-close')));
      await tester.pumpAndSettle();
      expect(find.text('Pelajaran'), findsOneWidget);

      await pump(tester, offer: LifetimeOfferSource.periodic);
      expect(find.byType(PremiumScreen), findsOneWidget);
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(find.text('Beranda'), findsOneWidget);
    });

    testWidgets('a purchase while the offer is open continues on', (
      tester,
    ) async {
      final billing = await pump(tester);
      billing.grant();
      await tester.pumpAndSettle();
      expect(find.text('Pelajaran'), findsOneWidget);
    });

    testWidgets('an account already Lifetime is not held on the offer', (
      tester,
    ) async {
      await pump(tester, lifetime: true);
      expect(find.text('Pelajaran'), findsOneWidget);
    });

    testWidgets('the normal premium page has no skip', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (_) => AuthController.test(_user()),
            ),
            billingControllerProvider.overrideWith((_) => _Billing()),
            lifetimeDisplayPriceProvider.overrideWith((_) async => 99000),
          ],
          child: const MaterialApp(home: PremiumScreen()),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('lifetime-offer-close')), findsNothing);
      expect(find.byKey(const Key('lifetime-offer-skip')), findsNothing);
    });
  });

  group('on app launch', () {
    Future<GoRouter> pumpHome(
      WidgetTester tester,
      ProviderContainer container,
    ) async {
      container.read(lifetimeOfferProvider).launchDelay = Duration.zero;
      final router = GoRouter(
        initialLocation: '/home',
        routes: [
          GoRoute(
            path: '/home',
            builder: (_, _) =>
                const LifetimeOfferLaunchCheck(child: Text('Beranda')),
          ),
          GoRoute(path: '/premium', builder: (_, _) => const Text('Offer')),
        ],
      );
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      return router;
    }

    testWidgets('shows once the interval has passed since it was last shown', (
      tester,
    ) async {
      final eleven = DateTime.now()
          .subtract(const Duration(days: 11))
          .millisecondsSinceEpoch;
      final container = await _container(
        prefs: {'lifetime_offer_last_shown_v1_u1': eleven},
      );
      await pumpHome(tester, container);
      expect(find.text('Offer'), findsOneWidget);
    });

    testWidgets('not within the interval, and not with interval 0', (
      tester,
    ) async {
      final three = DateTime.now()
          .subtract(const Duration(days: 3))
          .millisecondsSinceEpoch;
      await pumpHome(
        tester,
        await _container(prefs: {'lifetime_offer_last_shown_v1_u1': three}),
      );
      expect(find.text('Offer'), findsNothing);

      await pumpHome(
        tester,
        await _container(config: const LifetimeOfferConfig(intervalDays: 0)),
      );
      expect(find.text('Offer'), findsNothing);
    });

    testWidgets('only on launch: Beranda appearing again does not reopen it', (
      tester,
    ) async {
      final container = await _container(
        config: const LifetimeOfferConfig(intervalDays: 1),
      );
      await pumpHome(tester, container);
      expect(find.text('Offer'), findsOneWidget);
      await container
          .read(sharedPreferencesProvider)
          .remove('lifetime_offer_last_shown_v1_u1');
      await pumpHome(tester, container);
      expect(find.text('Offer'), findsNothing);
    });
  });
}
