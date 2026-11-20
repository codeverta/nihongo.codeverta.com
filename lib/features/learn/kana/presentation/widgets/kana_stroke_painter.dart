import 'package:flutter/material.dart';
import 'package:nihongo_master/features/learn/kana/domain/stroke_matcher.dart';

/// Paints a practice square: grid, faint guide, finished strokes, the
/// highlighted next stroke and the learner's live stroke. All strokes are in
/// the unit square and scaled to the canvas.
class KanaStrokePainter extends CustomPainter {
  KanaStrokePainter({
    required this.reference,
    required this.colors,
    this.showGuide = false,
    this.doneCount = 0,
    this.activeIndex,
    this.activeProgress = 1,
    this.showActiveStroke = false,
    this.showActiveStart = false,
    this.numberDone = false,
    this.live,
  });

  final List<Stroke> reference;
  final ColorScheme colors;
  final bool showGuide;
  final int doneCount;
  final int? activeIndex;
  final double activeProgress;
  final bool showActiveStroke;
  final bool showActiveStart;
  final bool numberDone;
  final Stroke? live;

  @override
  void paint(Canvas canvas, Size size) {
    final grid = Paint()
      ..color = colors.outlineVariant
      ..strokeWidth = 1;
    final mid = size.width / 2;
    for (var d = 0.0; d < size.width; d += 8) {
      canvas
        ..drawLine(Offset(mid, d), Offset(mid, d + 4), grid)
        ..drawLine(Offset(d, mid), Offset(d + 4, mid), grid);
    }

    final width = size.width * 0.05;
    Paint pen(Color color) => Paint()
      ..color = color
      ..strokeWidth = width
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;

    if (showGuide) {
      final guide = pen(colors.onSurface.withValues(alpha: 0.12));
      for (final stroke in reference) {
        _drawStroke(canvas, size, stroke, guide);
      }
    }
    final ink = pen(colors.onSurface);
    for (var i = 0; i < doneCount && i < reference.length; i++) {
      _drawStroke(canvas, size, reference[i], ink);
    }
    final active = activeIndex;
    if (active != null && active < reference.length) {
      final stroke = reference[active];
      if (showActiveStroke) {
        _drawStroke(
          canvas,
          size,
          StrokeMatcher.partial(stroke, activeProgress),
          pen(colors.primary.withValues(alpha: 0.85)),
        );
      }
      if (showActiveStart) {
        _startDot(canvas, size, stroke, active + 1, width);
      }
    }
    if (numberDone) {
      for (var i = 0; i < doneCount && i < reference.length; i++) {
        if (i != active) _number(canvas, size, reference[i], i + 1, width);
      }
    }
    final drawing = live;
    if (drawing != null && drawing.isNotEmpty) {
      _drawStroke(canvas, size, drawing, pen(colors.tertiary));
    }
  }

  void _drawStroke(Canvas canvas, Size size, Stroke stroke, Paint paint) {
    if (stroke.isEmpty) return;
    final first = _scale(stroke.first, size);
    if (stroke.length == 1) {
      canvas.drawCircle(
        first,
        paint.strokeWidth / 2,
        Paint()..color = paint.color,
      );
      return;
    }
    final path = Path()..moveTo(first.dx, first.dy);
    for (final point in stroke.skip(1)) {
      final p = _scale(point, size);
      path.lineTo(p.dx, p.dy);
    }
    canvas.drawPath(path, paint);
  }

  void _startDot(Canvas canvas, Size size, Stroke stroke, int n, double r) {
    final start = _scale(stroke.first, size);
    canvas.drawCircle(start, r * 0.9, Paint()..color = colors.error);
    _label(canvas, start + Offset(r * 1.6, -r * 1.6), '$n', colors.error);
  }

  void _number(Canvas canvas, Size size, Stroke stroke, int n, double r) {
    final start = _scale(stroke.first, size);
    _label(
      canvas,
      start + Offset(-r * 1.6, -r * 1.6),
      '$n',
      colors.onSurfaceVariant,
    );
  }

  void _label(Canvas canvas, Offset center, String text, Color color) {
    final painter = TextPainter(
      text: TextSpan(
        text: text,
        style: TextStyle(
          color: color,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    painter.paint(
      canvas,
      center - Offset(painter.width / 2, painter.height / 2),
    );
  }

  Offset _scale(Offset unit, Size size) =>
      Offset(unit.dx * size.width, unit.dy * size.height);

  @override
  bool shouldRepaint(covariant KanaStrokePainter old) => true;
}

/// Rounded square frame shared by the viewer and the practice pad.
class KanaSquare extends StatelessWidget {
  const KanaSquare({required this.child, this.maxSize = 280, super.key});

  final Widget child;
  final double maxSize;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, constraints) {
      final side = constraints.maxWidth.clamp(120.0, maxSize);
      return Center(
        child: SizedBox.square(
          dimension: side,
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(18),
            ),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(18),
              child: child,
            ),
          ),
        ),
      );
    },
  );
}
