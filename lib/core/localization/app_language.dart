import 'dart:async';
import 'dart:ui' show Locale, PlatformDispatcher;

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _manualLanguageKey = 'app_language_manual_v1';
const _initialLanguageKey = 'app_language_initial_v1';
const _languageInitializedKey = 'app_language_initialized_v1';

/// English is available to everyone by default. Build with
/// `--dart-define=ENGLISH_PUBLIC_RELEASE=false` to hide it again; internal QA can
/// still opt in with `ENGLISH_QA_PREVIEW=true` in a debug build.
abstract final class EnglishReleaseGate {
  static const publicRelease = bool.fromEnvironment(
    'ENGLISH_PUBLIC_RELEASE',
    defaultValue: true,
  );
  static const internalQaPreview = bool.fromEnvironment('ENGLISH_QA_PREVIEW');

  static bool get enabled => isEnabled(
    isPublicRelease: publicRelease,
    isInternalQa: internalQaPreview,
    isDebugBuild: kDebugMode,
  );

  static bool isEnabled({
    required bool isPublicRelease,
    required bool isInternalQa,
    required bool isDebugBuild,
  }) => isPublicRelease || (isInternalQa && isDebugBuild);
}

final deviceLocaleProvider = Provider<Locale>(
  (ref) => PlatformDispatcher.instance.locale,
);

final englishAvailabilityProvider = Provider<bool>(
  (ref) => EnglishReleaseGate.enabled,
);

/// The current app locale. Manual choices are stored only in local preferences.
final appLanguageProvider =
    StateNotifierProvider<AppLanguageController, Locale>(
      (ref) => AppLanguageController(
        ref.watch(sharedPreferencesProvider),
        deviceLocale: ref.watch(deviceLocaleProvider),
        englishAvailable: ref.watch(englishAvailabilityProvider),
      ),
    );

/// Keeps regional formatting from the device while the language choice stays
/// an independent, two-value preference.
final appPresentationLocaleProvider = Provider<Locale>((ref) {
  final language = ref.watch(appLanguageProvider).languageCode;
  final region = ref.watch(deviceLocaleProvider).countryCode;
  return Locale(language, region);
});

final appSupportedLocalesProvider = Provider<List<Locale>>((ref) {
  final region = ref.watch(deviceLocaleProvider).countryCode;
  return [
    if (region != null && region.isNotEmpty) Locale('id', region),
    if (region != null && region.isNotEmpty) Locale('en', region),
    ...AppLocalizations.supportedLocales,
  ];
});

class AppLanguageController extends StateNotifier<Locale> {
  AppLanguageController(
    this._preferences, {
    required Locale deviceLocale,
    required bool englishAvailable,
  }) : _englishAvailable = englishAvailable,
       super(
         _resolveInitialLocale(
           _preferences,
           deviceLocale: deviceLocale,
           englishAvailable: englishAvailable,
         ),
       ) {
    if (!_preferences.containsKey(_languageInitializedKey)) {
      unawaited(_persistInitialLocale(state.languageCode));
    }
  }

  final SharedPreferences _preferences;
  final bool _englishAvailable;

  Future<void> selectLanguage(String languageCode) async {
    if (languageCode != 'id' && languageCode != 'en') return;
    if (languageCode == 'en' && !_englishAvailable) return;

    state = Locale(languageCode);
    await _preferences.setString(_manualLanguageKey, languageCode);
    await _preferences.setString(_initialLanguageKey, languageCode);
    await _preferences.setBool(_languageInitializedKey, true);
  }

  static Locale _resolveInitialLocale(
    SharedPreferences preferences, {
    required Locale deviceLocale,
    required bool englishAvailable,
  }) {
    final manualChoice = preferences.getString(_manualLanguageKey);
    if (manualChoice == 'id') return const Locale('id');
    if (manualChoice == 'en' && englishAvailable) return const Locale('en');

    final wasInitialized =
        preferences.getBool(_languageInitializedKey) ?? false;
    final initialChoice = preferences.getString(_initialLanguageKey);
    if (wasInitialized && (initialChoice == 'id' || initialChoice == 'en')) {
      return _localeFor(initialChoice!, englishAvailable: englishAvailable);
    }

    // Existing app state means this install predates the locale setting. Keep
    // those users in Indonesian even when the device itself is set to English.
    final hasExistingAppState = preferences.getKeys().any(
      (key) =>
          key != _manualLanguageKey &&
          key != _initialLanguageKey &&
          key != _languageInitializedKey,
    );
    final detected = hasExistingAppState
        ? 'id'
        : (deviceLocale.languageCode.toLowerCase() == 'id' ? 'id' : 'en');
    return _localeFor(detected, englishAvailable: englishAvailable);
  }

  static Locale _localeFor(
    String languageCode, {
    required bool englishAvailable,
  }) {
    if (languageCode == 'en' && englishAvailable) return const Locale('en');
    return const Locale('id');
  }

  Future<void> _persistInitialLocale(String languageCode) async {
    await _preferences.setString(_initialLanguageKey, languageCode);
    await _preferences.setBool(_languageInitializedKey, true);
  }
}
