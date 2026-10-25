import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/auth/session_expiration.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

enum AuthenticationStatus { checking, guest, authenticated }

@immutable
class AuthenticatedUser {
  const AuthenticatedUser({
    required this.id,
    required this.displayName,
    required this.email,
    this.photoUrl,
    this.isPremium = false,
    this.isLifetime = false,
    this.isOfflinePremium = false,
  });

  factory AuthenticatedUser.fromJson(Map<String, dynamic> json) {
    return AuthenticatedUser(
      id: json['id'] as String,
      displayName: json['displayName'] as String,
      email: json['email'] as String,
      photoUrl: json['photoUrl'] as String?,
      isPremium:
          (json['entitlement'] as Map<String, dynamic>?)?['tier'] == 'premium',
      isLifetime:
          (json['entitlement'] as Map<String, dynamic>?)?['isLifetime'] == true,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'displayName': displayName,
    'email': email,
    if (photoUrl != null) 'photoUrl': photoUrl,
    'entitlement': {
      'tier': isPremium ? 'premium' : 'free',
      'isLifetime': isLifetime,
    },
  };

  AuthenticatedUser withOfflineAccess({required bool allowed}) =>
      AuthenticatedUser(
        id: id,
        displayName: displayName,
        email: email,
        photoUrl: photoUrl,
        isPremium: allowed,
        isLifetime: allowed,
        isOfflinePremium: allowed,
      );

  final String id;
  final String displayName;
  final String email;
  final String? photoUrl;
  final bool isPremium;
  final bool isLifetime;
  final bool isOfflinePremium;
}

@immutable
class AuthenticationState {
  const AuthenticationState({
    required this.status,
    this.user,
    this.errorMessage,
  });

  const AuthenticationState.checking()
    : this(status: AuthenticationStatus.checking);

  const AuthenticationState.guest({String? errorMessage})
    : this(status: AuthenticationStatus.guest, errorMessage: errorMessage);

  const AuthenticationState.authenticated(AuthenticatedUser user)
    : this(status: AuthenticationStatus.authenticated, user: user);

  final AuthenticationStatus status;
  final AuthenticatedUser? user;
  final String? errorMessage;

  bool get isAuthenticated => status == AuthenticationStatus.authenticated;
  bool get isLoading => status == AuthenticationStatus.checking;
}

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthenticationState>((ref) {
      final controller = AuthController(
        AuthRepository(
          ref.watch(dioProvider),
          ref.watch(sharedPreferencesProvider),
          offlineAccess: ref.watch(offlineAccessStoreProvider),
        ),
        GoogleAuthenticationService(),
      );
      ref.listen<int>(sessionExpirationEventProvider, (previous, next) {
        if (next > (previous ?? 0)) {
          controller.expireSession();
        }
      });
      unawaited(controller.restoreSession());
      return controller;
    });

class AuthController extends StateNotifier<AuthenticationState> {
  AuthController(this._repository, this._googleService)
    : super(const AuthenticationState.checking());

  @visibleForTesting
  AuthController.test(super.state) : _repository = null, _googleService = null;

  final AuthRepository? _repository;
  final GoogleAuthenticationService? _googleService;
  bool _showingCachedOfflineAccount = false;

  Future<void> restoreSession() async {
    final repository = _repository;
    if (repository == null) return;
    final cachedUser = repository.cachedUser;
    if (!repository.hasStoredSession) {
      if (cachedUser == null) {
        state = const AuthenticationState.guest();
      } else {
        _showingCachedOfflineAccount = true;
        state = AuthenticationState.authenticated(
          await repository.offlineCachedUser(cachedUser),
        );
      }
      return;
    }
    if (cachedUser != null) {
      _showingCachedOfflineAccount = true;
      state = AuthenticationState.authenticated(
        await repository.offlineCachedUser(cachedUser),
      );
    }
    try {
      final user = await repository.currentUser();
      _showingCachedOfflineAccount = user.isOfflinePremium;
      state = AuthenticationState.authenticated(user);
    } on AuthException catch (error) {
      if (error.statusCode == 401) {
        await repository.clearSession();
        state = const AuthenticationState.guest(
          errorMessage: sessionExpiredMessage,
        );
      } else if (cachedUser == null) {
        state = AuthenticationState.guest(errorMessage: error.message);
      }
    } on Object {
      if (cachedUser == null) state = const AuthenticationState.guest();
    }
  }

