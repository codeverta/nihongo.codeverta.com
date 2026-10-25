import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';

class ReadingException implements Exception {
  const ReadingException(this.message);
  final String message;
  @override
  String toString() => message;
}

class ReadingRepository {
  const ReadingRepository(this._dio, {OfflinePracticeService? offlinePractice})
    : _offlinePractice = offlinePractice;
  final Dio _dio;
  final OfflinePracticeService? _offlinePractice;
  Stream<OfflinePracticeSyncEvent> get syncEvents =>
      _offlinePractice?.syncEvents ??
      const Stream<OfflinePracticeSyncEvent>.empty();
  Future<ReadingCourse> getCourse({String level = 'N5'}) async =>
      ReadingCourse.fromJson(
        await _request(
          '/reading/course?level=${Uri.encodeQueryComponent(level)}',
        ),
      );
  Future<ReadingLesson> getLesson(String id) async => ReadingLesson.fromJson(
    await _request('/reading/lessons/${Uri.encodeComponent(id)}'),
  );
  Future<ReadingAttempt> startAttempt(String id) async {
    if (_offlinePractice != null) {
      try {
        return ReadingAttempt.fromJson(
          await _offlinePractice.start('reading', id),
        );
      } on DioException catch (error) {
        if (!_canUseLegacyOnlineQuiz(error)) rethrow;
      }
    }
    return ReadingAttempt.fromJson(
      await _request(
        '/reading/lessons/${Uri.encodeComponent(id)}/attempts',
        post: true,
      ),
    );
  }

  Future<ReadingFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async {
    if (_offlinePractice != null &&
        await _offlinePractice.ownsAttempt(attempt)) {
      return ReadingFeedback.fromJson(
        await _offlinePractice.answer(attempt, question, option),
      );
    }
    return ReadingFeedback.fromJson(
      await _request(
        '/reading/attempts/$attempt/answers',
        post: true,
        body: {'questionId': question, 'optionId': option},
      ),
    );
  }

  Future<ReadingResult> complete(String id) async {
    if (_offlinePractice != null && await _offlinePractice.ownsAttempt(id)) {
      return ReadingResult.fromJson(await _offlinePractice.complete(id));
    }
    return ReadingResult.fromJson(
      await _request('/reading/attempts/$id/complete', post: true),
    );
  }

  bool _canUseLegacyOnlineQuiz(DioException error) {
    final status = error.response?.statusCode;
    return status == 404 || status == 503;
  }

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
      final message =
          payload is Map<String, dynamic> &&
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
