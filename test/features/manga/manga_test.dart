import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/manga/data/manga_repository.dart';
import 'package:nihongo_master/features/manga/presentation/manga_screens.dart';

const _series = MangaSeries(
  id: 'one',
  title: 'Test Manga',
  japaneseTitle: '物語',
  summary: '日本語の物語です。',
  author: 'Test Author',
  coverUrl: '',
  sourceUrl: 'https://rawkuma.net/manga/test/',
  genres: ['Comedy', 'School Life'],
  chapterCount: 2,
);
MangaChapter _chapter(String status) => MangaChapter(
  id: 'chapter-one',
  seriesId: 'one',
  title: 'Chapter 1',
  number: '1',
  sourceUrl: 'https://rawkuma.net/manga/test/chapter-1.100/',
  status: status,
  pages: status == 'ready'
      ? const [
          MangaPage(
            url: 'https://our-cache.example/one.png',
            width: 800,
            height: 1200,
          ),
          MangaPage(
            url: 'https://our-cache.example/two.png',
            width: 800,
            height: 1200,
          ),
        ]
      : [],
);

class _Repository extends MangaRepository {
  _Repository() : super(Dio());
  final requests = <MangaFilter>[];
  int chapterCalls = 0;
  int prepareCalls = 0;
  bool ready = false;
  Completer<MangaChapter>? blocked;
  @override
  Future<MangaCatalog> list({
    int page = 1,
    String q = '',
    String genre = '',
  }) async {
    requests.add((page: page, q: q, genre: genre));
    return MangaCatalog(
      items: q == 'unknown' ? [] : [_series],
      total: 41,
      page: page,
    );
  }

  @override
  Future<MangaDetail> detail(String id) async =>
      MangaDetail(series: _series, chapters: [_chapter('available')]);
  @override
  Future<MangaChapter> chapter(String id) async {
    chapterCalls++;
    if (blocked != null) return blocked!.future;
    return _chapter(
      ready
          ? 'ready'
          : prepareCalls == 0
          ? 'available'
          : 'queued',
    );
  }

  @override
  Future<MangaChapter> prepare(String id) async {
    prepareCalls++;
    return _chapter('queued');
  }
}

Widget _app(
  _Repository repository, {
  String location = '/manga',
  Locale locale = const Locale('id'),
}) => ProviderScope(
  overrides: [mangaRepositoryProvider.overrideWithValue(repository)],
  child: MaterialApp.router(
    locale: locale,
    supportedLocales: AppLocalizations.supportedLocales,
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    routerConfig: GoRouter(
      initialLocation: location,
      routes: [
        GoRoute(path: '/manga', builder: (_, __) => const MangaCatalogScreen()),
        GoRoute(
          path: '/manga/chapters/:id',
          builder: (_, state) =>
              MangaReaderScreen(id: state.pathParameters['id']!),
        ),
        GoRoute(
          path: '/manga/:id',
          builder: (_, state) =>
              MangaDetailScreen(id: state.pathParameters['id']!),
        ),
      ],
    ),
  ),
);

