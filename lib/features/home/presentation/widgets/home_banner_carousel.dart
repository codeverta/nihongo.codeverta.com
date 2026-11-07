import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/shared/presentation/widgets/brand_artwork.dart';

class HomeBannerCarousel extends StatefulWidget {
  const HomeBannerCarousel({
    required this.banners,
    required this.onOpen,
    this.active = true,
    super.key,
  });

  final List<HomeBanner> banners;
  final ValueChanged<HomeBanner> onOpen;
  final bool active;

  @override
  State<HomeBannerCarousel> createState() => _HomeBannerCarouselState();
}

class _HomeBannerCarouselState extends State<HomeBannerCarousel>
    with WidgetsBindingObserver, SingleTickerProviderStateMixin {
  static const _interval = Duration(seconds: 5);

  /// Widget tests pump until no frame is scheduled; a looping 5 s fill would
  /// never settle there, so tests get a static indicator instead.
  static final bool _animateProgress = !WidgetsBinding.instance.runtimeType
      .toString()
      .contains('TestWidgetsFlutterBinding');

  late final AnimationController _progress = AnimationController(
    vsync: this,
    duration: _interval,
  );
  final _pages = PageController();
  Timer? _timer;
  int _index = 0;
  bool _dragging = false;
  bool _hovered = false;
  bool _resumed = true;
  bool _motionAllowed = true;
  bool _routeVisible = true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    final lifecycle = WidgetsBinding.instance.lifecycleState;
    _resumed = lifecycle == null || lifecycle == AppLifecycleState.resumed;
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final media = MediaQuery.maybeOf(context);
    _motionAllowed =
        !(media?.disableAnimations ?? false) &&
        !(media?.accessibleNavigation ?? false);
    _routeVisible = ModalRoute.of(context)?.isCurrent ?? true;
    _schedule();
  }

  @override
  void didUpdateWidget(HomeBannerCarousel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_index >= widget.banners.length) {
      _index = 0;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _pages.hasClients) _pages.jumpToPage(0);
      });
    }
    _schedule();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _resumed = state == AppLifecycleState.resumed;
    _schedule();
  }

  bool get _progressVisible =>
      _animateProgress && _motionAllowed && widget.banners.length > 1;

  void _schedule() {
    _timer?.cancel();
    _progress.stop();
    if (!widget.active ||
        !_resumed ||
        !_routeVisible ||
        !_motionAllowed ||
        _dragging ||
        _hovered ||
        widget.banners.length < 2) {
      return;
    }
    if (_progressVisible) unawaited(_progress.forward(from: 0));
    _timer = Timer(_interval, () {
      if (!mounted || !_pages.hasClients) return;
      unawaited(
        _pages.animateToPage(
          (_index + 1) % widget.banners.length,
          duration: const Duration(milliseconds: 400),
          curve: Curves.easeInOut,
        ),
      );
    });
  }

  void _pause() {
    _timer?.cancel();
    _progress.stop();
  }

  void _select(int index) {
    _pause();
    _schedule();
    if (!_pages.hasClients) return;
    if (_motionAllowed) {
      unawaited(
        _pages.animateToPage(
          index,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        ),
      );
    } else {
      _pages.jumpToPage(index);
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    _progress.dispose();
    WidgetsBinding.instance.removeObserver(this);
    _pages.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (widget.banners.isEmpty) return const SizedBox.shrink();
    return LayoutBuilder(
      builder: (context, constraints) {
        final textWidth = (constraints.maxWidth - 48).clamp(1, 460).toDouble();
        final scaler = MediaQuery.textScalerOf(context);
        double measure(
          String text,
          TextStyle style,
          int maxLines, [
          double? width,
        ]) {
          final painter =
              TextPainter(
                text: TextSpan(
                  text: text,
                  style: DefaultTextStyle.of(context).style.merge(style),
                ),
                textDirection: Directionality.of(context),
                textScaler: scaler,
                maxLines: maxLines,
                ellipsis: '…',
              )..layout(
                maxWidth: (width ?? textWidth)
                    .clamp(1, double.infinity)
                    .toDouble(),
              );
          final height = painter.height;
          painter.dispose();
          return height;
        }

        final height = widget.banners
            .map((banner) {
              final headingWidth = (banner.action == 'premium')
                  ? math.max<double>(1, textWidth - 102)
                  : textWidth;
              final headingHeight =
                  measure(
                    _eyebrow(banner.action).localized(context),
                    _eyebrowStyle,
                    1,
                    headingWidth,
                  ) +
                  12 +
                  measure(banner.title, _titleStyle, 3, headingWidth);
              return 40 +
                  math.max(
                    (banner.action == 'premium') ? 98 : 0,
                    headingHeight,
                  ) +
                  10 +
                  measure(banner.description, _descriptionStyle, 3) +
                  18 +
                  math.max(
                    48,
                    measure(
                          banner.buttonLabel,
                          _buttonStyle,
                          2,
                          textWidth - 60,
                        ) +
                        24,
                  );
            })
            .reduce(math.max);

        return MouseRegion(
          onEnter: (_) {
            _hovered = true;
            _schedule();
          },
          onExit: (_) {
            _hovered = false;
            _schedule();
          },
          child: Column(
            key: const Key('home-banner-carousel'),
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                height: height,
                child: NotificationListener<ScrollNotification>(
                  onNotification: (notification) {
                    if (notification is ScrollStartNotification &&
                        notification.dragDetails != null) {
                      _dragging = true;
                      _pause();
                    } else if (notification is ScrollEndNotification) {
                      _dragging = false;
                      _schedule();
                    }
                    return false;
                  },
                  child: PageView.builder(
                    key: const Key('home-banner-pages'),
                    controller: _pages,
                    itemCount: widget.banners.length,
                    allowImplicitScrolling: true,
                    onPageChanged: (index) {
                      setState(() => _index = index);
                      _schedule();
                    },
                    itemBuilder: (context, index) => _BannerCard(
                      banner: widget.banners[index],
                      textWidth: textWidth,
                      onTap: () {
                        _pause();
                        widget.onOpen(widget.banners[index]);
                      },
                    ),
                  ),
                ),
              ),
              if (widget.banners.length > 1)
                Wrap(
                  alignment: WrapAlignment.center,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    for (var i = 0; i < widget.banners.length; i++)
                      Semantics(
                        button: true,
                        selected: i == _index,
                        label:
                            'Banner ${i + 1} ${'dari'.localized(context)} ${widget.banners.length}: ${widget.banners[i].title}',
                        child: InkWell(
                          key: Key('home-banner-dot-$i'),
                          onTap: () => _select(i),
                          borderRadius: BorderRadius.circular(16),
                          child: AnimatedContainer(
                            duration: AppMotion.of(context, AppMotion.slow),
                            curve: AppMotion.standard,
                            width: i == _index ? 38 : 18,
                            height: 30,
                            alignment: Alignment.center,
                            child: _BannerDot(
                              active: i == _index,
                              progress: _progress,
                              showProgress: _progressVisible,
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
            ],
          ),
        );
      },
    );
  }
}

const _titleStyle = TextStyle(
  color: Colors.white,
  fontSize: 23,
  height: 1.17,
  fontWeight: FontWeight.w800,
);
const _descriptionStyle = TextStyle(
  color: Color(0xFFE7EAF6),
  fontSize: 14,
  height: 1.45,
);
const _eyebrowStyle = TextStyle(
  color: Color(0xFFE7EAF6),
  fontSize: 11,
  height: 1.4,
  letterSpacing: 1,
  fontWeight: FontWeight.w700,
);
const _buttonStyle = TextStyle(
  fontSize: 14,
  height: 1.3,
  fontWeight: FontWeight.w700,
);

String _eyebrow(String action) => switch (action) {
  'premium' => 'NIHONGO PREMIUM',
  'manga' => 'PREMIUM · LIFETIME ACCESS',
  'anki' => 'REVIEW SEDIKIT, INGAT LEBIH LAMA',
  'ssw' => 'KETERAMPILAN UNTUK JEPANG',
  'jobs' => 'PELUANG KARIER',
  _ => 'BELAJAR BAHASA JEPANG',
};

class _BannerCard extends StatelessWidget {
  const _BannerCard({
    required this.banner,
    required this.textWidth,
    required this.onTap,
  });
  final HomeBanner banner;
  final double textWidth;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final (color, asset) = switch (banner.action) {
      'premium' => (const Color(0xFF24285F), 'assets/images/exams/n5.webp'),
      'manga' => (
        const Color(0xFF30265D),
        'assets/images/explore/manga-reading.webp',
      ),
      'anki' => (
        const Color(0xFF145B60),
        'assets/images/explore/anki-flashcards.webp',
      ),
      'ssw' => (const Color(0xFF124E56), 'assets/images/ssw/ssw-careers.webp'),
      'jobs' => (
        const Color(0xFF443060),
        'assets/images/jobs/job-category-atlas.webp',
      ),
      _ => (const Color(0xFF272C75), 'assets/images/exams/n5.webp'),
    };
    final logicalWidth = MediaQuery.sizeOf(context).width;
    final width =
        ((logicalWidth > 0 ? logicalWidth : textWidth + 48) *
                MediaQuery.devicePixelRatioOf(context))
            .round()
            .clamp(64, 1600);
    Widget fallback() => Image.asset(
      asset,
      fit: BoxFit.cover,
      alignment: Alignment.centerRight,
      cacheWidth: width,
      excludeFromSemantics: true,
    );
    return ClipRRect(
      borderRadius: BorderRadius.circular(24),
      child: Material(
        color: color,
        child: InkWell(
          key: Key('home-banner-${banner.id}'),
          onTap: onTap,
          child: Stack(
            fit: StackFit.expand,
            children: [
              if (banner.action != 'premium' || banner.imageUrl.isNotEmpty)
                if (banner.imageUrl.isEmpty)
                  fallback()
                else
                  Image.network(
                    banner.imageUrl,
                    fit: BoxFit.cover,
                    alignment: Alignment.centerRight,
                    cacheWidth: width,
                    excludeFromSemantics: true,
                    errorBuilder: (context, error, stackTrace) => fallback(),
                    frameBuilder: (context, child, frame, loaded) =>
                        frame == null ? fallback() : child,
                  ),
              DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      color,
                      color.withValues(alpha: .94),
                      color.withValues(alpha: .52),
                    ],
                    stops: const [0, .38, 1],
                  ),
                ),
              ),
              if (banner.action == 'premium') ...[
                Positioned(
                  right: -90,
                  top: -110,
                  child: IgnorePointer(
                    child: Container(
                      width: 330,
                      height: 330,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: RadialGradient(
                          colors: [
                            const Color(0xFF7D79DC).withValues(alpha: .35),
                            const Color(0xFF7D79DC).withValues(alpha: 0),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
                Positioned(
                  right: -65,
                  bottom: -140,
                  child: IgnorePointer(
                    child: Container(
                      width: 280,
                      height: 280,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        border: Border.all(
                          color: Colors.white.withValues(alpha: .09),
                          width: 22,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 20,
                ),
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: SizedBox(
                    width: textWidth,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _eyebrow(banner.action).localized(context),
                                    style:
                                        (banner.action == 'premium' ||
                                            banner.action == 'manga')
                                        ? _eyebrowStyle.copyWith(
                                            color: const Color(0xFFFFD784),
                                          )
                                        : _eyebrowStyle,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  const SizedBox(height: 12),
                                  Text(
                                    banner.title.localized(context),
                                    style: _titleStyle,
                                    maxLines: 3,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                ],
                              ),
                            ),
                            if (banner.action == 'premium') ...[
                              const SizedBox(width: 14),
                              const Padding(
                                padding: EdgeInsets.only(bottom: 10),
                                child: BrandEmblem(premium: true),
                              ),
                            ],
                          ],
                        ),
                        const SizedBox(height: 10),
                        Text(
                          banner.description.localized(context),
                          style: _descriptionStyle,
                          maxLines: 3,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 18),
                        FilledButton.icon(
                          style: FilledButton.styleFrom(
                            backgroundColor:
                                (banner.action == 'premium' ||
                                    banner.action == 'manga')
                                ? SakuraColors.light.gold
                                : Colors.white,
                            foregroundColor:
                                (banner.action == 'premium' ||
                                    banner.action == 'manga')
                                ? SakuraColors.light.onGold
                                : color,
                            minimumSize: const Size(0, 48),
                            textStyle: Theme.of(
                              context,
                            ).textTheme.labelLarge?.merge(_buttonStyle),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                          ),
                          onPressed: onTap,
                          icon: const Icon(
                            Icons.arrow_forward_rounded,
                            size: 18,
                          ),
                          label: Text(
                            banner.buttonLabel.localized(context),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
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

/// Titik indikator: bulat kecil saat tidak aktif, memanjang menjadi pil
/// dengan isian merah yang berjalan selama jeda auto-slide saat aktif.
class _BannerDot extends StatelessWidget {
  const _BannerDot({
    required this.active,
    required this.progress,
    required this.showProgress,
  });

  final bool active;
  final Animation<double> progress;
  final bool showProgress;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return AnimatedContainer(
      duration: AppMotion.of(context, AppMotion.slow),
      curve: AppMotion.standard,
      width: active ? 26 : 6,
      height: 6,
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: active
            ? (showProgress
                  ? scheme.surfaceContainerHighest
                  : scheme.primary)
            : scheme.outlineVariant,
        borderRadius: BorderRadius.circular(6),
      ),
      child: active && showProgress
          ? AnimatedBuilder(
              animation: progress,
              builder: (context, _) => Align(
                alignment: Alignment.centerLeft,
                child: FractionallySizedBox(
                  widthFactor: progress.value.clamp(0, 1).toDouble(),
                  heightFactor: 1,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: scheme.primary,
                      borderRadius: BorderRadius.circular(6),
                    ),
                  ),
                ),
              ),
            )
          : null,
    );
  }
}
