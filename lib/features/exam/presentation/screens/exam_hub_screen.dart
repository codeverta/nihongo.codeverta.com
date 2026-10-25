import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';

class ExamHubScreen extends ConsumerStatefulWidget {
  const ExamHubScreen({this.initialCategory = 'JLPT', super.key});

  final String initialCategory;

  @override
  ConsumerState<ExamHubScreen> createState() => _ExamHubScreenState();
}

class _ExamHubScreenState extends ConsumerState<ExamHubScreen> {
  late String category = widget.initialCategory;

  @override
  void didUpdateWidget(covariant ExamHubScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.initialCategory != widget.initialCategory) {
      category = widget.initialCategory;
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = ref.watch(authControllerProvider);
    if (!auth.isAuthenticated) return const SizedBox.shrink();
    final overview = ref.watch(diagnosticOverviewProvider);
    final premium = auth.user?.isPremium ?? false;
    final colors = Theme.of(context).colorScheme;
    return RefreshIndicator(
      onRefresh: () => ref.refresh(diagnosticOverviewProvider.future),
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          Text(
            'Latihan Ujian'.localized(context),
            style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 6),
          Text(
            'Pilih kategori ujian dan latihan yang ingin kamu ikuti.'.localized(
              context,
            ),
            style: TextStyle(color: colors.onSurfaceVariant),
          ),
          const SizedBox(height: 20),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final name in ['JLPT', 'NAT Test', 'JFT Test', 'SSW'])
                ChoiceChip(
                  label: Text(name),
                  selected: category == name,
                  onSelected: (_) => setState(() => category = name),
                ),
            ],
          ),
          const SizedBox(height: 20),
          if (category == 'SSW') const SswEntryCard(exams: true),
          if (category == 'NAT Test') ...[
            _SectionTitle('NAT-TEST 5Q'.localized(context)),
            for (var paper = 1; paper <= 5; paper++) ...[
              if (paper > 1) const Divider(height: 1),
              _ExamRow(
                badge: '5Q',
                imageAsset: 'assets/images/exams/nat.webp',
                title: 'NAT 5Q Exam $paper',
                metadata: '43 soal · 60 menit'.localized(context),
                description: 'Kosakata, kanji, grammar & reading'.localized(
                  context,
                ),
                note: 'Latihan tertulis · Tanpa listening'.localized(context),
                locked: !premium,
                onTap: () => context.push(
                  premium
                      ? paper == 1
                            ? '/exams/nat/simulation'
                            : '/exams/nat-$paper/simulation'
                      : '/premium',
                ),
              ),
            ],
          ],
          if (category == 'JFT Test') ...[
            _SectionTitle('JFT-Basic'.localized(context)),
            for (final paperNo in [1, 2, 3, 4, 5]) ...[
              if (paperNo > 1) const Divider(height: 1),
              _ExamRow(
                badge: 'JFT',
                imageAsset: 'assets/images/exams/jft.webp',
                title: 'JFT Exam $paperNo',
                metadata: '36 soal · 45 menit'.localized(context),
                description: 'Kosakata, percakapan & reading'.localized(
                  context,
                ),
                note: 'Soal orisinal · Tanpa listening'.localized(context),
                locked: !premium,
                onTap: () => context.push(
                  premium ? '/exams/jft/exam-$paperNo' : '/premium',
                ),
              ),
            ],
          ],
          if (category == 'JLPT') ...[
            _SectionTitle('Simulasi JLPT'.localized(context)),
            const SizedBox(height: 4),
            _ExamRow(
              badge: 'N5',
              imageAsset: 'assets/images/exams/n5.webp',
              title: 'Simulasi N5'.localized(context),
              metadata: '38 soal · 40 menit'.localized(context),
              description: 'Kosakata, kanji & grammar'.localized(context),
              note: 'Tanpa reading & listening'.localized(context),
              onTap: () => context.push('/exams/n5/simulation'),
            ),
            const Divider(height: 1),
            _ExamRow(
              badge: 'N4',
              imageAsset: 'assets/images/exams/n4.webp',
              title: 'Simulasi N4'.localized(context),
              metadata: '57 soal · 80 menit'.localized(context),
              description: 'Kosakata, kanji, grammar & reading'.localized(
                context,
              ),
              note: 'Tanpa listening'.localized(context),
              locked: !premium,
              onTap: () =>
                  context.push(premium ? '/exams/n4/simulation' : '/premium'),
            ),
            const Divider(height: 1),
            _ExamRow(
              badge: 'N3',
              imageAsset: 'assets/images/exams/n3.webp',
              title: 'Simulasi N3'.localized(context),
              metadata: '74 soal · 100 menit'.localized(context),
              description: 'Kosakata, kanji, grammar & reading'.localized(
                context,
              ),
              note: 'Tanpa listening'.localized(context),
              locked: !premium,
              onTap: () =>
                  context.push(premium ? '/exams/n3/simulation' : '/premium'),
            ),
            const Divider(height: 1),
            _ExamRow(
              badge: 'N2',
              imageAsset: 'assets/images/exams/n3.webp',
              title: 'Simulasi N2'.localized(context),
              metadata: '64 soal · 105 menit'.localized(context),
              description: 'Kosakata, grammar & reading tingkat lanjut'
                  .localized(context),
              note: 'Bagian tertulis · Tanpa listening'.localized(context),
              locked: !premium,
              onTap: () =>
                  context.push(premium ? '/exams/n2/simulation' : '/premium'),
            ),
            const Divider(height: 1),
            _ExamRow(
              badge: 'N1',
              imageAsset: 'assets/images/exams/n3.webp',
              title: 'Simulasi N1'.localized(context),
              metadata: '64 soal · 110 menit'.localized(context),
              description: 'Kosakata, grammar & reading tingkat mahir'
                  .localized(context),
              note: 'Bagian tertulis · Tanpa listening'.localized(context),
              locked: !premium,
              onTap: () =>
                  context.push(premium ? '/exams/n1/simulation' : '/premium'),
            ),
            const SizedBox(height: 20),
            _SectionTitle('Cek kemampuan awal'.localized(context)),
            const SizedBox(height: 4),
            _ExamRow(
              icon: Icons.insights_outlined,
              imageAsset: 'assets/images/exams/diagnostic-n5.webp',
              title: 'Diagnostic N5',
              description: 'Temukan bagian yang perlu kamu latih.'.localized(
                context,
              ),
              onTap: () => context.push('/exams/n5'),
            ),
            const SizedBox(height: 20),
            _SectionTitle('Riwayat diagnostic'.localized(context)),
            const SizedBox(height: 12),
            overview.when(
              loading: () => const Padding(
                padding: EdgeInsets.all(20),
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (e, _) => ExamMessage(
                message: diagnosticError(e).localized(context),
                retry: () => ref.invalidate(diagnosticOverviewProvider),
              ),
              data: (data) {
                final history = diagnosticList(data['history']);
                final activeAttemptId = data['activeAttemptId'] as String;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (activeAttemptId.isNotEmpty) ...[
                      Material(
                        color: examInk.withValues(alpha: .06),
                        borderRadius: BorderRadius.circular(12),
                        clipBehavior: Clip.antiAlias,
                        child: ListTile(
                          leading: const Icon(
                            Icons.play_circle_outline,
                            color: examInk,
                          ),
                          title: Text(
                            'Lanjutkan diagnostic N5'.localized(context),
                          ),
                          subtitle: Text(
                            'Jawabanmu sudah tersimpan.'.localized(context),
                          ),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () => context.push(
                            '/exams/n5/attempt/$activeAttemptId',
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                    ],
                    if (history.isEmpty)
                      Text(
                        'Belum ada hasil. Riwayat akan muncul setelah kamu menyelesaikan diagnostic.'
                            .localized(context),
                        style: TextStyle(
                          color: colors.onSurfaceVariant,
                          height: 1.5,
                        ),
                      ),
                    for (var i = 0; i < history.length; i++) ...[
                      if (i > 0) const Divider(height: 1),
                      ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: const Text(
                          'Diagnostic N5',
                          style: TextStyle(fontWeight: FontWeight.w600),
                        ),
                        subtitle: Text(
                          '${history[i]['correct']}/${history[i]['total']} benar · ${_date(context, history[i]['completedAt'] as String)}',
                        ),
                        trailing: Text(
                          '${history[i]['score']}%',
                          style: const TextStyle(
                            color: examGreen,
                            fontSize: 20,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                        onTap: () => context.push(
                          '/exams/n5/attempt/${history[i]['attemptId']}',
                        ),
                      ),
                    ],
                  ],
                );
              },
            ),
          ],
        ],
      ),
    );
  }

  String _date(BuildContext context, String value) => regionalDate(
    DateTime.parse(value),
    localeName: appRegionLocaleName(context),
  );
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Text(
    text.localized(context),
    style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
  );
}

