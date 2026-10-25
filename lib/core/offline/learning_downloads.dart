import 'dart:async';

import 'package:dio/dio.dart';

import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_cache_interceptor.dart';

typedef CanStartLearningDownload = Future<bool> Function();

class LearningDownloadProgress {
  const LearningDownloadProgress({
    required this.courseKey,
    required this.total,
    required this.completed,
    required this.failed,
    required this.active,
    required this.isPreparing,
  });

  final String courseKey;
  final int total;
  final int completed;
  final int failed;
  final List<String> active;
  final bool isPreparing;

  int get remaining => total - completed - failed;
  double get fraction => total == 0 ? 1 : (completed + failed) / total;
}

class _DownloadTask {
  const _DownloadTask(this.path);

  final String path;
}

/// Preloads lesson text and its signed offline-practice package when a course
/// opens. The cache interceptor enforces account scope and premium access.
class LearningDownloads {
  LearningDownloads({
    required this.dio,
    required FutureOr<PersistentLearningCache> cache,
    required this.currentUserId,
    required this.canReadPremium,
    this.canStartDownload,
    this.maxConcurrent = 2,
    this.onProgress,
  }) : assert(maxConcurrent > 0, 'maxConcurrent must be greater than zero'),
       _cacheFuture = Future<PersistentLearningCache>.value(cache),
       _state = const LearningDownloadProgress(
         courseKey: '',
         total: 0,
         completed: 0,
         failed: 0,
         active: <String>[],
         isPreparing: false,
       );

  final Dio dio;
  final Future<PersistentLearningCache> _cacheFuture;
  final CurrentLearningUserId currentUserId;
  final CanReadPremiumCache canReadPremium;
  final CanStartLearningDownload? canStartDownload;
  final int maxConcurrent;
  final void Function(LearningDownloadProgress progress)? onProgress;

  final StreamController<LearningDownloadProgress> _progress =
      StreamController<LearningDownloadProgress>.broadcast();
  final Set<String> _active = <String>{};
  late LearningDownloadProgress _state;
  CancelToken? _cancelToken;
  int _generation = 0;
  bool _disposed = false;

  Stream<LearningDownloadProgress> get progress => _progress.stream;
  LearningDownloadProgress get state => _state;
  Future<PersistentLearningCache> get initializedCache => _cacheFuture;

  Future<LearningDownloadProgress> preloadCourse({
    required String type,
    required String level,
    required List<String> lessonIds,
    String? activeLessonId,
  }) async {
    if (_disposed) return _state;
    if (!_supportedType(type)) throw ArgumentError.value(type, 'type');
    await _cacheFuture;
    cancel();
    final generation = ++_generation;
    final courseKey = '$type:$level';
    _active.clear();

    final mayStart = canStartDownload == null || await canStartDownload!();
    if (!mayStart || generation != _generation || _disposed) {
      _publish(
        LearningDownloadProgress(
          courseKey: courseKey,
          total: 0,
          completed: 0,
          failed: 0,
          active: const <String>[],
          isPreparing: false,
        ),
      );
      return _state;
    }

    final userId = _normalizeUserId(await currentUserId());
    if (generation != _generation || _disposed) return _state;
    if (userId == null || userId.isEmpty) {
      return _runTasks(courseKey, const <_DownloadTask>[], userId, generation);
    }
    final uniqueIds = lessonIds.toSet().toList();
    if (activeLessonId != null && uniqueIds.remove(activeLessonId)) {
      uniqueIds.insert(0, activeLessonId);
    }

    final tasks = <_DownloadTask>[];
    for (final id in uniqueIds) {
      if (_isPremium(level, id) && !await canReadPremium(userId)) {
        continue;
      }
      final encodedId = Uri.encodeComponent(id);
      tasks
        ..add(_DownloadTask('/$type/lessons/$encodedId'))
        ..add(_DownloadTask('/offline/practice/$type/$encodedId'));
    }

    return _runTasks(courseKey, tasks, userId, generation);
  }

