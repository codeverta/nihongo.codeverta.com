import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/core/gamification/xp_gain.dart';

/// One locked item that an ad or Lifetime Access can open. [kind] uses the
/// server's names: grammar, kanji, reading, vocabulary, anki, simulation,
/// news, job.
@immutable
class UnlockTarget {
  const UnlockTarget(this.kind, this.itemId);

  /// The item a 402 `premium_required` response offers to unlock, or null
  /// when the response offers no ad (Lifetime only) or is something else.
  static UnlockTarget? fromError(Object? error) {
    if (error is! DioException || error.response?.statusCode != 402) {
      return null;
    }
    final data = error.response?.data;
    final body = data is Map ? data['error'] : null;
    if (body is! Map || body['unlockable'] != true) return null;
    final unlock = body['unlock'];
    if (unlock is! Map) return null;
    final kind = unlock['kind'];
    final itemId = unlock['itemId'];
    if (kind is! String ||
        itemId is! String ||
        kind.isEmpty ||
        itemId.isEmpty) {
      return null;
    }
    return UnlockTarget(kind, itemId);
  }

  /// True for a 402 `premium_required` response, whether or not an ad is
  /// offered. Lets screens show the Lifetime option instead of a generic error.
  static bool isLocked(Object? error) {
    if (error is! DioException || error.response?.statusCode != 402) {
      return false;
    }
    final data = error.response?.data;
    final body = data is Map ? data['error'] : null;
    return body is Map && body['code'] == 'premium_required';
  }

  final String kind;
  final String itemId;

  /// The key the server stores, for example `grammar:grammar-n3-01`.
  String get key => '$kind:$itemId';

  /// The group used by the remote switches and the gate flags.
  String get group => switch (kind) {
    'grammar' || 'kanji' || 'reading' || 'vocabulary' => 'lessons',
    'job' => 'jobs',
    _ => kind,
  };

  Map<String, String> toJson() => {'kind': kind, 'itemId': itemId};

  @override
  bool operator ==(Object other) =>
      other is UnlockTarget && other.kind == kind && other.itemId == itemId;

  @override
  int get hashCode => Object.hash(kind, itemId);

  @override
  String toString() => key;
}

/// What the account has opened and what is locked right now.
@immutable
class UnlockStatus {
  const UnlockStatus({
    this.items = const {},
    this.gates = const {},
    this.offered = const {},
    this.remaining = 0,
    this.fallbackRemaining = 0,
    this.dailyCap = 0,
  });

  factory UnlockStatus.fromJson(Map<String, dynamic> json) {
    Map<String, bool> flags(Object? value) => {
      if (value is Map)
        for (final entry in value.entries)
          if (entry.value is bool) '${entry.key}': entry.value as bool,
    };
    int number(Object? value) => value is num ? value.toInt() : 0;
    return UnlockStatus(
      items: {
        for (final item in (json['items'] as List?) ?? const [])
          if (item is String) item,
      },
      gates: flags(json['gates']),
      offered: flags(json['offered']),
      remaining: number(json['remaining']),
      fallbackRemaining: number(json['fallbackRemaining']),
      dailyCap: number(json['dailyCap']),
    );
  }

  /// Nothing known: draws no locks.
  static const unknown = UnlockStatus();

  final Set<String> items;

  /// Which groups (lessons, news, jobs) are locked for free accounts.
  final Map<String, bool> gates;

  /// Which groups an ad may open.
  final Map<String, bool> offered;
  final int remaining;
  final int fallbackRemaining;
  final int dailyCap;

  bool isOpen(UnlockTarget target) => items.contains(target.key);

  UnlockStatus withItem(UnlockTarget target, {int? remaining}) => UnlockStatus(
    items: {...items, target.key},
    gates: gates,
    offered: offered,
    remaining: remaining ?? this.remaining,
    fallbackRemaining: fallbackRemaining,
    dailyCap: dailyCap,
  );
}

