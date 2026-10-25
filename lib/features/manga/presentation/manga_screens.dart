import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/core/network/remote_asset_image.dart';
import 'package:nihongo_master/features/manga/data/manga_repository.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';
import 'package:url_launcher/url_launcher.dart';

const _ink = Color(0xFF232667);

String _retryDate(BuildContext context, DateTime value) =>
    regionalDateTime(value, localeName: appRegionLocaleName(context));

Future<void> _openSource(BuildContext context, String raw) async {
  final uri = Uri.tryParse(raw);
  var opened = false;
  try {
    if (uri != null && uri.scheme == 'https' && uri.host == 'rawkuma.net') {
      opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  } on Object {
    opened = false;
  }
  if (!opened && context.mounted) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Tautan sumber belum dapat dibuka.'.localized(context)),
      ),
    );
  }
}

class MangaEntryCard extends StatelessWidget {
  const MangaEntryCard({super.key});
  @override
  Widget build(BuildContext context) => ExploreFeatureCard(
    key: const Key('open-manga'),
    title: 'Manga Jepang'.localized(context),
    eyebrow: 'PREMIUM · LIFETIME ACCESS',
    description: 'Nikmati cerita dan latih membaca Jepang lewat panel aslinya'
        .localized(context),
    asset: 'assets/images/explore/manga-reading.webp',
    color: const Color(0xFF30265D),
    icon: Icons.auto_stories_rounded,
    premium: true,
    onTap: () => context.push('/manga'),
  );
}

class MangaArtwork extends StatelessWidget {
  const MangaArtwork({this.size = 92, super.key});
  final double size;
  @override
  Widget build(BuildContext context) => SizedBox.square(
    dimension: size,
    child: Stack(
      children: [
        Positioned.fill(
          child: Transform.rotate(
            angle: -.09,
            child: Container(
              decoration: BoxDecoration(
                color: const Color(0xFFFFD96A),
                borderRadius: BorderRadius.circular(14),
              ),
            ),
          ),
        ),
        Positioned.fill(
          left: 6,
          top: 5,
          child: Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: _ink, width: 2),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.forum_rounded, size: size * .38, color: _ink),
                const Text(
                  'まんが',
                  style: TextStyle(color: _ink, fontWeight: FontWeight.w900),
                ),
              ],
            ),
          ),
        ),
      ],
    ),
  );
}

class MangaCatalogScreen extends ConsumerStatefulWidget {
  const MangaCatalogScreen({super.key});
  @override
  ConsumerState<MangaCatalogScreen> createState() => _MangaCatalogScreenState();
}

