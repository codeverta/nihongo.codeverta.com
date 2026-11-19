import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/map/data/map_repository.dart';
import 'package:nihongo_master/features/map/domain/prefecture_profile.dart';
import 'package:nihongo_master/features/map/presentation/map_labels.dart';
import 'package:nihongo_master/features/map/presentation/prefecture_living_card.dart';
import 'package:url_launcher/url_launcher.dart';

String prefectureProfilePath(String code) => '/map/prefecture/$code';

Future<void> _openLink(String? url) async {
  final uri = Uri.tryParse(url ?? '');
  if (uri == null || !(uri.scheme == 'https' || uri.scheme == 'http')) return;
  await launchUrl(uri, mode: LaunchMode.externalApplication);
}

/// Short intro for the prefecture sheet with a link to the full profile.
class PrefectureProfileTeaser extends ConsumerWidget {
  const PrefectureProfileTeaser({required this.code, super.key});

  final String code;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(mapPrefectureProfilesProvider).valueOrNull?[code];
    final text = profile?.textFor(mapLanguage(context));
    if (profile == null || text == null) return const SizedBox.shrink();
    final theme = Theme.of(context);
    return Padding(
      key: const Key('prefecture-profile-teaser'),
      padding: const EdgeInsets.only(top: 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            text.intro,
            maxLines: 4,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.bodyMedium?.copyWith(height: 1.45),
          ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              key: const Key('prefecture-profile-open'),
              style: TextButton.styleFrom(padding: EdgeInsets.zero),
              onPressed: () => context.push(prefectureProfilePath(code)),
              icon: const Icon(Icons.menu_book_rounded, size: 18),
              label: Text('Baca profil lengkap'.localized(context)),
            ),
          ),
        ],
      ),
    );
  }
}

/// A Wikipedia-style page for one prefecture: names, an infobox of facts,
/// the article intro and a few sections, plus wages and attribution.
class PrefectureProfileScreen extends ConsumerWidget {
  const PrefectureProfileScreen({required this.code, super.key});

  final String code;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profiles = ref.watch(mapPrefectureProfilesProvider);
    final profile = profiles.valueOrNull?[code];
    if (profiles.isLoading && profile == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    if (profile == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.cloud_off_rounded, size: 40),
                const SizedBox(height: 12),
                Text(
                  'Profil prefektur belum dapat dimuat.'.localized(context),
                  textAlign: TextAlign.center,
                ),
                TextButton(
                  onPressed: () =>
                      ref.invalidate(mapPrefectureProfilesProvider),
                  child: Text('Coba lagi'.localized(context)),
                ),
              ],
            ),
          ),
        ),
      );
    }
    return _ProfileView(profile: profile);
  }
}

class _ProfileView extends StatelessWidget {
  const _ProfileView({required this.profile});

  final PrefectureProfile profile;

  @override
  Widget build(BuildContext context) {
    final lang = mapLanguage(context);
    final english = lang == 'en';
    final theme = Theme.of(context);
    final text = profile.textFor(lang);
    final fallback = text != null && text.lang != lang;
    final number = NumberFormat.decimalPattern(english ? 'en' : 'id');
    final population = profile.population;
    final year = profile.populationDate?.split('-').first;
    final area = profile.areaKm2;
    final density = profile.density;

    final facts = <(String, String)>[
      if (profile.capitalFor(lang) case final capital?)
        (
          'Ibu kota',
          '$capital${profile.capital['ja'] == null ? '' : ' (${profile.capital['ja']})'}',
        ),
      ('Wilayah', regionOf(profile.code)),
      if (population != null)
        (
          'Populasi',
          '${number.format(population)}${year == null ? '' : ' ($year)'}',
        ),
      if (area != null) ('Luas', '${number.format(area.round())} km²'),
      if (density != null)
        ('Kepadatan', '${number.format(density.round())} /km²'),
      ('Kode prefektur', 'JP-${profile.code}'),
    ];

    return Scaffold(
      appBar: AppBar(title: Text(profile.nameFor(lang))),
      body: SelectionArea(
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 32),
          children: [
            Text(
              profile.nameJa,
              key: const Key('prefecture-profile-name-ja'),
              style: AppTypography.japanese(
                fontSize: 34,
                fontWeight: FontWeight.w800,
                color: theme.colorScheme.onSurface,
              ),
            ),
            if (profile.nameKana.isNotEmpty)
              Text(
                profile.nameKana,
                style: AppTypography.japanese(
                  fontSize: 14,
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            const SizedBox(height: 16),
            _Infobox(facts: facts, website: profile.website),
            if (text != null) ...[
              const SizedBox(height: 20),
              if (fallback)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Text(
                    'Teks asli dalam bahasa Inggris'.localized(context),
                    key: const Key('prefecture-profile-english-note'),
                    style: theme.textTheme.labelMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ),
              Text(
                text.intro,
                style: theme.textTheme.bodyLarge?.copyWith(height: 1.55),
              ),
              for (final section in text.sections) ...[
                const SizedBox(height: 22),
                Text(
                  section.title,
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const Divider(height: 14),
                Text(
                  section.text,
                  style: theme.textTheme.bodyMedium?.copyWith(height: 1.55),
                ),
              ],
            ],
            PrefectureLivingCard(code: profile.code),
            const SizedBox(height: 24),
            _Attribution(text: text),
          ],
        ),
      ),
    );
  }
}

class _Infobox extends StatelessWidget {
  const _Infobox({required this.facts, required this.website});

  final List<(String, String)> facts;
  final String? website;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      key: const Key('prefecture-profile-infobox'),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerLow,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        children: [
          for (final (label, value) in facts)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 7),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: 118,
                    child: Text(
                      label.localized(context),
                      style: theme.textTheme.labelLarge?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      value,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          if (website != null && website!.isNotEmpty)
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                style: TextButton.styleFrom(padding: EdgeInsets.zero),
                onPressed: () => _openLink(website),
                icon: const Icon(Icons.open_in_new_rounded, size: 16),
                label: Text('Situs resmi prefektur'.localized(context)),
              ),
            ),
        ],
      ),
    );
  }
}

class _Attribution extends StatelessWidget {
  const _Attribution({required this.text});

  final ProfileText? text;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final style = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    return Column(
      key: const Key('prefecture-profile-attribution'),
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Teks dari Wikipedia (CC BY-SA 4.0), dipotong untuk aplikasi. Fakta dari Wikidata (CC0).'
              .localized(context),
          style: style,
        ),
        if (text != null && text!.url.isNotEmpty)
          TextButton.icon(
            key: const Key('prefecture-profile-wikipedia'),
            style: TextButton.styleFrom(padding: EdgeInsets.zero),
            onPressed: () => _openLink(text!.url),
            icon: const Icon(Icons.open_in_new_rounded, size: 16),
            label: Text(
              '${'Baca artikel lengkap di Wikipedia'.localized(context)} (${text!.lang})',
            ),
          ),
      ],
    );
  }
}
