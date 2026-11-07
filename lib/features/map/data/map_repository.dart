import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/map/domain/map_models.dart';
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

  Future<MapData> spots() async {
    final cached = preferences.getString(spotsKey);
    final etag = preferences.getString(spotsEtagKey);
    try {
      final response = await _getSpots(cached != null && etag != null ? etag : null);
      if (response.statusCode == 304 && cached != null) {
        return MapData(spots: _parseSpots(cached));
      }
      final body = '${response.data}';
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
