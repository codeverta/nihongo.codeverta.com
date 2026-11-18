import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/level_list/data/level_list_repository.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/grammar_level_list.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/kanji_level_list.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/level_pdf_button.dart';

class _FakeRepository extends LevelListRepository {
  _FakeRepository({this.pdfError}) : super(Dio());

  final LevelPdfException? pdfError;
  final pdfCalls = <(LevelListKind, String)>[];

  @override
  Future<List<GrammarListItem>> grammar(String level) async => const [
    GrammarListItem(
      lessonId: 'g1',
      lessonTitle: 'Dasar',
      lessonNumber: 1,
      pattern: 'は',
      meaning: 'penanda topik',
      formation: 'Nomina + は',
      explanation: 'Menandai topik kalimat.',
      commonMistake: 'Tertukar dengan が.',
      examples: [
        LevelListExample(
          japanese: '私は学生です。',
          reading: 'わたしはがくせいです。',
          meaning: 'Saya pelajar.',
        ),
      ],
    ),
    GrammarListItem(
      lessonId: 'g2',
      lessonTitle: 'Kepemilikan',
      lessonNumber: 2,
      pattern: 'の',
      meaning: 'milik',
      formation: 'Nomina + の + Nomina',
      explanation: '',
      commonMistake: '',
      examples: [],
    ),
  ];

  @override
  Future<List<KanjiListItem>> kanji(String level) async => const [
    KanjiListItem(
      lessonId: 'k1',
      lessonTitle: 'Alam',
      lessonNumber: 1,
      character: '山',
      meaning: 'gunung',
      onReading: 'サン',
      kunReading: 'やま',
      strokes: 3,
      mnemonic: 'Tiga puncak.',
      note: '',
      examples: [],
    ),
    KanjiListItem(
      lessonId: 'k1',
      lessonTitle: 'Alam',
      lessonNumber: 1,
      character: '川',
      meaning: 'sungai',
      onReading: 'セン',
      kunReading: 'かわ',
      strokes: 3,
      mnemonic: '',
      note: '',
      examples: [],
    ),
  ];

  @override
  Future<Uint8List> pdf(LevelListKind kind, String level) async {
    pdfCalls.add((kind, level));
    if (pdfError != null) throw pdfError!;
    return Uint8List.fromList([1, 2, 3]);
  }
}

typedef _Save = ({String name, int length});

