import 'dart:async';
import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

class LearningActivity {
  const LearningActivity({
    required this.kind,
    required this.lessonId,
    required this.title,
    required this.level,
    required this.position,
    required this.total,
    required this.accessedAt,
    this.completed = false,
    this.available = true,
    this.pending = false,
  });
  factory LearningActivity.fromJson(Map<String, dynamic> j) => LearningActivity(
    kind: j['kind'] as String,
    lessonId: j['lessonId'] as String,
    title: j['title'] as String,
    level: j['level'] as String,
    position: j['position'] as int,
    total: j['total'] as int,
    accessedAt: DateTime.parse(j['accessedAt'] as String),
    completed: j['completed'] == true,
    available: j['available'] != false,
    pending: j['pending'] == true,
  );
  final String kind;
  final String lessonId;
  final String title;
  final String level;
  final int position;
  final int total;
  final DateTime accessedAt;
  final bool completed;
  final bool available;
  final bool pending;
  String get key => '$kind:$lessonId';
  String get category => switch (kind) {
    'vocabulary' => 'Kosakata',
    'grammar' => 'Grammar',
    'kanji' => 'Kanji',
    _ => 'Reading',
  };
  String get unit => switch (kind) {
    'vocabulary' => 'Kata',
    'grammar' => 'Pola',
    'kanji' => 'Karakter',
    _ => 'Halaman',
  };
  String get coursePath =>
      '/learn/$kind?level=${Uri.encodeQueryComponent(level)}';
  String get lessonPath =>
      '/learn/$kind/lessons/${Uri.encodeComponent(lessonId)}';
  bool isLocked({required bool premium}) =>
      !premium && const ['N1', 'N2', 'N3'].contains(level);
  LearningActivity copy({
    int? position,
    bool? completed,
    bool? pending,
    DateTime? accessedAt,
  }) => LearningActivity(
    kind: kind,
    lessonId: lessonId,
    title: title,
    level: level,
    position: position ?? this.position,
    total: total,
    accessedAt: accessedAt ?? this.accessedAt,
    completed: completed ?? this.completed,
    available: available,
    pending: pending ?? this.pending,
  );
  Map<String, dynamic> toJson() => {
    'kind': kind,
    'lessonId': lessonId,
    'title': title,
    'level': level,
    'position': position,
    'total': total,
    'accessedAt': accessedAt.toUtc().toIso8601String(),
    'completed': completed,
    'available': available,
    'pending': pending,
  };
}

