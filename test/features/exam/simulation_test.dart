import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_attempt_screen.dart';

void main() {
  Future<void> open(WidgetTester tester, FakeSimulation repository) async {
    tester.view.physicalSize = const Size(360, 740);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [simulationRepositoryProvider.overrideWithValue(repository)],
        child:
            const MaterialApp(home: SimulationAttemptScreen(attemptId: 'test')),
      ),
    );
    await tester.pumpAndSettle();
  }

  testWidgets('small phone saves choice and mark, locks section, then reviews',
      (tester) async {
    final repository = FakeSimulation();
    await open(tester, repository);
    expect(find.text('Bagian 1 · 文字・語彙'), findsOneWidget);
    expect(find.text('Jawaban benar: あ'), findsNothing);
    final option = find.byKey(const ValueKey('simulation-option-v1-b'));
    await tester.ensureVisible(option);
    await tester.tap(option);
    await tester.pumpAndSettle();
    expect(repository.answers['v1'], 'v1-b');
    final flag = find.text('Tandai ragu');
    await tester.ensureVisible(flag);
    await tester.tap(flag);
    await tester.pumpAndSettle();
    expect(repository.flags, contains('v1'));

    await tester.tap(find.byTooltip('Peta soal'));
    await tester.pumpAndSettle();
    expect(find.text('1/2 dijawab · 1 ditandai'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, '2'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Kirim bagian'));
    await tester.pumpAndSettle();
    expect(find.textContaining('1 soal kosong'), findsOneWidget);
    await tester.tap(find.text('Kirim bagian').last);
    await tester.pumpAndSettle();
    expect(find.text('Bagian 2 · 文法'), findsOneWidget);
    expect(repository.submissions, ['vocabulary']);
    expect(find.textContaining('Kosakata dikunci.'), findsOneWidget);
    expect(find.text('Jawaban benar: あ'), findsNothing);
    await tester.tap(find.text('Kirim bagian'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Kirim bagian').last);
    await tester.pumpAndSettle();
    expect(repository.submissions, ['vocabulary', 'grammar']);
    expect(find.text('Simulasi selesai.'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets(
      'failed save is not shown as saved and confirmed server answer reloads',
      (tester) async {
    final repository = FakeSimulation()..failSave = true;
    await open(tester, repository);
    final option = find.byKey(const ValueKey('simulation-option-v1-b'));
    await tester.ensureVisible(option);
    await tester.tap(option);
    await tester.pumpAndSettle();
    expect(repository.answers, isEmpty);
    expect(
      find.textContaining('Pilihan belum dikonfirmasi tersimpan.'),
      findsOneWidget,
    );
    repository.failSave = false;
    await tester.ensureVisible(option);
    await tester.tap(option);
    await tester.pumpAndSettle();
    expect(repository.answers['v1'], 'v1-b');
    expect(
      find.textContaining('Pilihan belum dikonfirmasi tersimpan.'),
      findsNothing,
    );
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets('expiry sync opens grammar without submitting it',
      (tester) async {
    final repository = FakeSimulation()..expiredFirst = true;
    await open(tester, repository);
    await tester.pump(const Duration(seconds: 1));
    await tester.pumpAndSettle();
    expect(find.text('Bagian 2 · 文法'), findsOneWidget);
    expect(repository.submissions, isEmpty);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets('late network response after disposal does not update screen',
      (tester) async {
    final repository = FakeSimulation();
    await open(tester, repository);
    repository.pending = Completer<void>();
    final option = find.byKey(const ValueKey('simulation-option-v1-a'));
    await tester.ensureVisible(option);
    await tester.tap(option);
    await tester.pump();
    await tester.pumpWidget(const SizedBox());
    repository.pending!.complete();
    await tester.pump();
    expect(tester.takeException(), isNull);
  });
}

class FakeSimulation extends SimulationRepository {
  FakeSimulation() : super(Dio());
  String active = 'vocabulary';
  final answers = <String, String>{};
  final flags = <String>{};
  final submissions = <String>[];
  bool failSave = false;
  bool expiredFirst = false;
  int calls = 0;
  Completer<void>? pending;

  Json question(String id, int number) => {
        'id': id,
        'number': number,
        'section': active,
        'group': active == 'vocabulary' ? '文字・語彙 問題1' : '文法 問題2',
        'instructions': 'いちばん よい ものを ひとつ えらんで ください。',
        'prompt': 'これは [[本]]です。',
        'passage': '',
        'options': [
          for (final value in ['a', 'b', 'c', 'd'])
            {'id': '$id-$value', 'text': value == 'a' ? 'あ' : value},
        ],
      };

  @override
  Future<Json> session(String id) async {
    calls++;
    if (expiredFirst && calls > 1) active = 'grammar';
    final now = DateTime.now().toUtc();
    return {
      'attempt': {'id': id},
      'paperId': 'test',
      'activeSection': active,
      'serverTime': now.toIso8601String(),
      'expiresAt': now
          .add(Duration(seconds: expiredFirst && calls == 1 ? 0 : 1200))
          .toIso8601String(),
      'questions': active == 'vocabulary'
          ? [question('v1', 1), question('v2', 2)]
          : [question('g1', 22)],
      'answers': Map<String, dynamic>.from(answers),
      'marked': flags.toList(),
      'result': active == 'completed'
          ? {
              'score': 0,
              'correct': 0,
              'total': 3,
              'answered': 1,
              'durationSeconds': 20,
              'sections': [
                {'name': '文字・語彙 問題1', 'correct': 0, 'total': 2, 'score': 0},
              ],
              'reviews': <Json>[],
              'scoringNote': 'Persentase benar; bukan skor resmi JLPT.',
            }
          : null,
    };
  }

  @override
  Future<void> saveAnswer(
    String id,
    String question,
    String option, {
    required bool marked,
  }) async {
    if (pending != null) await pending!.future;
    if (failSave) throw Exception('Koneksi gagal.');
    if (option.isEmpty) {
      answers.remove(question);
    } else {
      answers[question] = option;
    }
    if (marked) {
      flags.add(question);
    } else {
      flags.remove(question);
    }
  }

  @override
  Future<Json> submitSection(String id, String section) async {
    submissions.add(section);
    active = section == 'vocabulary' ? 'grammar' : 'completed';
    answers.clear();
    flags.clear();
    return session(id);
  }
}
