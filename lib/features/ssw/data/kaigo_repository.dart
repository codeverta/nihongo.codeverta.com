import 'dart:convert';

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:shared_preferences/shared_preferences.dart';

final kaigoCurriculumProvider = FutureProvider<KaigoCurriculum>((ref) async {
  return _loadSswCurriculum(ref, 'kaigo');
});

final sswCurriculumProvider = FutureProvider.family<KaigoCurriculum, String>((
  ref,
  sector,
) async {
  if (sector == 'nursing-care') {
    return ref.watch(kaigoCurriculumProvider.future);
  }
  if (!sswActiveSectorIds.contains(sector)) {
    throw StateError('Materi belum tersedia');
  }
  return _loadSswCurriculum(ref, sector);
});

/// Selects the source curriculum without allowing English users to silently
/// fall back to Indonesian content while the translation is being reviewed.
String sswCurriculumAssetPath({
  required String sector,
  required String languageCode,
}) => languageCode == 'en'
    ? 'assets/content/ssw/en/$sector.json'
    : 'assets/content/ssw/$sector.json';

Future<KaigoCurriculum> _loadSswCurriculum(Ref ref, String sector) async {
  final languageCode = ref.watch(appLanguageProvider).languageCode;
  final assetPath = sswCurriculumAssetPath(
    sector: sector,
    languageCode: languageCode,
  );
  late final String content;
  try {
    content = await rootBundle.loadString(assetPath);
  } catch (error) {
    if (languageCode == 'en') {
      throw SswCurriculumTranslationUnavailable(sector, cause: error);
    }
    rethrow;
  }
  return KaigoCurriculum.fromJson(jsonDecode(content) as Map<String, dynamic>);
}

class SswCurriculumTranslationUnavailable implements Exception {
  const SswCurriculumTranslationUnavailable(this.sector, {this.cause});

  final String sector;
  final Object? cause;

  @override
  String toString() => 'English SSW curriculum is not ready for $sector.';
}

final sswProgressProvider =
    StateNotifierProvider.family<KaigoProgress, Set<String>, String>((
      ref,
      sector,
    ) {
      final userId = ref.watch(
        authControllerProvider.select((s) => s.user?.id),
      );
      return KaigoProgress(
        ref.watch(sharedPreferencesProvider),
        userId,
        namespace: sector == 'nursing-care' ? 'kaigo' : sector,
      );
    });

// Progress is device-local, scoped by account, and gives no gamification XP.
final kaigoProgressProvider = StateNotifierProvider<KaigoProgress, Set<String>>(
  (ref) {
    final userId = ref.watch(authControllerProvider.select((s) => s.user?.id));
    return KaigoProgress(ref.watch(sharedPreferencesProvider), userId);
  },
);

class KaigoProgress extends StateNotifier<Set<String>> {
  KaigoProgress(this.preferences, String? userId, {String namespace = 'kaigo'})
    : key = userId == null ? null : 'ssw_${namespace}_completed_v1_$userId',
      super(
        userId == null
            ? <String>{}
            : (preferences.getStringList(
                        'ssw_${namespace}_completed_v1_$userId',
                      ) ??
                      <String>[])
                  .toSet(),
      );

  final SharedPreferences preferences;
  final String? key;
  Future<void> _pending = Future<void>.value();

  Future<void> complete(String lessonId) {
    final operation = _pending.then((_) async {
      if (!mounted) return;
      final updated = {...state, lessonId};
      final storageKey = key;
      if (storageKey != null) {
        final saved = await preferences.setStringList(
          storageKey,
          updated.toList(),
        );
        if (!saved) throw StateError('Progress belum tersimpan. Coba lagi.');
      }
      if (mounted) state = updated;
    });
    _pending = operation.catchError((Object _) {});
    return operation;
  }
}
