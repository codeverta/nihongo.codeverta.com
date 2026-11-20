import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';

const _ink = Color(0xFF24285C);
const _red = Color(0xFFD84942);
const _pages = [
  _Page(
    asset: 'journey.webp',
    eyebrow: 'PERJALANAN BARU',
    title: 'Kuasai Bahasa Jepang',
    description:
        'Dari kata pertama sampai percakapan yang lebih percaya diri. Temukan ritme belajarmu, dari N5 hingga N1.',
    features: ['N5 → N1', 'Kosakata & kanji', 'Grammar & reading'],
    semanticLabel:
        'Ilustrasi buku belajar, sakura, dan perjalanan menuju Gunung Fuji',
  ),
  _Page(
    asset: 'exams.webp',
    eyebrow: 'LANGKAH YANG TERARAH',
    title: 'Siap untuk\nujian berikutnya.',
    description:
        'Pahami materinya, coba latihan soalnya, lalu ukur kemampuanmu lewat simulasi ujian.',
    features: ['JLPT', 'NAT', 'JFT-Basic'],
    semanticLabel:
        'Ilustrasi meja belajar, lembar ujian, dan medali pencapaian',
  ),
  _Page(
    asset: 'anywhere.webp',
    eyebrow: 'SEDIKIT, SETIAP HARI',
    title: 'Belajarmu ikut\nke mana pun.',
    description:
        'Buka lagi materi yang sudah tersimpan, bahkan saat offline. Lanjutkan perjalananmu, satu langkah setiap hari.',
    features: ['Materi tersimpan', 'Progres belajar'],
    semanticLabel:
        'Ilustrasi buku dan kartu belajar di perjalanan kereta Jepang',
  ),
];

/// Hooks for the optional last page that offers to turn on notifications.
/// The page only exists when this is provided, so the guide stays three pages
/// for people who cannot or already did enable them.
class OnboardingNotificationOptIn {
  const OnboardingNotificationOptIn({required this.enable, required this.skip});

  /// Shows the system permission dialog and subscribes to announcements.
  final Future<PushToggleResult> Function() enable;

