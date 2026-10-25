import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_attempt_screen.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  late KaigoCurriculum course;
  setUpAll(() async {
    course = KaigoCurriculum.fromJson(
      jsonDecode(
            await rootBundle.loadString(
              'assets/content/ssw/building-cleaning.json',
            ),
          )
          as Json,
    );
  });
  test('building cleaning course and premium paths have complete coverage', () {
    expect(course.modules.length, 10);
    expect(course.lessons.length, 30);
    expect(course.lessons.map((l) => l.id).toSet().length, 30);
    for (final l in course.lessons) {
      expect(
        l.sections.map((s) => s.body).join(' ').split(' ').length,
        greaterThan(80),
      );
      expect(l.vocabulary.length, 4);
      expect(l.dialogue.length, 2);
      expect(l.steps.length, 4);
      expect(l.checks.length, 2);
      for (final q in l.checks) {
        expect(q.options.length, 4);
        expect(q.options.toSet().length, 4);
        expect(q.answer, inInclusiveRange(0, 3));
        expect(q.explanation, isNotEmpty);
      }
    }
    expect(
      requiresPremium(Uri.parse('/ssw/building-cleaning/practice')),
      isTrue,
    );
    expect(requiresPremium(Uri.parse('/exams/ssw-kaigo/simulation')), isTrue);
    expect(
      requiresPremium(Uri.parse('/ssw/building-cleaning/lessons/cleaning-01')),
      isFalse,
    );
  });

  test(
    'English SSW curricula use a translated asset with no Indonesian fallback',
    () {
      expect(
        sswCurriculumAssetPath(sector: 'kaigo', languageCode: 'en'),
        'assets/content/ssw/en/kaigo.json',
      );
      expect(
        sswCurriculumAssetPath(sector: 'food-service', languageCode: 'en'),
        'assets/content/ssw/en/food-service.json',
      );
      expect(
        sswCurriculumAssetPath(sector: 'kaigo', languageCode: 'id'),
        'assets/content/ssw/kaigo.json',
      );
    },
  );

  test('all expanded SSW fields have sourced, readable curricula', () async {
    const sectors = [
      'automobile-maintenance',
      'aviation',
      'accommodation',
      'road-transport',
      'railway',
      'logistics-warehouse',
      'food-manufacturing',
      'food-service',
      'forestry',
      'wood-industry',
      'resource-recycling',
    ];
    final japaneseText = RegExp(r'[\u3040-\u30ff\u3400-\u9fff]');
    expect(
      sswActiveSectorIds,
      equals(sswSectors.map((sector) => sector.id).toSet()),
    );

    for (final sector in sectors) {
      expect(sswActiveSectorIds, contains(sector));
      final raw =
          jsonDecode(
                await rootBundle.loadString('assets/content/ssw/$sector.json'),
              )
              as Json;
      final course = KaigoCurriculum.fromJson(raw);
      expect(course.modules, hasLength(4), reason: sector);
      expect(course.lessons, hasLength(12), reason: sector);
      expect(raw['sources'], isNotEmpty, reason: '$sector source attribution');
      for (final module in course.modules) {
        expect(
          japaneseText.hasMatch('${module.title} ${module.description}'),
          isFalse,
          reason: '$sector/${module.id}: headings are plain text',
        );
      }
      for (final source in raw['sources'] as List) {
        final url = (source as Json)['url'] as String;
        expect(Uri.parse(url).scheme, 'https', reason: sector);
      }

      for (final lesson in course.lessons) {
        expect(lesson.checks, hasLength(3), reason: '$sector/${lesson.id}');
        expect(lesson.vocabulary, isNotEmpty, reason: '$sector/${lesson.id}');
        expect(lesson.dialogue, isNotEmpty, reason: '$sector/${lesson.id}');
        for (final word in [...lesson.vocabulary, ...lesson.dialogue]) {
          expect(word.reading, isNotEmpty, reason: '$sector/${lesson.id}');
        }

        final plainText = [
          lesson.title,
          ...lesson.objectives,
          lesson.scenario,
          lesson.response,
          lesson.caution,
          ...lesson.steps,
          ...lesson.sections.map((section) => section.body),
          ...lesson.checks.expand(
            (check) => [check.prompt, check.explanation, ...check.options],
          ),
        ].join(' ');
        expect(
          japaneseText.hasMatch(plainText),
          isFalse,
          reason:
              '$sector/${lesson.id}: Japanese text belongs in vocabulary/dialogue',
        );
      }
    }
  });

  test('SSW repositories send the correct distinct exam IDs', () async {
    for (final level in ['SSW-KAIGO', 'SSW-KAIGO-JP', 'SSW-CLEANING']) {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (o, h) {
            expect(o.data, {'examId': '${level.toLowerCase()}-simulation-1'});
            h.resolve(
              Response<Json>(
                requestOptions: o,
                data: {
                  'data': {'id': 'attempt'},
                },
              ),
            );
          },
        ),
      );
      expect(await SimulationRepository(dio, level: level).start(), 'attempt');
    }
  });

  testWidgets('cleaning phone reader saves progress independently from Kaigo', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(320, 740);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final router = GoRouter(
      initialLocation: '/ssw/building-cleaning',
      routes: [
        GoRoute(
          path: '/ssw/building-cleaning',
          builder: (_, __) =>
              const SswSectorScreen(sectorId: 'building-cleaning'),
        ),
        GoRoute(
          path: '/ssw/building-cleaning/lessons/:lessonId',
          builder: (_, s) => KaigoLessonScreen(
            sectorId: 'building-cleaning',
            lessonId: s.pathParameters['lessonId']!,
          ),
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(prefs),
          sswCurriculumProvider(
            'building-cleaning',
          ).overrideWith((ref) async => course),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(
              const AuthenticationState.authenticated(
                AuthenticatedUser(
                  id: 'cleaner',
                  displayName: 'Cleaner',
                  email: 'c@example.test',
                ),
              ),
            ),
          ),
        ],
        child: MaterialApp.router(
          routerConfig: router,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(
              context,
            ).copyWith(textScaler: const TextScaler.linear(1.4)),
            child: child!,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    Future<void> show(Finder f) async {
      await tester.scrollUntilVisible(
        f,
        300,
        scrollable: find.byType(Scrollable).first,
        maxScrolls: 120,
      );
      await tester.pumpAndSettle();
    }

    await show(find.byKey(const Key('kaigo-continue')));
    await tester.tap(find.byKey(const Key('kaigo-continue')));
    await tester.pumpAndSettle();
    expect(find.text(course.lessons.first.title), findsOneWidget);
    for (var i = 0; i < 2; i++) {
      final f = find.byKey(
        Key('kaigo-answer-$i-${course.lessons.first.checks[i].answer}'),
      );
      await show(f);
      await tester.tap(f);
      await tester.pumpAndSettle();
    }
    await show(find.byKey(const Key('kaigo-check')));
    await tester.tap(find.byKey(const Key('kaigo-check')));
    await tester.pumpAndSettle();
    await show(find.byKey(const Key('kaigo-complete')));
    await tester.tap(find.byKey(const Key('kaigo-complete')));
    await tester.pumpAndSettle();
    expect(prefs.getStringList('ssw_building-cleaning_completed_v1_cleaner'), [
      'cleaning-01',
    ]);
    expect(prefs.getStringList('ssw_kaigo_completed_v1_cleaner'), isNull);
    expect(find.text(course.lessons[1].title), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'practice phone retries failure, then reveals feedback only after check',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final source =
          jsonDecode(
                File(
                  'backend/internal/store/ssw_resources/assessments.json',
                ).readAsStringSync(),
              )
              as Json;
      final bank = diagnosticList(
        (source['banks'] as Json)['building-cleaning'],
      );
      final dio = Dio();
      var fail = true;
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (o, h) {
            if (o.method == 'GET') {
              if (o.path.contains('/offline/practice/')) {
                h.reject(
                  DioException(
                    requestOptions: o,
                    response: Response<Json>(
                      requestOptions: o,
                      statusCode: 503,
                    ),
                  ),
                );
                return;
              }
              h.resolve(
                Response<Json>(
                  requestOptions: o,
                  data: {
                    'data': {'questions': bank},
                  },
                ),
              );
              return;
            }
            if (fail) {
              fail = false;
              h.reject(
                DioException(
                  requestOptions: o,
                  type: DioExceptionType.connectionTimeout,
                ),
              );
              return;
            }
            final body = o.data as Json;
            final q = bank.firstWhere((q) => q['id'] == body['questionId']);
            h.resolve(
              Response<Json>(
                requestOptions: o,
                data: {
                  'data': {
                    'correct': q['answer'] == body['answer'],
                    'answer': q['answer'],
                    'explanation': q['explanation'],
                    'lessonId': q['lessonId'],
                  },
                },
              ),
            );
          },
        ),
      );
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            dioProvider.overrideWithValue(dio),
            offlineProvider.overrideWithValue(false),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                const AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'learner',
                    displayName: 'Learner',
                    email: 'l@example.test',
                    isPremium: true,
                  ),
                ),
              ),
            ),
          ],
          child: MaterialApp(
            home: const SswPracticeScreen(sector: 'building-cleaning'),
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.4)),
              child: child!,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.text('Dasar profesi dan mutu layanan'),
        180,
        scrollable: find.byType(Scrollable).first,
        maxScrolls: 20,
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Dasar profesi dan mutu layanan'));
      await tester.pumpAndSettle();
      expect(find.text('Jawaban:'), findsNothing);
      Future<void> show(Finder f) async {
        await tester.scrollUntilVisible(
          f,
          180,
          scrollable: find.byType(Scrollable).first,
          maxScrolls: 30,
        );
        await tester.pumpAndSettle();
      }

      await show(find.byKey(const Key('ssw-practice-answer-0')));
      await tester.tap(find.byKey(const Key('ssw-practice-answer-0')));
      await tester.pumpAndSettle();
      await show(find.text('Periksa jawaban'));
      await tester.tap(find.text('Periksa jawaban'));
      await tester.pumpAndSettle();
      expect(find.textContaining('Tidak dapat terhubung'), findsOneWidget);
      expect(find.text('Tepat'), findsNothing);
      await show(find.text('Periksa jawaban'));
      await tester.tap(find.text('Periksa jawaban'));
      await tester.pumpAndSettle();
      await show(find.text('Soal berikutnya'));
      expect(find.text('Tepat'), findsOneWidget);
      await tester.tap(find.text('Soal berikutnya'));
      await tester.pumpAndSettle();
      expect(find.text('Soal 2/6 · 1 benar'), findsOneWidget);
      expect(find.text('Tepat'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'Kaigo simulation displays the full 60 minute clock and SSW labels',
    (tester) async {
      final repo = _SswSessionRepository();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            simulationLevelRepositoryProvider(
              'SSW-KAIGO',
            ).overrideWithValue(repo),
          ],
          child: const MaterialApp(
            home: SimulationAttemptScreen(attemptId: 'a', level: 'SSW-KAIGO'),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Simulasi Kaigo · Keterampilan'), findsOneWidget);
      expect(find.text('Keterampilan Kaigo · 45 soal'), findsOneWidget);
      expect(
        tester.widget<Text>(find.byKey(const Key('simulation-timer'))).data,
        startsWith('60:'),
      );
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
    },
  );
}

class _SswSessionRepository extends SimulationRepository {
  _SswSessionRepository() : super(Dio(), level: 'SSW-KAIGO');
  @override
  Future<Json> session(String id) async {
    final now = DateTime.now().toUtc();
    return {
      'attempt': {'id': id},
      'paperId': 'ssw-kaigo-01',
      'activeSection': 'vocabulary',
      'serverTime': now.toIso8601String(),
      'expiresAt': now.add(const Duration(hours: 1)).toIso8601String(),
      'sections': [
        {
          'id': 'vocabulary',
          'name': 'Keterampilan Kaigo · 45 soal',
          'durationSeconds': 3600,
          'questionCount': 45,
        },
      ],
      'questions': [
        {
          'id': 'q',
          'number': 1,
          'section': 'vocabulary',
          'group': 'Dasar Kaigo',
          'instructions': 'Pilih jawaban.',
          'prompt': 'Apa prinsip utama Kaigo?',
          'options': [
            {'id': 'a', 'text': 'Hormati martabat pengguna'},
            {'id': 'b', 'text': 'Abaikan pilihan pengguna'},
          ],
        },
      ],
      'answers': <String, dynamic>{},
      'marked': <String>[],
    };
  }
}
