import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_visuals.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';

class SswEntryCard extends StatelessWidget {
  const SswEntryCard({super.key, this.exams = false});

  final bool exams;

  @override
  Widget build(BuildContext context) => ExploreFeatureCard(
    title: 'SSW / Tokutei Ginou',
    eyebrow: 'KETERAMPILAN & PERSIAPAN UJIAN'.localized(context),
    description: exams
        ? 'Latihan soal dan simulasi untuk ${sswSectors.length} bidang SSW'
              .localized(context)
        : 'Kenali bidangmu, pelajari materi, lalu coba simulasinya'.localized(
            context,
          ),
    asset: 'assets/images/ssw/ssw-careers.webp',
    color: const Color(0xFF125E60),
    icon: Icons.school_rounded,
    onTap: () => context.push(exams ? '/ssw?tab=test' : '/ssw'),
  );
}

class SswHubScreen extends StatefulWidget {
  const SswHubScreen({super.key, this.exams = false});

  final bool exams;

  @override
  State<SswHubScreen> createState() => _SswHubState();
}

class _SswHubState extends State<SswHubScreen> {
  String query = '';
  String filter = 'Semua';

  @override
  Widget build(BuildContext context) {
    final hasComingSoon = sswSectors.any((sector) => !sswAvailable(sector.id));
    final activeFilter = !hasComingSoon && filter == 'Segera hadir'
        ? 'Semua'
        : filter;
    final filtered = sswSectors
        .where(
          (sector) =>
              '${sector.name} ${sector.name.localized(context)} ${sector.japanese} ${sector.romaji} ${sector.keywords}'
                  .toLowerCase()
                  .contains(query.trim().toLowerCase()) &&
              (activeFilter == 'Semua' ||
                  (activeFilter == 'Tersedia' && sswAvailable(sector.id)) ||
                  (activeFilter == 'Segera hadir' &&
                      (!sswAvailable(sector.id)))),
        )
        .toList();
    return Scaffold(
      appBar: AppBar(title: Text('SSW / Tokutei Ginou')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const SswArtwork(),
          const SizedBox(height: 22),
          Text(
            'BELAJAR UNTUK DUNIA KERJA'.localized(context),
            style: TextStyle(
              color: Theme.of(context).colorScheme.primary,
              fontSize: 12,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.3,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Temukan bidangmu.\nBangun keterampilanmu.'.localized(context),
            style: Theme.of(
              context,
            ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 12),
          Text(
            widget.exams
                ? 'Latihan per topik dan simulasi belajar tersedia untuk ${sswSectors.length} bidang SSW. Paket aplikasi memakai soal orisinal dan dapat berbeda dari format tes resmi.'
                      .localized(context)
                : 'Kurikulum, kasus kerja, latihan per topik, dan simulasi belajar tersedia untuk ${sswSectors.length} bidang SSW. Mulai dari materi, lalu lanjutkan ke latihan dan simulasi.'
                      .localized(context),
            style: const TextStyle(height: 1.5),
          ),
          const SizedBox(height: 20),
          Card(
            color: Theme.of(context).colorScheme.primaryContainer,
            margin: EdgeInsets.zero,
            child: ListTile(
              contentPadding: const EdgeInsets.all(16),
              leading: Icon(
                Icons.volunteer_activism_rounded,
                color: Theme.of(context).colorScheme.onPrimaryContainer,
              ),
              title: Text(
                'Pelajari Kaigo'.localized(context),
                style: TextStyle(
                  fontWeight: FontWeight.w800,
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                ),
              ),
              subtitle: Text(
                'Kurikulum perawatan dan bahasa Jepang tersedia'.localized(
                  context,
                ),
                style: TextStyle(
                  height: 1.5,
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                ),
              ),
              trailing: Icon(
                Icons.arrow_forward_rounded,
                color: Theme.of(context).colorScheme.onPrimaryContainer,
              ),
              onTap: () => context.push(
                '/ssw/nursing-care${widget.exams ? '?tab=test' : ''}',
              ),
            ),
          ),
          const SizedBox(height: 20),
          TextField(
            onChanged: (value) => setState(() => query = value),
            decoration: InputDecoration(
              labelText: 'Cari bidang SSW'.localized(context),
              hintText: 'Kaigo, linen, manufaktur, konstruksi...'.localized(
                context,
              ),
              prefixIcon: Icon(Icons.search),
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final value in [
                'Semua',
                'Tersedia',
                if (hasComingSoon) 'Segera hadir',
              ])
                ChoiceChip(
                  label: Text(value.localized(context)),
                  selected: activeFilter == value,
                  onSelected: (_) => setState(() => filter = value),
                ),
            ],
          ),
          const SizedBox(height: 20),
          Text(
            '${filtered.length} ${'bidang'.localized(context)}',
            style: const TextStyle(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 8),
          if (filtered.isEmpty)
            Padding(
              padding: EdgeInsets.symmetric(vertical: 24),
              child: Text(
                'Bidang tidak ditemukan. Coba kata lain.'.localized(context),
              ),
            ),
          for (final sector in filtered)
            Card(
              child: ListTile(
                leading: CircleAvatar(
                  backgroundColor: sector.id == 'nursing-care'
                      ? Theme.of(context).colorScheme.primaryContainer
                      : Theme.of(context).colorScheme.secondaryContainer,
                  child: Icon(
                    sswSectorIcon(sector.id),
                    color: sector.id == 'nursing-care'
                        ? Theme.of(context).colorScheme.onPrimaryContainer
                        : Theme.of(context).colorScheme.onSecondaryContainer,
                  ),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                title: Text(
                  sector.name.localized(context),
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(sector.japanese),
                    const SizedBox(height: 2),
                    Text(
                      sector.romaji,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 6),
                    SswStatusBadge(available: sswAvailable(sector.id)),
                  ],
                ),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push(
                  '/ssw/${sector.id}${widget.exams ? '?tab=test' : ''}',
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class SswSectorScreen extends StatefulWidget {
  const SswSectorScreen({
    required this.sectorId,
    super.key,
    this.exams = false,
  });

  final String sectorId;
  final bool exams;

  @override
  State<SswSectorScreen> createState() => _SswSectorState();
}

class _SswSectorState extends State<SswSectorScreen> {
  late bool exams = widget.exams;

  @override
  void didUpdateWidget(covariant SswSectorScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.sectorId != widget.sectorId ||
        oldWidget.exams != widget.exams) {
      exams = widget.exams;
    }
  }

  @override
  Widget build(BuildContext context) {
    final sector = sswSectorById(widget.sectorId);
    return Scaffold(
      appBar: AppBar(title: Text('Bidang SSW'.localized(context))),
      body: sector == null
          ? Center(child: Text('Bidang tidak ditemukan.'.localized(context)))
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Text(
                  sector.name.localized(context),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  sector.japanese,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 4),
                Text(
                  sector.romaji,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 12),
                Align(
                  alignment: Alignment.centerLeft,
                  child: SswStatusBadge(available: sswAvailable(sector.id)),
                ),
                const SizedBox(height: 24),
                Wrap(
                  spacing: 8,
                  children: [
                    ChoiceChip(
                      label: Text('Materi'.localized(context)),
                      selected: !exams,
                      onSelected: (_) => setState(() => exams = false),
                    ),
                    ChoiceChip(
                      label: Text('Tes'.localized(context)),
                      selected: exams,
                      onSelected: (_) => setState(() => exams = true),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                if (exams && sswAvailable(sector.id)) ...[
                  SswAssessmentPanel(sector: sector.id),
                ] else if (exams) ...[
                  _FeaturePreview(
                    icon: Icons.quiz_outlined,
                    title: 'Latihan soal'.localized(context),
                    description: 'Latihan pemahaman keterampilan sesuai bidang.'
                        .localized(context),
                  ),
                  _FeaturePreview(
                    icon: Icons.timer_outlined,
                    title: 'Simulasi ujian'.localized(context),
                    description:
                        'Paket ujian dengan timer, hasil, dan pembahasan.'
                            .localized(context),
                  ),
                ] else if (sswAvailable(sector.id)) ...[
                  KaigoCurriculumPanel(sectorId: sector.id),
                ] else ...[
                  _FeaturePreview(
                    icon: Icons.menu_book_outlined,
                    title: 'Materi bidang'.localized(context),
                    description: 'Materi keterampilan untuk bidang pilihanmu.'
                        .localized(context),
                  ),
                  _FeaturePreview(
                    icon: Icons.translate,
                    title: 'Kosakata bidang'.localized(context),
                    description:
                        'Istilah bahasa Jepang yang digunakan di bidang ini.'
                            .localized(context),
                  ),
                ],
                const SizedBox(height: 12),
                if (!sswAvailable(sector.id))
                  Text(
                    'Konten belum tersedia. Kamu belum dapat memulai materi atau tes pada bidang ini.'
                        .localized(context),
                    style: TextStyle(height: 1.5),
                  ),
              ],
            ),
    );
  }
}

class _ComingSoonBadge extends StatelessWidget {
  const _ComingSoonBadge();

  @override
  Widget build(BuildContext context) => Text(
    'Coming Soon',
    style: TextStyle(
      color: Theme.of(context).colorScheme.primary,
      fontWeight: FontWeight.w700,
    ),
  );
}

class _FeaturePreview extends StatelessWidget {
  const _FeaturePreview({
    required this.icon,
    required this.title,
    required this.description,
  });

  final IconData icon;
  final String title;
  final String description;

  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: Theme.of(context).colorScheme.onSurfaceVariant),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 6),
                Text(description),
                const SizedBox(height: 8),
                const _ComingSoonBadge(),
              ],
            ),
          ),
        ],
      ),
    ),
  );
}
