import 'dart:async';
import 'dart:ui' show FontFeature;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';

/// Membungkus area yang bisa diketuk dengan umpan balik "tekan": mengecil
/// sedikit lalu memantul kembali. Tidak mengubah perilaku [onTap].
class Pressable extends StatefulWidget {
  const Pressable({
    required this.child,
    this.onTap,
    this.onLongPress,
    this.scale = 0.96,
    this.borderRadius,
    this.haptic = false,
    this.semanticsLabel,
    super.key,
  });

  final Widget child;
  final VoidCallback? onTap;
  final VoidCallback? onLongPress;
  final double scale;
  final BorderRadius? borderRadius;
  final bool haptic;
  final String? semanticsLabel;

  @override
  State<Pressable> createState() => _PressableState();
}

class _PressableState extends State<Pressable> {
  bool _down = false;

  void _set(bool value) {
    if (_down == value || !mounted) return;
    setState(() => _down = value);
  }

  @override
  Widget build(BuildContext context) {
    final enabled = widget.onTap != null || widget.onLongPress != null;
    final reduced = AppMotion.reduced(context);
    return AnimatedScale(
      scale: _down && enabled && !reduced ? widget.scale : 1,
      duration: _down ? AppMotion.fast : AppMotion.normal,
      curve: _down ? AppMotion.enter : AppMotion.spring,
      child: Material(
        type: MaterialType.transparency,
        child: InkWell(
          onTap: widget.onTap == null
              ? null
              : () {
                  if (widget.haptic) unawaited(HapticFeedback.selectionClick());
                  widget.onTap!();
                },
          onLongPress: widget.onLongPress,
          onHighlightChanged: _set,
          borderRadius: widget.borderRadius,
          child: widget.semanticsLabel == null
              ? widget.child
              : Semantics(label: widget.semanticsLabel, child: widget.child),
        ),
      ),
    );
  }
}

/// Memunculkan [child] dengan fade + geser naik, tertunda sesuai [index]
/// agar daftar tampil bertahap. Diputar sekali saat pertama dibangun, atau
/// lagi bila [replayKey] berubah.
class StaggeredEntrance extends StatefulWidget {
  const StaggeredEntrance({
    required this.child,
    this.index = 0,
    this.step = const Duration(milliseconds: 45),
    this.duration = AppMotion.slow,
    this.offset = 12,
    this.replayKey,
    super.key,
  });

  final Widget child;
  final int index;
  final Duration step;
  final Duration duration;
  final double offset;
  final Object? replayKey;

  @override
  State<StaggeredEntrance> createState() => _StaggeredEntranceState();
}

class _StaggeredEntranceState extends State<StaggeredEntrance>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: widget.duration,
  );
  late final CurvedAnimation _curve = CurvedAnimation(
    parent: _controller,
    curve: AppMotion.enter,
  );
  Timer? _timer;
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
  void didUpdateWidget(covariant StaggeredEntrance oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.replayKey != widget.replayKey) _play();
  }

  void _play() {
    _timer?.cancel();
    if (AppMotion.reduced(context)) {
      _controller.value = 1;
      return;
    }
    _controller.value = 0;
    final delay = widget.step * widget.index.clamp(0, 12);
    if (delay == Duration.zero) {
      unawaited(_controller.forward());
    } else {
      _timer = Timer(delay, () {
        if (mounted) unawaited(_controller.forward());
      });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    _curve.dispose();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final curved = _curve;
    return AnimatedBuilder(
      animation: curved,
      child: widget.child,
      builder: (context, child) => Opacity(
        opacity: curved.value,
        child: Transform.translate(
          offset: Offset(0, (1 - curved.value) * widget.offset),
          child: child,
        ),
      ),
    );
  }
}

/// Angka yang menghitung naik ke [value] saat pertama tampil atau berubah.
class AnimatedCount extends StatelessWidget {
  const AnimatedCount({
    required this.value,
    this.style,
    this.suffix = '',
    super.key,
  });

  final int value;
  final TextStyle? style;
  final String suffix;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: value.toDouble()),
      duration: AppMotion.of(context, const Duration(milliseconds: 900)),
      curve: Curves.easeOutCubic,
      builder: (context, current, _) => Text(
        '${current.round()}$suffix',
        style: (style ?? const TextStyle()).copyWith(
          fontFeatures: const [FontFeature.tabularFigures()],
        ),
      ),
    );
  }
}

/// Bar progres yang terisi dengan animasi.
class AnimatedProgressBar extends StatelessWidget {
  const AnimatedProgressBar({
    required this.value,
    this.color,
    this.background,
    this.height = 6,
    super.key,
  });

  final double value;
  final Color? color;
  final Color? background;
  final double height;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: value.clamp(0, 1).toDouble()),
      duration: AppMotion.of(context, const Duration(milliseconds: 900)),
      curve: AppMotion.enter,
      builder: (context, current, _) => ClipRRect(
        borderRadius: BorderRadius.circular(height),
        child: LinearProgressIndicator(
          value: current,
          minHeight: height,
          color: color ?? scheme.primary,
          backgroundColor: background ?? scheme.surfaceContainerHigh,
        ),
      ),
    );
  }
}

/// Wadah kartu bergaya Sakura Pagi (permukaan, sudut 18, bayangan lembut).
class SakuraCard extends StatelessWidget {
  const SakuraCard({
    required this.child,
    this.padding = const EdgeInsets.all(14),
    this.radius = 18,
    this.color,
    this.gradient,
    this.border = true,
    super.key,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final double radius;
  final Color? color;
  final Gradient? gradient;
  final bool border;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final sakura = SakuraColors.of(context);
    final dark = theme.brightness == Brightness.dark;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: gradient == null
            ? (color ?? (dark ? AppTheme.darkSurface : Colors.white))
            : null,
        gradient: gradient,
        borderRadius: BorderRadius.circular(radius),
        border: border && gradient == null
            ? Border.all(color: theme.colorScheme.outlineVariant, width: 0.8)
            : null,
        boxShadow: [
          BoxShadow(
            color: sakura.shadow.withValues(alpha: dark ? 0.35 : 0.08),
            blurRadius: 24,
            spreadRadius: -12,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Padding(padding: padding, child: child),
    );
  }
}

/// Kotak ikon kanji untuk kategori materi (語 文 字 読).
class KanjiGlyphTile extends StatelessWidget {
  const KanjiGlyphTile({
    required this.glyph,
    required this.color,
    this.size = 42,
    super.key,
  });

  final String glyph;
  final Color color;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.13),
        borderRadius: BorderRadius.circular(size * 0.29),
      ),
      child: Text(
        glyph,
        style: AppTypography.japanese(
          fontSize: size * 0.45,
          fontWeight: FontWeight.w700,
          color: color,
          height: 1,
        ),
      ),
    );
  }
}

/// Glyph kanji untuk id kategori materi.
String categoryGlyph(String id) => switch (id) {
  'vocabulary' => '語',
  'grammar' => '文',
  'kanji' => '字',
  'reading' => '読',
  'listening' => '聴',
  'speaking' => '話',
  _ => '学',
};
