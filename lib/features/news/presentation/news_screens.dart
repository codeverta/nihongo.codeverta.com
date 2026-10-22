import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/dictionary/data/dictionary_repository.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_details.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/news/presentation/furigana_text.dart';

const newsFuriganaPreferenceKey = legacyNewsFuriganaPreferenceKey;

String newsDate(DateTime date, {bool time = false}) {
  final d = date.toUtc().add(const Duration(hours: 7));
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'Mei',
    'Jun',
    'Jul',
    'Agu',
    'Sep',
    'Okt',
    'Nov',
    'Des',
  ];
  final day = '${d.day} ${months[d.month - 1]} ${d.year}';
  return time
      ? '$day, ${d.hour.toString().padLeft(2, '0')}.${d.minute.toString().padLeft(2, '0')} WIB'
      : day;
}

class NewsPreview extends ConsumerWidget {
  const NewsPreview({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Row(
        children: [
          Expanded(
            child: Text(
              'Berita Jepang',
              style: Theme.of(
                context,
              ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
          ),
          TextButton(
            onPressed: () => context.push('/news'),
            child: const Text('Lihat semua'),
          ),
        ],
      ),
      const Text('Baca Jepang, temukan kata baru.'),
      const SizedBox(height: 12),
      ref
          .watch(newsPreviewProvider)
          .when(
            loading: () => const LinearProgressIndicator(),
            error: (e, s) => _NewsError(
              onRetry: () => ref.invalidate(newsPreviewProvider),
              compact: true,
            ),
            data: (page) => page.items.isEmpty
                ? const _NewsEmpty()
                : Column(
                    children: [
                      for (final article in page.items)
                        NewsArticleCard(article: article),
                    ],
                  ),
          ),
    ],
  );
}

const _newsPageSize = 20;

const _newsIndigo = Color(0xFF1D2088);
const _newsCrimson = Color(0xFFE60012);

Color _kindColor(BuildContext context, String kind) {
  final dark = Theme.of(context).brightness == Brightness.dark;
  if (kind == 'easy') {
    return dark ? const Color(0xFF6FD6C8) : const Color(0xFF146B63);
  }
  return dark ? const Color(0xFFA9ACFF) : const Color(0xFF292A87);
}

/// Label tanggal relatif ("Hari ini", "Kemarin", "3 hari lalu").
String newsRelativeDate(DateTime date) {
  final now = DateTime.now().toUtc().add(const Duration(hours: 7));
  final today = DateTime.utc(now.year, now.month, now.day);
  final d = date.toUtc().add(const Duration(hours: 7));
  final day = DateTime.utc(d.year, d.month, d.day);
  final diff = today.difference(day).inDays;
  if (diff <= 0) return 'Hari ini';
  if (diff == 1) return 'Kemarin';
  if (diff < 7) return '$diff hari lalu';
  return newsDate(date);
}

class _KindBadge extends StatelessWidget {
  const _KindBadge({required this.article});
  final NewsArticle article;
  @override
  Widget build(BuildContext context) {
    final color = _kindColor(context, article.kind);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: .12),
        borderRadius: BorderRadius.circular(99),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            article.kind == 'easy' ? Icons.spa_rounded : Icons.public_rounded,
            size: 12,
            color: color,
          ),
          const SizedBox(width: 4),
          Text(
            article.kindLabel,
            style: TextStyle(
              color: color,
              fontSize: 11,
              fontWeight: FontWeight.w700,
              letterSpacing: .2,
            ),
          ),
        ],
      ),
    );
  }
}

class _NewsMeta extends StatelessWidget {
  const _NewsMeta({required this.article});
  final NewsArticle article;
  @override
  Widget build(BuildContext context) {
    final style = Theme.of(context).textTheme.bodySmall;
    final muted = style?.color?.withValues(alpha: .8);
    return Row(
      children: [
        Icon(Icons.newspaper_rounded, size: 13, color: muted),
        const SizedBox(width: 4),
        Flexible(
          child: Text(
            article.source,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: style,
          ),
        ),
        Text('  ·  ', style: style),
        Icon(Icons.schedule_rounded, size: 13, color: muted),
        const SizedBox(width: 4),
        Text(newsRelativeDate(article.sourceDate), style: style),
      ],
    );
  }
}

