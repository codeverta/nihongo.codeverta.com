import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/features/learn/grammar/data/grammar_repository.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice_store_io.dart';

void main() {
  test(
    'encrypted local outbox survives store recreation and clears by account',
    () async {
      final directory = await Directory.systemTemp.createTemp(
        'practice-outbox-',
      );
      addTearDown(() => directory.delete(recursive: true));
      final secrets = _MemorySecrets();
      final row = <String, dynamic>{
        'attemptId': 'attempt-1',
        'answers': [
          {'questionId': 'q1', 'optionId': 'private-answer'},
        ],
      };
      final first = EncryptedOfflinePracticeLocalStore(
        directoryPath: directory.path,
        secretStorage: secrets,
      );
      await first.write('user-a', [row]);

      final encrypted = base64UrlEncode(
        await Directory(
          directory.path,
        ).listSync().whereType<File>().single.readAsBytes(),
      );
      expect(encrypted, isNot(contains('private-answer')));

      final restarted = EncryptedOfflinePracticeLocalStore(
        directoryPath: directory.path,
        secretStorage: secrets,
      );
      expect(await restarted.read('user-a'), equals([row]));
      await restarted.clear('user-a');
      expect(await restarted.read('user-a'), isEmpty);
      expect(await secrets.read('offline_practice_key_v1_user-a'), isNull);
    },
  );

  test(
    'online quiz falls back to legacy routes when signing is unavailable',
    () async {
      var legacyCalls = 0;
      final dio = Dio(BaseOptions(baseUrl: 'https://test.invalid/v1'))
        ..interceptors.add(
          InterceptorsWrapper(
            onRequest: (options, handler) {
              if (options.path.contains('/offline/practice/')) {
                handler.reject(
                  DioException(
                    requestOptions: options,
                    response: Response<dynamic>(
                      requestOptions: options,
                      statusCode: 503,
                    ),
                  ),
                );
                return;
              }
              legacyCalls++;
              final path = options.path;
              final data = path.endsWith('/attempts')
                  ? {
                      'id': 'server-attempt',
                      'questions': [
                        {
                          'id': 'q1',
                          'prompt': 'Pick one.',
                          'options': [
                            {'id': 'o1', 'text': 'One'},
                          ],
                        },
                      ],
                    }
                  : path.endsWith('/answers')
                  ? {
                      'questionId': 'q1',
                      'correct': true,
                      'correctOptionId': 'o1',
                      'explanation': 'Correct.',
                    }
                  : {
                      'score': 100,
                      'correctAnswers': 1,
                      'totalQuestions': 1,
                      'passed': true,
                      'bestScore': 100,
                      'progress': 100,
                    };
              handler.resolve(
                Response<Map<String, dynamic>>(
                  requestOptions: options,
                  statusCode: 200,
                  data: {'data': data},
                ),
              );
            },
          ),
        );
      final service = OfflinePracticeService(
        dio: dio,
        store: _MemoryPracticeStore(),
        currentUserId: () => 'user-a',
        isOffline: () => false,
      );
      final repository = GrammarRepository(dio, offlinePractice: service);

      final attempt = await repository.startAttempt('grammar-n5-1');
      expect(attempt.id, 'server-attempt');
      final feedback = await repository.answer('server-attempt', 'q1', 'o1');
      expect(feedback.correct, isTrue);
      final result = await repository.complete('server-attempt');
      expect(result.verified, isTrue);
      expect(legacyCalls, 3);
      service.dispose();
    },
  );

  test(
    'durable offline attempt syncs after restart with a verified result',
    () async {
      const currentUser = 'user-a';
      var offline = true;
      final store = _MemoryPracticeStore();
      final server = _PracticeServer();
      final first = _service(
        store: store,
        server: server,
        user: () => currentUser,
        offline: () => offline,
      );

      final attempt = await first.start('grammar', 'grammar-n5-1');
      final question =
          (attempt['questions'] as List).first as Map<String, dynamic>;
      final options = question['options'] as List;
      final correctOptionId =
          (options.first as Map<String, dynamic>)['id'] as String;
      await first.answer(
        attempt['id'] as String,
        question['id'] as String,
        correctOptionId,
      );
      final local = await first.complete(attempt['id'] as String);
      expect(local['verified'], isFalse);
      expect(local['pendingSync'], isTrue);
      expect(local['passed'], isTrue);
      expect(await first.pendingCount, 1);
      first.dispose();

      // A new service instance represents an application restart. The encrypted
      // store survives and contains the package snapshot plus selected option ID.
      offline = false;
      final restarted = _service(
        store: store,
        server: server,
        user: () => currentUser,
        offline: () => offline,
      );
      final events = <OfflinePracticeSyncEvent>[];
      final subscription = restarted.syncEvents.listen(events.add);
      await restarted.flush();

      expect(await restarted.pendingCount, 0);
      expect(server.syncRequests, 1);
      expect(events, hasLength(1));
      expect(events.single.result['verified'], isTrue);
      expect(events.single.attemptId, attempt['id']);
      await restarted.flush();
      expect(server.syncRequests, 1, reason: 'synced attempts are not resent');

      await subscription.cancel();
      restarted.dispose();
    },
  );

  test(
    'attempts are isolated by account and logout clears that account queue',
    () async {
      var currentUser = 'user-a';
      var offline = true;
      final store = _MemoryPracticeStore();
      final server = _PracticeServer();
      final service = _service(
        store: store,
        server: server,
        user: () => currentUser,
        offline: () => offline,
      );
      final attempt = await service.start('grammar', 'grammar-n5-1');
      final question =
          (attempt['questions'] as List).first as Map<String, dynamic>;
      final option =
          ((question['options'] as List).first as Map<String, dynamic>)['id']
              as String;
      await service.answer(
        attempt['id'] as String,
        question['id'] as String,
        option,
      );
      await service.complete(attempt['id'] as String);

      currentUser = 'user-b';
      offline = false;
      expect(await service.pendingCount, 0);
      await service.flush();
      expect(server.syncRequests, 0);

      await service.clearAccount('user-a');
      currentUser = 'user-a';
      expect(await service.pendingCount, 0);
      await service.flush();
      expect(server.syncRequests, 0);
      service.dispose();
    },
  );

  test(
    'premium attempt feedback and queued answers stop when offline access expires',
    () async {
      var online = false;
      var canReadPremium = true;
      final store = _MemoryPracticeStore();
      final server = _PracticeServer();
      final service = _service(
        store: store,
        server: server,
        user: () => 'premium-user',
        offline: () => !online,
        canReadPremium: (_) async => canReadPremium,
        level: 'N3',
      );

      final attempt = await service.start('grammar', 'grammar-n5-1');
      final question =
          (attempt['questions'] as List).first as Map<String, dynamic>;
      await service.answer(
        attempt['id'] as String,
        question['id'] as String,
        'grammar-n5-1-o1',
      );
      final local = await service.complete(attempt['id'] as String);
      expect(local['pendingSync'], isTrue);

      // Simulate lease expiry or a successful server-side refund snapshot.
      canReadPremium = false;
      expect(await service.ownsAttempt(attempt['id'] as String), isFalse);
      await expectLater(
        service.answer(
          attempt['id'] as String,
          question['id'] as String,
          'grammar-n5-1-o1',
        ),
        throwsStateError,
      );
      await expectLater(
        service.complete(attempt['id'] as String),
        throwsStateError,
      );

      online = true;
      await service.flush();
      expect(server.syncRequests, 0);
      expect(await service.pendingCount, 1);
      service.dispose();
    },
  );

  test(
    'SSW signed practice supports partial module attempts without a pass',
    () async {
      var online = false;
      final store = _MemoryPracticeStore();
      final server = _PracticeServer();
      final service = _service(
        store: store,
        server: server,
        user: () => 'ssw-user',
        offline: () => !online,
        canReadPremium: (_) async => true,
        kind: 'ssw',
        lessonId: 'building-cleaning',
        level: 'SSW',
      );

      final bank = await service.loadBank('ssw', 'building-cleaning');
      expect(bank.single['moduleId'], 'cleaning');
      final attempt = await service.start(
        'ssw',
        'building-cleaning',
        scope: 'cleaning',
      );
      final question =
          (attempt['questions'] as List).first as Map<String, dynamic>;
      await service.answer(
        attempt['id'] as String,
        question['id'] as String,
        'ssw-q1-o1',
      );
      final result = await service.complete(attempt['id'] as String);
      expect(result['verified'], isFalse);
      expect(result['passed'], isFalse);
      expect(result['totalQuestions'], 1);
      expect(result['pendingSync'], isTrue);
      online = true;
      await service.flush();
      expect(server.syncRequests, 1);
      expect(await service.pendingCount, 0);
      service.dispose();
    },
  );
}

