import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

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

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({this.onCompleted, this.isPreview = false, super.key});

  final Future<void> Function()? onCompleted;
  final bool isPreview;

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _controller = PageController();
  int _current = 0;
  bool _busy = false;
  bool get _last => _current == _pages.length - 1;

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
                    itemCount: _pages.length,
                    onPageChanged: (page) => setState(() => _current = page),
                    itemBuilder: (context, index) =>
                        _PageBody(page: _pages[index], index: index),
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
                          for (var i = 0; i < _pages.length; i++)
                            Semantics(
                              selected: i == _current,
                              label: 'Halaman ${i + 1} dari ${_pages.length}'
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
  const _PageBody({required this.page, required this.index});
  final _Page page;
  final int index;

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
                '0${index + 1} / 03  ·  ${page.eyebrow.localized(context)}',
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
