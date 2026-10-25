import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice_store.dart';

// Keep the encrypted desktop/mobile outbox below 7 MiB. Browser storage uses
// base64, so its persisted payload remains below the cache's 10 MiB reserve.
const int maxOfflinePracticeOutboxBytes = 7 * 1024 * 1024;

final offlinePracticeServiceProvider = Provider<OfflinePracticeService>((ref) {
  final service = OfflinePracticeService(
    dio: ref.watch(dioProvider),
    store: EncryptedOfflinePracticeLocalStore(),
    currentUserId: () => ref.read(authControllerProvider).user?.id,
    isOffline: () => ref.read(offlineProvider),
    canReadPremium: (userId) =>
        ref.read(offlineAccessStoreProvider).canReadPremium(userId),
  );
  ref.onDispose(service.dispose);
  return service;
});

class OfflinePracticeSyncEvent {
  const OfflinePracticeSyncEvent({
    required this.userId,
    required this.attemptId,
    required this.result,
  });

  final String userId;
  final String attemptId;
  final Map<String, dynamic> result;
}

class OfflinePracticeUnavailableException implements Exception {
  const OfflinePracticeUnavailableException();

  @override
  String toString() =>
      'Kuis belum disiapkan untuk offline. Hubungkan internet dan buka kuis saat online untuk menyimpannya.';
}

/// Uses server-issued lesson banks for both online and offline lesson quizzes.
/// The encrypted outbox retains the signed package and selected option IDs;
/// server results are the
/// only results marked verified and eligible for official progress/rewards.
class OfflinePracticeService {
  OfflinePracticeService({
    required this.dio,
    required this.store,
    required this.currentUserId,
    required this.isOffline,
    Future<bool> Function(String userId)? canReadPremium,
    this.syncTimeout = const Duration(seconds: 3),
  }) : canReadPremium = canReadPremium ?? _denyPremium;

  final Dio dio;
  final OfflinePracticeLocalStore store;
  final String? Function() currentUserId;
  final bool Function() isOffline;
  final Future<bool> Function(String userId) canReadPremium;
  final Duration syncTimeout;

  final StreamController<OfflinePracticeSyncEvent> _syncEvents =
      StreamController<OfflinePracticeSyncEvent>.broadcast();
  final Map<String, int> _generations = {};
  final Map<String, Set<CancelToken>> _requests = {};
  final Map<String, Future<Map<String, dynamic>?>> _inflight = {};
  Future<void> _writes = Future<void>.value();
  Future<void>? _flushing;
  bool _disposed = false;

  Stream<OfflinePracticeSyncEvent> get syncEvents => _syncEvents.stream;

  Future<bool> ownsAttempt(String attemptId) async {
    final user = currentUserId();
    if (user == null) return false;
    final attempts = await store.read(user);
    final row = _attempt(attempts, attemptId);
    return currentUserId() == user &&
        row != null &&
        await _isAuthorizedRow(user, row);
  }

  Future<int> get pendingCount async {
    final user = currentUserId();
    if (user == null) return 0;
    final attempts = await store.read(user);
    return attempts.where((row) => row['status'] == 'pending').length;
  }

  Future<List<Map<String, dynamic>>> loadBank(
    String kind,
    String lessonId,
  ) async {
    final user = currentUserId();
    if (user == null || user.isEmpty) {
      throw StateError('Masuk untuk menyimpan latihan.');
    }
    if (!_supportedKind(kind)) {
      throw ArgumentError.value(kind, 'kind');
    }
    final generation = _generations[user] ?? 0;
    final package = await _fetchPackage(user, kind, lessonId, generation);
    return (package['questions'] as List<dynamic>)
        .whereType<Map<String, dynamic>>()
        .map(_learnerQuestion)
        .toList(growable: false);
  }

