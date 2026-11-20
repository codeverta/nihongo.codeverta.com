/// Hiragana & katakana chart data for absolute beginners.
///
/// Katakana is derived from hiragana by Unicode offset, so both scripts share
/// one table of romaji and tips.
library;

enum KanaScript { hiragana, katakana }

enum KanaSection { basic, dakuten, combo }

/// A bilingual beginner tip (Indonesian first, English for the English UI).
typedef KanaTip = ({String id, String en});

class Kana {
  const Kana._(this.char, this.romaji, this.section, this.tips, this.parts);

  /// The character (or two characters for yōon such as きゃ).
  final String char;
  final String romaji;
  final KanaSection section;
  final List<KanaTip> tips;

  /// The single characters a learner writes. One item for normal kana,
  /// two for yōon (き + ゃ).
  final List<String> parts;

  bool get isCombo => parts.length > 1;
}

const _katakanaOffset = 0x60;

String toKatakana(String hiragana) => String.fromCharCodes(
  hiragana.runes.map(
    (r) => r >= 0x3041 && r <= 0x3096 ? r + _katakanaOffset : r,
  ),
);

/// One chart row; `null` marks an empty cell (yi, ye, wu…).
typedef _Row = List<(String, String)?>;

const List<_Row> _basicRows = [
  [('あ', 'a'), ('い', 'i'), ('う', 'u'), ('え', 'e'), ('お', 'o')],
  [('か', 'ka'), ('き', 'ki'), ('く', 'ku'), ('け', 'ke'), ('こ', 'ko')],
  [('さ', 'sa'), ('し', 'shi'), ('す', 'su'), ('せ', 'se'), ('そ', 'so')],
  [('た', 'ta'), ('ち', 'chi'), ('つ', 'tsu'), ('て', 'te'), ('と', 'to')],
  [('な', 'na'), ('に', 'ni'), ('ぬ', 'nu'), ('ね', 'ne'), ('の', 'no')],
  [('は', 'ha'), ('ひ', 'hi'), ('ふ', 'fu'), ('へ', 'he'), ('ほ', 'ho')],
  [('ま', 'ma'), ('み', 'mi'), ('む', 'mu'), ('め', 'me'), ('も', 'mo')],
  [('や', 'ya'), null, ('ゆ', 'yu'), null, ('よ', 'yo')],
  [('ら', 'ra'), ('り', 'ri'), ('る', 'ru'), ('れ', 're'), ('ろ', 'ro')],
  [('わ', 'wa'), null, null, null, ('を', 'wo')],
  [('ん', 'n'), null, null, null, null],
];

const List<_Row> _dakutenRows = [
  [('が', 'ga'), ('ぎ', 'gi'), ('ぐ', 'gu'), ('げ', 'ge'), ('ご', 'go')],
  [('ざ', 'za'), ('じ', 'ji'), ('ず', 'zu'), ('ぜ', 'ze'), ('ぞ', 'zo')],
  [('だ', 'da'), ('ぢ', 'ji'), ('づ', 'zu'), ('で', 'de'), ('ど', 'do')],
  [('ば', 'ba'), ('び', 'bi'), ('ぶ', 'bu'), ('べ', 'be'), ('ぼ', 'bo')],
  [('ぱ', 'pa'), ('ぴ', 'pi'), ('ぷ', 'pu'), ('ぺ', 'pe'), ('ぽ', 'po')],
];

/// Small kana that build yōon and the double consonant.
const List<_Row> _smallRows = [
  [('ゃ', 'ya'), ('ゅ', 'yu'), ('ょ', 'yo'), ('っ', 'tsu'), null],
];

const _comboConsonants = [
  ('き', 'ky'),
  ('し', 'sh'),
  ('ち', 'ch'),
  ('に', 'ny'),
  ('ひ', 'hy'),
  ('み', 'my'),
  ('り', 'ry'),
  ('ぎ', 'gy'),
  ('じ', 'j'),
  ('び', 'by'),
  ('ぴ', 'py'),
];

