import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/jobs/data/jobs_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
import 'package:nihongo_master/features/map/presentation/map_labels.dart';

/// First jobs shown in a prefecture's bottom sheet.
final mapPrefectureJobsProvider = FutureProvider.autoDispose
    .family<JobPage, String>(
      (ref, prefecture) =>
          ref.watch(jobsRepositoryProvider).list({'prefecture': prefecture}),
    );

class JapanMapScreen extends ConsumerStatefulWidget {
  const JapanMapScreen({super.key});
  @override
  ConsumerState<JapanMapScreen> createState() => _JapanMapScreenState();
}

class _JapanMapScreenState extends ConsumerState<JapanMapScreen>
    with SingleTickerProviderStateMixin {
  final _transform = TransformationController();
  late final AnimationController _zoom = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 450),
  );
  Animation<Matrix4>? _zoomAnimation;
  bool _showSpots = true;
  bool _showJobs = true;
  final Set<String> _categories = {...mapCategories};
  String? _selected;

  @override
  void initState() {
    super.initState();
    _zoom.addListener(() {
      final animation = _zoomAnimation;
      if (animation != null) _transform.value = animation.value;
    });
  }

  @override
  void dispose() {
    _zoom.dispose();
    _transform.dispose();
    super.dispose();
  }

  void _animateTo(Matrix4 target) {
    _zoomAnimation = Matrix4Tween(begin: _transform.value, end: target).animate(
      CurvedAnimation(parent: _zoom, curve: AppMotion.standard),
    );
    unawaited(_zoom.forward(from: 0));
  }

  void _resetZoom() {
    setState(() => _selected = null);
    _animateTo(Matrix4.identity());
  }

  Size _viewport = Size.zero;

  /// Memperbesar/memperkecil di sekitar tengah layar (alternatif cubit).
  void _zoomBy(double factor) {
    final current = _transform.value.clone();
    final scale = current.getMaxScaleOnAxis();
    final next = (scale * factor).clamp(1.0, 10.0);
    if (next <= 1.0001) {
      _animateTo(Matrix4.identity());
      return;
    }
    final f = next / scale;
    final c = _viewport.center(Offset.zero);
    final around = Matrix4.translationValues(c.dx, c.dy, 0)
        .multiplied(Matrix4.diagonal3Values(f, f, 1))
        .multiplied(Matrix4.translationValues(-c.dx, -c.dy, 0));
    _animateTo(around.multiplied(current));
  }

  /// Zooms so [bounds] (viewport coordinates at scale 1) fills most of the
  /// viewport, keeping room at the bottom for the sheet.
  void _zoomTo(Rect bounds, Size viewport) {
    final scale = math
        .min(
          viewport.width * .7 / math.max(bounds.width, 1),
          viewport.height * .45 / math.max(bounds.height, 1),
        )
        .clamp(1.0, 8.0)
        .toDouble();
    final dx = viewport.width / 2 - bounds.center.dx * scale;
    final dy = viewport.height * .3 - bounds.center.dy * scale;
    _animateTo(
      Matrix4(scale, 0, 0, 0, 0, scale, 0, 0, 0, 0, 1, 0, dx, dy, 0, 1),
    );
  }

  Future<void> _openPrefecture(
    PrefectureShape prefecture,
    Rect bounds,
    Size viewport,
    List<MapSpot> spots,
    JobCounts counts,
  ) async {
    setState(() => _selected = prefecture.code);
    _zoomTo(bounds, viewport);
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (context) => _PrefectureSheet(
        prefecture: prefecture,
        spots: spots.where((s) => s.prefecture == prefecture.code).toList(),
        jobCount: counts[prefecture.code],
      ),
    );
    if (mounted) setState(() => _selected = null);
  }

  @override
  Widget build(BuildContext context) {
    final geometry = ref.watch(japanGeometryProvider);
    final spots = ref.watch(mapSpotsProvider);
    final counts =
        ref.watch(mapJobCountsProvider).valueOrNull ??
        const JobCounts(counts: {});
    final allSpots = spots.valueOrNull?.spots ?? const <MapSpot>[];
    final visibleSpots = _showSpots
        ? allSpots.where((s) => _categories.contains(s.category)).toList()
        : const <MapSpot>[];
    return Scaffold(
      appBar: AppBar(
        title: Text('Peta Jepang'.localized(context)),
        actions: [
          IconButton(
            tooltip: 'Tampilkan seluruh Jepang'.localized(context),
            onPressed: _resetZoom,
            icon: const Icon(Icons.zoom_out_map_rounded),
          ),
        ],
      ),
      body: Column(
        children: [
          _Filters(
            showSpots: _showSpots,
            showJobs: _showJobs,
            categories: _categories,
            onSpots: (value) => setState(() => _showSpots = value),
            onJobs: (value) => setState(() => _showJobs = value),
            onCategory: (category, value) => setState(() {
              value ? _categories.add(category) : _categories.remove(category);
            }),
          ),
          if (spots.valueOrNull?.fromCache ?? false)
            _Notice(
              icon: Icons.cloud_off_rounded,
              text: 'Offline: menampilkan data peta yang tersimpan.'
                  .localized(context),
            ),
          if (spots.hasError)
            _Notice(
              icon: Icons.error_outline_rounded,
              text: 'Tempat belum dapat dimuat. Periksa koneksi dan coba lagi.'
                  .localized(context),
              action: TextButton(
                onPressed: () => ref.invalidate(mapSpotsProvider),
                child: Text('Coba lagi'.localized(context)),
              ),
            ),
          Expanded(
            child: ColoredBox(
              color: SakuraColors.of(context).mapSea,
              child: Stack(
                children: [
                  Positioned.fill(
                    child: geometry.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => Center(
                child: Text('Peta tidak dapat ditampilkan.'.localized(context)),
              ),
              data: (geometry) => LayoutBuilder(
                builder: (context, constraints) {
                  final viewport = constraints.biggest;
                  final mapSize =
                      viewport.width / viewport.height > geometry.aspectRatio
                      ? Size(
                          viewport.height * geometry.aspectRatio,
                          viewport.height,
                        )
                      : Size(
                          viewport.width,
                          viewport.width / geometry.aspectRatio,
                        );
                  final origin = Offset(
                    (viewport.width - mapSize.width) / 2,
                    (viewport.height - mapSize.height) / 2,
                  );
                  _viewport = viewport;
                  return _MapCanvas(
                    key: ValueKey(mapSize),
                    geometry: geometry,
                    mapSize: mapSize,
                    origin: origin,
                    viewport: viewport,
                    transform: _transform,
                    selected: _selected,
                    spots: visibleSpots,
                    counts: _showJobs ? counts : const JobCounts(counts: {}),
                    onPrefecture: (prefecture, bounds) => _openPrefecture(
                      prefecture,
                      bounds.shift(origin),
                      viewport,
                      allSpots,
                      counts,
                    ),
                    onSpot: (spot) => context.push('/map/spot/${spot.id}'),
                  );
                },
              ),
                    ),
                  ),
                  Positioned(
                    right: 14,
                    bottom: 20,
                    child: _ZoomControls(
                      onZoomIn: () => _zoomBy(1.6),
                      onZoomOut: () => _zoomBy(1 / 1.6),
                      onReset: _resetZoom,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _Filters extends StatelessWidget {
  const _Filters({
    required this.showSpots,
    required this.showJobs,
    required this.categories,
    required this.onSpots,
    required this.onJobs,
    required this.onCategory,
  });
  final bool showSpots;
  final bool showJobs;
  final Set<String> categories;
  final ValueChanged<bool> onSpots;
  final ValueChanged<bool> onJobs;
  final void Function(String category, bool value) onCategory;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final onSelected = Theme.of(context).brightness == Brightness.dark
        ? AppTheme.darkBackground
        : Colors.white;
    Color layerIcon({required bool selected}) =>
        selected ? onSelected : scheme.onSurfaceVariant;
    return Column(
    children: [
      SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
        child: Row(
          children: [
            FilterChip(
              key: const Key('map-layer-spots'),
              avatar: Icon(
                Icons.place_rounded,
                size: 18,
                color: layerIcon(selected: showSpots),
              ),
              showCheckmark: false,
              label: Text('Tempat & budaya'.localized(context)),
              selected: showSpots,
              onSelected: onSpots,
            ),
            const SizedBox(width: 8),
            FilterChip(
              key: const Key('map-layer-jobs'),
              avatar: Icon(
                Icons.work_rounded,
                size: 18,
                color: layerIcon(selected: showJobs),
              ),
              showCheckmark: false,
              label: Text('Lowongan kerja'.localized(context)),
              selected: showJobs,
              onSelected: onJobs,
            ),
          ],
        ),
      ),
      AnimatedSize(
        duration: AppMotion.of(context, AppMotion.slow),
        curve: AppMotion.standard,
        alignment: Alignment.topCenter,
        child: showSpots
            ? SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
              child: Row(
                children: [
                  for (final category in mapCategories)
                    Padding(
                      padding: const EdgeInsets.only(right: 6),
                      child: FilterChip(
                        visualDensity: VisualDensity.compact,
                        showCheckmark: false,
                        selectedColor: categoryColor(
                          category,
                        ).withValues(alpha: .16),
                        side: BorderSide(
                          color: categories.contains(category)
                              ? categoryColor(category).withValues(alpha: .5)
                              : scheme.outlineVariant,
                        ),
                        labelStyle: TextStyle(
                          fontWeight: FontWeight.w700,
                          fontSize: 12.5,
                          color: categories.contains(category)
                              ? scheme.onSurface
                              : scheme.onSurfaceVariant,
                        ),
                        avatar: Icon(
                          categoryIcon(category),
                          size: 16,
                          color: categoryColor(category),
                        ),
                        label: Text(categoryLabel(context, category)),
                        selected: categories.contains(category),
                        onSelected: (value) => onCategory(category, value),
                      ),
                    ),
                ],
              ),
            )
            : const SizedBox(width: double.infinity),
      ),
    ],
  );
  }
}

class _ZoomControls extends StatelessWidget {
  const _ZoomControls({
    required this.onZoomIn,
    required this.onZoomOut,
    required this.onReset,
  });
  final VoidCallback onZoomIn;
  final VoidCallback onZoomOut;
  final VoidCallback onReset;

  @override
  Widget build(BuildContext context) {
    final sakura = SakuraColors.of(context);
    final scheme = Theme.of(context).colorScheme;
    Widget button(String tooltip, IconData icon, VoidCallback onTap, Key key) {
      return IconButton(
        key: key,
        tooltip: tooltip,
        onPressed: onTap,
        icon: Icon(icon, size: 20),
        color: scheme.onSurface,
      );
    }

    return DecoratedBox(
      decoration: BoxDecoration(
        color: sakura.glass,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: scheme.outlineVariant),
        boxShadow: [
          BoxShadow(
            color: sakura.shadow.withValues(alpha: 0.12),
            blurRadius: 18,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          button(
            'Perbesar'.localized(context),
            Icons.add_rounded,
            onZoomIn,
            const Key('map-zoom-in'),
          ),
          Container(width: 24, height: 1, color: scheme.outlineVariant),
          button(
            'Perkecil'.localized(context),
            Icons.remove_rounded,
            onZoomOut,
            const Key('map-zoom-out'),
          ),
          Container(width: 24, height: 1, color: scheme.outlineVariant),
          button(
            'Tampilkan seluruh Jepang'.localized(context),
            Icons.center_focus_strong_rounded,
            onReset,
            const Key('map-zoom-reset'),
          ),
        ],
      ),
    );
  }
}

/// Pin "jatuh" dan membesar saat pertama tampil.
class _PinEntrance extends StatelessWidget {
  const _PinEntrance({required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: 1),
      duration: AppMotion.of(context, const Duration(milliseconds: 420)),
      curve: AppMotion.spring,
      child: child,
      builder: (context, t, child) => Opacity(
        opacity: t.clamp(0, 1).toDouble(),
        child: Transform.translate(
          offset: Offset(0, (1 - t) * -10),
          child: Transform.scale(
            scale: 0.6 + 0.4 * t,
            alignment: Alignment.bottomCenter,
            child: child,
          ),
        ),
      ),
    );
  }
}

class _Notice extends StatelessWidget {
  const _Notice({required this.icon, required this.text, this.action});
  final IconData icon;
  final String text;
  final Widget? action;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(16, 2, 16, 8),
    child: Material(
      color: Theme.of(context).brightness == Brightness.dark
          ? AppTheme.darkSurface
          : Colors.white,
      borderRadius: BorderRadius.circular(14),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 6, 6, 6),
        child: Row(
        children: [
          Icon(icon, size: 18, color: Theme.of(context).colorScheme.primary),
          const SizedBox(width: 8),
          Expanded(child: Text(text, style: const TextStyle(fontSize: 13))),
          if (action != null) action!,
        ],
      ),
      ),
    ),
  );
}

class _MapCanvas extends StatefulWidget {
  const _MapCanvas({
    required this.geometry,
    required this.mapSize,
    required this.origin,
    required this.viewport,
    required this.transform,
    required this.selected,
    required this.spots,
    required this.counts,
    required this.onPrefecture,
    required this.onSpot,
    super.key,
  });
  final JapanGeometry geometry;
  final Size mapSize;
  final Offset origin;
  final Size viewport;
  final TransformationController transform;
  final String? selected;
  final List<MapSpot> spots;
  final JobCounts counts;
  final void Function(PrefectureShape prefecture, Rect bounds) onPrefecture;
  final ValueChanged<MapSpot> onSpot;

  @override
  State<_MapCanvas> createState() => _MapCanvasState();
}

class _MapCanvasState extends State<_MapCanvas> {
  late final Map<String, Path> _paths = widget.geometry.paths(widget.mapSize);

  PrefectureShape? _hit(Offset point) {
    for (final entry in _paths.entries) {
      if (entry.value.contains(point)) {
        return widget.geometry.byCode(entry.key);
      }
    }
    return null;
  }

  void _openPrefectureCode(String code) {
    final prefecture = widget.geometry.byCode(code);
    final path = _paths[code];
    if (prefecture != null && path != null) {
      widget.onPrefecture(prefecture, path.getBounds());
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final size = widget.mapSize;
    final jobPins = [
      for (final prefecture in widget.geometry.prefectures)
        if (widget.counts[prefecture.code] > 0)
          (
            prefecture,
            widget.geometry.toCanvas(
              prefecture.code,
              prefecture.capitalLat,
              prefecture.capitalLng,
              size,
            ),
          ),
    ];
    return InteractiveViewer(
      transformationController: widget.transform,
      minScale: 1,
      maxScale: 10,
      boundaryMargin: const EdgeInsets.all(120),
      child: SizedBox.fromSize(
        size: widget.viewport,
        child: Stack(
          children: [
            Positioned(
              left: widget.origin.dx,
              top: widget.origin.dy,
              width: size.width,
              height: size.height,
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTapUp: (details) {
                  final prefecture = _hit(details.localPosition);
                  if (prefecture != null) _openPrefectureCode(prefecture.code);
                },
                child: CustomPaint(
                  key: const Key('japan-map'),
                  size: size,
                  painter: _MapPainter(
                    paths: _paths,
                    selected: widget.selected,
                    inset: widget.geometry.insetFrame(size),
                    fill: SakuraColors.of(context).mapLand,
                    jobFill: jobPinColor.withValues(alpha: .07),
                    jobPrefectures: {
                      for (final pin in jobPins) pin.$1.code,
                    },
                    selectedFill: SakuraColors.of(context).mapLandActive,
                    selectedStroke: colors.primary,
                    stroke: SakuraColors.of(context).mapStroke,
                  ),
                ),
              ),
            ),
            Positioned.fill(
              child: ValueListenableBuilder<Matrix4>(
                valueListenable: widget.transform,
                builder: (context, matrix, _) {
                  final scale = matrix.getMaxScaleOnAxis();
                  final detailed = scale >= 2.2;
                  return Stack(
                    clipBehavior: Clip.none,
                    children: [
                      for (final spot in widget.spots)
                        _pin(
                          widget.origin +
                              widget.geometry.toCanvas(
                                spot.prefecture,
                                spot.lat,
                                spot.lng,
                                size,
                              ),
                          scale,
                          _SpotPin(
                            spot: spot,
                            detailed: detailed,
                            onTap: () => widget.onSpot(spot),
                          ),
                          detailed ? 30 : 16,
                        ),
                      for (final (prefecture, point) in jobPins)
                        _pin(
                          widget.origin + point + const Offset(10, -6),
                          scale,
                          _JobPin(
                            key: Key('job-pin-${prefecture.code}'),
                            count: widget.counts[prefecture.code],
                            onTap: () => _openPrefectureCode(prefecture.code),
                          ),
                          28,
                        ),
                    ],
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Places a pin centered on [point] that keeps its on-screen size while the
  /// map is zoomed.
  Widget _pin(Offset point, double scale, Widget child, double extent) {
    final box = extent * 3;
    return Positioned(
      left: point.dx - box / 2,
      top: point.dy - box / 2,
      width: box,
      height: box,
      child: Transform.scale(
        scale: 1 / scale,
        child: Center(child: _PinEntrance(child: child)),
      ),
    );
  }
}

class _MapPainter extends CustomPainter {
  _MapPainter({
    required this.paths,
    required this.selected,
    required this.inset,
    required this.fill,
    required this.jobFill,
    required this.jobPrefectures,
    required this.selectedFill,
    required this.selectedStroke,
    required this.stroke,
  });
  final Map<String, Path> paths;
  final String? selected;
  final Rect inset;
  final Color fill;
  final Color jobFill;
  final Set<String> jobPrefectures;
  final Color selectedFill;
  final Color selectedStroke;
  final Color stroke;

  @override
  void paint(Canvas canvas, Size size) {
    final base = Paint()..color = fill;
    final job = Paint()..color = jobFill;
    final highlight = Paint()..color = selectedFill;
    final line = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = .7
      ..color = stroke;
    for (final entry in paths.entries) {
      canvas.drawPath(entry.value, base);
      if (jobPrefectures.contains(entry.key)) canvas.drawPath(entry.value, job);
      if (entry.key == selected) canvas.drawPath(entry.value, highlight);
      canvas.drawPath(entry.value, line);
    }
    final selectedPath = selected == null ? null : paths[selected];
    if (selectedPath != null) {
      canvas.drawPath(
        selectedPath,
        Paint()
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1.6
          ..color = selectedStroke,
      );
    }
    if (inset != Rect.zero) {
      canvas.drawRRect(
        RRect.fromRectAndRadius(inset, const Radius.circular(8)),
        Paint()
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1
          ..color = stroke,
      );
    }
  }

  @override
  bool shouldRepaint(_MapPainter old) =>
      old.selected != selected ||
      old.fill != fill ||
      old.stroke != stroke ||
      old.jobPrefectures.length != jobPrefectures.length ||
      old.paths != paths;
}

class _SpotPin extends StatelessWidget {
  const _SpotPin({
    required this.spot,
    required this.detailed,
    required this.onTap,
  });
  final MapSpot spot;
  final bool detailed;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) {
    final color = categoryColor(spot.category);
    final extent = detailed ? 30.0 : 14.0;
    return Semantics(
      button: true,
      label: spot.nameFor(mapLanguage(context)),
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          width: extent,
          height: extent,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white, width: detailed ? 2 : 1.5),
            boxShadow: const [
              BoxShadow(color: Color(0x33000000), blurRadius: 3),
            ],
          ),
          child: detailed
              ? Icon(categoryIcon(spot.category), size: 16, color: Colors.white)
              : null,
        ),
      ),
    );
  }
}

class _JobPin extends StatelessWidget {
  const _JobPin({required this.count, required this.onTap, super.key});
  final int count;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => Semantics(
    button: true,
    label: '$count ${'lowongan'.localized(context)}',
    child: GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
        decoration: BoxDecoration(
          color: jobPinColor,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.white, width: 1.5),
          boxShadow: const [BoxShadow(color: Color(0x33000000), blurRadius: 3)],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.work_rounded, size: 12, color: Colors.white),
            const SizedBox(width: 3),
            Text(
              count > 999 ? '999+' : '$count',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class _PrefectureSheet extends ConsumerWidget {
  const _PrefectureSheet({
    required this.prefecture,
    required this.spots,
    required this.jobCount,
  });
  final PrefectureShape prefecture;
  final List<MapSpot> spots;
  final int jobCount;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final lang = mapLanguage(context);
    final theme = Theme.of(context);
    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: .5,
      minChildSize: .3,
      maxChildSize: .9,
      builder: (context, scroll) => ListView(
        controller: scroll,
        padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
        children: [
          Text(
            prefecture.nameJa,
            style: AppTypography.japanese(
              fontSize: 13,
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          Text(
            prefecture.name,
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.w800,
              letterSpacing: -0.4,
            ),
          ),
          const SizedBox(height: 10),
          Wrap(
            spacing: 6,
            runSpacing: 6,
            children: [
              _CountPill(
                text: '${spots.length} ${'tempat'.localized(context)}',
                background: theme.colorScheme.primaryContainer,
                foreground: theme.colorScheme.brightness == Brightness.dark
                    ? theme.colorScheme.primary
                    : theme.colorScheme.onPrimaryContainer,
              ),
              _CountPill(
                text: '$jobCount ${'lowongan'.localized(context)}',
                background: theme.colorScheme.secondaryContainer,
                foreground: theme.colorScheme.onSecondaryContainer,
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            'Tempat & budaya'.localized(context),
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          if (spots.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Text(
                'Belum ada tempat di prefektur ini.'.localized(context),
              ),
            ),
          for (final spot in spots)
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: Container(
                width: 42,
                height: 42,
                decoration: BoxDecoration(
                  color: categoryColor(spot.category).withValues(alpha: .14),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  categoryIcon(spot.category),
                  color: categoryColor(spot.category),
                ),
              ),
              title: Text(spot.nameFor(lang)),
              subtitle: Text(
                [
                  categoryLabel(context, spot.category),
                  if (spot.nameJa.isNotEmpty) spot.nameJa,
                ].join(' · '),
              ),
              trailing: const Icon(Icons.chevron_right_rounded),
              onTap: () {
                final router = GoRouter.of(context);
                Navigator.of(context).pop();
                unawaited(router.push('/map/spot/${spot.id}'));
              },
            ),
          const SizedBox(height: 16),
          Text(
            'Lowongan kerja'.localized(context),
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          if (jobCount == 0)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Text(
                'Belum ada lowongan di prefektur ini.'.localized(context),
              ),
            )
          else ...[
            ref
                .watch(mapPrefectureJobsProvider(prefecture.code))
                .when(
                  loading: () => const Padding(
                    padding: EdgeInsets.all(12),
                    child: Center(child: CircularProgressIndicator()),
                  ),
                  error: (_, __) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    child: Text(
                      'Lowongan belum dapat dimuat.'.localized(context),
                    ),
                  ),
                  data: (page) => Column(
                    children: [
                      for (final job in page.items.take(3))
                        ListTile(
                          contentPadding: EdgeInsets.zero,
                          leading: const Icon(Icons.work_outline_rounded),
                          title: Text(job.title),
                          subtitle: Text(job.company),
                          onTap: () {
                            final router = GoRouter.of(context);
                            Navigator.of(context).pop();
                            unawaited(router.push('/jobs/${job.id}'));
                          },
                        ),
                    ],
                  ),
                ),
            const SizedBox(height: 8),
            FilledButton.icon(
              key: const Key('map-all-jobs'),
              onPressed: () {
                final router = GoRouter.of(context);
                Navigator.of(context).pop();
                unawaited(router.push('/jobs?prefecture=${prefecture.code}'));
              },
              icon: const Icon(Icons.work_rounded),
              label: Text(
                '${'Lihat semua lowongan'.localized(context)} ($jobCount)',
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _CountPill extends StatelessWidget {
  const _CountPill({
    required this.text,
    required this.background,
    required this.foreground,
  });
  final String text;
  final Color background;
  final Color foreground;

  @override
  Widget build(BuildContext context) => DecoratedBox(
    decoration: BoxDecoration(
      color: background,
      borderRadius: BorderRadius.circular(999),
    ),
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w700,
          color: foreground,
        ),
      ),
    ),
  );
}
