import 'package:dio/dio.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

class HomeRepository {
  const HomeRepository(this._dio);

  final Dio _dio;

  Future<HomeDashboard> fetchDashboard() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/home');
      final body = response.data;
      final data = body?['data'];
      if (data is! Map<String, dynamic>) {
        throw const HomeException('Format data Beranda tidak valid.');
      }
      return HomeDashboard.fromJson(data);
    } on HomeException {
      rethrow;
    } on DioException catch (error) {
      throw HomeException(_message(error));
    }
  }

  String _message(DioException error) {
    final body = error.response?.data;
    if (body is Map<String, dynamic>) {
      final details = body['error'];
      if (details is Map<String, dynamic> && details['message'] is String) {
        return details['message'] as String;
      }
    }
    return 'Beranda tidak dapat dimuat. Periksa koneksi dan backend.';
  }
}

class HomeException implements Exception {
  const HomeException(this.message);

  final String message;

  @override
  String toString() => message;
}
