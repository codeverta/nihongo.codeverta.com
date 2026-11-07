import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
import 'package:nihongo_master/features/map/presentation/map_image.dart';
import 'package:nihongo_master/features/map/presentation/map_labels.dart';
import 'package:url_launcher/url_launcher.dart';

Uri googleMapsUri(MapSpot spot) => Uri.https('www.google.com', '/maps/search/', {
  'api': '1',
  'query': '${spot.lat},${spot.lng}',
});

Future<void> _open(String url) async {
  final uri = Uri.tryParse(url);
  if (uri == null || !(uri.scheme == 'https' || uri.scheme == 'http')) return;
  await launchUrl(uri, mode: LaunchMode.externalApplication);
}

class SpotDetailScreen extends ConsumerWidget {
  const SpotDetailScreen({required this.id, super.key});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(mapSpotsProvider);
    final counts =
        ref.watch(mapJobCountsProvider).valueOrNull ??
        const JobCounts(counts: {});
    final geometry = ref.watch(japanGeometryProvider).valueOrNull;
    return data.when(
      loading: () => Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (error, _) => Scaffold(
        appBar: AppBar(),
        body: Center(child: Text('$error'.localized(context))),
      ),
      data: (data) {
        final spot = data.spots.where((s) => s.id == id).firstOrNull;
        if (spot == null) {
          return Scaffold(
            appBar: AppBar(),
            body: Center(
              child: Text('Tempat ini tidak ditemukan.'.localized(context)),
            ),
          );
        }
        return _SpotDetail(
          spot: spot,
          jobCount: counts[spot.prefecture],
          prefectureName: geometry?.byCode(spot.prefecture)?.name ?? '',
        );
      },
    );
  }
}

class _SpotDetail extends StatelessWidget {
  const _SpotDetail({
    required this.spot,
    required this.jobCount,
    required this.prefectureName,
  });
  final MapSpot spot;
  final int jobCount;
  final String prefectureName;

