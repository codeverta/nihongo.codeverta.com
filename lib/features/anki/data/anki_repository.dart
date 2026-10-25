import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/anki/data/offline_anki_reviews.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

final offlineAnkiReviewsProvider = Provider(
  (ref) => OfflineAnkiReviews(
    dio: ref.watch(dioProvider),
    preferences: ref.watch(sharedPreferencesProvider),
    currentUserId: () => ref.read(authControllerProvider).user?.id,
    isOffline: () => ref.read(offlineProvider),
    canReadPremium: (user) =>
        ref.read(offlineAccessStoreProvider).canReadPremium(user),
  ),
);
final ankiRepositoryProvider = Provider(
  (ref) => AnkiRepository(
    ref.watch(dioProvider),
    offlineReviews: ref.watch(offlineAnkiReviewsProvider),
  ),
);
final ankiDecksProvider = FutureProvider.autoDispose<List<AnkiDeck>>((ref) {
  ref.watch(authControllerProvider.select((state) => state.user?.id));
  return ref.watch(ankiRepositoryProvider).decks();
});
final ankiDeckProvider = FutureProvider.autoDispose.family<AnkiDeck, String>((
  ref,
  id,
) {
  ref.watch(
    authControllerProvider.select((s) => (s.user?.id, s.user?.isPremium)),
  );
  return ref.watch(ankiRepositoryProvider).deck(id);
});
final ankiCreditsProvider = FutureProvider.autoDispose<Map<String, dynamic>>(
  (ref) => ref.watch(ankiRepositoryProvider).credits(),
);

class AnkiRepository {
  AnkiRepository(this.dio, {this.offlineReviews});
  final Dio dio;
  final OfflineAnkiReviews? offlineReviews;
  bool get hasPendingReviews => offlineReviews?.hasPending ?? false;
  Future<Map<String, dynamic>> request(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    try {
      final result = body == null
          ? await dio.get<Map<String, dynamic>>(path)
          : await dio.post<Map<String, dynamic>>(path, data: body);
      return result.data!['data'] as Map<String, dynamic>;
    } on DioException catch (error) {
      final status = error.response?.statusCode;
      final message = switch (status) {
        401 => 'Masuk untuk belajar dan menyimpan jadwal review.',
        402 => 'Deck ini memerlukan Lifetime Access.',
        404 => 'Deck tidak ditemukan.',
        409 => 'Jadwal kartu sudah berubah. Muat ulang sesi untuk melanjutkan.',
        _ => 'Belum dapat terhubung ke server. Coba kembali.',
      };
      throw AnkiFailure(message, status);
    }
  }

  Future<List<AnkiDeck>> decks() async =>
      ((await request('/anki/decks'))['decks'] as List<dynamic>)
          .map((d) => AnkiDeck.fromJson(d as Map<String, dynamic>))
          .toList();
  Future<AnkiDeck> deck(String id) async => AnkiDeck.fromJson(
    await request('/anki/decks/${Uri.encodeComponent(id)}'),
  );
  Future<AnkiSession> study(String id) async {
    final data = await request('/anki/decks/${Uri.encodeComponent(id)}/study');
    final pending = offlineReviews?.pendingCards(id) ?? const <int>{};
    return AnkiSession.fromJson({
      ...data,
      'cards': (data['cards'] as List)
          .where((card) => !pending.contains((card as Map)['index']))
          .toList(),
    });
  }

  Future<Map<String, dynamic>> review(
    String id,
    AnkiCard card,
    String rating,
    String requestId,
  ) async {
    final outbox = offlineReviews;
    if (outbox != null) {
      final metadata = await deck(id);
      try {
        return await outbox.submit(
          deckId: id,
          index: card.index,
          version: card.version,
          rating: rating,
          requestId: requestId,
          premium: metadata.premium,
        );
      } on DioException catch (error) {
        final status = error.response?.statusCode;
        throw AnkiFailure(
          status == 409
              ? 'Jadwal kartu berubah di perangkat lain. Muat ulang sesi untuk memakai jadwal terbaru.'
              : status == 402
              ? 'Verifikasi Lifetime Access diperlukan untuk deck ini.'
              : 'Penilaian belum tersimpan. Coba kembali.',
          status,
        );
      }
    }
    return request(
      '/anki/decks/${Uri.encodeComponent(id)}/review',
      body: {
        'index': card.index,
        'version': card.version,
        'rating': rating,
        'requestId': requestId,
      },
    );
  }

  Future<Map<String, dynamic>> credits() => request('/anki/credits');
}

class AnkiFailure implements Exception {
  const AnkiFailure(this.message, this.status);
  final String message;
  final int? status;
  @override
  String toString() => message;
}
