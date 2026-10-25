import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'package:cryptography/cryptography.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'package:nihongo_master/core/offline/learning_cache_persistence.dart';

const int defaultLearningCacheLimitBytes = 100 * 1024 * 1024;
const Duration learningCacheFreshness = Duration(hours: 24);
const String guestCacheAccountId = '__guest__';

final cachedResponseMetadataProvider = StateProvider<CachedResponseMetadata?>(
  (ref) => null,
);

enum CachedResourceKind { text, media }

class CachedResponseMetadata {
  const CachedResponseMetadata({
    required this.cacheKey,
    required this.cached,
    required this.offline,
    required this.stale,
    required this.updatedAt,
  });

  final String cacheKey;
  final bool cached;
  final bool offline;
  final bool stale;
  final DateTime? updatedAt;

  Map<String, Object?> toExtra() => <String, Object?>{
    'cached': cached,
    'offline': offline,
    'stale': stale,
    'updatedAt': updatedAt?.toIso8601String(),
  };
}

class CachedResponse {
  const CachedResponse({
    required this.body,
    required this.statusCode,
    required this.statusMessage,
    required this.headers,
    required this.contentType,
    required this.updatedAt,
    required this.stale,
    required this.requiresPremium,
    required this.premiumClassificationKnown,
  });

  final Uint8List body;
  final int statusCode;
  final String? statusMessage;
  final Map<String, String> headers;
  final String contentType;
  final DateTime updatedAt;
  final bool stale;
  final bool requiresPremium;
  final bool premiumClassificationKnown;
}

class CachedAsset {
  const CachedAsset({
    required this.bytes,
    required this.contentType,
    required this.updatedAt,
    required this.pinned,
  });

  final Uint8List bytes;
  final String contentType;
  final DateTime updatedAt;
  final bool pinned;
}

class LearningCacheStats {
  const LearningCacheStats({
    required this.totalBytes,
    required this.textBytes,
    required this.mediaBytes,
    required this.pinnedBytes,
    required this.entryCount,
    required this.limitBytes,
  });

  final int totalBytes;
  final int textBytes;
  final int mediaBytes;
  final int pinnedBytes;
  final int entryCount;
  final int limitBytes;

  double get usageRatio => limitBytes == 0 ? 0 : totalBytes / limitBytes;
}

abstract interface class LearningCacheCipher {
  Future<Uint8List> encrypt(Uint8List plaintext);
  Future<Uint8List> decrypt(Uint8List ciphertext);
}

/// AES-GCM encryption for persisted response bodies and downloaded media.
/// Its key is stored by the platform's secure storage implementation.
class SecureLearningCacheCipher implements LearningCacheCipher {
  SecureLearningCacheCipher({FlutterSecureStorage? secureStorage})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage();

  static const _keyName = 'nihongo_learning_cache_key_v1';
  static const _nonceLength = 12;
  static const _macLength = 16;

  final FlutterSecureStorage _secureStorage;
  final AesGcm _algorithm = AesGcm.with256bits();
  Future<SecretKey>? _keyFuture;

  Future<SecretKey> get _key => _keyFuture ??= _loadOrCreateKey();

  Future<SecretKey> _loadOrCreateKey() async {
    final stored = await _secureStorage.read(key: _keyName);
    if (stored != null) {
      try {
        final bytes = base64Url.decode(stored);
        if (bytes.length == 32) return SecretKey(bytes);
      } on FormatException {
        // Replace a damaged key and let old cache rows be discarded on read.
      }
    }
    final generated = await _algorithm.newSecretKey();
    await _secureStorage.write(
      key: _keyName,
      value: base64UrlEncode(await generated.extractBytes()),
    );
    return generated;
  }

  @override
  Future<Uint8List> encrypt(Uint8List plaintext) async {
    final box = await _algorithm.encrypt(plaintext, secretKey: await _key);
    return Uint8List.fromList(<int>[
      ...box.nonce,
      ...box.mac.bytes,
      ...box.cipherText,
    ]);
  }

  @override
  Future<Uint8List> decrypt(Uint8List ciphertext) async {
    if (ciphertext.length < _nonceLength + _macLength) {
      throw const FormatException('Invalid encrypted learning cache entry.');
    }
    final box = SecretBox.fromConcatenation(
      ciphertext,
      nonceLength: _nonceLength,
      macLength: _macLength,
    );
    final plaintext = await _algorithm.decrypt(box, secretKey: await _key);
    return Uint8List.fromList(plaintext);
  }
}

