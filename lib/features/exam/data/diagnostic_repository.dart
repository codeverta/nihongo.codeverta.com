import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

typedef Json = Map<String, dynamic>;

class DiagnosticRepository {
  const DiagnosticRepository(this.dio);
  final Dio dio;
  Future<Json> request(String path, {String method = 'GET', Json? body}) async {
    try {
      final response = await dio.request<Json>(
        path,
        data: body,
        options: Options(method: method),
      );
      final data = response.data?['data'];
      if (data is! Json) {
        throw Exception('Respons ujian tidak valid. Coba lagi.');
      }
      return data;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw Exception('Sesi berakhir. Silakan masuk kembali.');
      }
      if (e.response?.statusCode == 402) {
        throw Exception(
          'Lifetime Access diperlukan. Periksa status akses di halaman Premium.',
        );
      }
      if (e.response?.statusCode == 404) {
        throw Exception(
          'Sesi ujian tidak ditemukan. Kembali ke halaman Ujian untuk mulai lagi.',
        );
      }
      if (e.response?.statusCode == 422) {
        throw Exception(
          'Jawaban tidak dapat disimpan. Sesi sudah selesai atau waktunya habis.',
        );
      }
      throw Exception(
        'Ujian memerlukan koneksi internet. Hubungkan kembali internet lalu coba lagi; akun tetap masuk.',
      );
    }
  }

  Future<Json> overview() => request('/diagnostic/n5');
  Future<String> start() async =>
      (await request(
            '/exam-attempts',
            method: 'POST',
            body: {'examId': 'jlpt-n5-diagnostic'},
          ))['id']
          as String;
  Future<Json> session(String id) =>
      request('/exam-attempts/${Uri.encodeComponent(id)}');
  Future<void> answer(String id, String question, String option) async {
    await request(
      '/exam-attempts/${Uri.encodeComponent(id)}/answers',
      method: 'PUT',
      body: {'questionId': question, 'optionId': option},
    );
  }

  Future<Json> complete(String id) => request(
    '/exam-attempts/${Uri.encodeComponent(id)}/complete',
    method: 'POST',
  );
}

final diagnosticRepositoryProvider = Provider(
  (ref) => DiagnosticRepository(ref.watch(dioProvider)),
);
final diagnosticOverviewProvider = FutureProvider.autoDispose<Json>((ref) {
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  return ref.watch(diagnosticRepositoryProvider).overview();
});
List<Json> diagnosticList(dynamic value) =>
    (value as List).map((e) => Map<String, dynamic>.from(e as Map)).toList();
String diagnosticError(Object e) =>
    e.toString().replaceFirst('Exception: ', '');