  Future<bool> loginWithEmail(String email, String password) async {
    final repository = _repository;
    if (repository == null) return false;
    state = const AuthenticationState.checking();
    try {
      final user = await repository.loginWithEmail(email, password);
      _showingCachedOfflineAccount = false;
      state = AuthenticationState.authenticated(user);
      return true;
    } on AuthException catch (error) {
      state = AuthenticationState.guest(errorMessage: error.message);
      return false;
    }
  }

  Future<bool> loginWithGoogle() async {
    final repository = _repository;
    final googleService = _googleService;
    if (repository == null || googleService == null) return false;
    state = const AuthenticationState.checking();
    try {
      final idToken = await googleService.authenticate();
      final user = await repository.loginWithGoogle(idToken);
      _showingCachedOfflineAccount = false;
      state = AuthenticationState.authenticated(user);
      return true;
    } on AuthException catch (error) {
      state = AuthenticationState.guest(errorMessage: error.message);
      return false;
    } on GoogleSignInException catch (error) {
      final message = error.code == GoogleSignInExceptionCode.canceled
          ? 'Login Google dibatalkan.'
          : 'Google Sign-In gagal. Periksa konfigurasi OAuth.';
      state = AuthenticationState.guest(errorMessage: message);
      return false;
    }
  }

  Future<void> logout() async {
    final repository = _repository;
    final googleService = _googleService;
    LogoutSessionSnapshot? logoutSnapshot;
    var privateDataCleared = false;
    state = const AuthenticationState.checking();
    try {
      if (repository != null) {
        // The store gates the old account synchronously before this awaits the
        // secure token read, so in-flight cache responses cannot be attributed
        // to the account after logout begins.
        logoutSnapshot = await repository.beginLogout();
        await repository.clearPrivateData();
        privateDataCleared = true;
        await repository.logout(logoutSnapshot);
      }
    } finally {
      try {
        if (repository != null && !privateDataCleared) {
          await repository.clearPrivateData();
        }
        await googleService?.signOut();
      } on Object {
        // Logout lokal tidak boleh gagal karena provider/platform error.
      }
      state = const AuthenticationState.guest();
    }
  }

  void expireSession() {
    if (!mounted) return;
    unawaited(_repository?.clearSession());
    state = const AuthenticationState.guest(
      errorMessage: sessionExpiredMessage,
    );
  }

  void continueAsGuest() {
    if (!mounted || state.errorMessage != sessionExpiredMessage) return;
    state = const AuthenticationState.guest();
  }

  Future<void> refreshUser() async {
    final id = state.user?.id;
    if (_repository == null || id == null) return;
    try {
      final user = await _repository.currentUser();
      if (mounted && state.user?.id == id) {
        _showingCachedOfflineAccount = user.isOfflinePremium;
        state = AuthenticationState.authenticated(user);
      }
    } on AuthException catch (error) {
      if (error.statusCode == 401 && mounted && state.user?.id == id) {
        await _repository.clearSession();
        state = const AuthenticationState.guest();
      }
    } on Object {
      // Temporary outages do not destroy the login session. The server enforces access.
    }
  }

  /// Rechecks a cached offline lease on resume so the in-memory premium flag
  /// expires with its signed proof. Online server snapshots are left intact.
  Future<void> refreshOfflineEntitlement() async {
    final repository = _repository;
    final currentUser = state.user;
    if (repository == null ||
        currentUser == null ||
        !_showingCachedOfflineAccount) {
      return;
    }
    final updatedUser = await repository.offlineCachedUser(currentUser);
    if (mounted &&
        _showingCachedOfflineAccount &&
        state.user?.id == currentUser.id) {
      state = AuthenticationState.authenticated(updatedUser);
      _showingCachedOfflineAccount = updatedUser.isOfflinePremium;
    }
  }

  Future<void> uploadProfilePhoto(Uint8List compressedPhoto) async {
    final repository = _repository;
    final currentUser = state.user;
    if (repository == null || currentUser == null) {
      throw const AuthException('Masuk ke akun untuk mengubah foto profil.');
    }
    final updatedUser = await repository.uploadProfilePhoto(compressedPhoto);
    if (mounted && state.user?.id == currentUser.id) {
      state = AuthenticationState.authenticated(updatedUser);
    }
  }

