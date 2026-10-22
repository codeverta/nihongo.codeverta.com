import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/dictionary/data/dictionary_repository.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_controller.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_details.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

final cat = DictionaryEntry({
  'slug': '猫',
  'is_common': true,
  'jlpt': ['jlpt-n5'],
  'japanese': [
    {'word': '猫', 'reading': 'ねこ'},
    {'reading': 'ネコ'},
  ],
  'senses': [
    {
      'english_definitions': ['cat', 'feline'],
      'parts_of_speech': ['Noun'],
      'tags': <String>[],
      'see_also': <String>[],
      'antonyms': <String>[],
    },
    {
      'english_definitions': ['shamisen'],
      'parts_of_speech': ['Noun'],
      'tags': <String>[],
      'see_also': <String>[],
      'antonyms': <String>[],
    },
  ],
  'attribution': {'jmdict': true},
});
final dog = DictionaryEntry({
  'slug': '犬',
  'japanese': [
    {'word': '犬', 'reading': 'いぬ'},
  ],
  'senses': [
    {
      'english_definitions': ['dog'],
    },
  ],
});

class FakeDictionary extends DictionaryRepository {
  FakeDictionary() : super(Dio());
  Future<DictionaryPage> Function(String, int)? fetch;
  bool fail = false;
  bool failKanji = true;
  final calls = <(String, int, String?, bool)>[];
  final modes = <String>[];
  @override
  Future<Map<String, dynamic>> kanji(String character) async {
    if (failKanji) throw StateError('offline');
    return {
      'kanji': character,
      'stroke_count': 11,
      'jlpt': 3,
      'meanings': ['cat'],
      'on_readings': ['ビョウ'],
      'kun_readings': ['ねこ'],
      'name_readings': <String>[],
    };
  }

