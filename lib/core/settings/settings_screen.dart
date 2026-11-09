import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Pusat pengaturan. Aksi akun yang berisiko (logout dan hapus akun) sengaja
/// ditaruh di sini, terpisah dari menu profil, agar tidak tertekan tanpa sengaja.
class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(authControllerProvider).user;
    final error = Theme.of(context).colorScheme.error;
    return Scaffold(
      appBar: AppBar(title: Text('Pengaturan'.localized(context))),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const _LanguageCard(),
          const SizedBox(height: 12),
          Card(
            child: Column(
              children: [
                _SettingsTile(
                  icon: Icons.text_fields_rounded,
                  title: 'Pengaturan tampilan'.localized(context),
                  onTap: () => context.push('/settings/display'),
                ),
                const Divider(height: 1, indent: 68),
                _SettingsTile(
                  icon: Icons.notifications_outlined,
                  title: 'Pengaturan notifikasi'.localized(context),
                  onTap: () => context.push('/notifications'),
                ),
                const Divider(height: 1, indent: 68),
                _SettingsTile(
                  icon: Icons.offline_bolt_outlined,
                  title: 'Belajar offline & penyimpanan'.localized(context),
                  onTap: () => context.push('/settings/offline'),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          const Card(child: _AnalyticsSwitch()),
          const _AdPrivacyTile(),
          if (user != null) ...[
            const SizedBox(height: 32),
            Padding(
              padding: const EdgeInsets.only(left: 4, bottom: 8),
              child: Text(
                'Akun'.localized(context),
                style: Theme.of(
                  context,
                ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800),
              ),
            ),
            Card(
              child: Column(
                children: [
                  _SettingsTile(
                    key: const Key('settings-logout'),
                    icon: Icons.logout_rounded,
                    iconColor: error,
                    title: 'Logout'.localized(context),
                    onTap: () => _confirmLogout(context, ref),
                  ),
                  const Divider(height: 1, indent: 68),
                  _SettingsTile(
                    key: const Key('settings-delete-account'),
                    icon: Icons.delete_forever_outlined,
                    iconColor: error,
                    title: 'Hapus akun'.localized(context),
                    onTap: () => _confirmDelete(context, ref),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Future<void> _confirmLogout(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Keluar dari akun?'.localized(context)),
        content: Text(
          'Data pribadi di perangkat akan dihapus. Progres dan penilaian offline yang belum tersinkron juga akan hilang. Hubungkan internet sebelum logout untuk menyinkronkannya.'
              .localized(context),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Batal'.localized(context)),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text('Logout'.localized(context)),
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
        title: Text('Hapus akun permanen?'.localized(context)),
        content: Text(
          'Profil, progress, achievement, dan sesi login akan dihapus. Pembayaran store tidak otomatis direfund. Bukti transaksi minimum disimpan untuk restore dan mencegah penyalahgunaan. Setelah membuat akun baru, gunakan Restore dengan akun store pembeli.'
              .localized(context),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Batal'.localized(context)),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            ),
            onPressed: () => Navigator.pop(context, true),
            child: Text('Hapus akun'.localized(context)),
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
          SnackBar(
            content: Text(
              'Akun belum dihapus. Periksa koneksi lalu coba lagi.'.localized(
                context,
              ),
            ),
          ),
        );
      }
    }
  }
}

class _SettingsTile extends StatelessWidget {
  const _SettingsTile({
    required this.icon,
    required this.title,
    required this.onTap,
    this.iconColor,
    super.key,
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

class _AnalyticsSwitch extends ConsumerStatefulWidget {
  const _AnalyticsSwitch();

  @override
  ConsumerState<_AnalyticsSwitch> createState() => _AnalyticsSwitchState();
}

class _AnalyticsSwitchState extends ConsumerState<_AnalyticsSwitch> {
  late bool _enabled = ref.read(analyticsServiceProvider).enabled;

  @override
  Widget build(BuildContext context) {
    return SwitchListTile.adaptive(
      key: const Key('settings-analytics'),
      secondary: Icon(
        Icons.insights_outlined,
        color: Theme.of(context).colorScheme.primary,
      ),
      title: Text(
        'Bantu tingkatkan aplikasi'.localized(context),
        style: const TextStyle(fontWeight: FontWeight.w600),
      ),
      subtitle: Text(
        'Kirim data penggunaan anonim (fitur yang dibuka). Tidak berisi data pribadi.'
            .localized(context),
      ),
      value: _enabled,
      onChanged: (value) {
        setState(() => _enabled = value);
        unawaited(
          ref.read(analyticsServiceProvider).setEnabled(value: value),
        );
      },
    );
  }
}

/// Lets people reopen the Google consent form (required in some regions).
/// Hidden unless ads are active and the consent form says it must be offered.
class _AdPrivacyTile extends ConsumerWidget {
  const _AdPrivacyTile();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = ref.watch(adsReadyProvider).valueOrNull != null;
    if (!active) return const SizedBox.shrink();
    final runtime = ref.read(adsRuntimeProvider);
    return FutureBuilder<bool>(
      future: runtime.privacyOptionsRequired(),
      builder: (context, snapshot) {
        if (snapshot.data != true) return const SizedBox.shrink();
        return Padding(
          padding: const EdgeInsets.only(top: 12),
          child: Card(
            child: _SettingsTile(
              key: const Key('settings-ad-privacy'),
              icon: Icons.privacy_tip_outlined,
              title: 'Pilihan privasi iklan'.localized(context),
              onTap: () => unawaited(runtime.showPrivacyOptions()),
            ),
          ),
        );
      },
    );
  }
}

/// Quick language switch. The same choice also moves push notifications to the
/// matching language and is stored only on this device.
class _LanguageCard extends ConsumerWidget {
  const _LanguageCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(appLanguageProvider);
    final englishAvailable = ref.watch(englishAvailabilityProvider);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  Icons.language_rounded,
                  color: Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Bahasa aplikasi'.localized(context),
                    style: const TextStyle(fontWeight: FontWeight.w700),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Padding(
              padding: const EdgeInsets.only(left: 36),
              child: Text(
                'Berlaku untuk tampilan aplikasi, materi, dan notifikasi.'
                    .localized(context),
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: SegmentedButton<String>(
                key: const Key('settings-language'),
                showSelectedIcon: false,
                segments: [
                  ButtonSegment(
                    value: 'id',
                    label: Text('Bahasa Indonesia'.localized(context)),
                  ),
                  if (englishAvailable)
                    const ButtonSegment(value: 'en', label: Text('English')),
                ],
                selected: {locale.languageCode},
                onSelectionChanged: (value) {
                  unawaited(
                    ref
                        .read(appLanguageProvider.notifier)
                        .selectLanguage(value.first),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
