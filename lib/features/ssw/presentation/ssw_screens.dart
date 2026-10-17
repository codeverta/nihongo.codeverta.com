import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_visuals.dart';

class SswEntryCard extends StatelessWidget {
  const SswEntryCard({super.key, this.exams = false});

  final bool exams;

  @override
  Widget build(BuildContext context) => Card(
        margin: EdgeInsets.zero,
        child: ListTile(
          contentPadding: const EdgeInsets.all(16),
          leading: Icon(
            Icons.work_outline,
            color: Theme.of(context).colorScheme.primary,
          ),
          title: const Text(
            'SSW / Tokutei Ginou',
            style: TextStyle(fontWeight: FontWeight.w800),
          ),
          subtitle: Text(
            exams
                ? 'Kaigo & Pembersihan Gedung · Latihan dan simulasi'
                : '2 bidang tersedia · 19 bidang',
          ),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => context.push(exams ? '/ssw?tab=test' : '/ssw'),
        ),
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
    final filtered = sswSectors
        .where(
          (sector) =>
              '${sector.name} ${sector.japanese} ${sector.keywords}'
                  .toLowerCase()
                  .contains(query.trim().toLowerCase()) &&
              (filter == 'Semua' ||
                  (filter == 'Tersedia' && sswAvailable(sector.id)) ||
                  (filter == 'Segera hadir' && (!sswAvailable(sector.id)))),
        )
        .toList();
    return Scaffold(
      appBar: AppBar(title: const Text('SSW / Tokutei Ginou')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const SswArtwork(),
          const SizedBox(height: 22),
          Text(
            'BELAJAR UNTUK DUNIA KERJA',
            style: TextStyle(
              color: Theme.of(context).colorScheme.primary,
              fontSize: 12,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.3,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Temukan bidangmu.\nBangun keterampilanmu.',
            style: Theme.of(context)
                .textTheme
                .headlineSmall
                ?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 12),
          Text(
            widget.exams
                ? 'Latihan per topik dan simulasi Kaigo serta Pembersihan Gedung tersedia. Ukur pemahamanmu dan tinjau pembahasannya.'
                : 'Pelajari Kaigo dan Pembersihan Gedung melalui kurikulum, kasus kerja, serta bahasa Jepang. Lanjutkan dengan latihan dan simulasi.',
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
                'Pelajari Kaigo',
                style: TextStyle(
                  fontWeight: FontWeight.w800,
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                ),
              ),
              subtitle: Text(
                'Kurikulum perawatan dan bahasa Jepang tersedia',
                style: TextStyle(
                  height: 1.5,
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                ),
              ),
              trailing: Icon(
                Icons.arrow_forward_rounded,
                color: Theme.of(context).colorScheme.onPrimaryContainer,
              ),
              onTap: () => context
                  .push('/ssw/nursing-care${widget.exams ? '?tab=test' : ''}'),
            ),
          ),
          const SizedBox(height: 20),
          TextField(
            onChanged: (value) => setState(() => query = value),
            decoration: const InputDecoration(
              labelText: 'Cari bidang SSW',
              hintText: 'Kaigo, restoran, pertanian...',
              prefixIcon: Icon(Icons.search),
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final value in ['Semua', 'Tersedia', 'Segera hadir'])
                ChoiceChip(
                  label: Text(value),
                  selected: filter == value,
                  onSelected: (_) => setState(() => filter = value),
                ),
            ],
          ),
          const SizedBox(height: 20),
          Text(
            '${filtered.length} bidang',
            style: const TextStyle(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 8),
          if (filtered.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 24),
              child: Text('Bidang tidak ditemukan. Coba kata lain.'),
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
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                title: Text(
                  sector.name,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(sector.japanese),
                    const SizedBox(height: 6),
                    SswStatusBadge(
                      available: sswAvailable(sector.id),
                    ),
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
    if (oldWidget.sectorId != widget.sectorId || oldWidget.exams != widget.exams) {
      exams = widget.exams;
    }
  }

  @override
  Widget build(BuildContext context) {
    final sector = sswSectorById(widget.sectorId);
    return Scaffold(
      appBar: AppBar(title: const Text('Bidang SSW')),
      body: sector == null
          ? const Center(child: Text('Bidang tidak ditemukan.'))
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Text(
                  sector.name,
                  style: Theme.of(context)
                      .textTheme
                      .headlineSmall
                      ?.copyWith(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 8),
                Text(
                  sector.japanese,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 12),
                Align(
                  alignment: Alignment.centerLeft,
                  child: SswStatusBadge(
                    available: sswAvailable(sector.id),
                  ),
                ),
                const SizedBox(height: 24),
                Wrap(
                  spacing: 8,
                  children: [
                    ChoiceChip(
                      label: const Text('Materi'),
                      selected: !exams,
                      onSelected: (_) => setState(() => exams = false),
                    ),
                    ChoiceChip(
                      label: const Text('Tes'),
                      selected: exams,
                      onSelected: (_) => setState(() => exams = true),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                if (exams && sswAvailable(sector.id)) ...[
                  SswAssessmentPanel(sector: sector.id),
                ] else if (exams) ...[
                  const _FeaturePreview(
                    icon: Icons.quiz_outlined,
                    title: 'Latihan soal',
                    description:
                        'Latihan pemahaman keterampilan sesuai bidang.',
                  ),
                  const _FeaturePreview(
                    icon: Icons.timer_outlined,
                    title: 'Simulasi ujian',
                    description:
                        'Paket ujian dengan timer, hasil, dan pembahasan.',
                  ),
                ] else if (sswAvailable(sector.id)) ...[
                  KaigoCurriculumPanel(sectorId: sector.id),
                ] else ...[
                  const _FeaturePreview(
                    icon: Icons.menu_book_outlined,
                    title: 'Materi bidang',
                    description: 'Materi keterampilan untuk bidang pilihanmu.',
                  ),
                  const _FeaturePreview(
                    icon: Icons.translate,
                    title: 'Kosakata bidang',
                    description:
                        'Istilah bahasa Jepang yang digunakan di bidang ini.',
                  ),
                ],
                const SizedBox(height: 12),
                if (!sswAvailable(sector.id))
                  const Text(
                    'Konten belum tersedia. Kamu belum dapat memulai materi atau tes pada bidang ini.',
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
