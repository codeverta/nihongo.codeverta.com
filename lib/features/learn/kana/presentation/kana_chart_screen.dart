import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/learn/kana/data/kana_repository.dart';
import 'package:nihongo_master/features/learn/kana/domain/kana_models.dart';

String kanaSectionTitle(BuildContext context, KanaSection section) =>
    switch (section) {
      KanaSection.basic => 'Dasar',
      KanaSection.dakuten => 'Dakuten & handakuten',
      KanaSection.combo => 'Huruf kecil & gabungan (yōon)',
    }.localized(context);

String _sectionHint(BuildContext context, KanaSection section) =>
    switch (section) {
      KanaSection.basic => 'Mulai dari sini: 46 huruf utama.',
      KanaSection.dakuten => 'Tanda ゛ atau ゜ mengubah bunyi: か (ka) → が (ga).',
      KanaSection.combo =>
        'Huruf kecil ゃ ゅ ょ menempel pada huruf sebelumnya: き + ゃ = きゃ.',
    }.localized(context);

/// Tabel hiragana & katakana untuk pemula.
class KanaChartScreen extends ConsumerWidget {
  const KanaChartScreen({this.script = KanaScript.hiragana, super.key});

  final KanaScript script;

  @override
  Widget build(BuildContext context, WidgetRef ref) => DefaultTabController(
    length: 2,
    initialIndex: script.index,
    child: Scaffold(
      appBar: AppBar(
        title: const Text('Hiragana & Katakana'),
        bottom: const TabBar(
          tabs: [
            Tab(text: 'Hiragana'),
            Tab(text: 'Katakana'),
          ],
        ),
      ),
      body: const TabBarView(
        children: [
          _ChartView(script: KanaScript.hiragana),
          _ChartView(script: KanaScript.katakana),
        ],
      ),
    ),
  );
}

class _ChartView extends ConsumerWidget {
  const _ChartView({required this.script});

  final KanaScript script;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final chart = KanaChart.of(script);
    final practiced = ref.watch(kanaProgressProvider);
    final theme = Theme.of(context);
    return ListView(
      key: PageStorageKey('kana-${script.name}'),
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
      children: [
        Text(
          script == KanaScript.hiragana
              ? 'Hiragana dipakai untuk kata asli Jepang dan akhiran kata.'
                    .localized(context)
              : 'Katakana dipakai untuk kata serapan, nama asing, dan penekanan.'
                    .localized(context),
          style: theme.textTheme.bodyMedium,
        ),
        const SizedBox(height: 4),
        Text(
          'Ketuk huruf untuk melihat urutan goresan dan berlatih menulis.'
              .localized(context),
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
        for (final section in KanaSection.values) ...[
          const SizedBox(height: 22),
          _SectionHeader(
            title: kanaSectionTitle(context, section),
            hint: _sectionHint(context, section),
            done: chart
                .writable(section)
                .where((k) => practiced.contains(k.char))
                .length,
            total: chart.writable(section).length,
          ),
          const SizedBox(height: 10),
          for (final row in chart.sections[section]!)
            Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  for (final (index, cell) in row.indexed) ...[
                    if (index > 0) const SizedBox(width: 6),
                    Expanded(
                      child: cell == null
                          ? const SizedBox.shrink()
                          : _KanaCell(
                              kana: cell,
                              practiced: cell.parts.every(practiced.contains),
                            ),
                    ),
                  ],
                ],
              ),
            ),
        ],
      ],
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({
    required this.title,
    required this.hint,
    required this.done,
    required this.total,
  });

  final String title;
  final String hint;
  final int done;
  final int total;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              title,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
            Semantics(
              label: '${'Sudah dikuasai'.localized(context)} $done / $total',
              excludeSemantics: true,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: theme.colorScheme.secondaryContainer,
                  borderRadius: BorderRadius.circular(99),
                ),
                child: Text(
                  '✓ $done/$total',
                  style: TextStyle(
                    fontWeight: FontWeight.w700,
                    color: theme.colorScheme.onSecondaryContainer,
                  ),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 2),
        Text(
          hint,
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }
}

class _KanaCell extends StatelessWidget {
  const _KanaCell({required this.kana, required this.practiced});

  final Kana kana;
  final bool practiced;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Semantics(
      button: true,
      label: '${kana.char} ${kana.romaji}',
      excludeSemantics: true,
      child: Material(
        color: practiced
            ? colors.primaryContainer
            : colors.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        child: InkWell(
          key: Key('kana-cell-${kana.char}'),
          borderRadius: BorderRadius.circular(12),
          onTap: () => context.push(AppRoutes.kanaCharPath(kana.char)),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 2),
            child: Column(
              children: [
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Text(
                    kana.char,
                    style: AppTypography.japanese(
                      fontSize: 24,
                      fontWeight: FontWeight.w600,
                      height: 1.2,
                      color: colors.onSurface,
                    ),
                  ),
                ),
                FittedBox(
                  fit: BoxFit.scaleDown,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (practiced)
                        Icon(
                          Icons.check_rounded,
                          size: 12,
                          color: colors.primary,
                        ),
                      Text(
                        kana.romaji,
                        style: TextStyle(
                          fontSize: 11,
                          color: colors.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
