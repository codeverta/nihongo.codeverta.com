// Dio's legacy interceptor callbacks have `void` signatures; these overrides
// must await async cache reads and persistence before continuing the handler.
// ignore_for_file: avoid_void_async

import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';

import 'package:nihongo_master/core/offline/learning_cache.dart';

typedef CurrentLearningUserId = FutureOr<String?> Function();
typedef CanReadPremiumCache = FutureOr<bool> Function(String userId);
typedef LearningCacheMetadataCallback =
    void Function(CachedResponseMetadata metadata);

/// Set on a Dio `Options.extra` map when the caller needs an online refresh.
/// Successful responses still update the persistent cache.
const String offlineCacheBypassExtraKey = 'offlineCacheBypass';
const String offlineAccountIdExtraKey = 'offlineAccountId';

const String _bypassExtraKey = '_learningCacheBypass';
const String _accountExtraKey = '_learningCacheAccount';
const String _cacheOwnerExtraKey = '_learningCacheOwner';
const String _cacheKeyExtraKey = '_learningCacheKey';
const String _dedupeCompleterExtraKey = '_learningCacheDedupeCompleter';
const String _mutationPlanExtraKey = '_learningCacheMutationPlan';

/// Intercepts only known learning-content GET routes. Entries are scoped to the
/// account snapshot captured when the request starts and reauthorized on reads.
/// Disk cache work (secure-storage key, encryption, index rewrite) must never
/// keep a lesson on a spinner: past this budget the request continues without
/// the cache and slow writes finish in the background.
const _cacheBudget = Duration(seconds: 5);

/// A duplicate request waits this long for an identical in-flight request
/// before fetching on its own.
const _inFlightBudget = Duration(seconds: 15);

class LearningCacheInterceptor extends Interceptor {
  LearningCacheInterceptor({
    required this.dio,
    required this.cache,
    required this.currentUserId,
    required this.canReadPremium,
    this.onMetadata,
  }) : _cacheFuture = Future<PersistentLearningCache>.value(cache);

  final Dio dio;
  final FutureOr<PersistentLearningCache> cache;
  final Future<PersistentLearningCache> _cacheFuture;
  final CurrentLearningUserId currentUserId;
  final CanReadPremiumCache canReadPremium;
  final LearningCacheMetadataCallback? onMetadata;

  Future<PersistentLearningCache> get initializedCache => _cacheFuture;

  final Map<String, Completer<Response<dynamic>>> _inFlight =
      <String, Completer<Response<dynamic>>>{};
  final Set<String> _refreshing = <String>{};
  static const int _maxConcurrentRefreshes = 2;
  int _activeRefreshes = 0;
  final Queue<Completer<void>> _refreshQueue = Queue<Completer<void>>();

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final expectedAccountExtra = options.extra[offlineAccountIdExtraKey];
    final expectedAccount = expectedAccountExtra is String
        ? _accountId(expectedAccountExtra)
        : null;
    if (expectedAccount != null &&
        !await _matchesCurrentAccount(expectedAccount)) {
      handler.reject(_accountChanged(options));
      return;
    }

    final baseCacheKey = _requestCacheKey(options);
    final language = _acceptLanguage(options.headers);
    final cacheKey = baseCacheKey == null || language == null
        ? baseCacheKey
        : '$baseCacheKey\nlocale=$language';
    if (cacheKey == null) {
      final mutationPlan = _mutationPlan(options);
      if (mutationPlan != null) {
        try {
          final accountId =
              expectedAccount ?? _accountId(await currentUserId());
          if (!await _matchesCurrentAccount(accountId)) {
            handler.reject(_accountChanged(options));
            return;
          }
          options.extra[_accountExtraKey] = accountId;
          options.extra[_mutationPlanExtraKey] = mutationPlan;
        } on Object {
          // Mutation still runs; cache invalidation is a best-effort refresh.
        }
      }
      handler.next(options);
      return;
    }

