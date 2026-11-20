import 'package:flutter/foundation.dart';

/// Google's public AdMob test units. They never serve real ads or earn money.
const testBannerAdUnitId = 'ca-app-pub-3940256099942544/6300978111';
const testNativeAdUnitId = 'ca-app-pub-3940256099942544/2247696110';
const testRewardedAdUnitId = 'ca-app-pub-3940256099942544/5224354917';
const testInterstitialAdUnitId = 'ca-app-pub-3940256099942544/1033173712';

/// Default cards an Anki session needs before it counts as finished.
const defaultAnkiSessionMinCards = 10;

@immutable
class AdPlacementConfig {
  const AdPlacementConfig({
    this.enabled = false,
    this.unitId = '',
    this.every = 0,
  });

  factory AdPlacementConfig.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) return const AdPlacementConfig();
    final unit = json['unitId'];
    final every = json['every'];
    return AdPlacementConfig(
      enabled: json['enabled'] == true,
      unitId: unit is String ? unit.trim() : '',
      every: every is int ? every : 0,
    );
  }

  final bool enabled;
  final String unitId;
  final int every;

  bool get usable => enabled && unitId.isNotEmpty;

  AdPlacementConfig withUnit(String unit) =>
      AdPlacementConfig(enabled: enabled, unitId: unit, every: every);
}

/// The kinds of "finished studying" moments that may be followed by a
/// full-screen ad. Each has its own switch in the admin dashboard.
enum CompletionKind {
  /// A JLPT grammar, kanji, reading or vocabulary lesson quiz.
  lessons,

  /// A simulation, diagnostic, JFT, NAT or SSW exam.
  exams,

  /// An SSW practice session or a Kaigo lesson.
  practice,

  /// An Anki session past the minimum card count.
  anki,
}

/// Which "finished studying" moments may be followed by a full-screen ad.
/// Mirrors `interstitialKinds` in `/v1/ads-config`; a missing value counts as on.
@immutable
class CompletionKindsConfig {
  const CompletionKindsConfig({
    this.lessons = true,
    this.exams = true,
    this.practice = true,
    this.anki = true,
  });

  factory CompletionKindsConfig.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) return const CompletionKindsConfig();
    bool on(String key) => json[key] != false;
    return CompletionKindsConfig(
      lessons: on('lessons'),
      exams: on('exams'),
      practice: on('practice'),
      anki: on('anki'),
    );
  }

  final bool lessons;
  final bool exams;
  final bool practice;
  final bool anki;

  bool allows(CompletionKind kind) => switch (kind) {
    CompletionKind.lessons => lessons,
    CompletionKind.exams => exams,
    CompletionKind.practice => practice,
    CompletionKind.anki => anki,
  };
}

/// Which kinds of locked content a rewarded ad may open. Mirrors
/// `unlockKinds` in `/v1/ads-config`; a missing value counts as on.
@immutable
class UnlockKindsConfig {
  const UnlockKindsConfig({
    this.lessons = true,
    this.anki = true,
    this.simulation = true,
    this.news = true,
    this.jobs = true,
  });

  factory UnlockKindsConfig.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) return const UnlockKindsConfig();
    bool on(String key) => json[key] != false;
    return UnlockKindsConfig(
      lessons: on('lessons'),
      anki: on('anki'),
      simulation: on('simulation'),
      news: on('news'),
      jobs: on('jobs'),
    );
  }

  final bool lessons;
  final bool anki;
  final bool simulation;
  final bool news;
  final bool jobs;

  /// Whether an ad may open an item of [kind] (the server's kind names).
  bool allows(String kind) => switch (kind) {
    'grammar' || 'kanji' || 'reading' || 'vocabulary' => lessons,
    'anki' => anki,
    'simulation' => simulation,
    'news' => news,
    'job' => jobs,
    _ => false,
  };
}

/// Remote on/off switch for in-app ads. Everything is off by default so a
/// failed or missing response never shows an ad.
@immutable
class AdsConfig {
  const AdsConfig({
    this.enabled = false,
    this.banner = const AdPlacementConfig(),
    this.newsNative = const AdPlacementConfig(),
    this.jobsNative = const AdPlacementConfig(),
    this.aiChatRewarded = const AdPlacementConfig(),
    this.interstitial = const AdPlacementConfig(),
    this.dictionaryNative = const AdPlacementConfig(),
    this.quizResultBanner = const AdPlacementConfig(),
    this.unlockRewarded = const AdPlacementConfig(),
    this.unlockKinds = const UnlockKindsConfig(),
    this.interstitialKinds = const CompletionKindsConfig(),
    this.examTabBanner = const AdPlacementConfig(),
    this.practiceResultBanner = const AdPlacementConfig(),
    this.newsArticleNative = const AdPlacementConfig(),
    this.ankiSessionMinCards = defaultAnkiSessionMinCards,
    this.xpBoostRewarded = const AdPlacementConfig(),
    this.xpBoostKinds = const CompletionKindsConfig(),
  });

