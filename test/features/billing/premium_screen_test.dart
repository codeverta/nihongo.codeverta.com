import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:nihongo_master/features/billing/presentation/premium_screen.dart';

void main() {
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

    expect(find.text('Rp149.000'), findsOneWidget);
  });
}
