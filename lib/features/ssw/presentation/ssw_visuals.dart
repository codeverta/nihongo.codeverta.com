import 'package:flutter/material.dart';

class SswArtwork extends StatelessWidget {
  const SswArtwork({super.key});

  @override
  Widget build(BuildContext context) => ClipRRect(
        borderRadius: BorderRadius.circular(24),
        child: AspectRatio(
          aspectRatio: 1.5,
          child: Image.asset(
            'assets/images/ssw/ssw-careers.png',
            fit: BoxFit.cover,
            semanticLabel:
                'Caregiver dan lansia bersama pekerja berbagai bidang di Jepang',
          ),
        ),
      );
}

IconData sswSectorIcon(String id) => switch (id) {
      'nursing-care' => Icons.volunteer_activism_rounded,
      'building-cleaning' => Icons.cleaning_services_rounded,
      'linen-supply' => Icons.local_laundry_service_rounded,
      'industrial-manufacturing' => Icons.precision_manufacturing_rounded,
      'construction' => Icons.construction_rounded,
      'shipbuilding' => Icons.directions_boat_rounded,
      'automobile-maintenance' => Icons.car_repair_rounded,
      'aviation' => Icons.flight_rounded,
      'accommodation' => Icons.hotel_rounded,
      'road-transport' => Icons.local_shipping_rounded,
      'railway' => Icons.train_rounded,
      'logistics-warehouse' => Icons.warehouse_rounded,
      'agriculture' => Icons.agriculture_rounded,
      'fishery' => Icons.water_rounded,
      'food-manufacturing' => Icons.lunch_dining_rounded,
      'food-service' => Icons.restaurant_rounded,
      'forestry' => Icons.forest_rounded,
      'wood-industry' => Icons.carpenter_rounded,
      _ => Icons.recycling_rounded,
    };

class SswStatusBadge extends StatelessWidget {
  const SswStatusBadge({super.key, this.available = false});
  final bool available;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: available
            ? colors.primaryContainer
            : colors.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(30),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        child: Text(
          available ? 'Materi tersedia' : 'Coming Soon',
          style: TextStyle(
            fontSize: 12,
            color:
                available ? colors.onPrimaryContainer : colors.onSurfaceVariant,
            fontWeight: FontWeight.w700,
          ),
        ),
      ),
    );
  }
}
