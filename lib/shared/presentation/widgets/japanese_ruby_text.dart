import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';

/// Displays Japanese [text] with its complete kana [reading] above kanji.
///
/// Kana already present in [text] is kept at the text baseline and used as an
/// anchor to divide the reading among adjacent kanji groups.
class JapaneseRubyText extends StatelessWidget {
  const JapaneseRubyText({
    required this.text,
    required this.reading,
    this.style,
    this.rubyStyle,
    this.textAlign = TextAlign.start,
    super.key,
  });

  final String text;
  final String reading;
  final TextStyle? style;
  final TextStyle? rubyStyle;
  final TextAlign textAlign;

  @override
  Widget build(BuildContext context) {
    final baseStyle = DefaultTextStyle.of(context).style.merge(style);
    final resolvedRubyStyle =
        (rubyStyle ??
                baseStyle.copyWith(
                  fontSize: (baseStyle.fontSize ?? 16) * .48,
                  height: 1,
                  fontWeight: FontWeight.normal,
                ))
            .copyWith(color: rubyStyle?.color ?? baseStyle.color);
    final segments = alignJapaneseRuby(text, reading);
    return Text.rich(
      TextSpan(
        style: baseStyle,
        children: [
          for (final segment in segments)
            if (segment.ruby.isEmpty)
              TextSpan(text: segment.base)
            else
              WidgetSpan(
                alignment: PlaceholderAlignment.baseline,
                baseline: TextBaseline.alphabetic,
                child: _RubyAnnotation(
                  base: segment.base,
                  ruby: segment.ruby,
                  baseStyle: baseStyle,
                  rubyStyle: resolvedRubyStyle,
                ),
              ),
        ],
      ),
      textAlign: textAlign,
    );
  }
}

class _RubyAnnotation extends StatelessWidget {
  const _RubyAnnotation({
    required this.base,
    required this.ruby,
    required this.baseStyle,
    required this.rubyStyle,
  });

  final String base;
  final String ruby;
  final TextStyle baseStyle;
  final TextStyle rubyStyle;

