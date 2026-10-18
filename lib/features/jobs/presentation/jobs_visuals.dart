import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_visuals.dart';

/// Warm, editorial introduction to the Japan job board.
class JobBoardHero extends StatelessWidget {
  const JobBoardHero({super.key});

  static const _ink = Color(0xFF1D2088);
  static const _paper = Color(0xFFFFF7EA);

  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    decoration: BoxDecoration(
      color: _paper,
      borderRadius: BorderRadius.circular(24),
      border: Border.all(color: const Color(0xFFEDE3D4)),
    ),
    clipBehavior: Clip.antiAlias,
    child: Stack(
      children: [
        Positioned(
          top: 0,
          bottom: 0,
          right: 0,
          width: MediaQuery.sizeOf(context).width * .39,
          child: Image.asset(
            'assets/images/ssw/ssw-careers.png',
            fit: BoxFit.cover,
            alignment: const Alignment(.78, .5),
            semanticLabel: 'Ilustrasi pekerja dari berbagai bidang di Jepang',
          ),
        ),
        Positioned.fill(
          child: IgnorePointer(
            child: DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  colors: [
                    _paper,
                    _paper.withValues(alpha: .94),
                    _paper.withValues(alpha: 0),
                  ],
                  stops: const [0, .53, 1],
                ),
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(18, 18, 112, 18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: .8),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  child: Text(
                    'PELUANG KERJA',
                    style: TextStyle(
                      color: _ink,
                      fontSize: 11,
                      letterSpacing: .7,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                'Langkah berikutnya\nmenuju Jepang',
                style: TextStyle(
                  color: _ink,
                  fontSize: 21,
                  height: 1.12,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                'Temukan peluang dari berbagai sumber terpercaya.',
                style: TextStyle(color: _ink, height: 1.3),
              ),
            ],
          ),
        ),
      ],
    ),
  );
}

/// Horizontally browsable category shortcuts backed by the server's metadata.
class JobCategoryStrip extends StatelessWidget {
  const JobCategoryStrip({
    required this.categories,
    required this.selectedId,
    required this.onSelected,
    super.key,
  });

  final List<JobOption> categories;
  final String? selectedId;
  final ValueChanged<String?> onSelected;

