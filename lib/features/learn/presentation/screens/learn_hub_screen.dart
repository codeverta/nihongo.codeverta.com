import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

class LearnHubScreen extends ConsumerStatefulWidget {
  const LearnHubScreen({super.key});

  @override
  ConsumerState<LearnHubScreen> createState() => _LearnHubScreenState();
}

class _LearnHubScreenState extends ConsumerState<LearnHubScreen> {
  int _selectedLevel = 0;

  static const _levels = ['N5', 'N4', 'N3', 'N2', 'N1'];
  static const _numerals = ['五', '四', '三', '二', '一'];
  static const _subjects = [
    ('Vocabulary', 'Kosakata, contoh, kuis & progress', 'vocabulary'),
    ('Grammar', 'Pola, contoh kalimat, kuis & progress', 'grammar'),
    ('Kanji', 'Arti, bacaan, contoh kata & kuis', 'kanji'),
    ('Reading', 'Cerita bergambar, kosakata & kuis', 'reading'),
  ];

  VoidCallback? _subjectAction(int index, {required bool locked}) {
    final level = _levels[_selectedLevel];
    if (locked) return () => context.push('/premium');
    return switch (index) {
      0 => () => context.push('${AppRoutes.vocabularyPath}?level=$level'),
      1 when _selectedLevel <= 4 => () => context.push(
        '${AppRoutes.grammarPath}?level=$level',
      ),
      2 when _selectedLevel <= 4 => () => context.push(
        '${AppRoutes.kanjiPath}?level=$level',
      ),
      3 when _selectedLevel <= 4 => () => context.push(
        AppRoutes.readingCoursePath(level),
      ),
      _ => null,
    };
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // Every signed-in account can study every level (no Lifetime Access needed).
    final premium = ref.watch(authControllerProvider).isAuthenticated;
    final locked = _selectedLevel >= 2 && !premium;
    final level = _levels[_selectedLevel];
    return CustomScrollView(
      key: const PageStorageKey('learn-scroll'),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
          sliver: SliverList.list(
            children: [
              StaggeredEntrance(
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.baseline,
                  textBaseline: TextBaseline.alphabetic,
                  children: [
                    Flexible(
                      child: Text(
                        'Belajar'.localized(context),
                        style: theme.textTheme.headlineMedium?.copyWith(
                          fontWeight: FontWeight.w800,
                          letterSpacing: -0.8,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '勉強',
                      style: AppTypography.japanese(
                        fontSize: 14,
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 6),
              StaggeredEntrance(
                child: Text(
                  'Semua materi N5–N1 gratis. Masuk untuk mulai belajar.'
                      .localized(context),
                ),
              ),
              const SizedBox(height: 22),
              StaggeredEntrance(
                index: 1,
                child: Text(
                  'Pilih level JLPT'.localized(context),
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(height: 10),
              StaggeredEntrance(
                index: 1,
                child: _LevelSelector(
                  levels: _levels,
                  selected: _selectedLevel,
                  lockedFrom: premium ? null : 2,
                  onSelected: (index) =>
                      setState(() => _selectedLevel = index),
                ),
              ),
              const SizedBox(height: 18),
              StaggeredEntrance(
                index: 2,
                child: AnimatedSwitcher(
                  duration: AppMotion.of(context, AppMotion.slow),
                  // Kurva tanpa overshoot: FadeTransition tidak boleh > 1.
                  switchInCurve: AppMotion.standard,
                  switchOutCurve: AppMotion.exit,
                  transitionBuilder: (child, animation) => FadeTransition(
                    opacity: animation,
                    child: ScaleTransition(
                      scale: Tween(begin: 0.96, end: 1.0).animate(animation),
                      child: child,
                    ),
                  ),
                  child: locked
                      ? _LockedLevelCard(
                          key: ValueKey('locked-$level'),
                          level: level,
                          onPressed: () => context.push('/premium'),
                        )
                      : _VocabularyCallout(
                          key: ValueKey('callout-$level'),
                          level: level,
                          numeral: _numerals[_selectedLevel],
                          onPressed: () => context.push(
                            '${AppRoutes.vocabularyPath}?level=$level',
                          ),
                        ),
                ),
              ),
              const SizedBox(height: 22),
              Text(
                '${'Materi'.localized(context)} $level',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 12),
              for (final (index, subject) in _subjects.indexed)
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: StaggeredEntrance(
                    index: index + 3,
                    replayKey: _selectedLevel,
                    child: _SubjectTile(
                      title: subject.$1.localized(context),
                      subtitle: locked
                          ? 'Lifetime Access diperlukan'.localized(context)
                          : subject.$2.localized(context),
                      categoryId: subject.$3,
                      locked: locked,
                      onTap: _subjectAction(index, locked: locked),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

/// Pemilih level bersegmen dengan knob merah yang bergeser.
class _LevelSelector extends StatelessWidget {
  const _LevelSelector({
    required this.levels,
    required this.selected,
    required this.onSelected,
    this.lockedFrom,
  });

  final List<String> levels;
  final int selected;
  final ValueChanged<int> onSelected;
  final int? lockedFrom;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final dark = Theme.of(context).brightness == Brightness.dark;
    final knobColor = dark ? AppTheme.darkPrimary : AppTheme.primaryColor;
    return Container(
      height: 46,
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: scheme.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(999),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final width = constraints.maxWidth / levels.length;
          return Stack(
            children: [
              AnimatedPositioned(
                duration: AppMotion.of(context, AppMotion.slow),
                curve: AppMotion.spring,
                left: width * selected,
                top: 0,
                bottom: 0,
                width: width,
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    color: knobColor,
                    borderRadius: BorderRadius.circular(999),
                    boxShadow: [
                      BoxShadow(
                        color: knobColor.withValues(alpha: 0.3),
                        blurRadius: 12,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                ),
              ),
              Row(
                children: [
                  for (var i = 0; i < levels.length; i++)
                    Expanded(
                      child: Semantics(
                        button: true,
                        selected: i == selected,
                        child: InkWell(
                          key: Key('learn-level-${levels[i]}'),
                          borderRadius: BorderRadius.circular(999),
                          onTap: () => onSelected(i),
                          child: Center(
                            child: AnimatedDefaultTextStyle(
                              duration: AppMotion.of(context, AppMotion.normal),
                              style: TextStyle(
                                fontWeight: FontWeight.w800,
                                fontSize: 13,
                                color: i == selected
                                    ? Colors.white
                                    : scheme.onSurfaceVariant,
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  if (lockedFrom != null &&
                                      i >= lockedFrom!) ...[
                                    Icon(
                                      Icons.lock_rounded,
                                      size: 11,
                                      color: i == selected
                                          ? Colors.white
                                          : scheme.onSurfaceVariant,
                                    ),
                                    const SizedBox(width: 2),
                                  ],
                                  Text(levels[i]),
                                ],
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }
}

class _VocabularyCallout extends StatelessWidget {
  const _VocabularyCallout({
    required this.level,
    required this.numeral,
    required this.onPressed,
    super.key,
  });

  final String level;
  final String numeral;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final sakura = SakuraColors.of(context);
    return Pressable(
      onTap: onPressed,
      scale: 0.98,
      borderRadius: BorderRadius.circular(22),
      child: Container(
        clipBehavior: Clip.antiAlias,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(22),
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [sakura.heroStart, sakura.heroEnd],
          ),
        ),
        child: Stack(
          children: [
            Positioned(
              right: -6,
              top: -22,
              child: Text(
                numeral,
                style: AppTypography.japanese(
                  fontSize: 104,
                  fontWeight: FontWeight.w700,
                  color: Colors.white.withValues(alpha: 0.08),
                  height: 1,
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(18),
              child: Row(
                children: [
                  Container(
                    width: 56,
                    height: 56,
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.14),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      '語',
                      style: AppTypography.japanese(
                        fontSize: 26,
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                        height: 1,
                      ),
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'JLPT $level',
                          style: TextStyle(
                            color: sakura.gold,
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            letterSpacing: 1.2,
                          ),
                        ),
                        const SizedBox(height: 3),
                        Text(
                          'Vocabulary JLPT $level',
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w800,
                            fontSize: 17,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Pelajari kata, contoh, lalu uji pemahamanmu'
                              .localized(context),
                          style: TextStyle(
                            color: Colors.white.withValues(alpha: 0.78),
                            fontSize: 12.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    key: const Key('open-vocabulary-course'),
                    onPressed: onPressed,
                    style: IconButton.styleFrom(
                      backgroundColor: Colors.white,
                      foregroundColor: sakura.heroStart,
                    ),
                    icon: const Icon(Icons.arrow_forward_rounded),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Kartu ajakan Lifetime Access saat level N3–N1 dipilih tanpa akses.
class _LockedLevelCard extends StatelessWidget {
  const _LockedLevelCard({
    required this.level,
    required this.onPressed,
    super.key,
  });

  final String level;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final sakura = SakuraColors.of(context);
    final theme = Theme.of(context);
    return Container(
      key: const Key('learn-locked-level'),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: sakura.goldSoft,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: sakura.gold.withValues(alpha: 0.45)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: sakura.gold,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(Icons.lock_rounded, color: sakura.onGold, size: 20),
          ),
          const SizedBox(height: 12),
          Text(
            '${'Buka materi'.localized(context)} $level',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'Materi N3–N1 memerlukan Lifetime Access. Sekali bayar, belajar selamanya.'
                .localized(context),
            style: theme.textTheme.bodySmall,
          ),
          const SizedBox(height: 14),
          SizedBox(
            width: double.infinity,
            child: FilledButton(
              key: const Key('learn-locked-premium'),
              onPressed: onPressed,
              style: FilledButton.styleFrom(
                backgroundColor: sakura.gold,
                foregroundColor: sakura.onGold,
                minimumSize: const Size.fromHeight(46),
              ),
              child: Text('Lihat paket'.localized(context)),
            ),
          ),
        ],
      ),
    );
  }
}

class _SubjectTile extends StatelessWidget {
  const _SubjectTile({
    required this.title,
    required this.subtitle,
    required this.categoryId,
    required this.onTap,
    this.locked = false,
  });

  final String title;
  final String subtitle;
  final String categoryId;
  final VoidCallback? onTap;
  final bool locked;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = SakuraColors.of(context).category(categoryId);
    return Pressable(
      onTap: onTap,
      scale: 0.98,
      borderRadius: BorderRadius.circular(18),
      child: SakuraCard(
        padding: const EdgeInsets.fromLTRB(12, 12, 10, 12),
        child: Row(
          children: [
            Opacity(
              opacity: locked ? 0.55 : 1,
              child: KanjiGlyphTile(
                glyph: categoryGlyph(categoryId),
                color: color,
                size: 46,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title.localized(context),
                    style: theme.textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle.localized(context),
                    style: theme.textTheme.bodySmall,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            if (locked)
              Icon(
                Icons.lock_outline,
                color: SakuraColors.of(context).goldText,
              )
            else if (onTap == null)
              Chip(label: Text('Segera'.localized(context)))
            else
              Icon(
                Icons.chevron_right_rounded,
                color: theme.colorScheme.onSurfaceVariant,
              ),
          ],
        ),
      ),
    );
  }
}
