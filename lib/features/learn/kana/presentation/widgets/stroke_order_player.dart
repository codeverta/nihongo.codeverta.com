import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/learn/kana/domain/stroke_matcher.dart';
import 'package:nihongo_master/features/learn/kana/presentation/widgets/kana_stroke_painter.dart';

/// Animated stroke order with a step-by-step mode (◀ ▶).
class StrokeOrderPlayer extends StatefulWidget {
  const StrokeOrderPlayer({required this.strokes, super.key});

  final List<Stroke> strokes;

  @override
  State<StrokeOrderPlayer> createState() => _StrokeOrderPlayerState();
}

class _StrokeOrderPlayerState extends State<StrokeOrderPlayer>
    with SingleTickerProviderStateMixin {
  static const _strokeMs = 650;
  static const _pauseMs = 250;

  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: Duration(
      milliseconds: widget.strokes.length * (_strokeMs + _pauseMs),
    ),
  )..addListener(() => setState(() {}));

  /// When set, the player shows exactly this many strokes (step mode).
  int? _step;

  bool _started = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_started) {
      _started = true;
      _play();
    }
  }

  @override
  void didUpdateWidget(covariant StrokeOrderPlayer oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.strokes != widget.strokes) {
      _controller.duration = Duration(
        milliseconds: widget.strokes.length * (_strokeMs + _pauseMs),
      );
      _play();
    }
  }

  void _play() {
    _step = null;
    if (MediaQuery.maybeDisableAnimationsOf(context) ?? false) {
      _controller.value = 1;
    } else {
      _controller.forward(from: 0);
    }
  }

  void _goTo(int step) {
    _controller.stop();
    setState(() => _step = step.clamp(1, widget.strokes.length));
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final total = widget.strokes.length;
    final int done;
    final int active;
    final double progress;
    if (_step case final step?) {
      done = step - 1;
      active = step - 1;
      progress = 1;
    } else {
      final elapsed = _controller.value * total * (_strokeMs + _pauseMs);
      final index = (elapsed ~/ (_strokeMs + _pauseMs)).clamp(0, total - 1);
      final within = elapsed - index * (_strokeMs + _pauseMs);
      done = index;
      active = index;
      progress = (within / _strokeMs).clamp(0.0, 1.0);
    }
    final shown = _step ?? (active + 1);
    return Column(
      children: [
        Semantics(
          label: 'Animasi urutan goresan'.localized(context),
          child: KanaSquare(
            child: CustomPaint(
              key: const Key('kana-stroke-player'),
              painter: KanaStrokePainter(
                reference: widget.strokes,
                colors: colors,
                showGuide: true,
                doneCount: done,
                activeIndex: active,
                activeProgress: progress,
                showActiveStroke: true,
                showActiveStart: true,
                numberDone: true,
              ),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          alignment: WrapAlignment.center,
          crossAxisAlignment: WrapCrossAlignment.center,
          spacing: 4,
          children: [
            IconButton(
              tooltip: 'Goresan sebelumnya'.localized(context),
              onPressed: shown > 1 ? () => _goTo(shown - 1) : null,
              icon: const Icon(Icons.chevron_left_rounded),
            ),
            Text(
              '${'Goresan'.localized(context)} $shown / $total',
              style: const TextStyle(fontWeight: FontWeight.w700),
            ),
            IconButton(
              tooltip: 'Goresan berikutnya'.localized(context),
              onPressed: shown < total || _step == null
                  ? () => _goTo(_step == null ? 1 : shown + 1)
                  : null,
              icon: const Icon(Icons.chevron_right_rounded),
            ),
            IconButton(
              key: const Key('kana-replay'),
              tooltip: 'Putar ulang animasi'.localized(context),
              onPressed: () => setState(_play),
              icon: const Icon(Icons.replay_rounded),
            ),
          ],
        ),
      ],
    );
  }
}