Future<void> _pump(
  WidgetTester tester,
  Widget child, {
  required _FakeRepository repository,
  required List<_Save> saves,
  bool lifetime = false,
  bool offline = false,
}) async {
  tester.view.physicalSize = const Size(800, 1600);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  final router = GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => Scaffold(body: child),
      ),
      GoRoute(
        path: '/premium',
        builder: (context, state) => const Scaffold(body: Text('premium')),
      ),
      GoRoute(
        path: '/learn/grammar/lessons/:id',
        builder: (context, state) => const Scaffold(body: Text('lesson')),
      ),
    ],
  );
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        levelListRepositoryProvider.overrideWithValue(repository),
        levelPdfSaverProvider.overrideWithValue(
          (context, bytes, name) async =>
              saves.add((name: name, length: bytes.length)),
        ),
        offlineProvider.overrideWithValue(offline),
        authControllerProvider.overrideWith(
          (ref) => AuthController.test(
            AuthenticationState.authenticated(
              AuthenticatedUser(
                id: 'learner',
                displayName: 'Learner',
                email: 'learner@example.test',
                isLifetime: lifetime,
              ),
            ),
          ),
        ),
      ],
      child: MaterialApp.router(
        locale: const Locale('id'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        // The real theme: its full-width button defaults once blanked the
        // list, which the default test theme could not catch.
        theme: AppTheme.lightTheme,
        routerConfig: router,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('grammar list renders items and search filters', (tester) async {
    await _pump(
      tester,
      const GrammarLevelList(level: 'N5'),
      repository: _FakeRepository(),
      saves: [],
    );
    expect(find.text('2 pola grammar N5'), findsOneWidget);
    expect(
      find.byKey(const ValueKey('grammar-list-item-g1-は')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey('grammar-list-item-g2-の')),
      findsOneWidget,
    );

    await tester.enterText(
      find.byKey(const ValueKey('grammar-list-search-N5')),
      'milik',
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('grammar-list-item-g1-は')), findsNothing);
    expect(
      find.byKey(const ValueKey('grammar-list-item-g2-の')),
      findsOneWidget,
    );

    await tester.enterText(
      find.byKey(const ValueKey('grammar-list-search-N5')),
      'zzz',
    );
    await tester.pumpAndSettle();
    expect(find.text('Tidak ada hasil.'), findsOneWidget);
  });

  testWidgets('grammar item opens a sheet with formation and lesson button', (
    tester,
  ) async {
    await _pump(
      tester,
      const GrammarLevelList(level: 'N5'),
      repository: _FakeRepository(),
      saves: [],
    );
    await tester.tap(find.byKey(const ValueKey('grammar-list-item-g1-は')));
    await tester.pumpAndSettle();
    expect(find.text('Pembentukan'), findsOneWidget);
    expect(find.text('Nomina + は'), findsOneWidget);
    expect(find.text('Kesalahan umum'), findsOneWidget);
    expect(find.text('私は学生です。'), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey('level-list-open-lesson')));
    await tester.pumpAndSettle();
    expect(find.text('lesson'), findsOneWidget);
  });

  testWidgets('kanji grid renders and search by reading filters', (
    tester,
  ) async {
    await _pump(
      tester,
      const KanjiLevelList(level: 'N5'),
      repository: _FakeRepository(),
      saves: [],
    );
    expect(find.text('2 kanji N5'), findsOneWidget);
    expect(find.byKey(const ValueKey('kanji-list-item-山')), findsOneWidget);
    expect(find.byKey(const ValueKey('kanji-list-item-川')), findsOneWidget);

    await tester.enterText(
      find.byKey(const ValueKey('kanji-list-search-N5')),
      'かわ',
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('kanji-list-item-山')), findsNothing);
    expect(find.byKey(const ValueKey('kanji-list-item-川')), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey('kanji-list-item-川')));
    await tester.pumpAndSettle();
    expect(find.text('Bacaan on'), findsOneWidget);
    expect(find.text('Bacaan kun'), findsOneWidget);
    expect(
      find.byKey(const ValueKey('level-list-open-lesson')),
      findsOneWidget,
    );
  });

  testWidgets('free user sees the Lifetime upsell and no pdf call', (
    tester,
  ) async {
    final repository = _FakeRepository();
    await _pump(
      tester,
      const GrammarLevelList(level: 'N5'),
      repository: repository,
      saves: [],
    );
    await tester.tap(find.byKey(const ValueKey('level-pdf-button')));
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('level-pdf-upsell')), findsOneWidget);
    expect(repository.pdfCalls, isEmpty);

    await tester.tap(find.byKey(const ValueKey('level-pdf-upsell')));
    await tester.pumpAndSettle();
    expect(find.text('premium'), findsOneWidget);
  });

  testWidgets('Lifetime user downloads and saves the pdf', (tester) async {
    final repository = _FakeRepository();
    final saves = <_Save>[];
    await _pump(
      tester,
      const GrammarLevelList(level: 'N5'),
      repository: repository,
      saves: saves,
      lifetime: true,
    );
    await tester.tap(find.byKey(const ValueKey('level-pdf-button')));
    await tester.pumpAndSettle();
    expect(repository.pdfCalls, [(LevelListKind.grammar, 'N5')]);
    expect(saves, [(name: 'NihongoMaster-Grammar-N5.pdf', length: 3)]);
    expect(find.byKey(const ValueKey('level-pdf-upsell')), findsNothing);
  });

  testWidgets('Lifetime user downloads a simulation paper', (tester) async {
    final repository = _FakeRepository();
    final saves = <_Save>[];
    await _pump(
      tester,
      const LevelPdfButton(kind: LevelListKind.simulation, level: 'JFT-2'),
      repository: repository,
      saves: saves,
      lifetime: true,
    );
    await tester.tap(find.byKey(const ValueKey('level-pdf-button')));
    await tester.pumpAndSettle();
    expect(repository.pdfCalls, [(LevelListKind.simulation, 'JFT-2')]);
    expect(saves, [(name: 'NihongoMaster-Simulasi-JFT-2.pdf', length: 3)]);
  });

  test('simulation PDFs use their own endpoint', () {
    expect(LevelListKind.simulation.pdfPath('N3-4'), '/simulation-pdf/N3-4');
    expect(LevelListKind.grammar.pdfPath('N5'), '/grammar/levels/N5/pdf');
  });

  testWidgets('offline disables the pdf button', (tester) async {
    final repository = _FakeRepository();
    await _pump(
      tester,
      const KanjiLevelList(level: 'N5'),
      repository: repository,
      saves: [],
      lifetime: true,
      offline: true,
    );
    final button = tester.widget<OutlinedButton>(
      find.byKey(const ValueKey('level-pdf-button')),
    );
    expect(button.onPressed, isNull);
    expect(find.text('Butuh koneksi internet'), findsOneWidget);
  });

  testWidgets('server premium_required shows the upsell for Lifetime', (
    tester,
  ) async {
    final repository = _FakeRepository(
      pdfError: const LevelPdfException(
        LevelPdfException.premiumRequired,
        'Download PDF khusus pengguna Lifetime Access.',
      ),
    );
    final saves = <_Save>[];
    await _pump(
      tester,
      const KanjiLevelList(level: 'N5'),
      repository: repository,
      saves: saves,
      lifetime: true,
    );
    await tester.tap(find.byKey(const ValueKey('level-pdf-button')));
    await tester.pumpAndSettle();
    expect(repository.pdfCalls, [(LevelListKind.kanji, 'N5')]);
    expect(saves, isEmpty);
    expect(find.byKey(const ValueKey('level-pdf-upsell')), findsOneWidget);
  });
}
