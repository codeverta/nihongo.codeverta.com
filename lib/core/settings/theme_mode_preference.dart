import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const themeModePreferenceKey = 'display_theme_mode';

/// Mode tampilan aplikasi: ikuti sistem, terang, atau gelap.
final themeModeProvider = StateNotifierProvider<ThemeModeController, ThemeMode>(
  (ref) {
    SharedPreferences? preferences;
    try {
      preferences = ref.watch(sharedPreferencesProvider);
      // ignore: avoid_catching_errors
    } on StateError {
      // Isolated previews/tests run without persistent storage.
    }
    return ThemeModeController(preferences);
  },
);

class ThemeModeController extends StateNotifier<ThemeMode> {
  ThemeModeController(this._preferences) : super(_load(_preferences));

  final SharedPreferences? _preferences;

  static ThemeMode _load(SharedPreferences? preferences) {
    return switch (preferences?.getString(themeModePreferenceKey)) {
      'light' => ThemeMode.light,
      'dark' => ThemeMode.dark,
      _ => ThemeMode.system,
    };
  }

  void select(ThemeMode mode) {
    if (state == mode) return;
    state = mode;
    final preferences = _preferences;
    if (preferences == null) return;
    unawaited(preferences.setString(themeModePreferenceKey, mode.name));
  }
}
