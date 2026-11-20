import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/physics.dart';
import 'package:flutter/services.dart';

/// Which way a study card left the screen.
enum AnkiSwipeDirection { left, right }

/// A Tinder-style card stack for one flashcard.
///
/// The top card follows the finger horizontally (translation plus a slight
/// rotation). Released past [distanceThreshold] of the width or faster than
/// [velocityThreshold], it is thrown off screen and [onSwiped] runs; otherwise
/// it springs back to the centre. [onSwiped] returns whether the card was
/// accepted (the session moved on); when it was not, for example because the
/// review failed, the card comes back so the learner can retry.
///
/// Horizontal drags only, so vertical drags keep scrolling the page and a tap
/// (which never moves past the drag slop) always reaches [onTap].
class AnkiSwipeCardStack extends StatefulWidget {
  const AnkiSwipeCardStack({
    required this.child,
    required this.swipeEnabled,
    required this.onSwiped,
    required this.rightLabel,
    required this.leftLabel,
    this.next,
    this.onTap,
    this.distanceThreshold = .3,
    this.velocityThreshold = 900,
    super.key,
  });

  /// Content of the card on top.
  final Widget child;

  /// Preview of the following card, drawn behind the top card.
  final Widget? next;
  final bool swipeEnabled;
  final VoidCallback? onTap;
  final Future<bool> Function(AnkiSwipeDirection direction) onSwiped;
  final String rightLabel;
  final String leftLabel;

  /// Fraction of the card width that counts as a decisive swipe.
  final double distanceThreshold;

  /// Fling speed in logical pixels per second that counts as a swipe.
  final double velocityThreshold;

  @override
  State<AnkiSwipeCardStack> createState() => _AnkiSwipeCardStackState();
}

