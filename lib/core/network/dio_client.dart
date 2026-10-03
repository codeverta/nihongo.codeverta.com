import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';

final dioProvider = Provider<Dio>((ref) {
  final preferences = ref.watch(sharedPreferencesProvider);
  const configuredApiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080/v1',
  );
  final apiBaseUrl = _resolveApiBaseUrl(
    configuredApiBaseUrl,
    defaultTargetPlatform,
  );

  final dio = Dio(
    BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) {
        final token = preferences.getString(authTokenStorageKey);
        if (token != null && token.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
    ),
  );

  dio.interceptors.add(
    LogInterceptor(
      requestBody: true,
      responseBody: true,
    ),
  );

  return dio;
});

String _resolveApiBaseUrl(String configuredUrl, TargetPlatform platform) {
  final uri = Uri.tryParse(configuredUrl);
  if (uri == null || platform != TargetPlatform.android) return configuredUrl;

  final host = uri.host.toLowerCase();
  if (host != 'localhost' && host != '127.0.0.1' && host != '::1') {
    return configuredUrl;
  }

  // Inside Android Emulator, loopback belongs to the virtual device itself.
  // 10.0.2.2 is the emulator's route to the development host machine.
  return uri.replace(host: '10.0.2.2').toString();
}
