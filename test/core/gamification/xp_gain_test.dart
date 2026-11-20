import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/gamification/xp_gain.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/gamification/double_xp_button.dart';

void main() {
  test('reads the XP a result earned, or nothing for older responses', () {
    final gain = XpGain.fromJson({
      'awarded': 50,
      'total': 1250,
      'streak': 4,
      'completionId': 'grammar:attempt-1',
    })!;
    expect(gain.awarded, 50);
    expect(gain.total, 1250);
    expect(gain.streak, 4);
    expect(gain.completionId, 'grammar:attempt-1');
    expect(XpGain.fromJson(null), isNull);
    expect(XpGain.fromJson('x'), isNull);

    final result = GrammarResult.fromJson({
      'score': 90,
      'correctAnswers': 9,
      'totalQuestions': 10,
      'passed': true,
      'bestScore': 90,
      'progress': 10,
      'xp': {'awarded': 50, 'total': 50, 'streak': 1},
    });
    expect(result.xp?.awarded, 50);
    expect(
      GrammarResult.fromJson({
        'score': 90,
        'correctAnswers': 9,
        'totalQuestions': 10,
        'passed': true,
        'bestScore': 90,
        'progress': 10,
      }).xp,
      isNull,
    );
  });

  Future<void> pump(WidgetTester tester, XpGain? gain) => tester.pumpWidget(
    ProviderScope(
      overrides: [xpBoostPlacementProvider.overrideWith((ref) async => null)],
      child: MaterialApp(
        home: Scaffold(body: XpResultBadge(gain: gain)),
      ),
    ),
  );

  testWidgets('shows the XP earned and the streak', (tester) async {
    await pump(tester, const XpGain(awarded: 10, total: 60, streak: 3));
    expect(find.text('+10 XP'), findsOneWidget);
    expect(find.byKey(const Key('xp-streak')), findsOneWidget);
    expect(find.textContaining('3 '), findsOneWidget);
  });

  testWidgets('a completion that earned nothing still shows the streak', (
    tester,
  ) async {
    await pump(tester, const XpGain(awarded: 0, total: 60, streak: 2));
    expect(find.byKey(const Key('xp-awarded')), findsNothing);
    expect(find.byKey(const Key('xp-streak')), findsOneWidget);
  });

  testWidgets('draws nothing without a gain', (tester) async {
    await pump(tester, null);
    expect(find.byKey(const Key('xp-awarded')), findsNothing);
    expect(find.byKey(const Key('xp-streak')), findsNothing);
  });
}
