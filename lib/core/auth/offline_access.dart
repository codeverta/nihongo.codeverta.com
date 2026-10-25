import 'dart:convert';

import 'package:cryptography/cryptography.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const cachedAuthProfileStorageKey = 'auth_cached_profile_v1';
const offlineLeasePublicKey = String.fromEnvironment(
  'OFFLINE_ACCESS_PUBLIC_KEY',
);

const _offlineLeaseStorageKey = 'offline_access_lease_v1';
const _offlineLastSeenStorageKey = 'offline_access_last_seen_v1';
const _offlineIdentityStorageKey = 'offline_access_identity_v1';
const _offlineRevokedPreferenceKey = 'offline_access_revoked_v1';
const _sessionTokenSecureStorageKey = 'auth_session_token_v1';
const _sessionPresentStorageKey = 'auth_session_present_v1';
const _sessionRevokedStorageKey = 'auth_session_revoked_v1';

final offlineAccessStoreProvider = Provider<OfflineAccessStore>((ref) {
  return OfflineAccessStore(ref.watch(sharedPreferencesProvider));
});

/// Secure key/value storage used for signed entitlement proofs and clock state.
/// Tests can provide an in-memory implementation.
abstract interface class OfflineSecretStorage {
  Future<String?> read(String key);
  Future<void> write({required String key, required String value});
  Future<void> delete(String key);
}

class FlutterSecureOfflineSecretStorage implements OfflineSecretStorage {
  FlutterSecureOfflineSecretStorage([FlutterSecureStorage? storage])
    : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  @override
  Future<String?> read(String key) => _storage.read(key: key);

  @override
  Future<void> write({required String key, required String value}) =>
      _storage.write(key: key, value: value);

  @override
  Future<void> delete(String key) => _storage.delete(key: key);
}

/// Verifies an account-bound, Ed25519-signed offline lifetime entitlement.
// Keep this as an interface so callers can inject a deterministic verifier.
// ignore: one_member_abstracts
abstract interface class OfflineEntitlementVerifier {
  Future<bool> verify({
    required String proof,
    required String userId,
    required DateTime now,
  });
}

class SignedOfflineEntitlementVerifier implements OfflineEntitlementVerifier {
  SignedOfflineEntitlementVerifier(this.encodedPublicKey);

  final String encodedPublicKey;
  final Ed25519 _algorithm = Ed25519();

  @override
  Future<bool> verify({
    required String proof,
    required String userId,
    required DateTime now,
  }) async {
    if (encodedPublicKey.isEmpty || userId.isEmpty) return false;
    try {
      final parts = proof.split('.');
      if (parts.length != 3) return false;
      final headerBytes = _decodeBase64Url(parts[0]);
      final claimsBytes = _decodeBase64Url(parts[1]);
      final signatureBytes = _decodeBase64Url(parts[2]);
      final header = jsonDecode(utf8.decode(headerBytes));
      final claims = jsonDecode(utf8.decode(claimsBytes));
      if (header is! Map<String, dynamic> || claims is! Map<String, dynamic>) {
        return false;
      }
      if (header['alg'] != 'EdDSA' ||
          header['typ'] != 'JWT' ||
          header['kid'] != 'offline-access-v1' ||
          claims['iss'] != 'nihongo-master' ||
          claims['aud'] != 'nihongo-master-offline-premium' ||
          claims['sub'] != userId ||
          claims['tier'] != 'premium' ||
          claims['isLifetime'] != true ||
          claims['jti'] is! String ||
          (claims['jti'] as String).isEmpty) {
        return false;
      }

      final issuedAt = claims['iat'];
      final expiresAt = claims['exp'];
      if (issuedAt is! int || expiresAt is! int) return false;
      final issued = DateTime.fromMillisecondsSinceEpoch(
        issuedAt * 1000,
        isUtc: true,
      );
      final expiry = DateTime.fromMillisecondsSinceEpoch(
        expiresAt * 1000,
        isUtc: true,
      );
      final current = now.toUtc();
      if (issued.isAfter(current.add(const Duration(minutes: 5))) ||
          !expiry.isAfter(current) ||
          !expiry.isAfter(issued) ||
          expiry.difference(issued) > const Duration(days: 30)) {
        return false;
      }

      final publicBytes = _decodeBase64(encodedPublicKey);
      if (publicBytes.length != 32) return false;
      final publicKey = SimplePublicKey(publicBytes, type: KeyPairType.ed25519);
      return _algorithm.verify(
        utf8.encode('${parts[0]}.${parts[1]}'),
        signature: Signature(signatureBytes, publicKey: publicKey),
      );
    } on Object {
      return false;
    }
  }
}