/// Useful for isolated tests. Production callers should use the secure default.
class PlaintextLearningCacheCipher implements LearningCacheCipher {
  @override
  Future<Uint8List> decrypt(Uint8List ciphertext) async =>
      Uint8List.fromList(ciphertext);

  @override
  Future<Uint8List> encrypt(Uint8List plaintext) async =>
      Uint8List.fromList(plaintext);
}

class _CacheRecord {
  const _CacheRecord({
    required this.key,
    required this.accountId,
    required this.bodyId,
    required this.bytes,
    required this.kind,
    required this.pinned,
    required this.contentType,
    required this.updatedAt,
    required this.lastUsedAt,
    this.statusCode = 200,
    this.statusMessage,
    this.headers = const <String, String>{},
    this.requiresPremium = false,
    this.premiumClassificationKnown = false,
  });

  factory _CacheRecord.fromJson(Map<String, dynamic> json) => _CacheRecord(
    key: json['key'] as String,
    accountId: json['accountId'] as String,
    bodyId: json['bodyId'] as String,
    bytes: json['bytes'] as int,
    kind: CachedResourceKind.values.byName(json['kind'] as String),
    pinned: json['pinned'] as bool? ?? false,
    contentType: json['contentType'] as String? ?? 'application/octet-stream',
    updatedAt: DateTime.parse(json['updatedAt'] as String),
    lastUsedAt: DateTime.parse(json['lastUsedAt'] as String),
    statusCode: json['statusCode'] as int? ?? 200,
    statusMessage: json['statusMessage'] as String?,
    headers:
        (json['headers'] as Map?)?.map(
          (key, value) => MapEntry(key.toString(), value.toString()),
        ) ??
        const <String, String>{},
    requiresPremium: json['requiresPremium'] as bool? ?? false,
    premiumClassificationKnown:
        json['premiumClassificationKnown'] as bool? ?? false,
  );

  final String key;
  final String accountId;
  final String bodyId;
  final int bytes;
  final CachedResourceKind kind;
  final bool pinned;
  final String contentType;
  final DateTime updatedAt;
  final DateTime lastUsedAt;
  final int statusCode;
  final String? statusMessage;
  final Map<String, String> headers;
  final bool requiresPremium;
  final bool premiumClassificationKnown;

  _CacheRecord touched(DateTime now) => _CacheRecord(
    key: key,
    accountId: accountId,
    bodyId: bodyId,
    bytes: bytes,
    kind: kind,
    pinned: pinned,
    contentType: contentType,
    updatedAt: updatedAt,
    lastUsedAt: now,
    statusCode: statusCode,
    statusMessage: statusMessage,
    headers: headers,
    requiresPremium: requiresPremium,
    premiumClassificationKnown: premiumClassificationKnown,
  );

  Map<String, Object?> toJson() => <String, Object?>{
    'key': key,
    'accountId': accountId,
    'bodyId': bodyId,
    'bytes': bytes,
    'kind': kind.name,
    'pinned': pinned,
    'contentType': contentType,
    'updatedAt': updatedAt.toIso8601String(),
    'lastUsedAt': lastUsedAt.toIso8601String(),
    'statusCode': statusCode,
    'statusMessage': statusMessage,
    'headers': headers,
    'requiresPremium': requiresPremium,
    'premiumClassificationKnown': premiumClassificationKnown,
  };
}

/// Persistent encrypted HTTP content and media cache, scoped by account ID.
class PersistentLearningCache {
  PersistentLearningCache._({
    required LearningCachePersistence persistence,
    required LearningCacheCipher cipher,
    required this.maxBytes,
    required DateTime Function() clock,
  }) : _persistence = persistence,
       _cipher = cipher,
       _clock = clock;

