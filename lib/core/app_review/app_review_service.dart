import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:in_app_review/in_app_review.dart';
import 'package:nihongo_master/core/app_review/app_review_policy.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// App Store ID (angka setelah `id` pada URL App Store). Isi setelah aplikasi
/// terbit di App Store; selama kosong, tombol "Beri rating" disembunyikan di
/// iOS. Pop-up otomatis iOS tetap berjalan tanpa ID ini.
const appStoreId = '';

/// Pembungkus tipis package `in_app_review` agar mudah diganti saat tes.
abstract class StoreReviewClient {
  Future<bool> isAvailable();
  Future<void> requestReview();
  Future<void> openStoreListing({String? appStoreId});
}

class InAppReviewClient implements StoreReviewClient {
  const InAppReviewClient();
  @override
  Future<bool> isAvailable() => InAppReview.instance.isAvailable();
  @override
  Future<void> requestReview() => InAppReview.instance.requestReview();
  @override
  Future<void> openStoreListing({String? appStoreId}) =>
      InAppReview.instance.openStoreListing(appStoreId: appStoreId);
}

/// Platform tempat fitur rating store aktif (hanya Android dan iOS).
bool storeReviewSupported({TargetPlatform? platform, bool isWeb = kIsWeb}) {
  final target = platform ?? defaultTargetPlatform;
  return !isWeb &&
      (target == TargetPlatform.android || target == TargetPlatform.iOS);
}

/// Apakah tombol "Beri rating" bisa membuka halaman aplikasi di store.
bool storeListingAvailable({
  TargetPlatform? platform,
  bool isWeb = kIsWeb,
  String iosAppStoreId = appStoreId,
}) {
  final target = platform ?? defaultTargetPlatform;
  if (!storeReviewSupported(platform: target, isWeb: isWeb)) return false;
  return target == TargetPlatform.android || iosAppStoreId.isNotEmpty;
}

final storeReviewClientProvider = Provider<StoreReviewClient>(
  (ref) => const InAppReviewClient(),
);

final appReviewServiceProvider = Provider<AppReviewService>(
  (ref) => AppReviewService(
    ref.watch(sharedPreferencesProvider),
    ref.watch(storeReviewClientProvider),
  ),
);

/// Mencatat sinyal pemakaian secara lokal dan meminta pop-up rating bila
/// [AppReviewPolicy] terpenuhi. Semua kegagalan ditelan: fitur rating tidak
/// boleh mengganggu alur belajar.
class AppReviewService {
  AppReviewService(
    this._preferences,
    this._client, {
    this.policy = const AppReviewPolicy(),
    DateTime Function()? clock,
    bool? supported,
  }) : _clock = clock ?? DateTime.now,
       _supported = supported ?? storeReviewSupported();

  static const lessonsKey = 'app_review_completed_lessons_v1';
  static const activeDaysKey = 'app_review_active_days_v1';
  static const lastRequestedKey = 'app_review_last_requested_at_v1';

  final SharedPreferences _preferences;
  final StoreReviewClient _client;
  final AppReviewPolicy policy;
  final DateTime Function() _clock;
  final bool _supported;
  bool _requesting = false;

  int get completedLessons => _list(lessonsKey).length;
  int get activeDays => _list(activeDaysKey).length;
  DateTime? get lastRequestedAt =>
      DateTime.tryParse(_preferences.getString(lastRequestedKey) ?? '');

  /// Dipanggil saat aplikasi dibuka atau pengguna sedang belajar.
  Future<void> recordActiveDay() => _addCapped(
    activeDaysKey,
    AppReviewPolicy.dayKey(_clock()),
    policy.minimumActiveDays,
  );

  /// Dipanggil hanya saat sebuah pelajaran selesai untuk pertama kali.
  Future<void> recordLessonCompleted(String lessonKey) async {
    await recordActiveDay();
    await _addCapped(lessonsKey, lessonKey, policy.minimumLessons);
  }

  bool get isEligible =>
      _supported &&
      policy.isEligible(
        completedLessons: completedLessons,
        activeDays: activeDays,
        lastRequestedAt: lastRequestedAt,
        now: _clock(),
      );

  /// Meminta pop-up rating bawaan bila syarat terpenuhi. Mengembalikan `true`
  /// jika permintaan dikirim ke store (store tetap bisa memilih tidak
  /// menampilkannya).
  Future<bool> maybeRequestReview() async {
    if (_requesting) return false;
    _requesting = true;
    try {
      await recordActiveDay();
      if (!isEligible || !await _client.isAvailable()) return false;
      // Simpan dulu agar kegagalan di tengah jalan tidak memicu ulang terus.
      await _preferences.setString(
        lastRequestedKey,
        _clock().toUtc().toIso8601String(),
      );
      await _client.requestReview();
      return true;
    } on Object {
      return false;
    } finally {
      _requesting = false;
    }
  }

  List<String> _list(String key) {
    try {
      return _preferences.getStringList(key) ?? const [];
    } on Object {
      return const [];
    }
  }

  Future<void> _addCapped(String key, String value, int cap) async {
    final items = _list(key);
    if (items.length >= cap || items.contains(value)) return;
    try {
      await _preferences.setStringList(key, [...items, value]);
    } on Object {
      // Penyimpanan gagal: cukup lewati, coba lagi di kesempatan berikutnya.
    }
  }
}

/// Dipanggil layar daftar pelajaran setelah pengguna kembali dari pelajaran.
void requestAppReviewIfEligible(WidgetRef ref) {
  try {
    unawaited(ref.read(appReviewServiceProvider).maybeRequestReview());
  } on Object {
    // Provider belum siap (misalnya di tes): rating tidak boleh mengganggu.
  }
}
