import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test(
    'logout hides local account before network and sends captured revoke token',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final offlineAccess = OfflineAccessStore(
        preferences,
        secretStorage: _MemorySecrets(),
        verifier: _AcceptTestProof(),
      );
      await offlineAccess.saveSessionToken('session-to-revoke');
      await offlineAccess.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      });
      final adapter = _PausedLogoutAdapter();
      final dio = Dio(BaseOptions(baseUrl: 'https://api.example.com/v1'))
        ..httpClientAdapter = adapter;
      final controller = AuthController(
        AuthRepository(dio, preferences, offlineAccess: offlineAccess),
        null,
      );
      addTearDown(() {
        controller.dispose();
        dio.close(force: true);
      });

      final logout = controller.logout();
      await adapter.requestStarted.future;

      expect(offlineAccess.userId, isNull);
      expect(offlineAccess.cachedProfile, isNull);
      await offlineAccess.cacheProfile({
        'id': 'account-1',
        'displayName': 'Late response',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      });
      expect(offlineAccess.cachedProfile, isNull);
      expect(await offlineAccess.readSessionToken(), isNull);
      expect(adapter.path, '/auth/logout');
      expect(adapter.authorization, 'Bearer session-to-revoke');

      adapter.finish();
      await logout;
      expect(controller.state.status, AuthenticationStatus.guest);
    },
  );

  test(
    'cached account restores offline when no session token can be read',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final offlineAccess = OfflineAccessStore(
        preferences,
        secretStorage: _MemorySecrets(),
        verifier: _AcceptTestProof(),
      );
      await offlineAccess.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      });
      await offlineAccess.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'server-signed-test-proof',
      );
      final adapter = _CountingUnavailableAdapter();
      final dio = Dio(BaseOptions(baseUrl: 'https://api.example.com/v1'))
        ..httpClientAdapter = adapter;
      final controller = AuthController(
        AuthRepository(dio, preferences, offlineAccess: offlineAccess),
        null,
      );
      addTearDown(() {
        controller.dispose();
        dio.close(force: true);
      });

      await controller.restoreSession();

      expect(controller.state.user?.id, 'account-1');
      expect(controller.state.user?.isPremium, isFalse);
      expect(adapter.requests, 0);
    },
  );

  test(
    'offline restore keeps cached account and token after server 5xx',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      await preferences.setString(authTokenStorageKey, 'valid-session');
      final secrets = _MemorySecrets();
      final verifier = _AcceptTestProof();
      final offlineAccess = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: verifier,
      );
      await offlineAccess.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      });
      await offlineAccess.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'server-signed-test-proof',
      );

      final dio = Dio(BaseOptions(baseUrl: 'https://api.example.com/v1'))
        ..httpClientAdapter = _StatusAdapter(503);
      final repository = AuthRepository(
        dio,
        preferences,
        offlineAccess: offlineAccess,
      );
      final controller = AuthController(repository, null);
      addTearDown(() {
        controller.dispose();
        dio.close(force: true);
      });

      await controller.restoreSession();

      expect(controller.state.isAuthenticated, isTrue);
      expect(controller.state.user?.id, 'account-1');
      expect(controller.state.user?.isPremium, isTrue);
      expect(preferences.getString(authTokenStorageKey), isNull);
      expect(await offlineAccess.readSessionToken(), 'valid-session');

      verifier.isValid = false;
      await controller.refreshOfflineEntitlement();
      expect(controller.state.user?.isPremium, isFalse);
    },
  );

  test(
    'free auth profile does not revoke a still-valid signed billing lease',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final verifier = _AcceptTestProof();
      final offlineAccess = OfflineAccessStore(
        preferences,
        secretStorage: _MemorySecrets(),
        verifier: verifier,
      );
      await offlineAccess.saveSessionToken('valid-session');
      await offlineAccess.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'premium', 'isLifetime': true},
      });
      await offlineAccess.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'server-signed-test-proof',
      );
      final dio = Dio(BaseOptions(baseUrl: 'https://api.example.com/v1'))
        ..httpClientAdapter = _ProfileAdapter();
      final repository = AuthRepository(
        dio,
        preferences,
        offlineAccess: offlineAccess,
      );
      addTearDown(() => dio.close(force: true));

      final user = await repository.currentUser();

      expect(user.isPremium, isTrue);
      expect(await offlineAccess.canReadPremium('account-1'), isTrue);
    },
  );
}

class _AcceptTestProof implements OfflineEntitlementVerifier {
  bool isValid = true;

  @override
  Future<bool> verify({
    required String proof,
    required String userId,
    required DateTime now,
  }) async =>
      isValid && proof == 'server-signed-test-proof' && userId == 'account-1';
}

class _MemorySecrets implements OfflineSecretStorage {
  final Map<String, String> values = {};

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }

  @override
  Future<String?> read(String key) async => values[key];

  @override
  Future<void> write({required String key, required String value}) async {
    values[key] = value;
  }
}

class _StatusAdapter implements HttpClientAdapter {
  _StatusAdapter(this.statusCode);

  final int statusCode;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async => ResponseBody.fromString(
    '{}',
    statusCode,
    headers: {
      Headers.contentTypeHeader: ['application/json'],
    },
  );

  @override
  void close({bool force = false}) {}
}

class _ProfileAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async => ResponseBody.fromString(
    jsonEncode({
      'data': {
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      },
    }),
    200,
    headers: {
      Headers.contentTypeHeader: ['application/json'],
    },
  );

  @override
  void close({bool force = false}) {}
}

class _CountingUnavailableAdapter implements HttpClientAdapter {
  int requests = 0;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests++;
    return ResponseBody.fromString(
      '{}',
      503,
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

class _PausedLogoutAdapter implements HttpClientAdapter {
  final Completer<void> requestStarted = Completer<void>();
  final Completer<void> _release = Completer<void>();
  String? path;
  String? authorization;

  void finish() {
    if (!_release.isCompleted) _release.complete();
  }

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    path = options.path;
    authorization = options.headers['Authorization'] as String?;
    requestStarted.complete();
    await _release.future;
    return ResponseBody.fromString(
      '{}',
      200,
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
