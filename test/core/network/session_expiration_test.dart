import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/auth/session_expiration.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_cache_persistence_memory.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test(
    'revoked bearer token is cleared and its raw message is hidden',
    () async {
      SharedPreferences.setMockInitialValues({'is_first_time': false});
      final preferences = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          offlineAccessStoreProvider.overrideWithValue(
            OfflineAccessStore(
              preferences,
              secretStorage: _TestSecretStorage(),
            ),
          ),
          learningCacheProvider.overrideWith(
            (ref) => PersistentLearningCache.open(
              persistence: MemoryLearningCachePersistence(),
            ),
          ),
          apiBaseUrlProvider.overrideWithValue('https://api.example.com/v1'),
        ],
      );
      final adapter = _RevokedSessionAdapter();
      final dio = container.read(dioProvider)..httpClientAdapter = adapter;
      final authSubscription = container.listen(
        authControllerProvider,
        (_, _) {},
      );
      addTearDown(() {
        authSubscription.close();
        container.dispose();
        dio.close(force: true);
      });

      await preferences.setString(authTokenStorageKey, 'expired-token');
      DioException? failure;
      try {
        await dio.get<dynamic>('/vocabulary/lessons/n5-daily');
      } on DioException catch (error) {
        failure = error;
      }

      expect(failure, isNotNull);
      final responseData = failure!.response?.data as Map<String, dynamic>;
      final responseError = responseData['error'] as Map<String, dynamic>;
      expect(preferences.getString(authTokenStorageKey), isNull);
      expect(adapter.authorizationHeaders, contains('Bearer expired-token'));
      expect(responseError['message'], sessionExpiredMessage);
      expect(
        responseError['message'],
        isNot(contains('Bearer token tidak valid')),
      );
      expect(
        container.read(authControllerProvider).errorMessage,
        sessionExpiredMessage,
      );
    },
  );

  test('expired sessions redirect any protected page to login', () {
    expect(
      redirectToLoginIfSessionExpired(
        '/learn/vocabulary/lessons/n5-daily',
        sessionExpiredMessage,
      ),
      '/login',
    );
    expect(
      redirectToLoginIfSessionExpired('/login', sessionExpiredMessage),
      isNull,
    );
    expect(redirectToLoginIfSessionExpired('/home', 'Login gagal.'), isNull);
  });
}

class _TestSecretStorage implements OfflineSecretStorage {
  final values = <String, String>{};
  @override
  Future<String?> read(String key) async => values[key];
  @override
  Future<void> write({required String key, required String value}) async {
    values[key] = value;
  }

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }
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
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
