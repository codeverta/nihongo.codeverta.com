import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';

String formatYen(int yen) =>
    '¥${NumberFormat.decimalPattern('id').format(yen)}';

/// Rupiah in a short, readable form: "Rp 21,1 jt" / "Rp 950 rb".
String formatRupiahShort(int rupiah, {bool english = false}) {
  final format = NumberFormat.decimalPattern(english ? 'en' : 'id');
  if (rupiah >= 1000000) {
    final millions = (rupiah / 100000).round() / 10;
    return 'Rp ${format.format(millions)} ${english ? 'M' : 'jt'}';
  }
  final thousands = (rupiah / 1000).round();
  return 'Rp ${format.format(thousands)} ${english ? 'K' : 'rb'}';
}

/// Minimum wage, indicative monthly pay, rent and price level for one
/// prefecture, with sources. Hidden when the data is unavailable.
class PrefectureLivingCard extends ConsumerWidget {
  const PrefectureLivingCard({required this.code, super.key});

  final String code;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(mapPrefectureLivingProvider).valueOrNull;
    final living = data?[code];
    if (data == null || living == null || living.minimumWage <= 0) {
      return const SizedBox.shrink();
    }
    final theme = Theme.of(context);
    final english = AppLocalizations.of(context).locale.languageCode == 'en';
    final monthly = data.monthlyWage(living);
    final monthlyIdr = data.toRupiah(monthly);
    final rank =
        data.prefectures.values
            .where((p) => p.minimumWage > living.minimumWage)
            .length +
        1;
    final hours = data.meta.hoursPerDay % 1 == 0
        ? data.meta.hoursPerDay.toInt().toString()
        : data.meta.hoursPerDay.toString();
    final since = living.wageEffectiveDate;
    final index = living.priceIndex;

    return Container(
      key: const Key('prefecture-living-card'),
      margin: const EdgeInsets.only(top: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerLow,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.payments_outlined,
                size: 20,
                color: theme.colorScheme.primary,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Upah & biaya hidup'.localized(context),
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              Text(
                '#$rank / ${data.prefectures.length}',
                key: const Key('prefecture-wage-rank'),
                style: theme.textTheme.labelMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _Stat(
            label: 'Upah minimum',
            value:
                '${formatYen(living.minimumWage)} / ${'jam'.localized(context)}',
            note: since == null
                ? null
                : '${'Berlaku sejak'.localized(context)} ${DateFormat.yMMMd(english ? 'en' : 'id').format(since)}',
            emphasis: true,
          ),
          _Stat(
            label: 'Estimasi gaji bulanan',
            value:
                '≈ ${formatYen(monthly)}${monthlyIdr == null ? '' : '  ·  ${formatRupiahShort(monthlyIdr, english: english)}'}',
            note:
                '$hours ${'jam'.localized(context)} × ${data.meta.daysPerMonth} ${'hari kerja, sebelum pajak & asuransi'.localized(context)}',
          ),
          if (living.rent1kEstimate != null)
            _Stat(
              label: 'Perkiraan sewa 1K (±25 m²)',
              value:
                  '≈ ${formatYen(living.rent1kEstimate!)} / ${'bulan'.localized(context)}',
              note: data.meta.rentSurveyYear.isEmpty
                  ? null
                  : '${'Rata-rata sewa swasta'.localized(context)} · ${data.meta.rentSurveyYear}',
            ),
          if (index != null)
            _Stat(
              label: 'Indeks harga',
              value: NumberFormat('0.0', english ? 'en' : 'id').format(index),
              note: index >= 100.5
                  ? 'Lebih mahal dari rata-rata nasional (100)'.localized(
                      context,
                    )
                  : index <= 99.5
                  ? 'Lebih murah dari rata-rata nasional (100)'.localized(
                      context,
                    )
                  : 'Setara rata-rata nasional (100)'.localized(context),
            ),
          const SizedBox(height: 6),
          Text(
            [
              if (data.meta.wageFiscalYear.isNotEmpty)
                '${'Upah'.localized(context)}: MHLW ${data.meta.wageFiscalYear}',
              if (data.meta.rentSurveyYear.isNotEmpty)
                '${'Sewa & harga'.localized(context)}: 総務省 ${data.meta.rentSurveyYear}',
              if (monthlyIdr != null && data.meta.exchangeRateDate.isNotEmpty)
                '${'Kurs'.localized(context)} ${data.meta.exchangeRateDate}',
            ].join(' · '),
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          Text(
            'Angka indikatif, bukan tawaran gaji.'.localized(context),
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({
    required this.label,
    required this.value,
    this.note,
    this.emphasis = false,
  });

  final String label;
  final String value;
  final String? note;
  final bool emphasis;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label.localized(context),
            style: theme.textTheme.labelMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          Text(
            value,
            style:
                (emphasis
                        ? theme.textTheme.titleLarge
                        : theme.textTheme.titleMedium)
                    ?.copyWith(fontWeight: FontWeight.w800),
          ),
          if (note != null)
            Text(
              note!,
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
        ],
      ),
    );
  }
}

/// Color for a 0…1 wage share: cool (low) → warm (high).
Color wageHeatColor(double share, {required bool dark}) {
  const low = Color(0xFF9EC5E8);
  const mid = Color(0xFFF6D776);
  const high = Color(0xFFE8705A);
  final color = share < 0.5
      ? Color.lerp(low, mid, share * 2)!
      : Color.lerp(mid, high, (share - 0.5) * 2)!;
  return dark ? Color.lerp(color, Colors.black, 0.25)! : color;
}

/// Legend for the wage heatmap.
class WageHeatLegend extends ConsumerWidget {
  const WageHeatLegend({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(mapPrefectureLivingProvider).valueOrNull;
    if (data == null) return const SizedBox.shrink();
    final (low, high) = data.wageRange;
    final theme = Theme.of(context);
    final dark = theme.brightness == Brightness.dark;
    return Container(
      key: const Key('wage-heat-legend'),
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 10),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface.withValues(alpha: 0.92),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${'Upah minimum'.localized(context)} ${data.meta.wageFiscalYear}',
            style: theme.textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          Container(
            width: 150,
            height: 8,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(4),
              gradient: LinearGradient(
                colors: [
                  wageHeatColor(0, dark: dark),
                  wageHeatColor(0.5, dark: dark),
                  wageHeatColor(1, dark: dark),
                ],
              ),
            ),
          ),
          const SizedBox(height: 4),
          SizedBox(
            width: 150,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(formatYen(low), style: theme.textTheme.labelSmall),
                Text(formatYen(high), style: theme.textTheme.labelSmall),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
