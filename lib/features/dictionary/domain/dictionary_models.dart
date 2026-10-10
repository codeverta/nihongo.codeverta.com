List<String> dictionaryStrings(dynamic value) =>
    value is List ? value.whereType<String>().toList() : const [];

class DictionaryEntry {
  DictionaryEntry(this.json);
  final Map<String, dynamic> json;
  String get id => json['slug'] as String? ?? word;
  bool get common => json['is_common'] == true;
  List<String> get jlpt => dictionaryStrings(json['jlpt'])
      .map((tag) => tag.replaceFirst('jlpt-', '').toUpperCase())
      .toList();
  List<Map<String, dynamic>> get forms => _maps(json['japanese']);
  List<Map<String, dynamic>> get senses => _maps(json['senses']);
  String get word => forms.firstOrNull?['word'] as String? ?? reading;
  String get reading => forms.firstOrNull?['reading'] as String? ?? idFallback;
  String get idFallback => json['slug'] as String? ?? '';
  String get summary => senses.isEmpty
      ? ''
      : dictionaryStrings(senses.first['english_definitions']).join('; ');
  List<String> get kanji => word.runes
      .where(isDictionaryKanji)
      .map(String.fromCharCode)
      .toSet()
      .toList();
}

List<Map<String, dynamic>> _maps(dynamic value) =>
    value is List ? value.whereType<Map<String, dynamic>>().toList() : const [];

bool isDictionaryKanji(int rune) =>
    (rune >= 0x3400 && rune <= 0x4dbf) ||
    (rune >= 0x4e00 && rune <= 0x9fff) ||
    (rune >= 0xf900 && rune <= 0xfaff) ||
    (rune >= 0x20000 && rune <= 0x323af);

class DictionaryPage {
  const DictionaryPage({required this.entries, required this.hasMore});
  factory DictionaryPage.fromJson(Map<String, dynamic> json) => DictionaryPage(
        entries: _maps(json['entries']).map(DictionaryEntry.new).toList(),
        hasMore: json['hasMore'] == true,
      );
  final List<DictionaryEntry> entries;
  final bool hasMore;
}
