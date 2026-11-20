import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/leaderboard/data/leaderboard_repository.dart';
import 'package:nihongo_master/features/leaderboard/presentation/leaderboard_copy.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/profile_avatar.dart';

/// Global XP leaderboard: "Minggu ini" (default) and "Sepanjang masa".
class LeaderboardScreen extends StatelessWidget {
  const LeaderboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Leaderboard'),
          bottom: TabBar(
            tabs: [
              Tab(text: 'Minggu ini'.localized(context)),
              Tab(text: 'Sepanjang masa'.localized(context)),
            ],
          ),
        ),
        body: const SafeArea(
          top: false,
          child: TabBarView(
            children: [
              _LeaderboardTab(period: LeaderboardPeriod.week),
              _LeaderboardTab(period: LeaderboardPeriod.all),
            ],
          ),
        ),
      ),
    );
  }
}

class _LeaderboardTab extends ConsumerWidget {
  const _LeaderboardTab({required this.period});

  final LeaderboardPeriod period;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final page = ref.watch(leaderboardPageProvider(period));
    final mine = ref.watch(myLeaderboardProvider);
    final signedIn = ref.watch(
      authControllerProvider.select((s) => s.user != null),
    );
    Future<void> refresh() async {
      ref
        ..invalidate(leaderboardPageProvider(period))
        ..invalidate(myLeaderboardProvider);
      await ref.read(leaderboardPageProvider(period).future);
    }

    return page.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, _) => _Message(
        key: const Key('leaderboard-error'),
        icon: Icons.cloud_off_rounded,
        text: error is LeaderboardOffline
            ? 'Butuh koneksi internet'.localized(context)
            : 'Leaderboard belum dapat dimuat'.localized(context),
        onRetry: refresh,
      ),
      data: (cached) {
        final me = mine.valueOrNull?.data;
        return Column(
          children: [
            Expanded(
              child: RefreshIndicator(
                onRefresh: refresh,
                child: _Board(
                  period: period,
                  cached: cached,
                  me: me,
                  signedIn: signedIn,
                ),
              ),
            ),
            if (!signedIn)
              const _GuestPrompt()
            else if (me != null)
              _PinnedOwnRow(period: period, me: me, page: cached.data),
          ],
        );
      },
    );
  }
}

class _Board extends StatelessWidget {
  const _Board({
    required this.period,
    required this.cached,
    required this.me,
    required this.signedIn,
  });

  final LeaderboardPeriod period;
  final Cached<LeaderboardPage> cached;
  final MyLeaderboard? me;
  final bool signedIn;