  Future<Map<String, dynamic>> start(
    String kind,
    String lessonId, {
    String? scope,
  }) async {
    final user = currentUserId();
    if (user == null || user.isEmpty) {
      throw StateError('Masuk untuk menyimpan latihan.');
    }
    if (!_supportedKind(kind)) throw ArgumentError.value(kind, 'kind');
    if (kind != 'ssw' && scope != null) {
      throw ArgumentError.value(scope, 'scope');
    }
    final generation = _generations[user] ?? 0;
    final package = await _fetchPackage(user, kind, lessonId, generation);
    final questions = package['questions'] as List<dynamic>;
    Map<String, dynamic>? attemptRow;
    final attemptId = _newAttemptId();
    final row = <String, dynamic>{
      'attemptId': attemptId,
      'userId': user,
      'package': package,
      'answers': <Map<String, String>>[],
      'status': 'inProgress',
      'createdAt': DateTime.now().toUtc().toIso8601String(),
      if (scope != null) 'scope': scope,
    };
    await _mutate(user, (attempts) {
      if (kind == 'ssw') {
        final existing = attempts.where((entry) {
          if (entry['status'] != 'inProgress' || entry['scope'] != scope) {
            return false;
          }
          final existingPackage = entry['package'];
          if (existingPackage is! Map) return false;
          final existingManifest = existingPackage['manifest'];
          return existingManifest is Map &&
              existingManifest['kind'] == kind &&
              existingManifest['lessonId'] == lessonId &&
              existingManifest['contentHash'] ==
                  (package['manifest'] as Map<String, dynamic>)['contentHash'];
        }).firstOrNull;
        if (existing != null) {
          attemptRow = existing;
        } else {
          attempts
            ..removeWhere(
              (entry) =>
                  entry['status'] == 'inProgress' &&
                  entry['package'] is Map &&
                  (entry['package'] as Map)['manifest'] is Map &&
                  ((entry['package'] as Map)['manifest'] as Map)['kind'] ==
                      kind &&
                  ((entry['package'] as Map)['manifest'] as Map)['lessonId'] ==
                      lessonId,
            )
            ..add(row);
          attemptRow = row;
        }
      } else {
        attempts
          ..removeWhere(
            (entry) =>
                entry['status'] == 'inProgress' &&
                entry['package'] is Map &&
                (entry['package'] as Map)['manifest'] is Map &&
                ((entry['package'] as Map)['manifest'] as Map)['kind'] ==
                    kind &&
                ((entry['package'] as Map)['manifest'] as Map)['lessonId'] ==
                    lessonId,
          )
          ..add(row);
        attemptRow = row;
      }
      if (attempts.length > 500) {
        attempts.removeWhere((entry) => entry['status'] == 'synced');
        if (attempts.length > 500) {
          throw StateError('Ada terlalu banyak latihan yang belum tersinkron.');
        }
      }
    });
    _checkAccount(user, generation);
    final storedAnswers = (attemptRow!['answers'] as List<dynamic>)
        .map((value) => Map<String, String>.from(value as Map))
        .toList(growable: false);
    return <String, dynamic>{
      'id': attemptRow!['attemptId'],
      'lessonId': lessonId,
      'imageUrl': _firstString(questions, 'imageUrl'),
      'questions': [
        for (final value in questions)
          if (value is Map<String, dynamic>) _learnerQuestion(value),
      ],
      'storedAnswers': storedAnswers,
    };
  }

  Future<Map<String, dynamic>> _fetchPackage(
    String user,
    String kind,
    String lessonId,
    int generation,
  ) async {
    final requestStartedOffline = isOffline();
    final token = CancelToken();
    _track(user, token);
    try {
      late final Response<Map<String, dynamic>> response;
      try {
        response = await dio.get<Map<String, dynamic>>(
          '/offline/practice/$kind/${Uri.encodeComponent(lessonId)}',
          cancelToken: token,
          options: Options(extra: {'offlineAccountId': user}),
        );
      } on DioException catch (error) {
        final code = error.response?.statusCode;
        if (code == null ||
            (requestStartedOffline && (code == 404 || code == 503))) {
          throw const OfflinePracticeUnavailableException();
        }
        rethrow;
      }
      _checkAccount(user, generation);
      final package = _unwrap(response.data);
      final manifest = package['manifest'];
      final questions = package['questions'];
      if (manifest is! Map<String, dynamic> ||
          manifest['userId'] != user ||
          manifest['kind'] != kind ||
          manifest['lessonId'] != lessonId ||
          questions is! List ||
          questions.isEmpty) {
        throw const FormatException('Paket kuis offline tidak valid.');
      }
      if (_isPremiumManifest(manifest) && !await canReadPremium(user)) {
        // Premium content fetched over the network remains usable through the
        // existing online attempt APIs. Do not retain a local copy unless this
        // device still has a valid, account-bound offline entitlement.
        throw DioException(
          requestOptions: RequestOptions(
            path: '/offline/practice/$kind/$lessonId',
          ),
          response: Response<dynamic>(
            requestOptions: RequestOptions(
              path: '/offline/practice/$kind/$lessonId',
            ),
            statusCode: 503,
          ),
        );
      }
      _checkAccount(user, generation);
      return package;
    } finally {
      _untrack(user, token);
    }
  }

