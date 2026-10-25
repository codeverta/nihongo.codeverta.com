import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:cryptography/cryptography.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';
import 'package:path_provider/path_provider.dart';

export 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';

class EncryptedOfflinePracticeLocalStore implements OfflinePracticeLocalStore {
  EncryptedOfflinePracticeLocalStore({
    OfflineSecretStorage? secretStorage,
    this.directoryPath,
  }) : _secretStorage = secretStorage ?? FlutterSecureOfflineSecretStorage();

  final OfflineSecretStorage _secretStorage;
  final String? directoryPath;
  final AesGcm _cipher = AesGcm.with256bits();
  Future<Directory>? _directory;

  Future<Directory> get _root => _directory ??= _resolveDirectory();

  Future<Directory> _resolveDirectory() async {
    final path = directoryPath;
    final result = path == null
        ? Directory(
            '${(await getApplicationSupportDirectory()).path}${Platform.pathSeparator}offline_practice',
          )
        : Directory(path);
    await result.create(recursive: true);
    return result;
  }

  Future<String> _userHash(String userId) async {
    final hash = await Sha256().hash(utf8.encode(userId));
    return hash.bytes
        .map((byte) => byte.toRadixString(16).padLeft(2, '0'))
        .join();
  }

  Future<File> _file(String userId) async => File(
    '${(await _root).path}${Platform.pathSeparator}${await _userHash(userId)}.outbox',
  );

  String _keyName(String userId) => 'offline_practice_key_v1_$userId';

  Future<SecretKey> _loadKey(String userId, {required bool create}) async {
    final raw = await _secretStorage.read(_keyName(userId));
    if (raw != null) {
      try {
        final bytes = base64Url.decode(raw);
        if (bytes.length == 32) return SecretKey(bytes);
      } on FormatException {
        // Replace the unusable key. The corresponding outbox will be discarded.
      }
    }
    if (!create) throw const FormatException('Missing offline practice key.');
    final key = await _cipher.newSecretKey();
    await _secretStorage.write(
      key: _keyName(userId),
      value: base64UrlEncode(await key.extractBytes()),
    );
    return key;
  }

  @override
  Future<List<Map<String, dynamic>>> read(String userId) async {
    if (userId.isEmpty) return [];
    final file = await _file(userId);
    if (!file.existsSync()) return [];
    try {
      final bytes = await file.readAsBytes();
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
      // A missing or damaged key must fail closed instead of leaking plaintext.
      file.deleteSync();
      return [];
    }
  }

  @override
  Future<void> write(String userId, List<Map<String, dynamic>> attempts) async {
    if (userId.isEmpty) throw ArgumentError.value(userId, 'userId');
    final file = await _file(userId);
    final box = await _cipher.encrypt(
      utf8.encode(jsonEncode(attempts)),
      secretKey: await _loadKey(userId, create: true),
    );
    final contents = Uint8List.fromList(<int>[
      ...box.nonce,
      ...box.cipherText,
      ...box.mac.bytes,
    ]);
    final temporary = File('${file.path}.tmp');
    await temporary.writeAsBytes(contents, flush: true);
    if (file.existsSync()) file.deleteSync();
    await temporary.rename(file.path);
  }

  @override
  Future<void> clear(String userId) async {
    if (userId.isEmpty) return;
    final file = await _file(userId);
    if (file.existsSync()) file.deleteSync();
    await _secretStorage.delete(_keyName(userId));
  }
}
