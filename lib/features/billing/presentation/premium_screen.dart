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
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
        children: [
          _LifetimeHero(active: billing.premium),
          const SizedBox(height: 22),
          Text(
            'Lebih jauh dari sekadar belajar',
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 6),
          Text(
            'Dari persiapan JLPT sampai langkah kariermu di Jepang.',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 16),
          const _PremiumBenefit(
            icon: Icons.menu_book_rounded,
            title: 'Materi JLPT lanjutan',
            description:
                'Buka materi N3, N2, dan N1 yang tersedia beserta kuis pembelajarannya.',
          ),
          const _PremiumBenefit(
            icon: Icons.quiz_rounded,
            title: 'Latihan sampai siap ujian',
            description:
                'Akses seluruh tes yang tersedia, termasuk simulasi JLPT, JFT, NAT, dan SSW.',
          ),
          const _PremiumBenefit(
            icon: Icons.auto_stories_rounded,
            title: 'Baca Manga Jepang',
            description:
                'Nikmati koleksi manga berbahasa Jepang dan latih membaca lewat cerita.',
          ),
          const _PremiumBenefit(
            icon: Icons.work_rounded,
            title: 'Lanjut ke peluang kerja',
            description:
                'Buka tautan lamaran lowongan aktif dan lanjutkan prosesnya di situs sumber.',
          ),
          const SizedBox(height: 12),
          Container(
            key: const Key('lifetime-price-panel'),
            padding: const EdgeInsets.all(22),
            decoration: BoxDecoration(
              color: const Color(0xFFFFF7E7),
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: const Color(0xFFE8C887)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    Icon(
                      Icons.workspace_premium_rounded,
                      color: Color(0xFF825A14),
                    ),
                    SizedBox(width: 8),
                    Flexible(
                      child: Text(
                        'LIFETIME ACCESS',
                        style: TextStyle(
                          color: Color(0xFF825A14),
                          fontWeight: FontWeight.w800,
                          letterSpacing: 1,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Text(
                  billing.product?.price ??
                      (displayPriceIdr == null
                          ? 'Harga belum tersedia'
                          : _rupiah(displayPriceIdr)),
                  style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                    color: const Color(0xFF25264F),
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 4),
                const Text(
                  'Sekali bayar. Tanpa biaya bulanan.',
                  style: TextStyle(
                    color: Color(0xFF5C4C2C),
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 16),
                const _PriceNote(
                  'Akses dengan akun aplikasi yang sama di Android dan iOS',
                ),
                const _PriceNote('Tidak ada perpanjangan otomatis'),
                const _PriceNote(
                  'Harga akhir mengikuti store dan wilayah akun Anda',
                ),
              ],
            ),
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
                            (RegExp(r'^/jobs/[^/]+$').hasMatch(from) ||
                                from.startsWith('/learn/') ||
                                from.startsWith('/exams/') ||
                                from == '/manga' ||
                                from.startsWith('/manga/'))
                        ? from
                        : '/home',
                  );
                },
                child: const Text('Jelajahi akses saya'),
              ),
            FilledButton.icon(
              key: const Key('buy-lifetime'),
              style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(54),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
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
                billing.premium
                    ? 'Akses aktif'
                    : billing.pending
                    ? 'Menunggu pembayaran'
                    : 'Buka Lifetime Access',
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
          const SizedBox(height: 20),
          Text(
            'Materi N5–N4 dan tes N5 tetap gratis. Kuis di dalam lesson mengikuti akses materi.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const Divider(height: 32),
          Text(
            'Tentang pembelianmu',
            style: Theme.of(
              context,
            ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 10),
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
    var opened = false;
    try {
      opened = await launchUrl(
        Uri.parse(url),
        mode: LaunchMode.externalApplication,
      );
    } on Object {
      opened = false;
    }
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

class _LifetimeHero extends StatelessWidget {
  const _LifetimeHero({required this.active});
  final bool active;
  @override
  Widget build(BuildContext context) => Container(
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      color: const Color(0xFF20214F),
      borderRadius: BorderRadius.circular(28),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        AspectRatio(
          aspectRatio: 1.9,
          child: Image.asset(
            'assets/images/explore/lifetime-access.png',
            key: const Key('lifetime-illustration'),
            fit: BoxFit.cover,
            cacheWidth: 1200,
            semanticLabel:
                'Perjalanan belajar bahasa dan meraih peluang di Jepang',
            errorBuilder: (_, __, ___) => const Center(
              child: Icon(
                Icons.workspace_premium_rounded,
                size: 72,
                color: Color(0xFFFFD784),
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(22),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFD784),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      active
                          ? Icons.verified_rounded
                          : Icons.workspace_premium_rounded,
                      size: 16,
                      color: const Color(0xFF5A3E14),
                    ),
                    const SizedBox(width: 6),
                    Flexible(
                      child: Text(
                        active
                            ? 'LIFETIME AKTIF'
                            : 'SEKALI BAYAR · AKSES SELAMANYA',
                        style: const TextStyle(
                          color: Color(0xFF5A3E14),
                          fontWeight: FontWeight.w800,
                          fontSize: 10,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              Text(
                active
                    ? 'Lifetime Access aktif'
                    : 'Buka lebih banyak.\nMelangkah lebih jauh.',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  height: 1.2,
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                'Belajar, baca, latihan, dan temukan peluang dalam satu akses.',
                style: TextStyle(color: Color(0xFFDADBF0), height: 1.6),
              ),
            ],
          ),
        ),
      ],
    ),
  );
}

class _PremiumBenefit extends StatelessWidget {
  const _PremiumBenefit({
    required this.icon,
    required this.title,
    required this.description,
  });
  final IconData icon;
  final String title;
  final String description;
  @override
  Widget build(BuildContext context) => Card(
    margin: const EdgeInsets.only(bottom: 10),
    child: Padding(
      padding: const EdgeInsets.all(16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(11),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.primaryContainer,
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(
              icon,
              size: 23,
              color: Theme.of(context).colorScheme.onPrimaryContainer,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 5),
                Text(
                  description,
                  style: Theme.of(
                    context,
                  ).textTheme.bodySmall?.copyWith(height: 1.6),
                ),
              ],
            ),
          ),
        ],
      ),
    ),
  );
}

class _PriceNote extends StatelessWidget {
  const _PriceNote(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Icon(
          Icons.check_circle_outline_rounded,
          size: 17,
          color: Color(0xFF825A14),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(
              color: Color(0xFF5C4C2C),
              fontSize: 12,
              height: 1.4,
            ),
          ),
        ),
      ],
    ),
  );
}