void main() {
  testWidgets('manga catalog localizes controls and empty state', (
    tester,
  ) async {
    final repository = _Repository();
    await tester.pumpWidget(_app(repository, locale: const Locale('en')));
    await tester.pumpAndSettle();
    expect(find.text('Japanese Manga'), findsOneWidget);
    expect(
      find.text('Great stories.\nIn the original language.'),
      findsOneWidget,
    );
    await tester.enterText(find.byKey(const Key('manga-search')), 'unknown');
    await tester.testTextInput.receiveAction(TextInputAction.search);
    await tester.pumpAndSettle();
    expect(find.text('No manga found'), findsOneWidget);
    expect(
      find.text(
        'Curated titles will appear here. Try refreshing in a little while.',
      ),
      findsOneWidget,
    );
    await tester.pumpWidget(
      _app(repository, location: '/manga/one', locale: const Locale('en')),
    );
    await tester.pumpAndSettle();
    expect(find.text('物語'), findsOneWidget);
    expect(find.text('日本語の物語です。'), findsOneWidget);
    expect(find.text('About this story'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'catalog on a small phone filters stored results and resets pagination',
    (tester) async {
      tester.view.physicalSize = const Size(360, 640);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repo = _Repository();
      await tester.pumpWidget(_app(repo));
      await tester.pumpAndSettle();
      expect(find.text('Sumber & atribusi ↗'), findsOneWidget);
      await tester.scrollUntilVisible(
        find.byKey(const Key('manga-next-page')).hitTestable(),
        180,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(find.byKey(const Key('manga-next-page')));
      await tester.pumpAndSettle();
      expect(repo.requests.last.page, 2);
      await tester.scrollUntilVisible(
        find.byKey(const Key('manga-search')).hitTestable(),
        -180,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.enterText(find.byKey(const Key('manga-search')), 'unknown');
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle();
      expect(repo.requests.last, (page: 1, q: 'unknown', genre: ''));
      await tester.tap(find.byKey(const ValueKey('manga-genre-Sports')));
      await tester.pumpAndSettle();
      expect(repo.requests.last, (page: 1, q: 'unknown', genre: 'Sports'));
      expect(repo.prepareCalls, 0);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('detail opens reader without silently preparing a chapter', (
    tester,
  ) async {
    final repo = _Repository();
    await tester.pumpWidget(_app(repo, location: '/manga/one'));
    await tester.pumpAndSettle();
    expect(find.text('Test Author'), findsOneWidget);
    expect(find.text('Sumber: Rawkuma'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Mulai membaca'),
      150,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.tap(find.text('Mulai membaca'));
    await tester.pumpAndSettle();
    expect(find.text('Siapkan chapter'), findsOneWidget);
    expect(repo.chapterCalls, 1);
    expect(repo.prepareCalls, 0);
    await tester.tap(find.text('Siapkan chapter'));
    await tester.pumpAndSettle();
    expect(find.text('Chapter sedang disiapkan'), findsOneWidget);
    expect(repo.prepareCalls, 1);
    repo.ready = true;
    await tester.pump(const Duration(seconds: 15));
    await tester.pumpAndSettle();
    expect(tester.widget<PageView>(find.byType(PageView)).reverse, isTrue);
    expect(find.text('1 / 2'), findsOneWidget);
    expect(repo.prepareCalls, 1);
    await tester.tap(find.byTooltip('Halaman berikutnya'));
    await tester.pumpAndSettle();
    expect(find.text('2 / 2'), findsOneWidget);
    await tester.tap(find.byTooltip('Baca dengan scroll'));
    await tester.pumpAndSettle();
    expect(find.byType(PageView), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('disposing a pending reader cancels polling', (tester) async {
    final repo = _Repository()..prepareCalls = 1;
    await tester.pumpWidget(
      _app(repo, location: '/manga/chapters/chapter-one'),
    );
    await tester.pumpAndSettle();
    expect(repo.chapterCalls, 1);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    await tester.pump(const Duration(seconds: 45));
    expect(repo.chapterCalls, 1);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpAndSettle();
    expect(repo.chapterCalls, 2);
    await tester.pumpWidget(const SizedBox());
    await tester.pump(const Duration(seconds: 45));
    expect(repo.chapterCalls, 2);
    expect(tester.takeException(), isNull);
  });

  testWidgets('late reader response does not update a disposed widget', (
    tester,
  ) async {
    final repo = _Repository()..blocked = Completer<MangaChapter>();
    await tester.pumpWidget(
      _app(repo, location: '/manga/chapters/chapter-one'),
    );
    await tester.pump();
    await tester.pumpWidget(const SizedBox());
    repo.blocked!.complete(_chapter('available'));
    await tester.pump();
    expect(tester.takeException(), isNull);
  });
}
