import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_banner_content.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_banner_carousel.dart';
import 'package:nihongo_master/features/jobs/presentation/jobs_screens.dart';
import 'package:nihongo_master/features/learn/activity/continue_learning_section.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:nihongo_master/features/manga/presentation/manga_screens.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/news/presentation/news_screens.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/profile_avatar.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
import 'package:nihongo_master/shared/presentation/widgets/auth_gate.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({
    required this.openLearn,
    required this.openExam,
    required this.openProfile,
    this.active = true,
    this.openExamCategory,
    super.key,
  });

  final VoidCallback openLearn;
  final VoidCallback openExam;
  final VoidCallback openProfile;
  final bool active;
  final ValueChanged<String>? openExamCategory;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dashboard = ref.watch(homeDashboardProvider);
    final accountHasLifetimeAccess = ref.watch(
      authControllerProvider.select(
        (state) =>
            (state.user?.isLifetime ?? false) ||
            (state.user?.isPremium ?? false),
      ),
    );
    return dashboard.when(
      loading: () => ListView(
        padding: const EdgeInsets.all(20),
        children: const [
          ContinueLearningSection(),
          SizedBox(height: 240, child: _LoadingView()),
        ],
      ),
      error: (error, stackTrace) => ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const ContinueLearningSection(),
          _ErrorView(
            message: error.toString(),
            retry: () {
              ref.invalidate(homeDashboardProvider);
              ref.read(learningHistoryProvider.notifier).refresh();
            },
          ),
        ],
      ),
      data: (data) => RefreshIndicator(
        onRefresh: () async {
          await ref.read(learningHistoryProvider.notifier).refresh();
          await (ref
                ..invalidate(newsPreviewProvider)
                ..invalidate(homeDashboardProvider))
              .read(homeDashboardProvider.future);
        },
        child: _DashboardContent(
          data: data,
          active: active,
          accountHasLifetimeAccess: accountHasLifetimeAccess,
          openLearn: () => _authorize(context, openLearn),
          openExam: () => _authorize(context, openExam),
          openExamCategory: (category) => _authorize(context, () {
            if (openExamCategory != null) {
              openExamCategory!(category);
            } else {
              openExam();
            }
          }),
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
    required this.active,
    required this.accountHasLifetimeAccess,
    required this.openExamCategory,
    required this.openLearn,
    required this.openExam,
    required this.openProfile,
  });

  final HomeDashboard data;
  final bool active;
  final ValueChanged<String> openExamCategory;
  final bool accountHasLifetimeAccess;
  final VoidCallback openLearn;
  final VoidCallback openExam;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context) {
    final hasLifetimeAccess =
        accountHasLifetimeAccess ||
        (data.profile?.isLifetime ?? false) ||
        data.profile?.tier == 'premium';
    final banners = homeBannersFor(data, lifetime: hasLifetimeAccess);
    return CustomScrollView(
      key: const PageStorageKey('dashboard-scroll'),
      physics: const AlwaysScrollableScrollPhysics(),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 32),
          sliver: SliverList.list(
            children: [
              _DashboardHeader(data: data, openProfile: openProfile),
              const SizedBox(height: 22),
              if (banners.isNotEmpty) ...[
                HomeBannerCarousel(
                  key: ValueKey(banners.map((banner) => banner.id).join(':')),
                  banners: banners,
                  active: active,
                  onOpen: (banner) {
                    switch (banner.action) {
                      case 'premium':
                        context.push('/premium');
                      case 'learn':
                        openLearn();
                      case 'anki':
                        context.push('/anki');
                      case 'manga':
                        context.push('/manga');
                      case 'ssw':
                        context.push('/ssw');
                      case 'jobs':
                        context.push('/jobs');
                    }
                  },
                ),
                const SizedBox(height: 16),
              ],
              const ContinueLearningSection(),
              _SectionHeader(
                title: 'Mulai belajar',
                action: 'Lihat semua',
                onPressed: openLearn,
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  ActionChip(label: const Text('JLPT'), onPressed: openLearn),
                  ActionChip(
                    label: const Text('JFT'),
                    onPressed: () => openExamCategory('JFT Test'),
                  ),
                  ActionChip(
                    label: const Text('NAT'),
                    onPressed: () => openExamCategory('NAT Test'),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              _LearningMenuGrid(
                categories: data.categories
                    .where((c) => c.id != 'listening' && c.id != 'speaking')
                    .toList(growable: false),
                onPressed: openLearn,
              ),
              const SizedBox(height: 12),
              Card(
                child: ListTile(
                  leading: const Icon(Icons.quiz_outlined),
                  title: const Text('Latihan ujian'),
                  subtitle: const Text('Simulasi JLPT, JFT, dan NAT'),
                  trailing: const Icon(Icons.chevron_right_rounded),
                  onTap: openExam,
                ),
              ),
              const SizedBox(height: 28),
              Text(
                'Jelajahi Jepang',
                style: Theme.of(
                  context,
                ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 14),
              LayoutBuilder(
                builder: (context, constraints) {
                  final columns = constraints.maxWidth >= 600 ? 2 : 1;
                  final width =
                      (constraints.maxWidth - (columns - 1) * 16) / columns;
                  return Wrap(
                    spacing: 16,
                    runSpacing: 16,
                    children: [
                      SizedBox(width: width, child: const MangaEntryCard()),
                      SizedBox(width: width, child: const AnkiEntryCard()),
                      SizedBox(width: width, child: const SswEntryCard()),
                      SizedBox(width: width, child: const JobsEntryCard()),
                    ],
                  );
                },
              ),

              const SizedBox(height: 24),
              const NewsPreview(),
            ],
          ),
        ),
      ],
    );
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
        IconButton(
          onPressed: () => context.push('/notifications'),
          tooltip: 'Pengaturan notifikasi',
          icon: const Icon(Icons.notifications_outlined),
        ),
        if (profile != null)
          InkWell(
            onTap: openProfile,
            customBorder: const CircleBorder(),
            child: ProfileAvatar(
              radius: 24,
              displayName: profile.displayName,
              photoUrl: profile.photoUrl,
            ),
          ),
      ],
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
          // Keep the nested grid's scroll state separate from the dashboard.
          key: const PageStorageKey('dashboard-learning-grid'),
          primary: false,
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
                            style: const TextStyle(fontWeight: FontWeight.w700),
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
              style: Theme.of(
                context,
              ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
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
