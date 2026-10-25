import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_cache_interceptor.dart';
import 'package:nihongo_master/core/offline/learning_cache_persistence_memory.dart';
import 'package:nihongo_master/core/offline/learning_downloads.dart';

void main() {
  group('PersistentLearningCache', () {
    test('survives a cold reload and keeps account data isolated', () async {
      final persistence = MemoryLearningCachePersistence();
      final firstRun = await PersistentLearningCache.open(
        persistence: persistence,
        cipher: PlaintextLearningCacheCipher(),
      );
      await firstRun.writeResponse(
        accountId: 'account-a',
        key: '/home',
        body: Uint8List.fromList(utf8.encode('{"private":"A"}')),
        contentType: 'application/json',
        statusCode: 200,
        statusMessage: 'OK',
        headers: const <String, String>{},
        requiresPremium: false,
      );

      final coldReload = await PersistentLearningCache.open(
        persistence: persistence,
        cipher: PlaintextLearningCacheCipher(),
      );
      final ownResponse = await coldReload.readResponse(
        accountId: 'account-a',
        key: '/home',
      );
      final otherResponse = await coldReload.readResponse(
        accountId: 'account-b',
        key: '/home',
      );

      expect(utf8.decode(ownResponse!.body), '{"private":"A"}');
      expect(otherResponse, isNull);
    });

    test(
      'reopens encrypted-cache storage from disk after a cold reload',
      () async {
        final directory = await Directory.systemTemp.createTemp(
          'nihongo-learning-cache-',
        );
        try {
          final firstRun = await PersistentLearningCache.open(
            directoryPath: directory.path,
            cipher: PlaintextLearningCacheCipher(),
          );
          await _writeText(
            firstRun,
            'account-a',
            '/kanji/lessons/kanji-n5-01',
            'disk lesson',
          );
          await firstRun.dispose();

          final coldReload = await PersistentLearningCache.open(
            directoryPath: directory.path,
            cipher: PlaintextLearningCacheCipher(),
          );
          final response = await coldReload.readResponse(
            accountId: 'account-a',
            key: '/kanji/lessons/kanji-n5-01',
          );

          expect(utf8.decode(response!.body), 'disk lesson');
          await coldReload.dispose();
        } finally {
          await directory.delete(recursive: true);
        }
      },
    );

    test(
      'keeps old content readable after the 24 hour refresh window',
      () async {
        var now = DateTime.utc(2026, 10);
        final cache = await PersistentLearningCache.open(
          persistence: MemoryLearningCachePersistence(),
          cipher: PlaintextLearningCacheCipher(),
          clock: () => now,
        );
        await _writeText(
          cache,
          'user',
          '/grammar/lessons/grammar-n5-01',
          'lesson',
        );

        now = now.add(const Duration(days: 12));
        final stale = await cache.readResponse(
          accountId: 'user',
          key: '/grammar/lessons/grammar-n5-01',
        );

        expect(stale, isNotNull);
        expect(stale!.stale, isTrue);
        expect(utf8.decode(stale.body), 'lesson');
      },
    );

    test(
      'evicts media before least-used auto text and preserves pinned text',
      () async {
        var now = DateTime.utc(2026, 10);
        final cache = await PersistentLearningCache.open(
          persistence: MemoryLearningCachePersistence(),
          cipher: PlaintextLearningCacheCipher(),
          maxBytes: 8,
          clock: () => now,
        );
        await _writeText(cache, 'user', '/first', '1234');
        now = now.add(const Duration(minutes: 1));
        await _writeText(cache, 'user', '/second', '5678');
        now = now.add(const Duration(minutes: 1));
        await cache.readResponse(accountId: 'user', key: '/first');
        now = now.add(const Duration(minutes: 1));

        expect(
          await cache.writeAsset(
            userId: 'user',
            url: 'https://cdn.test/lesson.webp',
            bytes: Uint8List.fromList(<int>[1, 2, 3, 4]),
          ),
          isFalse,
        );
        expect(
          await cache.readResponse(accountId: 'user', key: '/second'),
          isNotNull,
        );
        expect(
          await cache.readAsset(
            userId: 'user',
            url: 'https://cdn.test/lesson.webp',
          ),
          isNull,
        );

        await cache.clearAll();
        await _writeText(cache, 'user', '/pinned', '1234', pinned: true);
        await _writeText(cache, 'user', '/automatic', '5678');
        expect(
          await cache.writeResponse(
            accountId: 'user',
            key: '/new-pinned',
            body: Uint8List.fromList(utf8.encode('12345')),
            contentType: 'application/json',
            statusCode: 200,
            statusMessage: 'OK',
            headers: const <String, String>{},
            requiresPremium: false,
            pinned: true,
          ),
          isFalse,
        );
        expect(
          await cache.readResponse(accountId: 'user', key: '/pinned'),
          isNotNull,
        );
        expect(
          await cache.readResponse(accountId: 'user', key: '/automatic'),
          isNotNull,
        );
        expect(
          await cache.writeAsset(
            userId: 'user',
            url: 'https://cdn.test/extra.webp',
            bytes: Uint8List.fromList(<int>[1, 2, 3, 4]),
          ),
          isFalse,
        );
        expect(
          await cache.readResponse(accountId: 'user', key: '/pinned'),
          isNotNull,
        );
        expect(
          await cache.readResponse(accountId: 'user', key: '/automatic'),
          isNotNull,
        );
        expect((await cache.stats()).totalBytes, 8);
      },
    );

    test(
      'clears only automatic media, retaining pinned media and text',
      () async {
        final cache = await PersistentLearningCache.open(
          persistence: MemoryLearningCachePersistence(),
          cipher: PlaintextLearningCacheCipher(),
        );
        await _writeText(
          cache,
          'user',
          '/grammar/lessons/grammar-n5-01',
          'text',
        );
        await cache.writeAsset(
          userId: 'user',
          url: 'https://cdn.test/automatic.webp',
          bytes: Uint8List.fromList(<int>[1]),
        );
        await cache.writeAsset(
          userId: 'user',
          url: 'https://cdn.test/pinned.webp',
          bytes: Uint8List.fromList(<int>[2]),
          pinned: true,
        );

        await cache.clearAutomaticMedia();

        expect(
          await cache.readResponse(
            accountId: 'user',
            key: '/grammar/lessons/grammar-n5-01',
          ),
          isNotNull,
        );
        expect(
          await cache.readAsset(
            userId: 'user',
            url: 'https://cdn.test/automatic.webp',
          ),
          isNull,
        );
        expect(
          (await cache.readAsset(
            userId: 'user',
            url: 'https://cdn.test/pinned.webp',
          ))?.pinned,
          isTrue,
        );
      },
    );

    test('clearAccount removes only that account and its media', () async {
      final cache = await PersistentLearningCache.open(
        persistence: MemoryLearningCachePersistence(),
        cipher: PlaintextLearningCacheCipher(),
      );
      await _writeText(cache, 'account-a', '/home', 'a');
      await _writeText(cache, 'account-b', '/home', 'b');
      await cache.writeAsset(
        userId: 'account-a',
        url: 'https://cdn.test/a.webp',
        bytes: Uint8List.fromList(<int>[1]),
      );

      await cache.clearAccount('account-a');

      expect(
        await cache.readResponse(accountId: 'account-a', key: '/home'),
        isNull,
      );
      expect(
        await cache.readAsset(
          userId: 'account-a',
          url: 'https://cdn.test/a.webp',
        ),
        isNull,
      );
      expect(
        utf8.decode(
          (await cache.readResponse(
            accountId: 'account-b',
            key: '/home',
          ))!.body,
        ),
        'b',
      );
    });
  });

  group('LearningCacheInterceptor', () {
    test('stores learning content separately for each app language', () async {
      final cache = await _newCache();
      var networkCalls = 0;
      final dio = _dio((options) async {
        networkCalls++;
        final language = options.headers['Accept-Language'] as String;
        return _jsonResponse('{"data":{"title":"$language lesson"}}');
      });
      dio.interceptors.add(_interceptor(dio, cache));
      const path = '/vocabulary/lessons/vocabulary-n5-01';

      final english = await dio.get<Map<String, dynamic>>(
        path,
        options: Options(headers: {'Accept-Language': 'en'}),
      );
      final englishCached = await dio.get<Map<String, dynamic>>(
        path,
        options: Options(headers: {'Accept-Language': 'en'}),
      );
      final indonesian = await dio.get<Map<String, dynamic>>(
        path,
        options: Options(headers: {'Accept-Language': 'id'}),
      );

      expect(english.data?['data']['title'], 'en lesson');
      expect(englishCached.data?['data']['title'], 'en lesson');
      expect(indonesian.data?['data']['title'], 'id lesson');
      expect(networkCalls, 2);
      expect(
        await cache.readResponse(accountId: 'user-a', key: '$path\nlocale=en'),
        isNotNull,
      );
      expect(
        await cache.readResponse(accountId: 'user-a', key: '$path\nlocale=id'),
        isNotNull,
      );
      dio.close(force: true);
    });

    test(
      'does not cache a response when logout wins cache initialization',
      () async {
        final cache = await _newCache();
        final delayedCache = Completer<PersistentLearningCache>();
        final response = Completer<ResponseBody>();
        final networkStarted = Completer<void>();
        final accountCapturedForResponse = Completer<void>();
        var userId = 'user-a';
        var userIdReads = 0;
        final dio = _dio((_) {
          networkStarted.complete();
          return response.future;
        });
        dio.interceptors.add(
          LearningCacheInterceptor(
            dio: dio,
            cache: delayedCache.future,
            currentUserId: () {
              userIdReads++;
              // A bypassed request does not wait for cache initialization.
              // The third read is onResponse's account snapshot, immediately
              // before it awaits the delayed cache future.
              if (userIdReads == 3 && !accountCapturedForResponse.isCompleted) {
                accountCapturedForResponse.complete();
              }
              return userId;
            },
            canReadPremium: (_) => false,
          ),
        );

        final request = dio.get<Map<String, dynamic>>(
          '/grammar/lessons/grammar-n5-01',
          options: Options(
            extra: <String, dynamic>{
              offlineAccountIdExtraKey: 'user-a',
              offlineCacheBypassExtraKey: true,
            },
          ),
        );
        await networkStarted.future;
        response.complete(
          _jsonResponse('{"data":{"title":"Lesson","content":{}}}'),
        );
        await accountCapturedForResponse.future;
        userId = 'user-b';
        delayedCache.complete(cache);

        await request;

        expect(
          await cache.readResponse(
            accountId: 'user-a',
            key: '/grammar/lessons/grammar-n5-01',
          ),
          isNull,
        );
        expect(
          await cache.readResponse(
            accountId: guestCacheAccountId,
            key: '/grammar/lessons/grammar-n5-01',
          ),
          isNull,
        );
      },
    );

    test(
      'does not recreate sanitized public content after logout cleanup',
      () async {
        final persistence = _GatedIndexWritePersistence();
        final cache = await PersistentLearningCache.open(
          persistence: persistence,
          cipher: PlaintextLearningCacheCipher(),
        );
        var userId = 'user-a';
        final dio = _dio(
          (_) async => _jsonResponse(
            '{"data":{"title":"Lesson","bestScore":70,'
            '"completed":true,"content":{"body":"Public text"}}}',
          ),
        );
        dio.interceptors.add(
          LearningCacheInterceptor(
            dio: dio,
            cache: cache,
            currentUserId: () => userId,
            canReadPremium: (_) => false,
          ),
        );

        final request = dio.get<Map<String, dynamic>>(
          '/grammar/lessons/grammar-n5-01',
        );
        await persistence.firstWriteStarted.future;
        userId = 'user-b';
        final logoutCleanup = cache.clearAccount('user-a');
        persistence.releaseFirstWrite.complete();
        await logoutCleanup;
        await request;

        expect(
          await cache.readResponse(
            accountId: 'user-a',
            key: '/grammar/lessons/grammar-n5-01',
          ),
          isNull,
        );
        expect(
          await cache.readResponse(
            accountId: guestCacheAccountId,
            key: '/grammar/lessons/grammar-n5-01',
          ),
          isNull,
        );
      },
    );

    test(
      'rejects a request stamped for an account that is no longer active',
      () async {
        final cache = await _newCache();
        var networkCalls = 0;
        final dio = _dio((_) async {
          networkCalls++;
          return _jsonResponse('{"data":{}}');
        });
        dio.interceptors.add(
          LearningCacheInterceptor(
            dio: dio,
            cache: cache,
            currentUserId: () => 'user-b',
            canReadPremium: (_) => false,
          ),
        );

        await expectLater(
          dio.get<Map<String, dynamic>>(
            '/vocabulary/lessons/vocabulary-n5-01',
            options: Options(
              extra: <String, dynamic>{offlineAccountIdExtraKey: 'user-a'},
            ),
          ),
          throwsA(
            isA<DioException>().having(
              (error) => error.type,
              'type',
              DioExceptionType.cancel,
            ),
          ),
        );

        expect(networkCalls, 0);
      },
    );

    test(
      'does not resolve a deduplicated request after account switch',
      () async {
        final cache = await _newCache();
        var userId = 'user-a';
        var userIdReads = 0;
        final secondRequestReachedCache = Completer<void>();
        final response = Completer<ResponseBody>();
        final networkStarted = Completer<void>();
        final dio = _dio((_) {
          if (!networkStarted.isCompleted) networkStarted.complete();
          return response.future;
        });
        dio.interceptors.add(
          LearningCacheInterceptor(
            dio: dio,
            cache: cache,
            currentUserId: () {
              userIdReads++;
              if (userIdReads == 6 && !secondRequestReachedCache.isCompleted) {
                secondRequestReachedCache.complete();
              }
              return userId;
            },
            canReadPremium: (_) => false,
          ),
        );
        final options = Options(
          extra: <String, dynamic>{offlineAccountIdExtraKey: 'user-a'},
        );

        final first = dio.get<Map<String, dynamic>>(
          '/vocabulary/lessons/vocabulary-n5-01',
          options: options,
        );
        await networkStarted.future;
        final deduplicated = dio.get<Map<String, dynamic>>(
          '/vocabulary/lessons/vocabulary-n5-01',
          options: options,
        );
        await secondRequestReachedCache.future;
        await Future<void>.delayed(Duration.zero);
        userId = 'user-b';
        response.complete(_jsonResponse('{"data":{"lesson":"user-a"}}'));

        await first;
        await expectLater(
          deduplicated,
          throwsA(
            isA<DioException>().having(
              (error) => error.type,
              'type',
              DioExceptionType.cancel,
            ),
          ),
        );
        expect(
          await cache.readResponse(
            accountId: 'user-b',
            key: '/vocabulary/lessons/vocabulary-n5-01',
          ),
          isNull,
        );
      },
    );

    test('deduplicates concurrent cold GETs', () async {
      final cache = await _newCache();
      var networkCalls = 0;
      final dio = _dio((options) async {
        networkCalls++;
        await Future<void>.delayed(const Duration(milliseconds: 20));
        return _jsonResponse('{"data":{"lesson":"ok"}}');
      });
      dio.interceptors.add(_interceptor(dio, cache));

      final responses = await Future.wait(
        <Future<Response<Map<String, dynamic>>>>[
          dio.get<Map<String, dynamic>>('/vocabulary/lessons/vocabulary-n5-01'),
          dio.get<Map<String, dynamic>>('/vocabulary/lessons/vocabulary-n5-01'),
        ],
      );

      expect(networkCalls, 1);
      expect(
        responses.map((response) => response.data?['data']),
        everyElement(<String, String>{'lesson': 'ok'}),
      );
    });

    test(
      'serves stale content immediately and retains it after an offline refresh',
      () async {
        var now = DateTime.utc(2026, 10);
        final cache = await PersistentLearningCache.open(
          persistence: MemoryLearningCachePersistence(),
          cipher: PlaintextLearningCacheCipher(),
          clock: () => now,
        );
        var offline = false;
        var networkCalls = 0;
        final offlineRefreshStarted = Completer<void>();
        final metadata = <CachedResponseMetadata>[];
        final dio = _dio((options) async {
          networkCalls++;
          if (offline) {
            if (!offlineRefreshStarted.isCompleted) {
              offlineRefreshStarted.complete();
            }
            throw DioException.connectionError(
              requestOptions: options,
              reason: 'offline',
            );
          }
          return _jsonResponse('{"data":{"lesson":"cached"}}');
        });
        dio.interceptors.add(
          _interceptor(dio, cache, onMetadata: metadata.add),
        );

        await dio.get<Map<String, dynamic>>(
          '/vocabulary/lessons/vocabulary-n5-01',
        );
        now = now.add(const Duration(days: 2));
        offline = true;
        final response = await dio.get<Map<String, dynamic>>(
          '/vocabulary/lessons/vocabulary-n5-01',
        );
        await offlineRefreshStarted.future;
        await Future<void>.delayed(const Duration(milliseconds: 30));

        expect(networkCalls, 2);
        expect(response.data?['data'], <String, String>{'lesson': 'cached'});
        expect(response.extra['offlineCache'], containsPair('cached', true));
        expect(response.extra['offlineCache'], containsPair('stale', true));
        expect(metadata.last.offline, isTrue);
        expect(
          utf8.decode(
            (await cache.readResponse(
              accountId: 'user-a',
              key: '/vocabulary/lessons/vocabulary-n5-01',
            ))!.body,
          ),
          '{"data":{"lesson":"cached"}}',
        );
      },
    );

    test('bounds queued stale refreshes to two concurrent requests', () async {
      var now = DateTime.utc(2026, 10);
      final cache = await PersistentLearningCache.open(
        persistence: MemoryLearningCachePersistence(),
        cipher: PlaintextLearningCacheCipher(),
        clock: () => now,
      );
      const paths = <String>[
        '/grammar/lessons/grammar-n5-01',
        '/grammar/lessons/grammar-n5-02',
        '/grammar/lessons/grammar-n5-03',
      ];
      for (final path in paths) {
        await _writeText(cache, 'user-a', path, '{"lesson":"cached"}');
      }
      now = now.add(const Duration(days: 2));

      final releases = <Completer<ResponseBody>>[];
      final twoRefreshesStarted = Completer<void>();
      final threeRefreshesStarted = Completer<void>();
      var active = 0;
      var maxActive = 0;
      final dio = _dio((_) {
        active++;
        maxActive = active > maxActive ? active : maxActive;
        final release = Completer<ResponseBody>();
        releases.add(release);
        if (releases.length == 2 && !twoRefreshesStarted.isCompleted) {
          twoRefreshesStarted.complete();
        }
        if (releases.length == 3 && !threeRefreshesStarted.isCompleted) {
          threeRefreshesStarted.complete();
        }
        return release.future.whenComplete(() => active--);
      });
      dio.interceptors.add(_interceptor(dio, cache));

      final cachedResponses = <Response<Map<String, dynamic>>>[];
      for (final path in paths) {
        cachedResponses.add(await dio.get<Map<String, dynamic>>(path));
      }
      await twoRefreshesStarted.future;
      expect(
        cachedResponses.every(
          (response) =>
              response.extra['offlineCache'] is Map &&
              (response.extra['offlineCache'] as Map)['stale'] == true,
        ),
        isTrue,
      );
      expect(releases, hasLength(2));
      expect(maxActive, 2);

      releases[0].complete(_jsonResponse('{"data":{"lesson":"fresh"}}'));
      await threeRefreshesStarted.future;
      expect(maxActive, 2);
      releases[1].complete(_jsonResponse('{"data":{"lesson":"fresh"}}'));
      releases[2].complete(_jsonResponse('{"data":{"lesson":"fresh"}}'));
      await Future<void>.delayed(const Duration(milliseconds: 30));

      expect(maxActive, 2);
      expect(active, 0);
    });

    test('drops premium fallback after local entitlement is revoked', () async {
      final cache = await _newCache();
      var premium = true;
      var networkCalls = 0;
      final dio = _dio((options) async {
        networkCalls++;
        if (!premium) {
          return _jsonResponse(
            '{"error":{"code":"premium_required"}}',
            status: 402,
          );
        }
        return _jsonResponse('{"data":{"questions":[]}}');
      });
      dio.interceptors.add(_interceptor(dio, cache, isPremium: () => premium));
      await dio.get<Map<String, dynamic>>('/ssw/building-cleaning/practice');
      premium = false;

      await expectLater(
        dio.get<Map<String, dynamic>>('/ssw/building-cleaning/practice'),
        throwsA(isA<DioException>()),
      );

      expect(networkCalls, 2);
      expect(
        await cache.readResponse(
          accountId: 'user-a',
          key: '/ssw/building-cleaning/practice',
        ),
        isNull,
      );
    });

    test(
      'caches SSW signed practice packages as premium account data',
      () async {
        final cache = await _newCache();
        var premium = true;
        var networkCalls = 0;
        const key = '/offline/practice/ssw/nursing-care';
        final dio = _dio((options) async {
          networkCalls++;
          if (!premium) {
            return _jsonResponse(
              '{"error":{"code":"premium_required"}}',
              status: 402,
            );
          }
          return _jsonResponse('{"manifest":{"signed":true},"questions":[]}');
        });
        dio.interceptors.add(
          _interceptor(dio, cache, isPremium: () => premium),
        );

        await dio.get<Map<String, dynamic>>(key);
        final saved = await cache.readResponse(accountId: 'user-a', key: key);
        expect(saved, isNotNull);
        expect(saved!.requiresPremium, isTrue);

        premium = false;
        await expectLater(
          dio.get<Map<String, dynamic>>(key),
          throwsA(isA<DioException>()),
        );
        expect(networkCalls, 2);
        expect(await cache.readResponse(accountId: 'user-a', key: key), isNull);
      },
    );

    test(
      'classifies unusual Anki deck IDs from response level and rechecks lease',
      () async {
        final cache = await _newCache();
        var premium = true;
        var networkCalls = 0;
        final dio = _dio((options) async {
          networkCalls++;
          if (!premium) {
            return _jsonResponse(
              '{"error":{"code":"premium_required"}}',
              status: 402,
            );
          }
          final deckId = options.uri.path.contains('/n2m/')
              ? 'n2m'
              : 'custom-deck';
          return _jsonResponse(
            deckId == 'n2m'
                ? '{"data":{"deckId":"n2m","level":"N2","cards":[]}}'
                : '{"data":{"deckId":"custom-deck",'
                      '"level":"advanced","requiredTier":"premium",'
                      '"cards":[]}}',
          );
        });
        dio.interceptors.add(
          _interceptor(dio, cache, isPremium: () => premium),
        );

        await dio.get<Map<String, dynamic>>('/anki/decks/n2m/study');
        final n2 = await cache.readResponse(
          accountId: 'user-a',
          key: '/anki/decks/n2m/study',
        );
        expect(n2!.requiresPremium, isTrue);
        expect(n2.premiumClassificationKnown, isTrue);

        await dio.get<Map<String, dynamic>>('/anki/decks/custom-deck/study');
        final custom = await cache.readResponse(
          accountId: 'user-a',
          key: '/anki/decks/custom-deck/study',
        );
        expect(custom!.requiresPremium, isTrue);
        expect(custom.premiumClassificationKnown, isTrue);

        premium = false;
        await expectLater(
          dio.get<Map<String, dynamic>>('/anki/decks/n2m/study'),
          throwsA(isA<DioException>()),
        );
        expect(networkCalls, 3);
        expect(
          await cache.readResponse(
            accountId: 'user-a',
            key: '/anki/decks/n2m/study',
          ),
          isNull,
        );
      },
    );

    test(
      'fails closed for unknown Anki metadata and allows explicit N4/N5',
      () async {
        final cache = await _newCache();
        await _writeText(
          cache,
          'user-a',
          '/anki/decks/n2m/study',
          '{"data":{"deckId":"n2m","cards":[{"front":"cached"}]}}',
        );
        var offline = false;
        var networkCalls = 0;
        final dio = _dio((options) async {
          networkCalls++;
          if (offline) {
            throw DioException.connectionError(
              requestOptions: options,
              reason: 'offline',
            );
          }
          final level = options.uri.path.contains('n4m') ? 'N4' : 'N5';
          return _jsonResponse(
            '{"data":{"level":"$level","requiredTier":"free",'
            '"cards":[]}}',
          );
        });
        dio.interceptors.add(_interceptor(dio, cache, isPremium: () => false));

        offline = true;
        await expectLater(
          dio.get<Map<String, dynamic>>('/anki/decks/n2m/study'),
          throwsA(isA<DioException>()),
        );
        expect(networkCalls, 1);
        expect(
          await cache.readResponse(
            accountId: 'user-a',
            key: '/anki/decks/n2m/study',
          ),
          isNull,
        );

        offline = false;
        await dio.get<Map<String, dynamic>>('/anki/decks/n4m/study');
        await dio.get<Map<String, dynamic>>('/anki/decks/n5m/study');
        offline = true;
        final n4 = await dio.get<Map<String, dynamic>>('/anki/decks/n4m/study');
        final n5 = await dio.get<Map<String, dynamic>>('/anki/decks/n5m/study');

        expect(n4.data?['data'], containsPair('level', 'N4'));
        expect(n5.data?['data'], containsPair('level', 'N5'));
        expect(networkCalls, 3);
        expect(
          (await cache.readResponse(
            accountId: 'user-a',
            key: '/anki/decks/n4m/study',
          ))!.requiresPremium,
          isFalse,
        );
      },
    );

    test(
      'import and delete invalidate personal deck cache only for the owner',
      () async {
        final cache = await _newCache();
        final dio = _dio((_) async => _jsonResponse('{"data":{}}'));
        dio.interceptors.add(_interceptor(dio, cache));
        for (final method in ['POST', 'DELETE']) {
          for (final account in ['user-a', 'user-b']) {
            await _writeText(cache, account, '/anki/decks', 'catalog');
            await _writeText(
              cache,
              account,
              '/anki/decks/user-example/study',
              'cards',
            );
          }
          await _writeText(cache, 'user-a', '/grammar/course?level=N5', 'keep');
          await dio.request<Map<String, dynamic>>(
            method == 'POST' ? '/anki/import' : '/anki/decks/user-example',
            options: Options(
              method: method,
              extra: {'offlineAccountId': 'user-a', 'offlineCacheBypass': true},
            ),
          );
          expect(
            await cache.readResponse(accountId: 'user-a', key: '/anki/decks'),
            isNull,
          );
          expect(
            await cache.readResponse(
              accountId: 'user-a',
              key: '/anki/decks/user-example/study',
            ),
            isNull,
          );
          expect(
            await cache.readResponse(accountId: 'user-b', key: '/anki/decks'),
            isNotNull,
          );
          expect(
            await cache.readResponse(
              accountId: 'user-a',
              key: '/grammar/course?level=N5',
            ),
            isNotNull,
          );
        }
      },
    );

    test(
      'successful lesson completion invalidates progress but retains practice content',
      () async {
        final cache = await _newCache();
        const user = 'user-a';
        const lessonKey = '/grammar/lessons/grammar-n5-01';
        await _writeText(cache, user, '/home', 'home');
        await _writeText(cache, user, '/grammar/course?level=N5', 'course');
        await _writeText(cache, user, lessonKey, 'private lesson');
        await _writeText(
          cache,
          guestCacheAccountId,
          lessonKey,
          'public lesson',
        );
        await _writeText(
          cache,
          user,
          '/offline/practice/grammar/grammar-n5-01',
          'package',
        );

        final dio = _dio((_) async => _jsonResponse('{"data":{}}'));
        dio.interceptors.add(_interceptor(dio, cache));
        await dio.post<Map<String, dynamic>>(
          '/grammar/attempts/attempt-1/complete',
        );

        expect(await cache.readResponse(accountId: user, key: '/home'), isNull);
        expect(
          await cache.readResponse(
            accountId: user,
            key: '/grammar/course?level=N5',
          ),
          isNull,
        );
        expect(
          await cache.readResponse(accountId: user, key: lessonKey),
          isNull,
        );
        expect(
          await cache.readResponse(
            accountId: guestCacheAccountId,
            key: lessonKey,
          ),
          isNull,
        );
        expect(
          await cache.readResponse(
            accountId: user,
            key: '/offline/practice/grammar/grammar-n5-01',
          ),
          isNotNull,
        );
      },
    );

    test(
      'retains sanitized public lesson details across account logout',
      () async {
        final cache = await _newCache();
        var userId = 'account-a';
        var online = true;
        final dio = _dio((options) async {
          if (!online) {
            throw DioException.connectionError(
              requestOptions: options,
              reason: 'offline',
            );
          }
          return _jsonResponse(
            '{"data":{"title":"Lesson","bestScore":70,"completed":true,'
            '"progress":{"current":3},"content":{"body":"Public text"}}}',
          );
        });
        dio.interceptors.add(
          LearningCacheInterceptor(
            dio: dio,
            cache: cache,
            currentUserId: () => userId,
            canReadPremium: (_) => false,
          ),
        );

        await dio.get<Map<String, dynamic>>('/grammar/lessons/grammar-n5-01');
        await cache.clearAccount('account-a');
        userId = 'account-b';
        online = false;

        final response = await dio.get<Map<String, dynamic>>(
          '/grammar/lessons/grammar-n5-01',
        );
        final data = response.data!['data'] as Map<String, dynamic>;

        expect(data['title'], 'Lesson');
        expect(data['content'], <String, String>{'body': 'Public text'});
        expect(data['bestScore'], 0);
        expect(data['completed'], isFalse);
        expect(data, isNot(contains('progress')));
        expect(response.extra['offlineCache'], containsPair('cached', true));
      },
    );
  });

  group('LearningDownloads', () {
    test('stops queued requests after a transport failure', () async {
      var requests = 0;
      final cache = await _newCache();
      final dio = _dio((options) async {
        requests++;
        throw DioException.connectionError(
          requestOptions: options,
          reason: 'offline',
        );
      });
      final downloads = LearningDownloads(
        dio: dio,
        cache: cache,
        currentUserId: () => 'user-a',
        canReadPremium: (_) => true,
        canStartDownload: () async => true,
        maxConcurrent: 1,
      );

      final progress = await downloads.preloadCourse(
        type: 'grammar',
        level: 'N5',
        lessonIds: const <String>['grammar-n5-01', 'grammar-n5-02'],
        activeLessonId: 'grammar-n5-02',
      );

      expect(requests, 1);
      expect(progress.failed, 1);
      expect(progress.completed, 0);
      expect(progress.remaining, 3);
      expect(progress.isPreparing, isFalse);
      await downloads.dispose();
    });

    test('continues after unsupported lesson package responses', () async {
      var requests = 0;
      final cache = await _newCache();
      final dio = _dio((options) async {
        requests++;
        return _jsonResponse(
          '{"error":{"code":"not_found"}}',
          status: options.path.contains('/offline/practice/') ? 404 : 200,
        );
      });
      final downloads = LearningDownloads(
        dio: dio,
        cache: cache,
        currentUserId: () => 'user-a',
        canReadPremium: (_) => true,
        maxConcurrent: 1,
      );

      final progress = await downloads.preloadCourse(
        type: 'grammar',
        level: 'N5',
        lessonIds: const <String>['grammar-n5-01', 'grammar-n5-02'],
      );

      expect(requests, 4);
      expect(progress.failed, 2);
      expect(progress.completed, 2);
      await downloads.dispose();
    });

    test(
      'preloads signed SSW packages only with an active premium lease',
      () async {
        final cache = await _newCache();
        var premium = true;
        final requestedPaths = <String>[];
        final dio = _dio((options) async {
          requestedPaths.add(options.uri.path);
          return _jsonResponse('{"manifest":{"signed":true},"questions":[]}');
        });
        final downloads = LearningDownloads(
          dio: dio,
          cache: cache,
          currentUserId: () => 'user-a',
          canReadPremium: (_) async => premium,
          canStartDownload: () async => true,
          maxConcurrent: 1,
        );

        final ready = await downloads.preloadSswSector(
          sectorId: 'nursing-care',
        );
        expect(ready.courseKey, 'ssw:nursing-care');
        expect(ready.completed, 1);
        expect(requestedPaths, <String>[
          '/v1/offline/practice/ssw/nursing-care',
        ]);

        premium = false;
        requestedPaths.clear();
        final unauthorized = await downloads.preloadSswSector(
          sectorId: 'food-service',
        );
        expect(unauthorized.total, 0);
        expect(requestedPaths, isEmpty);
        await downloads.dispose();
      },
    );

    test(
      'treats unknown Anki decks as premium until catalog level is supplied',
      () async {
        final cache = await _newCache();
        var networkCalls = 0;
        final dio = _dio((_) async {
          networkCalls++;
          return _jsonResponse('{"data":{"level":"N5","cards":[]}}');
        });
        final downloads = LearningDownloads(
          dio: dio,
          cache: cache,
          currentUserId: () => 'user-a',
          canReadPremium: (_) async => false,
          canStartDownload: () async => true,
          maxConcurrent: 1,
        );

        final unknown = await downloads.preloadDeck(deckId: 'n5m');
        expect(unknown.total, 0);
        expect(networkCalls, 0);

        final freeN5 = await downloads.preloadDeck(
          deckId: 'n5m',
          level: 'N5',
          requiredTier: 'free',
        );
        expect(freeN5.completed, 2);
        expect(networkCalls, 2);
        await downloads.dispose();
      },
    );
  });
}