class _NewsPlaceholder extends StatelessWidget {
  const _NewsPlaceholder({required this.kind, this.size = 30});
  final String kind;
  final double size;
  @override
  Widget build(BuildContext context) {
    final color = _kindColor(context, kind);
    return DecoratedBox(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [color.withValues(alpha: .22), color.withValues(alpha: .06)],
        ),
      ),
      child: Center(
        child: Text(
          kind == 'easy' ? '易' : '報',
          style: TextStyle(
            fontSize: size,
            fontWeight: FontWeight.w800,
            color: color.withValues(alpha: .55),
          ),
        ),
      ),
    );
  }
}

class _NewsImage extends ConsumerWidget {
  const _NewsImage({
    required this.article,
    this.width,
    this.height,
    this.cacheWidth = 300,
    this.placeholderSize = 30,
  });
  final NewsArticle article;
  final double? width;
  final double? height;
  final int cacheWidth;
  final double placeholderSize;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final placeholder = _NewsPlaceholder(
      kind: article.kind,
      size: placeholderSize,
    );
    final image = article.imageUrl;
    return SizedBox(
      width: width,
      height: height,
      child: image.isEmpty
          ? placeholder
          : Image.network(
              Uri.parse(
                ref.watch(apiBaseUrlProvider),
              ).resolve(image).toString(),
              fit: BoxFit.cover,
              cacheWidth: cacheWidth,
              loadingBuilder: (_, child, progress) =>
                  progress == null ? child : placeholder,
              errorBuilder: (_, e, s) => placeholder,
            ),
    );
  }
}

void _openArticle(BuildContext context, NewsArticle article) =>
    context.push('/news/${Uri.encodeComponent(article.id)}');

