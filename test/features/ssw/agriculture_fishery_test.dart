import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/learn/presentation/screens/learn_hub_screen.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  final courses = <String, KaigoCurriculum>{};
  final raw = <String, Json>{};
  setUpAll(() async {
    for (final sector in ['agriculture', 'fishery']) {
      raw[sector] = jsonDecode(
        await rootBundle.loadString('assets/content/ssw/$sector.json'),
      ) as Json;
      courses[sector] = KaigoCurriculum.fromJson(raw[sector]!);
    }
  });
  test(
      'both offline courses cover separate tracks and have usable complete lessons',
      () {
    for (final sector in courses.keys) {
      final course = courses[sector]!;
      expect(course.modules.length, 10);
      expect(course.lessons.length, sector == 'agriculture' ? 40 : 30);
      expect(
        course.lessons.map((l) => l.id).toSet().length,
        course.lessons.length,
      );
      final tracks = diagnosticList(raw[sector]!['modules'])
          .map((m) => m['track'])
          .toSet();
      expect(tracks.length, 2);
      for (final l in course.lessons) {
        expect(
          l.sections.map((s) => s.body).join(' ').split(' ').length,
          greaterThan(100),
        );
        expect(l.vocabulary.length, 4);
        expect(l.dialogue.length, 2);
        expect(l.steps.length, 4);
        expect(l.checks.length, 4);
        expect(l.scenario, isNotEmpty);
        expect(l.response, isNotEmpty);
        for (final q in l.checks) {
          expect(q.options.toSet().length, 4);
          expect(q.answer, inInclusiveRange(0, 3));
          expect(q.explanation, isNotEmpty);
        }
      }
      expect(sswActiveSectorIds.contains(sector), isTrue);
      for (final level in sswAssessmentLevels[sector]!) {
        expect(
          requiresPremium(
            Uri.parse('/exams/${level.toLowerCase()}/simulation'),
          ),
          isTrue,
        );
      }
      expect(requiresPremium(Uri.parse('/ssw/$sector/practice')), isTrue);
      expect(
        requiresPremium(
          Uri.parse('/ssw/$sector/lessons/${course.lessons.first.id}'),
        ),
        isFalse,
      );
    }
  });
  test(
      'backend practice keys and lesson links match the authored offline content',
      () {
    final resource = jsonDecode(
      File('backend/internal/store/ssw_resources/assessments.json')
          .readAsStringSync(),
    ) as Json;
    for (final sector in courses.keys) {
      final bank = diagnosticList((resource['banks'] as Json)[sector]);
      expect(bank.length, sector == 'agriculture' ? 160 : 120);
      expect(bank.map((q) => q['id']).toSet().length, bank.length);
      for (final q in bank) {
        final lesson = courses[sector]!.lessonById(q['lessonId'] as String)!;
        final original =
            lesson.checks.singleWhere((c) => c.prompt == q['prompt']);
        expect(
          (q['options'] as List)[q['answer'] as int],
          original.options[original.answer],
        );
      }
    }
  });
  testWidgets('Learn puts Anki and SSW below every JLPT subject',
      (tester) async {
    tester.view.physicalSize = const Size(420, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
        ],
        child: const MaterialApp(home: Scaffold(body: LearnHubScreen())),
      ),
    );
    await tester.pumpAndSettle();
    final ankiY = tester.getTopLeft(find.byType(AnkiEntryCard)).dy;
    for (final subject in ['Vocabulary', 'Grammar', 'Kanji', 'Reading']) {
      expect(tester.getTopLeft(find.text(subject)).dy, lessThan(ankiY));
    }
    expect(tester.getTopLeft(find.byType(SswEntryCard)).dy, greaterThan(ankiY));
    expect(tester.takeException(), isNull);
  });
  for (final sector in ['agriculture', 'fishery']) {
    testWidgets('$sector phone opens material and displays its two test tracks',
        (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      Widget app(Widget screen) => ProviderScope(
            overrides: [
              sharedPreferencesProvider.overrideWithValue(prefs),
              sswCurriculumProvider(sector)
                  .overrideWith((ref) async => courses[sector]!),
              authControllerProvider.overrideWith(
                (ref) => AuthController.test(
                  const AuthenticationState.authenticated(
                    AuthenticatedUser(
                      id: 'student',
                      displayName: 'Student',
                      email: 's@example.test',
                      isPremium: true,
                    ),
                  ),
                ),
              ),
            ],
            child: MaterialApp(
              home: screen,
              builder: (context, child) => MediaQuery(
                data: MediaQuery.of(context)
                    .copyWith(textScaler: const TextScaler.linear(1.4)),
                child: child!,
              ),
            ),
          );
      await tester.pumpWidget(app(SswSectorScreen(sectorId: sector)));
      await tester.pumpAndSettle();
      expect(
        find.text('Kurikulum ${sswSectorById(sector)!.name}'),
        findsOneWidget,
      );
      await tester.pumpWidget(
        app(
          KaigoLessonScreen(
            sectorId: sector,
            lessonId: courses[sector]!.lessons.first.id,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text(courses[sector]!.lessons.first.title), findsOneWidget);
      await tester
          .pumpWidget(app(SswSectorScreen(sectorId: sector, exams: true)));
      await tester.pumpAndSettle();
      for (final level in sswAssessmentLevels[sector]!) {
        await tester.scrollUntilVisible(
          find.text('Simulasi · ${sswExamTitle(level)}'),
          150,
          scrollable: find.byType(Scrollable).first,
          maxScrolls: 30,
        );
        expect(find.text('Simulasi · ${sswExamTitle(level)}'), findsOneWidget);
      }
      expect(tester.takeException(), isNull);
    });
  }
  test('new progress is account and sector scoped and removed on deletion',
      () async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final agriculture =
        KaigoProgress(prefs, 'student', namespace: 'agriculture');
    final fishery = KaigoProgress(prefs, 'student', namespace: 'fishery');
    final other = KaigoProgress(prefs, 'other', namespace: 'fishery');
    addTearDown(agriculture.dispose);
    addTearDown(fishery.dispose);
    addTearDown(other.dispose);
    await agriculture.complete('agri-01');
    await fishery.complete('fish-01');
    expect(agriculture.state, {'agri-01'});
    expect(fishery.state, {'fish-01'});
    expect(other.state, isEmpty);
    final restored = KaigoProgress(prefs, 'student', namespace: 'fishery');
    addTearDown(restored.dispose);
    expect(restored.state, fishery.state);
    expect(prefs.getKeys(), {
      'ssw_agriculture_completed_v1_student',
      'ssw_fishery_completed_v1_student',
    });
    await other.complete('fish-02');
    final dio = Dio();
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (o, h) =>
            h.resolve(Response<void>(requestOptions: o, statusCode: 204)),
      ),
    );
    await AuthRepository(dio, prefs).deleteAccount(userId: 'student');
    expect(prefs.getStringList('ssw_agriculture_completed_v1_student'), isNull);
    expect(prefs.getStringList('ssw_fishery_completed_v1_student'), isNull);
    expect(prefs.getStringList('ssw_fishery_completed_v1_other'), ['fish-02']);
  });
}
