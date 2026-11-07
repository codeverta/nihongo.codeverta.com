import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';
import 'package:url_launcher/url_launcher.dart';

class PremiumScreen extends ConsumerWidget {
  const PremiumScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authControllerProvider);
    final billing = ref.watch(billingControllerProvider);
    final english = AppLocalizations.of(context).locale.languageCode == 'en';
    final publicPrice = english
        ? null
        : ref.watch(lifetimeDisplayPriceProvider).valueOrNull;
    final displayPriceIdr = billing.priceIdr ?? publicPrice;
    final controller = ref.read(billingControllerProvider.notifier);
    final theme = Theme.of(context);
    final price =
        billing.product?.price ??
        (english
            ? (billing.busy
                  ? 'Memeriksa harga store…'.localized(context)
                  : 'Harga store belum tersedia'.localized(context))
            : displayPriceIdr == null
            ? 'Harga belum tersedia'.localized(context)
            : _rupiah(displayPriceIdr));

    return Scaffold(
      appBar: AppBar(title: Text('Lifetime Access'.localized(context))),
      body: Column(
        children: [
          Expanded(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(20, 4, 20, 20),
              children: [
                StaggeredEntrance(child: _LifetimeHero(active: billing.premium)),
                const SizedBox(height: 16),
                const _BenefitGrid(),
                const SizedBox(height: 14),
                const _PriceNote(
                  'Akses dengan akun aplikasi yang sama di Android dan iOS',
                ),
                const _PriceNote('Tidak ada perpanjangan otomatis'),
                const _PriceNote(
                  'Harga akhir mengikuti store dan wilayah akun Anda',
                ),
                const _PriceNote(
                  'Materi N5–N4 dan tes N5 tetap gratis. Kuis di dalam lesson mengikuti akses materi.',
                ),
                const SizedBox(height: 6),
                Theme(
                  data: theme.copyWith(dividerColor: Colors.transparent),
                  child: ExpansionTile(
                    tilePadding: EdgeInsets.zero,
                    childrenPadding: EdgeInsets.zero,
                    expandedCrossAxisAlignment: CrossAxisAlignment.start,
                    title: Text(
                      'Tentang pembelianmu'.localized(context),
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    children: [
                      Text(
                        'Pembayaran yang dibatalkan atau tertunda tidak mengaktifkan akses. Refund atau pencabutan oleh store mengunci premium kembali. Menghapus akun tidak mengembalikan uang; pembelian dapat dipulihkan dengan akun store pembeli setelah masuk ke akun aplikasi baru.'
                            .localized(context),
                        style: theme.textTheme.bodySmall,
                      ),
                      const SizedBox(height: 4),
                      Wrap(
                        children: [
                          TextButton(
                            onPressed: () => _help(
                              context,
                              'https://support.google.com/googleplay/answer/2479637',
                            ),
                            child: Text(
                              'Bantuan refund Google Play'.localized(context),
                            ),
                          ),
                          TextButton(
                            onPressed: () => _help(
                              context,
                              'https://support.apple.com/118223',
                            ),
                            child: Text(
                              'Bantuan refund App Store'.localized(context),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // Harga dan tombol beli selalu terlihat, tanpa perlu menggulir.
          _CheckoutBar(
            price: price,
            message: billing.message,
            authenticated: auth.isAuthenticated,
            children: [
              if (!auth.isAuthenticated)
                FilledButton(
                  onPressed: () => context.push('/login'),
                  style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(52),
                  ),
                  child: Text('Masuk untuk melanjutkan'.localized(context)),
                )
              else ...[
                if (billing.premium) ...[
                  FilledButton(
                    style: FilledButton.styleFrom(
                      minimumSize: const Size.fromHeight(52),
                    ),
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
                    child: Text('Jelajahi akses saya'.localized(context)),
                  ),
                  const SizedBox(height: 8),
                ],
                FilledButton.icon(
                  key: const Key('buy-lifetime'),
                  style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(52),
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
                        ? 'Akses aktif'.localized(context)
                        : billing.pending
                        ? 'Menunggu pembayaran'.localized(context)
                        : 'Buka Lifetime Access'.localized(context),
                  ),
                ),
                const SizedBox(height: 2),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Flexible(
                      child: TextButton(
                        onPressed: billing.busy ? null : controller.restore,
                        child: Text(
                          'Restore pembelian'.localized(context),
                          textAlign: TextAlign.center,
                        ),
                      ),
                    ),
                    Text('·', style: theme.textTheme.bodySmall),
                    Flexible(
                      child: TextButton(
                        onPressed: billing.busy ? null : controller.retry,
                        child: Text(
                          'Coba verifikasi lagi'.localized(context),
                          textAlign: TextAlign.center,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ],
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
        SnackBar(
          content: Text(
            'Tautan bantuan tidak dapat dibuka.'.localized(context),
          ),
        ),
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

/// Panel bawah yang menempel: harga + tombol beli.
class _CheckoutBar extends StatelessWidget {
  const _CheckoutBar({
    required this.price,
    required this.message,
    required this.authenticated,
    required this.children,
  });

  final String price;
  final String? message;
  final bool authenticated;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final sakura = SakuraColors.of(context);
    final dark = theme.brightness == Brightness.dark;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: dark ? AppTheme.darkSurface : Colors.white,
        border: Border(top: BorderSide(color: theme.colorScheme.outlineVariant)),
        boxShadow: [
          BoxShadow(
            color: sakura.shadow.withValues(alpha: dark ? 0.5 : 0.1),
            blurRadius: 24,
            offset: const Offset(0, -6),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          key: const Key('lifetime-price-panel'),
          padding: const EdgeInsets.fromLTRB(20, 14, 20, 6),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          price,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.headlineSmall?.copyWith(
                            fontWeight: FontWeight.w800,
                            letterSpacing: -0.5,
                          ),
                        ),
                        Text(
                          'Sekali bayar. Tanpa biaya bulanan.'.localized(
                            context,
                          ),
                          style: theme.textTheme.bodySmall,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  DecoratedBox(
                    decoration: BoxDecoration(
                      color: sakura.goldSoft,
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 5,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.workspace_premium_rounded,
                            size: 14,
                            color: sakura.goldText,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'LIFETIME',
                            style: TextStyle(
                              color: sakura.goldText,
                              fontSize: 10.5,
                              fontWeight: FontWeight.w800,
                              letterSpacing: 0.8,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
              if (message != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(
                    message!.localized(context),
                    key: const Key('billing-message'),
                    style: theme.textTheme.bodySmall,
                  ),
                ),
              const SizedBox(height: 12),
              ...children,
            ],
          ),
        ),
      ),
    );
  }
}

class _LifetimeHero extends StatelessWidget {
  const _LifetimeHero({required this.active});
  final bool active;
  @override
  Widget build(BuildContext context) {
    final sakura = SakuraColors.of(context);
    return Container(
      clipBehavior: Clip.antiAlias,
      height: 188,
      decoration: BoxDecoration(
        color: sakura.heroStart,
        borderRadius: BorderRadius.circular(24),
      ),
      child: Stack(
        fit: StackFit.expand,
        children: [
          Image.asset(
            'assets/images/explore/lifetime-access.webp',
            key: const Key('lifetime-illustration'),
            fit: BoxFit.cover,
            alignment: Alignment.centerRight,
            cacheWidth: 1200,
            semanticLabel:
                'Perjalanan belajar bahasa dan meraih peluang di Jepang'
                    .localized(context),
            errorBuilder: (_, __, ___) => Align(
              alignment: Alignment.centerRight,
              child: Padding(
                padding: const EdgeInsets.only(right: 20),
                child: Icon(
                  Icons.workspace_premium_rounded,
                  size: 72,
                  color: sakura.gold,
                ),
              ),
            ),
          ),
          DecoratedBox(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  sakura.heroStart,
                  sakura.heroStart.withValues(alpha: .9),
                  sakura.heroEnd.withValues(alpha: .25),
                ],
                stops: const [0, .45, 1],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                DecoratedBox(
                  decoration: BoxDecoration(
                    color: sakura.gold,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          active
                              ? Icons.verified_rounded
                              : Icons.workspace_premium_rounded,
                          size: 14,
                          color: sakura.onGold,
                        ),
                        const SizedBox(width: 5),
                        Flexible(
                          child: Text(
                            active
                                ? 'LIFETIME AKTIF'.localized(context)
                                : 'SEKALI BAYAR · AKSES SELAMANYA'.localized(
                                    context,
                                  ),
                            style: TextStyle(
                              color: sakura.onGold,
                              fontWeight: FontWeight.w800,
                              fontSize: 9.5,
                              letterSpacing: 0.4,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                FractionallySizedBox(
                  widthFactor: 0.72,
                  alignment: Alignment.centerLeft,
                  child: Text(
                    active
                        ? 'Lifetime Access aktif'.localized(context)
                        : 'Buka lebih banyak.\nMelangkah lebih jauh.'
                              .localized(context),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 21,
                      fontWeight: FontWeight.w800,
                      height: 1.2,
                      letterSpacing: -0.4,
                    ),
                  ),
                ),
                const SizedBox(height: 6),
                FractionallySizedBox(
                  widthFactor: 0.72,
                  alignment: Alignment.centerLeft,
                  child: Text(
                    'Belajar, baca, latihan, dan temukan peluang dalam satu akses.'
                        .localized(context),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: .8),
                      fontSize: 12,
                      height: 1.4,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Empat manfaat Lifetime Access dalam grid ringkas 2×2.
class _BenefitGrid extends StatelessWidget {
  const _BenefitGrid();

  static const _items = [
    (Icons.block_rounded, 'Bebas iklan', 'Tanpa iklan selamanya'),
    (Icons.download_for_offline_rounded, 'Unduh offline', 'Belajar tanpa internet'),
    (Icons.auto_stories_rounded, 'Baca manga', 'Koleksi manga Jepang'),
    (Icons.work_rounded, 'Lamar kerja', 'Tautan lowongan aktif'),
  ];

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final width = (constraints.maxWidth - 10) / 2;
        return Wrap(
          spacing: 10,
          runSpacing: 10,
          children: [
            for (final (index, item) in _items.indexed)
              SizedBox(
                width: width,
                child: StaggeredEntrance(
                  index: index + 1,
                  child: _BenefitTile(
                    icon: item.$1,
                    title: item.$2,
                    subtitle: item.$3,
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _BenefitTile extends StatelessWidget {
  const _BenefitTile({
    required this.icon,
    required this.title,
    required this.subtitle,
  });
  final IconData icon;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SakuraCard(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 34,
            height: 34,
            decoration: BoxDecoration(
              color: theme.colorScheme.primaryContainer,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              icon,
              size: 18,
              color: theme.brightness == Brightness.dark
                  ? theme.colorScheme.primary
                  : theme.colorScheme.onPrimaryContainer,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            title.localized(context),
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          Text(
            subtitle.localized(context),
            maxLines: 2,
            style: theme.textTheme.bodySmall?.copyWith(fontSize: 11.5),
          ),
        ],
      ),
    );
  }
}

class _PriceNote extends StatelessWidget {
  const _PriceNote(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 6),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(
          Icons.check_circle_rounded,
          size: 16,
          color: SakuraColors.of(context).success,
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            text.localized(context),
            style: Theme.of(
              context,
            ).textTheme.bodySmall?.copyWith(height: 1.4),
          ),
        ),
      ],
    ),
  );
}