OfflinePracticeService _service({
  required OfflinePracticeLocalStore store,
  required _PracticeServer server,
  required String? Function() user,
  required bool Function() offline,
  Future<bool> Function(String userId)? canReadPremium,
  String kind = 'grammar',
  String lessonId = 'grammar-n5-1',
  String level = 'N5',
}) {
  final dio = Dio(BaseOptions(baseUrl: 'https://test.invalid/v1'))
    ..interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (options.path.contains('/offline/practice-attempts/sync')) {
            server.syncRequests++;
            final body = Map<String, dynamic>.from(options.data as Map);
            final attemptId = body['attemptId'] as String;
            final package = body['package'] as Map<String, dynamic>;
            final manifest = package['manifest'] as Map<String, dynamic>;
            final questions = package['questions'] as List<dynamic>;
            final answers = body['answers'] as List<dynamic>;
            final correctAnswers = answers.where((answer) {
              final value = answer as Map<String, dynamic>;
              final question = questions
                  .whereType<Map<String, dynamic>>()
                  .firstWhere((item) => item['id'] == value['questionId']);
              return value['optionId'] == question['correctOptionId'];
            }).length;
            final ssw = manifest['kind'] == 'ssw';
            final totalQuestions = ssw ? answers.length : questions.length;
            final score = correctAnswers * 100 ~/ totalQuestions;
            handler.resolve(
              Response<Map<String, dynamic>>(
                requestOptions: options,
                statusCode: 200,
                data: {
                  'data': {
                    'attemptId': attemptId,
                    'lessonId': manifest['lessonId'],
                    'score': score,
                    'correctAnswers': correctAnswers,
                    'totalQuestions': totalQuestions,
                    'passed': !ssw && score >= 80,
                    'bestScore': score,
                    'progress': ssw ? 0 : score,
                    'canRetry': true,
                    'verified': true,
                  },
                },
              ),
            );
            return;
          }
          final userId = user();
          handler.resolve(
            Response<Map<String, dynamic>>(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'manifest': {
                    'schemaVersion': 1,
                    'packageId': 'package-v1',
                    'version': 'v1-test',
                    'contentHash': 'test-hash',
                    'kind': kind,
                    'lessonId': lessonId,
                    'level': level,
                    'userId': userId,
                    'issuedAt': DateTime.utc(2026).toIso8601String(),
                    'signature': 'server-signature',
                  },
                  'questions': [
                    {
                      'id': kind == 'ssw' ? 'ssw-q1' : '$lessonId-q1',
                      if (kind == 'ssw') 'moduleId': 'cleaning',
                      if (kind == 'ssw') 'group': 'cleaning',
                      if (kind == 'ssw') 'lessonId': lessonId,
                      'prompt': 'Choose the correct answer.',
                      'options': [
                        {
                          'id': kind == 'ssw' ? 'ssw-q1-o1' : '$lessonId-o1',
                          'text': 'Correct',
                        },
                        {
                          'id': kind == 'ssw' ? 'ssw-q1-o2' : '$lessonId-o2',
                          'text': 'Wrong',
                        },
                      ],
                      'correctOptionId': kind == 'ssw'
                          ? 'ssw-q1-o1'
                          : '$lessonId-o1',
                      'explanation': 'The first option is correct.',
                    },
                  ],
                },
              },
            ),
          );
        },
      ),
    );
  return OfflinePracticeService(
    dio: dio,
    store: store,
    currentUserId: user,
    isOffline: offline,
    canReadPremium: canReadPremium,
  );
}

class _PracticeServer {
  int syncRequests = 0;
}

class _MemorySecrets implements OfflineSecretStorage {
  final Map<String, String> _values = {};

  @override
  Future<void> delete(String key) async {
    _values.remove(key);
  }

  @override
  Future<String?> read(String key) async => _values[key];

  @override
  Future<void> write({required String key, required String value}) async {
    _values[key] = value;
  }
}

class _MemoryPracticeStore implements OfflinePracticeLocalStore {
  final Map<String, String> _records = {};

  @override
  Future<List<Map<String, dynamic>>> read(String userId) async {
    final raw = _records[userId];
    if (raw == null) return [];
    return (jsonDecode(raw) as List)
        .map((row) => Map<String, dynamic>.from(row as Map))
        .toList();
  }

  @override
  Future<void> write(String userId, List<Map<String, dynamic>> attempts) async {
    _records[userId] = jsonEncode(attempts);
  }

  @override
  Future<void> clear(String userId) async {
    _records.remove(userId);
  }
}
