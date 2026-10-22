import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/session_expiration.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets(
    'revoked bearer token is cleared and routes the user to login',
    (tester) async {
      SharedPreferences.setMockInitialValues({
        'is_first_time': false,
      });
      final preferences = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          apiBaseUrlProvider.overrideWithValue('https://api.example.com/v1'),
        ],
      );
      addTearDown(container.dispose);

      final adapter = _RevokedSessionAdapter();
      container.read(dioProvider).httpClientAdapter = adapter;
      final authSubscription = container.listen(
        authControllerProvider,
        (_, _) {},
      );
      addTearDown(authSubscription.close);
      final router = container.read(appRouterProvider)
        ..go('/attributions');

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pump();
      expect(
        router.routerDelegate.currentConfiguration.uri.path,
        '/attributions',
      );
      expect(
        container.read(authControllerProvider).status,
        AuthenticationStatus.guest,
      );

      await preferences.setString(authTokenStorageKey, 'expired-token');
      try {
        await container
            .read(dioProvider)
            .get<dynamic>('/vocabulary/lessons/n5-daily');
      } on DioException {
        // The endpoint intentionally returns 401 for the revoked token.
      }
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));
      await tester.pump();

      expect(preferences.getString(authTokenStorageKey), isNull);
      expect(
        adapter.authorizationHeaders,
        contains('Bearer expired-token'),
      );
      expect(
        container.read(authControllerProvider).errorMessage,
        sessionExpiredMessage,
      );
      expect(find.text('Masuk untuk mulai belajar'), findsOneWidget);
      expect(find.text(sessionExpiredMessage), findsOneWidget);
      expect(find.textContaining('Bearer token tidak valid'), findsNothing);
      expect(
        router.routerDelegate.currentConfiguration.uri.path,
        AppRoutes.loginPath,
      );
    },
  );
}

class _RevokedSessionAdapter implements HttpClientAdapter {
  final authorizationHeaders = <String?>[];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    String? authorization;
    for (final entry in options.headers.entries) {
      if (entry.key.toLowerCase() == 'authorization') {
        authorization = entry.value as String?;
      }
    }
    authorizationHeaders.add(authorization);
    return ResponseBody.fromString(
      jsonEncode({
        'error': {
          'code': 'unauthorized',
          'message': 'Bearer token tidak valid. Akses dicabut.',
        },
      }),
      401,
      headers: {Headers.contentTypeHeader: ['application/json']},
    );
  }

  @override
  void close({bool force = false}) {}
}