  Future<Map<String, dynamic>> answer(
    String attemptId,
    String questionId,
    String optionId,
  ) async {
    final user = currentUserId();
    if (user == null) throw StateError('Akun tidak tersedia.');
    final generation = _generations[user] ?? 0;
    await _authorizeStoredAttempt(user, attemptId);
    var feedback = <String, dynamic>{};
    await _mutate(user, (attempts) {
      final row = _attempt(attempts, attemptId);
      if (row == null || row['status'] != 'inProgress') {
        throw StateError('Latihan tidak lagi menerima jawaban.');
      }
      final package = row['package'] as Map<String, dynamic>;
      final questions = package['questions'] as List<dynamic>;
      final question = questions
          .whereType<Map<String, dynamic>>()
          .where((value) => value['id'] == questionId)
          .firstOrNull;
      if (question == null) throw ArgumentError.value(questionId, 'questionId');
      final options = question['options'] as List<dynamic>;
      if (!options.whereType<Map<String, dynamic>>().any(
        (value) => value['id'] == optionId,
      )) {
        throw ArgumentError.value(optionId, 'optionId');
      }
      final answers = (row['answers'] as List<dynamic>)
          .map((value) => Map<String, String>.from(value as Map))
          .toList();
      final prior = answers.where((value) => value['questionId'] == questionId);
      if (prior.isNotEmpty && prior.first['optionId'] != optionId) {
        throw StateError('Jawaban untuk soal ini sudah disimpan.');
      }
      if (prior.isEmpty) {
        answers.add({'questionId': questionId, 'optionId': optionId});
        row['answers'] = answers;
      }
      feedback = {
        'questionId': questionId,
        'correct': optionId == question['correctOptionId'],
        'correctOptionId': question['correctOptionId'],
        'explanation': question['explanation'],
      };
    });
    _checkAccount(user, generation);
    return feedback;
  }

  Future<Map<String, dynamic>> complete(String attemptId) async {
    final user = currentUserId();
    if (user == null) throw StateError('Akun tidak tersedia.');
    final generation = _generations[user] ?? 0;
    await _authorizeStoredAttempt(user, attemptId);
    Map<String, dynamic>? localResult;
    await _mutate(user, (attempts) {
      final row = _attempt(attempts, attemptId);
      if (row == null) throw StateError('Latihan tidak ditemukan.');
      if (row['status'] == 'synced') {
        localResult = Map<String, dynamic>.from(row['serverResult'] as Map);
        return;
      }
      final package = row['package'] as Map<String, dynamic>;
      final answers = (row['answers'] as List<dynamic>)
          .map((value) => Map<String, String>.from(value as Map))
          .toList();
      final questions = package['questions'] as List<dynamic>;
      final isSswPractice =
          (package['manifest'] as Map<String, dynamic>)['kind'] == 'ssw';
      if (answers.isEmpty ||
          (!isSswPractice && answers.length != questions.length)) {
        throw StateError('Jawab semua soal sebelum menyelesaikan latihan.');
      }
      localResult = _provisionalResult(attemptId, package, answers);
      row['status'] = 'pending';
      row['localResult'] = localResult;
    });
    _checkAccount(user, generation);
    if (!isOffline()) {
      try {
        await _syncOne(user, attemptId).timeout(syncTimeout);
      } on Object {
        // The encrypted pending record remains available for reconnect retries.
      }
    }
    _checkAccount(user, generation);
    final attempts = await store.read(user);
    final row = _attempt(attempts, attemptId);
    if (row?['status'] == 'synced') {
      return <String, dynamic>{
        ...Map<String, dynamic>.from(row!['serverResult'] as Map),
        'pendingSync': false,
      };
    }
    return <String, dynamic>{...localResult!, 'pendingSync': true};
  }

  Future<void> flush() {
    if (_disposed) return Future<void>.value();
    if (_flushing != null) return _flushing!;
    return _flushing = _flush().whenComplete(() => _flushing = null);
  }

