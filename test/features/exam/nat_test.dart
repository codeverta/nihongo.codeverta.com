import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_attempt_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_intro_screen.dart';

import 'simulation_test.dart' show FakeSimulation;

void main() {
  test('NAT repository uses its own overview and exam ID', () async {
    final requests = <RequestOptions>[];
    final dio = Dio();
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          requests.add(options);
          handler.resolve(
            Response<Json>(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {'id': 'nat-attempt', 'questionCount': 43},
              },
            ),
          );
        },
      ),
    );
    final repository = SimulationRepository(dio, level: 'NAT');
    expect((await repository.overview())['questionCount'], 43);
    expect(await repository.start(), 'nat-attempt');
    expect(requests[0].path, '/simulation/nat');
    expect(requests[1].method, 'POST');
    expect(requests[1].data, {'examId': 'nat-5q-full-1'});
  });
  testWidgets('NAT grammar and reading keeps the full 40 minute timer',
      (tester) async {
    final repository = FakeSimulation()
      ..active = 'grammar'
      ..durationSeconds = 2400;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          simulationLevelRepositoryProvider('NAT')
              .overrideWithValue(repository),
        ],
        child: const MaterialApp(
          home: SimulationAttemptScreen(attemptId: 'nat', level: 'NAT'),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('NAT 5Q Exam 1'), findsOneWidget);
    expect(find.text('Bagian 2 · 文法・読解'), findsOneWidget);
    final timer =
        tester.widget<Text>(find.byKey(const Key('simulation-timer')));
    expect(timer.data, startsWith('40:'));
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets('NAT intro shows both sections and reading content on a phone',
      (tester) async {
    tester.view.physicalSize = const Size(360, 740);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          simulationLevelOverviewProvider('NAT').overrideWith(
            (ref) async => {
              'activeAttemptId': '',
              'history': <Json>[],
            },
          ),
        ],
        child: const MaterialApp(home: SimulationIntroScreen(level: 'NAT')),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('NAT 5Q Exam 1'), findsOneWidget);
    expect(find.textContaining('43 soal · 60 menit'), findsOneWidget);
    expect(find.text('1. 文字・語彙 — 21 soal / 20 menit'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('2. 文法・読解 — 22 soal / 40 menit'),
      100,
    );
    expect(find.textContaining('5 reading'), findsOneWidget);
    expect(find.text('Mulai simulasi'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });
}
