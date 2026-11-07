import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// Language for spot text: English only when the app runs in English.
String mapLanguage(BuildContext context) =>
    Localizations.localeOf(context).languageCode == 'en' ? 'en' : 'id';

String categoryLabel(BuildContext context, String category) =>
    switch (category) {
      'culture' => 'Budaya'.localized(context),
      'shrine' => 'Kuil & Shrine'.localized(context),
      'nature' => 'Alam'.localized(context),
      'food' => 'Kuliner'.localized(context),
      'city' => 'Kota & modern'.localized(context),
      'history' => 'Sejarah'.localized(context),
      _ => category,
    };

IconData categoryIcon(String category) => switch (category) {
  'culture' => Icons.festival_rounded,
  'shrine' => Icons.temple_buddhist_rounded,
  'nature' => Icons.landscape_rounded,
  'food' => Icons.ramen_dining_rounded,
  'city' => Icons.location_city_rounded,
  'history' => Icons.castle_rounded,
  _ => Icons.place_rounded,
};

Color categoryColor(String category) => switch (category) {
  'culture' => const Color(0xFFC2410C),
  'shrine' => const Color(0xFFD7263D),
  'nature' => const Color(0xFF15803D),
  'food' => const Color(0xFFB45309),
  'city' => const Color(0xFF2563EB),
  'history' => const Color(0xFF7C3AED),
  _ => const Color(0xFF495057),
};

const jobPinColor = Color(0xFF2A2F6E);
