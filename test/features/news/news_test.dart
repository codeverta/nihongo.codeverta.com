import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/dictionary/data/dictionary_repository.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/news/presentation/furigana_text.dart';
import 'package:nihongo_master/features/news/presentation/news_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

const text = '😀東京で食べました。東京の文化について読みます。';
const tokens = [
  NewsToken(
    start: 0,
    end: 2,
    surface: '😀',
    base: '😀',
    reading: '',
    readingSource: 'unknown',
  ),
  NewsToken(
    start: 2,
    end: 4,
    surface: '東京',
    base: '東京',
    reading: 'とうきょう',
    readingSource: 'automatic',
  ),
  NewsToken(
    start: 4,
    end: 5,
    surface: 'で',
    base: 'で',
    reading: '',
    readingSource: 'unknown',
  ),
  NewsToken(
    start: 5,
    end: 7,
    surface: '食べ',
    base: '食べる',
    reading: 'たべ',
    readingSource: 'automatic',
  ),
  NewsToken(
    start: 7,
    end: 10,
    surface: 'ました',
    base: 'ます',
    reading: '',
    readingSource: 'unknown',
  ),
  NewsToken(
    start: 10,
    end: 11,
    surface: '。',
    base: '。',
    reading: '',
    readingSource: 'unknown',
  ),
  NewsToken(
    start: 11,
    end: 13,
    surface: '東京',
    base: '東京',
    reading: 'とうきょう',
    readingSource: 'automatic',
  ),
];

final article = NewsArticle({
  'id': 'one',
  'sourceId': 'global-voices',
  'sourceName': 'Global Voices 日本語',
  'kind': 'regular',
  'title': '文化のニュース',
  'titleTokens': <Map<String, dynamic>>[],
  'summary': text,
  'url': 'https://jp.globalvoices.org/2026/10/09/123/',
  'publishedAt': '2026-10-09T12:00:00Z',
  'fetchedAt': '2026-10-19T11:00:00Z',
  'license': 'CC BY 3.0',
  'licenseUrl': 'https://creativecommons.org/licenses/by/3.0/',
  'changes': 'Furigana ditambahkan oleh Nihongo Master.',
  'credits': [
    {
      'name': 'Author',
      'url': 'https://globalvoices.org/author/example/',
      'role': 'author',
    },
    {
      'name': 'Translator',
      'url': 'https://jp.globalvoices.org/author/example/',
      'role': 'translator',
    },
  ],
  'references': <Map<String, dynamic>>[],
  'paragraphs': [
    {
      'text': text,
      'kind': 'text',
      'tokens': [
        for (final t in tokens)
          {
            'start': t.start,
            'end': t.end,
            'surface': t.surface,
            'base': t.base,
            'reading': t.reading,
            'readingSource': t.readingSource,
          },
      ],
    },
  ],
});

class FakeNews extends NewsRepository {
  FakeNews() : super(Dio());
  int lists = 0;
  int details = 0;
  @override
  Future<NewsPage> list({
    int page = 1,
    int limit = 20,
    String source = '',
  }) async {
    lists++;
    return NewsPage(items: [article], total: 1, page: page, sources: []);
  }

  @override
  Future<NewsArticle> detail(String id) async {
    details++;
    return article;
  }
}

class NewsDictionary extends DictionaryRepository {
  NewsDictionary() : super(Dio());
  final calls = <(String, String)>[];
  @override
  Future<DictionaryPage> search(
    String query, {
    int page = 1,
    String? jlpt,
    bool common = false,
    String mode = 'auto',
  }) async {
    calls.add((query, mode));
    return DictionaryPage(
      entries: [
        DictionaryEntry({
          'slug': query,
          'japanese': [
            {'word': query, 'reading': 'とうきょう'},
          ],
          'senses': [
            {
              'english_definitions': ['Tokyo city'],
            },
          ],
        }),
      ],
      hasMore: false,
    );
  }
}