/// Persists the cached account profile and validates signed offline access.
/// A cached `isPremium` flag is never used as authorization.
class OfflineAccessStore {
  OfflineAccessStore(
    this._preferences, {
    OfflineSecretStorage? secretStorage,
    OfflineEntitlementVerifier? verifier,
    DateTime Function()? clock,
  }) : _secretStorage = secretStorage ?? FlutterSecureOfflineSecretStorage(),
       _verifier =
           verifier ?? SignedOfflineEntitlementVerifier(offlineLeasePublicKey),
       _clock = clock ?? DateTime.now;

  final SharedPreferences _preferences;
  final OfflineSecretStorage _secretStorage;
  final OfflineEntitlementVerifier _verifier;
  final DateTime Function() _clock;
  bool _localLogoutPending = false;
  String? _logoutAccountId;
  String? _logoutLegacyToken;
  int _authorizationGeneration = 0;
  bool _sessionRevokedInMemory = false;
  final Set<String> _revokedAccountsInMemory = {};

  String _accountKey(String prefix, String accountId) {
    final encoded = base64Url
        .encode(utf8.encode(accountId))
        .replaceAll('=', '');
    return '${prefix}_$encoded';
  }

  /// True when a secure token or a legacy token exists, without reading secure
  /// storage synchronously. Legacy tokens migrate on [readSessionToken].
  bool get hasPersistedSession =>
      !_localLogoutPending &&
      ((_preferences.getBool(_sessionPresentStorageKey) ?? false) ||
          (_preferences.getString(authTokenStorageKey)?.isNotEmpty ?? false));

  /// Reads the secure token and upgrades a legacy SharedPreferences token on
  /// first use. The legacy value is removed only after secure storage accepts it.
  Future<String?> readSessionToken() async {
    if (_localLogoutPending) return null;
    if (!hasPersistedSession) return null;
    if (_preferences.getBool(_sessionRevokedStorageKey) ?? false) return null;
    final generation = _authorizationGeneration;
    final secureToken = await _secretStorage.read(
      _sessionTokenSecureStorageKey,
    );
    if (!_sessionReadCurrent(generation)) return null;
    if (secureToken != null && secureToken.isNotEmpty) {
      await _preferences.setBool(_sessionPresentStorageKey, true);
      if (!_sessionReadCurrent(generation)) return null;
      return secureToken;
    }
    final legacyToken = _preferences.getString(authTokenStorageKey);
    if (legacyToken == null ||
        legacyToken.isEmpty ||
        !_sessionReadCurrent(generation)) {
      return null;
    }
    await saveSessionToken(legacyToken);
    return !_localLogoutPending && !_sessionRevokedInMemory
        ? legacyToken
        : null;
  }

  /// Stores a session token in platform secure storage and clears its legacy copy.
  Future<void> saveSessionToken(String token) async {
    if (token.isEmpty) {
      await clearSessionToken();
      return;
    }
    final generation = ++_authorizationGeneration;
    _sessionRevokedInMemory = true;
    await _secretStorage.write(
      key: _sessionTokenSecureStorageKey,
      value: token,
    );
    if (generation != _authorizationGeneration) return;
    await _preferences.setBool(_sessionPresentStorageKey, true);
    if (generation != _authorizationGeneration) return;
    await _preferences.setBool(_sessionRevokedStorageKey, false);
    if (generation != _authorizationGeneration) return;
    await _preferences.remove(authTokenStorageKey);
    if (generation != _authorizationGeneration) return;
    _localLogoutPending = false;
    _sessionRevokedInMemory = false;
    _logoutAccountId = null;
    _logoutLegacyToken = null;
  }

  /// Hides the active account synchronously before logout waits on storage or
  /// the network. Captured values are kept only in memory for server revocation.
  void beginLocalLogout() {
    if (_localLogoutPending) return;
    final rawProfile = _readCachedProfile();
    final id = rawProfile?['id'];
    _logoutAccountId = id is String && id.isNotEmpty ? id : null;
    if (_logoutAccountId != null) {
      _revokedAccountsInMemory.add(_logoutAccountId!);
    }
    _logoutLegacyToken = _preferences.getString(authTokenStorageKey);
    _localLogoutPending = true;
    _sessionRevokedInMemory = true;
    _authorizationGeneration++;
  }

