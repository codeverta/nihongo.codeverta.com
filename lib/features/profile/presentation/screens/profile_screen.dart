import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_image_compress/flutter_image_compress.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nihongo_master/core/widgets/legal_links.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/profile/presentation/screens/attributions_screen.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/gamification_widgets.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/profile_avatar.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  final _imagePicker = ImagePicker();
  bool _uploadingPhoto = false;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final user = ref.watch(authControllerProvider).user;
    if (user == null) {
      return const Center(
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('Masuk untuk melihat profil.'),
              LegalLinks(),
              AttributionsLink(),
            ],
          ),
        ),
      );
    }
    final gamification = ref.watch(gamificationProvider);

    return RefreshIndicator(
      onRefresh: () async {
        await ref.read(authControllerProvider.notifier).refreshUser();
        await ref
            .refresh(gamificationProvider.future)
            .then<void>((_) {}, onError: (Object error, StackTrace stack) {});
      },
      child: CustomScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
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
                        style: Theme.of(context).textTheme.headlineMedium
                            ?.copyWith(fontWeight: FontWeight.w900),
                      ),
                    ),
                    IconButton.filledTonal(
                      onPressed: () => context.push('/notifications'),
                      icon: const Icon(Icons.settings_outlined),
                      tooltip: 'Pengaturan notifikasi',
                    ),
                  ],
                ),
                const SizedBox(height: 22),
                _ProfileHeader(
                  user: user,
                  uploadingPhoto: _uploadingPhoto,
                  onChangePhoto: _uploadingPhoto ? null : _chooseProfilePhoto,
                ),
                const SizedBox(height: 8),
                Text(
                  'Foto dikompres sebelum diunggah. Tautan foto dapat dibuka oleh siapa pun yang memilikinya.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(height: 16),
                gamification.when(
                  loading: () =>
                      const Center(child: CircularProgressIndicator()),
                  error: (_, __) => GamificationError(
                    retry: () => ref.invalidate(gamificationProvider),
                  ),
                  data: (data) => GamificationStats(data: data),
                ),
                if (!user.isLifetime && !user.isPremium) ...[
                  const SizedBox(height: 24),
                  _LifetimeCard(colorScheme: colorScheme),
                ],
                const SizedBox(height: 28),
                Text(
                  'Gamification',
                  style: Theme.of(
                    context,
                  ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 14),
                gamification.when(
                  loading: () => const SizedBox.shrink(),
                  error: (_, __) => const SizedBox.shrink(),
                  data: (data) => Column(
                    children: [
                      for (final a in data.achievements.take(3))
                        AchievementTile(achievement: a),
                      TextButton.icon(
                        onPressed: () => context.push('/achievements'),
                        icon: const Icon(Icons.emoji_events_outlined),
                        label: Text(
                          'Lihat semua achievement (${data.unlockedCount}/${data.achievements.length})',
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                _ProfileMenu(
                  onLogout: () => _confirmLogout(context, ref),
                  onDelete: () => _confirmDelete(context, ref),
                ),
                const SizedBox(height: 16),
                const LegalLinks(),
                const AttributionsLink(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _chooseProfilePhoto() async {
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      showDragHandle: true,
      builder: (context) => SafeArea(
        child: Wrap(
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: const Text('Pilih dari galeri'),
              onTap: () => Navigator.pop(context, ImageSource.gallery),
            ),
            ListTile(
              leading: const Icon(Icons.photo_camera_outlined),
              title: const Text('Ambil foto'),
              onTap: () => Navigator.pop(context, ImageSource.camera),
            ),
          ],
        ),
      ),
    );
    if (source == null || !mounted) return;

    setState(() => _uploadingPhoto = true);
    try {
      final photo = await _imagePicker.pickImage(
        source: source,
        requestFullMetadata: false,
      );
      if (photo == null || !mounted) return;
      final compressed = await FlutterImageCompress.compressWithFile(
        photo.path,
        minWidth: 1024,
        minHeight: 1024,
        quality: 82,
      );
      if (compressed == null || compressed.isEmpty) {
        throw const AuthException(
          'Foto tidak dapat diproses. Pilih foto lain.',
        );
      }
      if (compressed.length > 8 * 1024 * 1024) {
        throw const AuthException('Foto masih terlalu besar. Pilih foto lain.');
      }
      await ref
          .read(authControllerProvider.notifier)
          .uploadProfilePhoto(Uint8List.fromList(compressed));
      ref.invalidate(homeDashboardProvider);
      _showPhotoMessage('Foto profil berhasil diperbarui.');
    } on AuthException catch (error) {
      _showPhotoMessage(error.message);
    } on Object {
      _showPhotoMessage(
        'Foto gagal diproses atau diunggah. Periksa koneksi lalu coba lagi.',
      );
    } finally {
      if (mounted) setState(() => _uploadingPhoto = false);
    }
  }

  void _showPhotoMessage(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
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

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Hapus akun permanen?'),
        content: const Text(
          'Profil, progress, achievement, dan sesi login akan dihapus. Pembayaran store tidak otomatis direfund. Bukti transaksi minimum disimpan untuk restore dan mencegah penyalahgunaan. Setelah membuat akun baru, gunakan Restore dengan akun store pembeli.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Hapus akun'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      await ref.read(authControllerProvider.notifier).deleteAccount();
      if (context.mounted) context.go('/home');
    } on Object {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Akun belum dihapus. Periksa koneksi lalu coba lagi.',
            ),
          ),
        );
      }
    }
  }
}

