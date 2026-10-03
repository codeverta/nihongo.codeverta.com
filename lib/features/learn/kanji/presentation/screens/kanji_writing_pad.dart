import 'package:flutter/material.dart';

/// A copy practice surface. It does not claim to grade stroke order.
class KanjiWritingPad extends StatefulWidget {
  const KanjiWritingPad({required this.character, super.key});
  final String character;

  @override
  State<KanjiWritingPad> createState() => _KanjiWritingPadState();
}

class _KanjiWritingPadState extends State<KanjiWritingPad> {
  final List<List<Offset>> _strokes = [];

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Expanded(
              child: Text(
                'Coba tulis ulang kanji',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w800),
              ),
            ),
            IconButton(
              tooltip: 'Hapus tulisan',
              onPressed: () => setState(_strokes.clear),
              icon: const Icon(Icons.refresh_rounded),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Center(
          child: Semantics(
            label: 'Area menyalin kanji ${widget.character}',
            child: SizedBox.square(
              dimension: 250,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: GestureDetector(
                  onPanStart: (d) => setState(
                    () => _strokes.add([d.localPosition]),
                  ),
                  onPanUpdate: (d) => setState(() {
                    if (_strokes.isNotEmpty) {
                      _strokes.last.add(d.localPosition);
                    }
                  }),
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      ColoredBox(color: colors.surfaceContainerHighest),
                      Center(
                        child: Text(
                          widget.character,
                          style: TextStyle(
                            fontSize: 175,
                            height: 1,
                            color: colors.onSurface.withValues(alpha: 0.13),
                          ),
                        ),
                      ),
                      CustomPaint(
                        painter: _WritingPainter(
                          _strokes,
                          colors.primary,
                          colors.outlineVariant,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _WritingPainter extends CustomPainter {
  _WritingPainter(this.strokes, this.ink, this.grid);
  final List<List<Offset>> strokes;
  final Color ink;
  final Color grid;

  @override
  void paint(Canvas canvas, Size size) {
    final gridPaint = Paint()
      ..color = grid
      ..strokeWidth = 1;
    canvas
      ..drawLine(
        Offset(size.width / 2, 0),
        Offset(size.width / 2, size.height),
        gridPaint,
      )
      ..drawLine(
        Offset(0, size.height / 2),
        Offset(size.width, size.height / 2),
        gridPaint,
      );
    final pen = Paint()
      ..color = ink
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;
    for (final stroke in strokes) {
      if (stroke.length == 1) {
        canvas.drawCircle(stroke.first, 2.5, Paint()..color = ink);
      } else if (stroke.isNotEmpty) {
        final path = Path()..moveTo(stroke.first.dx, stroke.first.dy);
        for (final point in stroke.skip(1)) {
          path.lineTo(point.dx, point.dy);
        }
        canvas.drawPath(path, pen);
      }
    }
  }

  @override
  bool shouldRepaint(covariant _WritingPainter oldDelegate) => true;
}
