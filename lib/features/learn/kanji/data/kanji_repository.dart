import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';

class KanjiException implements Exception {
  const KanjiException(this.message);
  final String message;
  @override
  String toString() => message;
}

class KanjiRepository {
  const KanjiRepository(this._dio);
  final Dio _dio;
  Future<KanjiCourse> getCourse({String level = 'N5'}) async =>
      KanjiCourse.fromJson(
        await _request(
          '/kanji/course?level=${Uri.encodeQueryComponent(level)}',
        ),
      );
  Future<KanjiLesson> getLesson(String id) async => KanjiLesson.fromJson(
        await _request('/kanji/lessons/${Uri.encodeComponent(id)}'),
      );
  Future<KanjiAttempt> startAttempt(String id) async => KanjiAttempt.fromJson(
        await _request(
          '/kanji/lessons/${Uri.encodeComponent(id)}/attempts',
          post: true,
        ),
      );
  Future<KanjiFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async =>
      KanjiFeedback.fromJson(
        await _request(
          '/kanji/attempts/$attempt/answers',
          post: true,
          body: {'questionId': question, 'optionId': option},
        ),
      );
  Future<KanjiResult> complete(String id) async => KanjiResult.fromJson(
        await _request('/kanji/attempts/$id/complete', post: true),
      );
  Future<Map<String, dynamic>> _request(
    String path, {
    bool post = false,
    Map<String, dynamic>? body,
  }) async {
    try {
      final response = post
          ? await _dio.post<Map<String, dynamic>>(path, data: body)
          : await _dio.get<Map<String, dynamic>>(path);
      final data = response.data?['data'];
      if (data is! Map<String, dynamic>) {
        throw const KanjiException(
          'Respons kanji tidak valid. Coba muat ulang.',
        );
      }
      return data;
    } on DioException catch (error) {
      if (error.response?.statusCode == 401) {
        throw const KanjiException('Sesi berakhir. Silakan masuk kembali.');
      }
      final payload = error.response?.data;
      final message = payload is Map<String, dynamic> &&
              payload['error'] is Map<String, dynamic>
          ? (payload['error'] as Map<String, dynamic>)['message']
          : null;
      throw KanjiException(
        message is String
            ? message
            : 'Materi kanji tidak dapat dimuat. Periksa koneksi dan backend.',
      );
    }
  }
}
