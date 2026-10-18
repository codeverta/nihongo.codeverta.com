import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';

final dictionaryRepositoryProvider = Provider<DictionaryRepository>((ref) {
  return DictionaryRepository(ref.watch(dioProvider));
});

class DictionaryRepository {
  DictionaryRepository(this.dio);
  final Dio dio;
  Future<DictionaryPage> search(
    String query, {
    int page = 1,
    String? jlpt,
    bool common = false,
    String mode = 'auto',
  }) async {
    final response = await dio.get<Map<String, dynamic>>(
      '/dictionary/search',
      queryParameters: {
        'q': query,
        'mode': mode,
        'page': page,
        'jlpt': jlpt,
        'common': common,
      },
      options: Options(receiveTimeout: const Duration(seconds: 15)),
    );
    return DictionaryPage.fromJson(
      response.data!['data'] as Map<String, dynamic>,
    );
  }

  Future<Map<String, dynamic>> kanji(String character) async {
    final response = await dio.get<Map<String, dynamic>>(
      '/dictionary/kanji/${Uri.encodeComponent(character)}',
      options: Options(receiveTimeout: const Duration(seconds: 15)),
    );
    return response.data!['data'] as Map<String, dynamic>;
  }
}

final dictionaryKanjiProvider = FutureProvider.autoDispose
    .family<Map<String, dynamic>, String>((ref, character) {
      return ref.watch(dictionaryRepositoryProvider).kanji(character);
    });
