import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/features/ai_chat/data/ai_chat_repository.dart';
import 'package:nihongo_master/features/ai_chat/presentation/ai_chat_sheet.dart';
import 'package:nihongo_master/features/ai_chat/presentation/ai_markdown.dart';
import 'package:shared_preferences/shared_preferences.dart';

typedef _Sent = ({AiChip? chip, String message});

class _FakeRepository extends AiChatRepository {
  _FakeRepository() : super(Dio());

  final sent = <_Sent>[];
  final adCredits = <bool>[];

  /// Responses for successive sends: a String answer or an AiChatException.
  final responses = <Object>[];
  List<AiChatMessage> stored = const [];

  @override
  Stream<AiStreamEvent> send({
    required AiLessonType lessonType,
    required String lessonId,
    AiChip? chip,
    String message = '',
  }) {
    sent.add((chip: chip, message: message));
    final next = responses.isEmpty ? 'Jawaban.' : responses.removeAt(0);
    if (next is AiChatException) return Stream.error(next);
    final answer = next as String;
    return Stream.fromIterable([
      for (final word in answer.split(' ')) AiDelta('$word '),
      const AiDone(cached: false, truncated: false),
    ]);
  }

  @override
  Future<void> grantAdCredit({required bool watched}) async {
    adCredits.add(watched);
  }

  @override
  Future<List<AiChatMessage>> history(AiLessonType t, String id) async =>
      stored;

  @override
  Future<void> clear(AiLessonType t, String id) async => stored = const [];
}

class _FakeShower implements RewardedAdShower {
  _FakeShower(this.outcome);
  final RewardedAdOutcome outcome;
  final shown = <String>[];

  @override
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  }) async {
    shown.add(unitId);
    return outcome;
  }
}

const _key = (type: AiLessonType.grammar, lessonId: 'grammar-n5-01');
const _adRequired = AiChatException(
  AiChatException.adRequired,
  'Kuota gratis hari ini sudah habis.',
);