class _MangaCatalogScreenState extends ConsumerState<MangaCatalogScreen> {
  final _search = TextEditingController();
  var _page = 1;
  var _query = '';
  var _genre = '';
  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final provider = mangaCatalogProvider((
      page: _page,
      q: _query,
      genre: _genre,
    ));
    final catalog = ref.watch(provider);
    final colors = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(title: Text('Manga Jepang'.localized(context))),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(provider.future),
        child: ListView(
          padding: const EdgeInsets.all(20),
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            Container(
              padding: const EdgeInsets.all(22),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [_ink, Color(0xFF565BB4)],
                ),
                borderRadius: BorderRadius.circular(26),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          '日本語・ORIGINAL',
                          style: TextStyle(
                            color: Color(0xFFFFD96A),
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            letterSpacing: 1.3,
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          'Cerita seru.\nBahasa aslinya.'.localized(context),
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 25,
                            fontWeight: FontWeight.w900,
                            height: 1.2,
                          ),
                        ),
                        const SizedBox(height: 10),
                        Text(
                          'Latih membaca Jepang lewat panel demi panel.'
                              .localized(context),
                          style: const TextStyle(
                            color: Colors.white,
                            height: 1.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  const MangaArtwork(),
                ],
              ),
            ),
            const SizedBox(height: 14),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton(
                style: TextButton.styleFrom(
                  textStyle: const TextStyle(fontSize: 11),
                  visualDensity: VisualDensity.compact,
                ),
                onPressed: () => context.push('/attributions'),
                child: Text('Sumber & atribusi ↗'.localized(context)),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _search,
              key: const Key('manga-search'),
              textInputAction: TextInputAction.search,
              onSubmitted: (v) => setState(() {
                _query = v.trim();
                _page = 1;
              }),
              decoration: InputDecoration(
                hintText: 'Cari judul manga…'.localized(context),
                prefixIcon: const Icon(Icons.search_rounded),
                suffixIcon: IconButton(
                  tooltip: 'Cari'.localized(context),
                  onPressed: () => setState(() {
                    _query = _search.text.trim();
                    _page = 1;
                  }),
                  icon: const Icon(Icons.arrow_forward_rounded),
                ),
              ),
            ),
            const SizedBox(height: 12),
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: [
                  for (final genre in [
                    '',
                    'Sports',
                    'Slice of Life',
                    'Comedy',
                    'Fantasy',
                    'School Life',
                  ])
                    Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: ChoiceChip(
                        key: ValueKey('manga-genre-$genre'),
                        label: Text(
                          genre.isEmpty ? 'Semua'.localized(context) : genre,
                        ),
                        selected: _genre == genre,
                        onSelected: (_) => setState(() {
                          _genre = genre;
                          _page = 1;
                        }),
                      ),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 18),
            catalog.when(
              loading: () => const Padding(
                padding: EdgeInsets.all(36),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (error, _) => _MangaError(
                error: error,
                retry: () => ref.invalidate(provider),
              ),
              data: (data) => data.items.isEmpty
                  ? Padding(
                      padding: const EdgeInsets.symmetric(vertical: 36),
                      child: Column(
                        children: [
                          Icon(
                            Icons.auto_stories_outlined,
                            size: 52,
                            color: colors.primary,
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _query.isNotEmpty || _genre.isNotEmpty
                                ? 'Manga belum ditemukan'.localized(context)
                                : 'Koleksi sedang disiapkan'.localized(context),
                            style: const TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Judul terkurasi akan tersedia di sini. Coba muat ulang beberapa saat lagi.'
                                .localized(context),
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    )
                  : Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${data.total} judul untuk dijelajahi'.localized(
                            context,
                          ),
                          style: const TextStyle(fontWeight: FontWeight.w800),
                        ),
                        const SizedBox(height: 14),
                        LayoutBuilder(
                          builder: (context, box) {
                            final columns = box.maxWidth >= 700 ? 4 : 2;
                            final width =
                                (box.maxWidth - 14 * (columns - 1)) / columns;
                            return Wrap(
                              spacing: 14,
                              runSpacing: 20,
                              children: [
                                for (final series in data.items)
                                  SizedBox(
                                    width: width,
                                    child: _MangaCard(series: series),
                                  ),
                              ],
                            );
                          },
                        ),
                        const SizedBox(height: 20),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            IconButton(
                              key: const Key('manga-previous-page'),
                              tooltip: 'Halaman sebelumnya'.localized(context),
                              onPressed: _page > 1
                                  ? () => setState(() => _page--)
                                  : null,
                              icon: const Icon(Icons.chevron_left),
                            ),
                            Text('Halaman $_page'.localized(context)),
                            IconButton(
                              key: const Key('manga-next-page'),
                              tooltip: 'Halaman berikutnya'.localized(context),
                              onPressed: _page * 20 < data.total
                                  ? () => setState(() => _page++)
                                  : null,
                              icon: const Icon(Icons.chevron_right),
                            ),
                          ],
                        ),
                      ],
                    ),
            ),
            const SizedBox(height: 16),
            Text(
              'Pilihan tanpa konten dewasa/ecchi. Katalog diperiksa mingguan.'
                  .localized(context),
              style: TextStyle(
                color: colors.onSurfaceVariant,
                fontSize: 12,
                height: 1.5,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MangaCard extends StatelessWidget {
  const _MangaCard({required this.series});
  final MangaSeries series;
  @override
  Widget build(BuildContext context) => InkWell(
    key: ValueKey('manga-${series.id}'),
    borderRadius: BorderRadius.circular(16),
    onTap: () => context.push('/manga/${Uri.encodeComponent(series.id)}'),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Hero(
          tag: 'manga-cover-${series.id}',
          child: ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: AspectRatio(
              aspectRatio: .7,
              child: RemoteAssetImage(
                asset: '',
                url: series.coverUrl,
                fit: BoxFit.cover,
              ),
            ),
          ),
        ),
        const SizedBox(height: 10),
        Text(
          series.title,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 4),
        Text(
          '${series.chapterCount} chapter · 日本語'.localized(context),
          style: TextStyle(
            fontSize: 12,
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        if (series.genres.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 5),
            child: Text(
              series.genres.take(2).join(' · '),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 11),
            ),
          ),
      ],
    ),
  );
}

