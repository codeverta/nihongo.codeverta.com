import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
import 'package:nihongo_master/features/map/domain/prefecture_living.dart';
import 'package:nihongo_master/features/map/domain/prefecture_profile.dart';
import 'package:shared_preferences/shared_preferences.dart';

final mapRepositoryProvider = Provider(
  (ref) => MapRepository(
    ref.watch(dioProvider),
    ref.watch(sharedPreferencesProvider),
  ),
);

final japanGeometryProvider = FutureProvider<JapanGeometry>((ref) async {
  final raw = await rootBundle.loadString('assets/map/japan_prefectures.json');
  return JapanGeometry.fromJson(jsonDecode(raw) as Map<String, dynamic>);
});

final mapSpotsProvider = FutureProvider.autoDispose<MapData>(
  (ref) => ref.watch(mapRepositoryProvider).spots(),
);

final mapJobCountsProvider = FutureProvider.autoDispose<JobCounts>(
  (ref) => ref.watch(mapRepositoryProvider).jobCounts(),
);

/// Minimum wage, rent and price level per prefecture, or null when neither
/// the server nor the device cache has it (the map then hides wage info).
final mapPrefectureLivingProvider =
    FutureProvider.autoDispose<PrefectureLivingData?>(
      (ref) => ref.watch(mapRepositoryProvider).prefectureLiving(),
    );

/// Wikipedia-style profiles of the 47 prefectures, or null when neither the
/// server nor the device cache has them.
final mapPrefectureProfilesProvider = FutureProvider<PrefectureProfiles?>(
  (ref) => ref.watch(mapRepositoryProvider).prefectureProfiles(),
);

class MapFailure implements Exception {
  const MapFailure(this.message);
  final String message;
  @override
  String toString() => message;
}

/// Loads map spots and job counts. The last successful responses are kept on
/// the device so the map keeps working offline; spots are revalidated with
/// their ETag so unchanged data is not downloaded again.
class MapRepository {
  MapRepository(this.dio, this.preferences);
  final Dio dio;
  final SharedPreferences preferences;

  static const spotsKey = 'map_spots_cache_v1';
  static const spotsEtagKey = 'map_spots_etag_v1';
  static const countsKey = 'map_job_counts_cache_v1';
  static const livingKey = 'map_prefecture_living_cache_v1';
  static const profilesKey = 'map_prefecture_profiles_cache_v1';
  static const profilesEtagKey = 'map_prefecture_profiles_etag_v1';

  /// The profiles rarely change, so they are revalidated with their ETag and
  /// kept on the device for offline reading.
  Future<PrefectureProfiles?> prefectureProfiles() async {
    final cached = preferences.getString(profilesKey);
    final etag = preferences.getString(profilesEtagKey);
    PrefectureProfiles? parse(String body, {bool fromCache = false}) {
      final json = jsonDecode(body) as Map<String, dynamic>;
      return PrefectureProfiles.fromJson(
        json['data'] as Map<String, dynamic>,
        fromCache: fromCache,
      );
    }

    try {
      final response = await dio.get<Object>(
        '/map/prefecture-profiles',
        options: Options(
          headers: {if (cached != null && etag != null) 'If-None-Match': etag},
          validateStatus: (status) => status == 200 || status == 304,
          responseType: ResponseType.plain,
          receiveTimeout: const Duration(seconds: 45),
        ),
      );
      if (response.statusCode == 304 && cached != null) return parse(cached);
      final raw = response.data;
      final body = raw is String ? raw : jsonEncode(raw);
      final profiles = parse(body);
      await preferences.setString(profilesKey, body);
      final newEtag = response.headers.value('etag');
      if (newEtag != null) {
        await preferences.setString(profilesEtagKey, newEtag);
      } else {
        await preferences.remove(profilesEtagKey);
      }
      return profiles;
    } on Object catch (error) {
      debugPrint('Prefecture profiles failed: $error');
      if (cached != null) {
        try {
          return parse(cached, fromCache: true);
        } on Object {
          await preferences.remove(profilesKey);
        }
      }
      return null;
    }
  }