class _ExamRow extends StatelessWidget {
  const _ExamRow({
    required this.title,
    required this.imageAsset,
    required this.description,
    required this.onTap,
    this.badge,
    this.icon,
    this.metadata,
    this.note,
    this.locked = false,
  });

  final String? badge;
  final IconData? icon;
  final String title;
  final String imageAsset;
  final String? metadata;
  final String description;
  final String? note;
  final bool locked;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Semantics(
      button: true,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 4),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: SizedBox(
                    width: 76,
                    height: 94,
                    child: Stack(
                      fit: StackFit.expand,
                      children: [
                        Image.asset(
                          imageAsset,
                          fit: BoxFit.cover,
                          cacheWidth: 256,
                          excludeFromSemantics: true,
                        ),
                        Positioned(
                          left: 6,
                          bottom: 6,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 7,
                              vertical: 3,
                            ),
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: .94),
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: badge != null
                                ? Text(
                                    badge!,
                                    style: const TextStyle(
                                      color: examInk,
                                      fontWeight: FontWeight.w800,
                                    ),
                                  )
                                : Icon(icon, color: examInk, size: 18),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title.localized(context),
                        style: const TextStyle(
                          fontSize: 17,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      if (metadata != null) ...[
                        const SizedBox(height: 4),
                        Text(
                          metadata!.localized(context),
                          style: const TextStyle(fontSize: 13),
                        ),
                      ],
                      const SizedBox(height: 4),
                      Text(
                        description.localized(context),
                        style: TextStyle(
                          fontSize: 13,
                          color: colors.onSurfaceVariant,
                        ),
                      ),
                      if (note != null)
                        Text(
                          note!.localized(context),
                          style: TextStyle(
                            fontSize: 12,
                            color: colors.onSurfaceVariant,
                          ),
                        ),
                      if (locked) ...[
                        const SizedBox(height: 6),
                        Text(
                          'Lifetime Access',
                          style: TextStyle(
                            color: colors.primary,
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Icon(
                    locked ? Icons.lock_outline : Icons.chevron_right,
                    size: 20,
                    color: colors.onSurfaceVariant,
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