  @override
  Widget build(BuildContext context) {
    final lang = mapLanguage(context);
    final theme = Theme.of(context);
    final summary = spot.summaryFor(lang);
    final tips = spot.tipsFor(lang);
    final phrases = spot.phrasesFor(lang);
    final image = spot.image;
    Widget heading(String text) => Padding(
      padding: const EdgeInsets.only(top: 24, bottom: 8),
      child: Text(
        text,
        style: theme.textTheme.titleMedium?.copyWith(
          fontWeight: FontWeight.w800,
        ),
      ),
    );
    Widget fallbackNote(LocalizedText text) => text.sourceLang == 'en' &&
            lang != 'en'
        ? Padding(
            padding: const EdgeInsets.only(bottom: 6),
            child: Text(
              'Teks asli dalam bahasa Inggris'.localized(context),
              key: const Key('english-fallback-note'),
              style: theme.textTheme.labelMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                fontStyle: FontStyle.italic,
              ),
            ),
          )
        : const SizedBox.shrink();
    return Scaffold(
      appBar: AppBar(title: Text(spot.nameFor(lang))),
      body: ListView(
        padding: const EdgeInsets.only(bottom: 32),
        children: [
          if (image != null) ...[
            AspectRatio(
              aspectRatio: 16 / 10,
              child: MapSpotImage(url: image.url),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 6, 20, 0),
              child: _Attribution(image: image),
            ),
          ],
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  spot.nameFor(lang),
                  style: theme.textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                if (spot.nameJa.isNotEmpty)
                  InkWell(
                    key: const Key('spot-japanese-name'),
                    onTap: () => context.push(
                      '/dictionary?q=${Uri.encodeQueryComponent(spot.nameJa)}',
                    ),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 4),
                      child: Wrap(
                        crossAxisAlignment: WrapCrossAlignment.end,
                        spacing: 8,
                        children: [
                          Text(
                            spot.nameJa,
                            style: theme.textTheme.titleLarge?.copyWith(
                              color: theme.colorScheme.primary,
                            ),
                          ),
                          if (spot.reading.isNotEmpty)
                            Text(
                              spot.reading,
                              style: theme.textTheme.bodyMedium,
                            ),
                          Icon(
                            Icons.menu_book_rounded,
                            size: 18,
                            color: theme.colorScheme.primary,
                          ),
                        ],
                      ),
                    ),
                  ),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  runSpacing: 6,
                  children: [
                    Chip(
                      avatar: Icon(
                        categoryIcon(spot.category),
                        size: 18,
                        color: categoryColor(spot.category),
                      ),
                      label: Text(categoryLabel(context, spot.category)),
                    ),
                    if (prefectureName.isNotEmpty)
                      Chip(
                        avatar: const Icon(Icons.map_outlined, size: 18),
                        label: Text(prefectureName),
                      ),
                  ],
                ),
                if (!summary.isEmpty) ...[
                  heading('Tentang tempat ini'.localized(context)),
                  fallbackNote(summary),
                  Text(summary.text, style: theme.textTheme.bodyLarge),
                ],
                if (!tips.isEmpty) ...[
                  heading('Tips berkunjung'.localized(context)),
                  fallbackNote(tips),
                  Text(tips.text, style: theme.textTheme.bodyLarge),
                ],
                if (phrases.isNotEmpty) ...[
                  heading('Kosakata & frasa'.localized(context)),
                  for (final phrase in phrases)
                    Card(
                      margin: const EdgeInsets.only(bottom: 8),
                      child: ListTile(
                        title: Text(
                          phrase.japanese,
                          style: theme.textTheme.titleMedium,
                        ),
                        subtitle: Text(
                          [
                            if (phrase.reading.isNotEmpty) phrase.reading,
                            phrase.meaningFor(lang)!,
                          ].join('\n'),
                        ),
                        trailing: IconButton(
                          tooltip: 'Cari di kamus'.localized(context),
                          icon: const Icon(Icons.menu_book_rounded),
                          onPressed: () => context.push(
                            '/dictionary?q=${Uri.encodeQueryComponent(phrase.japanese)}',
                          ),
                        ),
                      ),
                    ),
                ],
                const SizedBox(height: 24),
                FilledButton.icon(
                  key: const Key('open-google-maps'),
                  onPressed: () => unawaited(
                    launchUrl(
                      googleMapsUri(spot),
                      mode: LaunchMode.externalApplication,
                    ),
                  ),
                  icon: const Icon(Icons.directions_rounded),
                  label: Text('Buka di Google Maps'.localized(context)),
                ),
                if (jobCount > 0) ...[
                  const SizedBox(height: 8),
                  OutlinedButton.icon(
                    key: const Key('spot-prefecture-jobs'),
                    onPressed: () =>
                        context.push('/jobs?prefecture=${spot.prefecture}'),
                    icon: const Icon(Icons.work_rounded),
                    label: Text(
                      '${'Lowongan di prefektur ini'.localized(context)} ($jobCount)',
                    ),
                  ),
                ],
                if (spot.sources.isNotEmpty) ...[
                  heading('Sumber'.localized(context)),
                  for (final source in spot.sources)
                    InkWell(
                      onTap: () => unawaited(_open(source.url)),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 4),
                        child: Text(
                          '${_sourceName(source)} · ${source.license}',
                          style: theme.textTheme.bodySmall?.copyWith(
                            decoration: TextDecoration.underline,
                          ),
                        ),
                      ),
                    ),
                  const SizedBox(height: 4),
                  Text(
                    'Teks dari Wikipedia dan Wikivoyage berlisensi CC BY-SA 4.0; perubahan dapat dilakukan oleh tim Nihongo Master.'
                        .localized(context),
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  static String _sourceName(MapSource source) {
    final project = switch (source.kind) {
      'wikipedia' => 'Wikipedia',
      'wikivoyage' => 'Wikivoyage',
      'wikidata' => 'Wikidata',
      _ => source.kind,
    };
    return source.lang.isEmpty ? project : '$project (${source.lang})';
  }
}

class _Attribution extends StatelessWidget {
  const _Attribution({required this.image});
  final MapImage image;
  @override
  Widget build(BuildContext context) {
    final style = Theme.of(context).textTheme.bodySmall?.copyWith(
      color: Theme.of(context).colorScheme.onSurfaceVariant,
    );
    final text = [
      if (image.author.isNotEmpty) '© ${image.author}',
      if (image.license.isNotEmpty) image.license,
    ].join(' · ');
    final link = image.sourceUrl.isNotEmpty ? image.sourceUrl : image.licenseUrl;
    return InkWell(
      onTap: link.isEmpty ? null : () => unawaited(_open(link)),
      child: Text(
        '${'Foto'.localized(context)}: $text',
        style: style?.copyWith(
          decoration: link.isEmpty ? null : TextDecoration.underline,
        ),
      ),
    );
  }
}