class MangaDetailScreen extends ConsumerStatefulWidget {
  const MangaDetailScreen({required this.id, super.key});
  final String id;
  @override
  ConsumerState<MangaDetailScreen> createState() => _MangaDetailScreenState();
}

class _MangaDetailScreenState extends ConsumerState<MangaDetailScreen> {
  var _search = '';
  var _newest = false;
  @override
  Widget build(BuildContext context) {
    final provider = mangaDetailProvider(widget.id);
    return Scaffold(
      appBar: AppBar(title: Text('Detail manga'.localized(context))),
      body: ref
          .watch(provider)
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (error, _) => _MangaError(
              error: error,
              retry: () => ref.invalidate(provider),
            ),
            data: (data) {
              final s = data.series;
              final chapters = data.chapters
                  .where(
                    (c) =>
                        c.title.toLowerCase().contains(_search.toLowerCase()),
                  )
                  .toList();
              if (_newest) {
                final reversed = chapters.reversed.toList();
                chapters
                  ..clear()
                  ..addAll(reversed);
              }
              return ListView(
                padding: const EdgeInsets.all(20),
                children: [
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        width: 116,
                        child: Hero(
                          tag: 'manga-cover-${s.id}',
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(15),
                            child: AspectRatio(
                              aspectRatio: .7,
                              child: RemoteAssetImage(
                                asset: '',
                                url: s.coverUrl,
                                fit: BoxFit.cover,
                              ),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 18),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              s.title,
                              style: const TextStyle(
                                fontSize: 23,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                            if (s.japaneseTitle.isNotEmpty)
                              Padding(
                                padding: const EdgeInsets.only(top: 7),
                                child: Text(s.japaneseTitle),
                              ),
                            if (s.author.isNotEmpty)
                              Padding(
                                padding: const EdgeInsets.only(top: 9),
                                child: Text(s.author),
                              ),
                            const SizedBox(height: 12),
                            Text(
                              '${s.chapterCount} chapter · Jepang asli'
                                  .localized(context),
                              style: const TextStyle(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Wrap(
                    spacing: 6,
                    runSpacing: 4,
                    children: [
                      for (final genre in s.genres) Chip(label: Text(genre)),
                    ],
                  ),
                  _SourceLink(url: s.sourceUrl),
                  const SizedBox(height: 12),
                  Text(
                    'Tentang cerita'.localized(context),
                    style: const TextStyle(
                      fontWeight: FontWeight.w800,
                      fontSize: 19,
                    ),
                  ),
                  const SizedBox(height: 8),
                  SelectableText(
                    s.summary,
                    style: const TextStyle(height: 1.7),
                  ),
                  const SizedBox(height: 20),
                  if (data.chapters.isNotEmpty)
                    FilledButton.icon(
                      onPressed: () => context.push(
                        '/manga/chapters/${data.chapters.first.id}',
                      ),
                      icon: const Icon(Icons.menu_book_rounded),
                      label: Text('Mulai membaca'.localized(context)),
                    ),
                  const SizedBox(height: 22),
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          'Daftar chapter'.localized(context),
                          style: const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                      IconButton(
                        tooltip: _newest
                            ? 'Urutan terlama dahulu'.localized(context)
                            : 'Urutan terbaru dahulu'.localized(context),
                        onPressed: () => setState(() => _newest = !_newest),
                        icon: const Icon(Icons.sort_rounded),
                      ),
                    ],
                  ),
                  TextField(
                    onChanged: (v) => setState(() => _search = v),
                    decoration: InputDecoration(
                      hintText: 'Cari nomor chapter'.localized(context),
                      prefixIcon: const Icon(Icons.search),
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (chapters.isEmpty)
                    Padding(
                      padding: const EdgeInsets.all(24),
                      child: Text(
                        'Chapter tidak ditemukan.'.localized(context),
                      ),
                    ),
                  for (final ch in chapters)
                    ListTile(
                      key: ValueKey('manga-chapter-${ch.id}'),
                      contentPadding: EdgeInsets.zero,
                      leading: CircleAvatar(
                        child: Icon(
                          ch.status == 'ready'
                              ? Icons.check_rounded
                              : Icons.menu_book_outlined,
                        ),
                      ),
                      title: Text(
                        ch.title,
                        style: const TextStyle(fontWeight: FontWeight.w700),
                      ),
                      subtitle: Text(
                        ch.status == 'ready'
                            ? 'Siap dibaca'.localized(context)
                            : ch.pending
                            ? 'Sedang disiapkan'.localized(context)
                            : ch.status == 'failed'
                            ? 'Belum tersedia'.localized(context)
                            : 'Siapkan untuk dibaca'.localized(context),
                      ),
                      trailing: const Icon(Icons.chevron_right),
                      onTap: () => context.push('/manga/chapters/${ch.id}'),
                    ),
                ],
              );
            },
          ),
    );
  }
}

class MangaReaderScreen extends ConsumerStatefulWidget {
  const MangaReaderScreen({required this.id, super.key});
  final String id;
  @override
  ConsumerState<MangaReaderScreen> createState() => _MangaReaderScreenState();
}

class _MangaReaderScreenState extends ConsumerState<MangaReaderScreen>
    with WidgetsBindingObserver {
  MangaChapter? _chapter;
  Timer? _poll;
  var _polls = 0;
  var _version = 0;
  var _busy = false;
  var _foreground = true;
  var _continuous = false;
  var _page = 0;
  String? _error;
  final _pages = PageController();
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    unawaited(_load());
  }

  @override
  void dispose() {
    _version++;
    _poll?.cancel();
    _pages.dispose();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _foreground = state == AppLifecycleState.resumed;
    _poll?.cancel();
    if (state == AppLifecycleState.resumed && (_chapter?.pending ?? false)) {
      unawaited(_load());
    }
  }

  Future<void> _load({bool prepare = false}) async {
    if (_busy) return;
    final version = ++_version;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final repo = ref.read(mangaRepositoryProvider);
      final ch = prepare
          ? await repo.prepare(widget.id)
          : await repo.chapter(widget.id);
      if (!mounted || version != _version) return;
      setState(() {
        _chapter = ch;
        _busy = false;
      });
      _poll?.cancel();
      if (ch.pending && _polls < 20 && _foreground) {
        _polls++;
        _poll = Timer(const Duration(seconds: 15), () => unawaited(_load()));
      }
    } on Object catch (error) {
      if (mounted &&
          version == _version &&
          error is DioException &&
          [401, 402].contains(error.response?.statusCode)) {
        _poll?.cancel();
        setState(() {
          _chapter = null;
          _busy = false;
        });
        context.go(
          '/premium?from=${Uri.encodeQueryComponent('/manga/chapters/${widget.id}')}',
        );
        return;
      }
      if (mounted && version == _version) {
        setState(() {
          _error = 'Chapter belum dapat dimuat. Silakan coba lagi.';
          _busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final ch = _chapter;
    return Scaffold(
      backgroundColor: const Color(0xFF15151D),
      appBar: AppBar(
        backgroundColor: const Color(0xFF15151D),
        foregroundColor: Colors.white,
        title: Text(ch?.title ?? 'Pembaca manga'.localized(context)),
        actions: [
          if (ch?.ready ?? false)
            IconButton(
              tooltip: _continuous
                  ? 'Baca per halaman'.localized(context)
                  : 'Baca dengan scroll'.localized(context),
              onPressed: () => setState(() => _continuous = !_continuous),
              icon: Icon(
                _continuous
                    ? Icons.view_carousel_rounded
                    : Icons.view_stream_rounded,
              ),
            ),
          if (ch != null)
            IconButton(
              tooltip: 'Sumber: Rawkuma'.localized(context),
              onPressed: () => _openSource(context, ch.sourceUrl),
              icon: const Icon(Icons.open_in_new_rounded),
            ),
        ],
      ),
      body: ch?.ready ?? false
          ? Column(
              children: [
                Expanded(
                  child: _continuous
                      ? ListView.builder(
                          itemCount: ch!.pages.length,
                          itemBuilder: (context, i) => Column(
                            children: [
                              _PageImage(page: ch.pages[i]),
                              Padding(
                                padding: const EdgeInsets.all(8),
                                child: Text(
                                  '${i + 1} / ${ch.pages.length}',
                                  style: const TextStyle(
                                    color: Colors.white70,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        )
                      : PageView.builder(
                          controller: _pages,
                          reverse: true,
                          itemCount: ch!.pages.length,
                          onPageChanged: (i) => setState(() => _page = i),
                          itemBuilder: (context, i) => InteractiveViewer(
                            maxScale: 5,
                            child: Center(
                              child: _PageImage(
                                page: ch.pages[i],
                                contain: true,
                              ),
                            ),
                          ),
                        ),
                ),
                SafeArea(
                  top: false,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 5,
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            'Baca dari kanan ke kiri'.localized(context),
                            style: const TextStyle(
                              color: Colors.white70,
                              fontSize: 11,
                            ),
                          ),
                        ),
                        if (!_continuous) ...[
                          IconButton(
                            tooltip: 'Halaman sebelumnya'.localized(context),
                            color: Colors.white70,
                            disabledColor: Colors.white24,
                            onPressed: _page > 0
                                ? () => _pages.previousPage(
                                    duration: const Duration(milliseconds: 200),
                                    curve: Curves.easeOut,
                                  )
                                : null,
                            icon: const Icon(Icons.chevron_right),
                          ),
                          Text(
                            '${_page + 1} / ${ch.pages.length}',
                            style: const TextStyle(color: Colors.white),
                          ),
                          IconButton(
                            tooltip: 'Halaman berikutnya'.localized(context),
                            color: Colors.white70,
                            disabledColor: Colors.white24,
                            onPressed: _page + 1 < ch.pages.length
                                ? () => _pages.nextPage(
                                    duration: const Duration(milliseconds: 200),
                                    curve: Curves.easeOut,
                                  )
                                : null,
                            icon: const Icon(Icons.chevron_left),
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ],
            )
          : Center(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(28),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (_busy)
                      const CircularProgressIndicator(color: Colors.white)
                    else
                      Icon(
                        ch?.pending ?? false
                            ? Icons.hourglass_top_rounded
                            : Icons.auto_stories_outlined,
                        color: const Color(0xFFFFD96A),
                        size: 64,
                      ),
                    const SizedBox(height: 22),
                    Text(
                      ch?.pending ?? false
                          ? 'Chapter sedang disiapkan'.localized(context)
                          : 'Siap menjelajahi cerita?'.localized(context),
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 23,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      _error?.localized(context) ??
                          (ch?.pending ?? false
                              ? 'Halaman akan terbuka setelah semua gambar siap. Kamu juga bisa kembali dan membukanya lagi nanti.'
                                    .localized(context)
                              : ch?.status == 'failed'
                              ? 'Beberapa halaman belum siap. Cek kembali setelah waktu berikutnya.'
                                    .localized(context)
                              : 'Siapkan chapter ini agar tersedia untuk dibaca.'),
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        color: Colors.white70,
                        height: 1.7,
                      ),
                    ),
                    if (ch?.status == 'failed' && ch?.retryAt != null)
                      Padding(
                        padding: const EdgeInsets.only(top: 10),
                        child: Text(
                          'Coba kembali: ${_retryDate(context, ch!.retryAt!)}'
                              .localized(context),
                          style: const TextStyle(color: Colors.white70),
                        ),
                      ),
                    const SizedBox(height: 22),
                    if (ch?.canPrepare ?? false)
                      FilledButton(
                        onPressed: _busy
                            ? null
                            : () {
                                _polls = 0;
                                unawaited(_load(prepare: true));
                              },
                        child: Text('Siapkan chapter'.localized(context)),
                      ),
                    if (!_busy)
                      TextButton(
                        onPressed: () {
                          _polls = 0;
                          unawaited(_load());
                        },
                        child: Text(
                          'Muat ulang'.localized(context),
                          style: const TextStyle(color: Colors.white),
                        ),
                      ),
                    if (ch != null)
                      TextButton(
                        onPressed: () => _openSource(context, ch.sourceUrl),
                        child: Text(
                          'Sumber: Rawkuma ↗'.localized(context),
                          style: const TextStyle(
                            fontSize: 11,
                            color: Colors.white70,
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ),
    );
  }
}

class _PageImage extends StatefulWidget {
  const _PageImage({required this.page, this.contain = false});
  final MangaPage page;
  final bool contain;
  @override
  State<_PageImage> createState() => _PageImageState();
}

class _PageImageState extends State<_PageImage> {
  var _attempt = 0;
  @override
  Widget build(BuildContext context) {
    final page = widget.page;
    final image = Image.network(
      page.url,
      cacheWidth: math.min(page.width, 2048),
      key: ValueKey('${page.url}-$_attempt'),
      fit: widget.contain ? BoxFit.contain : BoxFit.fitWidth,
      width: double.infinity,
      semanticLabel: 'Halaman manga bahasa Jepang'.localized(context),
      loadingBuilder: (_, child, progress) => progress == null
          ? child
          : const SizedBox(
              height: 220,
              child: Center(
                child: CircularProgressIndicator(color: Colors.white),
              ),
            ),
      errorBuilder: (_, __, ___) => SizedBox(
        height: 220,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'Gambar belum dapat dimuat'.localized(context),
              style: const TextStyle(color: Colors.white70),
            ),
            TextButton(
              onPressed: () async {
                await NetworkImage(page.url).evict();
                if (mounted) setState(() => _attempt++);
              },
              child: Text('Coba lagi'.localized(context)),
            ),
          ],
        ),
      ),
    );
    if (widget.contain) return image;
    return AspectRatio(aspectRatio: page.width / page.height, child: image);
  }
}

class _SourceLink extends StatelessWidget {
  const _SourceLink({required this.url});
  final String url;
  @override
  Widget build(BuildContext context) => Align(
    alignment: Alignment.centerLeft,
    child: TextButton.icon(
      onPressed: () => _openSource(context, url),
      style: TextButton.styleFrom(
        textStyle: const TextStyle(fontSize: 11),
        visualDensity: VisualDensity.compact,
      ),
      icon: const Icon(Icons.open_in_new_rounded, size: 12),
      label: Text('Sumber: Rawkuma'.localized(context)),
    ),
  );
}

class _MangaError extends StatelessWidget {
  const _MangaError({required this.retry, required this.error});
  final VoidCallback retry;
  final Object error;
  @override
  Widget build(BuildContext context) {
    if (error is DioException &&
        [401, 402].contains((error as DioException).response?.statusCode)) {
      return Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            const Icon(
              Icons.workspace_premium_rounded,
              size: 44,
              color: Color(0xFF9A6500),
            ),
            const SizedBox(height: 12),
            Text(
              'Manga Jepang khusus pengguna Lifetime Access.'.localized(
                context,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: () => context.push('/premium?from=%2Fmanga'),
              child: Text('Lihat Lifetime Access'.localized(context)),
            ),
          ],
        ),
      );
    }
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Manga belum dapat dimuat. Periksa koneksi lalu coba kembali.'
                .localized(context),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 12),
          OutlinedButton(
            onPressed: retry,
            child: Text('Coba lagi'.localized(context)),
          ),
        ],
      ),
    );
  }
}
