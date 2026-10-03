import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';

class GrammarException implements Exception {
  const GrammarException(this.message);
  final String message;
  @override
  String toString() => message;
}

class GrammarRepository {
  const GrammarRepository(this._dio);
  final Dio _dio;
  Future<GrammarCourse> getCourse({String level = 'N5'}) async =>
      GrammarCourse.fromJson(
        await _request('/grammar/course?level=${Uri.encodeComponent(level)}'),
      );
  Future<GrammarLesson> getLesson(String id) async => GrammarLesson.fromJson(
        await _request('/grammar/lessons/${Uri.encodeComponent(id)}'),
      );
  Future<GrammarAttempt> startAttempt(String id) async =>
      GrammarAttempt.fromJson(
        await _request(
          '/grammar/lessons/${Uri.encodeComponent(id)}/attempts',
          post: true,
        ),
      );
  Future<GrammarFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async =>
      GrammarFeedback.fromJson(
        await _request(
          '/grammar/attempts/$attempt/answers',
          post: true,
          body: {'questionId': question, 'optionId': option},
        ),
      );
  Future<GrammarResult> complete(String id) async => GrammarResult.fromJson(
        await _request('/grammar/attempts/$id/complete', post: true),
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
        throw const GrammarException(
          'Respons grammar tidak valid. Coba muat ulang.',
        );
      }
      return data;
    } on DioException catch (error) {
      if (error.response?.statusCode == 401) {
        throw const GrammarException('Sesi berakhir. Silakan masuk kembali.');
      }
      final payload = error.response?.data;
      final message = payload is Map<String, dynamic> &&
              payload['error'] is Map<String, dynamic>
          ? (payload['error'] as Map<String, dynamic>)['message']
          : null;
      throw GrammarException(
        message is String
            ? message
            : 'Materi grammar tidak dapat dimuat. Periksa koneksi dan backend.',
      );
    }
  }
}
