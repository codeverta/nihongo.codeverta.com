import 'dart:math' as math;
import 'dart:ui';

/// A stroke as a polyline in a unit square (0..1 on both axes).
typedef Stroke = List<Offset>;

enum StrokeVerdict {
  /// Close enough to the expected stroke.
  correct,

  /// Shape is right but drawn from the end towards the start.
  reversed,

  /// Matches a later stroke: the learner skipped ahead.
  wrongOrder,

  /// Starts too far from the expected start point.
  wrongStart,

  /// Does not resemble the expected stroke.
  wrongShape,
}

/// Lenient, beginner-friendly comparison of one drawn stroke against the
/// reference strokes of a character. Coordinates are in the unit square.
class StrokeMatcher {
  const StrokeMatcher({
    this.shapeTolerance = 0.13,
    this.endpointTolerance = 0.2,
  });

  /// Average point distance allowed between drawn and reference stroke.
  final double shapeTolerance;

  /// Distance allowed between start (and end) points.
  final double endpointTolerance;

  static const _samples = 24;

  StrokeVerdict check(Stroke drawn, List<Stroke> reference, int index) {
    final expected = reference[index];
    final verdict = _compare(drawn, expected);
    if (verdict == StrokeVerdict.correct || verdict == StrokeVerdict.reversed) {
      return verdict;
    }
    for (var later = index + 1; later < reference.length; later++) {
      if (_compare(drawn, reference[later]) == StrokeVerdict.correct) {
        return StrokeVerdict.wrongOrder;
      }
    }
    return verdict;
  }

  StrokeVerdict _compare(Stroke drawn, Stroke expected) {
    final a = resample(drawn, _samples);
    final b = resample(expected, _samples);
    final forward = _meanDistance(a, b);
    final backward = _meanDistance(a, b.reversed.toList());
    if (forward <= shapeTolerance &&
        (a.first - b.first).distance <= endpointTolerance &&
        (a.last - b.last).distance <= endpointTolerance) {
      return StrokeVerdict.correct;
    }
    if (backward <= shapeTolerance && backward < forward) {
      return StrokeVerdict.reversed;
    }
    if ((a.first - b.first).distance > endpointTolerance &&
        forward <= shapeTolerance * 1.6) {
      return StrokeVerdict.wrongStart;
    }
    return StrokeVerdict.wrongShape;
  }

  static double _meanDistance(Stroke a, Stroke b) {
    var total = 0.0;
    for (var i = 0; i < a.length; i++) {
      total += (a[i] - b[i]).distance;
    }
    return total / a.length;
  }

  /// Evenly spaced points along [stroke] by arc length.
  static Stroke resample(Stroke stroke, int count) {
    if (stroke.isEmpty) return List.filled(count, Offset.zero);
    final length = strokeLength(stroke);
    if (stroke.length == 1 || length == 0) {
      return List.filled(count, stroke.first);
    }
    final step = length / (count - 1);
    final out = <Offset>[stroke.first];
    var target = step;
    var walked = 0.0;
    for (var i = 1; i < stroke.length && out.length < count; i++) {
      final a = stroke[i - 1];
      final b = stroke[i];
      final segment = (b - a).distance;
      while (segment > 0 &&
          walked + segment >= target &&
          out.length < count - 1) {
        final t = (target - walked) / segment;
        out.add(Offset.lerp(a, b, t)!);
        target += step;
      }
      walked += segment;
    }
    while (out.length < count) {
      out.add(stroke.last);
    }
    return out;
  }

  static double strokeLength(Stroke stroke) {
    var total = 0.0;
    for (var i = 1; i < stroke.length; i++) {
      total += (stroke[i] - stroke[i - 1]).distance;
    }
    return total;
  }

  /// Points of [stroke] up to [progress] (0..1) of its length.
  static Stroke partial(Stroke stroke, double progress) {
    if (progress >= 1 || stroke.length < 2) return stroke;
    final limit = strokeLength(stroke) * math.max(0, progress);
    final out = <Offset>[stroke.first];
    var walked = 0.0;
    for (var i = 1; i < stroke.length; i++) {
      final segment = (stroke[i] - stroke[i - 1]).distance;
      if (walked + segment >= limit) {
        final t = segment == 0 ? 0.0 : (limit - walked) / segment;
        out.add(Offset.lerp(stroke[i - 1], stroke[i], t)!);
        return out;
      }
      out.add(stroke[i]);
      walked += segment;
    }
    return out;
  }
}
