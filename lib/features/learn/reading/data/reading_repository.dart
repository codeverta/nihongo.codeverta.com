import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';

class ReadingException implements Exception {
  const ReadingException(this.message);
  final String message;
  @override
  String toString() => message;
}

class ReadingRepository {
  const ReadingRepository(this._dio);
  final Dio _dio;
  Future<ReadingCourse> getCourse({String level = 'N5'}) async =>
      ReadingCourse.fromJson(
        await _request(
          '/reading/course?level=${Uri.encodeQueryComponent(level)}',
        ),
      );
  Future<ReadingLesson> getLesson(String id) async => ReadingLesson.fromJson(
        await _request('/reading/lessons/${Uri.encodeComponent(id)}'),
      );
  Future<ReadingAttempt> startAttempt(String id) async =>
      ReadingAttempt.fromJson(
        await _request(
          '/reading/lessons/${Uri.encodeComponent(id)}/attempts',
          post: true,
        ),
      );
  Future<ReadingFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async =>
      ReadingFeedback.fromJson(
        await _request(
          '/reading/attempts/$attempt/answers',
          post: true,
          body: {'questionId': question, 'optionId': option},
        ),
      );
  Future<ReadingResult> complete(String id) async => ReadingResult.fromJson(
        await _request('/reading/attempts/$id/complete', post: true),
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
        throw const ReadingException(
          'Respons reading tidak valid. Coba muat ulang.',
        );
      }
      return data;
    } on DioException catch (error) {
      if (error.response?.statusCode == 401) {
        throw const ReadingException('Sesi berakhir. Silakan masuk kembali.');
      }
      final payload = error.response?.data;
      final message = payload is Map<String, dynamic> &&
              payload['error'] is Map<String, dynamic>
          ? (payload['error'] as Map<String, dynamic>)['message']
          : null;
      throw ReadingException(
        message is String
            ? message
            : 'Materi reading tidak dapat dimuat. Periksa koneksi dan backend.',
      );
    }
  }
}
