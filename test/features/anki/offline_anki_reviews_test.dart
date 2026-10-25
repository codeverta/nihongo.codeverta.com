import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/anki/data/offline_anki_reviews.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test(
    'premium online review works without retaining an offline lease',
    () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      final dio = Dio();
      addTearDown(dio.close);
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.resolve(
              Response(
                requestOptions: options,
                data: {
                  'data': {'saved': true},
                },
              ),
            );
          },
        ),
      );
      final service = OfflineAnkiReviews(
        dio: dio,
        preferences: prefs,
        currentUserId: () => 'alice',
        isOffline: () => false,
        canReadPremium: (_) async => false,
      );
      final result = await service.submit(
        deckId: 'n2m',
        index: 1,
        version: 0,
        rating: 'good',
        requestId: 'online-only',
        premium: true,
      );
      expect(result['saved'], true);
      expect(service.hasPending, false);
    },
  );

  test('a confirmed version conflict removes the stale local review', () async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final dio = Dio();
    addTearDown(dio.close);
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          handler.reject(
            DioException(
              requestOptions: options,
              type: DioExceptionType.badResponse,
              response: Response(requestOptions: options, statusCode: 409),
            ),
          );
        },
      ),
    );
    final service = OfflineAnkiReviews(
      dio: dio,
      preferences: prefs,
      currentUserId: () => 'alice',
      isOffline: () => false,
      canReadPremium: (_) async => true,
    );
    await expectLater(
      service.submit(
        deckId: 'n5',
        index: 1,
        version: 0,
        rating: 'good',
        requestId: 'outdated',
        premium: false,
      ),
      throwsA(isA<DioException>()),
    );
    expect(service.hasPending, false);
    expect(service.pendingCards('n5'), isEmpty);
  });

  test(
    'offline reviews survive recreation and sync with the original request ID',
    () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      var offline = true;
      final submitted = <Map<String, dynamic>>[];
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            submitted.add(Map<String, dynamic>.from(options.data as Map));
            expect(options.extra['offlineAccountId'], 'alice');
            handler.resolve(
              Response(
                requestOptions: options,
                data: {
                  'data': {'saved': true, 'version': 1},
                },
              ),
            );
          },
        ),
      );
      OfflineAnkiReviews service() => OfflineAnkiReviews(
        dio: dio,
        preferences: prefs,
        currentUserId: () => 'alice',
        isOffline: () => offline,
        canReadPremium: (_) async => true,
      );
      final first = service();
      final result = await first.submit(
        deckId: 'n5',
        index: 1,
        version: 0,
        rating: 'good',
        requestId: 'stable-id',
        premium: false,
      );
      expect(result['pendingSync'], true);
      expect(submitted, isEmpty);
      final reopened = service();
      expect(reopened.pendingCards('n5'), {1});
      offline = false;
      await reopened.flush();
      await reopened.flush();
      expect(submitted, hasLength(1));
      expect(submitted.single['requestId'], 'stable-id');
      expect(reopened.hasPending, false);
    },
  );

  test(
    'another account cannot flush queued reviews and logout removes them',
    () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      String? account = 'alice';
      var offline = true;
      var writes = 0;
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            writes++;
            handler.resolve(
              Response(
                requestOptions: options,
                data: {'data': <String, Object?>{}},
              ),
            );
          },
        ),
      );
      final service = OfflineAnkiReviews(
        dio: dio,
        preferences: prefs,
        currentUserId: () => account,
        isOffline: () => offline,
        canReadPremium: (_) async => false,
      );
      await service.submit(
        deckId: 'n5',
        index: 2,
        version: 0,
        rating: 'easy',
        requestId: 'id',
        premium: false,
      );
      account = 'bob';
      offline = false;
      await service.flush();
      expect(writes, 0);
      await service.clearAccount('alice');
      account = 'alice';
      expect(service.hasPending, false);
    },
  );
}
