import 'package:flutter/material.dart';
import 'package:nihongo_master/core/widgets/skeleton.dart';
import 'package:nihongo_master/features/learn/activity/continue_learning_section.dart';
import 'package:nihongo_master/features/news/presentation/news_screens.dart';

/// Placeholder shaped like the real Beranda (greeting, banner, progress,
/// quick menu, news) so a slow API feels like content arriving instead of
/// an empty screen with a spinner. "Lanjutkan belajar" comes from local
/// history, so it is shown for real even while the dashboard loads.
class HomeDashboardSkeleton extends StatelessWidget {
  const HomeDashboardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(
      key: const Key('home-loading'),
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 18, 20, 32),
      children: const [
        SkeletonShimmer(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _HeaderSkeleton(),
              SizedBox(height: 16),
              SkeletonBox(height: 150, radius: 20),
              SizedBox(height: 12),
              SkeletonBox(height: 72, radius: 18),
              SizedBox(height: 20),
              _QuickMenuSkeleton(),
            ],
          ),
        ),
        SizedBox(height: 22),
        ContinueLearningSection(),
        SizedBox(height: 24),
        SkeletonShimmer(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SkeletonBox(width: 120, height: 18),
              SizedBox(height: 8),
              SkeletonBox(width: 200, height: 12),
              SizedBox(height: 14),
              NewsPreviewSkeleton(),
            ],
          ),
        ),
      ],
    );
  }
}

class _HeaderSkeleton extends StatelessWidget {
  const _HeaderSkeleton();

  @override
  Widget build(BuildContext context) {
    return const Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SkeletonBox(width: 60, height: 11),
              SizedBox(height: 8),
              SkeletonBox(width: 170, height: 22),
              SizedBox(height: 8),
              SkeletonBox(width: 120, height: 11),
            ],
          ),
        ),
        SkeletonBox.circle(size: 40),
        SizedBox(width: 8),
        SkeletonBox.circle(size: 44),
      ],
    );
  }
}

class _QuickMenuSkeleton extends StatelessWidget {
  const _QuickMenuSkeleton();

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        // Mirrors HomeQuickMenu: four columns on phones, eight on tablets.
        final columns = constraints.maxWidth >= 600 ? 8 : 4;
        final spacing = constraints.maxWidth < 330 ? 2.0 : 6.0;
        final width =
            (constraints.maxWidth - spacing * (columns - 1)) / columns;
        return Wrap(
          spacing: spacing,
          runSpacing: 14,
          children: [
            for (var i = 0; i < 8; i++)
              SizedBox(
                width: width,
                child: const Column(
                  children: [
                    SkeletonBox(width: 56, height: 56, radius: 18),
                    SizedBox(height: 8),
                    SkeletonBox(width: 48, height: 10),
                  ],
                ),
              ),
          ],
        );
      },
    );
  }
}