    if (options.extra[_bypassExtraKey] == true ||
        options.extra[offlineCacheBypassExtraKey] == true) {
      try {
        final accountId = expectedAccount ?? _accountId(await currentUserId());
        if (!await _matchesCurrentAccount(accountId)) {
          handler.reject(_accountChanged(options));
          return;
        }
        options.extra[_accountExtraKey] = accountId;
      } on Object {
        // A forced network read remains usable if account metadata is unavailable.
      }
      options.extra[_cacheKeyExtraKey] = cacheKey;
      handler.next(options);
      return;
    }

    try {
      final accountId = expectedAccount ?? _accountId(await currentUserId());
      if (!await _matchesCurrentAccount(accountId)) {
        handler.reject(_accountChanged(options));
        return;
      }
      options.extra[_accountExtraKey] = accountId;
      options.extra[_cacheKeyExtraKey] = cacheKey;
      final store = await _cacheFuture.timeout(_cacheBudget);
      var ownerId = accountId;
      var cached = await store
          .readResponse(accountId: accountId, key: cacheKey)
          .timeout(_cacheBudget);
      if (cached == null &&
          accountId != guestCacheAccountId &&
          _isPublicLessonKey(cacheKey)) {
        cached = await store
            .readResponse(accountId: guestCacheAccountId, key: cacheKey)
            .timeout(_cacheBudget);
        if (cached != null) ownerId = guestCacheAccountId;
      }
      options.extra[_cacheOwnerExtraKey] = ownerId;
      if (cached != null) {
        final requiresPremium = _isAnkiDeckRequest(options)
            ? !cached.premiumClassificationKnown || cached.requiresPremium
            : cached.requiresPremium || _requiresPremium(options);
        final authorized =
            !requiresPremium ||
            (accountId != guestCacheAccountId &&
                await Future<bool>.value(
                  canReadPremium(accountId),
                ).timeout(_cacheBudget));
        final accountStillMatches = await _matchesCurrentAccount(accountId);
        if (authorized && accountStillMatches) {
          final metadata = CachedResponseMetadata(
            cacheKey: cacheKey,
            cached: true,
            offline: false,
            stale: cached.stale,
            updatedAt: cached.updatedAt,
          );
          _report(metadata);
          handler.resolve(_toResponse(options, cached, metadata));
          if (cached.stale) {
            unawaited(
              _refreshInBackground(
                options,
                accountId,
                cacheKey,
                requiresPremium: requiresPremium,
              ),
            );
          }
          return;
        }
        if (!accountStillMatches) {
          handler.reject(_accountChanged(options));
          return;
        }
        if (!authorized) {
          await store.removeResponse(accountId: ownerId, key: cacheKey);
        }
      }

      if (!await _matchesCurrentAccount(accountId)) {
        handler.reject(_accountChanged(options));
        return;
      }
      final scopedKey = '$accountId\n$cacheKey';
      final pending = _inFlight[scopedKey];
      if (pending != null) {
        try {
          final response = await pending.future.timeout(_inFlightBudget);
          if (!await _matchesCurrentAccount(accountId)) {
            handler.reject(_accountChanged(options));
            return;
          }
          handler.resolve(
            Response<dynamic>(
              requestOptions: options,
              data: response.data,
              statusCode: response.statusCode,
              statusMessage: response.statusMessage,
              headers: response.headers,
              isRedirect: response.isRedirect,
              redirects: response.redirects,
              extra: response.extra,
            ),
          );
        } on TimeoutException {
          // The shared request is stuck; fetch this one independently.
          handler.next(options);
        } on DioException catch (error) {
          handler.reject(
            await _matchesCurrentAccount(accountId)
                ? error.copyWith(requestOptions: options)
                : _accountChanged(options),
          );
        } on Object catch (error) {
          handler.reject(
            await _matchesCurrentAccount(accountId)
                ? DioException(requestOptions: options, error: error)
                : _accountChanged(options),
          );
        }
        return;
      }

      final completer = Completer<Response<dynamic>>();
      _inFlight[scopedKey] = completer;
      unawaited(
        completer.future.then<void>(
          (_) {},
          onError: (Object _, StackTrace __) {},
        ),
      );
      options.extra[_dedupeCompleterExtraKey] = scopedKey;
      handler.next(options);
    } on Object {
      // A cache or authorization storage failure must not stop an online read.
      handler.next(options);
    }
  }

  @override
  void onResponse(
    Response<dynamic> response,
    ResponseInterceptorHandler handler,
  ) async {
    final options = response.requestOptions;
    final cacheKey = options.extra[_cacheKeyExtraKey];
    final accountId = options.extra[_accountExtraKey];
    final dedupeKey = options.extra[_dedupeCompleterExtraKey];
    final mutationPlan = options.extra[_mutationPlanExtraKey];
    Future<void> persist() async {
      if (mutationPlan is String &&
          accountId is String &&
          response.statusCode != null &&
          response.statusCode! >= 200 &&
          response.statusCode! < 300 &&
          _accountId(await currentUserId()) == accountId) {
        final store = await _cacheFuture;
        if (!await _matchesCurrentAccount(accountId)) return;
        await store.invalidateResponsesForAccount(
          accountId: accountId,
          shouldInvalidate: (routeKey) =>
              _invalidateForMutation(routeKey, mutationPlan),
        );
        if (mutationPlan.startsWith('lesson-complete:')) {
          if (!await _matchesCurrentAccount(accountId)) return;
          await store.invalidateResponsesForAccount(
            accountId: guestCacheAccountId,
            shouldInvalidate: (routeKey) =>
                _isPublicLessonKey(routeKey) &&
                _invalidateForMutation(routeKey, mutationPlan),
          );
        }
      }
      if (cacheKey is String &&
          accountId is String &&
          response.statusCode != null &&
          response.statusCode! >= 200 &&
          response.statusCode! < 300) {
        final currentAccount = _accountId(await currentUserId());
        if (currentAccount == accountId) {
          final store = await _cacheFuture;
          // Cache initialization can take long enough for logout/account
          // switching to finish. Recheck immediately before persisting this
          // response under the account captured when the request began.
          if (!await _matchesCurrentAccount(accountId)) return;
          final body = _encodeBody(response.data);
          var stored = false;
          if (body != null && !_isErrorPayload(response.data)) {
            final contentType = _contentType(response.headers, response.data);
            final premiumClassification = _responsePremiumClassification(
              options,
              response.data,
            );
            stored = await store.writeResponse(
              accountId: accountId,
              key: cacheKey,
              body: body,
              contentType: contentType,
              statusCode: response.statusCode!,
              statusMessage: response.statusMessage,
              headers: _safeHeaders(response.headers),
              requiresPremium: premiumClassification.requiresPremium,
              premiumClassificationKnown: premiumClassification.isKnown,
            );
            final publicLesson = _isPublicLessonKey(cacheKey)
                ? _sanitizePublicLessonBody(body)
                : null;
            // The account-scoped write may itself wait on disk. Logout can
            // clear that account while it waits; do not repopulate the shared
            // sanitized copy from the response that crossed the logout.
            if (publicLesson != null && accountId != guestCacheAccountId) {
              if (!await _matchesCurrentAccount(accountId)) return;
              await store.writeResponse(
                accountId: guestCacheAccountId,
                key: cacheKey,
                body: publicLesson,
                contentType: 'application/json',
                statusCode: response.statusCode!,
                statusMessage: response.statusMessage,
                headers: _safeHeaders(response.headers),
                requiresPremium: premiumClassification.requiresPremium,
                premiumClassificationKnown: premiumClassification.isKnown,
              );
            }
          }
          if (!await _matchesCurrentAccount(accountId)) return;
          final updatedAt = DateTime.now();
          _report(
            CachedResponseMetadata(
              cacheKey: cacheKey,
              cached: false,
              offline: false,
              stale: false,
              updatedAt: updatedAt,
            ),
          );
          response.extra['offlineCache'] = <String, Object?>{
            'cached': false,
            'stored': stored,
            'offline': false,
            'stale': false,
            'updatedAt': updatedAt.toIso8601String(),
          };
        }
      }
    }

    try {
      await persist().timeout(_cacheBudget);
    } on Object {
      // Network success wins if disk persistence fails or is slow; a slow
      // write keeps running in the background.
    } finally {
      if (dedupeKey is String) {
        final completer = _inFlight.remove(dedupeKey);
        if (completer != null && !completer.isCompleted) {
          completer.complete(response);
        }
      }
      handler.next(response);
    }
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final options = err.requestOptions;
    final cacheKey = options.extra[_cacheKeyExtraKey];
    final accountId = options.extra[_accountExtraKey];
    final cacheOwner = options.extra[_cacheOwnerExtraKey];
    final dedupeKey = options.extra[_dedupeCompleterExtraKey];
    final statusCode = err.response?.statusCode;
    Future<void> record() async {
      if (cacheKey is String && accountId is String) {
        if (statusCode == 401 || statusCode == 403 || statusCode == 402) {
          final store = await _cacheFuture;
          await store.removeResponse(accountId: accountId, key: cacheKey);
          if (cacheOwner == guestCacheAccountId &&
              _isPublicLessonKey(cacheKey) &&
              _requiresPremium(options)) {
            await store.removeResponse(
              accountId: guestCacheAccountId,
              key: cacheKey,
            );
          }
          _report(
            CachedResponseMetadata(
              cacheKey: cacheKey,
              cached: false,
              offline: false,
              stale: false,
              updatedAt: null,
            ),
          );
        } else {
          final store = await _cacheFuture;
          final retained = await store.readResponse(
            accountId: accountId,
            key: cacheKey,
          );
          _report(
            CachedResponseMetadata(
              cacheKey: cacheKey,
              cached: retained != null,
              offline: retained != null,
              stale: retained?.stale ?? false,
              updatedAt: retained?.updatedAt,
            ),
          );
        }
      }
    }

    try {
      await record().timeout(_cacheBudget);
    } on Object {
      // Preserve the original network error.
    } finally {
      if (dedupeKey is String) {
        final completer = _inFlight.remove(dedupeKey);
        if (completer != null && !completer.isCompleted) {
          completer.completeError(err, StackTrace.current);
        }
      }
      handler.next(err);
    }
  }

  Future<void> _refreshInBackground(
    RequestOptions original,
    String accountId,
    String cacheKey, {
    required bool requiresPremium,
  }) async {
    final scopedKey = '$accountId\n$cacheKey';
    if (!_refreshing.add(scopedKey)) return;
    var acquiredRefreshSlot = false;
    try {
      if (_accountId(await currentUserId()) != accountId) return;
      await _acquireRefreshSlot();
      acquiredRefreshSlot = true;
      // A stale entry may have waited behind other refreshes. Revalidate the
      // account immediately before any queued network request is dispatched.
      if (_accountId(await currentUserId()) != accountId) return;
      if (requiresPremium &&
          (accountId == guestCacheAccountId ||
              !await canReadPremium(accountId))) {
        final store = await _cacheFuture;
        await store.removeResponse(accountId: accountId, key: cacheKey);
        return;
      }
      final extra = <String, dynamic>{
        ...original.extra,
        _bypassExtraKey: true,
        _accountExtraKey: accountId,
        _cacheKeyExtraKey: cacheKey,
      }..remove(_dedupeCompleterExtraKey);
      final options = original.copyWith(extra: extra);
      await dio.fetch<dynamic>(options);
    } on DioException catch (error) {
      final statusCode = error.response?.statusCode;
      if (_accountId(await currentUserId()) != accountId) return;
      if (statusCode == 401 || statusCode == 402 || statusCode == 403) {
        final store = await _cacheFuture;
        await store.removeResponse(accountId: accountId, key: cacheKey);
        if (original.extra[_cacheOwnerExtraKey] == guestCacheAccountId &&
            _isPublicLessonKey(cacheKey) &&
            _requiresPremium(original)) {
          await store.removeResponse(
            accountId: guestCacheAccountId,
            key: cacheKey,
          );
        }
        _report(
          CachedResponseMetadata(
            cacheKey: cacheKey,
            cached: false,
            offline: false,
            stale: false,
            updatedAt: null,
          ),
        );
      } else {
        final store = await _cacheFuture;
        final cached = await store.readResponse(
          accountId: accountId,
          key: cacheKey,
        );
        _report(
          CachedResponseMetadata(
            cacheKey: cacheKey,
            cached: cached != null,
            offline: cached != null,
            stale: cached?.stale ?? false,
            updatedAt: cached?.updatedAt,
          ),
        );
      }
    } on Object {
      if (_accountId(await currentUserId()) != accountId) return;
      final store = await _cacheFuture;
      final cached = await store.readResponse(
        accountId: accountId,
        key: cacheKey,
      );
      _report(
        CachedResponseMetadata(
          cacheKey: cacheKey,
          cached: cached != null,
          offline: cached != null,
          stale: cached?.stale ?? false,
          updatedAt: cached?.updatedAt,
        ),
      );
    } finally {
      if (acquiredRefreshSlot) _releaseRefreshSlot();
      _refreshing.remove(scopedKey);
    }
  }

  Future<void> _acquireRefreshSlot() async {
    if (_activeRefreshes < _maxConcurrentRefreshes) {
      _activeRefreshes++;
      return;
    }
    final waiter = Completer<void>();
    _refreshQueue.addLast(waiter);
    await waiter.future;
  }

  void _releaseRefreshSlot() {
    if (_refreshQueue.isNotEmpty) {
      _refreshQueue.removeFirst().complete();
      return;
    }
    _activeRefreshes--;
  }

  Response<dynamic> _toResponse(
    RequestOptions options,
    CachedResponse cached,
    CachedResponseMetadata metadata,
  ) {
    final data = options.responseType == ResponseType.bytes
        ? cached.body
        : options.responseType == ResponseType.plain
        ? utf8.decode(cached.body)
        : _decodeBody(cached.body, cached.contentType);
    return Response<dynamic>(
      requestOptions: options,
      data: data,
      statusCode: cached.statusCode,
      statusMessage: cached.statusMessage,
      headers: Headers.fromMap(<String, List<String>>{
        for (final entry in cached.headers.entries)
          entry.key: <String>[entry.value],
      }),
      extra: <String, dynamic>{
        ...options.extra,
        'offlineCache': metadata.toExtra(),
      },
    );
  }

  void _report(CachedResponseMetadata metadata) {
    try {
      onMetadata?.call(metadata);
    } on Object {
      // UI observers are optional and cannot fail a request.
    }
  }

  Future<bool> _matchesCurrentAccount(String accountId) async {
    try {
      return _accountId(await currentUserId()) == accountId;
    } on Object {
      return false;
    }
  }

  DioException _accountChanged(RequestOptions options) => DioException(
    requestOptions: options,
    type: DioExceptionType.cancel,
    error: 'The active account changed before this request could be served.',
  );
}

