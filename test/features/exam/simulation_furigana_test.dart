import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/simulation_text.dart';

void main() {
  test('parses furigana and underlined tested text', () {
    expect(parseSimulationText('{私|わたし}は[[毎週]]会います。'), const [
      SimulationSegment('私', ruby: 'わたし'),
      SimulationSegment('は'),
      SimulationSegment('毎週', underlined: true),
      SimulationSegment('会います。'),
    ]);
    expect(parseSimulationText('[[{今日|きょう}は雨]]'), const [
      SimulationSegment('今日', ruby: 'きょう', underlined: true),
      SimulationSegment('は雨', underlined: true),
    ]);
    expect(parseSimulationText('ひらがなだけ'), const [
      SimulationSegment('ひらがなだけ'),
    ]);
    expect(stripSimulationMarkup('{駅|えき}の[[前]]'), '駅の前');
  });

  test('maps extra paper level keys', () {
    expect(jlptBaseLevel('N5-2'), 'N5');
    expect(jlptBaseLevel('N4'), 'N4');
    expect(jlptBaseLevel('NAT-2'), 'NAT-2');
    expect(jlptPaperNumber('N4-3'), 3);
    expect(jlptPaperNumber('N5'), 1);
    expect(jlptPaperLabel('N5-2'), '02');
  });

  test('extra JLPT papers start their own exam IDs', () async {
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
                'data': {'id': 'attempt'},
              },
            ),
          );
        },
      ),
    );
    await SimulationRepository(dio, level: 'N5-2').start();
    await SimulationRepository(dio, level: 'N4-3').start();
    await SimulationRepository(dio, level: 'N4-3').overview();
    expect(requests[0].data, {'examId': 'jlpt-n5-simulation-2'});
    expect(requests[1].data, {'examId': 'jlpt-n4-simulation-3'});
    expect(requests[2].path, '/simulation/n4-3');
  });

  testWidgets('shows furigana above kanji and keeps plain text intact',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              SimulationText('{私|わたし}は[[毎週]]{会|あ}います。'),
              SimulationText('ひらがなです。'),
            ],
          ),
        ),
      ),
    );
    expect(find.text('わたし'), findsOneWidget);
    expect(find.text('私'), findsOneWidget);
    expect(find.text('あ'), findsOneWidget);
    expect(find.text('ひらがなです。'), findsOneWidget);
    expect(tester.takeException(), isNull);
    final reading = tester.getTopLeft(find.text('わたし')).dy;
    final base = tester.getTopLeft(find.text('私')).dy;
    expect(reading, lessThan(base));
  });
}
