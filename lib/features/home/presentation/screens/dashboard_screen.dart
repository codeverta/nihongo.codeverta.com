import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_banner_content.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_banner_carousel.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_quick_menu.dart';
import 'package:nihongo_master/features/learn/activity/continue_learning_section.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/news/presentation/news_screens.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/profile_avatar.dart';
import 'package:nihongo_master/shared/presentation/widgets/auth_gate.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

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
    final offline = ref.watch(offlineProvider);
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
            message: error.toString().localized(context),
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
          offline: offline || data.offline,
          active: active,
          accountHasLifetimeAccess: accountHasLifetimeAccess,
          openLearn: () => _authorize(context, openLearn),
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

class _OfflineDashboardNotice extends StatelessWidget {
  const _OfflineDashboardNotice();

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Container(
      key: const Key('offline-dashboard-notice'),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: colors.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          const Icon(Icons.cloud_off_rounded, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Offline · data terakhir tersimpan. Materi yang sudah disiapkan tetap bisa dibuka.'
                  .localized(context),
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        ],
      ),
    );
  }
}

class _DashboardContent extends StatelessWidget {
  const _DashboardContent({
    required this.data,
    required this.offline,
    required this.active,
    required this.accountHasLifetimeAccess,
    required this.openExamCategory,
    required this.openLearn,
    required this.openProfile,
  });

  final HomeDashboard data;
  final bool offline;
  final bool active;
  final ValueChanged<String> openExamCategory;
  final bool accountHasLifetimeAccess;
  final VoidCallback openLearn;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context) {
    // A stale dashboard is presentation data, never proof of premium access.
    final hasLifetimeAccess = accountHasLifetimeAccess;
    final banners = homeBannersFor(
      data,
      lifetime: hasLifetimeAccess,
      english: AppLocalizations.of(context).locale.languageCode == 'en',
    );
    return CustomScrollView(
      key: const PageStorageKey('dashboard-scroll'),
      physics: const AlwaysScrollableScrollPhysics(),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 32),
          sliver: SliverList.list(
            children: [
              StaggeredEntrance(
                child: _DashboardHeader(data: data, openProfile: openProfile),
              ),
              if (banners.isNotEmpty) ...[
                const SizedBox(height: 16),
                StaggeredEntrance(
                  index: 1,
                  child: HomeBannerCarousel(
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
                ),
              ],
              if (data.isAuthenticated && data.profile != null) ...[
                SizedBox(height: banners.isNotEmpty ? 8 : 16),
                StaggeredEntrance(
                  index: 2,
                  child: _ProgressSummary(
                    profile: data.profile!,
                    onTap: openProfile,
                  ),
                ),
              ],
              if (offline) ...[
                const SizedBox(height: 12),
                const _OfflineDashboardNotice(),
              ],
              const SizedBox(height: 20),
              HomeQuickMenu(
                openExamCategory: openExamCategory,
                hasLifetimeAccess: hasLifetimeAccess,
              ),
              const SizedBox(height: 22),
              const ContinueLearningSection(),
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
    final theme = Theme.of(context);
    final profile = data.profile;
    final name = profile?.displayName.trim();
    final hasName = name != null && name.isNotEmpty;
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (hasName)
                Text(
                  'おはよう',
                  style: AppTypography.japanese(
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              Text(
                hasName ? name! : 'Selamat datang!'.localized(context),
                key: const Key('home-greeting'),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w800,
                  letterSpacing: -0.6,
                  height: 1.15,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                data.isAuthenticated
                    ? [
                        if ((profile?.target ?? '').isNotEmpty)
                          profile!.target,
                        'Level ${profile?.level ?? 1}',
                      ].join(' • ')
                    : 'Jelajahi materi gratis, lalu masuk untuk menyimpan progress.'
                          .localized(context),
                style: theme.textTheme.bodySmall,
              ),
            ],
          ),
        ),
        _HeaderIconButton(
          onPressed: () => context.push('/notifications'),
          tooltip: 'Pengaturan notifikasi'.localized(context),
          icon: Icons.notifications_none_rounded,
        ),
        if (profile != null) ...[
          const SizedBox(width: 8),
          Pressable(
            onTap: openProfile,
            borderRadius: BorderRadius.circular(40),
            child: ProfileAvatar(
              radius: 22,
              displayName: profile.displayName,
              photoUrl: profile.photoUrl,
            ),
          ),
        ],
      ],
    );
  }
}