  static Future<PersistentLearningCache> open({
    LearningCachePersistence? persistence,
    LearningCacheCipher? cipher,
    int maxBytes = defaultLearningCacheLimitBytes,
    DateTime Function()? clock,
    String? directoryPath,
  }) async {
    if (maxBytes <= 0) throw ArgumentError.value(maxBytes, 'maxBytes');
    final cache = PersistentLearningCache._(
      persistence:
          persistence ??
          createLearningCachePersistence(directoryPath: directoryPath),
      cipher: cipher ?? SecureLearningCacheCipher(),
      maxBytes: maxBytes,
      clock: clock ?? DateTime.now,
    );
    await cache._loadIndex();
    return cache;
  }

  final LearningCachePersistence _persistence;
  final LearningCacheCipher _cipher;
  final DateTime Function() _clock;
  final int maxBytes;
  final Map<String, _CacheRecord> _records = <String, _CacheRecord>{};
  final StreamController<LearningCacheStats> _statsChanges =
      StreamController<LearningCacheStats>.broadcast();
  Future<void> _queue = Future<void>.value();

  Stream<LearningCacheStats> get statsChanges => _statsChanges.stream;

  Future<void> _loadIndex() async {
    final raw = await _persistence.readIndex();
    if (raw == null || raw.isEmpty) return;
    try {
      final decoded = jsonDecode(raw);
      if (decoded is! List) return;
      for (final item in decoded.whereType<Map<String, dynamic>>()) {
        final record = _CacheRecord.fromJson(item);
        _records[record.key] = record;
      }
    } on Object {
      _records.clear();
      await _persistence.clear();
    }
    final missing = <String>[];
    for (final record in _records.values) {
      if (await _persistence.readBody(record.bodyId) == null) {
        missing.add(record.key);
      }
    }
    for (final key in missing) {
      _records.remove(key);
    }
    if (missing.isNotEmpty) await _saveIndex();
  }

  Future<T> _locked<T>(Future<T> Function() action) {
    final result = Completer<T>();
    _queue = _queue.then((_) async {
      try {
        result.complete(await action());
      } on Object catch (error, stack) {
        result.completeError(error, stack);
      }
    });
    return result.future;
  }

  Future<void> _saveIndex() => _persistence.writeIndex(
    jsonEncode(_records.values.map((record) => record.toJson()).toList()),
  );

  String _scopedKey(String accountId, String key) =>
      jsonEncode(<String>[accountId, key]);

  String _newBodyId() {
    final random = Random.secure();
    return base64UrlEncode(
      List<int>.generate(18, (_) => random.nextInt(256)),
    ).replaceAll('=', '');
  }

  Future<CachedResponse?> readResponse({
    required String accountId,
    required String key,
    DateTime? now,
  }) => _locked(() async {
    final scopedKey = _scopedKey(accountId, key);
    final record = _records[scopedKey];
    if (record == null || record.kind != CachedResourceKind.text) return null;
    final encrypted = await _persistence.readBody(record.bodyId);
    if (encrypted == null) {
      _records.remove(scopedKey);
      await _saveIndex();
      return null;
    }
    Uint8List body;
    try {
      body = await _cipher.decrypt(encrypted);
    } on Object {
      await _persistence.deleteBody(record.bodyId);
      _records.remove(scopedKey);
      await _saveIndex();
      return null;
    }
    final touched = record.touched(_clock());
    _records[scopedKey] = touched;
    await _saveIndex();
    _emitStats();
    final current = now ?? _clock();
    return CachedResponse(
      body: body,
      statusCode: record.statusCode,
      statusMessage: record.statusMessage,
      headers: record.headers,
      contentType: record.contentType,
      updatedAt: record.updatedAt,
      stale: current.difference(record.updatedAt) > learningCacheFreshness,
      requiresPremium: record.requiresPremium,
      premiumClassificationKnown: record.premiumClassificationKnown,
    );
  });

  Future<bool> writeResponse({
    required String accountId,
    required String key,
    required Uint8List body,
    required String contentType,
    required int statusCode,
    required String? statusMessage,
    required Map<String, String> headers,
    required bool requiresPremium,
    bool premiumClassificationKnown = false,
    bool pinned = false,
  }) => _write(
    accountId: accountId,
    key: key,
    body: body,
    kind: CachedResourceKind.text,
    pinned: pinned,
    contentType: contentType,
    statusCode: statusCode,
    statusMessage: statusMessage,
    headers: headers,
    requiresPremium: requiresPremium,
    premiumClassificationKnown: premiumClassificationKnown,
  );

