import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:nihongo_master/core/network/api_base_url.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Sends bounded, sanitized diagnostics without user credentials or request data.
abstract final class ErrorReporter {
  static Dio? _client;
  static int _pendingReports = 0;
  static SharedPreferences? _preferences;
  static Future<void> _queueTail = Future<void>.value();
  static const _pendingQueueKey = 'pending_error_diagnostics_v1';
  static const _maxQueuedReports = 20;

  static void initialize(SharedPreferences preferences) {
    _preferences = preferences;
    unawaited(flushPending());
  }

  static Future<void> flushPending() async {
    final pending = await _readQueue();
    for (final encoded in pending) {
      try {
        final decoded = jsonDecode(encoded);
        if (decoded is! Map<String, dynamic>) {
          await _removeFromQueue(encoded);
          continue;
        }
        await _clientFor().post<void>('diagnostics/errors', data: decoded);
        await _removeFromQueue(encoded);
      } catch (_) {
        // Keep unsent reports for the next launch if the API is still offline.
        return;
      }
    }
  }

  static Future<void> capture(
    Object error,
    StackTrace stackTrace, {
    String severity = 'error',
  }) async {
    await _send(
      severity: severity,
      errorType: error.runtimeType.toString(),
      message: _sanitize(error.toString()),
      stackTrace: _sanitize(stackTrace.toString()),
      route: '',
    );
  }

  static void captureNetworkFailure(DioException error) {
    final statusCode = error.response?.statusCode;
    final message = statusCode == null
        ? 'Network request failed (${error.type.name})'
        : 'HTTP $statusCode response';
    final stackTrace = error.stackTrace.toString();
    unawaited(
      _send(
        severity: 'error',
        errorType: 'network/${error.type.name}',
        message: message,
        stackTrace: _sanitize(stackTrace),
        route: error.requestOptions.uri.path,
      ),
    );
  }

  static Future<void> _send({
    required String severity,
    required String errorType,
    required String message,
    required String stackTrace,
    required String route,
  }) async {
    if (_pendingReports >= 5 || message.trim().isEmpty) return;
    _pendingReports++;
    try {
      final payload = <String, dynamic>{
        'severity': severity,
        'errorType': _truncate(errorType, 160),
        'message': _truncate(message, 2000),
        'stackTrace': _truncate(stackTrace, 10000),
        'platform': kIsWeb ? 'web' : defaultTargetPlatform.name,
        'appVersion': const String.fromEnvironment(
          'APP_VERSION',
          defaultValue: '1.0.5+6',
        ),
        'route': _truncate(route, 512),
      };
      final encoded = jsonEncode(payload);
      await _mutateQueue((items) {
        items.add(encoded);
        if (items.length > _maxQueuedReports) {
          items.removeRange(0, items.length - _maxQueuedReports);
        }
      });
      try {
        await _clientFor().post<void>('diagnostics/errors', data: payload);
        await _removeFromQueue(encoded);
      } catch (_) {
        // Keep the report locally and retry it at the next app launch.
      }
    } catch (_) {
      // Diagnostics must never cause a second app error or delay normal flows.
    } finally {
      _pendingReports--;
    }
  }

  static Dio _clientFor() => _client ??= Dio(
    BaseOptions(
      baseUrl: resolveApiBaseUrl(configuredApiBaseUrl, defaultTargetPlatform),
      connectTimeout: const Duration(seconds: 3),
      receiveTimeout: const Duration(seconds: 3),
      sendTimeout: const Duration(seconds: 3),
      headers: const {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ),
  );

  static Future<List<String>> _readQueue() async {
    final previous = _queueTail;
    final next = Completer<void>();
    _queueTail = next.future;
    await previous;
    try {
      return _preferences?.getStringList(_pendingQueueKey) ?? <String>[];
    } finally {
      next.complete();
    }
  }

  static Future<void> _removeFromQueue(String encoded) =>
      _mutateQueue((items) => items.remove(encoded));

  static Future<void> _mutateQueue(void Function(List<String>) mutation) async {
    final previous = _queueTail;
    final next = Completer<void>();
    _queueTail = next.future;
    await previous;
    try {
      final preferences = _preferences;
      if (preferences == null) return;
      final items = preferences.getStringList(_pendingQueueKey) ?? <String>[];
      mutation(items);
      await preferences.setStringList(_pendingQueueKey, items);
    } finally {
      next.complete();
    }
  }

  static String _sanitize(String value) {
    return value
        .replaceAll(
          RegExp(r'[\w.+-]+@[\w.-]+\.[a-z]{2,}', caseSensitive: false),
          '[redacted-email]',
        )
        .replaceAll(
          RegExp(r'bearer\s+[a-z0-9._~+/-]+=*', caseSensitive: false),
          'Bearer [redacted]',
        )
        .replaceAll(
          RegExp(
            r'\beyJ[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\b',
          ),
          '[redacted-token]',
        )
        .replaceAll(
          RegExp(
            r'(token|password|secret|authorization)([=: ]+)[^\s&]+',
            caseSensitive: false,
          ),
          r'$1$2[redacted]',
        )
        .replaceAll(
          RegExp(r'https?://[^\s<>]+', caseSensitive: false),
          '[redacted-url]',
        );
  }

  static String _truncate(String value, int maxRunes) {
    final runes = value.runes;
    if (runes.length <= maxRunes) return value;
    return String.fromCharCodes(runes.take(maxRunes));
  }
}