class _ProfileHeader extends StatelessWidget {
  const _ProfileHeader({
    required this.user,
    required this.uploadingPhoto,
    required this.onChangePhoto,
  });

  final AuthenticatedUser? user;
  final bool uploadingPhoto;
  final VoidCallback? onChangePhoto;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Row(
      children: [
        Semantics(
          button: onChangePhoto != null,
          label: 'Ubah foto profil',
          child: InkWell(
            onTap: onChangePhoto,
            customBorder: const CircleBorder(),
            child: Stack(
              children: [
                ProfileAvatar(
                  radius: 38,
                  displayName: user?.displayName ?? '',
                  photoUrl: user?.photoUrl,
                ),
                Positioned(
                  right: 0,
                  bottom: 0,
                  child: IgnorePointer(
                    child: CircleAvatar(
                      radius: 13,
                      backgroundColor: colorScheme.primary,
                      child: uploadingPhoto
                          ? const SizedBox(
                              width: 15,
                              height: 15,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : const Icon(
                              Icons.edit_rounded,
                              size: 14,
                              color: Colors.white,
                            ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                user?.displayName ?? 'Nihongo Learner',
                style: Theme.of(
                  context,
                ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w900),
              ),
              const SizedBox(height: 4),
              Text(
                user?.email ?? 'Memuat profil...',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 7),
              _PlanBadge(
                premium:
                    (user?.isPremium ?? false) || (user?.isLifetime ?? false),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _PlanBadge extends StatelessWidget {
  const _PlanBadge({required this.premium});
  final bool premium;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: const Color(0xFFE59E24).withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(99),
      ),
      child: Text(
        premium ? 'LIFETIME MEMBER' : 'FREE MEMBER',
        style: const TextStyle(
          color: Color(0xFFB87500),
          fontWeight: FontWeight.w800,
          fontSize: 10,
          letterSpacing: 0.8,
        ),
      ),
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
            onPressed: () => context.push('/premium'),
            style: TextButton.styleFrom(foregroundColor: Colors.white),
            child: const Text('Lihat'),
          ),
        ],
      ),
    );
  }
}

class _ProfileMenu extends StatelessWidget {
  const _ProfileMenu({required this.onLogout, required this.onDelete});

  final VoidCallback onLogout;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          _ProfileMenuTile(
            icon: Icons.bar_chart_rounded,
            title: 'Statistik pembelajaran',
            onTap: () => context.push('/achievements'),
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
            icon: Icons.credit_card_rounded,
            title: 'Paket & pembayaran',
            onTap: () => context.push('/premium'),
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.delete_forever_outlined,
            title: 'Hapus akun',
            iconColor: Theme.of(context).colorScheme.error,
            onTap: onDelete,
          ),
          const Divider(height: 1, indent: 68),
          _ProfileMenuTile(
            icon: Icons.help_outline_rounded,
            title: 'Bantuan',
            onTap: () => openLegalPage(context, '/support'),
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
    this.iconColor,
  });

  final IconData icon;
  final String title;
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
      trailing: const Icon(Icons.chevron_right_rounded),
    );
  }
}