  Future<void> _flush() async {
    final user = currentUserId();
    if (user == null || isOffline()) return;
    final generation = _generations[user] ?? 0;
    final attempts = await store.read(user);
    for (final row in attempts.where((entry) => entry['status'] == 'pending')) {
      if (isOffline() || currentUserId() != user) return;
      if (!await _isAuthorizedRow(user, row)) continue;
      try {
        await _syncOne(user, row['attemptId'] as String);
      } on Object {
        return;
      }
      if (generation != (_generations[user] ?? 0)) return;
    }
  }

  Future<Map<String, dynamic>?> _syncOne(String user, String attemptId) {
    final key = '$user:$attemptId';
    final current = _inflight[key];
    if (current != null) return current;
    final pending = _performSync(user, attemptId);
    _inflight[key] = pending;
    return pending.whenComplete(() {
      if (identical(_inflight[key], pending)) _inflight.remove(key);
    });
  }

  Future<Map<String, dynamic>?> _performSync(
    String user,
    String attemptId,
  ) async {
    if (currentUserId() != user || isOffline()) return null;
    final generation = _generations[user] ?? 0;
    final attempts = await store.read(user);
    final row = _attempt(attempts, attemptId);
    if (row == null || row['status'] != 'pending') {
      return row?['serverResult'] is Map
          ? Map<String, dynamic>.from(row!['serverResult'] as Map)
          : null;
    }
    if (!await _isAuthorizedRow(user, row)) return null;
    final token = CancelToken();
    _track(user, token);
    try {
      final response = await dio
          .post<Map<String, dynamic>>(
            '/offline/practice-attempts/sync',
            data: {
              'attemptId': attemptId,
              'package': row['package'],
              'answers': row['answers'],
            },
            cancelToken: token,
            options: Options(extra: {'offlineAccountId': user}),
          )
          .timeout(
            syncTimeout,
            onTimeout: () {
              token.cancel('Offline practice sync timed out.');
              throw TimeoutException('Offline practice sync timed out.');
            },
          );
      _checkAccount(user, generation);
      final result = _unwrap(response.data);
      if (result['verified'] != true || result['attemptId'] != attemptId) {
        throw const FormatException('Hasil latihan server tidak valid.');
      }
      await _mutate(user, (currentAttempts) {
        final current = _attempt(currentAttempts, attemptId);
        if (current == null || current['status'] == 'synced') return;
        current
          ..['status'] = 'synced'
          ..['serverResult'] = result
          ..remove('package')
          ..remove('answers')
          ..remove('localResult');
      });
      _checkAccount(user, generation);
      final event = OfflinePracticeSyncEvent(
        userId: user,
        attemptId: attemptId,
        result: Map<String, dynamic>.from(result),
      );
      if (!_disposed) _syncEvents.add(event);
      return result;
    } finally {
      _untrack(user, token);
    }
  }

  Future<void> clearAccount(String userId) async {
    if (userId.isEmpty) return;
    _generations[userId] = (_generations[userId] ?? 0) + 1;
    for (final token in _requests[userId] ?? const <CancelToken>{}) {
      token.cancel('Account signed out.');
    }
    await _writes;
    await store.clear(userId);
  }

  Future<void> _authorizeStoredAttempt(String userId, String attemptId) async {
    if (currentUserId() != userId) throw StateError('Akun berubah.');
    final attempts = await store.read(userId);
    final row = _attempt(attempts, attemptId);
    if (row == null || !await _isAuthorizedRow(userId, row)) {
      throw StateError('Akses offline ke latihan ini sudah tidak tersedia.');
    }
    _checkAccount(userId, _generations[userId] ?? 0);
  }

  Future<bool> _isAuthorizedRow(String userId, Map<String, dynamic> row) async {
    if (row['userId'] != userId) return false;
    final package = row['package'];
    if (package is! Map<String, dynamic>) {
      // A successfully synced result has had its private package removed.
      return row['status'] == 'synced';
    }
    final manifest = package['manifest'];
    if (manifest is! Map<String, dynamic> || manifest['userId'] != userId) {
      return false;
    }
    return !_isPremiumManifest(manifest) || await canReadPremium(userId);
  }