  /// Reads the pre-logout token despite the local logout gate. This exists only
  /// so `/auth/logout` can revoke the captured server session after local data
  /// has already been hidden and deleted.
  Future<String?> captureSessionTokenForLogout() async {
    beginLocalLogout();
    try {
      final secureToken = await _secretStorage.read(
        _sessionTokenSecureStorageKey,
      );
      if (secureToken != null && secureToken.isNotEmpty) return secureToken;
    } on Object {
      // Fall back to a legacy token captured before local removal.
    }
    final legacyToken = _logoutLegacyToken;
    return legacyToken != null && legacyToken.isNotEmpty ? legacyToken : null;
  }

  /// Revokes the local session marker before deleting the secure token so a
  /// failed keychain deletion cannot resurrect a stale session.
  Future<void> clearSessionToken() async {
    _sessionRevokedInMemory = true;
    _authorizationGeneration++;
    await _preferences.setBool(_sessionRevokedStorageKey, true);
    await _preferences.setBool(_sessionPresentStorageKey, false);
    await _preferences.remove(authTokenStorageKey);
    try {
      await _secretStorage.delete(_sessionTokenSecureStorageKey);
    } on Object {
      // The revocation marker above keeps a retained keychain value unusable.
    }
  }

  String? get userId {
    if (_localLogoutPending) return null;
    final profile = _readCachedProfile();
    final id = profile?['id'];
    return id is String && id.isNotEmpty ? id : null;
  }

  /// Returns the signed lease expiry only while the current account's proof is
  /// still valid. Secure storage and the pinned verifier remain authoritative.
  Future<DateTime?> get offlineLeaseExpiresAt async {
    final activeUserId = userId;
    if (activeUserId == null || !await canReadPremium(activeUserId)) {
      return null;
    }
    try {
      final proof = await _secretStorage.read(
        _accountKey(_offlineLeaseStorageKey, activeUserId),
      );
      if (proof == null) return null;
      final parts = proof.split('.');
      if (parts.length != 3) return null;
      final claims = jsonDecode(utf8.decode(_decodeBase64Url(parts[1])));
      if (claims is! Map<String, dynamic> || claims['exp'] is! int) {
        return null;
      }
      return DateTime.fromMillisecondsSinceEpoch(
        (claims['exp'] as int) * 1000,
        isUtc: true,
      );
    } on Object {
      return null;
    }
  }

  Map<String, dynamic>? get cachedProfile {
    if (_localLogoutPending) return null;
    return _readCachedProfile();
  }

  Map<String, dynamic>? _readCachedProfile() {
    final raw = _preferences.getString(cachedAuthProfileStorageKey);
    if (raw == null || raw.isEmpty) return null;
    try {
      final value = jsonDecode(raw);
      return value is Map<String, dynamic> ? value : null;
    } on Object {
      return null;
    }
  }

  Future<void> cacheProfile(Map<String, dynamic> profile) async {
    if (_localLogoutPending) return;
    final id = profile['id'];
    if (id is! String || id.isEmpty) return;
    final currentId = userId;
    if (currentId != null && currentId != id) return;
    final safeProfile = <String, dynamic>{
      'id': id,
      'displayName': profile['displayName'] as String? ?? '',
      'email': profile['email'] as String? ?? '',
      if (profile['photoUrl'] is String) 'photoUrl': profile['photoUrl'],
      'entitlement': _safeEntitlement(),
    };
    await _preferences.setString(
      cachedAuthProfileStorageKey,
      jsonEncode(safeProfile),
    );
    try {
      await _secretStorage.write(
        key: _accountKey(_offlineIdentityStorageKey, id),
        value: id,
      );
    } on Object {
      // Online account use remains available; offline access fails closed.
    }
  }