  /// Preloads the account's deck detail and current study session on deck open.
  Future<LearningDownloadProgress> preloadDeck({
    required String deckId,
    String? level,
    String? requiredTier,
  }) async {
    if (_disposed) return _state;
    await _cacheFuture;
    cancel();
    final generation = ++_generation;
    final courseKey = 'anki:$deckId';
    final mayStart = canStartDownload == null || await canStartDownload!();
    if (!mayStart || generation != _generation || _disposed) {
      _publish(
        LearningDownloadProgress(
          courseKey: courseKey,
          total: 0,
          completed: 0,
          failed: 0,
          active: const <String>[],
          isPreparing: false,
        ),
      );
      return _state;
    }

    final userId = _normalizeUserId(await currentUserId());
    if (generation != _generation || _disposed) return _state;
    final premium = _isPremiumDeck(level: level, requiredTier: requiredTier);
    final authorized =
        !premium ||
        (userId != null && userId.isNotEmpty && await canReadPremium(userId));
    final tasks = <_DownloadTask>[];
    if (authorized) {
      final encodedId = Uri.encodeComponent(deckId);
      tasks.add(_DownloadTask('/anki/decks/$encodedId'));
      if (userId != null && userId.isNotEmpty) {
        tasks.add(_DownloadTask('/anki/decks/$encodedId/study'));
      }
    }
    return _runTasks(courseKey, tasks, userId, generation);
  }

  /// Preloads a signed ordinary-practice package for an SSW sector.
  /// Simulations and their answer submissions use different routes and are
  /// never part of this offline learning package.
  Future<LearningDownloadProgress> preloadSswSector({
    required String sectorId,
  }) async {
    if (_disposed) return _state;
    await _cacheFuture;
    cancel();
    final generation = ++_generation;
    final courseKey = 'ssw:$sectorId';
    final mayStart = canStartDownload == null || await canStartDownload!();
    if (!mayStart || generation != _generation || _disposed) {
      _publish(
        LearningDownloadProgress(
          courseKey: courseKey,
          total: 0,
          completed: 0,
          failed: 0,
          active: const <String>[],
          isPreparing: false,
        ),
      );
      return _state;
    }

    final userId = _normalizeUserId(await currentUserId());
    if (generation != _generation || _disposed) return _state;
    final authorized =
        userId != null && userId.isNotEmpty && await canReadPremium(userId);
    final tasks = authorized
        ? <_DownloadTask>[
            _DownloadTask(
              '/offline/practice/ssw/${Uri.encodeComponent(sectorId)}',
            ),
          ]
        : const <_DownloadTask>[];
    return _runTasks(courseKey, tasks, userId, generation);
  }

