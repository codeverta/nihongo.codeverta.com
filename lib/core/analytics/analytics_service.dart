import 'dart:async';

import 'package:firebase_analytics/firebase_analytics.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Maps a route pattern (for example `/learn/kanji/lessons/:lessonId`) to a
/// stable, low-cardinality feature name for Google Analytics reports.
String analyticsFeatureForPath(String path) {
  const prefixes = <String, String>{
    '/learn/vocabulary': 'vocabulary',
    '/learn/kanji': 'kanji',
    '/learn/grammar': 'grammar',
    '/learn/reading': 'reading',
    '/anki': 'anki',
    '/dictionary': 'dictionary',
    '/news': 'news',
    '/manga': 'manga',
    '/jobs': 'jobs',
    '/ssw': 'ssw',
    '/exams': 'exam',
    '/premium': 'premium',
    '/achievements': 'statistics',
    '/settings': 'settings',
    '/notifications': 'settings',
    '/attributions': 'settings',
    '/home': 'home',
    '/login': 'login',
    '/onboarding': 'onboarding',
  };
  for (final entry in prefixes.entries) {
    if (path == entry.key || path.startsWith('${entry.key}/')) {
      return entry.value;
    }
  }
  return 'other';
}

/// Thin wrapper around Firebase Analytics. It never throws and does nothing
/// when Firebase is unavailable (web, iOS without config, tests) or when the
/// user turned usage sharing off in Settings.
class AnalyticsService {
  AnalyticsService(this._preferences);

  static const enabledKey = 'analytics_enabled_v1';

  final SharedPreferences _preferences;
  FirebaseAnalytics? _analytics;
  String? _lastScreen;

  bool get enabled => _preferences.getBool(enabledKey) ?? true;
  bool get _ready => _analytics != null && enabled;

  Future<void> initialize() async {
    if (kIsWeb || defaultTargetPlatform != TargetPlatform.android) return;
    try {
      if (Firebase.apps.isEmpty) {
        await Firebase.initializeApp().timeout(const Duration(seconds: 15));
      }
      final analytics = FirebaseAnalytics.instance;
      await analytics.setAnalyticsCollectionEnabled(enabled);
      _analytics = analytics;
    } on Object catch (error) {
      debugPrint('Analytics unavailable: $error');
    }
  }

  Future<void> setEnabled({required bool value}) async {
    await _preferences.setBool(enabledKey, value);
    try {
      await _analytics?.setAnalyticsCollectionEnabled(value);
      if (!value) await _analytics?.resetAnalyticsData();
    } on Object catch (error) {
      debugPrint('Analytics toggle failed: $error');
    }
  }

  /// Logs `screen_view` and `feature_open` once per navigation change.
  void logScreen(String path) {
    if (!_ready || path.isEmpty || path == _lastScreen) return;
    _lastScreen = path;
    final feature = analyticsFeatureForPath(path);
    unawaited(
      _safe(
        () => _analytics!.logScreenView(
          screenName: path,
          screenClass: feature,
        ),
      ),
    );
    logEvent('feature_open', {'feature': feature, 'screen': path});
  }

  void logTab(String tab) {
    logEvent('tab_select', {'tab': tab});
    _lastScreen = null;
  }

  void logEvent(String name, [Map<String, Object>? parameters]) {
    if (!_ready) return;
    unawaited(_safe(() => _analytics!.logEvent(name: name, parameters: parameters)));
  }

  void setUserProperty(String name, String? value) {
    if (!_ready) return;
    unawaited(_safe(() => _analytics!.setUserProperty(name: name, value: value)));
  }

  Future<void> _safe(Future<void> Function() call) async {
    try {
      await call();
    } on Object catch (error) {
      debugPrint('Analytics event failed: $error');
    }
  }
}

final analyticsServiceProvider = Provider<AnalyticsService>(
  (ref) => AnalyticsService(ref.watch(sharedPreferencesProvider)),
);

/// Initializes analytics once and keeps non-identifying user properties
/// (plan and language) current. Watch it from the root widget.
final analyticsBootstrapProvider = Provider<void>((ref) {
  final service = ref.watch(analyticsServiceProvider);
  unawaited(
    service.initialize().then((_) {
      final user = ref.read(authControllerProvider).user;
      service
        ..setUserProperty('plan', _plan(user?.isPremium, user?.isLifetime))
        ..setUserProperty(
          'app_language',
          ref.read(appLanguageProvider).languageCode,
        );
    }),
  );
  ref
    ..listen(
      authControllerProvider.select((s) => s.user),
      (previous, next) => service.setUserProperty(
        'plan',
        _plan(next?.isPremium, next?.isLifetime),
      ),
    )
    ..listen(
      appLanguageProvider,
      (previous, next) =>
          service.setUserProperty('app_language', next.languageCode),
    );
});

String _plan(bool? premium, bool? lifetime) {
  if (lifetime ?? false) return 'lifetime';
  if (premium ?? false) return 'premium';
  return 'free';
}
