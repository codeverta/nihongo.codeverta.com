import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/shared/presentation/widgets/auth_gate.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({
    required this.openLearn,
    required this.openExam,
    required this.openProfile,
    super.key,
  });

  final VoidCallback openLearn;
  final VoidCallback openExam;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dashboard = ref.watch(homeDashboardProvider);
    return dashboard.when(
      loading: () => const _LoadingView(),
      error: (error, stackTrace) => _ErrorView(
        message: error.toString(),
        retry: () => ref.invalidate(homeDashboardProvider),
      ),
      data: (data) => RefreshIndicator(
        onRefresh: () => ref.refresh(homeDashboardProvider.future),
        child: _DashboardContent(
          data: data,
          openLearn: () => _authorize(context, openLearn),
          openExam: () => _authorize(context, openExam),
          openProfile: () => _authorize(context, openProfile),
        ),
      ),
    );
  }

  Future<void> _authorize(
    BuildContext context,
    VoidCallback destination,
  ) async {
    if (await AuthGate.requireLogin(context) && context.mounted) destination();
  }
}

class _DashboardContent extends StatelessWidget {
  const _DashboardContent({
    required this.data,
    required this.openLearn,
    required this.openExam,
    required this.openProfile,
  });

  final HomeDashboard data;
  final VoidCallback openLearn;
  final VoidCallback openExam;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context) {
    final lifetime = data.plans.where((plan) => plan.billingType == 'lifetime');
    final premiumPlan = lifetime.isNotEmpty
        ? lifetime.first
        : data.plans.where((plan) => plan.recommended).firstOrNull;
    final jlptTracks = data.tracks
        .where((track) => track.system.toUpperCase() == 'JLPT')
        .toList(growable: false);

    return CustomScrollView(
      key: const PageStorageKey('dashboard-scroll'),
      physics: const AlwaysScrollableScrollPhysics(),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 32),
          sliver: SliverList.list(
            children: [
              _DashboardHeader(data: data, openProfile: openProfile),
              if (premiumPlan != null && data.profile?.tier != 'premium') ...[
                const SizedBox(height: 22),
                _PremiumBanner(plan: premiumPlan, onPressed: openProfile),
              ],
              const SizedBox(height: 22),
              _DailyGoalCard(goal: data.dailyGoal, onPressed: openLearn),
              const SizedBox(height: 28),
              _SectionHeader(
                title: 'Mulai belajar',
                action: 'Lihat semua',
                onPressed: openLearn,
              ),
              const SizedBox(height: 14),
              _LearningMenuGrid(
                categories: data.categories,
                onPressed: openLearn,
              ),
              if (jlptTracks.isNotEmpty) ...[
                const SizedBox(height: 28),
                _SectionHeader(
                  title: 'Jalur JLPT',
                  action: 'Detail level',
                  onPressed: openLearn,
                ),
                const SizedBox(height: 14),
                _JlptLevelList(tracks: jlptTracks, onPressed: openLearn),
              ],
              if (data.exams.isNotEmpty) ...[
                const SizedBox(height: 28),
                _SectionHeader(
                  title: 'Latihan ujian',
                  action: 'Semua ujian',
                  onPressed: openExam,
                ),
                const SizedBox(height: 14),
                _ExamQuickList(exams: data.exams, onPressed: openExam),
              ],
              if (data.achievements.isNotEmpty) ...[
                const SizedBox(height: 28),
                _AchievementPreview(
                  achievement: _nextAchievement(data.achievements),
                  onPressed: openProfile,
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  Achievement _nextAchievement(List<Achievement> achievements) {
    return achievements.reduce((current, next) {
      final currentRemaining = current.target - current.progress;
      final nextRemaining = next.target - next.progress;
      return nextRemaining < currentRemaining ? next : current;
    });
  }
}

class _DashboardHeader extends StatelessWidget {
  const _DashboardHeader({required this.data, required this.openProfile});

  final HomeDashboard data;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context) {
    final profile = data.profile;
    final name = profile?.displayName.trim();
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                name == null || name.isEmpty
                    ? 'Selamat datang!'
                    : 'おはよう, $name!',
                key: const Key('home-greeting'),
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
              ),
              const SizedBox(height: 5),
              Text(
                data.isAuthenticated
                    ? '${profile?.target ?? ''} • Level ${profile?.level ?? 1} • ${profile?.xp ?? 0} XP'
                    : 'Jelajahi materi gratis, lalu masuk untuk menyimpan progress.',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            ],
          ),
        ),
        if (profile != null)
          InkWell(
            onTap: openProfile,
            customBorder: const CircleBorder(),
            child: CircleAvatar(
              radius: 24,
              backgroundImage: profile.photoUrl?.isNotEmpty ?? false
                  ? NetworkImage(profile.photoUrl!)
                  : null,
              child: profile.photoUrl?.isNotEmpty ?? false
                  ? null
                  : Text(
                      profile.displayName.isEmpty
                          ? 'N'
                          : profile.displayName.characters.first.toUpperCase(),
                      style: const TextStyle(fontWeight: FontWeight.w800),
                    ),
            ),
          ),
      ],
    );
  }
}

