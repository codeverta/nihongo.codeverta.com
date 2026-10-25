import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:in_app_purchase/in_app_purchase.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:nihongo_master/features/billing/presentation/premium_screen.dart';

void main() {
  testWidgets(
    'store price wins and pending checkout stays disabled on a narrow screen',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (_) => AuthController.test(_signedIn),
            ),
            billingControllerProvider.overrideWith(
              (_) => _StaticBilling(
                BillingState(
                  pending: true,
                  ready: true,
                  priceIdr: 149000,
                  product: ProductDetails(
                    id: 'lifetime',
                    title: 'Lifetime',
                    description: '',
                    price: '¥1,000',
                    rawPrice: 1000,
                    currencyCode: 'JPY',
                  ),
                ),
              ),
            ),
            lifetimeDisplayPriceProvider.overrideWith((_) async => 99000),
          ],
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.5)),
              child: child!,
            ),
            home: const PremiumScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.byKey(const Key('buy-lifetime')),
        300,
        scrollable: find.byType(Scrollable).first,
      );
      expect(
        tester
            .widget<FilledButton>(find.byKey(const Key('buy-lifetime')))
            .onPressed,
        isNull,
      );
      expect(find.text('Menunggu pembayaran'), findsOneWidget);
      expect(find.text('¥1,000'), findsOneWidget);
      expect(find.text('Rp149.000'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets('verified premium returns to the originating job detail', (
    tester,
  ) async {
    final router = GoRouter(
      initialLocation: '/premium?from=%2Fjobs%2Fone',
      routes: [
        GoRoute(path: '/premium', builder: (_, __) => const PremiumScreen()),
        GoRoute(
          path: '/jobs/one',
          builder: (_, __) => const Scaffold(body: Text('Job destination')),
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (_) => AuthController.test(_signedIn),
          ),
          billingControllerProvider.overrideWith(
            (_) => _StaticBilling(const BillingState(premium: true)),
          ),
          lifetimeDisplayPriceProvider.overrideWith((_) async => 99000),
        ],
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Jelajahi akses saya'),
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.tap(find.text('Jelajahi akses saya'));
    await tester.pumpAndSettle();
    expect(find.text('Job destination'), findsOneWidget);
  });
  testWidgets('premium page shows the public backend price to guests', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          billingControllerProvider.overrideWith(
            (ref) => BillingController(
              repository: BillingRepository(Dio()),
              gateway: null,
              platform: '',
              onChanged: () async {},
            ),
          ),
          lifetimeDisplayPriceProvider.overrideWith((ref) async => 149000),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const PremiumScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.scrollUntilVisible(
      find.byKey(const Key('lifetime-price-panel')),
      250,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Rp149.000'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('English premium screen never shows an IDR-only price', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          billingControllerProvider.overrideWith(
            (_) => _StaticBilling(
              const BillingState(priceIdr: 149000, ready: false),
            ),
          ),
          lifetimeDisplayPriceProvider.overrideWith((ref) async => 149000),
        ],
        child: MaterialApp(
          locale: const Locale('en'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: const PremiumScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byKey(const Key('lifetime-price-panel')),
      250,
      scrollable: find.byType(Scrollable).first,
    );

    expect(find.text('Store price unavailable'), findsOneWidget);
    expect(find.text('Rp149.000'), findsNothing);
    expect(find.text('Lebih jauh dari sekadar belajar'), findsNothing);
    expect(tester.takeException(), isNull);
  });
}

const _signedIn = AuthenticationState.authenticated(
  AuthenticatedUser(
    id: 'buyer',
    displayName: 'Buyer',
    email: 'buyer@example.com',
  ),
);

class _StaticBilling extends BillingController {
  _StaticBilling(BillingState initial)
    : super(
        repository: BillingRepository(Dio()),
        gateway: null,
        platform: '',
        onChanged: () async {},
      ) {
    state = initial;
  }
}
