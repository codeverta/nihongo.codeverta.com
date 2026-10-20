import 'package:flutter/material.dart';

/// Illustrated entry point shared by the home career cards.
class ExploreFeatureCard extends StatelessWidget {
  const ExploreFeatureCard({
    required this.title,
    required this.description,
    required this.eyebrow,
    required this.asset,
    required this.color,
    required this.icon,
    required this.onTap,
    this.premium = false,
    super.key,
  });
  final String title;
  final String description;
  final String eyebrow;
  final String asset;
  final Color color;
  final IconData icon;
  final VoidCallback onTap;
  final bool premium;

  @override
  Widget build(BuildContext context) => Card(
    margin: EdgeInsets.zero,
    clipBehavior: Clip.antiAlias,
    child: InkWell(
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            height: 144,
            width: double.infinity,
            child: Stack(
              fit: StackFit.expand,
              children: [
                Image.asset(
                  asset,
                  fit: BoxFit.cover,
                  cacheWidth: 1000,
                  excludeFromSemantics: true,
                  errorBuilder: (_, __, ___) => ColoredBox(
                    color: color,
                    child: Icon(icon, color: Colors.white, size: 60),
                  ),
                ),
                DecoratedBox(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [
                        Colors.transparent,
                        color.withValues(alpha: .88),
                      ],
                    ),
                  ),
                ),
                if (premium)
                  Positioned(
                    right: 14,
                    top: 14,
                    child: DecoratedBox(
                      decoration: BoxDecoration(
                        color: const Color(0xFFFFD784),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Padding(
                        padding: EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 7,
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.workspace_premium_rounded,
                              color: Color(0xFF5C3F12),
                              size: 16,
                            ),
                            SizedBox(width: 4),
                            Text(
                              'Premium',
                              style: TextStyle(
                                color: Color(0xFF5C3F12),
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                Positioned(
                  left: 18,
                  bottom: 14,
                  right: 18,
                  child: Text(
                    eyebrow,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.w800,
                      fontSize: 11,
                      letterSpacing: 1,
                    ),
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(18),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: Theme.of(context).textTheme.titleMedium
                            ?.copyWith(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        description,
                        style: Theme.of(
                          context,
                        ).textTheme.bodySmall?.copyWith(height: 1.5),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 12),
                CircleAvatar(
                  backgroundColor: color.withValues(alpha: .1),
                  child: Icon(
                    Icons.arrow_forward_rounded,
                    color: color,
                    size: 20,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    ),
  );
}