class NewsArticleCard extends StatelessWidget {
  const NewsArticleCard({required this.article, super.key});
  final NewsArticle article;
  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        key: ValueKey('news-${article.id}'),
        onTap: () => _openArticle(context, article),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: _NewsImage(
                  article: article,
                  width: 88,
                  height: 100,
                  cacheWidth: 264,
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _KindBadge(article: article),
                    const SizedBox(height: 7),
                    Text(
                      article.title,
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w700,
                        height: 1.45,
                      ),
                    ),
                    const SizedBox(height: 8),
                    _NewsMeta(article: article),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _FeaturedNewsCard extends StatelessWidget {
  const _FeaturedNewsCard({required this.article});
  final NewsArticle article;
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 18),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        key: ValueKey('news-${article.id}'),
        onTap: () => _openArticle(context, article),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              height: 180,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  _NewsImage(
                    article: article,
                    cacheWidth: 960,
                    placeholderSize: 72,
                  ),
                  if (article.imageUrl.isNotEmpty)
                    const DecoratedBox(
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          begin: Alignment.topCenter,
                          end: Alignment.bottomCenter,
                          colors: [Colors.transparent, Color(0x66000000)],
                        ),
                      ),
                    ),
                  Positioned(
                    left: 12,
                    top: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: _newsCrimson,
                        borderRadius: BorderRadius.circular(99),
                      ),
                      child: const Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.bolt_rounded,
                            size: 14,
                            color: Colors.white,
                          ),
                          SizedBox(width: 3),
                          Text(
                            'Terbaru',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _KindBadge(article: article),
                  const SizedBox(height: 8),
                  Text(
                    article.title,
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                      height: 1.45,
                    ),
                  ),
                  if (article.summary.isNotEmpty) ...[
                    const SizedBox(height: 6),
                    Text(
                      article.summary,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        height: 1.5,
                        color: theme.textTheme.bodySmall?.color,
                      ),
                    ),
                  ],
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(child: _NewsMeta(article: article)),
                      Text(
                        'Baca',
                        style: TextStyle(
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      Icon(
                        Icons.arrow_forward_rounded,
                        size: 16,
                        color: theme.colorScheme.primary,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Animasi masuk ringan (fade + slide) bertahap per item.
class _FadeIn extends StatelessWidget {
  const _FadeIn({required this.index, required this.child});
  final int index;
  final Widget child;
  @override
  Widget build(BuildContext context) => TweenAnimationBuilder<double>(
    tween: Tween(begin: 0, end: 1),
    duration: Duration(milliseconds: 260 + 40 * index.clamp(0, 8)),
    curve: Curves.easeOutCubic,
    builder: (_, t, child) => Opacity(
      opacity: t,
      child: Transform.translate(offset: Offset(0, 16 * (1 - t)), child: child),
    ),
    child: child,
  );
}

class _NewsHero extends StatelessWidget {
  const _NewsHero({required this.page, required this.source});
  final NewsPage? page;
  final String source;
  @override
  Widget build(BuildContext context) {
    final sources = (page?.sources ?? const <Map<String, dynamic>>[])
        .where((s) => source.isEmpty || s['id'] == source)
        .toList();
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 4, 16, 4),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(22),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [_newsIndigo, Color(0xFF3A2C9E)],
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            right: -28,
            top: -28,
            child: Container(
              width: 130,
              height: 130,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _newsCrimson.withValues(alpha: .85),
              ),
            ),
          ),
          Positioned(
            right: 14,
            bottom: -18,
            child: Text(
              '新聞',
              style: TextStyle(
                fontSize: 76,
                fontWeight: FontWeight.w900,
                color: Colors.white.withValues(alpha: .08),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 18, 18, 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '日本のニュース',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 22,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Latih membaca dengan berita asli. Pilih kata untuk melihat artinya.',
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: .82),
                    height: 1.4,
                    fontSize: 13,
                  ),
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 6,
                  children: [
                    if (page != null)
                      _HeroChip(
                        icon: Icons.article_rounded,
                        label: '${page!.total} artikel',
                      ),
                    for (final s in sources)
                      _HeroChip(
                        icon: s['lastSuccessAt'] == null
                            ? Icons.hourglass_top_rounded
                            : Icons.sync_rounded,
                        dot: s['lastSuccessAt'] == null
                            ? const Color(0xFFFFC857)
                            : const Color(0xFF5BE49B),
                        label:
                            '${s['name']} · ${s['lastSuccessAt'] == null ? 'menunggu sinkronisasi' : newsDate(DateTime.parse(s['lastSuccessAt'] as String), time: true)}',
                      ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _HeroChip extends StatelessWidget {
  const _HeroChip({required this.icon, required this.label, this.dot});
  final IconData icon;
  final String label;
  final Color? dot;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: .14),
      borderRadius: BorderRadius.circular(99),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (dot != null) ...[
          Container(
            width: 7,
            height: 7,
            decoration: BoxDecoration(color: dot, shape: BoxShape.circle),
          ),
          const SizedBox(width: 5),
        ] else ...[
          Icon(icon, size: 13, color: Colors.white),
          const SizedBox(width: 4),
        ],
        Flexible(
          child: Text(
            label,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 11,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      ],
    ),
  );
}

class _FilterBarDelegate extends SliverPersistentHeaderDelegate {
  _FilterBarDelegate({required this.selected, required this.onSelected});
  final String selected;
  final ValueChanged<String> onSelected;

  static const _choices = [
    ('', 'Semua', Icons.auto_awesome_rounded, 'all'),
    ('global-voices', 'Berita dunia', Icons.public_rounded, 'regular'),
    ('osaka', 'Info Jepang · Easy', Icons.spa_rounded, 'easy'),
  ];

  @override
  double get minExtent => 60;
  @override
  double get maxExtent => 60;

  @override
  Widget build(BuildContext context, double shrinkOffset, bool overlaps) {
    final theme = Theme.of(context);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.scaffoldBackgroundColor,
        boxShadow: overlaps
            ? [
                BoxShadow(
                  color: Colors.black.withValues(alpha: .06),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ]
            : null,
      ),
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        children: [
          for (final (id, label, icon, kind) in _choices)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: _FilterPill(
                label: label,
                icon: icon,
                color: kind == 'all'
                    ? theme.colorScheme.primary
                    : _kindColor(context, kind),
                selected: selected == id,
                onTap: () => onSelected(id),
              ),
            ),
        ],
      ),
    );
  }

  @override
  bool shouldRebuild(_FilterBarDelegate old) => old.selected != selected;
}

