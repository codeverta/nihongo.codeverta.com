import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/gamification_widgets.dart';

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
        appBar: AppBar(title: const Text('Achievements')),
        body: Center(
          child: FilledButton(
            onPressed: () => context.push('/login'),
            child: const Text('Masuk untuk melihat achievements'),
          ),
        ),
      );
    }
    return Scaffold(
      appBar: AppBar(title: const Text('Achievements')),
      body: ref.watch(gamificationProvider).when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (_, __) => GamificationError(
              retry: () => ref.invalidate(gamificationProvider),
            ),
            data: (data) => RefreshIndicator(
              onRefresh: () async {
                await ref.refresh(gamificationProvider.future).then<void>(
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
                    '${data.unlockedCount}/${data.achievements.length} achievement terbuka',
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    children: [
                      for (final (index, label)
                          in ['Semua', 'Terbuka', 'Terkunci'].indexed)
                        ChoiceChip(
                          label: Text(label),
                          selected: filter == index,
                          onSelected: (_) => setState(() => filter = index),
                        ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  for (final a in data.achievements.where(
                    (a) =>
                        filter == 0 ||
                        (filter == 1 ? a.isUnlocked : !a.isUnlocked),
                  ))
                    AchievementTile(achievement: a),
                  if (data.achievements
                      .where(
                        (a) =>
                            filter == 0 ||
                            (filter == 1 ? a.isUnlocked : !a.isUnlocked),
                      )
                      .isEmpty)
                    const Padding(
                      padding: EdgeInsets.all(20),
                      child: Text('Belum ada achievement dalam kategori ini.'),
                    ),
                ],
              ),
            ),
          ),
    );
  }
}