class LearningHistoryRepository {
  LearningHistoryRepository(this.dio);
  final Dio dio;
  Future<List<LearningActivity>> fetch() async {
    final response = await dio.get<Map<String, dynamic>>(
      '/me/learning-history',
      options: Options(receiveTimeout: const Duration(seconds: 3)),
    );
    final data = response.data!['data'] as Map<String, dynamic>;
    return (data['items'] as List<dynamic>)
        .map((e) => LearningActivity.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<LearningActivity> save(LearningActivity item) async {
    final response = await dio.put<Map<String, dynamic>>(
      '/me/learning-history/${Uri.encodeComponent(item.kind)}/${Uri.encodeComponent(item.lessonId)}',
      data: {
        'position': item.position,
        'accessedAt': item.accessedAt.toUtc().toIso8601String(),
      },
      options: Options(receiveTimeout: const Duration(seconds: 3)),
    );
    return LearningActivity.fromJson(
      response.data!['data'] as Map<String, dynamic>,
    );
  }
}

final learningHistoryRepositoryProvider = Provider(
  (ref) => LearningHistoryRepository(ref.watch(dioProvider)),
);
final learningHistoryProvider =
    StateNotifierProvider<LearningHistoryController, List<LearningActivity>>((
      ref,
    ) {
      final id = ref.watch(authControllerProvider.select((s) => s.user?.id));
      final controller = LearningHistoryController(
        ref.watch(sharedPreferencesProvider),
        ref.watch(learningHistoryRepositoryProvider),
        id,
      );
      if (id != null) unawaited(Future<void>.microtask(controller.refresh));
      return controller;
    });

class LearningHistoryController extends StateNotifier<List<LearningActivity>> {
  LearningHistoryController(this.preferences, this.repository, this.userId)
    : super([]) {
    if (userId == null) return;
    try {
      final raw = preferences.getString(storageKey);
      if (raw != null) {
        state = (jsonDecode(raw) as List<dynamic>)
            .map((e) => LearningActivity.fromJson(e as Map<String, dynamic>))
            .toList();
      }
    } on Object {
      state = [];
    }
  }
  final SharedPreferences preferences;
  final LearningHistoryRepository repository;
  final String? userId;
  String get storageKey => 'learning_history_v1_$userId';
  Timer? _timer;
  Future<void>? _refresh;
  int _revision = 0;
  DateTime? _lastWrite;
  Future<void> _writes = Future.value();

  DateTime _timestamp() {
    var now = DateTime.now().toUtc();
    if (_lastWrite != null && !now.isAfter(_lastWrite!)) {
      now = _lastWrite!.add(const Duration(microseconds: 1));
    }
    return _lastWrite = now;
  }

  void _store(List<LearningActivity> items) {
    if (!mounted) return;
    items.sort((a, b) => b.accessedAt.compareTo(a.accessedAt));
    state = List.unmodifiable(items);
    final snapshot = jsonEncode(items.map((e) => e.toJson()).toList());
    _writes = _writes
        .then((_) async {
          await preferences.setString(storageKey, snapshot);
        })
        .catchError((Object _) {});
  }

  void _replace(LearningActivity item) =>
      _store([item, ...state.where((e) => e.key != item.key)]);

  // Called only after a lesson successfully loads. Sync errors cannot block learning.
  Future<int> open({
    required String kind,
    required String lessonId,
    required String title,
    required int total,
    required bool completed,
    String? level,
    bool Function()? isActive,
  }) async {
    if (!mounted || userId == null) return 0;
    await refresh();
    if (!mounted || (isActive != null && !isActive())) return 0;
    final old = state.where((e) => e.key == '$kind:$lessonId').firstOrNull;
    final position = completed ? 0 : (old?.position ?? 0).clamp(0, total - 1);
    _replace(
      LearningActivity(
        kind: kind,
        lessonId: lessonId,
        title: title,
        level:
            level ??
            RegExp(
              'n[1-5]',
              caseSensitive: false,
            ).firstMatch(lessonId)?.group(0)?.toUpperCase() ??
            'N5',
        position: position,
        total: total,
        completed: completed,
        accessedAt: _timestamp(),
        pending: true,
      ),
    );
    _schedule();
    return position;
  }

  void position(String kind, String id, int position) {
    if (!mounted) return;
    final old = state.where((e) => e.key == '$kind:$id').firstOrNull;
    if (old == null || position < 0 || position >= old.total) return;
    _replace(
      old.copy(position: position, accessedAt: _timestamp(), pending: true),
    );
    _schedule();
  }

  void complete(String kind, String id) {
    if (!mounted) return;
    final old = state.where((e) => e.key == '$kind:$id').firstOrNull;
    if (old == null) return;
    _replace(
      old.copy(completed: true, accessedAt: _timestamp(), pending: true),
    );
    _schedule();
  }

  void _schedule() {
    _revision++;
    _timer?.cancel();
    _timer = Timer(
      const Duration(milliseconds: 600),
      () => unawaited(refresh()),
    );
  }

  Future<void> refresh() {
    if (!mounted || userId == null) return Future.value();
    if (_refresh != null) return _refresh!;
    final revision = _revision;
    return _refresh = _sync().whenComplete(() {
      _refresh = null;
      if (mounted && revision != _revision && state.any((e) => e.pending)) {
        _schedule();
      }
    });
  }

  Future<void> _sync() async {
    try {
      final remote = await repository.fetch().timeout(
        const Duration(seconds: 3),
      );
      if (!mounted) return;
      final merged = {for (final item in remote) item.key: item};
      for (final local in state) {
        final server = merged[local.key];
        if (local.pending &&
            (server == null || local.accessedAt.isAfter(server.accessedAt))) {
          merged[local.key] = server == null
              ? local
              : server.copy(
                  position: local.position.clamp(0, server.total - 1),
                  accessedAt: local.accessedAt,
                  pending: true,
                );
        }
      }
      _store(merged.values.toList());
      for (final item in state.where((e) => e.pending).toList()) {
        if (!mounted) return;
        try {
          final saved = await repository
              .save(item)
              .timeout(const Duration(seconds: 3));
          if (!mounted) return;
          final current = state.where((e) => e.key == item.key).firstOrNull;
          if (current?.accessedAt == item.accessedAt) _replace(saved);
        } on DioException catch (e) {
          // A removed or newly locked lesson must not be retried forever.
          if (mounted &&
              [403, 404].contains(e.response?.statusCode) &&
              state.where((e) => e.key == item.key).firstOrNull?.accessedAt ==
                  item.accessedAt) {
            _replace(
              LearningActivity(
                kind: item.kind,
                lessonId: item.lessonId,
                title: item.title,
                level: item.level,
                position: item.position,
                total: item.total,
                accessedAt: item.accessedAt,
                completed: item.completed,
                available: e.response?.statusCode != 404,
              ),
            );
          }
        }
      }
    } on Object {
      /* Keep cached history and pending writes for the next retry. */
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }
}