class _FilterPill extends StatelessWidget {
  const _FilterPill({
    required this.label,
    required this.icon,
    required this.color,
    required this.selected,
    required this.onTap,
  });
  final String label;
  final IconData icon;
  final Color color;
  final bool selected;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fg = selected ? Colors.white : theme.colorScheme.onSurface;
    final bg = selected
        ? (theme.brightness == Brightness.dark && color != _newsCrimson
              ? color.withValues(alpha: .35)
              : color)
        : theme.colorScheme.surface;
    return Semantics(
      selected: selected,
      button: true,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(99),
          onTap: onTap,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            curve: Curves.easeOut,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: bg,
              borderRadius: BorderRadius.circular(99),
              border: Border.all(
                color: selected
                    ? Colors.transparent
                    : theme.dividerColor.withValues(alpha: .6),
              ),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(icon, size: 16, color: selected ? Colors.white : color),
                const SizedBox(width: 6),
                Text(
                  label,
                  style: TextStyle(
                    color: fg,
                    fontSize: 13,
                    fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SectionLabel extends StatelessWidget {
  const _SectionLabel(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 10, top: 2),
    child: Row(
      children: [
        Container(
          width: 4,
          height: 16,
          decoration: BoxDecoration(
            color: _newsCrimson,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 8),
        Text(
          text,
          style: Theme.of(
            context,
          ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
        ),
      ],
    ),
  );
}

class _NewsPager extends StatelessWidget {
  const _NewsPager({
    required this.page,
    required this.pages,
    required this.onChange,
  });
  final int page;
  final int pages;
  final ValueChanged<int> onChange;
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(top: 6),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
        child: Row(
          children: [
            IconButton.filledTonal(
              tooltip: 'Sebelumnya',
              onPressed: page > 1 ? () => onChange(page - 1) : null,
              icon: const Icon(Icons.chevron_left_rounded),
            ),
            Expanded(
              child: Text.rich(
                TextSpan(
                  children: [
                    const TextSpan(text: 'Halaman '),
                    TextSpan(
                      text: '$page',
                      style: TextStyle(
                        fontWeight: FontWeight.w800,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    TextSpan(text: ' dari $pages'),
                  ],
                ),
                textAlign: TextAlign.center,
              ),
            ),
            IconButton.filledTonal(
              tooltip: 'Berikutnya',
              onPressed: page < pages ? () => onChange(page + 1) : null,
              icon: const Icon(Icons.chevron_right_rounded),
            ),
          ],
        ),
      ),
    );
  }
}

class _NewsSkeleton extends StatefulWidget {
  const _NewsSkeleton();
  @override
  State<_NewsSkeleton> createState() => _NewsSkeletonState();
}

class _NewsSkeletonState extends State<_NewsSkeleton>
    with SingleTickerProviderStateMixin {
  late final _c = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 900),
  )..repeat(reverse: true);
  @override
  void dispose() {
    _c.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final base = Theme.of(context).colorScheme.onSurface;
    Widget bar(double w, double h) => Container(
      width: w,
      height: h,
      decoration: BoxDecoration(
        color: base.withValues(alpha: .08),
        borderRadius: BorderRadius.circular(8),
      ),
    );
    return FadeTransition(
      opacity: Tween(begin: .45, end: 1.0).animate(_c),
      child: Column(
        children: [
          Container(
            height: 180,
            margin: const EdgeInsets.only(bottom: 18),
            decoration: BoxDecoration(
              color: base.withValues(alpha: .08),
              borderRadius: BorderRadius.circular(16),
            ),
          ),
          for (var i = 0; i < 4; i++)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  bar(88, 100),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        bar(90, 14),
                        const SizedBox(height: 10),
                        bar(double.infinity, 14),
                        const SizedBox(height: 6),
                        bar(180, 14),
                        const SizedBox(height: 14),
                        bar(120, 10),
                      ],
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class NewsListScreen extends ConsumerStatefulWidget {
  const NewsListScreen({super.key});
  @override
  ConsumerState<NewsListScreen> createState() => _NewsListScreenState();
}

class _NewsListScreenState extends ConsumerState<NewsListScreen> {
  int _page = 1;
  String _source = '';
  final _scroll = ScrollController();
  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  void _changePage(int page) {
    setState(() => _page = page);
    if (_scroll.hasClients) _scroll.jumpTo(0);
  }

  @override
  Widget build(BuildContext context) {
    final filter = (page: _page, source: _source);
    final provider = newsPageProvider(filter);
    final state = ref.watch(provider);
    final page = state.valueOrNull;

    Widget body;
    if (state.isLoading && page == null) {
      body = const SliverToBoxAdapter(child: _NewsSkeleton());
    } else if (state.hasError && page == null) {
      body = SliverFillRemaining(
        hasScrollBody: false,
        child: _NewsError(onRetry: () => ref.invalidate(provider)),
      );
    } else {
      final data = page!;
      final items = data.items;
      final featured = _page == 1 && items.isNotEmpty ? items.first : null;
      final rest = featured == null ? items : items.skip(1).toList();
      final pages = (data.total / _newsPageSize).ceil().clamp(1, 9999);
      body = SliverList.list(
        children: [
          if (items.isEmpty) const _NewsEmpty(),
          if (featured != null)
            _FadeIn(index: 0, child: _FeaturedNewsCard(article: featured)),
          if (rest.isNotEmpty)
            _SectionLabel(
              featured == null ? 'Daftar berita' : 'Berita lainnya',
            ),
          for (final (i, article) in rest.indexed)
            _FadeIn(
              index: i + 1,
              child: NewsArticleCard(article: article),
            ),
          if (pages > 1)
            _NewsPager(page: _page, pages: pages, onChange: _changePage),
          const SizedBox(height: 14),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(
                Icons.info_outline_rounded,
                size: 14,
                color: Theme.of(context).textTheme.bodySmall?.color,
              ),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  'Tanggal mengikuti terbitan sumber. Berita reguler memakai teks Jepang asli; info Easy memakai bahasa yang lebih sederhana.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ),
            ],
          ),
        ],
      );
    }

    return Scaffold(
      body: RefreshIndicator(
        edgeOffset: kToolbarHeight + MediaQuery.paddingOf(context).top,
        onRefresh: () async {
          ref
            ..invalidate(newsPreviewProvider)
            ..invalidate(provider);
          try {
            await ref.read(provider.future);
          } on Object catch (_) {
            // Error ditampilkan lewat state provider.
          }
        },
        child: CustomScrollView(
          controller: _scroll,
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            SliverAppBar(
              pinned: true,
              title: const Text('Berita Jepang'),
              leading: BackButton(
                onPressed: () =>
                    context.canPop() ? context.pop() : context.go('/home'),
              ),
            ),
            SliverToBoxAdapter(
              child: _NewsHero(page: page, source: _source),
            ),
            SliverPersistentHeader(
              pinned: true,
              delegate: _FilterBarDelegate(
                selected: _source,
                onSelected: (id) {
                  if (id == _source) return;
                  setState(() {
                    _source = id;
                    _page = 1;
                  });
                },
              ),
            ),
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 6, 16, 32),
              sliver: body,
            ),
          ],
        ),
      ),
    );
  }
}

class NewsReaderScreen extends ConsumerStatefulWidget {
  const NewsReaderScreen({required this.id, super.key});
  final String id;
  @override
  ConsumerState<NewsReaderScreen> createState() => _NewsReaderScreenState();
}

class _NewsReaderScreenState extends ConsumerState<NewsReaderScreen> {
  @override
  Widget build(BuildContext context) {
    final furigana = ref.watch(displayPreferencesProvider).furigana;
    final displayPreferencesController = ref.read(
      displayPreferencesProvider.notifier,
    );
    final provider = newsArticleProvider(widget.id);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Baca berita'),
        actions: [
          IconButton(
            key: const Key('news-furigana-toggle'),
            tooltip: furigana ? 'Matikan furigana' : 'Tampilkan furigana',
            isSelected: furigana,
            onPressed: () =>
                displayPreferencesController.setFurigana(enabled: !furigana),
            icon: const Icon(Icons.text_format_rounded),
            selectedIcon: const Icon(Icons.text_fields_rounded),
          ),
        ],
      ),
      body: ref
          .watch(provider)
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, s) =>
                _NewsError(onRetry: () => ref.invalidate(provider)),
            data: (article) => ListView(
              padding: const EdgeInsets.fromLTRB(22, 8, 22, 40),
              children: [
                Text(
                  article.kindLabel,
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.primary,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                FuriganaText(
                  text: article.title,
                  tokens: article.titleTokens,
                  showFurigana: furigana,
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.w800,
                  ),
                  onLookup: (tokens) => _lookup(context, tokens),
                ),
                const SizedBox(height: 12),
                Text(
                  '${article.source} · ${article.dateLabel} ${newsDate(article.sourceDate)}',
                ),
                Text(
                  'Disinkronkan ${newsDate(article.fetchedAt, time: true)}',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                if (article.imageUrl.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  ClipRRect(
                    borderRadius: BorderRadius.circular(14),
                    child: Image.network(
                      Uri.parse(
                        ref.watch(apiBaseUrlProvider),
                      ).resolve(article.imageUrl).toString(),
                      fit: BoxFit.contain,
                      cacheWidth: 1280,
                      errorBuilder: (_, e, s) => const SizedBox.shrink(),
                    ),
                  ),
                ],
                const SizedBox(height: 14),
                const Text('Pilih kata atau kalimat, lalu tekan “Cari arti”.'),
                if (furigana)
                  Text(
                    'Furigana dapat dibuat otomatis dan bisa keliru, terutama pada nama.',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                const Divider(height: 30),
                for (final paragraph in article.paragraphs)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 18),
                    child: FuriganaText(
                      text: paragraph.text,
                      tokens: paragraph.tokens,
                      showFurigana: furigana,
                      style: TextStyle(
                        fontSize: paragraph.kind == 'heading' ? 22 : 20,
                        fontWeight: paragraph.kind == 'heading'
                            ? FontWeight.bold
                            : FontWeight.normal,
                      ),
                      onLookup: (tokens) => _lookup(context, tokens),
                    ),
                  ),
                const Divider(height: 30),
                Text(
                  'Sumber & kredit',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                _NewsLink(
                  label: '${article.source} · Artikel asli',
                  url: article.url,
                ),
                for (final credit in article.credits)
                  _NewsLink(
                    label:
                        '${credit['role'] == 'translator'
                            ? 'Penerjemah'
                            : credit['role'] == 'publisher'
                            ? 'Penerbit'
                            : 'Penulis'}: ${credit['name']}',
                    url: credit['url'] as String,
                  ),
                _NewsLink(label: article.license, url: article.licenseUrl),
                Text(
                  article.changes,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                if (article.imageCredit.isNotEmpty)
                  Text(
                    'Gambar: ${article.imageCredit} · ${article.imageLicense}',
                  ),
                if (article.imageSourceUrl.isNotEmpty)
                  _NewsLink(
                    label: 'Sumber gambar',
                    url: article.imageSourceUrl,
                  ),
                if (article.imageLicenseUrl.isNotEmpty)
                  _NewsLink(
                    label: 'Lisensi gambar · ${article.imageLicense}',
                    url: article.imageLicenseUrl,
                  ),
                if (article.references.isNotEmpty)
                  ExpansionTile(
                    tilePadding: EdgeInsets.zero,
                    title: const Text('Referensi dalam artikel'),
                    children: [
                      for (final link in article.references)
                        _NewsLink(
                          label: link['name'] as String,
                          url: link['url'] as String,
                        ),
                    ],
                  ),
              ],
            ),
          ),
    );
  }
}

