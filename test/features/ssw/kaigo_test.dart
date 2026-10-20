import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  late KaigoCurriculum course;
  setUpAll(() async {
    course = KaigoCurriculum.fromJson(
      jsonDecode(await rootBundle.loadString('assets/content/ssw/kaigo.json'))
          as Map<String, dynamic>,
    );
  });
  test(
    'offline curriculum has unique navigable lessons and valid study checks',
    () {
      expect(course.modules.length, 12);
      expect(course.lessons.length, 36);
      expect(course.lessons.map((l) => l.id).toSet().length, 36);
      expect(course.lessons.expand((l) => l.checks).length, 72);
      for (final module in course.modules) {
        expect(module.lessons.length, 3);
        for (final lesson in module.lessons) {
          expect(course.lessonById(lesson.id), same(lesson));
          expect(lesson.objectives.length, greaterThanOrEqualTo(2));
          expect(
            lesson.sections.map((s) => s.body).join(' ').split(' ').length,
            greaterThan(100),
          );
          expect(lesson.steps.length, 4);
          expect(lesson.scenario, isNotEmpty);
          expect(lesson.response, isNotEmpty);
          expect(lesson.vocabulary.length, greaterThanOrEqualTo(4));
          expect(lesson.dialogue.length, greaterThanOrEqualTo(2));
          for (final check in lesson.checks) {
            expect(check.options.toSet().length, check.options.length);
            expect(check.answer, inInclusiveRange(0, check.options.length - 1));
            expect(check.explanation, isNotEmpty);
          }
        }
      }
      expect(course.lessonById('missing'), isNull);
    },
  );
  test(
    'progress persists per account, saves concurrently, and is removed on deletion',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final first = KaigoProgress(preferences, 'first');
      final other = KaigoProgress(preferences, 'other');
      final guest = KaigoProgress(preferences, null);
      addTearDown(first.dispose);
      addTearDown(other.dispose);
      addTearDown(guest.dispose);
      await Future.wait([
        first.complete('kaigo-01'),
        first.complete('kaigo-02'),
      ]);
      await first.complete('kaigo-01');
      expect(first.state, {'kaigo-01', 'kaigo-02'});
      final restored = KaigoProgress(preferences, 'first');
      addTearDown(restored.dispose);
      expect(restored.state, first.state);
      expect(other.state, isEmpty);
      await other.complete('kaigo-03');
      await guest.complete('kaigo-04');
      expect(preferences.getKeys(), {
        'ssw_kaigo_completed_v1_first',
        'ssw_kaigo_completed_v1_other',
      });
      final dio = Dio(BaseOptions(baseUrl: 'https://example.test/v1'));
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            expect(options.path, '/me/account');
            handler.resolve(
              Response<void>(requestOptions: options, statusCode: 204),
            );
          },
        ),
      );
      await AuthRepository(dio, preferences).deleteAccount(userId: 'first');
      expect(preferences.getStringList('ssw_kaigo_completed_v1_first'), isNull);
      expect(preferences.getStringList('ssw_kaigo_completed_v1_other'), [
        'kaigo-03',
      ]);
    },
  );
  testWidgets(
    'Kaigo phone reader retries wrong answers and saves before the next lesson',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final router = GoRouter(
        initialLocation: '/ssw/nursing-care',
        routes: [
          GoRoute(
            path: '/ssw/nursing-care',
            builder: (_, __) => const SswSectorScreen(sectorId: 'nursing-care'),
          ),
          GoRoute(
            path: '/ssw/nursing-care/lessons/:lessonId',
            builder: (_, state) =>
                KaigoLessonScreen(lessonId: state.pathParameters['lessonId']!),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPreferencesProvider.overrideWithValue(preferences),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                const AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'student',
                    displayName: 'Student',
                    email: 's@example.test',
                  ),
                ),
              ),
            ),
            kaigoCurriculumProvider.overrideWith((ref) async => course),
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
      Future<void> show(Finder target, {double delta = 350}) async {
        await tester.scrollUntilVisible(
          target,
          delta,
          scrollable: find.byType(Scrollable).first,
          maxScrolls: 120,
        );
        await tester.pumpAndSettle();
      }

      await show(find.byKey(const Key('kaigo-continue')));
      await tester.tap(find.byKey(const Key('kaigo-continue')));
      await tester.pumpAndSettle();
      expect(find.text(course.lessons.first.title), findsOneWidget);
      final furiganaToggle = find.byKey(const Key('ssw-furigana-toggle'));
      expect(tester.widget<FilterChip>(furiganaToggle).selected, isTrue);
      await show(find.text('かいご'));
      expect(find.text('かいご'), findsOneWidget);
      await show(furiganaToggle, delta: -350);
      await tester.tap(furiganaToggle);
      await tester.pumpAndSettle();
      await show(find.text('介護'));
      expect(find.text('かいご'), findsNothing);
      expect(find.byKey(const Key('kaigo-complete')), findsNothing);
      for (var i = 0; i < 2; i++) {
        final option = find.byKey(Key('kaigo-answer-$i-2'));
        await show(option);
        await tester.tap(option);
        await tester.pumpAndSettle();
      }
      await show(find.byKey(const Key('kaigo-check')));
      await tester.tap(find.byKey(const Key('kaigo-check')));
      await tester.pumpAndSettle();
      await show(find.byKey(const Key('kaigo-retry')));
      expect(find.byKey(const Key('kaigo-complete')), findsNothing);
      await tester.tap(find.byKey(const Key('kaigo-retry')));
      await tester.pumpAndSettle();
      for (var i = 0; i < 2; i++) {
        final option = find.byKey(
          Key('kaigo-answer-$i-${course.lessons.first.checks[i].answer}'),
        );
        await show(option, delta: i == 0 ? -350 : 350);
        await tester.tap(option);
        await tester.pumpAndSettle();
      }
      await show(find.byKey(const Key('kaigo-check')));
      await tester.tap(find.byKey(const Key('kaigo-check')));
      await tester.pumpAndSettle();
      await show(find.byKey(const Key('kaigo-complete')));
      await tester.tap(find.byKey(const Key('kaigo-complete')));
      await tester.pumpAndSettle();
      expect(preferences.getStringList('ssw_kaigo_completed_v1_student'), [
        'kaigo-01',
      ]);
      expect(find.text(course.lessons[1].title), findsOneWidget);
      expect(tester.takeException(), isNull);
      router.go('/ssw/nursing-care/lessons/missing');
      await tester.pumpAndSettle();
      expect(find.text('Lesson tidak ditemukan.'), findsOneWidget);
    },
  );
}
