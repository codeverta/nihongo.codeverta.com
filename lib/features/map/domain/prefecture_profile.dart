import 'package:flutter/foundation.dart';

/// One titled section of a prefecture article (History, Geography, …).
@immutable
class ProfileSection {
  const ProfileSection({required this.title, required this.text});

  factory ProfileSection.fromJson(Map<String, dynamic> json) => ProfileSection(
    title: json['title'] as String? ?? '',
    text: json['text'] as String? ?? '',
  );

  final String title;
  final String text;
}

/// Wikipedia text for a prefecture in one language.
@immutable
class ProfileText {
  const ProfileText({
    required this.lang,
    required this.title,
    required this.url,
    required this.intro,
    required this.sections,
  });

  factory ProfileText.fromJson(String lang, Map<String, dynamic> json) =>
      ProfileText(
        lang: lang,
        title: json['title'] as String? ?? '',
        url: json['url'] as String? ?? '',
        intro: json['intro'] as String? ?? '',
        sections: [
          for (final item in (json['sections'] as List?) ?? const [])
            if (item is Map<String, dynamic>) ProfileSection.fromJson(item),
        ].where((s) => s.text.trim().isNotEmpty).toList(),
      );

  final String lang;
  final String title;
  final String url;
  final String intro;
  final List<ProfileSection> sections;

  bool get isEmpty => intro.trim().isEmpty;
}

/// A Wikipedia-style profile of one prefecture: facts from Wikidata and text
/// from Wikipedia (CC BY-SA 4.0).
@immutable
class PrefectureProfile {
  const PrefectureProfile({
    required this.code,
    required this.nameJa,
    required this.nameKana,
    required this.nameId,
    required this.nameEn,
    required this.capital,
    required this.texts,
    this.population,
    this.populationDate,
    this.areaKm2,
    this.website,
  });

  factory PrefectureProfile.fromJson(Map<String, dynamic> json) {
    final capital = json['capital'];
    final text = json['text'];
    return PrefectureProfile(
      code: json['code'] as String? ?? '',
      nameJa: json['nameJa'] as String? ?? '',
      nameKana: json['nameKana'] as String? ?? '',
      nameId: json['nameId'] as String? ?? '',
      nameEn: json['nameEn'] as String? ?? '',
      capital: {
        if (capital is Map)
          for (final entry in capital.entries)
            if (entry.value is String && (entry.value as String).isNotEmpty)
              '${entry.key}': entry.value as String,
      },
      population: (json['population'] as num?)?.toInt(),
      populationDate: json['populationDate'] as String?,
      areaKm2: (json['areaKm2'] as num?)?.toDouble(),
      website: json['website'] as String?,
      texts: {
        if (text is Map)
          for (final entry in text.entries)
            if (entry.value is Map<String, dynamic>)
              '${entry.key}': ProfileText.fromJson(
                '${entry.key}',
                entry.value as Map<String, dynamic>,
              ),
      },
    );
  }

  final String code;
  final String nameJa;
  final String nameKana;
  final String nameId;
  final String nameEn;

  /// Capital name per language (`ja`, `id`, `en`).
  final Map<String, String> capital;
  final int? population;
  final String? populationDate;
  final double? areaKm2;
  final String? website;
  final Map<String, ProfileText> texts;

  String nameFor(String lang) =>
      (lang == 'en' ? nameEn : nameId).trim().isNotEmpty
      ? (lang == 'en' ? nameEn : nameId)
      : nameEn;

  String? capitalFor(String lang) => capital[lang] ?? capital['en'];

  /// Text in [lang], falling back to English (the caller labels fallbacks).
  ProfileText? textFor(String lang) {
    final own = texts[lang];
    if (own != null && !own.isEmpty) return own;
    final english = texts['en'];
    return english != null && !english.isEmpty ? english : null;
  }

  /// People per km², when both figures are known.
  double? get density => population != null && areaKm2 != null && areaKm2! > 0
      ? population! / areaKm2!
      : null;
}

/// All 47 profiles keyed by JIS code.
@immutable
class PrefectureProfiles {
  const PrefectureProfiles({required this.byCode, this.fromCache = false});

  factory PrefectureProfiles.fromJson(
    Map<String, dynamic> json, {
    bool fromCache = false,
  }) => PrefectureProfiles(
    byCode: {
      for (final item in (json['prefectures'] as List?) ?? const [])
        if (item is Map<String, dynamic>)
          item['code'] as String? ?? '': PrefectureProfile.fromJson(item),
    },
    fromCache: fromCache,
  );

  final Map<String, PrefectureProfile> byCode;
  final bool fromCache;

  PrefectureProfile? operator [](String code) => byCode[code];
}

/// The eight traditional regions, from the JIS prefecture code.
String regionOf(String code) {
  final n = int.tryParse(code) ?? 0;
  return switch (n) {
    1 => 'Hokkaido',
    <= 7 => 'Tohoku',
    <= 14 => 'Kanto',
    <= 23 => 'Chubu',
    <= 30 => 'Kansai',
    <= 35 => 'Chugoku',
    <= 39 => 'Shikoku',
    _ => 'Kyushu & Okinawa',
  };
}
