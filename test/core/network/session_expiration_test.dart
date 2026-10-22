import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
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

  test(
    'revoked bearer token is cleared and routes the user to login',
    () async {
      SharedPreferences.setMockInitialValues({
        authTokenStorageKey: 'expired-token',
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
      final dio = container.read(dioProvider)..httpClientAdapter = adapter;
      final authSubscription = container.listen(
        authControllerProvider,
        (_, _) {},
      );
      addTearDown(authSubscription.close);
      final router = container.read(appRouterProvider)
        ..go(AppRoutes.vocabularyLessonPath('n5-daily'));

      await _waitFor(
        () => container.read(authControllerProvider).isAuthenticated,
      );
      expect(container.read(authControllerProvider).isAuthenticated, isTrue);

      late DioException failure;
      try {
        await dio.get<Map<String, dynamic>>('/vocabulary/levels');
        fail('Expected the API request to be rejected.');
      } on DioException catch (error) {
        failure = error;
      }

      expect(adapter.authorizationHeaders.last, 'Bearer expired-token');
      expect(container.read(sessionExpirationEventProvider), 1);
      await _waitFor(() {
        return container.read(authControllerProvider).status ==
            AuthenticationStatus.guest;
      });
      await _waitFor(
        () => router.routeInformationProvider.value.uri.path ==
            AppRoutes.loginPath,
      );

      expect(preferences.getString(authTokenStorageKey), isNull);
      expect(
        container.read(authControllerProvider).errorMessage,
        sessionExpiredMessage,
      );
      final responseBody = failure.response!.data as Map<String, dynamic>;
      final errorDetails = responseBody['error'] as Map<String, dynamic>;
      expect(errorDetails['message'] as String, sessionExpiredMessage);
      expect(
        router.routeInformationProvider.value.uri.path,
        AppRoutes.loginPath,
      );

      container.read(authControllerProvider.notifier).continueAsGuest();
      router.go(AppRoutes.homePath);
      expect(
        router.routeInformationProvider.value.uri.path,
        AppRoutes.homePath,
      );
      expect(container.read(authControllerProvider).errorMessage, isNull);
    },
  );
}

Future<void> _waitFor(bool Function() condition) async {
  for (var attempt = 0; attempt < 50; attempt++) {
    if (condition()) return;
    await Future<void>.delayed(const Duration(milliseconds: 10));
  }
  fail('Condition did not become true in time.');
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
    if (options.path.endsWith('/auth/me')) {
      return ResponseBody.fromString(
        jsonEncode({
          'data': {
            'id': 'user-1',
            'displayName': 'Learner',
            'email': 'learner@example.com',
            'entitlement': {'tier': 'free', 'isLifetime': false},
          },
        }),
        200,
        headers: {Headers.contentTypeHeader: ['application/json']},
      );
    }

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
