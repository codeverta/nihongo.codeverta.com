import 'package:flutter/material.dart';

/// Marks the tested word without revealing its reading or translation.
class SimulationText extends StatelessWidget {
  const SimulationText(this.text, {this.style, super.key});
  final String text;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) {
    final spans = <TextSpan>[];
    var offset = 0;
    for (final match in RegExp(r'\[\[(.*?)\]\]').allMatches(text)) {
      spans
        ..add(TextSpan(text: text.substring(offset, match.start)))
        ..add(
          TextSpan(
            text: match.group(1),
            style: const TextStyle(
              decoration: TextDecoration.underline,
              decorationThickness: 2,
              fontWeight: FontWeight.w800,
            ),
          ),
        );
      offset = match.end;
    }
    spans.add(TextSpan(text: text.substring(offset)));
    return Text.rich(TextSpan(children: spans), style: style);
  }
}