class _HeaderIconButton extends StatefulWidget {
  const _HeaderIconButton({
    required this.onPressed,
    required this.tooltip,
    required this.icon,
  });

  final VoidCallback onPressed;
  final String tooltip;
  final IconData icon;

  @override
  State<_HeaderIconButton> createState() => _HeaderIconButtonState();
}

class _HeaderIconButtonState extends State<_HeaderIconButton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ring = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 600),
  );

  @override
  void dispose() {
    _ring.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Tooltip(
      message: widget.tooltip,
      child: Pressable(
        borderRadius: BorderRadius.circular(40),
        semanticsLabel: widget.tooltip,
        onTap: () {
          if (!AppMotion.reduced(context)) _ring.forward(from: 0);
          widget.onPressed();
        },
        child: SakuraCard(
          padding: const EdgeInsets.all(10),
          radius: 40,
          child: AnimatedBuilder(
            animation: _ring,
            builder: (context, child) {
              final t = _ring.value;
              final angle = 0.3 * math.sin(t * math.pi * 5) * (1 - t);
              return Transform.rotate(angle: angle, child: child);
            },
            child: Icon(widget.icon, size: 20, color: scheme.onSurface),
          ),
        ),
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({required this.profile, required this.onTap});

  final HomeProfile profile;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final sakura = SakuraColors.of(context);
    final scheme = theme.colorScheme;
    Widget stat({
      required IconData icon,
      required Color color,
      required Widget value,
      required String label,
    }) {
      return Expanded(
        child: Row(
          children: [
            Container(
              width: 34,
              height: 34,
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.14),
                borderRadius: BorderRadius.circular(11),
              ),
              child: Icon(icon, size: 18, color: color),
            ),
            const SizedBox(width: 8),
            Flexible(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  DefaultTextStyle.merge(
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w800,
                      height: 1.1,
                    ),
                    child: value,
                  ),
                  Text(
                    label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: scheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    return Pressable(
      key: const Key('home-progress-summary'),
      onTap: onTap,
      scale: 0.98,
      borderRadius: BorderRadius.circular(18),
      child: SakuraCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                stat(
                  icon: Icons.local_fire_department_rounded,
                  color: scheme.primary,
                  value: AnimatedCount(value: profile.streakDays),
                  label: 'Streak hari'.localized(context),
                ),
                stat(
                  icon: Icons.bolt_rounded,
                  color: sakura.goldText,
                  value: AnimatedCount(value: profile.xp),
                  label: 'XP',
                ),
                stat(
                  icon: Icons.emoji_events_rounded,
                  color: sakura.grammar,
                  value: Text('${profile.level}'),
                  label: 'Level',
                ),
              ],
            ),
            const SizedBox(height: 10),
            Text(
              profile.streakDays > 0
                  ? 'Pertahankan streak dengan menyelesaikan kuis hari ini.'
                        .localized(context)
                  : 'Selesaikan 1 kuis atau ujian hari ini untuk memulai streak.'
                        .localized(context),
              style: theme.textTheme.bodySmall,
            ),
          ],
        ),
      ),
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
              'Beranda belum dapat dimuat'.localized(context),
              style: Theme.of(
                context,
              ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 7),
            Text(message.localized(context), textAlign: TextAlign.center),
            const SizedBox(height: 18),
            FilledButton.icon(
              onPressed: retry,
              icon: const Icon(Icons.refresh_rounded),
              label: Text('Coba lagi'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}
