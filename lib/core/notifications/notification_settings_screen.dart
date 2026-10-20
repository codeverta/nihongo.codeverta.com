import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';

class NotificationSettingsScreen extends ConsumerWidget {
  const NotificationSettingsScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(pushNotificationsProvider);
    final controller = ref.read(pushNotificationsProvider.notifier);
    return Scaffold(
      appBar: AppBar(title: const Text('Notifikasi')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const Icon(Icons.notifications_outlined, size: 52),
          const SizedBox(height: 16),
          Text(
            'Pengumuman Nihongo Master',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          const Text(
            'Terima berita aplikasi dan informasi belajar. Anda dapat mengubah pilihan ini kapan saja. Pengumuman tidak memuat data pribadi akun Anda.',
          ),
          const SizedBox(height: 16),
          SwitchListTile.adaptive(
            contentPadding: EdgeInsets.zero,
            title: const Text('Terima pengumuman'),
            subtitle: Text(
              state.enabled && state.granted && !state.pendingSync
                  ? 'Aktif pada perangkat ini'
                  : state.enabled
                  ? 'Dipilih aktif; periksa izin dan koneksi perangkat'
                  : 'Nonaktif',
            ),
            value: state.enabled,
            onChanged: state.available && !state.busy
                ? (enabled) => controller.setEnabled(enabled: enabled)
                : null,
          ),
          if (state.busy) const LinearProgressIndicator(),
          if (state.message != null)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16),
              child: Text(state.message!),
            ),
          if (state.available) ...[
            TextButton.icon(
              onPressed: controller.openSystemSettings,
              icon: const Icon(Icons.settings_outlined),
              label: const Text('Pengaturan notifikasi Android'),
            ),
            if (state.pendingSync)
              OutlinedButton(
                onPressed: state.busy ? null : controller.refresh,
                child: const Text('Coba sambungkan lagi'),
              ),
          ],
        ],
      ),
    );
  }
}
