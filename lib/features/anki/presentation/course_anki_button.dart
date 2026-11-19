import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// A built-in Anki deck that drills the same material as a course.
class CourseAnkiDeck {
  const CourseAnkiDeck({
    required this.id,
    required this.title,
    required this.subtitle,
  });

  final String id;

  /// Deck name without the level, as Indonesian source copy; localized with
  /// `.localized(context)`.
  final String title;
  final String subtitle;
}

/// The built-in decks matching a course (`kanji` or `vocabulary`) and JLPT
/// level. Grammar and reading have no matching deck. IDs come from the
/// backend Anki catalog (backend/internal/anki/resources/catalog.json).
List<CourseAnkiDeck> courseAnkiDecks(String type, String level) {
  final n = level.toLowerCase();
  if (!const ['n5', 'n4', 'n3', 'n2', 'n1'].contains(n)) return const [];
  switch (type) {
    case 'kanji':
      return [
        CourseAnkiDeck(
          id: 'k_$n',
          title: 'Penggunaan Kanji',
          subtitle: 'Latih bacaan dan arti kanji dengan kartu',
        ),
      ];
    case 'vocabulary':
      return [
        CourseAnkiDeck(
          id: n,
          title: 'Bacaan Kosakata',
          subtitle: 'Kartu: kanji → cara baca',
        ),
        // Meaning decks exist only for these levels in the catalog.
        if (const ['n5', 'n4', 'n2'].contains(n))
          CourseAnkiDeck(
            id: '${n}m',
            title: 'Arti Kosakata',
            subtitle: 'Kartu: kata → arti',
          ),
      ];
  }
  return const [];
}

/// App-bar action that opens the Anki deck for the course level, or lets the
/// user pick when several decks match.
class CourseAnkiButton extends StatelessWidget {
  const CourseAnkiButton({required this.type, required this.level, super.key});

  final String type;
  final String level;

  Future<void> _open(BuildContext context, List<CourseAnkiDeck> decks) async {
    if (decks.length == 1) {
      await context.push('/anki/decks/${decks.single.id}');
      return;
    }
    final picked = await showModalBottomSheet<CourseAnkiDeck>(
      context: context,
      showDragHandle: true,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 0, 24, 8),
              child: Text(
                'Pilih deck Anki'.localized(context),
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
            for (final deck in decks)
              ListTile(
                key: ValueKey('course-anki-deck-${deck.id}'),
                leading: const Icon(Icons.style_outlined),
                title: Text('$level · ${deck.title.localized(context)}'),
                subtitle: Text(deck.subtitle.localized(context)),
                trailing: const Icon(Icons.chevron_right_rounded),
                onTap: () => Navigator.pop(context, deck),
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
    if (picked != null && context.mounted) {
      await context.push('/anki/decks/${picked.id}');
    }
  }

  @override
  Widget build(BuildContext context) {
    final decks = courseAnkiDecks(type, level);
    if (decks.isEmpty) return const SizedBox.shrink();
    return IconButton(
      key: const ValueKey('course-anki-button'),
      tooltip: 'Latihan kartu Anki'.localized(context),
      icon: const Icon(Icons.style_outlined),
      onPressed: () => _open(context, decks),
    );
  }
}
