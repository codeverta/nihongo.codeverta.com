import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/gamification_widgets.dart';

/// With dozens of achievements the flat server order is hard to read: earned
/// ones come first (newest on top), then locked ones from closest to done.
/// Ties keep the server's order.
List<Achievement> orderAchievements(List<Achievement> items) {
  double ratio(Achievement a) =>
      a.target > 0 ? (a.progress / a.target).clamp(0.0, 1.0) : 0;
  final indexed = items.indexed.toList()
    ..sort((x, y) {
      final a = x.$2;
      final b = y.$2;
      if (a.isUnlocked != b.isUnlocked) return a.isUnlocked ? -1 : 1;
      final byRank = a.isUnlocked
          ? b.unlockedAt!.compareTo(a.unlockedAt!)
          : ratio(b).compareTo(ratio(a));
      return byRank != 0 ? byRank : x.$1.compareTo(y.$1);
    });
  return [for (final (_, a) in indexed) a];
}

class AchievementsScreen extends ConsumerStatefulWidget {
  const AchievementsScreen({super.key});
  @override
  ConsumerState<AchievementsScreen> createState() => _AchievementsState();
}

class _AchievementsState extends ConsumerState<AchievementsScreen> {
  int filter = 0;
  @override
  Widget build(BuildContext context) {
    if (!ref.watch(authControllerProvider).isAuthenticated) {
      return Scaffold(
        appBar: AppBar(title: Text('Achievements'.localized(context))),
        body: Center(
          child: FilledButton(
            onPressed: () => context.push('/login'),
            child: Text('Masuk untuk melihat achievements'.localized(context)),
          ),
        ),
      );
    }
    return Scaffold(
      appBar: AppBar(title: Text('Achievements'.localized(context))),
      body: ref
          .watch(gamificationProvider)
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (_, __) => GamificationError(
              retry: () => ref.invalidate(gamificationProvider),
            ),
            data: (data) => _list(context, data),
          ),
    );
  }

  Widget _list(BuildContext context, GamificationSummary data) {
    final achievements = orderAchievements(data.achievements);
    return RefreshIndicator(
      onRefresh: () async {
        await ref
            .refresh(gamificationProvider.future)
            .then<void>(
              (_) {},
              // The provider renders the error and retry action.
              onError: (Object error, StackTrace stack) {},
            );
      },
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(20),
        children: [
          GamificationStats(data: data),
          const SizedBox(height: 16),
          Text(
            '${data.unlockedCount}/${achievements.length} achievement terbuka'
                .localized(context),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            children: [
              for (final (index, label) in [
                'Semua',
                'Terbuka',
                'Terkunci',
              ].indexed)
                ChoiceChip(
                  label: Text(label.localized(context)),
                  selected: filter == index,
                  onSelected: (_) => setState(() => filter = index),
                ),
            ],
          ),
          const SizedBox(height: 12),
          for (final a in achievements.where(
            (a) => filter == 0 || (filter == 1 ? a.isUnlocked : !a.isUnlocked),
          ))
            AchievementTile(achievement: a),
          if (achievements
              .where(
                (a) =>
                    filter == 0 || (filter == 1 ? a.isUnlocked : !a.isUnlocked),
              )
              .isEmpty)
            Padding(
              padding: EdgeInsets.all(20),
              child: Text(
                'Belum ada achievement dalam kategori ini.'.localized(context),
              ),
            ),
        ],
      ),
    );
  }
}