class _NewsLink extends StatelessWidget {
  const _NewsLink({required this.label, required this.url});
  final String label;
  final String url;
  @override
  Widget build(BuildContext context) => TextButton(
    onPressed: () {
      final uri = Uri.tryParse(url);
      if (uri?.scheme == 'https') openDictionaryUrl(context, uri!);
    },
    style: TextButton.styleFrom(
      alignment: Alignment.centerLeft,
      padding: const EdgeInsets.symmetric(vertical: 8),
    ),
    child: Text(label),
  );
}

void _lookup(BuildContext context, List<NewsToken> tokens) {
  showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    useSafeArea: true,
    builder: (_) => FractionallySizedBox(
      heightFactor: .85,
      child: _NewsDictionary(tokens: tokens),
    ),
  );
}

class _NewsDictionary extends ConsumerStatefulWidget {
  const _NewsDictionary({required this.tokens});
  final List<NewsToken> tokens;
  @override
  ConsumerState<_NewsDictionary> createState() => _NewsDictionaryState();
}

class _NewsDictionaryState extends ConsumerState<_NewsDictionary> {
  NewsToken? _selected;
  Future<DictionaryPage>? _results;
  @override
  void initState() {
    super.initState();
    if (widget.tokens.length == 1) _search(widget.tokens.single);
  }

