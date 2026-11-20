import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/leaderboard/data/leaderboard_repository.dart';
import 'package:nihongo_master/features/leaderboard/presentation/leaderboard_copy.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

const leaderboardPath = '/leaderboard';

/// Rank line inside the Beranda progress card. Only calls `/me/leaderboard`.
class HomeLeaderboardRankRow extends ConsumerWidget {
  const HomeLeaderboardRankRow({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final copy = LeaderboardCopy.of(context);
    final me = ref.watch(myLeaderboardProvider).valueOrNull?.data;
    final String text;
    if (me == null) {
      text = 'Lihat leaderboard minggu ini'.localized(context);
    } else if (!me.week.ranked) {
      text =
          'Belum masuk peringkat minggu ini — selesaikan 1 lesson untuk ikut bersaing'
              .localized(context);
    } else if (me.hidden) {
      text = copy.hiddenRank(me.week.rank);
    } else {
      text = copy.homeRank(me.week.rank, me.week.gapToNext);
    }
    return _LeaderboardLine(
      key: const Key('home-leaderboard-row'),
      text: text,
      onTap: () => context.push(leaderboardPath),
    );
  }
}

/// Slim row shown to guests where the progress card would be.
class HomeLeaderboardGuestRow extends StatelessWidget {
  const HomeLeaderboardGuestRow({super.key});

  @override
  Widget build(BuildContext context) {
    return SakuraCard(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: _LeaderboardLine(
        key: const Key('home-leaderboard-guest-row'),
        text: 'Lihat leaderboard minggu ini'.localized(context),
        onTap: () => context.push(leaderboardPath),
      ),
    );
  }
}

class _LeaderboardLine extends StatelessWidget {
  const _LeaderboardLine({required this.text, required this.onTap, super.key});

  final String text;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Semantics(
      button: true,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: ConstrainedBox(
          constraints: const BoxConstraints(minHeight: 44),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 6),
            child: Row(
              children: [
                const Text('🏆', style: TextStyle(fontSize: 18)),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    text,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Icon(
                  Icons.chevron_right_rounded,
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// "Minggu lalu kamu finis #N" card, shown once per closed week until closed.
class HomeLastWeekCard extends ConsumerWidget {
  const HomeLastWeekCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final last = ref.watch(myLeaderboardProvider).valueOrNull?.data.lastWeek;
    final dismissed = ref.watch(dismissedLastWeekProvider);
    if (last == null || last.rank <= 0 || dismissed == last.weekStart) {
      return const SizedBox.shrink();
    }
    final theme = Theme.of(context);
    final sakura = SakuraColors.of(context);
    return Padding(
      padding: const EdgeInsets.only(top: 12),
      child: SakuraCard(
        key: const Key('home-last-week-card'),
        color: sakura.goldSoft,
        padding: const EdgeInsets.fromLTRB(14, 8, 4, 8),
        child: Row(
          children: [
            const Text('🏆', style: TextStyle(fontSize: 22)),
            const SizedBox(width: 10),
            Expanded(
              child: InkWell(
                onTap: () => context.push(leaderboardPath),
                child: Text(
                  LeaderboardCopy.of(context).lastWeek(last.rank, last.xp),
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ),
            IconButton(
              key: const Key('home-last-week-dismiss'),
              tooltip: 'Tutup'.localized(context),
              icon: const Icon(Icons.close_rounded),
              onPressed: () => ref
                  .read(dismissedLastWeekProvider.notifier)
                  .dismiss(last.weekStart),
            ),
          ],
        ),
      ),
    );
  }
}