  Future<void> deleteAccount() async {
    final repository = _repository;
    if (repository == null) return;
    await repository.deleteAccount(userId: state.user?.id);
    await logout();
  }
}

class AuthRepository {
  AuthRepository(
    this._dio,
    this._preferences, {
    OfflineAccessStore? offlineAccess,
  }) : _offlineAccess = offlineAccess;

  final Dio _dio;
  final SharedPreferences _preferences;
  final OfflineAccessStore? _offlineAccess;

  bool get hasStoredSession =>
      _offlineAccess?.hasPersistedSession ??
      (_preferences.getString(authTokenStorageKey)?.isNotEmpty ?? false);

  AuthenticatedUser? get cachedUser {
    final raw =
        _offlineAccess?.cachedProfile ?? _cachedProfileFromPreferences();
    if (raw == null) return null;
    try {
      return AuthenticatedUser.fromJson(raw);
    } on Object {
      return null;
    }
  }

  Future<AuthenticatedUser> offlineCachedUser(AuthenticatedUser user) async {
    final allowed = await _offlineAccess?.canReadPremium(user.id) ?? false;
    return user.withOfflineAccess(allowed: allowed);
  }

  Future<AuthenticatedUser> loginWithEmail(
    String email,
    String password,
  ) async {
    return _authenticate('/auth/login', {
      'email': email.trim(),
      'password': password,
    });
  }

  Future<AuthenticatedUser> loginWithGoogle(String idToken) async {
    return _authenticate('/auth/google', {'idToken': idToken});
  }

  Future<AuthenticatedUser> currentUser() async {
    try {
      await _offlineAccess?.readSessionToken();
      final response = await _dio.get<Map<String, dynamic>>('/auth/me');
      final user = AuthenticatedUser.fromJson(_data(response.data));
      await _offlineAccess?.cacheProfile(user.toJson());
      // /auth/me is not a purchase verification. Keep an unexpired, signed
      // lease until a successful billing response confirms free/refunded state.
      if (await _offlineAccess?.canReadPremium(user.id) ?? false) {
        return user.withOfflineAccess(allowed: true);
      }
      return user;
    } on DioException catch (error) {
      throw AuthException(
        _messageFromDio(error),
        statusCode: error.response?.statusCode,
      );
    }
  }

