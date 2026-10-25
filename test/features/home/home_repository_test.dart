import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/home/data/home_repository.dart';

void main() {
  test(
    'cached dashboard keeps offline metadata and the last saved progress',
    () async {
      final dio = Dio(BaseOptions(baseUrl: 'https://api.example.test/v1'));
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.resolve(
              Response<Map<String, dynamic>>(
                requestOptions: options,
                data: _response,
                extra: {
                  'offlineCache': {
                    'cached': true,
                    'offline': true,
                    'updatedAt': '2026-10-23T01:00:00Z',
                  },
                },
              ),
            );
          },
        ),
      );
      final saved = await HomeRepository(dio).fetchDashboard();
      expect(saved.fromCache, true);
      expect(saved.offline, true);
      expect(saved.cachedAt, DateTime.utc(2026, 10, 23, 1));
      expect(saved.tracks.single.progress, 35);
    },
  );

  test('fetchDashboard parses the complete home API response', () async {
    late RequestOptions capturedRequest;
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:8080/v1'))
      ..interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            capturedRequest = options;
            handler.resolve(
              Response<Map<String, dynamic>>(
                requestOptions: options,
                statusCode: 200,
                data: _response,
              ),
            );
          },
        ),
      );

    final dashboard = await HomeRepository(dio).fetchDashboard();

    expect(capturedRequest.path, '/home');
    expect(dashboard.isAuthenticated, isTrue);
    expect(dashboard.profile?.displayName, 'API Learner');
    expect(dashboard.dailyGoal.completion, 0.5);
    expect(dashboard.categories.single.name, 'Vocabulary from API');
    expect(dashboard.tracks.single.progress, 35);
    expect(dashboard.exams.single.questionCount, 25);
    expect(dashboard.plans.single.billingType, 'lifetime');
    expect(dashboard.achievements.single.target, 1);
    expect(
      dashboard.banners[0].imageUrl,
      'http://localhost:8080/home-banner-images/photo.png',
    );
    expect(dashboard.banners[1].imageUrl, 'https://cdn.example.com/ssw.jpg');
    expect(dashboard.banners[2].imageUrl, '');
  });

  test('fetchDashboard rejects an invalid envelope', () async {
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:8080/v1'))
      ..interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) => handler.resolve(
            Response<Map<String, dynamic>>(
              requestOptions: options,
              statusCode: 200,
              data: const {'unexpected': true},
            ),
          ),
        ),
      );

    expect(HomeRepository(dio).fetchDashboard(), throwsA(isA<HomeException>()));
  });
}

const _response = <String, dynamic>{
  'data': <String, dynamic>{
    'banners': [
      {
        'id': 'learn',
        'title': 'Belajar',
        'imageUrl': '/home-banner-images/photo.png',
        'action': 'learn',
        'buttonLabel': 'Belajar',
      },
      {
        'id': 'ssw',
        'title': 'SSW',
        'imageUrl': 'https://cdn.example.com/ssw.jpg',
        'action': 'ssw',
        'buttonLabel': 'SSW',
      },
      {
        'id': 'jobs',
        'title': 'Kerja',
        'imageUrl': '',
        'action': 'jobs',
        'buttonLabel': 'Lowongan',
      },
    ],
    'isAuthenticated': true,
    'profile': <String, dynamic>{
      'id': 'api-user',
      'displayName': 'API Learner',
      'email': 'api@example.com',
      'target': 'JLPT N5',
      'level': 1,
      'xp': 50,
      'streakDays': 2,
      'entitlement': <String, dynamic>{'tier': 'free', 'isLifetime': false},
    },
    'dailyGoal': <String, dynamic>{
      'minutesLearned': 10,
      'targetMinutes': 20,
      'streakDays': 2,
    },
    'categories': <Map<String, dynamic>>[
      {
        'id': 'vocabulary',
        'name': 'Vocabulary from API',
        'description': 'Kosakata',
        'lessonCount': 240,
        'requiredTier': 'free',
      },
    ],
    'tracks': <Map<String, dynamic>>[
      {
        'id': 'jlpt-n5',
        'system': 'JLPT',
        'level': 'N5',
        'title': 'JLPT N5',
        'description': 'Fondasi',
        'progress': 35,
        'requiredTier': 'free',
      },
    ],
    'exams': <Map<String, dynamic>>[
      {
        'id': 'exam-1',
        'system': 'JLPT',
        'level': 'N5',
        'title': 'Diagnostic',
        'questionCount': 25,
        'durationMinutes': 30,
        'requiredTier': 'free',
      },
    ],
    'plans': <Map<String, dynamic>>[
      {
        'id': 'lifetime',
        'name': 'Lifetime',
        'billingType': 'lifetime',
        'priceIdr': 1499000,
        'features': <String>['Semua materi'],
        'recommended': true,
      },
    ],
    'achievements': <Map<String, dynamic>>[
      {
        'id': 'first',
        'name': 'First Step',
        'description': 'Pelajaran pertama',
        'icon': 'footprints',
        'progress': 0,
        'target': 1,
      },
    ],
  },
};