  @override
  Widget build(BuildContext context) {
    if (categories.isEmpty) return const SizedBox.shrink();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: Text(
            'Jelajahi bidang',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w800,
              color: const Color(0xFF1D2088),
            ),
          ),
        ),
        SizedBox(
          height: 150,
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                for (final entry in categories.asMap().entries)
                  Padding(
                    padding: EdgeInsets.only(
                      right: entry.key == categories.length - 1 ? 0 : 10,
                    ),
                    child: _JobCategoryCard(
                      item: entry.value,
                      selected: entry.value.id == selectedId,
                      onTap: () => onSelected(
                        entry.value.id == selectedId ? null : entry.value.id,
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _JobCategoryCard extends StatelessWidget {
  const _JobCategoryCard({
    required this.item,
    required this.selected,
    required this.onTap,
  });

  final JobOption item;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    const palettes = [
      (Color(0xFFE4ECFF), Color(0xFF1D2088)),
      (Color(0xFFE5F3E8), Color(0xFF275A38)),
      (Color(0xFFFFEBDD), Color(0xFF8A4218)),
      (Color(0xFFECE7F8), Color(0xFF503A80)),
    ];
    final palette =
        palettes[(_jobCategoryArtworkCells[item.id] ?? 18) % palettes.length];
    return SizedBox(
      width: 154,
      child: Semantics(
        button: true,
        selected: selected,
        label: 'Kategori ${item.label}',
        child: Material(
          color: Colors.white,
          clipBehavior: Clip.antiAlias,
          elevation: selected ? 2 : 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(18),
            side: BorderSide(
              color: selected ? palette.$2 : palette.$1,
              width: selected ? 2 : 1,
            ),
          ),
          child: InkWell(
            onTap: onTap,
            child: Column(
              children: [
                SizedBox(
                  width: double.infinity,
                  height: 100,
                  child: _JobCategoryArtwork(
                    categoryId: item.id,
                    fallbackColor: palette.$1,
                  ),
                ),
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 7,
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            item.label,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              color: palette.$2,
                              fontSize: 12,
                              height: 1.1,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
                        if (selected)
                          Icon(
                            Icons.check_circle_rounded,
                            color: palette.$2,
                            size: 18,
                          ),
                      ],
                    ),
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

IconData _jobCategoryIcon(String id) => switch (id) {
  'engineering' => Icons.code_rounded,
  'office' => Icons.business_center_outlined,
  'hospitality' => sswSectorIcon('accommodation'),
  'manufacturing' => sswSectorIcon('industrial-manufacturing'),
  'logistics' => sswSectorIcon('logistics-warehouse'),
  'transportation' => sswSectorIcon('road-transport'),
  'education' => Icons.school_outlined,
  'creative' => Icons.palette_outlined,
  'sales' => Icons.storefront_outlined,
  'finance' => Icons.account_balance_outlined,
  'healthcare' => Icons.medical_services_outlined,
  'general' => Icons.work_outline_rounded,
  _ => sswSectorIcon(id),
};

/// A generated contact sheet keeps the individual category artwork consistent
/// while rendering only the matching occupation in each card.
class _JobCategoryArtwork extends StatefulWidget {
  const _JobCategoryArtwork({
    required this.categoryId,
    required this.fallbackColor,
  });

  final String categoryId;
  final Color fallbackColor;

  @override
  State<_JobCategoryArtwork> createState() => _JobCategoryArtworkState();
}

class _JobCategoryArtworkState extends State<_JobCategoryArtwork> {
  ui.Image? _image;
  ImageStream? _stream;
  ImageStreamListener? _listener;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_stream != null) return;
    _stream = const AssetImage(
      'assets/images/jobs/job-category-atlas.png',
    ).resolve(createLocalImageConfiguration(context));
    _listener = ImageStreamListener((info, _) {
      if (mounted) setState(() => _image = info.image);
    });
    _stream!.addListener(_listener!);
  }

  @override
  void dispose() {
    if (_listener != null) _stream?.removeListener(_listener!);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ColoredBox(
    color: widget.fallbackColor,
    child: _image == null
        ? Center(
            child: Icon(
              _jobCategoryIcon(widget.categoryId),
              color: const Color(0xFF1D2088),
            ),
          )
        : CustomPaint(
            painter: _JobCategoryCropPainter(
              image: _image!,
              cell: _jobCategoryArtworkCells[widget.categoryId] ?? 18,
            ),
            child: const SizedBox.expand(),
          ),
  );
}

const _jobCategoryArtworkCells = <String, int>{
  'engineering': 0,
  'office': 1,
  'hospitality': 2,
  'nursing-care': 3,
  'building-cleaning': 4,
  'manufacturing': 5,
  'construction': 6,
  'agriculture': 7,
  'fishery': 8,
  'food-service': 9,
  'food-manufacturing': 10,
  'logistics': 11,
  'transportation': 12,
  'education': 13,
  'creative': 14,
  'sales': 15,
  'finance': 16,
  'healthcare': 17,
  'general': 18,
};

class _JobCategoryCropPainter extends CustomPainter {
  const _JobCategoryCropPainter({required this.image, required this.cell});

  final ui.Image image;
  final int cell;

  @override
  void paint(Canvas canvas, Size size) {
    const columns = 4;
    const rows = 5;
    final cellWidth = image.width / columns;
    final cellHeight = image.height / rows;
    final col = cell % columns;
    final row = cell ~/ columns;
    final source = Rect.fromLTWH(
      col * cellWidth + 4,
      row * cellHeight + 4,
      cellWidth - 8,
      cellHeight - 8,
    );
    final fitted = applyBoxFit(BoxFit.cover, source.size, size);
    final input = Alignment.center.inscribe(fitted.source, source);
    final output = Alignment.center.inscribe(
      fitted.destination,
      Offset.zero & size,
    );
    canvas.drawImageRect(
      image,
      input,
      output,
      Paint()..filterQuality = FilterQuality.high,
    );
  }

  @override
  bool shouldRepaint(_JobCategoryCropPainter oldDelegate) =>
      oldDelegate.image != image || oldDelegate.cell != cell;
}