  Future<CachedAsset?> readAsset({
    required String userId,
    required String url,
  }) => _locked(() async {
    final record = _records[_scopedKey(userId, 'asset:$url')];
    if (record == null || record.kind != CachedResourceKind.media) return null;
    final encrypted = await _persistence.readBody(record.bodyId);
    if (encrypted == null) return null;
    try {
      final bytes = await _cipher.decrypt(encrypted);
      _records[record.key] = record.touched(_clock());
      await _saveIndex();
      _emitStats();
      return CachedAsset(
        bytes: bytes,
        contentType: record.contentType,
        updatedAt: record.updatedAt,
        pinned: record.pinned,
      );
    } on Object {
      await _persistence.deleteBody(record.bodyId);
      _records.remove(record.key);
      await _saveIndex();
      return null;
    }
  });

  Future<bool> writeAsset({
    required String userId,
    required String url,
    required Uint8List bytes,
    String contentType = 'application/octet-stream',
    bool pinned = false,
  }) => _write(
    accountId: userId,
    key: 'asset:$url',
    body: bytes,
    kind: CachedResourceKind.media,
    pinned: pinned,
    contentType: contentType,
    statusCode: 200,
    statusMessage: null,
    headers: const <String, String>{},
    requiresPremium: false,
  );

  /// Changes whether an existing response is protected from automatic LRU
  /// eviction. A missing response is left unchanged.
  Future<bool> setResponsePinned({
    required String accountId,
    required String key,
    required bool pinned,
  }) => _setPinned(accountId, key, pinned);

  /// Changes whether an existing asset is protected from automatic LRU
  /// eviction. A missing asset is left unchanged.
  Future<bool> setAssetPinned({
    required String userId,
    required String url,
    required bool pinned,
  }) => _setPinned(userId, 'asset:$url', pinned);

  Future<bool> _setPinned(String accountId, String key, bool pinned) =>
      _locked(() async {
        final scopedKey = _scopedKey(accountId, key);
        final record = _records[scopedKey];
        if (record == null || record.pinned == pinned) return record != null;
        _records[scopedKey] = _CacheRecord(
          key: record.key,
          accountId: record.accountId,
          bodyId: record.bodyId,
          bytes: record.bytes,
          kind: record.kind,
          pinned: pinned,
          contentType: record.contentType,
          updatedAt: record.updatedAt,
          lastUsedAt: record.lastUsedAt,
          statusCode: record.statusCode,
          statusMessage: record.statusMessage,
          headers: record.headers,
          requiresPremium: record.requiresPremium,
          premiumClassificationKnown: record.premiumClassificationKnown,
        );
        await _saveIndex();
        _emitStats();
        return true;
      });

  Future<void> removeAsset({required String userId, required String url}) =>
      removeResponse(accountId: userId, key: 'asset:$url');

  Future<bool> _write({
    required String accountId,
    required String key,
    required Uint8List body,
    required CachedResourceKind kind,
    required bool pinned,
    required String contentType,
    required int statusCode,
    required String? statusMessage,
    required Map<String, String> headers,
    required bool requiresPremium,
    bool premiumClassificationKnown = false,
  }) => _locked(() async {
    final scopedKey = _scopedKey(accountId, key);
    final old = _records[scopedKey];
    final effectivePinned = pinned || (old?.pinned ?? false);
    final encrypted = await _cipher.encrypt(body);
    final bodyId = _newBodyId();
    final newRecord = _CacheRecord(
      key: scopedKey,
      accountId: accountId,
      bodyId: bodyId,
      bytes: encrypted.length,
      kind: kind,
      pinned: effectivePinned,
      contentType: contentType,
      updatedAt: _clock(),
      lastUsedAt: _clock(),
      statusCode: statusCode,
      statusMessage: statusMessage,
      headers: Map<String, String>.unmodifiable(headers),
      requiresPremium: requiresPremium,
      premiumClassificationKnown: premiumClassificationKnown,
    );
    await _persistence.writeBody(bodyId, encrypted);

    final candidates = _records.values
        .where((record) => record.key != scopedKey)
        .toList();
    var projected =
        candidates.fold<int>(0, (sum, record) => sum + record.bytes) +
        newRecord.bytes;
    final evictionOrder =
        <_CacheRecord>[
          ...candidates.where((record) => !record.pinned),
          if (!effectivePinned) newRecord,
        ]..sort((a, b) {
          final kindOrder = (a.kind == CachedResourceKind.media ? 0 : 1)
              .compareTo(b.kind == CachedResourceKind.media ? 0 : 1);
          return kindOrder != 0
              ? kindOrder
              : a.lastUsedAt.compareTo(b.lastUsedAt);
        });
    final evicted = <_CacheRecord>[];
    var incomingEvicted = false;
    for (final record in evictionOrder) {
      if (projected <= maxBytes) break;
      projected -= record.bytes;
      if (record.key == newRecord.key) {
        incomingEvicted = true;
      } else {
        evicted.add(record);
      }
    }
    if (projected > maxBytes || incomingEvicted) {
      await _persistence.deleteBody(bodyId);
      return false;
    }

    _records.remove(scopedKey);
    _records[scopedKey] = newRecord;
    for (final record in evicted) {
      _records.remove(record.key);
      await _persistence.deleteBody(record.bodyId);
    }
    if (old != null) await _persistence.deleteBody(old.bodyId);
    await _saveIndex();
    _emitStats();
    return true;
  });

