import 'package:dio/dio.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';

class VocabularyRepository {
  const VocabularyRepository(this._dio);

  final Dio _dio;

  Future<List<VocabularyLevel>> getLevels() async {
    final data = await _getList('/vocabulary/levels');
    return data.map(VocabularyLevel.fromJson).toList(growable: false);
  }

  Future<VocabularyLesson> getLesson(String lessonId) async {
    return VocabularyLesson.fromJson(
      await _get('/vocabulary/lessons/$lessonId'),
    );
  }

  Future<VocabularyAttempt> startAttempt(String lessonId) async {
    return VocabularyAttempt.fromJson(
      await _post('/vocabulary/lessons/$lessonId/attempts'),
    );
  }

  Future<AnswerFeedback> answer({
    required String attemptId,
    required String questionId,
    required String optionId,
  }) async {
    return AnswerFeedback.fromJson(
      await _post(
        '/vocabulary/attempts/$attemptId/answers',
        body: {'questionId': questionId, 'optionId': optionId},
      ),
    );
  }

  Future<VocabularyResult> complete(String attemptId) async {
    return VocabularyResult.fromJson(
      await _post('/vocabulary/attempts/$attemptId/complete'),
    );
  }

  Future<Map<String, dynamic>> _get(String path) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(path);
      return _data(response.data);
    } on DioException catch (error) {
      throw VocabularyException(_message(error));
    }
  }

  Future<List<Map<String, dynamic>>> _getList(String path) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(path);
      final data = response.data?['data'];
      if (data is! List<dynamic>) throw const VocabularyException();
      return data.whereType<Map<String, dynamic>>().toList(growable: false);
    } on DioException catch (error) {
      throw VocabularyException(_message(error));
    }
  }

  Future<Map<String, dynamic>> _post(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(path, data: body);
      return _data(response.data);
    } on DioException catch (error) {
      throw VocabularyException(_message(error));
    }
  }

  Map<String, dynamic> _data(Map<String, dynamic>? body) {
    final data = body?['data'];
    if (data is! Map<String, dynamic>) throw const VocabularyException();
    return data;
  }

  String _message(DioException error) {
    final body = error.response?.data;
    if (body is Map<String, dynamic> && body['error'] is Map<String, dynamic>) {
      final message = (body['error'] as Map<String, dynamic>)['message'];
      if (message is String) return message;
    }
    return 'Materi vocabulary tidak dapat dimuat. Periksa koneksi Anda.';
  }
}

class VocabularyException implements Exception {
  const VocabularyException([
    this.message = 'Respons vocabulary dari server tidak valid.',
  ]);

  final String message;

  @override
  String toString() => message;
}
