import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/jft_intro_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_attempt_screen.dart';

import 'simulation_test.dart' show FakeSimulation;

class FakeJft extends FakeSimulation {
  FakeJft() {
    durationSeconds = 2700;
  }

  @override
  Future<Json> submitSection(String id, String section) async {
    submissions.add(section);
    active = switch (section) {
      'vocabulary' => 'conversation',
      'conversation' => 'reading',
      _ => 'completed',
    };
    answers.clear();
    flags.clear();
    return session(id);
  }
}

void main() {
  test('JFT repository uses JFT overview and real exam ID', () async {
    final requests = <RequestOptions>[];
    final dio = Dio();
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (request, handler) {
          requests.add(request);
          handler.resolve(
            Response<Json>(
              requestOptions: request,
              data: {
                'data': {'id': 'jft-attempt'},
              },
            ),
          );
        },
      ),
    );
    final repository = SimulationRepository(dio, level: 'JFT');
    await repository.overview();
    expect(await repository.start(), 'jft-attempt');
    expect(requests[0].path, '/simulation/jft');
    expect(requests[1].data, {'examId': 'jft-a2-full-1'});
  });

  testWidgets('JFT phone shows written scope and one shared timer',
      (tester) async {
    tester.view.physicalSize = const Size(320, 740);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          simulationLevelOverviewProvider('JFT').overrideWith(
            (ref) async => {'activeAttemptId': '', 'history': <Json>[]},
          ),
        ],
        child: const MaterialApp(home: JftIntroScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('JFT Exam 1'), findsOneWidget);
    expect(find.textContaining('36 soal · 45 menit'), findsOneWidget);
    expect(find.textContaining('Listening belum termasuk'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.textContaining('12 soal grammar dan ungkapan'),
      100,
    );
    await tester.pumpAndSettle();
    expect(find.textContaining('12 soal grammar dan ungkapan'), findsOneWidget);
    expect(find.text('Mulai Exam 1'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('JFT moves through all three sections to results',
      (tester) async {
    final repository = FakeJft();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          simulationLevelRepositoryProvider('JFT')
              .overrideWithValue(repository),
        ],
        child: const MaterialApp(
          home: SimulationAttemptScreen(attemptId: 'jft', level: 'JFT'),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('JFT Exam 1'), findsOneWidget);
    expect(
      tester.widget<Text>(find.byKey(const Key('simulation-timer'))).data,
      startsWith('45:'),
    );
    for (final expected in [
      'Bagian 2 · 会話と表現',
      'Bagian 3 · 読解',
      'Hasil JFT Exam 1',
    ]) {
      if (find.text('Berikutnya').evaluate().isNotEmpty) {
        await tester.tap(find.text('Berikutnya'));
        await tester.pumpAndSettle();
      }
      await tester.tap(find.text('Kirim bagian'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Kirim bagian').last);
      await tester.pumpAndSettle();
      expect(find.text(expected), findsOneWidget);
      expect(tester.takeException(), isNull);
    }
    expect(repository.submissions, ['vocabulary', 'conversation', 'reading']);
    expect(
      find.textContaining('JFT Exam 1 · Kosakata, Percakapan & Reading'),
      findsOneWidget,
    );
    await tester.pumpWidget(const SizedBox());
  });
}