class _PremiumBanner extends StatelessWidget {
  const _PremiumBanner({required this.plan, required this.onPressed});

  final SubscriptionPlan plan;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final feature = plan.features.isEmpty ? plan.name : plan.features.first;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(24),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [colors.secondary, const Color(0xFF3538A8)],
        ),
        boxShadow: [
          BoxShadow(
            color: colors.secondary.withValues(alpha: 0.22),
            blurRadius: 24,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(Icons.auto_awesome_rounded, color: Color(0xFFFFD166)),
              SizedBox(width: 8),
              Text(
                'NIHONGO PREMIUM',
                style: TextStyle(
                  color: Color(0xFFFFE7A3),
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.1,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            plan.name,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 5),
          Text(
            '$feature • ${_rupiah(plan.priceIdr)}',
            style: TextStyle(color: Colors.white.withValues(alpha: 0.78)),
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: onPressed,
            style: FilledButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: colors.secondary,
            ),
            child: const Text('Lihat paket'),
          ),
        ],
      ),
    );
  }
}

class _DailyGoalCard extends StatelessWidget {
  const _DailyGoalCard({required this.goal, required this.onPressed});

  final HomeDailyGoal goal;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final percent = (goal.completion * 100).round();
    return Card(
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Row(
            children: [
              SizedBox.square(
                dimension: 62,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    CircularProgressIndicator(
                      value: goal.completion,
                      strokeWidth: 7,
                      strokeCap: StrokeCap.round,
                      backgroundColor: colors.primary.withValues(alpha: 0.1),
                    ),
                    Text(
                      '$percent%',
                      style: const TextStyle(fontWeight: FontWeight.w800),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Target harian',
                      style: Theme.of(context)
                          .textTheme
                          .titleMedium
                          ?.copyWith(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${goal.minutesLearned} dari ${goal.targetMinutes} menit selesai',
                    ),
                    const SizedBox(height: 7),
                    Text(
                      '${goal.streakDays} hari beruntun',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                            color: const Color(0xFFE86A33),
                            fontWeight: FontWeight.w700,
                          ),
                    ),
                  ],
                ),
              ),
              Icon(Icons.chevron_right_rounded, color: colors.outline),
            ],
          ),
        ),
      ),
    );
  }
}

class _LearningMenuGrid extends StatelessWidget {
  const _LearningMenuGrid({required this.categories, required this.onPressed});

  final List<LearningCategory> categories;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final columns = constraints.maxWidth >= 650 ? 3 : 2;
        return GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: categories.length,
          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: columns,
            crossAxisSpacing: 12,
            mainAxisSpacing: 12,
            childAspectRatio: 1.48,
          ),
          itemBuilder: (context, index) => _LearningMenuCard(
            category: categories[index],
            onPressed: onPressed,
          ),
        );
      },
    );
  }
}

class _LearningMenuCard extends StatelessWidget {
  const _LearningMenuCard({required this.category, required this.onPressed});

