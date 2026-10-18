import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

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
    with WidgetsBindingObserver {
  final _pages = PageController();
  Timer? _timer;
  int _index = 0;
  bool _paused = false;
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

  void _schedule() {
    _timer?.cancel();
    if (!widget.active ||
        !_resumed ||
        !_routeVisible ||
        !_motionAllowed ||
        _paused ||
        _hovered ||
        widget.banners.length < 2) {
      return;
    }
    _timer = Timer(const Duration(seconds: 5), () {
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
    if (!_paused) setState(() => _paused = true);
  }

  void _select(int index) {
    _pause();
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
              return 48 +
                  measure(_eyebrow(banner.action), _eyebrowStyle, 1) +
                  12 +
                  measure(banner.title, _titleStyle, 3) +
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
                child: NotificationListener<ScrollStartNotification>(
                  onNotification: (notification) {
                    if (notification.dragDetails != null) _pause();
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
                            'Banner ${i + 1} dari ${widget.banners.length}: ${widget.banners[i].title}',
                        child: InkWell(
                          key: Key('home-banner-dot-$i'),
                          onTap: () => _select(i),
                          borderRadius: BorderRadius.circular(24),
                          child: SizedBox(
                            width: 44,
                            height: 48,
                            child: Center(
                              child: AnimatedContainer(
                                duration: const Duration(milliseconds: 180),
                                width: i == _index ? 24 : 8,
                                height: 8,
                                decoration: BoxDecoration(
                                  color: i == _index
                                      ? Theme.of(context).colorScheme.primary
                                      : Theme.of(
                                          context,
                                        ).colorScheme.outlineVariant,
                                  borderRadius: BorderRadius.circular(8),
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    if (_motionAllowed)
                      IconButton(
                        key: const Key('home-banner-toggle'),
                        tooltip: _paused ? 'Putar banner' : 'Jeda banner',
                        icon: Icon(
                          _paused
                              ? Icons.play_arrow_rounded
                              : Icons.pause_rounded,
                        ),
                        onPressed: () {
                          setState(() => _paused = !_paused);
                          _schedule();
                        },
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
      'premium' => (const Color(0xFF292A87), 'assets/images/exams/n5.png'),
      'ssw' => (const Color(0xFF124E56), 'assets/images/ssw/ssw-careers.png'),
      'jobs' => (
        const Color(0xFF443060),
        'assets/images/jobs/job-category-atlas.png',
      ),
      _ => (const Color(0xFF272C75), 'assets/images/exams/n5.png'),
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
              Padding(
                padding: const EdgeInsets.all(24),
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: SizedBox(
                    width: textWidth,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          _eyebrow(banner.action),
                          style: _eyebrowStyle,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 12),
                        Text(
                          banner.title,
                          style: _titleStyle,
                          maxLines: 3,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 10),
                        Text(
                          banner.description,
                          style: _descriptionStyle,
                          maxLines: 3,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 18),
                        FilledButton.icon(
                          style: FilledButton.styleFrom(
                            backgroundColor: Colors.white,
                            foregroundColor: color,
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
                            banner.buttonLabel,
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
