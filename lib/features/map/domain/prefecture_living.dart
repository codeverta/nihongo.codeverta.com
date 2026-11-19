import 'package:flutter/foundation.dart';

/// Official minimum wage, rent and price level for one prefecture.
@immutable
class PrefectureLiving {
  const PrefectureLiving({
    required this.code,
    required this.nameJa,
    required this.minimumWage,
    this.wageEffectiveDate,
    this.rentPer33m2,
    this.rent1kEstimate,
    this.priceIndex,
  });

  factory PrefectureLiving.fromJson(Map<String, dynamic> json) =>
      PrefectureLiving(
        code: json['code'] as String? ?? '',
        nameJa: json['nameJa'] as String? ?? '',
        minimumWage: (json['minimumWage'] as num?)?.toInt() ?? 0,
        wageEffectiveDate: DateTime.tryParse(
          json['wageEffectiveDate'] as String? ?? '',
        ),
        rentPer33m2: (json['rentPer33m2'] as num?)?.toInt(),
        rent1kEstimate: (json['rent1kEstimate'] as num?)?.toInt(),
        priceIndex: (json['priceIndex'] as num?)?.toDouble(),
      );

  final String code;
  final String nameJa;

  /// Hourly minimum wage in yen (地域別最低賃金).
  final int minimumWage;
  final DateTime? wageEffectiveDate;

  /// Average private rent per 3.3 m² per month (official statistics).
  final int? rentPer33m2;

  /// Indicative monthly rent for a ~25 m² 1K flat derived from [rentPer33m2].
  final int? rent1kEstimate;

  /// Regional price index, national average = 100.
  final double? priceIndex;
}

@immutable
class PrefectureLivingMeta {
  const PrefectureLivingMeta({
    required this.wageFiscalYear,
    required this.wageSource,
    required this.rentMeasure,
    required this.rentSurveyYear,
    required this.rentSource,
    required this.priceIndexYear,
    required this.priceIndexSource,
    required this.idrPerJpy,
    required this.exchangeRateDate,
    required this.hoursPerDay,
    required this.daysPerMonth,
  });

  factory PrefectureLivingMeta.fromJson(Map<String, dynamic> json) {
    String text(String key) => json[key] as String? ?? '';
    return PrefectureLivingMeta(
      wageFiscalYear: text('wageFiscalYear'),
      wageSource: text('wageSource'),
      rentMeasure: text('rentMeasure'),
      rentSurveyYear: text('rentSurveyYear'),
      rentSource: text('rentSource'),
      priceIndexYear: text('priceIndexYear'),
      priceIndexSource: text('priceIndexSource'),
      idrPerJpy: (json['idrPerJpy'] as num?)?.toDouble() ?? 0,
      exchangeRateDate: text('exchangeRateDate'),
      hoursPerDay: (json['hoursPerDay'] as num?)?.toDouble() ?? 8,
      daysPerMonth: (json['daysPerMonth'] as num?)?.toInt() ?? 21,
    );
  }

  final String wageFiscalYear;
  final String wageSource;
  final String rentMeasure;
  final String rentSurveyYear;
  final String rentSource;
  final String priceIndexYear;
  final String priceIndexSource;
  final double idrPerJpy;
  final String exchangeRateDate;
  final double hoursPerDay;
  final int daysPerMonth;
}

@immutable
class PrefectureLivingData {
  const PrefectureLivingData({
    required this.meta,
    required this.prefectures,
    this.fromCache = false,
  });

  factory PrefectureLivingData.fromJson(
    Map<String, dynamic> json, {
    bool fromCache = false,
  }) {
    final items = json['prefectures'];
    return PrefectureLivingData(
      meta: PrefectureLivingMeta.fromJson(
        json['meta'] as Map<String, dynamic>? ?? const {},
      ),
      prefectures: {
        if (items is List)
          for (final item in items)
            if (item is Map<String, dynamic>)
              item['code'] as String? ?? '': PrefectureLiving.fromJson(item),
      },
      fromCache: fromCache,
    );
  }

  final PrefectureLivingMeta meta;
  final Map<String, PrefectureLiving> prefectures;
  final bool fromCache;

  PrefectureLiving? operator [](String code) => prefectures[code];

  /// Indicative full-time monthly pay at the minimum wage, in yen.
  int monthlyWage(PrefectureLiving prefecture) =>
      (prefecture.minimumWage * meta.hoursPerDay * meta.daysPerMonth).round();

  /// Yen converted to rupiah at the admin-set reference rate, or null when
  /// no rate is configured.
  int? toRupiah(int yen) =>
      meta.idrPerJpy > 0 ? (yen * meta.idrPerJpy).round() : null;

  /// Lowest and highest minimum wage, for the heatmap legend.
  (int, int) get wageRange {
    final wages = [
      for (final p in prefectures.values)
        if (p.minimumWage > 0) p.minimumWage,
    ];
    if (wages.isEmpty) return (0, 0);
    wages.sort();
    return (wages.first, wages.last);
  }

  /// 0 (lowest wage) … 1 (highest wage) for a prefecture, or null.
  double? wageShare(String code) {
    final wage = prefectures[code]?.minimumWage;
    final (low, high) = wageRange;
    if (wage == null || wage <= 0 || high <= low) return null;
    return (wage - low) / (high - low);
  }
}
