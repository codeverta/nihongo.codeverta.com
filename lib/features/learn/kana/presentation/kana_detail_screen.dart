import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/learn/kana/data/kana_repository.dart';
import 'package:nihongo_master/features/learn/kana/domain/kana_models.dart';
import 'package:nihongo_master/features/learn/kana/presentation/kana_chart_screen.dart';
import 'package:nihongo_master/features/learn/kana/presentation/widgets/stroke_order_player.dart';

String kanaTipText(BuildContext context, KanaTip tip) =>
    AppLocalizations.of(context).locale.languageCode == 'en' ? tip.en : tip.id;

/// Previous/next character within the same chart section.
(Kana?, Kana?) kanaNeighbours(Kana kana, {bool writableOnly = false}) {
  final chart = KanaChart.of(scriptOf(kana.char));
  final list = writableOnly
      ? chart.writable(kana.section)
      : chart.list(kana.section);
  final index = list.indexWhere((k) => k.char == kana.char);
  if (index < 0) return (null, null);
  return (
    index > 0 ? list[index - 1] : null,
    index < list.length - 1 ? list[index + 1] : null,
  );
}

/// Detail huruf: urutan goresan, romaji, tips, dan pintu ke latihan menulis.
class KanaDetailScreen extends ConsumerWidget {
  const KanaDetailScreen({required this.char, super.key});

  final String char;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final kana = KanaChart.of(scriptOf(char)).find(char);
    if (kana == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(child: Text('Huruf tidak ditemukan.'.localized(context))),
      );
    }
    final (previous, next) = kanaNeighbours(kana);
    final practiced = ref.watch(kanaProgressProvider);
    final strokes = ref.watch(kanaStrokesProvider);
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(kanaSectionTitle(context, kana.section)),
        actions: [
          IconButton(
            tooltip: 'Huruf sebelumnya'.localized(context),
            onPressed: previous == null
                ? null
                : () => context.pushReplacement(
                    AppRoutes.kanaCharPath(previous.char),
                  ),
            icon: const Icon(Icons.arrow_back_ios_new_rounded, size: 18),
          ),
          IconButton(
            tooltip: 'Huruf berikutnya'.localized(context),
            onPressed: next == null
                ? null
                : () => context.pushReplacement(
                    AppRoutes.kanaCharPath(next.char),
                  ),
            icon: const Icon(Icons.arrow_forward_ios_rounded, size: 18),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
        children: [
          Center(
            child: Text(
              kana.char,
              key: const Key('kana-detail-char'),
              style: AppTypography.japanese(
                fontSize: 56,
                fontWeight: FontWeight.w600,
                height: 1.1,
              ),
            ),
          ),
          Center(
            child: Text(
              kana.romaji,
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.w800,
                color: theme.colorScheme.primary,
              ),
            ),
          ),
          const SizedBox(height: 16),
          if (kana.isCombo)
            ..._comboParts(context, kana, practiced)
          else
            strokes.when(
              data: (data) => StrokeOrderPlayer(
                key: ValueKey(kana.char),
                strokes: data[kana.char] ?? const [],
              ),
              loading: () => const SizedBox(
                height: 200,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (_, _) => Text(
                'Data urutan goresan belum dapat dimuat.'.localized(context),
              ),
            ),
          const SizedBox(height: 16),
          for (final tip in kana.tips)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(
                    Icons.lightbulb_outline_rounded,
                    size: 20,
                    color: theme.colorScheme.tertiary,
                  ),
                  const SizedBox(width: 8),
                  Expanded(child: Text(kanaTipText(context, tip))),
                ],
              ),
            ),
          if (!kana.isCombo) ...[
            const SizedBox(height: 8),
            if (practiced.contains(kana.char))
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  children: [
                    Icon(Icons.check_circle, color: theme.colorScheme.primary),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        'Sudah bisa ditulis dari ingatan.'.localized(context),
                      ),
                    ),
                  ],
                ),
              ),
            FilledButton.icon(
              key: const Key('kana-start-practice'),
              onPressed: () =>
                  context.push(AppRoutes.kanaPracticePath(kana.char)),
              icon: const Icon(Icons.draw_rounded),
              label: Text('Latihan menulis'.localized(context)),
            ),
          ],
          const SizedBox(height: 20),
          Text(
            'Data urutan goresan: KanjiVG (CC BY-SA 3.0).'.localized(context),
            textAlign: TextAlign.center,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }

  List<Widget> _comboParts(
    BuildContext context,
    Kana kana,
    Set<String> practiced,
  ) => [
    Text(
      'Tulis per huruf. Huruf kedua ditulis kecil di pojok kiri bawah kotak berikutnya.'
          .localized(context),
    ),
    const SizedBox(height: 12),
    for (final part in kana.parts)
      Card(
        child: ListTile(
          key: Key('kana-part-$part'),
          leading: Text(part, style: AppTypography.japanese(fontSize: 28)),
          title: Text(KanaChart.of(scriptOf(part)).find(part)?.romaji ?? part),
          trailing: practiced.contains(part)
              ? const Icon(Icons.check_circle_outline_rounded)
              : const Icon(Icons.chevron_right_rounded),
          onTap: () => context.push(AppRoutes.kanaCharPath(part)),
        ),
      ),
  ];
}
