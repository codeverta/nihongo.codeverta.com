import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

/// Beranda quick menu: a 4×2 icon grid that replaces the large explore cards
/// and the old "Mulai belajar" section so every feature is reachable without
/// scrolling. Belajar, Ujian and Kamus stay in the bottom navigation.
class HomeQuickMenu extends StatelessWidget {
  const HomeQuickMenu({
    required this.openExamCategory,
    required this.hasLifetimeAccess,
    super.key,
  });

  final ValueChanged<String> openExamCategory;

  /// Hides the Premium crown on Manga for Lifetime Access accounts.
  final bool hasLifetimeAccess;

  @override
  Widget build(BuildContext context) {
    final items = <HomeQuickMenuItem>[
      HomeQuickMenuItem(
        id: 'jft',
        label: 'JFT',
        light: const Color(0xFFD7263D),
        dark: const Color(0xFFFF8593),
        accentLight: const Color(0xFFE0A100),
        accentDark: const Color(0xFFF4C66B),
        onTap: () => openExamCategory('JFT Test'),
      ),
      HomeQuickMenuItem(
        id: 'nat',
        label: 'NAT',
        light: const Color(0xFF3B42A8),
        dark: const Color(0xFFA9B0FF),
        onTap: () => openExamCategory('NAT Test'),
      ),
      HomeQuickMenuItem(
        id: 'manga',
        label: 'Manga',
        light: const Color(0xFF7C3AED),
        dark: const Color(0xFFC4A5FF),
        premium: !hasLifetimeAccess,
        onTap: () => context.push('/manga'),
      ),
      HomeQuickMenuItem(
        id: 'anki',
        label: 'Anki',
        light: const Color(0xFF0F8A7E),
        dark: const Color(0xFF5EEAD4),
        onTap: () => context.push('/anki'),
      ),
      HomeQuickMenuItem(
        id: 'ssw',
        label: 'SSW',
        light: const Color(0xFFB45309),
        dark: const Color(0xFFF4C66B),
        onTap: () => context.push('/ssw'),
      ),
      HomeQuickMenuItem(
        id: 'jobs',
        label: 'Lowongan',
        light: const Color(0xFF2563EB),
        dark: const Color(0xFF93C5FD),
        onTap: () => context.push('/jobs'),
      ),
      HomeQuickMenuItem(
        id: 'map',
        label: 'Peta',
        light: const Color(0xFF15803D),
        dark: const Color(0xFF4ADE80),
        onTap: () => context.push('/map'),
      ),
      HomeQuickMenuItem(
        id: 'news',
        label: 'Berita',
        light: const Color(0xFFC2410C),
        dark: const Color(0xFFFDBA74),
        onTap: () => context.push('/news'),
      ),
    ];
    return LayoutBuilder(
      builder: (context, constraints) {
        // Four columns on phones; tablets get one row of eight.
        final columns = constraints.maxWidth >= 600 ? 8 : 4;
        final spacing = constraints.maxWidth < 330 ? 2.0 : 6.0;
        final width = (constraints.maxWidth - spacing * (columns - 1)) / columns;
        return Wrap(
          key: const Key('home-quick-menu'),
          spacing: spacing,
          runSpacing: 14,
          children: [
            for (var i = 0; i < items.length; i++)
              SizedBox(
                width: width,
                child: StaggeredEntrance(
                  index: i,
                  step: const Duration(milliseconds: 35),
                  child: _QuickMenuTile(item: items[i]),
                ),
              ),
          ],
        );
      },
    );
  }
}

@immutable
class HomeQuickMenuItem {
  const HomeQuickMenuItem({
    required this.id,
    required this.label,
    required this.light,
    required this.dark,
    required this.onTap,
    this.accentLight,
    this.accentDark,
    this.premium = false,
  });

  final String id;
  final String label;
  final Color light;
  final Color dark;
  final Color? accentLight;
  final Color? accentDark;
  final bool premium;
  final VoidCallback onTap;
}

class _QuickMenuTile extends StatelessWidget {
  const _QuickMenuTile({required this.item});

  final HomeQuickMenuItem item;