  /// Records "Nanti saja" without any system dialog.
  final Future<void> Function() skip;
}

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({
    this.onCompleted,
    this.isPreview = false,
    this.notificationOptIn,
    super.key,
  });

  final Future<void> Function()? onCompleted;
  final bool isPreview;
  final OnboardingNotificationOptIn? notificationOptIn;

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _controller = PageController();
  int _current = 0;
  bool _busy = false;
  bool _enabling = false;
  bool _notificationSeen = false;
  PushToggleResult? _notificationResult;
  bool get _hasNotificationPage => widget.notificationOptIn != null;
  int get _pageCount => _pages.length + (_hasNotificationPage ? 1 : 0);
  bool get _last => _current == _pageCount - 1;

  Future<void> _enableNotifications() async {
    final optIn = widget.notificationOptIn;
    if (optIn == null || _enabling) return;
    setState(() => _enabling = true);
    PushToggleResult result;
    try {
      result = await optIn.enable();
    } on Object {
      result = PushToggleResult.failed;
    }
    if (!mounted) return;
    setState(() {
      _enabling = false;
      _notificationResult = result;
    });
  }

  Future<void> _skipNotifications() async {
    final optIn = widget.notificationOptIn;
    if (optIn == null || _notificationResult != null) return;
    setState(() => _notificationResult = PushToggleResult.off);
    try {
      await optIn.skip();
    } on Object {
      // Never block onboarding on notification bookkeeping.
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _move(int index) {
    if (_busy) return;
    _controller.animateToPage(
      index,
      duration: const Duration(milliseconds: 380),
      curve: Curves.easeOutCubic,
    );
  }

  Future<void> _finish() async {
    if (_busy) return;
    if (widget.isPreview) {
      if (context.canPop()) {
        context.pop();
      } else {
        context.go('/home');
      }
      return;
    }
    // Reaching the notification page and leaving without choosing counts as
    // "Nanti saja", so the person is never asked twice.
    if (_notificationSeen && _notificationResult == null) {
      await _skipNotifications();
    }
    if (!mounted) return;
    setState(() => _busy = true);
    try {
      await widget.onCompleted?.call();
      if (mounted) context.go('/home');
    } on Object {
      if (!mounted) return;
      setState(() => _busy = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Panduan belum dapat diselesaikan. Coba lagi.'.localized(context),
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final dark = colors.brightness == Brightness.dark;
    final foreground = dark ? colors.onSurface : _ink;
    return Scaffold(
      backgroundColor: dark ? colors.surface : const Color(0xFFFAF8F4),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 600),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(24, 12, 16, 4),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(9),
                        decoration: BoxDecoration(
                          color: _red,
                          borderRadius: BorderRadius.circular(13),
                        ),
                        child: const Icon(
                          Icons.auto_stories_rounded,
                          color: Colors.white,
                          size: 21,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          'Nihongo\nMaster',
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: Theme.of(context).textTheme.titleSmall
                              ?.copyWith(
                                fontWeight: FontWeight.w800,
                                height: 1.12,
                                color: foreground,
                              ),
                        ),
                      ),
                      if (widget.isPreview)
                        IconButton(
                          key: const Key('onboarding-close'),
                          tooltip: 'Tutup panduan'.localized(context),
                          onPressed: _finish,
                          icon: const Icon(Icons.close_rounded),
                        )
                      else
                        TextButton(
                          key: const Key('onboarding-skip'),
                          onPressed: _busy ? null : _finish,
                          style: TextButton.styleFrom(
                            foregroundColor: foreground,
                          ),
                          child: Text('Lewati'.localized(context)),
                        ),
                    ],
                  ),
                ),
                Expanded(
                  child: PageView.builder(
                    key: const Key('onboarding-pages'),
                    controller: _controller,
                    itemCount: _pageCount,
                    onPageChanged: (page) => setState(() {
                      _current = page;
                      if (page >= _pages.length) _notificationSeen = true;
                    }),
                    itemBuilder: (context, index) => index < _pages.length
                        ? _PageBody(
                            page: _pages[index],
                            index: index,
                            total: _pageCount,
                          )
                        : _NotificationPage(
                            index: index,
                            total: _pageCount,
                            result: _notificationResult,
                            enabling: _enabling,
                            onEnable: _enableNotifications,
                            onSkip: _skipNotifications,
                          ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.fromLTRB(24, 12, 24, 20),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          for (var i = 0; i < _pageCount; i++)
                            Semantics(
                              selected: i == _current,
                              label: 'Halaman ${i + 1} dari $_pageCount'
                                  .localized(context),
                              child: InkWell(
                                key: Key('onboarding-page-$i'),
                                onTap: () => _move(i),
                                borderRadius: BorderRadius.circular(20),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 4,
                                    vertical: 14,
                                  ),
                                  child: AnimatedContainer(
                                    duration: const Duration(milliseconds: 250),
                                    width: i == _current ? 42 : 10,
                                    height: 6,
                                    decoration: BoxDecoration(
                                      color: i == _current
                                          ? _red
                                          : foreground.withValues(alpha: 0.15),
                                      borderRadius: BorderRadius.circular(20),
                                    ),
                                  ),
                                ),
                              ),
                            ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          if (_current > 0) ...[
                            IconButton.outlined(
                              key: const Key('onboarding-back'),
                              tooltip: 'Kembali'.localized(context),
                              onPressed: _busy
                                  ? null
                                  : () => _move(_current - 1),
                              style: IconButton.styleFrom(
                                minimumSize: const Size(56, 56),
                                foregroundColor: foreground,
                              ),
                              icon: const Icon(Icons.arrow_back_rounded),
                            ),
                            const SizedBox(width: 12),
                          ],
                          Expanded(
                            child: FilledButton(
                              key: const Key('onboarding-next'),
                              onPressed: _busy
                                  ? null
                                  : _last
                                  ? _finish
                                  : () => _move(_current + 1),
                              style: FilledButton.styleFrom(
                                backgroundColor: dark ? colors.primary : _ink,
                                foregroundColor: dark
                                    ? colors.onPrimary
                                    : Colors.white,
                                minimumSize: const Size.fromHeight(56),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 20,
                                  vertical: 16,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(20),
                                ),
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  if (_busy)
                                    const SizedBox.square(
                                      dimension: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                      ),
                                    ),
                                  if (_busy) const SizedBox(width: 10),
                                  Flexible(
                                    child: Text(
                                      (_last
                                              ? widget.isPreview
                                                    ? 'Kembali ke aplikasi'
                                                    : 'Mulai Belajar'
                                              : 'Selanjutnya')
                                          .localized(context),
                                      textAlign: TextAlign.center,
                                    ),
                                  ),
                                  const SizedBox(width: 10),
                                  Icon(
                                    _last
                                        ? Icons.check_rounded
                                        : Icons.arrow_forward_rounded,
                                    size: 20,
                                  ),
                                ],
                              ),
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
        ),
      ),
    );
  }
}

