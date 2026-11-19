import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// Animates a soft shimmer across every [SkeletonBox] below it. Uses one
/// controller for the whole group so all placeholders move together, and
/// stays still when the user turned animations off in system settings.
class SkeletonShimmer extends StatefulWidget {
  const SkeletonShimmer({required this.child, super.key});

  final Widget child;

  @override
  State<SkeletonShimmer> createState() => _SkeletonShimmerState();
}

class _SkeletonShimmerState extends State<SkeletonShimmer>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 1300),
  );

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (MediaQuery.disableAnimationsOf(context)) {
      _controller.stop();
    } else if (!_controller.isAnimating) {
      _controller.repeat();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final highlight = scheme.surface.withValues(alpha: 0.55);
    return Semantics(
      // Announce one loading state instead of many empty shapes.
      label: 'Memuat…'.localized(context),
      excludeSemantics: true,
      child: AnimatedBuilder(
        animation: _controller,
        child: widget.child,
        builder: (context, child) {
          if (!_controller.isAnimating) return child!;
          final t = _controller.value * 2 - 0.5;
          return ShaderMask(
            blendMode: BlendMode.srcATop,
            shaderCallback: (bounds) => LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [Colors.transparent, highlight, Colors.transparent],
              stops: [
                (t - 0.25).clamp(0.0, 1.0),
                t.clamp(0.0, 1.0),
                (t + 0.25).clamp(0.0, 1.0),
              ],
            ).createShader(bounds),
            child: child,
          );
        },
      ),
    );
  }
}

/// A rounded placeholder block. Width defaults to the available width.
class SkeletonBox extends StatelessWidget {
  const SkeletonBox({
    this.width = double.infinity,
    this.height = 14,
    this.radius = 8,
    super.key,
  });

  /// A circular placeholder, e.g. for an avatar.
  const SkeletonBox.circle({required double size, super.key})
    : width = size,
      height = size,
      radius = size / 2;

  final double width;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: scheme.onSurface.withValues(alpha: 0.07),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
  }
}