  /// Applies only a successful authoritative response from `GET /billing` or
  /// `POST /billing/verify`. A free result revokes a cached lease immediately.
  Future<void> applyServerSnapshot({
    required String userId,
    required bool premium,
    required bool lifetime,
    String? signedLease,
  }) async {
    if (!_hasCurrentSession(userId)) return;
    final generation = ++_authorizationGeneration;
    final revokedKey = _accountKey(_offlineRevokedPreferenceKey, userId);
    if (!premium || !lifetime) {
      _revokedAccountsInMemory.add(userId);
      await _preferences.setBool(revokedKey, true);
      if (!_isSnapshotCurrent(userId, generation)) return;
      await _secretStorage.delete(_accountKey(_offlineLeaseStorageKey, userId));
      if (!_isSnapshotCurrent(userId, generation)) return;
      await _secretStorage.delete(
        _accountKey(_offlineLastSeenStorageKey, userId),
      );
      return;
    }
    final identityKey = _accountKey(_offlineIdentityStorageKey, userId);
    if (await _secretStorage.read(identityKey) != userId) return;
    if (!_isSnapshotCurrent(userId, generation)) return;
    final proof = signedLease;
    if (proof == null || proof.isEmpty) return;
    final now = _clock().toUtc();
    if (!await _verifier.verify(proof: proof, userId: userId, now: now)) {
      return;
    }
    if (!_isSnapshotCurrent(userId, generation) ||
        await _secretStorage.read(identityKey) != userId ||
        !_isSnapshotCurrent(userId, generation)) {
      return;
    }
    await _secretStorage.write(
      key: _accountKey(_offlineLeaseStorageKey, userId),
      value: proof,
    );
    if (!_isSnapshotCurrent(userId, generation)) {
      await _discardStaleLeaseIfRevoked(userId, proof);
      return;
    }
    await _recordLatestTime(userId, now, generation: generation);
    if (!_isSnapshotCurrent(userId, generation)) return;
    await _preferences.setBool(revokedKey, false);
    if (!_isSnapshotCurrent(userId, generation)) {
      // A stale billing operation may finish its preference write after a
      // logout or refund started. Restore the revocation marker if a newer
      // authoritative operation has not already revalidated premium.
      if (_revokedAccountsInMemory.contains(userId)) {
        await _preferences.setBool(revokedKey, true);
      }
      return;
    }
    _revokedAccountsInMemory.remove(userId);
  }

  Future<void> _discardStaleLeaseIfRevoked(
    String userId,
    String staleProof,
  ) async {
    if (!_revokedAccountsInMemory.contains(userId)) return;
    try {
      final leaseKey = _accountKey(_offlineLeaseStorageKey, userId);
      if (await _secretStorage.read(leaseKey) == staleProof &&
          _revokedAccountsInMemory.contains(userId)) {
        await _secretStorage.delete(leaseKey);
      }
    } on Object {
      // The revoked marker still gates access if best-effort cleanup fails.
    }
  }

  Future<bool> canReadPremium(String userId) async {
    if (!_canReadPremiumNow(userId)) return false;
    final generation = _authorizationGeneration;
    try {
      final identityKey = _accountKey(_offlineIdentityStorageKey, userId);
      if (await _secretStorage.read(identityKey) != userId) {
        return false;
      }
      if (!_canReadPremiumAtGeneration(userId, generation)) return false;
      final now = _clock().toUtc();
      final lastSeenKey = _accountKey(_offlineLastSeenStorageKey, userId);
      final lastSeenRaw = await _secretStorage.read(lastSeenKey);
      if (lastSeenRaw != null) {
        final lastSeen = DateTime.tryParse(lastSeenRaw)?.toUtc();
        if (lastSeen == null ||
            now.isBefore(lastSeen.subtract(const Duration(minutes: 5)))) {
          return false;
        }
      }
      if (!_canReadPremiumAtGeneration(userId, generation)) return false;
      final proof = await _secretStorage.read(
        _accountKey(_offlineLeaseStorageKey, userId),
      );
      if (proof == null || proof.isEmpty) return false;
      final valid = await _verifier.verify(
        proof: proof,
        userId: userId,
        now: now,
      );
      if (!valid || !_canReadPremiumAtGeneration(userId, generation)) {
        return false;
      }
      // Re-read the binding after signature verification. Logout, refund, or
      // an account switch may have happened while secure storage was pending.
      if (await _secretStorage.read(identityKey) != userId ||
          !_canReadPremiumAtGeneration(userId, generation)) {
        return false;
      }
      if (lastSeenRaw == null || now.isAfter(DateTime.parse(lastSeenRaw))) {
        if (!_canReadPremiumAtGeneration(userId, generation)) return false;
        await _secretStorage.write(
          key: lastSeenKey,
          value: now.toIso8601String(),
        );
      }
      return _canReadPremiumAtGeneration(userId, generation);
    } on Object {
      return false;
    }
  }

