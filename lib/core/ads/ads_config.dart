import 'package:flutter/foundation.dart';

/// Google's public AdMob test units. They never serve real ads or earn money.
const testBannerAdUnitId = 'ca-app-pub-3940256099942544/6300978111';
const testNativeAdUnitId = 'ca-app-pub-3940256099942544/2247696110';

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
  });

  factory AdsConfig.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) return const AdsConfig();
    return AdsConfig(
      enabled: json['enabled'] == true,
      banner: AdPlacementConfig.fromJson(json['banner']),
      newsNative: AdPlacementConfig.fromJson(json['newsNative']),
      jobsNative: AdPlacementConfig.fromJson(json['jobsNative']),
    );
  }

  static const off = AdsConfig();

  final bool enabled;
  final AdPlacementConfig banner;
  final AdPlacementConfig newsNative;
  final AdPlacementConfig jobsNative;

  /// Debug builds must never request real ads, so swap in Google test units.
  AdsConfig withTestUnits() => AdsConfig(
    enabled: enabled,
    banner: banner.withUnit(testBannerAdUnitId),
    newsNative: newsNative.withUnit(testNativeAdUnitId),
    jobsNative: jobsNative.withUnit(testNativeAdUnitId),
  );
}
