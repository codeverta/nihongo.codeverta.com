import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';

/// Ruby is painted over actual native selection boxes. The selectable string
/// contains only source text, so copying and UTF-16 selection offsets stay intact.
class FuriganaText extends StatefulWidget {
  const FuriganaText({
    required this.text,
    required this.tokens,
    required this.showFurigana,
    required this.onLookup,
    this.style,
    super.key,
  });
  final String text;
  final List<NewsToken> tokens;
  final bool showFurigana;
  final TextStyle? style;
  final ValueChanged<List<NewsToken>> onLookup;
  @override
  State<FuriganaText> createState() => _FuriganaTextState();
}

class _FuriganaTextState extends State<FuriganaText> {
  final _textKey = GlobalKey();
  final _paintKey = GlobalKey();

  @override
  Widget build(BuildContext context) {
    final style = (Theme.of(context).textTheme.bodyLarge ?? const TextStyle())
        .merge(widget.style)
        .copyWith(height: widget.showFurigana ? 2.3 : 1.8);
    final size = style.fontSize ?? 20;
    final rubyStyle = style.copyWith(
      fontSize: size * .48,
      height: 1,
      fontWeight: FontWeight.normal,
      letterSpacing: 0,
    );
    final scaler = MediaQuery.textScalerOf(context);
    final spans = <TextSpan>[];
    var offset = 0;
    for (final token in widget.tokens) {
      if (token.start < offset ||
          token.end > widget.text.length ||
          token.start >= token.end) {
        continue;
      }
      if (token.start > offset) {
        spans.add(TextSpan(text: widget.text.substring(offset, token.start)));
      }
      var tokenStyle = style;
      if (widget.showFurigana && token.reading.isNotEmpty) {
        final base = TextPainter(
          text: TextSpan(text: token.surface, style: style),
          textDirection: TextDirection.ltr,
          textScaler: scaler,
        )..layout();
        final ruby = TextPainter(
          text: TextSpan(text: token.reading, style: rubyStyle),
          textDirection: TextDirection.ltr,
          textScaler: scaler,
        )..layout();
        final extra = math.max(0, ruby.width - base.width);
        if (extra > 0) {
          tokenStyle = style.copyWith(
            letterSpacing:
                (style.letterSpacing ?? 0) +
                extra / math.max(1, token.surface.runes.length),
          );
        }
        base.dispose();
        ruby.dispose();
      }
      spans.add(
        TextSpan(
          text: widget.text.substring(token.start, token.end),
          style: tokenStyle,
        ),
      );
      offset = token.end;
    }
    if (offset < widget.text.length) {
      spans.add(TextSpan(text: widget.text.substring(offset)));
    }
    return CustomPaint(
      key: _paintKey,
      foregroundPainter: widget.showFurigana
          ? _RubyPainter(
              textKey: _textKey,
              paintKey: _paintKey,
              tokens: widget.tokens,
              style: rubyStyle,
              scaler: scaler,
            )
          : null,
      child: Padding(
        padding: EdgeInsets.only(
          top: widget.showFurigana ? scaler.scale(size * .55) : 0,
        ),
        child: SelectableText.rich(
          TextSpan(style: style, children: spans),
          key: _textKey,
          textDirection: TextDirection.ltr,
          contextMenuBuilder: (context, state) =>
              AdaptiveTextSelectionToolbar.buttonItems(
                anchors: state.contextMenuAnchors,
                buttonItems: [
                  ContextMenuButtonItem(
                    label: 'Cari arti'.localized(context),
                    onPressed: () {
                      final selection = state.textEditingValue.selection;
                      final tokens = newsSelectionTokens(
                        widget.tokens,
                        selection,
                      );
                      state.hideToolbar();
                      if (tokens.isNotEmpty) widget.onLookup(tokens);
                    },
                  ),
                  ...state.contextMenuButtonItems,
                ],
              ),
        ),
      ),
    );
  }
}

List<NewsToken> newsSelectionTokens(
  List<NewsToken> tokens,
  TextSelection selection,
) {
  if (!selection.isValid || selection.isCollapsed) return [];
  final seen = <String>{};
  return tokens
      .where(
        (t) =>
            t.end > selection.start &&
            t.start < selection.end &&
            RegExp(r'[\p{L}\p{N}]', unicode: true).hasMatch(t.surface) &&
            seen.add(t.base),
      )
      .toList();
}

RenderEditable? _editable(RenderObject? object) {
  if (object is RenderEditable) return object;
  RenderEditable? result;
  object?.visitChildren((child) {
    result ??= _editable(child);
  });
  return result;
}

class _RubyPainter extends CustomPainter {
  _RubyPainter({
    required this.textKey,
    required this.paintKey,
    required this.tokens,
    required this.style,
    required this.scaler,
  });
  final GlobalKey textKey;
  final GlobalKey paintKey;
  final List<NewsToken> tokens;
  final TextStyle style;
  final TextScaler scaler;
  @override
  void paint(Canvas canvas, Size size) {
    final editable = _editable(textKey.currentContext?.findRenderObject());
    final parent = paintKey.currentContext?.findRenderObject();
    if (editable == null || parent is! RenderBox || !editable.hasSize) return;
    final origin = editable.localToGlobal(Offset.zero, ancestor: parent);
    for (final token in tokens) {
      if (token.reading.isEmpty) continue;
      final boxes = editable.getBoxesForSelection(
        TextSelection(baseOffset: token.start, extentOffset: token.end),
      );
      final total = boxes.fold<double>(0, (n, b) => n + b.right - b.left);
      if (total <= 0) continue;
      final reading = token.reading.runes.toList();
      var cursor = 0;
      var accumulated = 0.0;
      for (var i = 0; i < boxes.length; i++) {
        final box = boxes[i];
        accumulated += box.right - box.left;
        final end = i == boxes.length - 1
            ? reading.length
            : (reading.length * accumulated / total).round();
        if (end <= cursor) continue;
        final painter = TextPainter(
          text: TextSpan(
            text: String.fromCharCodes(reading.sublist(cursor, end)),
            style: style,
          ),
          textDirection: TextDirection.ltr,
          textScaler: scaler,
          maxLines: 1,
        )..layout();
        final width = box.right - box.left;
        final scale = math.min<double>(1, width / math.max(1, painter.width));
        final x = origin.dx + box.left + (width - painter.width * scale) / 2;
        final y = math.max<double>(0, origin.dy + box.top - painter.height);
        canvas
          ..save()
          ..translate(x, y)
          ..scale(scale, 1);
        painter.paint(canvas, Offset.zero);
        canvas.restore();
        painter.dispose();
        cursor = end;
      }
    }
  }

  @override
  bool shouldRepaint(covariant _RubyPainter oldDelegate) => true;
}