  Future<void> removeResponse({
    required String accountId,
    required String key,
  }) => _locked(() async {
    final scopedKey = _scopedKey(accountId, key);
    final record = _records.remove(scopedKey);
    if (record != null) {
      await _persistence.deleteBody(record.bodyId);
      await _saveIndex();
      _emitStats();
    }
  });

  /// Removes account-scoped text snapshots selected by their logical route key.
  /// Media assets and pinned content remain available.
  Future<void> invalidateResponsesForAccount({
    required String accountId,
    required bool Function(String routeKey) shouldInvalidate,
  }) => _locked(() async {
    final matching = _records.values.where((record) {
      if (record.accountId != accountId ||
          record.kind != CachedResourceKind.text) {
        return false;
      }
      try {
        final decoded = jsonDecode(record.key);
        if (decoded is List && decoded.length == 2 && decoded[1] is String) {
          return shouldInvalidate(decoded[1] as String);
        }
      } on Object {
        return false;
      }
      return false;
    }).toList();
    for (final record in matching) {
      _records.remove(record.key);
      await _persistence.deleteBody(record.bodyId);
    }
    if (matching.isNotEmpty) {
      await _saveIndex();
      _emitStats();
    }
  });

  Future<void> clearAccount(String userId) => _locked(() async {
    final matching = _records.values
        .where((record) => record.accountId == userId)
        .toList();
    for (final record in matching) {
      _records.remove(record.key);
      await _persistence.deleteBody(record.bodyId);
    }
    await _saveIndex();
    _emitStats();
  });

  /// Clears automatically cached media while retaining lesson text and any
  /// media the user explicitly pinned for offline use.
  Future<void> clearAutomaticMedia() => _locked(() async {
    final automaticMedia = _records.values
        .where(
          (record) => record.kind == CachedResourceKind.media && !record.pinned,
        )
        .toList();
    for (final record in automaticMedia) {
      _records.remove(record.key);
      await _persistence.deleteBody(record.bodyId);
    }
    if (automaticMedia.isNotEmpty) {
      await _saveIndex();
      _emitStats();
    }
  });

  Future<void> clearAll() => _locked(() async {
    _records.clear();
    await _persistence.clear();
    await _saveIndex();
    _emitStats();
  });

  Future<LearningCacheStats> stats() => _locked(() async => _buildStats());

  LearningCacheStats _buildStats() {
    var text = 0;
    var media = 0;
    var pinned = 0;
    for (final record in _records.values) {
      if (record.kind == CachedResourceKind.text) {
        text += record.bytes;
      } else {
        media += record.bytes;
      }
      if (record.pinned) pinned += record.bytes;
    }
    return LearningCacheStats(
      totalBytes: text + media,
      textBytes: text,
      mediaBytes: media,
      pinnedBytes: pinned,
      entryCount: _records.length,
      limitBytes: maxBytes,
    );
  }

  void _emitStats() {
    if (!_statsChanges.isClosed) _statsChanges.add(_buildStats());
  }

  Future<void> dispose() => _statsChanges.close();
}
