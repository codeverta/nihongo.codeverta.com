import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';

class MapEntryCard extends StatelessWidget {
  const MapEntryCard({super.key});
  @override
  Widget build(BuildContext context) => ExploreFeatureCard(
    title: 'Peta Jepang'.localized(context),
    eyebrow: 'BUDAYA, TEMPAT, DAN LOWONGAN'.localized(context),
    description:
        'Kenali budaya tiap daerah dan lihat lowongan kerja di peta Jepang'
            .localized(context),
    asset: 'assets/images/explore/japan-map.webp',
    color: const Color(0xFF1D2088),
    icon: Icons.map_rounded,
    onTap: () => context.push('/map'),
  );
}