  Future<PrefectureLivingData?> prefectureLiving() async {
    try {
      final response = await dio.get<Object>('/map/prefecture-living');
      final raw = response.data;
      final json = raw is String ? jsonDecode(raw) : raw;
      final data =
          (json as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final parsed = PrefectureLivingData.fromJson(data);
      if (parsed.prefectures.isEmpty) throw const FormatException('empty');
      await preferences.setString(livingKey, jsonEncode(data));
      return parsed;
    } on Object catch (error) {
      debugPrint('Prefecture living data failed: $error');
      final cached = preferences.getString(livingKey);
      if (cached != null) {
        try {
          return PrefectureLivingData.fromJson(
            jsonDecode(cached) as Map<String, dynamic>,
            fromCache: true,
          );
        } on Object {
          await preferences.remove(livingKey);
        }
      }
      return null;
    }
  }

  Future<MapData> spots() async {
    final cached = preferences.getString(spotsKey);
    final etag = preferences.getString(spotsEtagKey);
    try {
      final response = await _getSpots(
        cached != null && etag != null ? etag : null,
      );
      if (response.statusCode == 304 && cached != null) {
        return MapData(spots: _parseSpots(cached));
      }
      // `responseType: plain` should give a String, but an interceptor or Dio
      // version may already have decoded it. Interpolating a decoded Map would
      // produce `{data: ...}`, which is not JSON, so encode it back instead.
      final raw = response.data;
      final body = raw is String ? raw : jsonEncode(raw);
      final spots = _parseSpots(body);
      await preferences.setString(spotsKey, body);
      final newEtag = response.headers.value('etag');
      if (newEtag != null) {
        await preferences.setString(spotsEtagKey, newEtag);
      } else {
        await preferences.remove(spotsEtagKey);
      }
      return MapData(spots: spots);
    } on Object catch (error) {
      debugPrint('Map spots failed: $error');
      if (cached != null) {
        try {
          return MapData(spots: _parseSpots(cached), fromCache: true);
        } on Object {
          await preferences.remove(spotsKey);
        }
      }
      throw const MapFailure(
        'Peta belum dapat dimuat. Periksa koneksi dan coba lagi.',
      );
    }
  }

  /// The spot list is one large document, so it gets a longer read timeout
  /// than normal API calls and one retry for transient failures.
  Future<Response<Object>> _getSpots(String? etag) async {
    Future<Response<Object>> request() => dio.get<Object>(
      '/map/spots',
      options: Options(
        headers: {if (etag != null) 'If-None-Match': etag},
        validateStatus: (status) => status == 200 || status == 304,
        responseType: ResponseType.plain,
        receiveTimeout: const Duration(seconds: 45),
        sendTimeout: const Duration(seconds: 20),
      ),
    );
    try {
      return await request();
    } on DioException catch (error) {
      final status = error.response?.statusCode ?? 0;
      final transient =
          error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout ||
          error.type == DioExceptionType.connectionError ||
          status >= 500;
      if (!transient) rethrow;
      await Future<void>.delayed(const Duration(seconds: 1));
      return request();
    }
  }

  Future<JobCounts> jobCounts() async {
    try {
      final response = await dio.get<Map<String, dynamic>>(
        '/jobs/prefecture-counts',
      );
      final data = response.data!['data'] as Map<String, dynamic>;
      final counts = _parseCounts(data['counts']);
      await preferences.setString(countsKey, jsonEncode(counts));
      return JobCounts(counts: counts);
    } on Object {
      final cached = preferences.getString(countsKey);
      if (cached != null) {
        try {
          return JobCounts(
            counts: _parseCounts(jsonDecode(cached)),
            fromCache: true,
          );
        } on Object {
          await preferences.remove(countsKey);
        }
      }
      return const JobCounts(counts: {}, fromCache: true);
    }
  }

  static List<MapSpot> _parseSpots(String body) {
    final json = jsonDecode(body) as Map<String, dynamic>;
    final data = json['data'] as Map<String, dynamic>;
    return [
      for (final item in data['spots'] as List)
        if (item is Map<String, dynamic>) MapSpot.fromJson(item),
    ].where((spot) => spot.id.isNotEmpty).toList();
  }

  static Map<String, int> _parseCounts(Object? json) {
    if (json is! Map) return {};
    return {
      for (final entry in json.entries)
        if (entry.value is num && (entry.value as num) > 0)
          '${entry.key}': (entry.value as num).toInt(),
    };
  }
}
