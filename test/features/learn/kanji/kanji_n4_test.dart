import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/kanji/data/kanji_repository.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/controllers/kanji_controller.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/screens/kanji_course_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

late SharedPreferences _preferences;

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    _preferences = await SharedPreferences.getInstance();
  });

  test('repository requests the selected level and defaults to N5', () async {
    final dio = Dio();
    addTearDown(dio.close);
    final requested = <String?>[];
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final level = options.uri.queryParameters['level'];
          requested.add(level);
          handler.resolve(
            Response<Map<String, dynamic>>(
              requestOptions: options,
              data: {
                'data': {
                  'level': level,
                  'title': 'Kanji JLPT $level',
                  'progress': 0,
                  'completedLessons': 0,
                  'lessons': <dynamic>[],
                },
              },
            ),
          );
        },
      ),
    );
    final repository = KanjiRepository(dio);
    expect((await repository.getCourse()).level, 'N5');
    expect((await repository.getCourse(level: 'N4')).level, 'N4');
    expect(requested, ['N5', 'N4']);
  });

  testWidgets('N4 catalog uses N4 provider and displays actual content counts',
      (tester) async {
    tester.view.physicalSize = const Size(360, 640);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final requested = <String>[];
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(_preferences),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          offlineProvider.overrideWithValue(true),
          kanjiCourseProvider.overrideWith((ref, level) async {
            requested.add(level);
            return _course(level);
          }),
        ],
        child: const MaterialApp(home: KanjiCourseScreen(level: 'N4')),
      ),
    );
    await tester.pumpAndSettle();
    expect(requested, ['N4']);
    expect(find.text('Kanji JLPT N4'), findsOneWidget);
    expect(find.text('34 lesson • 166 kanji • 332 soal'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await tester.scrollUntilVisible(
      find.byKey(const Key('kanji-search')),
      250,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.enterText(find.byKey(const Key('kanji-search')), 'Pertemuan');
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('kanji-lesson-kanji-n4-01')), findsNothing);
    await tester.scrollUntilVisible(
      find.byKey(const Key('kanji-lesson-kanji-n4-32')),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Pertemuan dan perjalanan'), findsOneWidget);
    await tester.ensureVisible(find.byKey(const Key('kanji-n4-sources')));
    await tester.tap(find.text('Tentang materi dan sumber'));
    await tester.pumpAndSettle();
    expect(find.textContaining('James William Breen'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('existing catalog screen continues to request N5 by default',
      (tester) async {
    final requested = <String>[];
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(_preferences),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          offlineProvider.overrideWithValue(true),
          kanjiCourseProvider.overrideWith((ref, level) async {
            requested.add(level);
            return _course(level);
          }),
        ],
        child: const MaterialApp(home: KanjiCourseScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(requested, ['N5']);
    expect(find.text('Kanji JLPT N5'), findsOneWidget);
    expect(find.text('20 lesson • 100 kanji • 200 soal'), findsOneWidget);
  });
}

KanjiCourse _course(String level) {
  final raw = jsonDecode(
    File('backend/data/kanji/${level.toLowerCase()}.json').readAsStringSync(),
  ) as List<dynamic>;
  final lessons = raw.cast<Map<String, dynamic>>();
  return KanjiCourse(
    level: level,
    title: 'Kanji JLPT $level',
    progress: 0,
    completedLessons: 0,
    lessons: [
      for (final (index, value) in lessons.indexed)
        KanjiLessonSummary(
          id: value['ID'] as String,
          order: index + 1,
          chapter: value['Chapter'] as String,
          title: value['Title'] as String,
          characterCount: (value['Characters'] as List<dynamic>).length,
          questionCount: (value['Questions'] as List<dynamic>).length,
          bestScore: 0,
          completed: false,
        ),
    ],
  );
}
