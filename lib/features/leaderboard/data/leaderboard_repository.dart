import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Leaderboard periods served by `GET /v1/leaderboard?period=`.
enum LeaderboardPeriod {
  week('week'),
  all('all');

  const LeaderboardPeriod(this.apiValue);
  final String apiValue;
}

class LeaderboardEntry {
  const LeaderboardEntry({
    required this.rank,
    required this.displayName,
    required this.tag,
    required this.xp,
    this.photoUrl,
  });

  factory LeaderboardEntry.fromJson(Map<String, dynamic> json) =>
      LeaderboardEntry(
        rank: json['rank'] as int? ?? 0,
        displayName: json['displayName'] as String? ?? '',
        tag: json['tag'] as String? ?? '',
        xp: json['xp'] as int? ?? 0,
        photoUrl: json['photoUrl'] as String?,
      );

  final int rank;

  /// Empty when the name is missing or looks like an email; show the tag.
  final String displayName;
  final String tag;
  final int xp;
  final String? photoUrl;

  Map<String, dynamic> toJson() => {
    'rank': rank,
    'displayName': displayName,
    'tag': tag,
    'xp': xp,
    if (photoUrl != null) 'photoUrl': photoUrl,
  };
}

class LeaderboardPage {
  const LeaderboardPage({
    required this.period,
    required this.weekEndsAt,
    required this.participants,
    required this.entries,
  });

  factory LeaderboardPage.fromJson(Map<String, dynamic> json) =>
      LeaderboardPage(
        period: json['period'] as String? ?? 'week',
        weekEndsAt:
            DateTime.tryParse(json['weekEndsAt'] as String? ?? '')?.toUtc() ??
            DateTime.now().toUtc(),
        participants: json['participants'] as int? ?? 0,
        entries: (json['entries'] as List<dynamic>? ?? const [])
            .map((e) => LeaderboardEntry.fromJson(e as Map<String, dynamic>))
            .toList(growable: false),
      );

  final String period;
  final DateTime weekEndsAt;
  final int participants;
  final List<LeaderboardEntry> entries;

  Map<String, dynamic> toJson() => {
    'period': period,
    'weekEndsAt': weekEndsAt.toIso8601String(),
    'participants': participants,
    'entries': entries.map((e) => e.toJson()).toList(),
  };
}

class LeaderboardStanding {
  const LeaderboardStanding({
    required this.xp,
    required this.rank,
    required this.participants,
    required this.gapToNext,
  });

  factory LeaderboardStanding.fromJson(Map<String, dynamic>? json) =>
      LeaderboardStanding(
        xp: json?['xp'] as int? ?? 0,
        rank: json?['rank'] as int? ?? 0,
        participants: json?['participants'] as int? ?? 0,
        gapToNext: json?['gapToNext'] as int? ?? 0,
      );

  final int xp;

  /// 0 when the user has no XP in the period.
  final int rank;
  final int participants;

  /// XP needed to pass the person one rank above; 0 at rank 1.
  final int gapToNext;
  bool get ranked => rank > 0;

  Map<String, dynamic> toJson() => {
    'xp': xp,
    'rank': rank,
    'participants': participants,
    'gapToNext': gapToNext,
  };
}

class WeeklyResult {
  const WeeklyResult({
    required this.weekStart,
    required this.rank,
    required this.xp,
  });

  factory WeeklyResult.fromJson(Map<String, dynamic> json) => WeeklyResult(
    weekStart: json['weekStart'] as String? ?? '',
    rank: json['rank'] as int? ?? 0,
    xp: json['xp'] as int? ?? 0,
  );

  final String weekStart;
  final int rank;
  final int xp;

  Map<String, dynamic> toJson() => {
    'weekStart': weekStart,
    'rank': rank,
    'xp': xp,
  };
}

class MyLeaderboard {
  const MyLeaderboard({
    required this.hidden,
    required this.displayName,
    required this.tag,
    required this.weekEndsAt,
    required this.week,
    required this.all,
    this.photoUrl,
    this.lastWeek,
  });

  factory MyLeaderboard.fromJson(Map<String, dynamic> json) => MyLeaderboard(
    hidden: json['hidden'] as bool? ?? false,
    displayName: json['displayName'] as String? ?? '',
    tag: json['tag'] as String? ?? '',
    photoUrl: json['photoUrl'] as String?,
    weekEndsAt:
        DateTime.tryParse(json['weekEndsAt'] as String? ?? '')?.toUtc() ??
        DateTime.now().toUtc(),
    week: LeaderboardStanding.fromJson(json['week'] as Map<String, dynamic>?),
    all: LeaderboardStanding.fromJson(json['all'] as Map<String, dynamic>?),
    lastWeek: json['lastWeek'] is Map<String, dynamic>
        ? WeeklyResult.fromJson(json['lastWeek'] as Map<String, dynamic>)
        : null,
  );

