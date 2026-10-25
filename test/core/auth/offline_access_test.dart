import 'dart:async';
import 'dart:convert';

import 'package:cryptography/cryptography.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test(
    'signed lease is account-bound, expires, and refund revokes it',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final secretStorage = _MemorySecrets();
      final seed = List<int>.generate(32, (index) => index + 1);
      final algorithm = Ed25519();
      final keyPair = await algorithm.newKeyPairFromSeed(seed);
      final publicKey = await keyPair.extractPublicKey();
      final now = DateTime.utc(2026, 10, 23, 1);
      var clock = now;
      final store = OfflineAccessStore(
        preferences,
        secretStorage: secretStorage,
        verifier: SignedOfflineEntitlementVerifier(
          base64Encode(publicKey.bytes),
        ),
        clock: () => clock,
      );
      await store.saveSessionToken('session-token');
      await store.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
        'entitlement': {'tier': 'free', 'isLifetime': false},
      });
      final lease = await _signedLease(
        algorithm: algorithm,
        keyPair: keyPair,
        userId: 'account-1',
        issuedAt: now,
      );
      final leaseParts = lease.split('.');
      final signature = leaseParts[2];
      final changedSignature =
          '${signature.startsWith('A') ? 'B' : 'A'}${signature.substring(1)}';
      expect(
        await SignedOfflineEntitlementVerifier(
          base64Encode(publicKey.bytes),
        ).verify(
          proof: '${leaseParts[0]}.${leaseParts[1]}.$changedSignature',
          userId: 'account-1',
          now: now,
        ),
        isFalse,
      );

      await store.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: lease,
      );
      expect(await store.canReadPremium('account-1'), isTrue);
      expect(
        await store.offlineLeaseExpiresAt,
        now.add(const Duration(days: 30)),
      );
      expect(await store.canReadPremium('another-account'), isFalse);

      clock = now.add(const Duration(days: 30));
      expect(await store.canReadPremium('account-1'), isFalse);

      clock = now.add(const Duration(days: 1));
      await store.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: lease,
      );
      expect(await store.canReadPremium('account-1'), isTrue);
      await store.applyServerSnapshot(
        userId: 'account-1',
        premium: false,
        lifetime: false,
      );
      expect(await store.canReadPremium('account-1'), isFalse);
    },
  );

  test('missing pinned key and tampered proof fail closed', () async {
    final verifier = SignedOfflineEntitlementVerifier('');
    expect(
      await verifier.verify(
        proof: 'header.claims.signature',
        userId: 'account-1',
        now: DateTime.utc(2026, 10, 23),
      ),
      isFalse,
    );
  });

  test('legacy session token migrates to secure storage', () async {
    SharedPreferences.setMockInitialValues({
      authTokenStorageKey: 'legacy-session-token',
    });
    final preferences = await SharedPreferences.getInstance();
    final secrets = _MemorySecrets();
    final store = OfflineAccessStore(
      preferences,
      secretStorage: secrets,
      verifier: _AcceptTestProof(),
    );

    expect(await store.readSessionToken(), 'legacy-session-token');
    expect(secrets.values['auth_session_token_v1'], 'legacy-session-token');
    expect(preferences.getString(authTokenStorageKey), isNull);
    expect(store.hasPersistedSession, isTrue);
  });

  test('cached premium profile flags do not grant offline access', () async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    final store = OfflineAccessStore(
      preferences,
      secretStorage: _MemorySecrets(),
      verifier: _AcceptTestProof(),
    );
    await store.cacheProfile({
      'id': 'account-1',
      'displayName': 'Learner',
      'email': 'learner@example.com',
      'entitlement': {'tier': 'premium', 'isLifetime': true},
    });

    expect(store.cachedProfile?['entitlement'], {
      'tier': 'free',
      'isLifetime': false,
    });
    expect(await store.canReadPremium('account-1'), isFalse);
  });

  test('revoked session immediately gates an otherwise valid lease', () async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    final store = OfflineAccessStore(
      preferences,
      secretStorage: _MemorySecrets(),
      verifier: _AcceptTestProof(),
    );
    await store.saveSessionToken('valid-session');
    await store.cacheProfile({
      'id': 'account-1',
      'displayName': 'Learner',
      'email': 'learner@example.com',
    });
    await store.applyServerSnapshot(
      userId: 'account-1',
      premium: true,
      lifetime: true,
      signedLease: 'proof',
    );
    expect(await store.canReadPremium('account-1'), isTrue);

    await store.clearSessionToken();

    expect(await store.canReadPremium('account-1'), isFalse);
  });

  test(
    'refund during lease verification invalidates the pending access check',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final secrets = _MemorySecrets();
      final setupStore = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: _AcceptTestProof(),
      );
      await setupStore.saveSessionToken('session-token');
      await setupStore.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
      });
      await setupStore.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'proof',
      );

      final verifier = _BlockingVerifier();
      final store = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: verifier,
      );
      final accessCheck = store.canReadPremium('account-1');
      await verifier.started.future;
      await store.applyServerSnapshot(
        userId: 'account-1',
        premium: false,
        lifetime: false,
      );
      verifier.release();

      expect(await accessCheck, isFalse);
      expect(await store.canReadPremium('account-1'), isFalse);
    },
  );

  test(
    'logout during lease verification invalidates the pending access check',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final secrets = _MemorySecrets();
      final setupStore = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: _AcceptTestProof(),
      );
      await setupStore.saveSessionToken('session-token');
      await setupStore.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
      });
      await setupStore.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'proof',
      );

      final verifier = _BlockingVerifier();
      final store = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: verifier,
      );
      final accessCheck = store.canReadPremium('account-1');
      await verifier.started.future;
      store.beginLocalLogout();
      verifier.release();

      expect(await accessCheck, isFalse);
    },
  );

  test('verified lease arriving after logout is never written', () async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    final secrets = _MemorySecrets();
    final verifier = _BlockingVerifier();
    final store = OfflineAccessStore(
      preferences,
      secretStorage: secrets,
      verifier: verifier,
    );
    await store.saveSessionToken('session-token');
    await store.cacheProfile({
      'id': 'account-1',
      'displayName': 'Learner',
      'email': 'learner@example.com',
    });

    final snapshot = store.applyServerSnapshot(
      userId: 'account-1',
      premium: true,
      lifetime: true,
      signedLease: 'proof',
    );
    await verifier.started.future;
    store.beginLocalLogout();
    final clearing = store.clearPrivateData();
    verifier.release();
    await Future.wait([snapshot, clearing]);

    expect(
      secrets.values.keys.where(
        (key) => key.startsWith('offline_access_lease_v1_'),
      ),
      isEmpty,
    );
    expect(await store.canReadPremium('account-1'), isFalse);
  });

  test('clock rollback beyond tolerance fails closed', () async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    final seed = List<int>.generate(32, (index) => index + 1);
    final algorithm = Ed25519();
    final keyPair = await algorithm.newKeyPairFromSeed(seed);
    final publicKey = await keyPair.extractPublicKey();
    final secrets = _MemorySecrets();
    final issuedAt = DateTime.utc(2026, 10, 23);
    var clock = issuedAt;
    final store = OfflineAccessStore(
      preferences,
      secretStorage: secrets,
      verifier: SignedOfflineEntitlementVerifier(base64Encode(publicKey.bytes)),
      clock: () => clock,
    );
    await store.saveSessionToken('session-token');
    await store.cacheProfile({
      'id': 'account-1',
      'displayName': 'Learner',
      'email': 'learner@example.com',
    });
    await store.applyServerSnapshot(
      userId: 'account-1',
      premium: true,
      lifetime: true,
      signedLease: await _signedLease(
        algorithm: algorithm,
        keyPair: keyPair,
        userId: 'account-1',
        issuedAt: issuedAt,
      ),
    );
    expect(await store.canReadPremium('account-1'), isTrue);
    clock = issuedAt.subtract(const Duration(hours: 1));
    expect(await store.canReadPremium('account-1'), isFalse);
  });

  test(
    'logout removes private account state and keeps public content',
    () async {
      SharedPreferences.setMockInitialValues({
        authTokenStorageKey: 'private-session',
        'dictionary_library_v1_account-1': '{}',
        'learning_history_v1_account-1': '[]',
        'anki_reviews_v1_account-1': '[]',
        'ssw_kaigo_completed_v1_account-1': <String>['module-1'],
        'public_catalog_cache': 'shared',
      });
      final preferences = await SharedPreferences.getInstance();
      final secrets = _MemorySecrets();
      final store = OfflineAccessStore(
        preferences,
        secretStorage: secrets,
        verifier: _AcceptTestProof(),
      );
      await store.cacheProfile({
        'id': 'account-1',
        'displayName': 'Learner',
        'email': 'learner@example.com',
      });
      await store.applyServerSnapshot(
        userId: 'account-1',
        premium: true,
        lifetime: true,
        signedLease: 'proof',
      );
      expect(await store.canReadPremium('account-1'), isTrue);

      await store.clearPrivateData();

      expect(store.userId, isNull);
      expect(preferences.getString(authTokenStorageKey), isNull);
      expect(preferences.getString('dictionary_library_v1_account-1'), isNull);
      expect(preferences.getString('learning_history_v1_account-1'), isNull);
      expect(preferences.getString('anki_reviews_v1_account-1'), isNull);
      expect(
        preferences.getStringList('ssw_kaigo_completed_v1_account-1'),
        isNull,
      );
      expect(preferences.getString('public_catalog_cache'), 'shared');
      expect(await store.canReadPremium('account-1'), isFalse);
    },
  );
}

