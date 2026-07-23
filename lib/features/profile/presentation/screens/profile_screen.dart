import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final user = ref.watch(authControllerProvider).user;

    return CustomScrollView(
      key: const PageStorageKey('profile-scroll'),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
          sliver: SliverList.list(
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      'Profil',
                      style:
                          Theme.of(context).textTheme.headlineMedium?.copyWith(
                                fontWeight: FontWeight.w900,
                              ),
                    ),
                  ),
                  IconButton.filledTonal(
                    onPressed: () {},
                    icon: const Icon(Icons.settings_outlined),
                    tooltip: 'Pengaturan',
                  ),
                ],
              ),
              const SizedBox(height: 22),
              _ProfileHeader(user: user),
              const SizedBox(height: 16),
              const _StatsRow(),
              const SizedBox(height: 24),
              _LifetimeCard(colorScheme: colorScheme),
              const SizedBox(height: 28),
              Text(
                'Gamification',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w800,
                    ),
              ),
              const SizedBox(height: 14),
              const _AchievementsCard(),
              const SizedBox(height: 24),
              _ProfileMenu(
                onLogout: () => _confirmLogout(context, ref),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _confirmLogout(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Keluar dari akun?'),
        content: const Text(
          'Progress lokal tetap aman dan dapat disinkronkan saat masuk kembali.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Logout'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    await ref.read(authControllerProvider.notifier).logout();
    if (context.mounted) context.go('/home');
  }
}

class _ProfileHeader extends StatelessWidget {
  const _ProfileHeader({required this.user});

  final AuthenticatedUser? user;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Row(
      children: [
        Stack(
          children: [
            CircleAvatar(
              radius: 38,
              backgroundColor: colorScheme.primary.withValues(alpha: 0.10),
              child: Text(
                (user?.displayName.isNotEmpty ?? false)
                    ? user!.displayName.characters.first.toUpperCase()
                    : 'N',
                style: TextStyle(
                  color: colorScheme.primary,
                  fontSize: 30,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            Positioned(
              right: 0,
              bottom: 0,
              child: CircleAvatar(
                radius: 13,
                backgroundColor: colorScheme.primary,
                child: const Icon(
                  Icons.edit_rounded,
                  size: 14,
                  color: Colors.white,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                user?.displayName ?? 'Nihongo Learner',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w900,
                    ),
              ),
              const SizedBox(height: 4),
              Text(
                user?.email ?? 'Memuat profil...',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 7),
              const _PlanBadge(),
            ],
          ),
        ),
      ],
    );
  }
}

class _PlanBadge extends StatelessWidget {
  const _PlanBadge();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: const Color(0xFFE59E24).withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(99),
      ),
      child: const Text(
        'FREE MEMBER',
        style: TextStyle(
          color: Color(0xFFB87500),
          fontWeight: FontWeight.w800,
          fontSize: 10,
          letterSpacing: 0.8,
        ),
      ),
    );
  }
}

class _StatsRow extends StatelessWidget {
  const _StatsRow();

  @override
  Widget build(BuildContext context) {
    return const Card(
      child: Padding(
        padding: EdgeInsets.symmetric(vertical: 18),
        child: Row(
          children: [
            Expanded(child: _Stat(value: '7', label: 'Streak', icon: '🔥')),
            SizedBox(height: 40, child: VerticalDivider()),
            Expanded(child: _Stat(value: '2.480', label: 'XP', icon: '⚡')),
            SizedBox(height: 40, child: VerticalDivider()),
            Expanded(child: _Stat(value: '12', label: 'Badge', icon: '🏆')),
          ],
        ),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.value, required this.label, required this.icon});

  final String value;
  final String label;
  final String icon;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(icon, style: const TextStyle(fontSize: 20)),
        const SizedBox(height: 5),
        Text(
          value,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w900,
              ),
        ),
        Text(label, style: Theme.of(context).textTheme.bodySmall),
      ],
    );
  }
}

class _LifetimeCard extends StatelessWidget {
  const _LifetimeCard({required this.colorScheme});

  final ColorScheme colorScheme;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [colorScheme.secondary, const Color(0xFF3538A8)],
        ),
        borderRadius: BorderRadius.circular(22),
      ),
      child: Row(
        children: [
          Container(
            width: 54,
            height: 54,
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.13),
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Icon(
              Icons.all_inclusive_rounded,
              color: Color(0xFFFFD166),
              size: 30,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Lifetime Access',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 17,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  'Sekali bayar, belajar selamanya.',
                  style: TextStyle(color: Colors.white.withValues(alpha: 0.72)),
                ),
              ],
            ),
          ),
          TextButton(
            onPressed: () {},
            style: TextButton.styleFrom(foregroundColor: Colors.white),
            child: const Text('Lihat'),
          ),
        ],
      ),
    );
  }
}

class _AchievementsCard extends StatelessWidget {
  const _AchievementsCard();

  @override
  Widget build(BuildContext context) {
    const achievements = [
      ('🔥', '7-Day Streak'),
      ('あ', 'Kana Master'),
      ('📚', 'Bookworm'),
      ('🔒', 'Locked'),
    ];

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Row(
              children: achievements.map((item) {
                final locked = item.$2 == 'Locked';
                return Expanded(
                  child: Opacity(
                    opacity: locked ? 0.4 : 1,
                    child: Column(
                      children: [
                        Container(
                          width: 52,
                          height: 52,
                          alignment: Alignment.center,
                          decoration: BoxDecoration(
                            color: Theme.of(context)
                                .colorScheme
                                .primary
                                .withValues(alpha: 0.08),
                            shape: BoxShape.circle,
                          ),
                          child: Text(
                            item.$1,
                            style: const TextStyle(fontSize: 25),
                          ),
                        ),
                        const SizedBox(height: 7),
                        Text(
                          item.$2,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          textAlign: TextAlign.center,
                          style: Theme.of(context).textTheme.labelSmall,
                        ),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 14),
            const Divider(),
            TextButton.icon(
              onPressed: () {},
              icon: const Icon(Icons.emoji_events_outlined),
              label: const Text('Lihat semua achievement'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ProfileMenu extends StatelessWidget {
  const _ProfileMenu({required this.onLogout});

  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          _ProfileMenuTile(
            icon: Icons.bar_chart_rounded,
            title: 'Statistik pembelajaran',
            onTap: () {},
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.logout_rounded,
            title: 'Logout',
            iconColor: Theme.of(context).colorScheme.error,
            onTap: onLogout,
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.download_for_offline_outlined,
            title: 'Materi offline',
            trailing: '3 tersimpan',
            onTap: () {},
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.credit_card_rounded,
            title: 'Paket & pembayaran',
            onTap: () {},
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.help_outline_rounded,
            title: 'Bantuan',
            onTap: () {},
          ),
        ],
      ),
    );
  }
}

class _ProfileMenuTile extends StatelessWidget {
  const _ProfileMenuTile({
    required this.icon,
    required this.title,
    required this.onTap,
    this.trailing,
    this.iconColor,
  });

  final IconData icon;
  final String title;
  final String? trailing;
  final VoidCallback onTap;
  final Color? iconColor;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 3),
      leading: Icon(
        icon,
        color: iconColor ?? Theme.of(context).colorScheme.primary,
      ),
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (trailing != null)
            Text(trailing!, style: Theme.of(context).textTheme.bodySmall),
          const SizedBox(width: 4),
          const Icon(Icons.chevron_right_rounded),
        ],
      ),
    );
  }
}
