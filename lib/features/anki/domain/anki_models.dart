class AnkiCard {
  AnkiCard.fromJson(Map<String, dynamic> json)
      : index = json['index'] as int,
        front = json['front'] as String,
        answers = List<String>.from(json['answers'] as List),
        meaning = json['meaning'] as String? ?? '',
        imageUrl = json['imageUrl'] as String? ?? '',
        version = json['version'] as int? ?? 0,
        intervals = Map<String, String>.from(json['intervals'] as Map? ?? {});
  final int index;
  final String front;
  final List<String> answers;
  final String meaning;
  final String imageUrl;
  final int version;
  final Map<String, String> intervals;
}

class AnkiDeck {
  AnkiDeck.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        title = json['title'] as String,
        category = json['category'] as String,
        level = json['level'] as String? ?? '',
        description = json['description'] as String,
        instructions = json['instructions'] as String,
        originalName = json['originalName'] as String,
        sourceUrl =
            json['sourceUrl'] as String? ?? 'https://github.com/mistval/kotoba',
        originalDescription = json['originalDescription'] as String? ?? '',
        meaningLanguage = json['meaningLanguage'] as String? ?? '',
        cardCount = json['cardCount'] as int,
        sourceCount = json['sourceCount'] as int,
        skippedCount = json['skippedCount'] as int,
        available = json['available'] as bool,
        unavailableReason = json['unavailableReason'] as String? ?? '',
        locked = json['locked'] as bool? ?? false,
        preview = (json['preview'] as List<dynamic>? ?? [])
            .map((c) => AnkiCard.fromJson(c as Map<String, dynamic>))
            .toList(),
        seen = (json['stats'] as Map<String, dynamic>?)?['seen'] as int? ?? 0,
        due = (json['stats'] as Map<String, dynamic>?)?['due'] as int? ?? 0,
        newCards = (json['stats'] as Map<String, dynamic>?)?['new'] as int? ??
            json['cardCount'] as int,
        nextDue = DateTime.tryParse(
          (json['stats'] as Map<String, dynamic>?)?['nextDue'] as String? ?? '',
        );
  final String id;
  final String title;
  final String category;
  final String level;
  final String description;
  final String instructions;
  final String originalName;
  final String sourceUrl;
  final String originalDescription;
  final String meaningLanguage;
  final int cardCount;
  final int sourceCount;
  final int skippedCount;
  final bool available;
  final bool locked;
  final String unavailableReason;
  final List<AnkiCard> preview;
  final int seen;
  final int due;
  final int newCards;
  final DateTime? nextDue;
  bool get premium => const ['N1', 'N2', 'N3'].contains(level);
}

class AnkiSession {
  AnkiSession.fromJson(Map<String, dynamic> json)
      : title = json['title'] as String? ?? 'Belajar flashcard',
        instructions = json['instructions'] as String? ?? '',
        cards = (json['cards'] as List<dynamic>)
            .map((c) => AnkiCard.fromJson(c as Map<String, dynamic>))
            .toList(),
        due = (json['stats'] as Map<String, dynamic>)['due'] as int,
        nextDue = DateTime.tryParse(
          (json['stats'] as Map<String, dynamic>)['nextDue'] as String? ?? '',
        );
  final List<AnkiCard> cards;
  final String title;
  final String instructions;
  final int due;
  final DateTime? nextDue;
}