class _Page {
  const _Page({
    required this.asset,
    required this.eyebrow,
    required this.title,
    required this.description,
    required this.features,
    required this.semanticLabel,
  });
  final String asset;
  final String eyebrow;
  final String title;
  final String description;
  final List<String> features;
  final String semanticLabel;
}

class _PageBody extends StatelessWidget {
  const _PageBody({
    required this.page,
    required this.index,
    required this.total,
  });
  final _Page page;
  final int index;
  final int total;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final ink = colors.brightness == Brightness.dark ? colors.onSurface : _ink;
    return LayoutBuilder(
      builder: (context, constraints) {
        final imageHeight = (constraints.maxHeight * 0.52).clamp(145.0, 330.0);
        return SingleChildScrollView(
          key: Key('onboarding-content-$index'),
          padding: const EdgeInsets.fromLTRB(24, 12, 24, 8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(32),
                child: Image.asset(
                  'assets/images/onboarding/${page.asset}',
                  key: Key('onboarding-art-$index'),
                  width: double.infinity,
                  height: imageHeight,
                  fit: BoxFit.cover,
                  semanticLabel: page.semanticLabel.localized(context),
                ),
              ),
              const SizedBox(height: 24),
              Text(
                '0${index + 1} / 0$total  ·  ${page.eyebrow.localized(context)}',
                style: theme.textTheme.labelSmall?.copyWith(
                  color: colors.brightness == Brightness.dark
                      ? colors.primary
                      : _red,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.2,
                ),
              ),
              const SizedBox(height: 12),
              Text(
                page.title.localized(context),
                style: theme.textTheme.headlineMedium?.copyWith(
                  color: ink,
                  fontSize: 32,
                  fontWeight: FontWeight.w800,
                  letterSpacing: -1.1,
                  height: 1.16,
                ),
              ),
              const SizedBox(height: 14),
              Text(
                page.description.localized(context),
                style: theme.textTheme.bodyMedium?.copyWith(
                  height: 1.65,
                  color: ink.withValues(alpha: 0.7),
                ),
              ),
              const SizedBox(height: 18),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (final feature in page.features)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 7,
                      ),
                      decoration: BoxDecoration(
                        color: ink.withValues(alpha: 0.05),
                        borderRadius: BorderRadius.circular(30),
                        border: Border.all(color: ink.withValues(alpha: 0.08)),
                      ),
                      child: Text(
                        feature.localized(context),
                        style: theme.textTheme.labelSmall?.copyWith(
                          color: ink,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }
}

/// Last page: explains what notifications are for, then asks. The system
/// permission dialog only appears after "Aktifkan notifikasi" is tapped.
class _NotificationPage extends StatelessWidget {
  const _NotificationPage({
    required this.index,
    required this.total,
    required this.result,
    required this.enabling,
    required this.onEnable,
    required this.onSkip,
  });

  final int index;
  final int total;
  final PushToggleResult? result;
  final bool enabling;
  final VoidCallback onEnable;
  final VoidCallback onSkip;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final dark = colors.brightness == Brightness.dark;
    final ink = dark ? colors.onSurface : _ink;
    final message = switch (result) {
      PushToggleResult.on => 'Notifikasi aktif. Terima kasih!',
      PushToggleResult.denied =>
        'Izin belum diberikan. Kamu bisa mengaktifkannya nanti lewat pengaturan Android.',
      PushToggleResult.failed =>
        'Notifikasi belum dapat diaktifkan sekarang. Coba lagi nanti di Pengaturan.',
      _ => null,
    };
    // "off" means the person chose "Nanti saja": nothing more to say or ask.
    final answered = result != null && result != PushToggleResult.failed;
    return LayoutBuilder(
      builder: (context, constraints) => SingleChildScrollView(
        key: Key('onboarding-content-$index'),
        padding: const EdgeInsets.fromLTRB(24, 12, 24, 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              key: const Key('onboarding-notification-art'),
              width: double.infinity,
              // Compact on short screens so the buttons stay in view.
              height: (constraints.maxHeight * 0.17).clamp(64.0, 130.0),
              decoration: BoxDecoration(
                color: ink.withValues(alpha: 0.06),
                borderRadius: BorderRadius.circular(32),
              ),
              child: Icon(
                Icons.notifications_active_rounded,
                size: (constraints.maxHeight * 0.1).clamp(40.0, 64.0),
                color: _red,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              '0${index + 1} / 0$total  ·  ${'TETAP TERHUBUNG'.localized(context)}',
              style: theme.textTheme.labelSmall?.copyWith(
                color: dark ? colors.primary : _red,
                fontWeight: FontWeight.w800,
                letterSpacing: 1.2,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              'Jangan ketinggalan kabar penting'.localized(context),
              style: theme.textTheme.headlineMedium?.copyWith(
                color: ink,
                fontSize: 32,
                fontWeight: FontWeight.w800,
                letterSpacing: -1.1,
                height: 1.16,
              ),
            ),
            const SizedBox(height: 14),
            Text(
              'Aktifkan notifikasi untuk menerima pengumuman, materi baru, dan info penting dari Nihongo Master. Tanpa spam, dan bisa dimatikan kapan saja di Pengaturan.'
                  .localized(context),
              style: theme.textTheme.bodyMedium?.copyWith(
                height: 1.65,
                color: ink.withValues(alpha: 0.7),
              ),
            ),
            const SizedBox(height: 18),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final feature in const [
                  'Pengumuman',
                  'Materi baru',
                  'Info penting',
                ])
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 7,
                    ),
                    decoration: BoxDecoration(
                      color: ink.withValues(alpha: 0.05),
                      borderRadius: BorderRadius.circular(30),
                      border: Border.all(color: ink.withValues(alpha: 0.08)),
                    ),
                    child: Text(
                      feature.localized(context),
                      style: theme.textTheme.labelSmall?.copyWith(
                        color: ink,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 20),
            if (message != null)
              Text(
                message.localized(context),
                key: const Key('onboarding-notification-result'),
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: ink,
                ),
              ),
            if (!answered) ...[
              if (message != null) const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  key: const Key('onboarding-notification-enable'),
                  onPressed: enabling ? null : onEnable,
                  icon: enabling
                      ? const SizedBox.square(
                          dimension: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.notifications_active_outlined),
                  label: Text('Aktifkan notifikasi'.localized(context)),
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size.fromHeight(52),
                    foregroundColor: ink,
                  ),
                ),
              ),
              Align(
                child: TextButton(
                  key: const Key('onboarding-notification-skip'),
                  onPressed: enabling ? null : onSkip,
                  child: Text('Nanti saja'.localized(context)),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