Future<PersistentLearningCache> _newCache() => PersistentLearningCache.open(
  persistence: MemoryLearningCachePersistence(),
  cipher: PlaintextLearningCacheCipher(),
);

Future<void> _writeText(
  PersistentLearningCache cache,
  String accountId,
  String key,
  String text, {
  bool pinned = false,
}) async {
  await cache.writeResponse(
    accountId: accountId,
    key: key,
    body: Uint8List.fromList(utf8.encode(text)),
    contentType: 'application/json',
    statusCode: 200,
    statusMessage: 'OK',
    headers: const <String, String>{},
    requiresPremium: false,
    pinned: pinned,
  );
}

Dio _dio(Future<ResponseBody> Function(RequestOptions options) fetch) {
  return Dio(BaseOptions(baseUrl: 'https://api.test/v1'))
    ..httpClientAdapter = _TestAdapter(fetch);
}

LearningCacheInterceptor _interceptor(
  Dio dio,
  PersistentLearningCache cache, {
  bool Function()? isPremium,
  void Function(CachedResponseMetadata metadata)? onMetadata,
}) => LearningCacheInterceptor(
  dio: dio,
  cache: cache,
  currentUserId: () => 'user-a',
  canReadPremium: (_) => isPremium?.call() ?? false,
  onMetadata: onMetadata,
);

ResponseBody _jsonResponse(String body, {int status = 200}) =>
    ResponseBody.fromString(
      body,
      status,
      headers: const <String, List<String>>{
        'content-type': <String>['application/json'],
      },
    );

class _TestAdapter implements HttpClientAdapter {
  _TestAdapter(this._fetch);

  final Future<ResponseBody> Function(RequestOptions options) _fetch;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) => _fetch(options);

  @override
  void close({bool force = false}) {}
}

class _GatedIndexWritePersistence extends MemoryLearningCachePersistence {
  final Completer<void> firstWriteStarted = Completer<void>();
  final Completer<void> releaseFirstWrite = Completer<void>();
  bool _hasBlockedFirstWrite = false;

  @override
  Future<void> writeIndex(String contents) async {
    if (!_hasBlockedFirstWrite) {
      _hasBlockedFirstWrite = true;
      firstWriteStarted.complete();
      await releaseFirstWrite.future;
    }
    await super.writeIndex(contents);
  }
}