  @override
  Future<DictionaryPage> search(
    String query, {
    int page = 1,
    String? jlpt,
    bool common = false,
    String mode = 'auto',
  }) async {
    calls.add((query, page, jlpt, common));
    modes.add(mode);
    if (fail) throw StateError('offline');
    if (fetch != null) return fetch!(query, page);
    return DictionaryPage(
      entries: page == 1 ? [cat] : [cat, dog],
      hasMore: page == 1,
    );
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('kanji failure can retry and legacy JLPT is labeled accurately', (
    tester,
  ) async {
    final repository = FakeDictionary();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [dictionaryRepositoryProvider.overrideWithValue(repository)],
        child: MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => TextButton(
                onPressed: () => showKanjiDetails(context, '猫'),
                child: const Text('Kanji'),
              ),
            ),
          ),
        ),
      ),
    );
    await tester.tap(find.text('Kanji'));
    await tester.pumpAndSettle();
    expect(
      find.text('Detail kanji belum dapat dimuat dari sumber kamus.'),
      findsOneWidget,
    );
    repository.failKanji = false;
    await tester.tap(find.text('Coba lagi'));
    await tester.pumpAndSettle();
    expect(find.text('11 goresan'), findsOneWidget);
    expect(find.text('JLPT lama 3'), findsOneWidget);
    expect(find.text('ビョウ'), findsOneWidget);
    expect(find.text('ねこ'), findsOneWidget);
    expect(find.text('JLPT N3'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  test(
    'kana-only entries retain their reading; every sense remains available',
    () {
      final kana = DictionaryEntry({
        'slug': 'ありがとう',
        'japanese': [
          {'reading': 'ありがとう'},
        ],
        'senses': <Map<String, dynamic>>[],
      });
      expect(kana.word, 'ありがとう');
      expect(cat.jlpt, ['N5']);
      expect(cat.senses.length, 2);
      expect(cat.kanji, ['猫']);
    },
  );

  test(
    'stale searches cannot replace newer results; filters reach the repository',
    () async {
      final old = Completer<DictionaryPage>();
      final repository = FakeDictionary()
        ..fetch = (query, page) async => query == 'old'
            ? await old.future
            : DictionaryPage(entries: [dog], hasMore: false);
      final controller = DictionaryController(repository);
      addTearDown(controller.dispose);
      final pending = controller.search('old');
      await controller.search('new');
      old.complete(DictionaryPage(entries: [cat], hasMore: false));
      await pending;
      expect(controller.state.entries.single.id, '犬');
      await controller.filter('N3', common: true);
      expect(repository.calls.last, ('new', 1, 'N3', true));
      await controller.filter(null, common: false);
      expect(repository.calls.last, ('new', 1, null, false));
    },
  );

  test(
    'pagination deduplicates words and preserves results after a failed page',
    () async {
      final repository = FakeDictionary();
      final controller = DictionaryController(repository);
      addTearDown(controller.dispose);
      await controller.search('cat');
      repository.fail = true;
      await controller.more();
      expect(controller.state.entries.single.id, '猫');
      expect(controller.state.page, 1);
      expect(controller.state.error, isNotNull);
      repository.fail = false;
      await controller.more();
      expect(controller.state.entries.map((e) => e.id), ['猫', '犬']);
      expect(controller.state.hasMore, isFalse);
      expect(controller.state.error, isNull);
    },
  );

  test(
    'empty and oversized queries never fetch; clear invalidates pending results',
    () async {
      final result = Completer<DictionaryPage>();
      final repository = FakeDictionary()..fetch = (_, page) => result.future;
      final controller = DictionaryController(repository);
      addTearDown(controller.dispose);
      await controller.search(' ');
      await controller.search('猫' * 101);
      expect(repository.calls, isEmpty);
      final pending = controller.search('cat');
      await controller.search('');
      result.complete(DictionaryPage(entries: [cat], hasMore: false));
      await pending;
      expect(controller.state.entries, isEmpty);
      expect(controller.state.query, '');
    },
  );

  test(
    'mode reaches API and survives filters, pagination and search races',
    () async {
      final repository = FakeDictionary();
      final controller = DictionaryController(repository);
      addTearDown(controller.dispose);
      await controller.search('pan');
      await controller.setMode('japanese');
      expect(repository.modes.last, 'japanese');
      await controller.filter('N5', common: true);
      await controller.more();
      expect(repository.modes.last, 'japanese');
      expect(repository.calls.last, ('pan', 2, 'N5', true));
      final old = Completer<DictionaryPage>();
      repository.fetch = (_, page) => old.future;
      final pending = controller.setMode('english');
      repository.fetch = (_, page) async =>
          DictionaryPage(entries: [dog], hasMore: false);
      await controller.setMode('auto');
      old.complete(DictionaryPage(entries: [cat], hasMore: false));
      await pending;
      expect(controller.state.mode, 'auto');
      expect(controller.state.entries.single.id, dog.id);
    },
  );

  test('repository sends mode as a backend query parameter', () async {
    final dio = Dio(BaseOptions(baseUrl: 'https://api.example/v1'));
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          expect(options.queryParameters['mode'], 'english');
          expect(options.queryParameters['q'], 'pan');
          handler.resolve(
            Response<Map<String, dynamic>>(
              requestOptions: options,
              data: {
                'data': {'entries': <dynamic>[], 'hasMore': false},
              },
              statusCode: 200,
            ),
          );
        },
      ),
    );
    await DictionaryRepository(dio).search('pan', mode: 'english');
    dio.close();
  });

  test(
    'applying filters uses one request and can clear the JLPT level',
    () async {
      final repository = FakeDictionary();
      final controller = DictionaryController(repository);
      addTearDown(controller.dispose);
      await controller.search('cat');
      final before = repository.calls.length;
      await controller.filter(
        'N5',
        common: true,
        mode: 'japanese',
        text: 'neko',
      );
      expect(repository.calls.length, before + 1);
      expect(repository.calls.last, ('neko', 1, 'N5', true));
      expect(repository.modes.last, 'japanese');
      await controller.filter(null, common: false, mode: 'auto');
      expect(repository.calls.last, ('neko', 1, null, false));
      expect(repository.modes.last, 'auto');
    },
  );

  test(
    'favorites persist offline, concurrent saves survive, accounts are isolated',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final library = DictionaryLibraryController(preferences, 'buyer');
      addTearDown(library.dispose);
      await Future.wait([
        library.toggle(cat),
        library.toggle(dog),
        library.remember('neko'),
      ]);
      expect(library.state.saved.length, 2);
      final restored = DictionaryLibraryController(preferences, 'buyer');
      final other = DictionaryLibraryController(preferences, 'other');
      addTearDown(restored.dispose);
      addTearDown(other.dispose);
      expect(restored.state.saved.first.word, '犬');
      expect(restored.state.recent, ['neko']);
      expect(other.state.saved, isEmpty);
      await restored.clearHistory();
      expect(restored.state.recent, isEmpty);
      expect(restored.state.saved.length, 2);
      await restored.toggle(cat);
      expect(restored.state.saved.single.word, '犬');
      await other.toggle(cat);
      final dio = Dio(BaseOptions(baseUrl: 'https://api.example/v1'));
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            expect(options.method, 'DELETE');
            expect(options.path, '/me/account');
            handler.resolve(
              Response<void>(requestOptions: options, statusCode: 204),
            );
          },
        ),
      );
      await AuthRepository(dio, preferences).deleteAccount(userId: 'buyer');
      expect(preferences.getString('dictionary_library_v1_buyer'), isNull);
      expect(preferences.getString('dictionary_library_v1_other'), isNotNull);
    },
  );

  testWidgets(
    'filter sheet cancels drafts and resets applied filters on a narrow phone',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final repository = FakeDictionary();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPreferencesProvider.overrideWithValue(preferences),
            dictionaryRepositoryProvider.overrideWithValue(repository),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(const AuthenticationState.guest()),
            ),
          ],
          child: MaterialApp(
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.4)),
              child: child!,
            ),
            home: const DictionaryScreen(),
          ),
        ),
      );
      await tester.enterText(find.byType(TextField), 'neko');
      await tester.tap(find.text('Filter'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ChoiceChip, 'Inggris'));
      await tester.tap(find.byTooltip('Tutup filter'));
      await tester.pumpAndSettle();
      expect(repository.calls, isEmpty);
      expect(find.text('Bahasa otomatis'), findsOneWidget);
      await tester.tap(find.text('Filter'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.byType(DropdownButtonFormField<String>));
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text('N5').last);
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Kata umum saja'));
      await tester.tap(find.text('Kata umum saja'));
      await tester.ensureVisible(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      expect(repository.calls.single, ('neko', 1, 'N5', true));
      expect(find.text('Filter · 2'), findsOneWidget);
      await tester.ensureVisible(find.text('Filter · 2'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Filter · 2'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Reset filter'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Reset filter'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      expect(repository.calls.length, 2);
      expect(repository.calls.last, ('neko', 1, null, false));
      await tester.scrollUntilVisible(find.text('Filter'), -100);
      await tester.pumpAndSettle();
      expect(find.text('Filter'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'phone searches real-shaped entries, saves and reads all senses',
    (tester) async {
      tester.view.physicalSize = const Size(360, 800);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final repository = FakeDictionary();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPreferencesProvider.overrideWithValue(preferences),
            dictionaryRepositoryProvider.overrideWithValue(repository),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(const AuthenticationState.guest()),
            ),
          ],
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: const DictionaryScreen(),
          ),
        ),
      );
      await tester.enterText(find.byType(TextField), 'neko');
      await tester.tap(find.text('Cari di kamus'));
      await tester.pumpAndSettle();
      expect(find.byType(ChoiceChip), findsNothing);
      await tester.tap(find.text('Filter'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ChoiceChip, 'Inggris'));
      // Editing is a draft; it must not search until Apply is pressed.
      expect(repository.modes.last, 'auto');
      await tester.tap(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      expect(repository.modes.last, 'english');
      await tester.enterText(find.byType(TextField), 'konnichiwa');
      await tester.tap(find.text('Filter · 1'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ChoiceChip, 'Jepang'));
      await tester.tap(find.text('Terapkan filter'));
      await tester.pumpAndSettle();
      expect(repository.modes.last, 'japanese');
      expect(repository.calls.last.$1, 'konnichiwa');
      await tester.ensureVisible(find.text('猫'));
      expect(find.text('ねこ'), findsOneWidget);
      expect(find.text('cat; feline'), findsOneWidget);
      await tester.tap(find.byTooltip('Simpan 猫'));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Hapus 猫 dari favorit'), findsOneWidget);
      await tester.tap(find.text('猫'));
      await tester.pumpAndSettle();
      expect(find.text('2. shamisen'), findsOneWidget);
      expect(find.text('Bentuk dan bacaan lain'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.tap(find.byTooltip('Tutup detail'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.byTooltip('Kata tersimpan'));
      await tester.tap(find.byTooltip('Kata tersimpan'));
      await tester.pumpAndSettle();
      expect(
        find.text('Favorit tersimpan di perangkat • 1/60'),
        findsOneWidget,
      );
    },
  );
}