  @override
  Widget build(BuildContext context) {
    final copy = LeaderboardCopy.of(context);
    final entries = cached.data.entries;
    final myRank = me == null || me!.hidden ? 0 : me!.standing(period).rank;
    final podium = entries.take(3).toList();
    final rest = entries.skip(3).toList();
    return ListView(
      key: PageStorageKey('leaderboard-${period.apiValue}'),
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
      children: [
        _StatusLine(period: period, cached: cached),
        const SizedBox(height: 12),
        if (signedIn && me != null && me!.hidden)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: Text(
              'Kamu disembunyikan dari leaderboard. Ubah di Pengaturan.'
                  .localized(context),
              key: const Key('leaderboard-hidden-note'),
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        if (entries.isEmpty)
          _Message(
            key: const Key('leaderboard-empty'),
            icon: Icons.emoji_events_outlined,
            text: period == LeaderboardPeriod.week
                ? 'Belum ada yang mendapat XP minggu ini. Jadilah yang pertama!'
                      .localized(context)
                : 'Belum ada peringkat.'.localized(context),
          )
        else ...[
          _Podium(entries: podium, myRank: myRank),
          const SizedBox(height: 12),
          for (final entry in rest)
            _EntryRow(
              key: ValueKey('leaderboard-row-${entry.rank}'),
              rank: entry.rank,
              name: copy.name(entry.displayName, entry.tag),
              photoUrl: entry.photoUrl,
              xp: entry.xp,
              highlighted: entry.rank == myRank,
            ),
          if (cached.data.participants > entries.length)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(
                copy.english
                    ? 'Showing the top ${entries.length} of ${cached.data.participants} learners'
                    : 'Menampilkan ${entries.length} teratas dari ${cached.data.participants} pelajar',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
        ],
      ],
    );
  }
}

/// Countdown when live; "Terakhir diperbarui HH.MM" when showing the cache.
class _StatusLine extends StatefulWidget {
  const _StatusLine({required this.period, required this.cached});

  final LeaderboardPeriod period;
  final Cached<LeaderboardPage> cached;

  @override
  State<_StatusLine> createState() => _StatusLineState();
}

class _StatusLineState extends State<_StatusLine> {
  Timer? _ticker;

  @override
  void initState() {
    super.initState();
    _ticker = Timer.periodic(const Duration(minutes: 1), (_) {
      if (mounted) setState(() {});
    });
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final copy = LeaderboardCopy.of(context);
    final theme = Theme.of(context);
    final String text;
    final IconData icon;
    if (widget.cached.fromCache) {
      text = copy.lastUpdated(widget.cached.fetchedAt);
      icon = Icons.cloud_off_rounded;
    } else if (widget.period == LeaderboardPeriod.week) {
      text = copy.endsIn(
        widget.cached.data.weekEndsAt.difference(DateTime.now().toUtc()),
      );
      icon = Icons.timer_outlined;
    } else {
      text = 'Total XP sejak bergabung'.localized(context);
      icon = Icons.all_inclusive_rounded;
    }
    return Row(
      key: const Key('leaderboard-status'),
      children: [
        Icon(icon, size: 18, color: theme.colorScheme.onSurfaceVariant),
        const SizedBox(width: 6),
        Expanded(child: Text(text, style: theme.textTheme.bodyMedium)),
      ],
    );
  }
}

class _Podium extends StatelessWidget {
  const _Podium({required this.entries, required this.myRank});

  final List<LeaderboardEntry> entries;
  final int myRank;

  @override
  Widget build(BuildContext context) {
    // Classic order: 2nd, 1st, 3rd.
    final order = [
      if (entries.length > 1) entries[1],
      entries[0],
      if (entries.length > 2) entries[2],
    ];
    return Row(
      key: const Key('leaderboard-podium'),
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        for (final entry in order)
          Expanded(
            child: _PodiumSpot(entry: entry, mine: entry.rank == myRank),
          ),
      ],
    );
  }
}

class _PodiumSpot extends StatelessWidget {
  const _PodiumSpot({required this.entry, required this.mine});

