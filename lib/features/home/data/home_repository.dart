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
      // Uploaded and bundled server images use root-relative URLs. Resolve them
      // against the actual API host (including Android emulator host rewriting).
      final banners = data['banners'];
      final resolved = <String, dynamic>{...data};
      if (banners is List) {
        final base = Uri.parse(_dio.options.baseUrl);
        resolved['banners'] = banners.whereType<Map<String, dynamic>>().map((
          banner,
        ) {
          final image = banner['imageUrl'] as String? ?? '';
          return <String, dynamic>{
            ...banner,
            'imageUrl': image.isEmpty ? '' : base.resolve(image).toString(),
          };
        }).toList();
      }
      return HomeDashboard.fromJson(resolved);
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
