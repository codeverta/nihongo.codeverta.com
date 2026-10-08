import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/network/remote_asset_image.dart';

const _pageCount = 3;
const _homePath = '/home';

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({
    this.onCompleted,
    super.key,
  });

  final Future<void> Function()? onCompleted;

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _pageController = PageController();
  int _currentPage = 0;
  bool _isCompleting = false;

  bool get _isLastPage => _currentPage == _pageCount - 1;

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  Future<void> _nextPage() async {
    await _pageController.nextPage(
      duration: const Duration(milliseconds: 420),
      curve: Curves.easeOutCubic,
    );
  }

  Future<void> _completeOnboarding() async {
    if (_isCompleting) return;

    setState(() => _isCompleting = true);
    await widget.onCompleted?.call();

    if (mounted) context.go(_homePath);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final horizontalPadding =
                (constraints.maxWidth * 0.07).clamp(20.0, 32.0);

            return Column(
              children: [
                SizedBox(
                  height: 56,
                  child: Padding(
                    padding: EdgeInsets.symmetric(
                      horizontal: horizontalPadding - 8,
                    ),
                    child: Align(
                      alignment: Alignment.centerRight,
                      child: AnimatedOpacity(
                        opacity: _isLastPage ? 0 : 1,
                        duration: const Duration(milliseconds: 200),
                        child: IgnorePointer(
                          ignoring: _isLastPage || _isCompleting,
                          child: TextButton(
                            onPressed: _completeOnboarding,
                            child: const Text('Lewati (Skip)'),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
                Expanded(
                  child: PageView(
                    controller: _pageController,
                    physics: const BouncingScrollPhysics(),
                    onPageChanged: (index) {
                      setState(() => _currentPage = index);
                    },
                    children: const [
                      _MasterJapanesePage(),
                      _ExamPreparationPage(),
                      _LearnAnywherePage(),
                    ],
                  ),
                ),
                _PageIndicator(
                  currentPage: _currentPage,
                  pageCount: _pageCount,
                ),
                const SizedBox(height: 24),
                Padding(
                  padding: EdgeInsets.fromLTRB(
                    horizontalPadding,
                    0,
                    horizontalPadding,
                    20,
                  ),
                  child: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 250),
                    switchInCurve: Curves.easeOut,
                    switchOutCurve: Curves.easeIn,
                    child: _isLastPage
                        ? SizedBox(
                            key: const ValueKey('start-button'),
                            width: double.infinity,
                            child: ElevatedButton.icon(
                              onPressed:
                                  _isCompleting ? null : _completeOnboarding,
                              icon: _isCompleting
                                  ? const SizedBox.square(
                                      dimension: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2.2,
                                        color: Colors.white,
                                      ),
                                    )
                                  : const Icon(Icons.rocket_launch_rounded),
                              label: const Text('Mulai Belajar'),
                            ),
                          )
                        : Align(
                            key: const ValueKey('next-button'),
                            alignment: Alignment.centerRight,
                            child: FilledButton.icon(
                              onPressed: _nextPage,
                              style: FilledButton.styleFrom(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 22,
                                  vertical: 16,
                                ),
                                shape: const StadiumBorder(),
                              ),
                              iconAlignment: IconAlignment.end,
                              icon: const Icon(Icons.arrow_forward_rounded),
                              label: const Text('Selanjutnya'),
                            ),
                          ),
                  ),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _MasterJapanesePage extends StatelessWidget {
  const _MasterJapanesePage();

  @override
  Widget build(BuildContext context) {
    return const _OnboardingPageLayout(
      semanticLabel: 'Ilustrasi perjalanan belajar bahasa Jepang',
      title: 'Kuasai Bahasa Jepang',
      subtitle: 'Belajar dari N5 hingga N1 dengan metode interaktif.',
      artwork: _JourneyArtwork(),
    );
  }
}

class _ExamPreparationPage extends StatelessWidget {
  const _ExamPreparationPage();

  @override
  Widget build(BuildContext context) {
    return const _OnboardingPageLayout(
      semanticLabel: 'Ilustrasi persiapan ujian bahasa Jepang',
      eyebrow: 'TARGET YANG JELAS',
      title: 'Persiapan Ujian\nTepat Sasaran',
      subtitle: 'Materi khusus untuk JLPT, NAT, dan JFT-Basic.',
      artwork: _IconArtwork(
        icon: Icons.fact_check_rounded,
        supportingIcon: Icons.workspace_premium_rounded,
      ),
    );
  }
}

class _LearnAnywherePage extends StatelessWidget {
  const _LearnAnywherePage();

  @override
  Widget build(BuildContext context) {
    return const _OnboardingPageLayout(
      semanticLabel: 'Ilustrasi belajar secara offline kapan saja',
      eyebrow: 'TEMAN BELAJARMU',
      title: 'Belajar Kapan Saja',
      subtitle: 'Akses materi offline dan raih pencapaian belajarmu.',
      artwork: _IconArtwork(
        icon: Icons.offline_bolt_rounded,
        supportingIcon: Icons.military_tech_rounded,
      ),
    );
  }
}

class _OnboardingPageLayout extends StatelessWidget {
  const _OnboardingPageLayout({
    required this.semanticLabel,
    required this.title,
    required this.subtitle,
    required this.artwork,
    this.eyebrow,
  });

  final String semanticLabel;
  final String? eyebrow;
  final String title;
  final String subtitle;
  final Widget artwork;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;

    return LayoutBuilder(
      builder: (context, constraints) {
        final horizontalPadding =
            (constraints.maxWidth * 0.075).clamp(24.0, 36.0);
        final artworkHeight =
            (constraints.maxHeight * 0.56).clamp(240.0, 390.0);

        return SingleChildScrollView(
          padding: EdgeInsets.fromLTRB(
            horizontalPadding,
            4,
            horizontalPadding,
            20,
          ),
          child: Column(
            children: [
              Semantics(
                image: true,
                label: semanticLabel,
                child: SizedBox(
                  height: artworkHeight,
                  width: double.infinity,
                  child: artwork,
                ),
              ),
              const SizedBox(height: 24),
              if (eyebrow != null) ...[
                Text(
                  eyebrow!,
                  style: textTheme.labelMedium?.copyWith(
                    color: Theme.of(context).colorScheme.primary,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 1.8,
                  ),
                ),
                const SizedBox(height: 10),
              ],
              Text(
                title,
                textAlign: TextAlign.center,
                style: textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.w800,
                  height: 1.16,
                  letterSpacing: -0.6,
                ),
              ),
              const SizedBox(height: 14),
              ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 340),
                child: Text(
                  subtitle,
                  textAlign: TextAlign.center,
                  style: textTheme.bodyLarge?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.62),
                    height: 1.55,
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _JourneyArtwork extends StatelessWidget {
  const _JourneyArtwork();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return DecoratedBox(
      decoration: BoxDecoration(
        color: colorScheme.primary.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(32),
        border: Border.all(
          color: colorScheme.primary.withValues(alpha: 0.10),
        ),
        boxShadow: [
          BoxShadow(
            color: colorScheme.primary.withValues(alpha: 0.08),
            blurRadius: 32,
            offset: const Offset(0, 14),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(31),
        child: const RemoteAssetImage(
          asset: 'assets/images/onboarding-japanese-journey.png',
          url: '',
          fit: BoxFit.cover,
          alignment: Alignment(0, 0.22),
        ),
      ),
    );
  }
}

class _IconArtwork extends StatelessWidget {
  const _IconArtwork({
    required this.icon,
    required this.supportingIcon,
  });

  final IconData icon;
  final IconData supportingIcon;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Stack(
      alignment: Alignment.center,
      children: [
        Container(
          width: 270,
          height: 270,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: RadialGradient(
              colors: [
                colorScheme.primary.withValues(alpha: 0.16),
                colorScheme.primary.withValues(alpha: 0.03),
              ],
            ),
          ),
        ),
        Transform.rotate(
          angle: -0.06,
          child: Container(
            width: 202,
            height: 218,
            decoration: BoxDecoration(
              color: colorScheme.surface,
              borderRadius: BorderRadius.circular(36),
              border: Border.all(
                color: colorScheme.outlineVariant.withValues(alpha: 0.55),
              ),
              boxShadow: [
                BoxShadow(
                  color: colorScheme.shadow.withValues(alpha: 0.10),
                  blurRadius: 30,
                  offset: const Offset(0, 16),
                ),
              ],
            ),
            child: Icon(icon, size: 92, color: colorScheme.primary),
          ),
        ),
        Positioned(
          right: 32,
          bottom: 32,
          child: Container(
            width: 72,
            height: 72,
            decoration: BoxDecoration(
              color: colorScheme.secondary,
              shape: BoxShape.circle,
              border: Border.all(color: colorScheme.surface, width: 5),
              boxShadow: [
                BoxShadow(
                  color: colorScheme.shadow.withValues(alpha: 0.16),
                  blurRadius: 18,
                  offset: const Offset(0, 8),
                ),
              ],
            ),
            child: Icon(
              supportingIcon,
              size: 34,
              color: colorScheme.onSecondary,
            ),
          ),
        ),
      ],
    );
  }
}

class _PageIndicator extends StatelessWidget {
  const _PageIndicator({
    required this.currentPage,
    required this.pageCount,
  });

  final int currentPage;
  final int pageCount;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Semantics(
      label: 'Halaman ${currentPage + 1} dari $pageCount',
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: List.generate(pageCount, (index) {
          final isActive = index == currentPage;

          return AnimatedContainer(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeOutCubic,
            width: isActive ? 30 : 8,
            height: 8,
            margin: const EdgeInsets.symmetric(horizontal: 4),
            decoration: BoxDecoration(
              color:
                  isActive ? colorScheme.primary : colorScheme.outlineVariant,
              borderRadius: BorderRadius.circular(99),
            ),
          );
        }),
      ),
    );
  }
}