Future<void> _pump(
  WidgetTester tester, {
  required _FakeRepository repository,
  RewardedAdShower? shower,
  AdPlacementConfig? placement,
  AdPlacementConfig? lifetimePlacement,
  bool offline = false,
}) async {
  // Wide enough that every quick-action chip is laid out.
  tester.view
    ..physicalSize = const Size(1600, 2400)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  SharedPreferences.setMockInitialValues({});
  final preferences = await SharedPreferences.getInstance();
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        aiChatRepositoryProvider.overrideWithValue(repository),
        rewardedAdShowerProvider.overrideWithValue(
          shower ?? _FakeShower(RewardedAdOutcome.watched),
        ),
        aiChatRewardedPlacementProvider.overrideWith((ref) async => placement),
        aiChatLifetimeRewardedPlacementProvider.overrideWith(
          (ref) async => lifetimePlacement,
        ),
        offlineProvider.overrideWithValue(offline),
        analyticsServiceProvider.overrideWithValue(
          AnalyticsService(preferences),
        ),
      ],
      child: const MaterialApp(
        home: Scaffold(body: AiChatSheet(chatKey: _key)),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

const _placement = AdPlacementConfig(
  enabled: true,
  unitId: testRewardedAdUnitId,
);

void main() {
  testWidgets('a quick action sends the chip and shows the streamed answer', (
    tester,
  ) async {
    final repository = _FakeRepository()..responses.add('Ringkasan pelajaran');
    await _pump(tester, repository: repository);
    await tester.tap(find.byKey(const ValueKey('ai-chip-summary')));
    await tester.pumpAndSettle();
    expect(repository.sent.single.chip, AiChip.summary);
    expect(repository.sent.single.message, isEmpty);
    expect(find.textContaining('Ringkasan pelajaran'), findsOneWidget);
  });

  testWidgets('free text and sentence correction send the typed text', (
    tester,
  ) async {
    final repository = _FakeRepository();
    await _pump(tester, repository: repository);
    await tester.enterText(
      find.byKey(const ValueKey('ai-input')),
      'Apa itu は?',
    );
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('ai-chip-correct')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const ValueKey('ai-input')), '私は学生です');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    expect(repository.sent, [
      (chip: null, message: 'Apa itu は?'),
      (chip: AiChip.correct, message: '私は学生です'),
    ]);
  });

  testWidgets('after the free quota an opt-in ad unlocks the held message', (
    tester,
  ) async {
    final repository = _FakeRepository()
      ..responses.addAll([_adRequired, 'Jawaban setelah iklan']);
    final shower = _FakeShower(RewardedAdOutcome.watched);
    await _pump(
      tester,
      repository: repository,
      shower: shower,
      placement: _placement,
    );
    await tester.enterText(find.byKey(const ValueKey('ai-input')), 'Satu lagi');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();

    // Never auto-played: the user must choose to watch.
    expect(shower.shown, isEmpty);
    expect(find.byKey(const ValueKey('ai-ad-watch')), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('ai-ad-watch')));
    await tester.pumpAndSettle();

    expect(shower.shown, [testRewardedAdUnitId]);
    expect(repository.adCredits, [true]);
    expect(repository.sent.map((s) => s.message), ['Satu lagi', 'Satu lagi']);
    expect(find.textContaining('Jawaban setelah iklan'), findsOneWidget);
  });

  testWidgets('cancelling the ad sheet sends nothing', (tester) async {
    final repository = _FakeRepository()..responses.add(_adRequired);
    final shower = _FakeShower(RewardedAdOutcome.watched);
    await _pump(
      tester,
      repository: repository,
      shower: shower,
      placement: _placement,
    );
    await tester.enterText(find.byKey(const ValueKey('ai-input')), 'Satu lagi');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('ai-ad-cancel')));
    await tester.pumpAndSettle();
    expect(shower.shown, isEmpty);
    expect(repository.adCredits, isEmpty);
    expect(repository.sent, hasLength(1));
  });

  testWidgets(
    'an ad closed early grants nothing; an unavailable ad still unlocks',
    (tester) async {
      final repository = _FakeRepository()..responses.add(_adRequired);
      await _pump(
        tester,
        repository: repository,
        shower: _FakeShower(RewardedAdOutcome.dismissed),
        placement: _placement,
      );
      await tester.enterText(find.byKey(const ValueKey('ai-input')), 'x');
      await tester.tap(find.byKey(const ValueKey('ai-send')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('ai-ad-watch')));
      await tester.pumpAndSettle();
      expect(repository.adCredits, isEmpty);

      final failOpen = _FakeRepository()..responses.addAll([_adRequired, 'ok']);
      await _pump(
        tester,
        repository: failOpen,
        shower: _FakeShower(RewardedAdOutcome.unavailable),
        placement: _placement,
      );
      await tester.enterText(find.byKey(const ValueKey('ai-input')), 'x');
      await tester.tap(find.byKey(const ValueKey('ai-send')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('ai-ad-watch')));
      await tester.pumpAndSettle();
      expect(failOpen.adCredits, [false]);
      expect(failOpen.sent, hasLength(2));
    },
  );

  testWidgets(
    'without a rewarded placement (iOS, ads off) the user sees the Lifetime upsell',
    (tester) async {
      final repository = _FakeRepository()..responses.add(_adRequired);
      final shower = _FakeShower(RewardedAdOutcome.watched);
      await _pump(tester, repository: repository, shower: shower);
      await tester.enterText(find.byKey(const ValueKey('ai-input')), 'x');
      await tester.tap(find.byKey(const ValueKey('ai-send')));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('ai-ad-watch')), findsNothing);
      expect(find.byKey(const ValueKey('ai-upsell-lifetime')), findsOneWidget);
      expect(shower.shown, isEmpty);
    },
  );

  testWidgets(
    'a Lifetime user at the daily limit is told to come back tomorrow',
    (tester) async {
      final repository = _FakeRepository()
        ..responses.add(
          const AiChatException(AiChatException.premiumLimit, 'limit'),
        );
      await _pump(tester, repository: repository, placement: _placement);
      await tester.enterText(find.byKey(const ValueKey('ai-input')), 'x');
      await tester.tap(find.byKey(const ValueKey('ai-send')));
      await tester.pumpAndSettle();
      expect(
        find.text('Batas harian tercapai, coba lagi besok.'),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('ai-ad-watch')), findsNothing);
    },
  );

  testWidgets(
    'a Lifetime user past the daily messages can watch an ad for 5 more',
    (tester) async {
      final repository = _FakeRepository()
        ..responses.addAll([
          const AiChatException(
            AiChatException.premiumAdOffer,
            'habis',
            messagesPerAd: 5,
          ),
          'Jawaban Lifetime',
        ]);
      final shower = _FakeShower(RewardedAdOutcome.watched);
      await _pump(
        tester,
        repository: repository,
        shower: shower,
        lifetimePlacement: _placement,
      );
      await tester.enterText(find.byKey(const ValueKey('ai-input')), 'Lagi');
      await tester.tap(find.byKey(const ValueKey('ai-send')));
      await tester.pumpAndSettle();
      // Optional: nothing plays until the user chooses to watch.
      expect(shower.shown, isEmpty);
      expect(
        find.text(
          'Tonton iklan singkat untuk menambah 5 pesan. Pertanyaan ini langsung dikirim.',
        ),
        findsOneWidget,
      );
      await tester.tap(find.byKey(const ValueKey('ai-ad-watch')));
      await tester.pumpAndSettle();
      expect(repository.adCredits, [true]);
      expect(repository.sent, hasLength(2));
      expect(find.textContaining('Jawaban Lifetime'), findsOneWidget);
    },
  );

  testWidgets('a Lifetime user gets nothing for an ad that did not load', (
    tester,
  ) async {
    final repository = _FakeRepository()
      ..responses.add(
        const AiChatException(
          AiChatException.premiumAdOffer,
          'habis',
          messagesPerAd: 5,
        ),
      );
    await _pump(
      tester,
      repository: repository,
      shower: _FakeShower(RewardedAdOutcome.unavailable),
      lifetimePlacement: _placement,
    );
    await tester.enterText(find.byKey(const ValueKey('ai-input')), 'Lagi');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('ai-ad-watch')));
    await tester.pumpAndSettle();
    expect(repository.adCredits, isEmpty);
    expect(repository.sent, hasLength(1));
  });

  testWidgets('Lifetime on iOS (no ad placement) stops at the daily limit', (
    tester,
  ) async {
    final repository = _FakeRepository()
      ..responses.add(
        const AiChatException(
          AiChatException.premiumAdOffer,
          'habis',
          messagesPerAd: 5,
        ),
      );
    await _pump(tester, repository: repository, placement: _placement);
    await tester.enterText(find.byKey(const ValueKey('ai-input')), 'x');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('ai-ad-watch')), findsNothing);
    expect(
      find.text('Batas harian tercapai, coba lagi besok.'),
      findsOneWidget,
    );
  });

  testWidgets('an upstream failure offers a retry of the same message', (
    tester,
  ) async {
    final repository = _FakeRepository()
      ..responses.addAll([
        const AiChatException(
          AiChatException.upstream,
          'Jawaban gagal dibuat.',
        ),
        'Berhasil',
      ]);
    await _pump(tester, repository: repository);
    await tester.enterText(find.byKey(const ValueKey('ai-input')), 'Halo');
    await tester.tap(find.byKey(const ValueKey('ai-send')));
    await tester.pumpAndSettle();
    expect(find.text('Jawaban gagal dibuat.'), findsOneWidget);
    await tester.tap(find.text('Coba lagi'));
    await tester.pumpAndSettle();
    expect(repository.sent.map((s) => s.message), ['Halo', 'Halo']);
    expect(find.textContaining('Berhasil'), findsOneWidget);
  });

  testWidgets('offline disables sending and explains why', (tester) async {
    final repository = _FakeRepository();
    await _pump(tester, repository: repository, offline: true);
    expect(find.byKey(const ValueKey('ai-offline')), findsOneWidget);
    final input = tester.widget<TextField>(
      find.byKey(const ValueKey('ai-input')),
    );
    expect(input.enabled, isFalse);
    await tester.tap(find.byKey(const ValueKey('ai-chip-summary')));
    await tester.pumpAndSettle();
    expect(repository.sent, isEmpty);
  });

  testWidgets('stored history is shown when the sheet opens', (tester) async {
    final repository = _FakeRepository()
      ..stored = const [
        AiChatMessage(role: AiRole.user, content: '', chip: AiChip.summary),
        AiChatMessage(role: AiRole.assistant, content: '## Inti\n- poin satu'),
      ];
    await _pump(tester, repository: repository);
    expect(find.text('Ringkas'), findsWidgets);
    expect(find.text('Inti'), findsOneWidget);
    expect(find.text('poin satu'), findsOneWidget);
  });

  group('restricted markdown', () {
    test('parses headings, lists and rules', () {
      final blocks = parseAiMarkdown(
        '# Judul\n\n- satu\n  - dua\n1. pertama\n---\nteks',
      );
      expect(blocks.map((b) => b.kind), [
        AiBlockKind.heading,
        AiBlockKind.spacer,
        AiBlockKind.bullet,
        AiBlockKind.bullet,
        AiBlockKind.numbered,
        AiBlockKind.rule,
        AiBlockKind.paragraph,
      ]);
      expect(blocks[3].level, 1);
      expect(blocks[4].number, 1);
    });

    test('links, images and HTML never survive', () {
      final text = parseAiMarkdown(
        'Klik [di sini](https://evil.example) ![x](https://evil.example/a.png) '
        '<a href="https://evil.example">tag</a><script>alert(1)</script>',
      ).map((b) => b.text).join();
      expect(text, isNot(contains('http')));
      expect(text, isNot(contains('<')));
      expect(text, contains('di sini'));
    });

    testWidgets('renders bold text without any tappable link', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(body: AiMarkdown('**は** vs [が](https://x.y)')),
        ),
      );
      expect(find.textContaining('は'), findsOneWidget);
      expect(find.byType(InkWell), findsNothing);
      expect(find.byType(GestureDetector), findsNothing);
    });
  });

  test('server-sent events are parsed across chunk boundaries', () async {
    const raw =
        'event: delta\ndata: {"text":"こん"}\n\n'
        'event: delta\ndata: {"text":"にちは"}\n\n'
        'event: done\ndata: {"cached":true}\n\n';
    final bytes = utf8.encode(raw);
    final chunks = Stream<List<int>>.fromIterable([
      bytes.sublist(0, 7),
      bytes.sublist(7, 40),
      bytes.sublist(40),
    ]);
    final events = await parseServerSentEvents(chunks).toList();
    expect(events.map((e) => e.name), ['delta', 'delta', 'done']);
    expect(events[1].data['text'], 'にちは');
    expect(events[2].data['cached'], isTrue);
  });

  test('chips map to the backend wire names', () {
    expect(AiChip.values.map((c) => c.wire), [
      'summary',
      'explain',
      'examples',
      'quiz',
      'correct',
    ]);
    expect(AiChip.correct.needsText, isTrue);
    expect(AiChip.summary.needsText, isFalse);
  });

  test('rewarded placement is parsed and swapped for the test unit', () {
    final config = AdsConfig.fromJson(const {
      'enabled': true,
      'aiChatRewarded': {
        'enabled': true,
        'unitId': 'ca-app-pub-1234567890123456/1234567890',
      },
    });
    expect(config.aiChatRewarded.usable, isTrue);
    expect(config.withTestUnits().aiChatRewarded.unitId, testRewardedAdUnitId);
  });

  test('controller key equality keeps one conversation per lesson', () {
    const a = (type: AiLessonType.kanji, lessonId: 'k1');
    const b = (type: AiLessonType.kanji, lessonId: 'k1');
    expect(a == b, isTrue);
  });
}