  factory AdsConfig.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) return const AdsConfig();
    return AdsConfig(
      enabled: json['enabled'] == true,
      banner: AdPlacementConfig.fromJson(json['banner']),
      newsNative: AdPlacementConfig.fromJson(json['newsNative']),
      jobsNative: AdPlacementConfig.fromJson(json['jobsNative']),
      aiChatRewarded: AdPlacementConfig.fromJson(json['aiChatRewarded']),
      interstitial: AdPlacementConfig.fromJson(json['interstitial']),
      dictionaryNative: AdPlacementConfig.fromJson(json['dictionaryNative']),
      quizResultBanner: AdPlacementConfig.fromJson(json['quizResultBanner']),
      unlockRewarded: AdPlacementConfig.fromJson(json['unlockRewarded']),
      unlockKinds: UnlockKindsConfig.fromJson(json['unlockKinds']),
      interstitialKinds: CompletionKindsConfig.fromJson(
        json['interstitialKinds'],
      ),
      examTabBanner: AdPlacementConfig.fromJson(json['examTabBanner']),
      practiceResultBanner: AdPlacementConfig.fromJson(
        json['practiceResultBanner'],
      ),
      newsArticleNative: AdPlacementConfig.fromJson(json['newsArticleNative']),
      xpBoostRewarded: AdPlacementConfig.fromJson(json['xpBoostRewarded']),
      xpBoostKinds: CompletionKindsConfig.fromJson(json['xpBoostKinds']),
      ankiSessionMinCards: switch (json['ankiSessionMinCards']) {
        final int n when n >= 1 && n <= 100 => n,
        _ => defaultAnkiSessionMinCards,
      },
    );
  }

  static const off = AdsConfig();

  final bool enabled;
  final AdPlacementConfig banner;
  final AdPlacementConfig newsNative;
  final AdPlacementConfig jobsNative;

  /// Opt-in rewarded ad that unlocks one Tanya AI message after the daily
  /// free quota.
  final AdPlacementConfig aiChatRewarded;

  /// Full-screen ad at the next natural break after [AdPlacementConfig.every]
  /// minutes of foreground use since the last full-screen ad.
  final AdPlacementConfig interstitial;

  /// Native ad inserted every N dictionary search results.
  final AdPlacementConfig dictionaryNative;

  /// Banner on lesson quiz and exam result screens.
  final AdPlacementConfig quizResultBanner;

  /// Opt-in rewarded ad that opens one locked item, verified by the server
  /// through AdMob server-side verification.
  final AdPlacementConfig unlockRewarded;

  /// Which kinds of item [unlockRewarded] may open.
  final UnlockKindsConfig unlockKinds;

  /// Which completion moments [interstitial] may follow.
  final CompletionKindsConfig interstitialKinds;

  /// Banner above the bottom navigation on the Exam tab.
  final AdPlacementConfig examTabBanner;

  /// Banner at the end of an Anki session, SSW practice or Kaigo lesson.
  final AdPlacementConfig practiceResultBanner;

  /// One native ad at the end of a news article.
  final AdPlacementConfig newsArticleNative;

  /// Cards an Anki session needs before it counts as finished (for the
  /// full-screen ad; the server uses the same number for XP).
  final int ankiSessionMinCards;

  /// Opt-in rewarded ad on result screens that doubles that activity's XP,
  /// verified by the server through AdMob server-side verification.
  final AdPlacementConfig xpBoostRewarded;

  /// Which activities offer [xpBoostRewarded].
  final CompletionKindsConfig xpBoostKinds;

  /// Debug builds must never request real ads, so swap in Google test units.
  AdsConfig withTestUnits() => AdsConfig(
    enabled: enabled,
    banner: banner.withUnit(testBannerAdUnitId),
    newsNative: newsNative.withUnit(testNativeAdUnitId),
    jobsNative: jobsNative.withUnit(testNativeAdUnitId),
    aiChatRewarded: aiChatRewarded.withUnit(testRewardedAdUnitId),
    interstitial: interstitial.withUnit(testInterstitialAdUnitId),
    dictionaryNative: dictionaryNative.withUnit(testNativeAdUnitId),
    quizResultBanner: quizResultBanner.withUnit(testBannerAdUnitId),
    unlockRewarded: unlockRewarded.withUnit(testRewardedAdUnitId),
    unlockKinds: unlockKinds,
    interstitialKinds: interstitialKinds,
    examTabBanner: examTabBanner.withUnit(testBannerAdUnitId),
    practiceResultBanner: practiceResultBanner.withUnit(testBannerAdUnitId),
    newsArticleNative: newsArticleNative.withUnit(testNativeAdUnitId),
    ankiSessionMinCards: ankiSessionMinCards,
    xpBoostRewarded: xpBoostRewarded.withUnit(testRewardedAdUnitId),
    xpBoostKinds: xpBoostKinds,
  );
}