  Future<void> _mutate(
    String user,
    void Function(List<Map<String, dynamic>> attempts) change,
  ) {
    final generation = _generations[user] ?? 0;
    final operation = _writes.then((_) async {
      if (_disposed || (_generations[user] ?? 0) != generation) return;
      final attempts = await store.read(user);
      if (_disposed || (_generations[user] ?? 0) != generation) return;
      change(attempts);
      if (_disposed || (_generations[user] ?? 0) != generation) return;
      if (utf8.encode(jsonEncode(attempts)).length >
          maxOfflinePracticeOutboxBytes) {
        throw StateError(
          'Antrean latihan terlalu besar. Hubungkan internet untuk menyinkronkan latihan yang tertunda.',
        );
      }
      await store.write(user, attempts);
    });
    _writes = operation.catchError((Object _) {});
    return operation;
  }

  void _checkAccount(String user, int generation) {
    if (_disposed ||
        currentUserId() != user ||
        (_generations[user] ?? 0) != generation) {
      throw StateError('Akun berubah selama latihan berlangsung.');
    }
  }

  void _track(String user, CancelToken token) =>
      (_requests[user] ??= <CancelToken>{}).add(token);

  void _untrack(String user, CancelToken token) {
    _requests[user]?.remove(token);
    if (_requests[user]?.isEmpty ?? false) _requests.remove(user);
  }

  void dispose() {
    _disposed = true;
    for (final tokens in _requests.values) {
      for (final token in tokens) {
        token.cancel('Offline practice service disposed.');
      }
    }
    _requests.clear();
    unawaited(_syncEvents.close());
  }

  static Map<String, dynamic> _unwrap(Map<String, dynamic>? body) {
    final value = body?['data'];
    if (value is Map<String, dynamic>) return value;
    throw const FormatException('Respons latihan offline tidak valid.');
  }

  static Map<String, dynamic> _learnerQuestion(Map<String, dynamic> value) => {
    for (final entry in value.entries)
      if (entry.key != 'correctOptionId' && entry.key != 'explanation')
        entry.key: entry.value,
  };

  static String _firstString(List<dynamic> questions, String key) {
    for (final value in questions.whereType<Map<String, dynamic>>()) {
      final candidate = value[key];
      if (candidate is String && candidate.isNotEmpty) return candidate;
    }
    return '';
  }

  static Map<String, dynamic> _provisionalResult(
    String attemptId,
    Map<String, dynamic> package,
    List<Map<String, String>> answers,
  ) {
    final manifest = package['manifest'] as Map<String, dynamic>;
    final questions = package['questions'] as List<dynamic>;
    final byQuestion = {
      for (final answer in answers) answer['questionId']!: answer['optionId']!,
    };
    final correct = questions
        .whereType<Map<String, dynamic>>()
        .where((question) => byQuestion.containsKey(question['id']))
        .where(
          (question) =>
              byQuestion[question['id']] == question['correctOptionId'],
        )
        .length;
    final kind = manifest['kind'];
    final total = kind == 'ssw' ? answers.length : questions.length;
    final score = total == 0 ? 0 : correct * 100 ~/ total;
    return {
      'attemptId': attemptId,
      'lessonId': manifest['lessonId'],
      'score': score,
      'correctAnswers': correct,
      'totalQuestions': total,
      'passed': kind != 'ssw' && score >= 80,
      'bestScore': 0,
      'progress': 0,
      'canRetry': true,
      'verified': false,
    };
  }

  static Map<String, dynamic>? _attempt(
    List<Map<String, dynamic>> attempts,
    String attemptId,
  ) => attempts.where((row) => row['attemptId'] == attemptId).firstOrNull;

  static bool _supportedKind(String kind) =>
      const {'vocabulary', 'grammar', 'kanji', 'reading', 'ssw'}.contains(kind);

  static bool _isPremiumManifest(Map<String, dynamic> manifest) {
    final level = (manifest['level'] as String? ?? '').toUpperCase();
    return manifest['kind'] == 'ssw' ||
        const {'N3', 'N2', 'N1'}.contains(level);
  }

  static Future<bool> _denyPremium(String _) async => false;

  static String _newAttemptId() {
    final random = Random.secure();
    final bytes = List<int>.generate(16, (_) => random.nextInt(256));
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    final hex = bytes
        .map((byte) => byte.toRadixString(16).padLeft(2, '0'))
        .join();
    return '${hex.substring(0, 8)}-${hex.substring(8, 12)}-'
        '${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}';
  }
}
