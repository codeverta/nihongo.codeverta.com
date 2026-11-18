import 'package:flutter/material.dart';

/// Renders the small Markdown subset the tutor is asked to use: headings,
/// bullet and numbered lists, bold, inline code and rules.
///
/// It is deliberately restricted for safety: there is no HTML, no images and
/// no tappable links. Markdown links render as their label only, so a model
/// tricked into emitting a link cannot send the user anywhere.
class AiMarkdown extends StatelessWidget {
  const AiMarkdown(this.source, {super.key});

  final String source;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final base = theme.textTheme.bodyMedium!.copyWith(height: 1.45);
    final children = <Widget>[];
    for (final block in parseAiMarkdown(source)) {
      final spans = inlineSpans(block.text, base, theme);
      switch (block.kind) {
        case AiBlockKind.heading:
          children.add(
            Padding(
              padding: EdgeInsets.only(
                top: children.isEmpty ? 0 : 10,
                bottom: 4,
              ),
              child: Text.rich(
                TextSpan(children: spans),
                style: base.copyWith(
                  fontWeight: FontWeight.w700,
                  fontSize: (base.fontSize ?? 14) + (block.level == 1 ? 4 : 2),
                ),
              ),
            ),
          );
        case AiBlockKind.bullet:
        case AiBlockKind.numbered:
          children.add(
            Padding(
              padding: EdgeInsets.only(left: 4.0 + block.level * 14, bottom: 2),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: 22,
                    child: Text(
                      block.kind == AiBlockKind.bullet
                          ? '•'
                          : '${block.number}.',
                      style: base,
                    ),
                  ),
                  Expanded(
                    child: Text.rich(TextSpan(children: spans), style: base),
                  ),
                ],
              ),
            ),
          );
        case AiBlockKind.rule:
          children.add(const Divider(height: 16));
        case AiBlockKind.spacer:
          if (children.isNotEmpty) children.add(const SizedBox(height: 6));
        case AiBlockKind.paragraph:
          children.add(
            Padding(
              padding: const EdgeInsets.only(bottom: 2),
              child: Text.rich(TextSpan(children: spans), style: base),
            ),
          );
      }
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: children,
    );
  }
}

enum AiBlockKind { heading, bullet, numbered, rule, spacer, paragraph }

class AiBlock {
  const AiBlock(this.kind, this.text, {this.level = 0, this.number = 0});
  final AiBlockKind kind;
  final String text;
  final int level;
  final int number;
}

final _heading = RegExp(r'^(#{1,6})\s+(.*)$');
final _bullet = RegExp(r'^(\s*)[-*•]\s+(.*)$');
final _numbered = RegExp(r'^(\s*)(\d{1,3})[.)]\s+(.*)$');
final _rule = RegExp(r'^\s*([-*_])(\s*\1){2,}\s*$');
final _link = RegExp(r'!?\[([^\]]*)\]\([^)]*\)');
final _tag = RegExp('<[^>]{1,200}>');

/// Splits Markdown into line-level blocks. Exposed for tests.
List<AiBlock> parseAiMarkdown(String source) {
  final blocks = <AiBlock>[];
  for (var line in source.replaceAll('\r\n', '\n').split('\n')) {
    // Never interpret HTML; drop tags and keep image/link labels only.
    line = line
        .replaceAllMapped(_link, (m) => m.group(1) ?? '')
        .replaceAll(_tag, '');
    if (line.trim().isEmpty) {
      if (blocks.isEmpty || blocks.last.kind != AiBlockKind.spacer) {
        blocks.add(const AiBlock(AiBlockKind.spacer, ''));
      }
      continue;
    }
    if (_rule.hasMatch(line)) {
      blocks.add(const AiBlock(AiBlockKind.rule, ''));
      continue;
    }
    final heading = _heading.firstMatch(line);
    if (heading != null) {
      blocks.add(
        AiBlock(
          AiBlockKind.heading,
          heading.group(2)!.trim(),
          level: heading.group(1)!.length,
        ),
      );
      continue;
    }
    final bullet = _bullet.firstMatch(line);
    if (bullet != null) {
      blocks.add(
        AiBlock(
          AiBlockKind.bullet,
          bullet.group(2)!.trim(),
          level: bullet.group(1)!.length ~/ 2,
        ),
      );
      continue;
    }
    final numbered = _numbered.firstMatch(line);
    if (numbered != null) {
      blocks.add(
        AiBlock(
          AiBlockKind.numbered,
          numbered.group(3)!.trim(),
          level: numbered.group(1)!.length ~/ 2,
          number: int.parse(numbered.group(2)!),
        ),
      );
      continue;
    }
    blocks.add(AiBlock(AiBlockKind.paragraph, line.trim()));
  }
  while (blocks.isNotEmpty && blocks.last.kind == AiBlockKind.spacer) {
    blocks.removeLast();
  }
  return blocks;
}

final _inline = RegExp(r'\*\*(.+?)\*\*|__(.+?)__|`([^`]+)`');

/// Bold (`**x**`, `__x__`) and inline code. Everything else is plain text.
List<InlineSpan> inlineSpans(String text, TextStyle base, ThemeData theme) {
  final spans = <InlineSpan>[];
  var index = 0;
  for (final match in _inline.allMatches(text)) {
    if (match.start > index) {
      spans.add(TextSpan(text: text.substring(index, match.start)));
    }
    final bold = match.group(1) ?? match.group(2);
    if (bold != null) {
      spans.add(
        TextSpan(
          text: bold,
          style: const TextStyle(fontWeight: FontWeight.w700),
        ),
      );
    } else {
      spans.add(
        TextSpan(
          text: match.group(3),
          style: TextStyle(
            fontFamily: 'monospace',
            backgroundColor: theme.colorScheme.surfaceContainerHighest,
          ),
        ),
      );
    }
    index = match.end;
  }
  if (index < text.length) spans.add(TextSpan(text: text.substring(index)));
  return spans;
}