  void _search(NewsToken token) {
    _selected = token;
    _results = token.base.runes.length > 100
        ? Future.error(StateError('Pilih kata yang lebih pendek.'))
        : ref
              .read(dictionaryRepositoryProvider)
              .search(token.base, mode: 'japanese');
  }

  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(22, 0, 22, 32),
    children: [
      Text('Kamus bacaan', style: Theme.of(context).textTheme.titleLarge),
      const SizedBox(height: 8),
      if (widget.tokens.length > 1) ...[
        const Text(
          'Pilih kata untuk melihat arti. Bentuk dasar dipakai untuk mencari kata yang berubah bentuk.',
        ),
        Wrap(
          spacing: 8,
          runSpacing: 4,
          children: [
            for (final token in widget.tokens)
              ChoiceChip(
                label: Text(
                  token.surface == token.base
                      ? token.surface
                      : '${token.surface} → ${token.base}',
                ),
                selected: identical(_selected, token),
                onSelected: (_) => setState(() => _search(token)),
              ),
          ],
        ),
      ],
      if (_results != null) ...[
        Text(
          _selected!.surface == _selected!.base
              ? _selected!.base
              : '${_selected!.surface} → ${_selected!.base}',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        FutureBuilder<DictionaryPage>(
          future: _results,
          builder: (context, snapshot) {
            if (snapshot.connectionState != ConnectionState.done) {
              return const Padding(
                padding: EdgeInsets.all(32),
                child: Center(child: CircularProgressIndicator()),
              );
            }
            if (snapshot.hasError) {
              return Column(
                children: [
                  const Text(
                    'Arti belum dapat dimuat. Periksa koneksi dan coba kembali.',
                  ),
                  TextButton(
                    onPressed: () => setState(() => _search(_selected!)),
                    child: const Text('Coba lagi'),
                  ),
                ],
              );
            }
            final entries = snapshot.data!.entries;
            if (entries.isEmpty) {
              return const Padding(
                padding: EdgeInsets.all(16),
                child: Text(
                  'Kata belum ditemukan di kamus. Nama orang atau tempat tidak selalu tersedia.',
                ),
              );
            }
            return Column(
              children: [
                for (final entry in entries)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    title: Text('${entry.word}  ${entry.reading}'),
                    subtitle: Text(entry.summary),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => showDictionaryWord(context, entry),
                  ),
              ],
            );
          },
        ),
      ],
    ],
  );
}

class _NewsError extends StatelessWidget {
  const _NewsError({required this.onRetry, this.compact = false});
  final VoidCallback onRetry;
  final bool compact;
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: EdgeInsets.all(compact ? 8 : 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (!compact) ...[
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withValues(alpha: .1),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.wifi_off_rounded,
                  size: 34,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(height: 14),
            ],
            const Text(
              'Berita belum dapat dimuat. Periksa koneksi dan coba kembali.',
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            FilledButton.tonalIcon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh_rounded, size: 18),
              label: const Text('Coba lagi'),
            ),
          ],
        ),
      ),
    );
  }
}

class _NewsEmpty extends StatelessWidget {
  const _NewsEmpty();
  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).textTheme.bodySmall?.color;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 28),
      child: Column(
        children: [
          Text(
            '空',
            style: TextStyle(
              fontSize: 44,
              fontWeight: FontWeight.w800,
              color: color?.withValues(alpha: .35),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Belum ada artikel tersedia. Berita akan muncul setelah sinkronisasi sumber berhasil.',
            textAlign: TextAlign.center,
            style: TextStyle(color: color),
          ),
        ],
      ),
    );
  }
}
