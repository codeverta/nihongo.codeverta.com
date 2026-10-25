import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Small account-scoped write-ahead queue. The server still owns scheduling;
/// requestId makes a retry after an uncertain timeout the same operation.
class OfflineAnkiReviews {
  OfflineAnkiReviews({
    required this.dio,
    required this.preferences,
    required this.currentUserId,
    required this.isOffline,
    required this.canReadPremium,
  });

  final Dio dio;
  final SharedPreferences preferences;
  final String? Function() currentUserId;
  final bool Function() isOffline;
  final Future<bool> Function(String) canReadPremium;
  Future<void> _writes = Future.value();
  Future<void>? _flushing;
  final Map<String, int> _generations = {};

  String _key(String user) => 'anki_reviews_v1_$user';

  List<Map<String, dynamic>> _read(String user) {
    try {
      return (jsonDecode(preferences.getString(_key(user)) ?? '[]') as List)
          .map((e) => Map<String, dynamic>.from(e as Map))
          .toList();
    } on Object {
      return [];
    }
  }

  bool get hasPending =>
      currentUserId() != null && _read(currentUserId()!).isNotEmpty;

  Set<int> pendingCards(String deckId) {
    final user = currentUserId();
    if (user == null) return {};
    return _read(
      user,
    ).where((e) => e['deckId'] == deckId).map((e) => e['index'] as int).toSet();
  }

  Future<void> _mutate(
    String user,
    void Function(List<Map<String, dynamic>>) action,
  ) {
    final generation = _generations[user] ?? 0;
    final operation = _writes.then((_) async {
      if ((_generations[user] ?? 0) != generation) return;
      final entries = _read(user);
      action(entries);
      if (!await preferences.setString(_key(user), jsonEncode(entries))) {
        throw StateError('Penilaian offline belum tersimpan. Coba lagi.');
      }
    });
    _writes = operation.catchError((Object _) {});
    return operation;
  }

  Future<Map<String, dynamic>> submit({
    required String deckId,
    required int index,
    required int version,
    required String rating,
    required String requestId,
    required bool premium,
  }) async {
    final user = currentUserId();
    if (user == null) throw StateError('Masuk untuk menyimpan review.');
    final canQueue = !premium || await canReadPremium(user);
    if (!canQueue && isOffline()) {
      throw DioException(
        requestOptions: RequestOptions(path: '/anki'),
        response: Response(
          requestOptions: RequestOptions(path: '/anki'),
          statusCode: 402,
        ),
      );
    }
    if (currentUserId() != user) throw StateError('Akun sudah berubah.');
    final entry = <String, dynamic>{
      'deckId': deckId,
      'index': index,
      'version': version,
      'rating': rating,
      'requestId': requestId,
      'premium': premium,
    };
    // Online authorization belongs to the server. A missing/expired offline
    // proof must not prevent a verified subscriber from reviewing online, but
    // it must not create a local premium queue on an uncertain network failure.
    if (!canQueue) return _send(user, entry);
    await _mutate(user, (entries) {
      if (entries.any((e) => e['requestId'] == requestId)) return;
      if (entries.length >= 1000) {
        throw StateError(
          'Antrean review penuh. Hubungkan internet untuk sinkronisasi.',
        );
      }
      entries.add(entry);
    });
    if (currentUserId() != user) throw StateError('Akun sudah berubah.');
    if (isOffline()) return {'pendingSync': true};
    try {
      return await _send(user, entry);
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if ([404, 409, 422].contains(status)) {
        await _mutate(
          user,
          (entries) =>
              entries.removeWhere((item) => item['requestId'] == requestId),
        );
      }
      if (status != null && status < 500) rethrow;
      return {'pendingSync': true};
    }
  }

  Future<Map<String, dynamic>> _send(
    String user,
    Map<String, dynamic> entry,
  ) async {
    if (currentUserId() != user) throw StateError('Akun sudah berubah.');
    final generation = _generations[user] ?? 0;
    final response = await dio.post<Map<String, dynamic>>(
      '/anki/decks/${Uri.encodeComponent(entry['deckId'] as String)}/review',
      data: {
        for (final key in ['index', 'version', 'rating', 'requestId'])
          key: entry[key],
      },
      options: Options(extra: {'offlineAccountId': user}),
    );
    if (currentUserId() != user || generation != (_generations[user] ?? 0)) {
      throw StateError('Akun sudah berubah.');
    }
    await _mutate(
      user,
      (entries) =>
          entries.removeWhere((e) => e['requestId'] == entry['requestId']),
    );
    return Map<String, dynamic>.from(response.data?['data'] as Map? ?? {});
  }

  Future<void> flush() {
    if (_flushing != null) return _flushing!;
    return _flushing = _flush().whenComplete(() => _flushing = null);
  }

  Future<void> _flush() async {
    final user = currentUserId();
    if (user == null || isOffline()) return;
    for (final entry in _read(user)) {
      if (currentUserId() != user) return;
      if (entry['premium'] == true && !await canReadPremium(user)) continue;
      try {
        await _send(user, entry);
      } on DioException catch (e) {
        // Another device may already have changed the schedule. Never overwrite
        // that newer schedule using an old cached card version.
        if ([404, 409, 422].contains(e.response?.statusCode)) {
          await _mutate(
            user,
            (entries) => entries.removeWhere(
              (a) => a['requestId'] == entry['requestId'],
            ),
          );
          continue;
        }
        return;
      } on Object {
        return;
      }
    }
  }

  Future<void> clearAccount(String user) async {
    _generations[user] = (_generations[user] ?? 0) + 1;
    await _writes;
    await preferences.remove(_key(user));
  }
}