const _tips = <String, List<KanaTip>>{
  'あ': [(id: 'Dibaca "a" seperti pada "apa".', en: 'Said "a" as in "father".')],
  'う': [
    (
      id: 'Bibir tidak dimonyongkan; "u" yang ringan.',
      en: 'Lips stay relaxed, not rounded; a light "u".',
    ),
  ],
  'し': [(id: 'Dibaca "shi", bukan "si".', en: 'Said "shi", not "si".')],
  'ち': [
    (
      id: 'Dibaca "chi" seperti "ci" pada "cicak".',
      en: 'Said "chi" as in "cheese".',
    ),
  ],
  'つ': [
    (
      id: 'Dibaca "tsu": huruf t dan s diucapkan menyatu.',
      en: 'Said "tsu": t and s blended together, as in "cats".',
    ),
  ],
  'ふ': [
    (
      id: 'Antara "fu" dan "hu"; tiup udara pelan lewat bibir.',
      en: 'Between "fu" and "hu"; blow air gently through your lips.',
    ),
  ],
  'ら': [
    (
      id: 'Baris ら: lidah menyentuh langit-langit sekali, antara "r" dan "l".',
      en: 'The ら row: tap the tongue once, between "r" and "l".',
    ),
  ],
  'を': [
    (
      id: 'Dibaca "o"; hanya dipakai sebagai partikel objek.',
      en: 'Said "o"; only used as the object particle.',
    ),
  ],
  'ん': [
    (
      id: 'Satu-satunya huruf konsonan tunggal; dibaca "n" atau "ng".',
      en: 'The only stand-alone consonant; said "n" or "ng".',
    ),
  ],
  'は': [
    (
      id: 'Dibaca "wa" bila menjadi partikel topik (わたしは).',
      en: 'Said "wa" when it is the topic particle (わたしは).',
    ),
  ],
  'へ': [
    (
      id: 'Dibaca "e" bila menjadi partikel arah.',
      en: 'Said "e" when it is the direction particle.',
    ),
  ],
  'ぢ': [
    (
      id: 'Bunyinya sama dengan じ; jarang dipakai.',
      en: 'Sounds the same as じ; rarely used.',
    ),
  ],
  'づ': [
    (
      id: 'Bunyinya sama dengan ず; jarang dipakai.',
      en: 'Sounds the same as ず; rarely used.',
    ),
  ],
  'っ': [
    (
      id: 'っ kecil menggandakan konsonan berikutnya: きって = kitte.',
      en: 'Small っ doubles the next consonant: きって = kitte.',
    ),
  ],
  'ゃ': [
    (
      id: 'Ditulis kecil di belakang huruf baris i: き + ゃ = きゃ (kya).',
      en: 'Written small after an i-row kana: き + ゃ = きゃ (kya).',
    ),
  ],
  'ゅ': [
    (
      id: 'Ditulis kecil di belakang huruf baris i: し + ゅ = しゅ (shu).',
      en: 'Written small after an i-row kana: し + ゅ = しゅ (shu).',
    ),
  ],
  'ょ': [
    (
      id: 'Ditulis kecil di belakang huruf baris i: ち + ょ = ちょ (cho).',
      en: 'Written small after an i-row kana: ち + ょ = ちょ (cho).',
    ),
  ],
};

/// Look-alike warnings, keyed by the exact character in either script.
const _lookAlikes = <String, KanaTip>{
  'ぬ': (
    id: 'Mirip め; ぬ punya lingkaran kecil di ujung.',
    en: 'Looks like め; ぬ ends with a small loop.',
  ),
  'め': (
    id: 'Mirip ぬ; め tidak punya lingkaran di ujung.',
    en: 'Looks like ぬ; め has no loop at the end.',
  ),
  'わ': (
    id: 'Mirip ね dan れ; わ berakhir dengan lengkung ke dalam.',
    en: 'Looks like ね and れ; わ ends with an inward curve.',
  ),
  'ね': (
    id: 'Mirip わ dan れ; ね berakhir dengan lingkaran kecil.',
    en: 'Looks like わ and れ; ね ends with a small loop.',
  ),
  'れ': (
    id: 'Mirip わ dan ね; ujung れ melengkung ke luar seperti kaki.',
    en: 'Looks like わ and ね; れ ends with an outward kick.',
  ),
  'る': (
    id: 'Mirip ろ; る punya lingkaran di bawah.',
    en: 'Looks like ろ; る has a loop at the bottom.',
  ),
  'ろ': (
    id: 'Mirip る; ろ tidak punya lingkaran di bawah.',
    en: 'Looks like る; ろ has no loop at the bottom.',
  ),
  'さ': (
    id: 'Cermin dari ち; garis bawah さ membuka ke kiri.',
    en: 'A mirror of ち; the bottom of さ opens to the left.',
  ),
  'ち': (
    id: 'Cermin dari さ; garis bawah ち membuka ke kanan.',
    en: 'A mirror of さ; the bottom of ち opens to the right.',
  ),
  'シ': (
    id: 'Mirip ツ; titik シ berjajar ke bawah dan goresan terakhir naik dari bawah.',
    en: 'Looks like ツ; シ dots stack down and the last stroke sweeps up from below.',
  ),
  'ツ': (
    id: 'Mirip シ; titik ツ berjajar ke samping dan goresan terakhir turun dari atas.',
    en: 'Looks like シ; ツ dots sit side by side and the last stroke falls from the top.',
  ),
  'ソ': (
    id: 'Mirip ン; goresan kedua ソ turun dari atas.',
    en: 'Looks like ン; the second stroke of ソ falls from the top.',
  ),
  'ン': (
    id: 'Mirip ソ; goresan kedua ン naik dari bawah.',
    en: 'Looks like ソ; the second stroke of ン sweeps up from below.',
  ),
  'ク': (
    id: 'Mirip タ; ク tidak punya garis tengah.',
    en: 'Looks like タ; ク has no middle stroke.',
  ),
  'タ': (
    id: 'Mirip ク; タ punya garis pendek di tengah.',
    en: 'Looks like ク; タ has a short middle stroke.',
  ),
  'ウ': (
    id: 'Mirip ワ; ウ punya garis pendek di atas.',
    en: 'Looks like ワ; ウ has a short stroke on top.',
  ),
  'ワ': (
    id: 'Mirip ウ; ワ tidak punya garis di atas.',
    en: 'Looks like ウ; ワ has no stroke on top.',
  ),
};