  final bool hidden;
  final String displayName;
  final String tag;
  final String? photoUrl;
  final DateTime weekEndsAt;
  final LeaderboardStanding week;
  final LeaderboardStanding all;
  final WeeklyResult? lastWeek;

  LeaderboardStanding standing(LeaderboardPeriod period) =>
      period == LeaderboardPeriod.week ? week : all;

  Map<String, dynamic> toJson() => {
    'hidden': hidden,
    'displayName': displayName,
    'tag': tag,
    if (photoUrl != null) 'photoUrl': photoUrl,
    'weekEndsAt': weekEndsAt.toIso8601String(),
    'week': week.toJson(),
    'all': all.toJson(),
    'lastWeek': lastWeek?.toJson(),
  };
}

/// A response plus when it was fetched. [fromCache] is true when the network
/// failed and the last saved copy is shown instead.
class Cached<T> {
  const Cached(this.data, this.fetchedAt, {this.fromCache = false});
  final T data;
  final DateTime fetchedAt;
  final bool fromCache;
}

/// Thrown when there is no connection and nothing was cached yet.
class LeaderboardOffline implements Exception {
  const LeaderboardOffline();
  @override
  String toString() => 'Butuh koneksi internet';
}

class LeaderboardRepository {
  LeaderboardRepository(this.dio, this.prefs);
  final Dio dio;
  final SharedPreferences prefs;

  static String pageCacheKey(LeaderboardPeriod period) =>
      'leaderboard_cache_${period.apiValue}';
  static String meCacheKey(String userId) => 'leaderboard_me_cache_$userId';

  Future<Cached<LeaderboardPage>> fetchPage(LeaderboardPeriod period) =>
      _cached(
        pageCacheKey(period),
        () => dio.get<Map<String, dynamic>>(
          '/leaderboard',
          queryParameters: {'period': period.apiValue},
        ),
        LeaderboardPage.fromJson,
        (page) => page.toJson(),
      );

  Future<Cached<MyLeaderboard>> fetchMine(String userId) => _cached(
    meCacheKey(userId),
    () => dio.get<Map<String, dynamic>>('/me/leaderboard'),
    MyLeaderboard.fromJson,
    (me) => me.toJson(),
  );

  Future<void> setHidden({required bool hidden}) async {
    await dio.put<Map<String, dynamic>>(
      '/me/leaderboard/settings',
      data: {'hidden': hidden},
    );
  }

  Future<Cached<T>> _cached<T>(
    String key,
    Future<Response<Map<String, dynamic>>> Function() request,
    T Function(Map<String, dynamic>) parse,
    Map<String, dynamic> Function(T) serialize,
  ) async {
    try {
      final response = await request();
      final data = parse(response.data!['data'] as Map<String, dynamic>);
      final now = DateTime.now();
      await prefs.setString(
        key,
        jsonEncode({'at': now.toIso8601String(), 'data': serialize(data)}),
      );
      return Cached(data, now);
    } on DioException catch (error) {
      // Only fall back to the cache when the server could not be reached.
      if (error.response != null) rethrow;
      final raw = prefs.getString(key);
      if (raw == null) throw const LeaderboardOffline();
      try {
        final saved = jsonDecode(raw) as Map<String, dynamic>;
        return Cached(
          parse(saved['data'] as Map<String, dynamic>),
          DateTime.parse(saved['at'] as String),
          fromCache: true,
        );
      } on Object {
        throw const LeaderboardOffline();
      }
    }
  }
}

final leaderboardRepositoryProvider = Provider(
  (ref) => LeaderboardRepository(
    ref.watch(dioProvider),
    ref.watch(sharedPreferencesProvider),
  ),
);

final leaderboardPageProvider = FutureProvider.autoDispose
    .family<Cached<LeaderboardPage>, LeaderboardPeriod>(
      (ref, period) =>
          ref.watch(leaderboardRepositoryProvider).fetchPage(period),
    );

/// The viewer's own standing, or null for guests (who are never ranked).
final myLeaderboardProvider =
    FutureProvider.autoDispose<Cached<MyLeaderboard>?>((ref) {
      final userId = ref.watch(
        authControllerProvider.select((s) => s.user?.id),
      );
      if (userId == null) return null;
      return ref.watch(leaderboardRepositoryProvider).fetchMine(userId);
    });

const _dismissedLastWeekKey = 'leaderboard_last_week_dismissed';

/// The week (Monday date) whose "Minggu lalu kamu finis #N" card was closed.
class DismissedLastWeekNotifier extends StateNotifier<String> {
  DismissedLastWeekNotifier(this.prefs)
    : super(prefs.getString(_dismissedLastWeekKey) ?? '');
  final SharedPreferences prefs;

  Future<void> dismiss(String weekStart) async {
    state = weekStart;
    await prefs.setString(_dismissedLastWeekKey, weekStart);
  }
}

final dismissedLastWeekProvider =
    StateNotifierProvider<DismissedLastWeekNotifier, String>(
      (ref) => DismissedLastWeekNotifier(ref.watch(sharedPreferencesProvider)),
    );
