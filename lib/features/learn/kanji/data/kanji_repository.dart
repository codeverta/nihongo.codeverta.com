import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';

class KanjiException implements Exception {
  const KanjiException(this.message);
  final String message;
  @override
  String toString() => message;
}

class KanjiRepository {
  const KanjiRepository(this._dio, {OfflinePracticeService? offlinePractice})
    : _offlinePractice = offlinePractice;
  final Dio _dio;
  final OfflinePracticeService? _offlinePractice;
  Stream<OfflinePracticeSyncEvent> get syncEvents =>
      _offlinePractice?.syncEvents ??
      const Stream<OfflinePracticeSyncEvent>.empty();
  Future<KanjiCourse> getCourse({String level = 'N5'}) async =>
      KanjiCourse.fromJson(
        await _request(
          '/kanji/course?level=${Uri.encodeQueryComponent(level)}',
        ),
      );
  Future<KanjiLesson> getLesson(String id) async => KanjiLesson.fromJson(
    await _request('/kanji/lessons/${Uri.encodeComponent(id)}'),
  );
  Future<KanjiAttempt> startAttempt(String id) async {
    if (_offlinePractice != null) {
      try {
        return KanjiAttempt.fromJson(await _offlinePractice.start('kanji', id));
      } on DioException catch (error) {
        if (!_canUseLegacyOnlineQuiz(error)) rethrow;
      }
    }
    return KanjiAttempt.fromJson(
      await _request(
        '/kanji/lessons/${Uri.encodeComponent(id)}/attempts',
        post: true,
      ),
    );
  }

  Future<KanjiFeedback> answer(
    String attempt,
    String question,
    String option,
  ) async {
    if (_offlinePractice != null &&
        await _offlinePractice.ownsAttempt(attempt)) {
      return KanjiFeedback.fromJson(
        await _offlinePractice.answer(attempt, question, option),
      );
    }
    return KanjiFeedback.fromJson(
      await _request(
        '/kanji/attempts/$attempt/answers',
        post: true,
        body: {'questionId': question, 'optionId': option},
      ),
    );
  }

  Future<KanjiResult> complete(String id) async {
    if (_offlinePractice != null && await _offlinePractice.ownsAttempt(id)) {
      return KanjiResult.fromJson(await _offlinePractice.complete(id));
    }
    return KanjiResult.fromJson(
      await _request('/kanji/attempts/$id/complete', post: true),
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
      final message =
          payload is Map<String, dynamic> &&
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