  final LearningCategory category;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final color = _categoryColor(category.id);
    return Card(
      child: InkWell(
        key: Key('category-${category.id}'),
        onTap: onPressed,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(13),
          child: Row(
            children: [
              Container(
                width: 45,
                height: 45,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Icon(_categoryIcon(category.id), color: color),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Flexible(
                          child: Text(
                            category.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                        if (category.isPremium) ...[
                          const SizedBox(width: 4),
                          const Icon(
                            Icons.workspace_premium_rounded,
                            size: 15,
                            color: Color(0xFFE59E24),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 3),
                    Text(
                      '${category.lessonCount} pelajaran',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _JlptLevelList extends StatelessWidget {
  const _JlptLevelList({required this.tracks, required this.onPressed});

  final List<LearningTrack> tracks;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return SizedBox(
      height: 142,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: tracks.length,
        separatorBuilder: (context, index) => const SizedBox(width: 12),
        itemBuilder: (context, index) {
          final track = tracks[index];
          final color = Color.lerp(
            primary,
            const Color(0xFF1D2088),
            index / tracks.length,
          )!;
          return InkWell(
            onTap: onPressed,
            borderRadius: BorderRadius.circular(20),
            child: Container(
              width: 140,
              padding: const EdgeInsets.all(15),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: color.withValues(alpha: 0.18)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        track.level,
                        style: TextStyle(
                          color: color,
                          fontSize: 24,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      if (track.isPremium)
                        const Icon(
                          Icons.lock_rounded,
                          size: 17,
                          color: Color(0xFFE59E24),
                        ),
                    ],
                  ),
                  Text(
                    track.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                  const Spacer(),
                  LinearProgressIndicator(
                    value: track.progress / 100,
                    borderRadius: BorderRadius.circular(99),
                    backgroundColor: color.withValues(alpha: 0.1),
                    color: color,
                  ),
                  const SizedBox(height: 7),
                  Text(
                    track.isPremium ? 'Premium' : '${track.progress}% selesai',
                    style: Theme.of(context)
                        .textTheme
                        .labelSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

class _ExamQuickList extends StatelessWidget {
  const _ExamQuickList({required this.exams, required this.onPressed});

  final List<HomeExam> exams;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 152,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: exams.length,
        separatorBuilder: (context, index) => const SizedBox(width: 10),
        itemBuilder: (context, index) => SizedBox(
          width: 164,
          child: Card(
            child: InkWell(
              onTap: onPressed,
              borderRadius: BorderRadius.circular(16),
              child: Padding(
                padding: const EdgeInsets.all(13),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(
                          Icons.assignment_rounded,
                          color: _examColor(exams[index].system),
                        ),
                        const Spacer(),
                        if (exams[index].isPremium)
                          const Icon(
                            Icons.workspace_premium_rounded,
                            size: 16,
                            color: Color(0xFFE59E24),
                          ),
                      ],
                    ),
                    const SizedBox(height: 9),
                    Text(
                      exams[index].title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontWeight: FontWeight.w800),
                    ),
                    const Spacer(),
                    Text(
                      '${exams[index].questionCount} soal • ${exams[index].durationMinutes} menit',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _AchievementPreview extends StatelessWidget {
  const _AchievementPreview({
    required this.achievement,
    required this.onPressed,
  });

  final Achievement achievement;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final remaining = (achievement.target - achievement.progress)
        .clamp(0, achievement.target);
    return Material(
      color: colors.primary.withValues(alpha: 0.055),
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(20),
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Row(
            children: [
              Icon(Icons.emoji_events_rounded, color: colors.primary, size: 42),
              const SizedBox(width: 15),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      achievement.name,
                      style: Theme.of(context)
                          .textTheme
                          .titleSmall
                          ?.copyWith(fontWeight: FontWeight.w800),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '$remaining langkah lagi • ${achievement.description}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right_rounded),
            ],
          ),
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({
    required this.title,
    required this.action,
    required this.onPressed,
  });

  final String title;
  final String action;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Text(
            title,
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                  letterSpacing: -0.25,
                ),
          ),
        ),
        TextButton(onPressed: onPressed, child: Text(action)),
      ],
    );
  }
}

class _LoadingView extends StatelessWidget {
  const _LoadingView();

  @override
  Widget build(BuildContext context) {
    return const Center(
      key: Key('home-loading'),
      child: CircularProgressIndicator(),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.retry});

  final String message;
  final VoidCallback retry;

  @override
  Widget build(BuildContext context) {
    return Center(
      key: const Key('home-error'),
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.cloud_off_rounded,
              size: 54,
              color: Theme.of(context).colorScheme.primary,
            ),
            const SizedBox(height: 14),
            Text(
              'Beranda belum dapat dimuat',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 7),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 18),
            FilledButton.icon(
              onPressed: retry,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Coba lagi'),
            ),
          ],
        ),
      ),
    );
  }
}

IconData _categoryIcon(String id) => switch (id) {
      'vocabulary' => Icons.translate_rounded,
      'grammar' => Icons.menu_book_rounded,
      'listening' => Icons.headphones_rounded,
      'reading' => Icons.chrome_reader_mode_rounded,
      'kanji' => Icons.brush_rounded,
      'speaking' => Icons.record_voice_over_rounded,
      _ => Icons.school_rounded,
    };

Color _categoryColor(String id) => switch (id) {
      'vocabulary' => const Color(0xFFE85555),
      'grammar' => const Color(0xFF5B67D7),
      'listening' => const Color(0xFF16A085),
      'reading' => const Color(0xFFE59E24),
      'kanji' => const Color(0xFF9B59B6),
      _ => const Color(0xFF2E86C1),
    };

Color _examColor(String system) => switch (system.toUpperCase()) {
      'JLPT' => const Color(0xFFE60012),
      'NAT' => const Color(0xFF1D2088),
      _ => const Color(0xFF168C72),
    };

String _rupiah(int value) {
  final digits = value.toString();
  final result = StringBuffer();
  for (var index = 0; index < digits.length; index++) {
    if (index > 0 && (digits.length - index) % 3 == 0) result.write('.');
    result.write(digits[index]);
  }
  return 'Rp$result';
}