/// The result of asking the server to start an unlock.
@immutable
class UnlockIntent {
  const UnlockIntent({
    required this.alreadyUnlocked,
    this.nonce = '',
    this.userId = '',
  });
  final bool alreadyUnlocked;
  final String nonce;
  final String userId;
}

/// A refusal the user should read: daily limit, ads unavailable, and so on.
class UnlockException implements Exception {
  const UnlockException(this.code, this.message);

  factory UnlockException.from(DioException error) {
    final data = error.response?.data;
    final body = data is Map ? data['error'] : null;
    if (body is Map) {
      return UnlockException(
        '${body['code'] ?? 'unknown'}',
        '${body['message'] ?? ''}'.trim(),
      );
    }
    return const UnlockException('network', '');
  }

  static const dailyLimit = 'unlock_daily_limit';
  static const fallbackExhausted = 'unlock_fallback_exhausted';
  static const unavailable = 'unlock_unavailable';
  static const expired = 'intent_expired';

  final String code;
  final String message;

  @override
  String toString() => message.isEmpty ? code : message;
}

/// Talks to `/v1/unlocks`. The server decides everything; an ad only counts
/// once Google's signed callback has reached it.
class UnlockRepository {
  UnlockRepository(this._dio);
  final Dio _dio;

  Future<UnlockStatus> status() async {
    final response = await _dio.get<Map<String, dynamic>>('/unlocks');
    final data = response.data?['data'];
    return data is Map<String, dynamic>
        ? UnlockStatus.fromJson(data)
        : UnlockStatus.unknown;
  }

  Future<UnlockIntent> createIntent(UnlockTarget target) async {
    final data = await _post('/unlocks/intents', target.toJson());
    return UnlockIntent(
      alreadyUnlocked: data['alreadyUnlocked'] == true,
      nonce: '${data['nonce'] ?? ''}',
      userId: '${data['userId'] ?? ''}',
    );
  }

  /// True once the server has recorded the unlock; false while Google's
  /// callback has not arrived yet (the server already waited a few seconds).
  Future<bool> claim(String nonce) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/unlocks/claim',
        data: {'nonce': nonce},
        options: Options(receiveTimeout: const Duration(seconds: 20)),
      );
      final data = response.data?['data'];
      return response.statusCode == 200 &&
          data is Map &&
          data['unlocked'] == true;
    } on DioException catch (error) {
      throw UnlockException.from(error);
    }
  }

  /// Doubles the XP of [completionId] without an ad (Lifetime only).
  Future<XpGain?> claimFreeBoost(String completionId) async {
    final data = await _post(
      '/xp-boosts/${Uri.encodeComponent(completionId)}/free',
      const <String, Object>{},
    );
    return XpGain.fromJson(data['xp']);
  }

  /// Opens the item without an ad, when no ad could be loaded. The server
  /// allows only a few of these per day.
  Future<void> fallback(UnlockTarget target) async {
    await _post('/unlocks/fallback', target.toJson());
  }

  Future<Map<String, dynamic>> _post(String path, Object body) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(path, data: body);
      final data = response.data?['data'];
      return data is Map<String, dynamic> ? data : const {};
    } on DioException catch (error) {
      throw UnlockException.from(error);
    }
  }
}

final unlockRepositoryProvider = Provider<UnlockRepository>(
  (ref) => UnlockRepository(ref.watch(dioProvider)),
);

/// The signed-in account's unlocks and the server's gates. Guests and failures
/// yield [UnlockStatus.unknown], which draws no locks: the server still
/// refuses locked content, and the screen then offers the unlock sheet.
final unlockStatusProvider = FutureProvider<UnlockStatus>((ref) async {
  final user = ref.watch(authControllerProvider.select((s) => s.user?.id));
  if (user == null) return UnlockStatus.unknown;
  try {
    return await ref.watch(unlockRepositoryProvider).status();
  } on Object catch (error) {
    debugPrint('Unlock status unavailable: $error');
    return UnlockStatus.unknown;
  }
});
