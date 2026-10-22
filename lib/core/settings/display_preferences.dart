import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const furiganaPreferenceKey = 'display_furigana_enabled';
const translationPreferenceKey = 'display_translation_enabled';
const legacyNewsFuriganaPreferenceKey = 'news_furigana_enabled';

final displayPreferencesProvider =
    StateNotifierProvider<DisplayPreferencesController, DisplayPreferences>((
      ref,
    ) {
      SharedPreferences? preferences;
      try {
        preferences = ref.watch(sharedPreferencesProvider);
        // ignore: avoid_catching_errors
      } on StateError {
        // Keep isolated previews usable; the app root supplies persistent storage.
      }
      return DisplayPreferencesController(preferences);
    });

class DisplayPreferences {
  const DisplayPreferences({required this.furigana, required this.translation});

  final bool furigana;
  final bool translation;
}

class DisplayPreferencesController extends StateNotifier<DisplayPreferences> {
  DisplayPreferencesController(this._preferences)
    : super(_loadPreferences(_preferences)) {
    if (_preferences != null &&
        !_preferences.containsKey(furiganaPreferenceKey) &&
        _preferences.containsKey(legacyNewsFuriganaPreferenceKey)) {
      _persist(furiganaPreferenceKey, state.furigana);
    }
  }

  final SharedPreferences? _preferences;

  static DisplayPreferences _loadPreferences(SharedPreferences? preferences) {
    return DisplayPreferences(
      furigana:
          preferences?.getBool(furiganaPreferenceKey) ??
          preferences?.getBool(legacyNewsFuriganaPreferenceKey) ??
          true,
      translation: preferences?.getBool(translationPreferenceKey) ?? true,
    );
  }

  void setFurigana({required bool enabled}) {
    if (state.furigana == enabled) return;
    state = DisplayPreferences(
      furigana: enabled,
      translation: state.translation,
    );
    _persist(furiganaPreferenceKey, enabled);
    // Keep the old News key in sync for installs that upgrade incrementally.
    _persist(legacyNewsFuriganaPreferenceKey, enabled);
  }

  void setTranslation({required bool enabled}) {
    if (state.translation == enabled) return;
    state = DisplayPreferences(furigana: state.furigana, translation: enabled);
    _persist(translationPreferenceKey, enabled);
  }

  void _persist(String key, bool value) {
    if (_preferences == null) return;
    unawaited(_writePreference(key, value));
  }

  Future<void> _writePreference(String key, bool value) async {
    await _preferences!.setBool(key, value);
  }
}