  final LeaderboardEntry entry;
  final bool mine;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final sakura = SakuraColors.of(context);
    final copy = LeaderboardCopy.of(context);
    final medal = const {1: '🥇', 2: '🥈', 3: '🥉'}[entry.rank] ?? '';
    final height = const {1: 36.0, 2: 22.0, 3: 12.0}[entry.rank] ?? 0;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(medal, style: const TextStyle(fontSize: 22)),
          const SizedBox(height: 4),
          ProfileAvatar(
            radius: entry.rank == 1 ? 30 : 24,
            displayName: copy.name(entry.displayName, entry.tag),
            photoUrl: entry.photoUrl,
          ),
          const SizedBox(height: 6),
          Text(
            copy.name(entry.displayName, entry.tag),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            textAlign: TextAlign.center,
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w700,
              color: mine ? theme.colorScheme.primary : null,
            ),
          ),
          Text(
            copy.xp(entry.xp),
            style: theme.textTheme.labelMedium?.copyWith(
              color: sakura.goldText,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 6),
          Container(
            height: height + 8,
            decoration: BoxDecoration(
              color: mine
                  ? theme.colorScheme.primaryContainer
                  : sakura.goldSoft,
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(10),
              ),
            ),
            alignment: Alignment.center,
            child: Text(
              '#${entry.rank}',
              style: theme.textTheme.labelLarge?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _EntryRow extends StatelessWidget {
  const _EntryRow({
    required this.rank,
    required this.name,
    required this.photoUrl,
    required this.xp,
    this.highlighted = false,
    this.trailingLabel,
    super.key,
  });

  final int rank;
  final String name;
  final String? photoUrl;
  final int xp;
  final bool highlighted;
  final String? trailingLabel;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 3),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: highlighted ? scheme.primaryContainer : scheme.surfaceContainer,
        borderRadius: BorderRadius.circular(14),
        border: highlighted ? Border.all(color: scheme.primary) : null,
      ),
      child: Row(
        children: [
          SizedBox(
            width: 40,
            child: FittedBox(
              fit: BoxFit.scaleDown,
              alignment: Alignment.centerLeft,
              child: Text(
                rank > 0 ? '#$rank' : '–',
                style: theme.textTheme.labelLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),
          const SizedBox(width: 6),
          ProfileAvatar(radius: 16, displayName: name, photoUrl: photoUrl),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: highlighted ? FontWeight.w800 : FontWeight.w600,
                  ),
                ),
                if (trailingLabel != null)
                  Text(
                    trailingLabel!,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: scheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 88),
            child: FittedBox(
              fit: BoxFit.scaleDown,
              alignment: Alignment.centerRight,
              child: Text(
                '$xp XP',
                style: theme.textTheme.labelLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                  color: SakuraColors.of(context).goldText,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// The viewer's own row, pinned under the list when it is not visible in the
/// top 50 (or when hidden, as "posisimu jika ditampilkan").
class _PinnedOwnRow extends StatelessWidget {
  const _PinnedOwnRow({
    required this.period,
    required this.me,
    required this.page,
  });

  final LeaderboardPeriod period;
  final MyLeaderboard me;
  final LeaderboardPage page;

  @override
  Widget build(BuildContext context) {
    final standing = me.standing(period);
    final copy = LeaderboardCopy.of(context);
    final visibleInList =
        !me.hidden && standing.ranked && standing.rank <= page.entries.length;
    if (visibleInList) return const SizedBox.shrink();
    final String label;
    if (!standing.ranked) {
      label = period == LeaderboardPeriod.week
          ? 'Belum masuk peringkat minggu ini — selesaikan 1 lesson untuk ikut bersaing'
                .localized(context)
          : 'Kumpulkan XP untuk masuk peringkat'.localized(context);
    } else if (me.hidden) {
      label = copy.hiddenRank(standing.rank);
    } else {
      label = copy.rankOf(standing.rank, standing.participants);
    }
    return Material(
      key: const Key('leaderboard-own-row'),
      elevation: 6,
      color: Theme.of(context).colorScheme.surface,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 6, 16, 10),
        child: _EntryRow(
          rank: me.hidden ? 0 : standing.rank,
          name: copy.name(me.displayName, me.tag),
          photoUrl: me.photoUrl,
          xp: standing.xp,
          highlighted: true,
          trailingLabel: label,
        ),
      ),
    );
  }
}

class _GuestPrompt extends StatelessWidget {
  const _GuestPrompt();

  @override
  Widget build(BuildContext context) {
    return Material(
      key: const Key('leaderboard-guest-prompt'),
      elevation: 6,
      color: Theme.of(context).colorScheme.surface,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
        child: Row(
          children: [
            const Text('🏆', style: TextStyle(fontSize: 22)),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                'Masuk untuk ikut peringkat'.localized(context),
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
              ),
            ),
            const SizedBox(width: 8),
            FilledButton(
              onPressed: () => context.push('/login'),
              child: Text('Masuk'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}

class _Message extends StatelessWidget {
  const _Message({
    required this.icon,
    required this.text,
    this.onRetry,
    super.key,
  });

  final IconData icon;
  final String text;
  final Future<void> Function()? onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: Theme.of(context).colorScheme.primary),
            const SizedBox(height: 12),
            Text(text, textAlign: TextAlign.center),
            if (onRetry != null) ...[
              const SizedBox(height: 14),
              FilledButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh_rounded),
                label: Text('Coba lagi'.localized(context)),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