  Future<LearningDownloadProgress> _runTasks(
    String courseKey,
    List<_DownloadTask> tasks,
    String? userId,
    int generation,
  ) async {
    _cancelToken = CancelToken();
    _publish(
      LearningDownloadProgress(
        courseKey: courseKey,
        total: tasks.length,
        completed: 0,
        failed: 0,
        active: const <String>[],
        isPreparing: tasks.isNotEmpty,
      ),
    );
    if (tasks.isEmpty) return _state;

    var nextIndex = 0;
    var completed = 0;
    var failed = 0;
    var stopQueue = false;
    Future<void> worker() async {
      while (true) {
        if (_disposed ||
            generation != _generation ||
            stopQueue ||
            (_cancelToken?.isCancelled ?? false)) {
          return;
        }
        if (_accountId(await currentUserId()) != _accountId(userId)) {
          stopQueue = true;
          _cancelToken?.cancel('Learning account changed.');
          return;
        }
        if (canStartDownload != null) {
          var allowed = false;
          try {
            allowed = await canStartDownload!();
          } on Object {
            allowed = false;
          }
          if (!allowed) {
            stopQueue = true;
            _cancelToken?.cancel('Download connectivity policy changed.');
            return;
          }
          if (_accountId(await currentUserId()) != _accountId(userId)) {
            stopQueue = true;
            _cancelToken?.cancel('Learning account changed.');
            return;
          }
        }
        if (nextIndex >= tasks.length) return;
        final task = tasks[nextIndex++];
        _active.add(task.path);
        _publish(
          LearningDownloadProgress(
            courseKey: courseKey,
            total: tasks.length,
            completed: completed,
            failed: failed,
            active: List<String>.unmodifiable(_active),
            isPreparing: true,
          ),
        );
        try {
          await dio.get<dynamic>(
            task.path,
            cancelToken: _cancelToken,
            options: Options(
              extra: <String, dynamic>{
                if (userId != null) offlineAccountIdExtraKey: userId,
              },
            ),
          );
          completed++;
        } on DioException catch (error) {
          failed++;
          if (error.response == null) {
            stopQueue = true;
            _cancelToken?.cancel('Learning download lost its connection.');
          }
        } on Object {
          failed++;
          stopQueue = true;
          _cancelToken?.cancel('Learning download failed.');
        } finally {
          _active.remove(task.path);
          if (generation == _generation && !_disposed) {
            _publish(
              LearningDownloadProgress(
                courseKey: courseKey,
                total: tasks.length,
                completed: completed,
                failed: failed,
                active: List<String>.unmodifiable(_active),
                isPreparing:
                    _active.isNotEmpty ||
                    (!stopQueue && nextIndex < tasks.length),
              ),
            );
          }
        }
      }
    }

    await Future.wait<void>(
      List<Future<void>>.generate(
        maxConcurrent.clamp(1, tasks.length),
        (_) => worker(),
      ),
    );
    if (generation == _generation && !_disposed) {
      _publish(
        LearningDownloadProgress(
          courseKey: courseKey,
          total: tasks.length,
          completed: completed,
          failed: failed,
          active: const <String>[],
          isPreparing: false,
        ),
      );
    }
    return _state;
  }

  void cancel() {
    _generation++;
    _cancelToken?.cancel('Learning download cancelled.');
    _cancelToken = null;
    _active.clear();
    if (_state.isPreparing) {
      _publish(
        LearningDownloadProgress(
          courseKey: _state.courseKey,
          total: _state.total,
          completed: _state.completed,
          failed: _state.failed,
          active: const <String>[],
          isPreparing: false,
        ),
      );
    }
  }

  Future<void> dispose() async {
    if (_disposed) return;
    cancel();
    _disposed = true;
    await _progress.close();
  }

  void _publish(LearningDownloadProgress progress) {
    _state = progress;
    if (!_progress.isClosed) _progress.add(progress);
    try {
      onProgress?.call(progress);
    } on Object {
      // UI observers are optional and do not interrupt preloading.
    }
  }
}

bool _supportedType(String type) =>
    const <String>{'vocabulary', 'grammar', 'kanji', 'reading'}.contains(type);

String _accountId(String? id) => id == null || id.isEmpty ? '__guest__' : id;

String? _normalizeUserId(String? id) =>
    id == null || id.trim().isEmpty ? null : id;

bool _isPremium(String level, String lessonId) {
  final fromLevel = RegExp(r'^N[1-3]$', caseSensitive: false).hasMatch(level);
  final match = RegExp(
    r'(?:^|[-_])N[1-3](?:[-_]|$)',
    caseSensitive: false,
  ).hasMatch(lessonId);
  return fromLevel || match;
}

bool _isPremiumDeck({String? level, String? requiredTier}) {
  final metadata = classifyAnkiPremiumMetadata(
    level: level,
    requiredTier: requiredTier,
  );
  if (metadata != null) return metadata;
  // IDs can be non-obvious (for example, `n2m`), so unknown decks fail closed.
  return true;
}
