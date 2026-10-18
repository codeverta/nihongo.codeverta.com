import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:url_launcher/url_launcher.dart';

class PremiumScreen extends ConsumerWidget {
  const PremiumScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authControllerProvider);
    final billing = ref.watch(billingControllerProvider);
    final publicPrice = ref.watch(lifetimeDisplayPriceProvider).valueOrNull;
    final displayPriceIdr = billing.priceIdr ?? publicPrice;
    final controller = ref.read(billingControllerProvider.notifier);
    return Scaffold(
      appBar: AppBar(title: const Text('Lifetime Access')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Icon(Icons.all_inclusive_rounded, size: 64),
          const SizedBox(height: 20),
          Text(
            billing.premium
                ? 'Lifetime Access aktif'
                : 'Sekali bayar, akses selamanya',
            style: Theme.of(context).textTheme.headlineSmall,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          Text(
            displayPriceIdr == null
                ? billing.product?.price ?? 'Harga belum tersedia'
                : _rupiah(displayPriceIdr),
            style: Theme.of(context).textTheme.headlineMedium,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          const Text(
            'Tanpa biaya bulanan atau perpanjangan otomatis. Harga akhir mengikuti store dan wilayah akun Anda.',
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          const ListTile(
            leading: Icon(Icons.menu_book),
            title: Text('Materi N3, N2, dan N1 yang tersedia'),
          ),
          const ListTile(
            leading: Icon(Icons.quiz),
            title: Text('Seluruh tes yang tersedia'),
          ),
          const ListTile(
            leading: Icon(Icons.devices),
            title: Text(
              'Akses di Android dan iOS dengan akun aplikasi yang sama',
            ),
          ),
          const Text(
            'Materi N5–N4 dan tes N5 tetap gratis. Kuis di dalam lesson mengikuti akses materi.',
          ),
          const SizedBox(height: 20),
          if (billing.message != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(billing.message!, key: const Key('billing-message')),
            ),
          if (!auth.isAuthenticated)
            FilledButton(
              onPressed: () => context.push('/login'),
              child: const Text('Masuk untuk melanjutkan'),
            )
          else ...[
            if (billing.premium)
              FilledButton(
                onPressed: () {
                  final from = GoRouterState.of(
                    context,
                  ).uri.queryParameters['from'];
                  context.go(
                    from != null &&
                            (from.startsWith('/learn/') ||
                                from.startsWith('/exams/'))
                        ? from
                        : '/home',
                  );
                },
                child: const Text('Lanjutkan belajar'),
              ),
            FilledButton.icon(
              key: const Key('buy-lifetime'),
              onPressed:
                  billing.ready &&
                      !billing.busy &&
                      !billing.pending &&
                      !billing.premium
                  ? controller.buy
                  : null,
              icon: billing.busy
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.lock_open),
              label: Text(
                billing.premium ? 'Akses aktif' : 'Beli Lifetime Access',
              ),
            ),
            const SizedBox(height: 10),
            OutlinedButton(
              onPressed: billing.busy ? null : controller.restore,
              child: const Text('Restore pembelian'),
            ),
            TextButton(
              onPressed: billing.busy ? null : controller.retry,
              child: const Text('Coba verifikasi lagi'),
            ),
          ],
          const Divider(height: 32),
          const Text(
            'Pembayaran yang dibatalkan atau tertunda tidak mengaktifkan akses. Refund atau pencabutan oleh store mengunci premium kembali. Menghapus akun tidak mengembalikan uang; pembelian dapat dipulihkan dengan akun store pembeli setelah masuk ke akun aplikasi baru.',
          ),
          const SizedBox(height: 8),
          TextButton(
            onPressed: () => _help(
              context,
              'https://support.google.com/googleplay/answer/2479637',
            ),
            child: const Text('Bantuan refund Google Play'),
          ),
          TextButton(
            onPressed: () => _help(context, 'https://support.apple.com/118223'),
            child: const Text('Bantuan refund App Store'),
          ),
        ],
      ),
    );
  }

  Future<void> _help(BuildContext context, String url) async {
    final opened = await launchUrl(
      Uri.parse(url),
      mode: LaunchMode.externalApplication,
    );
    if (!opened && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tautan bantuan tidak dapat dibuka.')),
      );
    }
  }
}

String _rupiah(int value) {
  final digits = value.toString();
  final result = StringBuffer();
  for (var index = 0; index < digits.length; index++) {
    if (index > 0 && (digits.length - index) % 3 == 0) result.write('.');
    result.write(digits[index]);
  }
  return 'Rp$result';
}
