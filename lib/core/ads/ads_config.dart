import 'package:flutter/foundation.dart';

/// Google's public AdMob test units. They never serve real ads or earn money.
const testBannerAdUnitId = 'ca-app-pub-3940256099942544/6300978111';
const testNativeAdUnitId = 'ca-app-pub-3940256099942544/2247696110';
const testRewardedAdUnitId = 'ca-app-pub-3940256099942544/5224354917';
const testInterstitialAdUnitId = 'ca-app-pub-3940256099942544/1033173712';

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
  );
}