bool _isPublicLessonKey(String key) {
  final path = Uri.tryParse(key)?.path;
  if (path == null) return false;
  final segments = path.split('/').where((part) => part.isNotEmpty).toList();
  return segments.length == 3 &&
      const <String>{
        'vocabulary',
        'grammar',
        'kanji',
        'reading',
      }.contains(segments[0]) &&
      segments[1] == 'lessons';
}

Uint8List? _sanitizePublicLessonBody(Uint8List body) {
  try {
    final decoded = jsonDecode(utf8.decode(body));
    final sanitized = _removePrivateLessonFields(decoded);
    return Uint8List.fromList(utf8.encode(jsonEncode(sanitized)));
  } on Object {
    return null;
  }
}

Object? _removePrivateLessonFields(Object? value) {
  const privateFields = <String>{
    'attempts',
    'completedat',
    'completion',
    'iscompleted',
    'lastattempt',
    'lastattemptat',
    'lessonprogress',
    'locked',
    'nextreviewat',
    'progress',
    'progressdata',
    'result',
    'review',
    'score',
    'stats',
    'streak',
    'totalxp',
    'user',
    'userid',
    'userprogress',
    'xp',
  };
  if (value is Map) {
    return <String, Object?>{
      for (final entry in value.entries)
        if (!privateFields.contains(entry.key.toString().toLowerCase()))
          entry.key.toString(): switch (entry.key.toString().toLowerCase()) {
            'bestscore' => 0,
            'completed' => false,
            _ => _removePrivateLessonFields(entry.value),
          },
    };
  }
  if (value is List) {
    return value.map<Object?>(_removePrivateLessonFields).toList();
  }
  return value;
}

