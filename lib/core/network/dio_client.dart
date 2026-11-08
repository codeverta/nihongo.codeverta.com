import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/auth/session_expiration.dart';
import 'package:nihongo_master/core/diagnostics/error_reporter.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/device_timezone.dart';
import 'package:nihongo_master/core/network/api_base_url.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_cache_interceptor.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';

final apiBaseUrlProvider = Provider<String>((ref) {
  return resolveApiBaseUrl(configuredApiBaseUrl, defaultTargetPlatform);
});

final dioProvider = Provider<Dio>((ref) {
  final access = ref.watch(offlineAccessStoreProvider);
  final apiBaseUrl = ref.watch(apiBaseUrlProvider);
  final cache = ref.watch(learningCacheProvider.future);
  var disposed = false;

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
      onRequest: (options, handler) async {
        options.headers['Accept-Language'] = ref
            .read(appLanguageProvider)
            .languageCode;
        if (options.path.endsWith('/auth/google')) {
          final timezone = await ref.read(
            deviceTimeZoneIdentifierProvider.future,
          );
          if (timezone != null) {
            options.headers['X-Device-Timezone'] = timezone;
          }
        }
        final expectedAccount =
            options.extra['offlineAccountId'] ??
            access.userId ??
            guestCacheAccountId;
        options.extra['offlineAccountId'] = expectedAccount;
        if (expectedAccount != (access.userId ?? guestCacheAccountId)) {
          handler.reject(
            DioException(
              requestOptions: options,
              type: DioExceptionType.cancel,
              error: 'Akun sudah berubah.',
            ),
          );
          return;
        }
        String? token;
        try {
          // Secure storage can stall on some devices; never hang a request.
          token = await access.readSessionToken().timeout(
            const Duration(seconds: 8),
          );
        } on Object catch (error) {
          handler.reject(DioException(requestOptions: options, error: error));
          return;
        }
        if (expectedAccount != (access.userId ?? guestCacheAccountId)) {
          handler.reject(
            DioException(
              requestOptions: options,
              type: DioExceptionType.cancel,
              error: 'Akun sudah berubah.',
            ),
          );
          return;
        }
        if (token != null && token.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $token';
        } else {
          options.headers.removeWhere(
            (key, _) => key.toLowerCase() == 'authorization',
          );
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        final metadata = response.extra['offlineCache'];
        if (!disposed && (metadata is! Map || metadata['cached'] != true)) {
          ref.read(backendReachableProvider.notifier).state = true;
        }
        handler.next(response);
      },
      onError: (error, handler) async {
        final statusCode = error.response?.statusCode;
        final requestToken = _bearerToken(error.requestOptions);
        String? storedToken;
        try {
          storedToken = await access.readSessionToken();
        } on Object {
          // A Keychain outage does not prove a stored login was revoked.
        }
        if (statusCode == 401 &&
            requestToken != null &&
            requestToken == storedToken &&
            !_isAuthenticationSubmission(error.requestOptions.path)) {
          await access.clearSessionToken();
          final response = error.response;
          if (response != null) {
            response.data = {
              'error': {
                'code': 'SESSION_EXPIRED',
                'message': sessionExpiredMessage,
              },
            };
          }
          if (!disposed) {
            ref.read(sessionExpirationEventProvider.notifier).state++;
          }
        }
        if (error.type != DioExceptionType.cancel &&
            (statusCode == null || statusCode >= 500)) {
          if (!disposed && statusCode == null) {
            ref.read(backendReachableProvider.notifier).state = false;
          }
          ErrorReporter.captureNetworkFailure(error);
        }
        handler.next(error);
      },
    ),
  );

  dio.interceptors.add(
    LearningCacheInterceptor(
      dio: dio,
      cache: cache,
      currentUserId: () async => access.userId,
      canReadPremium: access.canReadPremium,
      onMetadata: (metadata) {
        if (!disposed) {
          ref.read(cachedResponseMetadataProvider.notifier).state = metadata;
        }
      },
    ),
  );
  // Retry an idempotent GET once on a transient failure so a brief network
  // hiccup does not leave a lesson on an error screen.
  dio.interceptors.add(
    InterceptorsWrapper(
      onError: (error, handler) async {
        final options = error.requestOptions;
        final transient =
            error.type == DioExceptionType.connectionTimeout ||
            error.type == DioExceptionType.sendTimeout ||
            error.type == DioExceptionType.receiveTimeout ||
            error.type == DioExceptionType.connectionError ||
            error.error is TimeoutException;
        if (disposed ||
            !transient ||
            options.method.toUpperCase() != 'GET' ||
            options.extra['networkRetried'] == true ||
            (options.cancelToken?.isCancelled ?? false)) {
          handler.next(error);
          return;
        }
        options.extra['networkRetried'] = true;
        try {
          handler.resolve(await dio.fetch<dynamic>(options));
        } on DioException catch (retryError) {
          handler.next(retryError);
        }
      },
    ),
  );
  ref.onDispose(() {
    disposed = true;
    dio.close(force: true);
  });

  if (kDebugMode) {
    dio.interceptors.add(
      LogInterceptor(requestHeader: false, responseHeader: false),
    );
  }

  return dio;
});

String? _bearerToken(RequestOptions request) {
  for (final entry in request.headers.entries) {
    if (entry.key.toLowerCase() == 'authorization' && entry.value is String) {
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
