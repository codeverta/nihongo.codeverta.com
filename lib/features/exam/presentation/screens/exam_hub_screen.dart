import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

class ExamHubScreen extends ConsumerStatefulWidget {
  const ExamHubScreen({this.initialCategory = 'JLPT', super.key});

  final String initialCategory;

  @override
  ConsumerState<ExamHubScreen> createState() => _ExamHubScreenState();
}

class _ExamHubScreenState extends ConsumerState<ExamHubScreen> {
  late String category = widget.initialCategory;
  String _jlptLevel = 'N5';

  void _selectLevel(String level) {
    if (level == _jlptLevel) return;
    setState(() => _jlptLevel = level);
  }

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
    // Every signed-in account can open all exams (no Lifetime Access needed).
    final premium = auth.isAuthenticated;
    final colors = Theme.of(context).colorScheme;
    return RefreshIndicator(
      onRefresh: () => ref.refresh(diagnosticOverviewProvider.future),
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        physics: const AlwaysScrollableScrollPhysics(),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              Flexible(
                child: Text(
                  'Latihan Ujian'.localized(context),
                  style: const TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.w800,
                    letterSpacing: -0.8,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                '試験',
                style: AppTypography.japanese(
                  fontSize: 14,
                  color: colors.onSurfaceVariant,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            'Pilih kategori ujian dan latihan yang ingin kamu ikuti.'.localized(
              context,
            ),
            style: TextStyle(color: colors.onSurfaceVariant),
          ),
          const SizedBox(height: 20),
          _CategoryTabs(
            categories: const ['JLPT', 'NAT Test', 'JFT Test', 'SSW'],
            selected: category,
            onSelected: (name) => setState(() => category = name),
          ),
          if (category == 'JLPT') ...[
            const SizedBox(height: 14),
            _LevelJumpChips(
              selected: _jlptLevel,
              lockedFrom: premium ? null : 'N3',
              onSelected: _selectLevel,
            ),
          ],
          const SizedBox(height: 18),
          if (category == 'SSW') const SswEntryCard(exams: true),
          if (category == 'NAT Test') ...[
            _SectionTitle('NAT-TEST 5Q'.localized(context)),
            for (var paper = 1; paper <= 10; paper++) ...[
              if (paper > 1) const SizedBox(height: 10),
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
            for (final paperNo in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
              if (paperNo > 1) const SizedBox(height: 10),
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
            _SectionTitle(
              '${'Simulasi JLPT'.localized(context)} $_jlptLevel',
            ),
            const SizedBox(height: 4),
            if (_jlptLevel == 'N5') ...[
              KeyedSubtree(
                key: const ValueKey('exam-first-N5'),
                child: _ExamRow(
                  badge: 'N5',
                  imageAsset: 'assets/images/exams/n5.webp',
                  title: 'Simulasi N5'.localized(context),
                  metadata: '38 soal · 40 menit'.localized(context),
                  description: 'Kosakata, kanji & grammar'.localized(context),
                  note: 'Tanpa reading & listening'.localized(context),
                  onTap: () => context.push('/exams/n5/simulation'),
                ),
              ),
              for (final paper in [2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
                const SizedBox(height: 10),
                _ExamRow(
                  badge: 'N5',
                  imageAsset: 'assets/images/exams/n5.webp',
                  title:
                      '${'Simulasi N5'.localized(context)} · ${'Paket'.localized(context)} ${paper.toString().padLeft(2, '0')}',
                  metadata: '38 soal · 40 menit'.localized(context),
                  description: 'Kosakata, kanji & grammar · dengan furigana'
                      .localized(context),
                  onTap: () => context.push('/exams/n5-$paper/simulation'),
                ),
              ],
            ],
            if (_jlptLevel == 'N4') ...[
              KeyedSubtree(
                key: const ValueKey('exam-first-N4'),
                child: _ExamRow(
                  badge: 'N4',
                  imageAsset: 'assets/images/exams/n4.webp',
                  title: 'Simulasi N4'.localized(context),
                  metadata: '57 soal · 80 menit'.localized(context),
                  description: 'Kosakata, kanji, grammar & reading'.localized(
                    context,
                  ),
                  note: 'Tanpa listening'.localized(context),
                  onTap: () => context.push('/exams/n4/simulation'),
                ),
              ),
              for (final paper in [2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
                const SizedBox(height: 10),
                _ExamRow(
                  badge: 'N4',
                  imageAsset: 'assets/images/exams/n4.webp',
                  title:
                      '${'Simulasi N4'.localized(context)} · ${'Paket'.localized(context)} ${paper.toString().padLeft(2, '0')}',
                  metadata: '57 soal · 80 menit'.localized(context),
                  description: 'Kosakata, kanji, grammar & reading · dengan furigana'
                      .localized(context),
                  note: 'Gratis · Tanpa listening'.localized(context),
                  onTap: () => context.push('/exams/n4-$paper/simulation'),
                ),
              ],
            ],
            if (_jlptLevel == 'N3') ...[
              KeyedSubtree(
                key: const ValueKey('exam-first-N3'),
                child: _ExamRow(
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
              ),
              for (final paper in [2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
                const SizedBox(height: 10),
                _ExamRow(
                  badge: 'N3',
                  imageAsset: 'assets/images/exams/n3.webp',
                  title:
                      '${'Simulasi N3'.localized(context)} · ${'Paket'.localized(context)} ${paper.toString().padLeft(2, '0')}',
                  metadata: '74 soal · 100 menit'.localized(context),
                  description: 'Kosakata, kanji, grammar & reading · dengan furigana'
                      .localized(context),
                  note: 'Gratis · Tanpa listening'.localized(context),
                  onTap: () => context.push('/exams/n3-$paper/simulation'),
                ),
              ],
            ],
            if (_jlptLevel == 'N2') ...[
              KeyedSubtree(
                key: const ValueKey('exam-first-N2'),
                child: _ExamRow(
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
              ),
              for (final paper in [2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
                const SizedBox(height: 10),
                _ExamRow(
                  badge: 'N2',
                  imageAsset: 'assets/images/exams/n3.webp',
                  title:
                      '${'Simulasi N2'.localized(context)} · ${'Paket'.localized(context)} ${paper.toString().padLeft(2, '0')}',
                  metadata: '64 soal · 105 menit'.localized(context),
                  description: 'Kosakata, grammar & reading tingkat lanjut'
                      .localized(context),
                  note: 'Gratis · Tanpa listening'.localized(context),
                  onTap: () => context.push('/exams/n2-$paper/simulation'),
                ),
              ],
            ],
            if (_jlptLevel == 'N1') ...[
              KeyedSubtree(
                key: const ValueKey('exam-first-N1'),
                child: _ExamRow(
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
              ),
              for (final paper in [2, 3, 4, 5, 6, 7, 8, 9, 10]) ...[
                const SizedBox(height: 10),
                _ExamRow(
                  badge: 'N1',
                  imageAsset: 'assets/images/exams/n3.webp',
                  title:
                      '${'Simulasi N1'.localized(context)} · ${'Paket'.localized(context)} ${paper.toString().padLeft(2, '0')}',
                  metadata: '64 soal · 110 menit'.localized(context),
                  description: 'Kosakata, grammar & reading tingkat mahir'
                      .localized(context),
                  note: 'Gratis · Tanpa listening'.localized(context),
                  onTap: () => context.push('/exams/n1-$paper/simulation'),
                ),
              ],
            ],
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
                          '${history[i]['correct']}/${history[i]['total']} ${'benar'.localized(context)} · ${_date(context, history[i]['completedAt'] as String)}',
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

class _CategoryTabs extends StatelessWidget {
  const _CategoryTabs({
    required this.categories,
    required this.selected,
    required this.onSelected,
  });

  final List<String> categories;
  final String selected;
  final ValueChanged<String> onSelected;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final underline = Theme.of(context).brightness == Brightness.dark
        ? AppTheme.darkPrimaryText
        : AppTheme.primaryColor;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: colors.outlineVariant)),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            for (final name in categories)
              Semantics(
                button: true,
                selected: name == selected,
                child: InkWell(
                  key: Key('exam-category-$name'),
                  onTap: () => onSelected(name),
                  borderRadius: BorderRadius.circular(10),
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(4, 8, 18, 0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        AnimatedDefaultTextStyle(
                          duration: AppMotion.of(context, AppMotion.normal),
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w800,
                            color: name == selected
                                ? colors.onSurface
                                : colors.onSurfaceVariant,
                          ),
                          child: Text(name),
                        ),
                        const SizedBox(height: 8),
                        AnimatedContainer(
                          duration: AppMotion.of(context, AppMotion.slow),
                          // Tanpa overshoot: kurva spring membuat lebar
                          // sempat negatif (error BoxConstraints).
                          curve: AppMotion.standard,
                          height: 3,
                          width: name == selected ? 28 : 0,
                          decoration: BoxDecoration(
                            color: underline,
                            borderRadius: const BorderRadius.vertical(
                              top: Radius.circular(3),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Chip N5–N1 untuk memfilter daftar simulasi JLPT.
class _LevelJumpChips extends StatelessWidget {
  const _LevelJumpChips({
    required this.selected,
    required this.onSelected,
    this.lockedFrom,
  });

  final String selected;
  final ValueChanged<String> onSelected;
  final String? lockedFrom;

  static const _levels = ['N5', 'N4', 'N3', 'N2', 'N1'];

  @override
  Widget build(BuildContext context) {
    final lockIndex = lockedFrom == null ? null : _levels.indexOf(lockedFrom!);
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          for (final (index, level) in _levels.indexed)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: _LevelPill(
                key: Key('exam-level-$level'),
                label: level,
                selected: level == selected,
                locked: lockIndex != null && index >= lockIndex,
                onTap: () => onSelected(level),
              ),
            ),
        ],
      ),
    );
  }
}

class _LevelPill extends StatelessWidget {
  const _LevelPill({
    required this.label,
    required this.selected,
    required this.locked,
    required this.onTap,
    super.key,
  });

  final String label;
  final bool selected;
  final bool locked;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final dark = Theme.of(context).brightness == Brightness.dark;
    final background = selected
        ? (dark ? const Color(0xFFA9B0FF) : AppTheme.secondaryColor)
        : (dark ? AppTheme.darkSurface2 : Colors.white);
    final foreground = selected
        ? (dark ? AppTheme.darkBackground : Colors.white)
        : scheme.onSurface;
    return Semantics(
      button: true,
      selected: selected,
      child: Pressable(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: AnimatedContainer(
          duration: AppMotion.of(context, AppMotion.normal),
          curve: AppMotion.standard,
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: background,
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? background : scheme.outlineVariant,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              AnimatedSize(
                duration: AppMotion.of(context, AppMotion.normal),
                curve: AppMotion.standard,
                child: selected
                    ? Padding(
                        padding: const EdgeInsets.only(right: 4),
                        child: Icon(
                          Icons.check_rounded,
                          size: 16,
                          color: foreground,
                        ),
                      )
                    : const SizedBox.shrink(),
              ),
              Text(
                label,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w800,
                  color: foreground,
                ),
              ),
              if (locked) ...[
                const SizedBox(width: 3),
                Icon(
                  Icons.lock_outline_rounded,
                  size: 13,
                  color: foreground.withValues(alpha: 0.8),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
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
    final sakura = SakuraColors.of(context);
    final free = note?.startsWith('Gratis') ?? false;
    return StaggeredEntrance(
      offset: 8,
      replayKey: title,
      child: Semantics(
        button: true,
        child: Pressable(
          onTap: onTap,
          scale: 0.98,
          borderRadius: BorderRadius.circular(18),
          child: SakuraCard(
            padding: const EdgeInsets.all(12),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: SizedBox(
                    width: 64,
                    height: 80,
                    child: Stack(
                      fit: StackFit.expand,
                      children: [
                        ColorFiltered(
                          colorFilter: locked
                              ? const ColorFilter.matrix(_grayscale)
                              : const ColorFilter.mode(
                                  Colors.transparent,
                                  BlendMode.dst,
                                ),
                          child: Image.asset(
                            imageAsset,
                            fit: BoxFit.cover,
                            cacheWidth: 256,
                            excludeFromSemantics: true,
                          ),
                        ),
                        Positioned(
                          left: 5,
                          bottom: 5,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 6,
                              vertical: 2,
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
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                    ),
                                  )
                                : Icon(icon, color: examInk, size: 16),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title.localized(context),
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                          height: 1.25,
                        ),
                      ),
                      if (metadata != null) ...[
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            Icon(
                              Icons.schedule_rounded,
                              size: 13,
                              color: colors.onSurfaceVariant,
                            ),
                            const SizedBox(width: 4),
                            Flexible(
                              child: Text(
                                metadata!.localized(context),
                                style: TextStyle(
                                  fontSize: 12,
                                  color: colors.onSurfaceVariant,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                      const SizedBox(height: 4),
                      Text(
                        description.localized(context),
                        style: TextStyle(fontSize: 12.5, color: colors.onSurface),
                      ),
                      if (note != null || locked) ...[
                        const SizedBox(height: 6),
                        Wrap(
                          spacing: 4,
                          runSpacing: 4,
                          children: [
                            if (locked)
                              _Tag(
                                text: 'Lifetime Access',
                                background: sakura.goldSoft,
                                foreground: sakura.goldText,
                              ),
                            if (note != null)
                              _Tag(
                                text: note!.localized(context),
                                background: free
                                    ? sakura.successSoft
                                    : colors.surfaceContainerHigh,
                                foreground: free
                                    ? sakura.success
                                    : colors.onSurfaceVariant,
                              ),
                          ],
                        ),
                      ],
                    ],
                  ),
                ),
                const SizedBox(width: 6),
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: locked
                      ? Container(
                          width: 26,
                          height: 26,
                          decoration: BoxDecoration(
                            color: sakura.goldSoft,
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            Icons.lock_outline,
                            size: 14,
                            color: sakura.goldText,
                          ),
                        )
                      : Icon(
                          Icons.chevron_right,
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

class _Tag extends StatelessWidget {
  const _Tag({
    required this.text,
    required this.background,
    required this.foreground,
  });

  final String text;
  final Color background;
  final Color foreground;

  @override
  Widget build(BuildContext context) => DecoratedBox(
    decoration: BoxDecoration(
      color: background,
      borderRadius: BorderRadius.circular(6),
    ),
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 10.5,
          fontWeight: FontWeight.w700,
          color: foreground,
        ),
      ),
    ),
  );
}

const _grayscale = <double>[
  0.33, 0.59, 0.11, 0, 0, //
  0.33, 0.59, 0.11, 0, 0, //
  0.33, 0.59, 0.11, 0, 0, //
  0, 0, 0, 1, 0, //
];