  @override
  Widget build(BuildContext context) {
    final scaler = MediaQuery.textScalerOf(context);
    final direction = Directionality.of(context);
    final basePainter = TextPainter(
      text: TextSpan(text: base, style: baseStyle),
      textDirection: direction,
      textScaler: scaler,
      locale: Localizations.maybeLocaleOf(context),
    )..layout();
    final rubyPainter = TextPainter(
      text: TextSpan(text: ruby, style: rubyStyle),
      textDirection: direction,
      textScaler: scaler,
      locale: Localizations.maybeLocaleOf(context),
    )..layout();
    final baseBaseline = basePainter.computeDistanceToActualBaseline(
      TextBaseline.alphabetic,
    );
    final rubyHeight = rubyPainter.height;
    final width = basePainter.width > rubyPainter.width
        ? basePainter.width
        : rubyPainter.width;
    basePainter.dispose();
    rubyPainter.dispose();

    return _RubyBaseline(
      baseline: rubyHeight + baseBaseline,
      child: Padding(
        padding: EdgeInsets.only(top: rubyHeight),
        child: SizedBox(
          width: width,
          child: Stack(
            alignment: Alignment.topCenter,
            clipBehavior: Clip.none,
            children: [
              Text(base, style: baseStyle, softWrap: false),
              Positioned(
                top: -rubyHeight,
                left: 0,
                right: 0,
                child: Center(
                  child: Text(ruby, style: rubyStyle, softWrap: false),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _RubyBaseline extends SingleChildRenderObjectWidget {
  const _RubyBaseline({required this.baseline, required super.child});

  final double baseline;

  @override
  RenderObject createRenderObject(BuildContext context) =>
      _RenderRubyBaseline(baseline);

  @override
  void updateRenderObject(
    BuildContext context,
    covariant _RenderRubyBaseline renderObject,
  ) {
    renderObject.baseline = baseline;
  }
}

class _RenderRubyBaseline extends RenderProxyBox {
  _RenderRubyBaseline(double baseline) : _baseline = baseline;

  double _baseline;

  double get baseline => _baseline;

  set baseline(double value) {
    if (value == _baseline) return;
    _baseline = value;
    markNeedsLayout();
  }

  @override
  double? computeDistanceToActualBaseline(TextBaseline textBaseline) =>
      _baseline;
}

/// A base-text run and the kana reading aligned above it, if any.
class JapaneseRubySegment {
  const JapaneseRubySegment(this.base, this.ruby);

  final String base;
  final String ruby;
}

/// Aligns a complete kana reading to kanji groups, using written kana as
/// anchors. For example, `食べる` + `たべる` becomes ruby `た` over `食` and
/// leaves `べる` at the baseline. Unmatched kanji groups are left unannotated.
List<JapaneseRubySegment> alignJapaneseRuby(String text, String reading) {
  final runes = text.runes.toList();
  if (runes.isEmpty) return const [];
  final groups = <({String base, _JapaneseRun kind})>[];
  var current = String.fromCharCode(runes.first);
  var kind = _kindOf(runes.first);
  for (final rune in runes.skip(1)) {
    final nextKind = _kindOf(rune);
    if (nextKind == kind) {
      current += String.fromCharCode(rune);
    } else {
      groups.add((base: current, kind: kind));
      current = String.fromCharCode(rune);
      kind = nextKind;
    }
  }
  groups.add((base: current, kind: kind));

  final rubyByGroup = List<String>.filled(groups.length, '');
  final readingRunes = _normalizedReading(reading).runes.toList();
  var cursor = 0;
  final alignedGroups = <int>{};
  final anchorPositions = List<int?>.filled(groups.length, null);
  var anchorSearchBefore = readingRunes.length;
  for (var i = groups.length - 1; i >= 0; i--) {
    final group = groups[i];
    if (group.kind != _JapaneseRun.kana) continue;
    final anchor = _toHiragana(group.base).runes.toList();
    final at = _lastIndexOf(readingRunes, anchor, anchorSearchBefore);
    if (at < 0) continue;
    anchorPositions[i] = at;
    anchorSearchBefore = at;
  }
  for (var i = 0; i < groups.length; i++) {
    final group = groups[i];
    if (group.kind != _JapaneseRun.kana) continue;
    final at = anchorPositions[i];
    if (at == null) continue;
    if (i > 0 && groups[i - 1].kind == _JapaneseRun.kanji) {
      rubyByGroup[i - 1] = String.fromCharCodes(
        readingRunes.sublist(cursor, at),
      );
      alignedGroups.add(i - 1);
    }
    alignedGroups.add(i);
    cursor = at + group.base.runes.length;
  }
  if (cursor < readingRunes.length && groups.isNotEmpty) {
    final significantGroups = groups
        .asMap()
        .entries
        .where((entry) => entry.value.kind != _JapaneseRun.punctuation)
        .toList();
    final lastKanji =
        significantGroups.isNotEmpty &&
            significantGroups.last.value.kind == _JapaneseRun.kanji
        ? significantGroups.last.key
        : -1;
    if (lastKanji >= 0 && !alignedGroups.contains(lastKanji)) {
      rubyByGroup[lastKanji] = String.fromCharCodes(
        readingRunes.sublist(cursor),
      );
      alignedGroups.add(lastKanji);
    }
  }
  if (alignedGroups.isEmpty) {
    final kanjiRuns = groups
        .where((group) => group.kind == _JapaneseRun.kanji)
        .length;
    if (kanjiRuns == 1 &&
        groups.every(
          (group) =>
              group.kind == _JapaneseRun.kanji ||
              group.kind == _JapaneseRun.punctuation,
        )) {
      return [
        for (var i = 0; i < groups.length; i++)
          JapaneseRubySegment(
            groups[i].base,
            groups[i].kind == _JapaneseRun.kanji
                ? String.fromCharCodes(readingRunes)
                : '',
          ),
      ];
    }
    return [JapaneseRubySegment(text, '')];
  }
  return [
    for (var i = 0; i < groups.length; i++)
      JapaneseRubySegment(groups[i].base, rubyByGroup[i]),
  ];
}

int _lastIndexOf(List<int> source, List<int> target, int before) {
  if (target.isEmpty) return before;
  var lastPossibleStart = before - target.length;
  if (lastPossibleStart > source.length - target.length) {
    lastPossibleStart = source.length - target.length;
  }
  for (var i = lastPossibleStart; i >= 0; i--) {
    var matches = true;
    for (var j = 0; j < target.length; j++) {
      if (source[i + j] != target[j]) {
        matches = false;
        break;
      }
    }
    if (matches) return i;
  }
  return -1;
}

enum _JapaneseRun { kanji, kana, punctuation, plain }

_JapaneseRun _kindOf(int rune) {
  if (_isKana(rune)) return _JapaneseRun.kana;
  if (_isKanji(rune)) return _JapaneseRun.kanji;
  if (_isPunctuation(rune) || _isWhitespace(rune)) {
    return _JapaneseRun.punctuation;
  }
  return _JapaneseRun.plain;
}

bool _isKana(int rune) =>
    (rune >= 0x3040 && rune <= 0x30ff) || (rune >= 0xff66 && rune <= 0xff9d);

bool _isKanji(int rune) =>
    (rune >= 0x3400 && rune <= 0x4dbf) ||
    (rune >= 0x4e00 && rune <= 0x9fff) ||
    (rune >= 0xf900 && rune <= 0xfaff) ||
    (rune >= 0x20000 && rune <= 0x3134f) ||
    rune == 0x3005;

bool _isWhitespace(int rune) => String.fromCharCode(rune).trim().isEmpty;

bool _isPunctuation(int rune) =>
    (rune >= 0x3000 && rune <= 0x303f) ||
    (rune >= 0xff01 && rune <= 0xff0f) ||
    (rune >= 0xff1a && rune <= 0xff20) ||
    (rune >= 0xff3b && rune <= 0xff40) ||
    (rune >= 0xff5b && rune <= 0xff65) ||
    ',.!?;:()[]{}\u2010\u2014\u2026'.runes.contains(rune);

String _normalizedReading(String value) =>
    _toHiragana(value).replaceAll(RegExp(r'[\s。、，．！？・「」『』（）()「」]'), '');

String _toHiragana(String value) => String.fromCharCodes(
  value.runes.map(
    (rune) => rune >= 0x30a1 && rune <= 0x30f6 ? rune - 0x60 : rune,
  ),
);
