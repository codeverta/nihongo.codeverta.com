import 'package:flutter/material.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

/// Adds ruby readings to Japanese vocabulary and example text known by a lesson.
///
/// The quiz API sends prompts and choices as plain strings. This widget uses
/// the lesson's word bank to annotate matching Japanese terms and examples
/// without requiring a separate reading for Indonesian text.
class VocabularyRubyText extends StatelessWidget {
  const VocabularyRubyText({
    required this.text,
    required this.words,
    required this.enabled,
    this.style,
    this.rubyStyle,
    this.textAlign = TextAlign.start,
    super.key,
  });

  final String text;
  final List<VocabularyWord> words;
  final bool enabled;
  final TextStyle? style;
  final TextStyle? rubyStyle;
  final TextAlign textAlign;

  @override
  Widget build(BuildContext context) {
    final baseStyle = DefaultTextStyle.of(context).style.merge(style);
    if (!enabled) {
      return Text(text, style: baseStyle, textAlign: textAlign);
    }

    final candidates = _annotations(words);
    if (candidates.isEmpty) {
      return Text(text, style: baseStyle, textAlign: textAlign);
    }

    final spans = <InlineSpan>[];
    final plainText = StringBuffer();
    void flushPlainText() {
      if (plainText.isEmpty) return;
      spans.add(TextSpan(text: plainText.toString()));
      plainText.clear();
    }

    var cursor = 0;
    while (cursor < text.length) {
      final match = _findNextAnnotation(text, cursor, candidates);
      if (match == null) {
        plainText.write(text.substring(cursor));
        break;
      }

      plainText.write(text.substring(cursor, match.start));
      flushPlainText();
      spans.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: JapaneseRubyText(
            text: match.annotation.text,
            reading: match.annotation.reading,
            style: baseStyle,
            rubyStyle: rubyStyle,
          ),
        ),
      );
      cursor = _skipInlineReading(
        text,
        match.start + match.annotation.text.length,
        match.annotation.reading,
      );
    }
    flushPlainText();

    if (spans.isEmpty) {
      return Text(text, style: baseStyle, textAlign: textAlign);
    }
    return Text.rich(
      TextSpan(style: baseStyle, children: spans),
      textAlign: textAlign,
    );
  }
}

List<_RubyAnnotation> _annotations(List<VocabularyWord> words) {
  final unique = <String, _RubyAnnotation>{};
  for (final word in words) {
    _addAnnotation(unique, word.term, word.reading);
    for (final example in word.examples) {
      _addAnnotation(unique, example.japanese, example.reading);
    }
  }
  final result = unique.values.toList()
    ..sort((a, b) => b.text.length.compareTo(a.text.length));
  return result;
}

void _addAnnotation(
  Map<String, _RubyAnnotation> annotations,
  String text,
  String reading,
) {
  if (text.isEmpty || reading.isEmpty || annotations.containsKey(text)) return;
  if (!alignJapaneseRuby(
    text,
    reading,
  ).any((segment) => segment.ruby.isNotEmpty)) {
    return;
  }
  annotations[text] = _RubyAnnotation(text, reading);
}

_AnnotationMatch? _findNextAnnotation(
  String text,
  int from,
  List<_RubyAnnotation> candidates,
) {
  _AnnotationMatch? best;
  for (final annotation in candidates) {
    final start = text.indexOf(annotation.text, from);
    if (start < 0) continue;
    if (best == null ||
        start < best.start ||
        (start == best.start &&
            annotation.text.length > best.annotation.text.length)) {
      best = _AnnotationMatch(start, annotation);
    }
  }
  return best;
}

int _skipInlineReading(String text, int from, String reading) {
  var openingIndex = from;
  if (text.startsWith(' ', openingIndex)) openingIndex++;
  final opening = text.startsWith('(', openingIndex)
      ? '('
      : text.startsWith('（', openingIndex)
      ? '（'
      : null;
  if (opening == null) return from;

  final readingStart = openingIndex + opening.length;
  if (!text.startsWith(reading, readingStart)) return from;
  final closingIndex = readingStart + reading.length;
  final closing = opening == '(' ? ')' : '）';
  if (!text.startsWith(closing, closingIndex)) return from;
  return closingIndex + closing.length;
}

class _RubyAnnotation {
  const _RubyAnnotation(this.text, this.reading);

  final String text;
  final String reading;
}

class _AnnotationMatch {
  const _AnnotationMatch(this.start, this.annotation);

  final int start;
  final _RubyAnnotation annotation;
}