Future<String> _signedLease({
  required Ed25519 algorithm,
  required KeyPair keyPair,
  required String userId,
  required DateTime issuedAt,
}) async {
  String encode(Object value) =>
      base64Url.encode(utf8.encode(jsonEncode(value))).replaceAll('=', '');
  final header = encode({
    'alg': 'EdDSA',
    'typ': 'JWT',
    'kid': 'offline-access-v1',
  });
  final payload = encode({
    'iss': 'nihongo-master',
    'sub': userId,
    'aud': 'nihongo-master-offline-premium',
    'tier': 'premium',
    'isLifetime': true,
    'iat': issuedAt.millisecondsSinceEpoch ~/ 1000,
    'exp':
        issuedAt.add(const Duration(days: 30)).millisecondsSinceEpoch ~/ 1000,
    'jti': 'android:purchase-1',
  });
  final message = '$header.$payload';
  final signature = await algorithm.sign(
    utf8.encode(message),
    keyPair: keyPair,
  );
  return '$message.${base64Url.encode(signature.bytes).replaceAll('=', '')}';
}

class _MemorySecrets implements OfflineSecretStorage {
  final Map<String, String> values = {};

  @override
  Future<void> delete(String key) async => values.remove(key);

  @override
  Future<String?> read(String key) async => values[key];

  @override
  Future<void> write({required String key, required String value}) async {
    values[key] = value;
  }
}

class _AcceptTestProof implements OfflineEntitlementVerifier {
  @override
  Future<bool> verify({
    required String proof,
    required String userId,
    required DateTime now,
  }) async => proof == 'proof' && userId == 'account-1';
}

class _BlockingVerifier implements OfflineEntitlementVerifier {
  final Completer<void> started = Completer<void>();
  final Completer<void> _release = Completer<void>();

  void release() {
    if (!_release.isCompleted) _release.complete();
  }

  @override
  Future<bool> verify({
    required String proof,
    required String userId,
    required DateTime now,
  }) async {
    if (!started.isCompleted) started.complete();
    await _release.future;
    return proof == 'proof' && userId == 'account-1';
  }
}