void main() {
  test('publisher calendar date stays intact across JST and WIB', () {
    final a = NewsArticle({
      ...article.json,
      'publishedAt': '2026-10-13T15:00:00Z',
      'originalDate': '2026-10-14',
    });
    expect(newsDate(a.sourceDate), '14 Okt 2026');
    final cached = NewsArticle({
      ...article.json,
      'publishedAt': '2026-10-13T15:00:00Z',
    });
    expect(newsDate(cached.sourceDate), '14 Okt 2026');
  });
  var clipboard = '';
  setUp(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(SystemChannels.platform, (call) async {
          if (call.method == 'Clipboard.setData') {
            clipboard = (call.arguments as Map)['text'] as String;
          }
          if (call.method == 'Clipboard.getData') {
            return {'text': clipboard};
          }
          if (call.method == 'Clipboard.hasStrings') {
            return {'value': clipboard.isNotEmpty};
          }
          return null;
        });
  });
  tearDown(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(SystemChannels.platform, null);
  });
  test(
    'selection uses UTF-16 bounds, expands partial words and deduplicates lemma',
    () {
      expect(
        newsSelectionTokens(
          tokens,
          const TextSelection(baseOffset: 3, extentOffset: 6),
        ).map((t) => t.base),
        ['東京', 'で', '食べる'],
      );
      expect(
        newsSelectionTokens(
          tokens,
          const TextSelection(baseOffset: 13, extentOffset: 2),
        ).where((t) => t.base == '東京'),
        hasLength(1),
      );
      expect(
        newsSelectionTokens(
          tokens,
          const TextSelection(baseOffset: 0, extentOffset: 2),
        ),
        isEmpty,
      );
      expect(
        newsSelectionTokens(tokens, const TextSelection.collapsed(offset: 3)),
        isEmpty,
      );
      expect(requiresPremium(Uri.parse('/news/one')), isFalse);
    },
  );

  Future<SharedPreferences> prefs() async {
    SharedPreferences.setMockInitialValues({});
    return SharedPreferences.getInstance();
  }

  Future<void> reader(
    WidgetTester tester,
    SharedPreferences preferences,
    NewsDictionary dictionary, {
    double scale = 1,
  }) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          newsRepositoryProvider.overrideWithValue(FakeNews()),
          dictionaryRepositoryProvider.overrideWithValue(dictionary),
        ],
        child: MaterialApp(
          home: MediaQuery(
            data: MediaQueryData(textScaler: TextScaler.linear(scale)),
            child: const NewsReaderScreen(id: 'one'),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
  }

  EditableTextState paragraphState(WidgetTester tester) =>
      tester.state<EditableTextState>(
        find.descendant(
          of: find.byType(FuriganaText).at(1),
          matching: find.byType(EditableText),
        ),
      );

  Future<void> select(WidgetTester tester, int start, int end) async {
    final state = paragraphState(tester);
    state
      ..userUpdateTextEditingValue(
        state.widget.controller.value.copyWith(
          selection: TextSelection(baseOffset: start, extentOffset: end),
        ),
        SelectionChangedCause.longPress,
      )
      ..showToolbar();
    await tester.pumpAndSettle();
  }

  testWidgets(
    'native selection copies base text with ruby and large font across lines',
    (tester) async {
      tester.view
        ..physicalSize = const Size(360, 900)
        ..devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await reader(tester, await prefs(), NewsDictionary(), scale: 1.7);
      final state = paragraphState(tester);
      const selection = TextSelection(baseOffset: 2, extentOffset: 13);
      final boxes = state.renderEditable.getBoxesForSelection(selection);
      expect(boxes.length, greaterThan(1));
      state
        ..userUpdateTextEditingValue(
          state.widget.controller.value.copyWith(selection: selection),
          SelectionChangedCause.longPress,
        )
        ..copySelection(SelectionChangedCause.toolbar);
      await tester.pump();
      expect(
        (await Clipboard.getData(Clipboard.kTextPlain))?.text,
        '東京で食べました。東京',
      );
      expect(state.widget.controller.text, text);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'word lookup uses existing Japanese dictionary and sentence waits for a word',
    (tester) async {
      final dictionary = NewsDictionary();
      await reader(tester, await prefs(), dictionary);
      await select(tester, 5, 7);
      await tester.tap(find.text('Cari arti'));
      await tester.pumpAndSettle();
      expect(dictionary.calls, [('食べる', 'japanese')]);
      expect(find.text('Tokyo city'), findsOneWidget);
      Navigator.of(tester.element(find.text('Kamus bacaan'))).pop();
      await tester.pumpAndSettle();
      await select(tester, 2, 13);
      await tester.tap(find.text('Cari arti'));
      await tester.pumpAndSettle();
      expect(dictionary.calls, hasLength(1));
      await tester.tap(find.widgetWithText(ChoiceChip, '東京'));
      await tester.pumpAndSettle();
      expect(dictionary.calls.last, ('東京', 'japanese'));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('furigana toggle persists on reopening the reader', (
    tester,
  ) async {
    final preferences = await prefs();
    final dictionary = NewsDictionary();
    await reader(tester, preferences, dictionary);
    expect(
      tester.widget<FuriganaText>(find.byType(FuriganaText).first).showFurigana,
      isTrue,
    );
    await tester.tap(find.byKey(const Key('news-furigana-toggle')));
    await tester.pumpAndSettle();
    expect(preferences.getBool(newsFuriganaPreferenceKey), isFalse);
    await tester.pumpWidget(const SizedBox());
    await reader(tester, preferences, dictionary);
    expect(
      tester.widget<FuriganaText>(find.byType(FuriganaText).first).showFurigana,
      isFalse,
    );
    await tester.scrollUntilVisible(
      find.text('Penerjemah: Translator'),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.text('Penerjemah: Translator'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('preview navigates to list and article without login', (
    tester,
  ) async {
    final repository = FakeNews();
    final router = GoRouter(
      routes: [
        GoRoute(
          path: '/',
          builder: (_, state) =>
              const Scaffold(body: SingleChildScrollView(child: NewsPreview())),
        ),
        GoRoute(path: '/news', builder: (_, state) => const NewsListScreen()),
        GoRoute(
          path: '/news/:id',
          builder: (_, state) =>
              NewsReaderScreen(id: state.pathParameters['id']!),
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(await prefs()),
          newsRepositoryProvider.overrideWithValue(repository),
        ],
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Lihat semua'));
    await tester.pumpAndSettle();
    expect(find.text('Berita Jepang'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('news-one')));
    await tester.pumpAndSettle();
    expect(find.text('Baca berita'), findsOneWidget);
    expect(repository.details, 1);
    expect(tester.takeException(), isNull);
  });
}