  Future<AuthenticatedUser> uploadProfilePhoto(
    Uint8List compressedPhoto,
  ) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/me/profile/photo',
        data: FormData.fromMap({
          'photo': MultipartFile.fromBytes(
            compressedPhoto,
            filename: 'profile.jpg',
          ),
        }),
        options: Options(contentType: Headers.multipartFormDataContentType),
      );
      final user = AuthenticatedUser.fromJson(_data(response.data));
      await _offlineAccess?.cacheProfile(user.toJson());
      return user;
    } on DioException catch (error) {
      throw AuthException(
        _messageFromDio(error),
        statusCode: error.response?.statusCode,
      );
    }
  }

  Future<LogoutSessionSnapshot> beginLogout() async {
    final accountId = _offlineAccess?.userId ?? cachedUser?.id;
    if (_offlineAccess != null) {
      _offlineAccess.beginLocalLogout();
      final token = await _offlineAccess.captureSessionTokenForLogout();
      return LogoutSessionSnapshot(accountId: accountId, token: token);
    }
    return LogoutSessionSnapshot(
      accountId: accountId,
      token: _preferences.getString(authTokenStorageKey),
    );
  }

  Future<void> logout(LogoutSessionSnapshot snapshot) async {
    final token = snapshot.token;
    if (token == null || token.isEmpty) return;
    try {
      // This one request intentionally bypasses normal session/account
      // interceptors: local identity has already been removed. It uses the
      // token captured before clearing, so the server can still revoke it.
      final logoutDio = Dio(_dio.options)
        ..httpClientAdapter = _dio.httpClientAdapter;
      await logoutDio.post<void>(
        '/auth/logout',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
          extra: {
            if (snapshot.accountId != null) 'accountId': snapshot.accountId,
          },
        ),
      );
    } on DioException {
      // Local logout must still finish when the network is unavailable.
    }
  }

  Future<void> deleteAccount({String? userId}) async {
    await _dio.delete<void>('/me/account', data: {'confirmation': 'DELETE'});
    if (userId != null) {
      await _preferences.remove('dictionary_library_v1_$userId');
      await _preferences.remove('learning_history_v1_$userId');
      await _preferences.remove('ssw_kaigo_completed_v1_$userId');
      await _preferences.remove('ssw_building-cleaning_completed_v1_$userId');
      await _preferences.remove('ssw_agriculture_completed_v1_$userId');
      await _preferences.remove('ssw_fishery_completed_v1_$userId');
    }
    await clearPrivateData();
  }

  Future<void> clearSession() async {
    if (_offlineAccess != null) {
      await _offlineAccess.clearSessionToken();
    } else {
      await _preferences.remove(authTokenStorageKey);
    }
    await _offlineAccess?.clearCachedProfile();
  }

  Future<void> clearPrivateData() async {
    if (_offlineAccess != null) {
      await _offlineAccess.clearPrivateData();
    } else {
      await _preferences.remove(authTokenStorageKey);
      await _preferences.remove(cachedAuthProfileStorageKey);
    }
  }

  Future<AuthenticatedUser> _authenticate(
    String path,
    Map<String, dynamic> body,
  ) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(path, data: body);
      final data = _data(response.data);
      final token = data['token'] as String;
      final user = AuthenticatedUser.fromJson(
        data['user'] as Map<String, dynamic>,
      );
      final previousId = _offlineAccess?.userId;
      if (previousId != null && previousId != user.id) {
        await _offlineAccess!.clearPrivateData();
      }
      if (_offlineAccess != null) {
        await _offlineAccess.saveSessionToken(token);
      } else {
        await _preferences.setString(authTokenStorageKey, token);
      }
      await _offlineAccess?.cacheProfile(user.toJson());
      return user;
    } on DioException catch (error) {
      throw AuthException(_messageFromDio(error));
    }
  }

  Map<String, dynamic> _data(Map<String, dynamic>? response) {
    if (response == null || response['data'] is! Map<String, dynamic>) {
      throw const AuthException('Response autentikasi tidak valid.');
    }
    return response['data'] as Map<String, dynamic>;
  }

  String _messageFromDio(DioException error) {
    final response = error.response?.data;
    if (response is Map<String, dynamic>) {
      final details = response['error'];
      if (details is Map<String, dynamic> && details['message'] is String) {
        return details['message'] as String;
      }
    }
    return 'Tidak dapat terhubung ke server. Pastikan backend aktif.';
  }

  Map<String, dynamic>? _cachedProfileFromPreferences() {
    final raw = _preferences.getString(cachedAuthProfileStorageKey);
    if (raw == null || raw.isEmpty) return null;
    try {
      final profile = jsonDecode(raw);
      return profile is Map<String, dynamic> ? profile : null;
    } on Object {
      return null;
    }
  }
}

class GoogleAuthenticationService {
  static const _serverClientId = String.fromEnvironment(
    'GOOGLE_SERVER_CLIENT_ID',
  );

  bool _initialized = false;

  Future<String> authenticate() async {
    if (_serverClientId.isEmpty) {
      throw const AuthException('GOOGLE_SERVER_CLIENT_ID belum dikonfigurasi.');
    }
    final signIn = GoogleSignIn.instance;
    if (!_initialized) {
      await signIn.initialize(serverClientId: _serverClientId);
      _initialized = true;
    }
    if (!signIn.supportsAuthenticate()) {
      throw const AuthException(
        'Google Sign-In interaktif belum didukung pada platform ini.',
      );
    }
    final account = await signIn.authenticate();
    final idToken = account.authentication.idToken;
    if (idToken == null || idToken.isEmpty) {
      throw const AuthException('Google tidak mengembalikan ID token.');
    }
    return idToken;
  }

  Future<void> signOut() async {
    if (!_initialized && _serverClientId.isNotEmpty) {
      await GoogleSignIn.instance.initialize(serverClientId: _serverClientId);
      _initialized = true;
    }
    if (_initialized) await GoogleSignIn.instance.signOut();
  }
}

class AuthException implements Exception {
  const AuthException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;
}

@immutable
class LogoutSessionSnapshot {
  const LogoutSessionSnapshot({required this.accountId, required this.token});

  final String? accountId;
  final String? token;
}
