import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/session_expiration.dart';
import 'package:nihongo_master/core/diagnostics/error_reporter.dart';
import 'package:nihongo_master/core/network/api_base_url.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';

final apiBaseUrlProvider = Provider<String>((ref) {
  return resolveApiBaseUrl(configuredApiBaseUrl, defaultTargetPlatform);
});

final dioProvider = Provider<Dio>((ref) {
  final preferences = ref.watch(sharedPreferencesProvider);
  final apiBaseUrl = ref.watch(apiBaseUrlProvider);

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
      onError: (error, handler) async {
        final statusCode = error.response?.statusCode;
        final requestToken = _bearerToken(error.requestOptions);
        final storedToken = preferences.getString(authTokenStorageKey);
        if (statusCode == 401 &&
            requestToken != null &&
            requestToken == storedToken &&
            !_isAuthenticationSubmission(error.requestOptions.path)) {
          await preferences.remove(authTokenStorageKey);
          final response = error.response;
          if (response != null) {
            response.data = {
              'error': {
                'code': 'SESSION_EXPIRED',
                'message': sessionExpiredMessage,
              },
            };
          }
          ref.read(sessionExpirationEventProvider.notifier).state++;
        }
        if (error.type != DioExceptionType.cancel &&
            (statusCode == null || statusCode >= 500)) {
          ErrorReporter.captureNetworkFailure(error);
        }
        handler.next(error);
      },
    ),
  );

  if (kDebugMode) {
    dio.interceptors.add(
      LogInterceptor(requestHeader: false, responseHeader: false),
    );
  }

  return dio;
});

String? _bearerToken(RequestOptions request) {
  for (final entry in request.headers.entries) {
    if (entry.key.toLowerCase() == 'authorization' &&
        entry.value is String) {
      final value = (entry.value as String).trim();
      if (value.toLowerCase().startsWith('bearer ')) {
        return value.substring(7).trim();
      }
    }
  }
  return null;
}

bool _isAuthenticationSubmission(String path) =>
    path.endsWith('/auth/login') || path.endsWith('/auth/google');
