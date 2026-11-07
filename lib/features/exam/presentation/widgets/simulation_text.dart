import 'package:flutter/material.dart';

/// One piece of simulation text: plain text, or a base text with furigana.
class SimulationSegment {
  const SimulationSegment(this.text, {this.ruby, this.underlined = false});
  final String text;
  final String? ruby;
  final bool underlined;

  @override
  bool operator ==(Object other) =>
      other is SimulationSegment &&
      other.text == text &&
      other.ruby == ruby &&
      other.underlined == underlined;

  @override
  int get hashCode => Object.hash(text, ruby, underlined);

  @override
  String toString() => 'SimulationSegment($text, $ruby, $underlined)';
}

final _markup = RegExp(r'\[\[|\]\]|\{([^|{}]+)\|([^{}]+)\}');

/// Parses `[[tested text]]` (underlined) and `{漢字|かんじ}` (furigana).
/// Both may be combined; text outside any markup is returned unchanged.
List<SimulationSegment> parseSimulationText(String text) {
  final segments = <SimulationSegment>[];
  var underlined = false;
  var offset = 0;
  void plain(String value) {
    if (value.isNotEmpty) {
      segments.add(SimulationSegment(value, underlined: underlined));
    }
  }

  for (final match in _markup.allMatches(text)) {
    plain(text.substring(offset, match.start));
    final token = match.group(0)!;
    if (token == '[[') {
      underlined = true;
    } else if (token == ']]') {
      underlined = false;
    } else {
      segments.add(
        SimulationSegment(
          match.group(1)!,
          ruby: match.group(2)!,
          underlined: underlined,
        ),
      );
    }
    offset = match.end;
  }
  plain(text.substring(offset));
  return segments;
}

/// Removes markup, leaving the text a reader would see without furigana.
String stripSimulationMarkup(String text) =>
    parseSimulationText(text).map((s) => s.text).join();

/// Base level of a JLPT level key: `N5-2` -> `N5`. Other keys are unchanged.
String jlptBaseLevel(String level) {
  final match = RegExp(r'^(N[1-5])-\d+$').firstMatch(level);
  return match?.group(1) ?? level;
}

/// Paper number of a JLPT level key: `N5-2` -> 2, `N5` -> 1.
int jlptPaperNumber(String level) {
  final match = RegExp(r'^N[1-5]-(\d+)$').firstMatch(level);
  return match == null ? 1 : int.tryParse(match.group(1)!) ?? 1;
}

/// Two-digit paper label, e.g. `02`.
String jlptPaperLabel(String level) =>
    jlptPaperNumber(level).toString().padLeft(2, '0');

/// Marks the tested word and shows furigana above kanji.
class SimulationText extends StatelessWidget {
  const SimulationText(this.text, {this.style, super.key});
  final String text;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) {
    final base = DefaultTextStyle.of(context).style.merge(style);
    final fontSize = base.fontSize ?? 14;
    const underline = TextStyle(
      decoration: TextDecoration.underline,
      decorationThickness: 2,
      fontWeight: FontWeight.w800,
    );
    final spans = <InlineSpan>[];
    for (final segment in parseSimulationText(text)) {
      final segmentStyle = segment.underlined ? underline : null;
      final ruby = segment.ruby;
      if (ruby == null) {
        spans.add(TextSpan(text: segment.text, style: segmentStyle));
        continue;
      }
      spans.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: Semantics(
            label: '${segment.text} ($ruby)',
            excludeSemantics: true,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              // Reverse order so the column's baseline is the base text's.
              verticalDirection: VerticalDirection.up,
              children: [
                Text(
                  segment.text,
                  style: base.merge(segmentStyle).copyWith(height: 1.15),
                ),
                Text(
                  ruby,
                  style: base
                      .copyWith(
                        fontSize: fontSize * 0.5,
                        height: 1.1,
                        fontWeight: FontWeight.w500,
                        decoration: TextDecoration.none,
                      ),
                ),
              ],
            ),
          ),
        ),
      );
    }
    return Text.rich(TextSpan(children: spans), style: style);
  }
}