class _AnkiSwipeCardStackState extends State<AnkiSwipeCardStack>
    with SingleTickerProviderStateMixin {
  // Horizontal offset of the top card in logical pixels; unbounded so it can
  // follow the finger, spring past zero and fly beyond the screen edge.
  late final AnimationController _offset = AnimationController.unbounded(
    vsync: this,
  );
  double _width = 1;
  bool _pastThreshold = false;
  bool _dismissing = false;

  bool get _reduceMotion => MediaQuery.disableAnimationsOf(context);
  bool get _canDrag => widget.swipeEnabled && !_dismissing;

  @override
  void didUpdateWidget(AnkiSwipeCardStack oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Swipe switched off mid-drag (e.g. the card flipped back): settle.
    if (!widget.swipeEnabled && !_dismissing && _offset.value != 0) {
      _springBack(0);
    }
  }

  @override
  void dispose() {
    _offset.dispose();
    super.dispose();
  }

  void _onDragStart(DragStartDetails details) {
    if (!_canDrag) return;
    _offset.stop();
  }

  void _onDragUpdate(DragUpdateDetails details) {
    if (!_canDrag) return;
    _offset.value += details.primaryDelta ?? details.delta.dx;
    final past = _offset.value.abs() >= _width * widget.distanceThreshold;
    if (past != _pastThreshold) {
      _pastThreshold = past;
      if (past) unawaited(HapticFeedback.lightImpact());
    }
  }

  void _onDragEnd(DragEndDetails details) {
    if (!_canDrag) return;
    final velocity =
        details.primaryVelocity ?? details.velocity.pixelsPerSecond.dx;
    final dx = _offset.value;
    final farEnough = dx.abs() >= _width * widget.distanceThreshold;
    // A fling counts only when it goes the same way the card already leans,
    // so a quick flick back towards the centre cancels the swipe.
    final fastEnough =
        velocity.abs() >= widget.velocityThreshold &&
        (dx == 0 || velocity.sign == dx.sign) &&
        dx.abs() >= 12;
    _pastThreshold = false;
    if (farEnough || fastEnough) {
      final direction = (fastEnough ? velocity : dx) > 0
          ? AnkiSwipeDirection.right
          : AnkiSwipeDirection.left;
      unawaited(_dismiss(direction, velocity));
    } else {
      _springBack(velocity);
    }
  }

  void _onDragCancel() {
    if (!_canDrag) return;
    _pastThreshold = false;
    _springBack(0);
  }

  void _springBack(double velocity) {
    if (_reduceMotion) {
      _offset.value = 0;
      return;
    }
    const spring = SpringDescription(mass: 1, stiffness: 420, damping: 26);
    unawaited(
      _offset.animateWith(SpringSimulation(spring, _offset.value, 0, velocity)),
    );
  }

  Future<void> _dismiss(AnkiSwipeDirection direction, double velocity) async {
    _dismissing = true;
    final sign = direction == AnkiSwipeDirection.right ? 1.0 : -1.0;
    final target = sign * _width * 1.5;
    if (_reduceMotion) {
      _offset.value = target;
    } else {
      // Faster flings leave sooner; keep it between 140 and 280 ms.
      final remaining = (target - _offset.value).abs();
      final speed = math.max(velocity.abs(), 1600);
      final ms = (remaining / speed * 1000).clamp(140, 280).round();
      await _offset.animateTo(
        target,
        duration: Duration(milliseconds: ms),
        curve: Curves.easeOutCubic,
      );
    }
    if (!mounted) return;
    final accepted = await widget.onSwiped(direction);
    if (!mounted) return;
    _dismissing = false;
    if (!accepted) _springBack(0);
  }

  @override
  Widget build(BuildContext context) {
    final reduceMotion = _reduceMotion;
    final top = RepaintBoundary(child: widget.child);
    return LayoutBuilder(
      builder: (context, constraints) {
        _width = constraints.maxWidth.isFinite && constraints.maxWidth > 0
            ? constraints.maxWidth
            : MediaQuery.sizeOf(context).width;
        return Stack(
          clipBehavior: Clip.none,
          children: [
            if (widget.next != null)
              Positioned.fill(
                child: IgnorePointer(
                  child: AnimatedBuilder(
                    animation: _offset,
                    // The next card rises as the top card moves away.
                    builder: (context, child) {
                      final t = (_offset.value.abs() / (_width * .6)).clamp(
                        0.0,
                        1.0,
                      );
                      return Transform.translate(
                        offset: Offset(0, 14 * (1 - t)),
                        child: Transform.scale(
                          scale: .94 + .06 * t,
                          alignment: Alignment.topCenter,
                          child: Opacity(opacity: .55 + .45 * t, child: child),
                        ),
                      );
                    },
                    child: _NextCardFadeIn(
                      reduceMotion: reduceMotion,
                      child: widget.next!,
                    ),
                  ),
                ),
              ),
            GestureDetector(
              key: const Key('anki-swipe-card'),
              behavior: HitTestBehavior.opaque,
              onTap: widget.onTap,
              onHorizontalDragStart: widget.swipeEnabled ? _onDragStart : null,
              onHorizontalDragUpdate: widget.swipeEnabled
                  ? _onDragUpdate
                  : null,
              onHorizontalDragEnd: widget.swipeEnabled ? _onDragEnd : null,
              onHorizontalDragCancel: widget.swipeEnabled
                  ? _onDragCancel
                  : null,
              child: AnimatedBuilder(
                animation: _offset,
                builder: (context, child) {
                  final dx = _offset.value;
                  final progress =
                      (dx.abs() / (_width * widget.distanceThreshold)).clamp(
                        0.0,
                        1.0,
                      );
                  final angle = reduceMotion ? 0.0 : dx / _width * .22;
                  return Transform.translate(
                    offset: Offset(dx, dx.abs() * .04),
                    child: Transform.rotate(
                      angle: angle,
                      alignment: Alignment.bottomCenter,
                      child: Stack(
                        children: [
                          child!,
                          if (dx != 0)
                            Positioned.fill(
                              child: IgnorePointer(
                                child: _SwipeOverlay(
                                  right: dx > 0,
                                  progress: progress,
                                  label: dx > 0
                                      ? widget.rightLabel
                                      : widget.leftLabel,
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  );
                },
                child: top,
              ),
            ),
          ],
        );
      },
    );
  }
}

class _NextCardFadeIn extends StatelessWidget {
  const _NextCardFadeIn({required this.reduceMotion, required this.child});
  final bool reduceMotion;
  final Widget child;

  @override
  Widget build(BuildContext context) => TweenAnimationBuilder<double>(
    tween: Tween(begin: reduceMotion ? 1 : 0, end: 1),
    duration: reduceMotion ? Duration.zero : const Duration(milliseconds: 260),
    curve: Curves.easeOut,
    builder: (context, value, child) => Opacity(opacity: value, child: child),
    child: child,
  );
}

class _SwipeOverlay extends StatelessWidget {
  const _SwipeOverlay({
    required this.right,
    required this.progress,
    required this.label,
  });
  final bool right;
  final double progress;
  final String label;

  @override
  Widget build(BuildContext context) {
    final color = right ? const Color(0xFF2E9E5B) : const Color(0xFFD94848);
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        color: color.withValues(alpha: .08 * progress),
        border: Border.all(
          color: color.withValues(alpha: .7 * progress),
          width: 3,
        ),
      ),
      child: Align(
        alignment: right ? Alignment.topLeft : Alignment.topRight,
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Opacity(
            opacity: progress,
            child: Transform.rotate(
              angle: right ? -.2 : .2,
              child: Container(
                key: Key(right ? 'anki-swipe-right' : 'anki-swipe-left'),
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.surface,
                  border: Border.all(color: color, width: 3),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  label.toUpperCase(),
                  style: TextStyle(
                    color: color,
                    fontWeight: FontWeight.w900,
                    fontSize: 20,
                    letterSpacing: 1.2,
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