  static const _tile = 56.0;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final color = isDark ? item.dark : item.light;
    final accent = isDark
        ? (item.accentDark ?? const Color(0xFFFF8593))
        : (item.accentLight ?? const Color(0xFFD7263D));
    final surface = theme.colorScheme.surface;
    final label = item.label.localized(context);
    return Pressable(
      key: Key('home-quick-${item.id}'),
      onTap: item.onTap,
      scale: 0.9,
      borderRadius: BorderRadius.circular(18),
      semanticsLabel: item.premium
          ? '$label · ${'Premium'.localized(context)}'
          : label,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          SizedBox(
            width: _tile + 6,
            height: _tile + 6,
            child: Stack(
              clipBehavior: Clip.none,
              alignment: Alignment.center,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(18),
                  child: SizedBox(
                    width: _tile,
                    height: _tile,
                    child: Stack(
                      children: [
                        Positioned.fill(
                          child: ColoredBox(
                            color: color.withValues(alpha: isDark ? 0.16 : 0.12),
                          ),
                        ),
                        // Soft "pressed edge" along the bottom of the tile.
                        Positioned(
                          left: 0,
                          right: 0,
                          bottom: 0,
                          height: 3,
                          child: ColoredBox(
                            color: color.withValues(alpha: 0.22),
                          ),
                        ),
                        Center(
                          child: SvgPicture.string(
                            homeQuickMenuSvg(
                              item.id,
                              color: color,
                              accent: accent,
                              surface: surface,
                            ),
                            width: 30,
                            height: 30,
                            excludeFromSemantics: true,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                if (item.premium)
                  Positioned(
                    top: -2,
                    right: -2,
                    child: _PremiumCrown(ring: theme.scaffoldBackgroundColor),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            textAlign: TextAlign.center,
            style: theme.textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.w700,
              color: theme.colorScheme.onSurface,
            ),
          ),
        ],
      ),
    );
  }
}

class _PremiumCrown extends StatelessWidget {
  const _PremiumCrown({required this.ring});

  final Color ring;

  @override
  Widget build(BuildContext context) {
    final sakura = SakuraColors.of(context);
    return Container(
      key: const Key('home-quick-premium'),
      width: 21,
      height: 21,
      decoration: BoxDecoration(
        color: sakura.gold,
        shape: BoxShape.circle,
        border: Border.all(color: ring, width: 2.5),
      ),
      alignment: Alignment.center,
      child: Icon(Icons.workspace_premium_rounded, size: 12, color: sakura.onGold),
    );
  }
}

String _hex(Color color) {
  final argb = color.toARGB32();
  return '#${(argb & 0xFFFFFF).toRadixString(16).padLeft(6, '0')}';
}

/// Builds the custom duotone icon for [id] (32×32 grid) with theme colors.
String homeQuickMenuSvg(
  String id, {
  required Color color,
  required Color accent,
  required Color surface,
}) {
  final body = _homeQuickMenuIcons[id] ?? '';
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" '
          'fill="none" stroke="{C}" stroke-width="1.9" '
          'stroke-linecap="round" stroke-linejoin="round">$body</svg>'
      .replaceAll('{C}', _hex(color))
      .replaceAll('{A}', _hex(accent))
      .replaceAll('{S}', _hex(surface));
}

const _homeQuickMenuIcons = <String, String>{
  'jft':
      '<path fill="{C}" fill-opacity="0.24" d="M8 3.5h11l5 5V25a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V5.5a2 2 0 0 1 2-2z"/><path d="M19 3.5v5h5M10 11h8M10 15h5"/><path fill="{A}" stroke="none" d="m17.5 25 -1.5 5 3-1.5 3 1.5-1.5-5z"/><circle fill="{A}" stroke="none" cx="21" cy="21.5" r="5.5"/><path stroke="{S}" stroke-width="2" d="m18.6 21.6 1.7 1.7 3.2-3.4"/>',
  'nat':
      '<rect fill="{C}" fill-opacity="0.24" x="4" y="5" width="16" height="22" rx="2.5"/><circle cx="9" cy="11" r="1.8"/><circle cx="15" cy="11" r="1.8" fill="{A}" stroke="none"/><circle cx="9" cy="16" r="1.8" fill="{A}" stroke="none"/><circle cx="15" cy="16" r="1.8"/><circle cx="9" cy="21" r="1.8"/><path fill="{S}" d="m24.5 9.5 3 3-8.5 8.5-4 1 1-4z"/><path d="m22.5 11.5 3 3"/>',
  'manga':
      '<rect fill="{C}" fill-opacity="0.24" x="3.5" y="8" width="19" height="19.5" rx="2.5"/><path d="M3.5 18h9M12.5 27.5 14 8"/><path fill="{S}" d="M22 2.5c4.1 0 7.5 2.6 7.5 5.8S26.1 14 22 14c-.7 0-1.4-.1-2-.2l-3.6 2 .9-3.4c-1.7-1-2.8-2.5-2.8-4.2 0-3.1 3.4-5.7 7.5-5.7z"/><path stroke="{A}" stroke-width="2.6" d="M19.3 8.3h.01M22 8.3h.01M24.7 8.3h.01"/><path d="m6.5 22 2.5 2.5M7.5 20.5l3.2 1.2"/>',
  'anki':
      '<rect fill="{C}" fill-opacity="0.24" x="10" y="4" width="16" height="20" rx="2.5" transform="rotate(12 18 14)"/><rect fill="{S}" x="5" y="8" width="16" height="20" rx="2.5"/><circle fill="{A}" stroke="none" cx="13" cy="15.5" r="3.2"/><path d="M9.5 22.5h7"/>',
  'ssw':
      '<path fill="{C}" fill-opacity="0.24" d="M6.5 21v-1.5a9.5 9.5 0 0 1 19 0V21z"/><path d="M16 10v7M11.6 11.6c-1 1.6-1.6 3.6-1.6 5.9M20.4 11.6c1 1.6 1.6 3.6 1.6 5.9"/><rect fill="{A}" stroke="none" x="3" y="21" width="26" height="4" rx="2"/>',
  'jobs':
      '<rect fill="{C}" fill-opacity="0.24" x="4" y="10" width="24" height="16" rx="3"/><path d="M12 10V8a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><g fill="{A}" stroke="none" transform="translate(16 18.5)"><ellipse rx="1.7" ry="2.7" cy="-2.6"/><ellipse rx="1.7" ry="2.7" cy="-2.6" transform="rotate(72)"/><ellipse rx="1.7" ry="2.7" cy="-2.6" transform="rotate(144)"/><ellipse rx="1.7" ry="2.7" cy="-2.6" transform="rotate(216)"/><ellipse rx="1.7" ry="2.7" cy="-2.6" transform="rotate(288)"/></g><circle fill="{S}" stroke="none" cx="16" cy="18.5" r="1"/>',
  'map':
      '<path fill="{C}" fill-opacity="0.24" d="m3.5 8 7.5-3 10 3 7.5-3v19.5l-7.5 3-10-3-7.5 3z"/><path d="M11 5v19.5M21 8v19.5" opacity=".45"/><path stroke-width="3" d="M8.5 23.5c3-1.4 5.3-3.4 7.3-6s3.2-5 4.2-7.8"/><path stroke-width="3" d="M22 7.3c.9-.6 2.2-.6 2.9.2"/><path stroke-width="3" d="M6.4 24.8h.3"/><path fill="{A}" stroke="none" d="M15 7.5a3.6 3.6 0 0 0-3.6 3.6c0 2.7 3.6 6.3 3.6 6.3s3.6-3.6 3.6-6.3A3.6 3.6 0 0 0 15 7.5z"/><circle fill="{S}" stroke="none" cx="15" cy="11.1" r="1.3"/>',
  'news':
      '<path fill="{C}" fill-opacity="0.24" d="M5 7h17v18a3 3 0 0 1-3 3H8a3 3 0 0 1-3-3z"/><path d="M22 12h4v13a3 3 0 0 1-3 3h-4"/><circle fill="{A}" stroke="none" cx="10.5" cy="13" r="2.7"/><path d="M15.5 11.5h3M15.5 14.5h3M9 19.5h9.5M9 23h6.5"/>',
};