String _accountId(String? userId) =>
    userId == null || userId.trim().isEmpty ? guestCacheAccountId : userId;

String? _acceptLanguage(Map<String, dynamic> headers) {
  for (final entry in headers.entries) {
    if (entry.key.toLowerCase() != 'accept-language' ||
        entry.value is! String) {
      continue;
    }
    final language = (entry.value as String).trim().toLowerCase();
    if (language == 'id' || language == 'en') return language;
  }
  return null;
}

String? _requestCacheKey(RequestOptions options) {
  if (options.method.toUpperCase() != 'GET') return null;
  final path = _apiPath(options.uri.path);
  final segments = path.split('/').where((part) => part.isNotEmpty).toList();
  if (path == '/home') {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  if (segments.length == 2 &&
      const <String>{
        'vocabulary',
        'grammar',
        'kanji',
        'reading',
      }.contains(segments[0]) &&
      const <String>{'levels', 'course'}.contains(segments[1])) {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  if (segments.length == 3 &&
      const <String>{
        'vocabulary',
        'grammar',
        'kanji',
        'reading',
      }.contains(segments[0]) &&
      segments[1] == 'lessons') {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  if (segments.length == 3 &&
      segments[0] == 'ssw' &&
      segments[2] == 'practice') {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  if (segments.length == 2 && segments[0] == 'anki' && segments[1] == 'decks' ||
      segments.length == 3 && segments[0] == 'anki' && segments[1] == 'decks' ||
      segments.length == 4 &&
          segments[0] == 'anki' &&
          segments[1] == 'decks' &&
          segments[3] == 'study') {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  if (segments.length == 4 &&
      segments[0] == 'offline' &&
      segments[1] == 'practice' &&
      const <String>{
        'vocabulary',
        'grammar',
        'kanji',
        'reading',
        'ssw',
      }.contains(segments[2])) {
    return _canonicalUrlKey(path, options.uri.queryParametersAll);
  }
  return null;
}

String? _mutationPlan(RequestOptions options) {
  if (const <String>{
    'POST',
    'PUT',
    'PATCH',
    'DELETE',
  }.contains(options.method.toUpperCase())) {
    final path = _apiPath(options.uri.path);
    final segments = path.split('/').where((part) => part.isNotEmpty).toList();
    if (path == '/anki/import' ||
        options.method.toUpperCase() == 'DELETE' &&
            segments.length == 3 &&
            segments[0] == 'anki' &&
            segments[1] == 'decks') {
      return 'anki-library';
    }
    if (segments.length == 4 &&
        segments[0] == 'anki' &&
        segments[1] == 'decks' &&
        segments[3] == 'review') {
      return 'anki-review:${segments[2]}';
    }
    if (segments.length == 4 &&
        const <String>{
          'vocabulary',
          'grammar',
          'kanji',
          'reading',
        }.contains(segments[0]) &&
        segments[1] == 'attempts' &&
        segments[3] == 'complete') {
      return 'lesson-complete:${segments[0]}';
    }
    if (path == '/offline/practice-attempts/sync') return 'practice-sync';
    if (path == '/billing/verify') return 'billing-verify';
  }
  return null;
}

bool _invalidateForMutation(String routeKey, String plan) {
  final path = Uri.parse(routeKey).path;
  if (path == '/home') return true;
  if (plan == 'anki-library') {
    return path == '/anki/decks' || path.startsWith('/anki/decks/');
  }
  if (plan == 'practice-sync') {
    return const <String>{
      'vocabulary',
      'grammar',
      'kanji',
      'reading',
    }.any((kind) => _isCourseSnapshot(path, kind));
  }
  if (plan == 'billing-verify') {
    return path == '/vocabulary/levels' ||
        path == '/anki/decks' ||
        const <String>{
          'vocabulary',
          'grammar',
          'kanji',
          'reading',
        }.any((kind) => _isCourseSnapshot(path, kind));
  }
  if (plan.startsWith('anki-review:')) {
    final deckId = Uri.encodeComponent(plan.substring('anki-review:'.length));
    return path == '/anki/decks' ||
        path == '/anki/decks/$deckId' ||
        path == '/anki/decks/$deckId/study';
  }
  if (plan.startsWith('lesson-complete:')) {
    final kind = plan.substring('lesson-complete:'.length);
    return _isCourseSnapshot(path, kind) ||
        path.startsWith('/$kind/lessons/') ||
        (kind == 'vocabulary' && path == '/vocabulary/levels');
  }
  return false;
}

bool _isCourseSnapshot(String path, String kind) =>
    path == '/$kind/course' || path.startsWith('/$kind/course/');

class _PremiumClassification {
  const _PremiumClassification({
    required this.requiresPremium,
    required this.isKnown,
  });

  final bool requiresPremium;
  final bool isKnown;
}

bool _isAnkiDeckRequest(RequestOptions options) {
  final path = _apiPath(options.uri.path);
  final segments = path.split('/').where((part) => part.isNotEmpty).toList();
  return segments.length == 3 &&
          segments[0] == 'anki' &&
          segments[1] == 'decks' ||
      segments.length == 4 &&
          segments[0] == 'anki' &&
          segments[1] == 'decks' &&
          segments[3] == 'study';
}

_PremiumClassification _responsePremiumClassification(
  RequestOptions options,
  Object? responseBody,
) {
  if (_isAnkiDeckRequest(options)) {
    final classification = _ankiPremiumFromPayload(responseBody);
    return classification == null
        ? const _PremiumClassification(requiresPremium: true, isKnown: false)
        : _PremiumClassification(
            requiresPremium: classification,
            isKnown: true,
          );
  }
  return _PremiumClassification(
    requiresPremium: _requiresPremium(options),
    isKnown: true,
  );
}

bool? _ankiPremiumFromPayload(Object? value) {
  var sawFree = false;
  var sawPremium = false;

  void visit(Object? node) {
    if (node is Map) {
      final classification = classifyAnkiPremiumMetadata(
        level: node['level'],
        requiredTier: node['requiredTier'] ?? node['required_tier'],
      );
      if (classification ?? false) {
        sawPremium = true;
      } else if (classification != null) {
        sawFree = true;
      }
      for (final child in node.values) {
        visit(child);
      }
    } else if (node is Iterable) {
      for (final child in node) {
        visit(child);
      }
    }
  }

  visit(value);
  if (sawPremium) return true;
  if (sawFree) return false;
  return null;
}

/// Classifies trusted Anki metadata. `null` means the server did not provide a
/// recognized level or tier, which callers must treat as premium/fail-closed.
bool? classifyAnkiPremiumMetadata({Object? level, Object? requiredTier}) {
  final levelName = level is String ? level.trim().toUpperCase() : '';
  final tierName = requiredTier is String
      ? requiredTier.trim().toLowerCase()
      : '';
  if (tierName == 'premium' ||
      tierName == 'lifetime' ||
      tierName == 'paid' ||
      const <String>{'N1', 'N2', 'N3'}.contains(levelName)) {
    return true;
  }
  if (const <String>{'N4', 'N5'}.contains(levelName) ||
      tierName == 'free' ||
      tierName == 'basic' ||
      tierName == 'standard') {
    return false;
  }
  return null;
}

String _apiPath(String path) => path == '/v1'
    ? '/'
    : path.startsWith('/v1/')
    ? path.substring(3)
    : path;

String _canonicalUrlKey(String path, Map<String, List<String>> query) {
  final keys = query.keys.toList()..sort();
  final canonicalQuery = <String, List<String>>{
    for (final key in keys) key: (query[key] ?? <String>[]).toList()..sort(),
  };
  return canonicalQuery.isEmpty
      ? path
      : '$path?${Uri(queryParameters: <String, String>{for (final entry in canonicalQuery.entries) entry.key: entry.value.join(',')}).query}';
}

bool _requiresPremium(RequestOptions options) {
  final path = _apiPath(options.uri.path);
  final segments = path.split('/').where((part) => part.isNotEmpty).toList();
  if (segments.length == 3 &&
      segments[0] == 'ssw' &&
      segments[2] == 'practice') {
    return true;
  }
  if (segments.length == 4 &&
      segments[0] == 'offline' &&
      segments[1] == 'practice' &&
      segments[2] == 'ssw') {
    return true;
  }
  if (_isAnkiDeckRequest(options)) return true;
  String? level;
  if (options.uri.queryParameters['level'] != null) {
    level = options.uri.queryParameters['level'];
  } else if (segments.length >= 3 && segments[1] == 'lessons') {
    level = _levelInId(segments[2]);
  } else if (segments.length >= 4 &&
      segments[0] == 'offline' &&
      segments[1] == 'practice') {
    level = _levelInId(segments[3]);
  }
  return level != null &&
      RegExp(r'^N[1-3]$', caseSensitive: false).hasMatch(level);
}

String? _levelInId(String id) {
  final match = RegExp(
    r'(?:^|[-_])(N[1-5])(?:[-_]|$)',
    caseSensitive: false,
  ).firstMatch(id);
  return match?.group(1);
}

Uint8List? _encodeBody(Object? data) {
  try {
    if (data is Uint8List) return Uint8List.fromList(data);
    if (data is List<int>) return Uint8List.fromList(data);
    if (data is String) return Uint8List.fromList(utf8.encode(data));
    if (data is Map || data is List) {
      return Uint8List.fromList(utf8.encode(jsonEncode(data)));
    }
  } on Object {
    return null;
  }
  return null;
}

bool _isErrorPayload(Object? data) => data is Map && data['error'] != null;

String _contentType(Headers headers, Object? data) {
  final values = headers.map.entries
      .where((entry) => entry.key.toLowerCase() == 'content-type')
      .expand((entry) => entry.value);
  final contentType = values.firstOrNull;
  if (contentType != null) return contentType;
  if (data is Map || data is List) return 'application/json';
  return 'text/plain; charset=utf-8';
}

Map<String, String> _safeHeaders(Headers headers) {
  const safeNames = <String>{'content-type', 'etag', 'last-modified'};
  return <String, String>{
    for (final entry in headers.map.entries)
      if (safeNames.contains(entry.key.toLowerCase()) && entry.value.isNotEmpty)
        entry.key.toLowerCase(): entry.value.join(', '),
  };
}

Object? _decodeBody(Uint8List bytes, String contentType) {
  final text = utf8.decode(bytes);
  if (contentType.toLowerCase().contains('json')) return jsonDecode(text);
  return text;
}
