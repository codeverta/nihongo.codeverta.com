import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
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
          title: Text('Hapus materi tersimpan?'.localized(context)),
          content: Text(
            'Materi perlu diunduh kembali untuk digunakan offline. Progres belajar dan penilaian yang menunggu sinkronisasi tetap disimpan.'
                .localized(context),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text('Batal'.localized(context)),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: Text('Hapus materi'.localized(context)),
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
                  ? 'Gambar otomatis dibersihkan. Teks materi tetap tersedia.'.localized(context)
                  : 'Materi tersimpan sudah dihapus.'.localized(context),
            ),
          ),
        );
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Penyimpanan belum dapat dibersihkan. Coba lagi.'.localized(
                context,
              ),
            ),
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
      appBar: AppBar(
        title: Text('Belajar offline & penyimpanan'.localized(context)),
      ),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            'Teks materi disiapkan saat membuka course dan langsung dipakai saat belajar berikutnya. Batas cache otomatis 100 MB; gambar otomatis dibersihkan lebih dahulu ketika ruang penuh.'
                .localized(context),
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
                    Text(
                      '${'Teks materi'.localized(context)}: ${mb(value.textBytes)}',
                    ),
                    Text(
                      '${'Gambar'.localized(context)}: ${mb(value.mediaBytes)}',
                    ),
                    if (value.pinnedBytes > 0)
                      Text(
                        '${'Unduhan pilihanmu'.localized(context)}: ${mb(value.pinnedBytes)}',
                      ),
                  ],
                ),
              ),
            ),
            error: (_, __) => ListTile(
              leading: const Icon(Icons.storage_outlined),
              title: Text(
                'Ukuran penyimpanan belum dapat dibaca.'.localized(context),
              ),
            ),
            loading: () => const LinearProgressIndicator(),
          ),
          const SizedBox(height: 12),
          SwitchListTile.adaptive(
            contentPadding: EdgeInsets.zero,
            title: Text('Unduhan besar lewat data seluler'.localized(context)),
            subtitle: Text(
              'Teks dan ilustrasi kecil tetap boleh disiapkan. Gambar besar otomatis menunggu Wi-Fi jika pilihan ini mati.'
                  .localized(context),
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
              title: Text('Premium offline aktif'.localized(context)),
              subtitle: Text(
                '${'Verifikasi berlaku sampai'.localized(context)} ${expiry.day}/${expiry.month}/${expiry.year}. ${'Diperbarui otomatis saat terhubung.'.localized(context)}',
              ),
            ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: _busy ? null : () => _clear(imagesOnly: true),
            icon: const Icon(Icons.image_not_supported_outlined),
            label: Text('Bersihkan gambar otomatis'.localized(context)),
          ),
          const SizedBox(height: 8),
          TextButton(
            onPressed: _busy ? null : () => _clear(imagesOnly: false),
            child: Text('Hapus semua materi tersimpan'.localized(context)),
          ),
          const SizedBox(height: 16),
          Text(
            'Materi yang belum selesai disiapkan membutuhkan internet saat pertama dibuka. Simulasi ujian tetap memerlukan koneksi. Logout menghapus data pribadi lokal, termasuk progres yang belum tersinkron.'
                .localized(context),
          ),
        ],
      ),
    );
  }
}
