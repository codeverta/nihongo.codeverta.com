import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/offline/download_preferences.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';

final offlineLeaseExpiryProvider = FutureProvider.autoDispose<DateTime?>((ref) {
  return ref.watch(offlineAccessStoreProvider).offlineLeaseExpiresAt;
});

class OfflineSettingsScreen extends ConsumerStatefulWidget {
  const OfflineSettingsScreen({super.key});

  @override
  ConsumerState<OfflineSettingsScreen> createState() => _OfflineSettingsState();
}

class _OfflineSettingsState extends ConsumerState<OfflineSettingsScreen> {
  bool _busy = false;

  Future<void> _clear({required bool imagesOnly}) async {
    if (_busy) return;
    if (!imagesOnly) {
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Hapus materi tersimpan?'),
          content: const Text(
            'Materi perlu diunduh kembali untuk digunakan offline. Progres belajar dan penilaian yang menunggu sinkronisasi tetap disimpan.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Batal'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Hapus materi'),
            ),
          ],
        ),
      );
      if (confirmed != true || !mounted) return;
    }
    setState(() => _busy = true);
    try {
      final downloads = ref.read(learningDownloadsProvider).valueOrNull;
      downloads?.cancel();
      final cache = await ref.read(learningCacheProvider.future);
      if (imagesOnly) {
        await cache.clearAutomaticMedia();
      } else {
        await cache.clearAll();
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              imagesOnly
                  ? 'Gambar otomatis dibersihkan. Teks materi tetap tersedia.'
                  : 'Materi tersimpan sudah dihapus.',
            ),
          ),
        );
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Penyimpanan belum dapat dibersihkan. Coba lagi.'),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final stats = ref.watch(learningCacheStatsProvider);
    final allowMobile = ref.watch(allowLargeMobileDownloadsProvider);
    final expiry = ref.watch(offlineLeaseExpiryProvider).valueOrNull?.toLocal();
    String mb(int bytes) => '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    return Scaffold(
      appBar: AppBar(title: const Text('Belajar offline & penyimpanan')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          const Text(
            'Teks materi disiapkan saat membuka course dan langsung dipakai saat belajar berikutnya. Batas cache otomatis 100 MB; gambar otomatis dibersihkan lebih dahulu ketika ruang penuh.',
          ),
          const SizedBox(height: 16),
          stats.when(
            data: (value) => Card(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${mb(value.totalBytes)} / 100 MB',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: 12),
                    LinearProgressIndicator(
                      value: value.usageRatio.clamp(0, 1),
                    ),
                    const SizedBox(height: 16),
                    Text('Teks materi: ${mb(value.textBytes)}'),
                    Text('Gambar: ${mb(value.mediaBytes)}'),
                    if (value.pinnedBytes > 0)
                      Text('Unduhan pilihanmu: ${mb(value.pinnedBytes)}'),
                  ],
                ),
              ),
            ),
            error: (_, __) => const ListTile(
              leading: Icon(Icons.storage_outlined),
              title: Text('Ukuran penyimpanan belum dapat dibaca.'),
            ),
            loading: () => const LinearProgressIndicator(),
          ),
          const SizedBox(height: 12),
          SwitchListTile.adaptive(
            contentPadding: EdgeInsets.zero,
            title: const Text('Unduhan besar lewat data seluler'),
            subtitle: const Text(
              'Teks dan ilustrasi kecil tetap boleh disiapkan. Gambar besar otomatis menunggu Wi-Fi jika pilihan ini mati.',
            ),
            value: allowMobile,
            onChanged: (value) => ref
                .read(allowLargeMobileDownloadsProvider.notifier)
                .setEnabled(enabled: value),
          ),
          if (expiry != null)
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.workspace_premium_outlined),
              title: const Text('Premium offline aktif'),
              subtitle: Text(
                'Verifikasi berlaku sampai ${expiry.day}/${expiry.month}/${expiry.year}. Diperbarui otomatis saat terhubung.',
              ),
            ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _busy ? null : () => _clear(imagesOnly: true),
            icon: const Icon(Icons.image_not_supported_outlined),
            label: const Text('Bersihkan gambar otomatis'),
          ),
          const SizedBox(height: 8),
          TextButton(
            onPressed: _busy ? null : () => _clear(imagesOnly: false),
            child: const Text('Hapus semua materi tersimpan'),
          ),
          const SizedBox(height: 16),
          const Text(
            'Materi yang belum selesai disiapkan membutuhkan internet saat pertama dibuka. Simulasi ujian tetap memerlukan koneksi. Logout menghapus data pribadi lokal, termasuk progres yang belum tersinkron.',
          ),
        ],
      ),
    );
  }
}