  /// Logout clears account-scoped local data while leaving public content alone.
  Future<void> clearPrivateData() async {
    beginLocalLogout();
    final activeUserId = _logoutAccountId;
    if (activeUserId != null) {
      await _preferences.setBool(
        _accountKey(_offlineRevokedPreferenceKey, activeUserId),
        true,
      );
    }
    await clearSessionToken();
    await _preferences.remove(cachedAuthProfileStorageKey);
    if (activeUserId != null) {
      final suffix = '_$activeUserId';
      final privateKeys = _preferences.getKeys().where((key) {
        if (!key.endsWith(suffix)) return false;
        return key.startsWith('dictionary_library_v1_') ||
            key.startsWith('learning_history_v1_') ||
            key.startsWith('anki_reviews_v1_') ||
            (key.startsWith('ssw_') && key.contains('_completed_v1_'));
      }).toList();
      for (final key in privateKeys) {
        await _preferences.remove(key);
      }
    }
    try {
      if (activeUserId != null) {
        await _deleteLease(activeUserId);
        await _secretStorage.delete(
          _accountKey(_offlineIdentityStorageKey, activeUserId),
        );
      }
      // Clean the pre-account-keyed layout if an app was upgraded mid-session.
      await _secretStorage.delete(_offlineLeaseStorageKey);
      await _secretStorage.delete(_offlineLastSeenStorageKey);
      await _secretStorage.delete(_offlineIdentityStorageKey);
    } on Object {
      // The revocation marker above prevents a retained secure entry being used.
    }
  }

  Future<void> clearCachedProfile() => _clearCachedProfileAndInvalidate();

  Future<void> _clearCachedProfileAndInvalidate() async {
    _authorizationGeneration++;
    await _preferences.remove(cachedAuthProfileStorageKey);
  }

  Future<void> _recordLatestTime(
    String userId,
    DateTime value, {
    required int generation,
  }) async {
    final key = _accountKey(_offlineLastSeenStorageKey, userId);
    final raw = await _secretStorage.read(key);
    if (!_isSnapshotCurrent(userId, generation)) return;
    final lastSeen = raw == null ? null : DateTime.tryParse(raw)?.toUtc();
    if (lastSeen == null || value.isAfter(lastSeen)) {
      await _secretStorage.write(key: key, value: value.toIso8601String());
    }
  }

  bool _hasCurrentSession(String userId) =>
      userId.isNotEmpty &&
      !_localLogoutPending &&
      !_sessionRevokedInMemory &&
      this.userId == userId &&
      hasPersistedSession &&
      !(_preferences.getBool(_sessionRevokedStorageKey) ?? false);

  bool _sessionReadCurrent(int generation) =>
      generation == _authorizationGeneration &&
      !_localLogoutPending &&
      !_sessionRevokedInMemory &&
      hasPersistedSession &&
      !(_preferences.getBool(_sessionRevokedStorageKey) ?? false);

  bool _isSnapshotCurrent(String userId, int generation) =>
      generation == _authorizationGeneration && _hasCurrentSession(userId);

  bool _canReadPremiumNow(String userId) =>
      _hasCurrentSession(userId) &&
      !_revokedAccountsInMemory.contains(userId) &&
      !(_preferences.getBool(
            _accountKey(_offlineRevokedPreferenceKey, userId),
          ) ??
          false);

  bool _canReadPremiumAtGeneration(String userId, int generation) =>
      generation == _authorizationGeneration && _canReadPremiumNow(userId);

  Future<void> _deleteLease(String userId) async {
    await _secretStorage.delete(_accountKey(_offlineLeaseStorageKey, userId));
    await _secretStorage.delete(
      _accountKey(_offlineLastSeenStorageKey, userId),
    );
  }
}

// The profile cache never persists entitlement flags. Restoration adds premium
// only after the account-bound secure lease has been verified.
Map<String, dynamic> _safeEntitlement() => const {
  'tier': 'free',
  'isLifetime': false,
};

List<int> _decodeBase64Url(String value) => _decodeBase64(value);

List<int> _decodeBase64(String value) {
  final normalized = value.replaceAll('-', '+').replaceAll('_', '/');
  final padded = normalized.padRight(
    normalized.length + (4 - normalized.length % 4) % 4,
    '=',
  );
  return base64.decode(padded);
}
