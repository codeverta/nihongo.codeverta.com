import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/branding/app_branding.dart';

/// The same application identity used by startup, with immediate asset fallback.
class BrandEmblem extends ConsumerWidget {
  const BrandEmblem({this.size = 88, this.premium = false, super.key});

  final double size;
  final bool premium;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final branding =
        ref.watch(appBrandingProvider).valueOrNull ??
        ref.watch<AppBranding>(cachedBrandingProvider);
    return ExcludeSemantics(
      child: SizedBox.square(
        dimension: size,
        child: Stack(
          clipBehavior: Clip.none,
          children: [
            Positioned.fill(
              child: Container(
                padding: EdgeInsets.all(size * .09),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(size * .25),
                  border: Border.all(color: const Color(0xFFE8E8F4)),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF191C66).withValues(alpha: .16),
                      blurRadius: 24,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(size * .16),
                  child: BrandIcon(branding: branding, size: size * .82),
                ),
              ),
            ),
            if (premium)
              Positioned(
                right: -5,
                bottom: -5,
                child: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFD784),
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: const Color(0xFF343594),
                      width: 3,
                    ),
                  ),
                  child: const Icon(
                    Icons.workspace_premium_rounded,
                    color: Color(0xFF5A370D),
                    size: 21,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class LoginArtwork extends StatelessWidget {
  const LoginArtwork({super.key});

  @override
  Widget build(BuildContext context) => ExcludeSemantics(
    child: SizedBox(
      height: 164,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(28),
        child: Stack(
          alignment: Alignment.center,
          children: [
            const Positioned.fill(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [Color(0xFFEEEEFF), Color(0xFFFDF1EE)],
                  ),
                ),
              ),
            ),
            const Positioned(
              right: -28,
              top: -65,
              child: _ArtCircle(size: 180, color: Color(0xFFE5E3FA)),
            ),
            const Positioned(
              left: -40,
              bottom: -90,
              child: _ArtCircle(size: 180, color: Color(0xFFF9DEDB)),
            ),
            Positioned(
              left: 30,
              top: 28,
              child: Transform.rotate(
                angle: -.16,
                child: const _KanaTile(text: 'あ', color: Color(0xFFCD344D)),
              ),
            ),
            Positioned(
              right: 30,
              bottom: 24,
              child: Transform.rotate(
                angle: .14,
                child: const _KanaTile(text: '学', color: Color(0xFF35378D)),
              ),
            ),
            Transform.rotate(angle: -.05, child: const BrandEmblem(size: 104)),
          ],
        ),
      ),
    ),
  );
}

class _ArtCircle extends StatelessWidget {
  const _ArtCircle({required this.size, required this.color});
  final double size;
  final Color color;

  @override
  Widget build(BuildContext context) => Container(
    width: size,
    height: size,
    decoration: BoxDecoration(shape: BoxShape.circle, color: color),
  );
}

class _KanaTile extends StatelessWidget {
  const _KanaTile({required this.text, required this.color});
  final String text;
  final Color color;

  @override
  Widget build(BuildContext context) => Container(
    width: 46,
    height: 52,
    alignment: Alignment.center,
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: .85),
      borderRadius: BorderRadius.circular(13),
      border: Border.all(color: Colors.white),
    ),
    child: Text(
      text,
      style: TextStyle(fontSize: 27, fontWeight: FontWeight.w700, color: color),
    ),
  );
}