Kana _single(String hira, String romaji, KanaSection section, KanaScript s) {
  final char = s == KanaScript.hiragana ? hira : toKatakana(hira);
  return Kana._(
    char,
    romaji,
    section,
    [...?_tips[hira], ?_lookAlikes[char]],
    [char],
  );
}

Kana _comboKana(String head, String tail, String romaji, KanaScript s) {
  String conv(String c) => s == KanaScript.hiragana ? c : toKatakana(c);
  return Kana._(
    conv(head) + conv(tail),
    romaji,
    KanaSection.combo,
    const [
      (
        id: 'Gabungan dua huruf dibaca satu ketukan. Huruf kedua ditulis kecil.',
        en: 'Two kana read as one beat. The second one is written small.',
      ),
    ],
    [conv(head), conv(tail)],
  );
}

KanaChart _buildChart(KanaScript script) {
  List<List<Kana?>> rows(List<_Row> source, KanaSection section) => [
    for (final row in source)
      [
        for (final cell in row)
          cell == null ? null : _single(cell.$1, cell.$2, section, script),
      ],
  ];
  final small = rows(_smallRows, KanaSection.combo);
  final combos = <List<Kana?>>[
    for (final (head, sound) in _comboConsonants)
      [
        for (final (tail, vowel) in const [('ゃ', 'a'), ('ゅ', 'u'), ('ょ', 'o')])
          _comboKana(head, tail, '$sound$vowel', script),
      ],
  ];
  return KanaChart._(script, {
    KanaSection.basic: rows(_basicRows, KanaSection.basic),
    KanaSection.dakuten: rows(_dakutenRows, KanaSection.dakuten),
    KanaSection.combo: [...small, ...combos],
  });
}

/// The chart for one script, grouped into rows. `null` cells are blanks.
class KanaChart {
  KanaChart._(this.script, this.sections);

  factory KanaChart.of(KanaScript script) =>
      script == KanaScript.hiragana ? _hiragana : _katakana;

  static final _hiragana = _buildChart(KanaScript.hiragana);
  static final _katakana = _buildChart(KanaScript.katakana);

  final KanaScript script;
  final Map<KanaSection, List<List<Kana?>>> sections;

  /// Every cell of a section in reading order, without blanks.
  List<Kana> list(KanaSection section) => [
    for (final row in sections[section]!) ...row.nonNulls,
  ];

  /// Every single character the learner can practise writing in a section.
  List<Kana> writable(KanaSection section) =>
      list(section).where((k) => !k.isCombo).toList();

  /// Finds the section a character belongs to.
  Kana? find(String char) {
    for (final section in KanaSection.values) {
      for (final kana in list(section)) {
        if (kana.char == char) return kana;
      }
    }
    return null;
  }
}

KanaScript scriptOf(String char) {
  final rune = char.runes.first;
  return rune >= 0x30A0 ? KanaScript.katakana : KanaScript.hiragana;
}
