import 'dart:convert';

import 'package:cryptography/cryptography.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';
import 'package:shared_preferences/shared_preferences.dart';

export 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';

/// Encrypted browser storage for the offline attempt outbox.
class EncryptedOfflinePracticeLocalStore implements OfflinePracticeLocalStore {
  EncryptedOfflinePracticeLocalStore({
    OfflineSecretStorage? secretStorage,
    SharedPreferences? preferences,
  }) : _secretStorage = secretStorage ?? FlutterSecureOfflineSecretStorage(),
       _preferences = preferences;

  final OfflineSecretStorage _secretStorage;
  final SharedPreferences? _preferences;
  final AesGcm _cipher = AesGcm.with256bits();
  Future<SharedPreferences>? _preferencesFuture;

  Future<SharedPreferences> get _prefs =>
      _preferencesFuture ??= _loadPreferences();

  Future<SharedPreferences> _loadPreferences() async =>
      _preferences ?? await SharedPreferences.getInstance();

  Future<String> _keyName(String userId) async {
    final hash = await Sha256().hash(utf8.encode(userId));
    final value = hash.bytes
        .map((byte) => byte.toRadixString(16).padLeft(2, '0'))
        .join();
    return 'offline_practice_key_v1_$value';
  }

  String _storageName(String userId) => 'offline_practice_v1_$userId';

  Future<SecretKey> _loadKey(String userId, {required bool create}) async {
    final name = await _keyName(userId);
    final raw = await _secretStorage.read(name);
    if (raw != null) {
      try {
        final bytes = base64Url.decode(raw);
        if (bytes.length == 32) return SecretKey(bytes);
      } on FormatException {
        // The unreadable outbox is discarded by read().
      }
    }
    if (!create) throw const FormatException('Missing offline practice key.');
    final key = await _cipher.newSecretKey();
    await _secretStorage.write(
      key: name,
      value: base64UrlEncode(await key.extractBytes()),
    );
    return key;
  }

  @override
  Future<List<Map<String, dynamic>>> read(String userId) async {
    if (userId.isEmpty) return [];
    final prefs = await _prefs;
    final raw = prefs.getString(_storageName(userId));
    if (raw == null) return [];
    try {
      final bytes = base64Url.decode(raw);
      final box = SecretBox.fromConcatenation(
        bytes,
        nonceLength: 12,
        macLength: 16,
      );
      final clearText = await _cipher.decrypt(
        box,
        secretKey: await _loadKey(userId, create: false),
      );
      final value = jsonDecode(utf8.decode(clearText));
      if (value is! List) throw const FormatException('Invalid outbox.');
      return value
          .whereType<Map<Object?, Object?>>()
          .map(Map<String, dynamic>.from)
          .toList(growable: true);
    } on Object {
      await prefs.remove(_storageName(userId));
      return [];
    }
  }

  @override
  Future<void> write(String userId, List<Map<String, dynamic>> attempts) async {
    if (userId.isEmpty) throw ArgumentError.value(userId, 'userId');
    final prefs = await _prefs;
    final box = await _cipher.encrypt(
      utf8.encode(jsonEncode(attempts)),
      secretKey: await _loadKey(userId, create: true),
    );
    final contents = <int>[...box.nonce, ...box.cipherText, ...box.mac.bytes];
    if (!await prefs.setString(
      _storageName(userId),
      base64UrlEncode(contents),
    )) {
      throw StateError('Offline practice outbox could not be saved.');
    }
  }

  @override
  Future<void> clear(String userId) async {
    if (userId.isEmpty) return;
    final prefs = await _prefs;
    await prefs.remove(_storageName(userId));
    await _secretStorage.delete(await _keyName(userId));
  }
}
