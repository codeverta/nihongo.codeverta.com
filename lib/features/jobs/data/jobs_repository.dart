import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';

final jobsRepositoryProvider = Provider(
  (ref) => JobsRepository(ref.watch(dioProvider)),
);
final jobsMetaProvider = FutureProvider.autoDispose(
  (ref) => ref.watch(jobsRepositoryProvider).meta(),
);
final jobDetailProvider = FutureProvider.autoDispose.family<JapanJob, String>(
  (ref, id) => ref.watch(jobsRepositoryProvider).detail(id),
);
final jobSourcesProvider = FutureProvider.autoDispose(
  (ref) => ref.watch(jobsRepositoryProvider).sources(),
);

class JobsRepository {
  JobsRepository(this.dio);
  final Dio dio;

  Future<Map<String, dynamic>> _get(
    String path, [
    Map<String, dynamic>? query,
  ]) async {
    try {
      final response = await dio.get<Map<String, dynamic>>(
        path,
        queryParameters: query,
      );
      return response.data!['data'] as Map<String, dynamic>;
    } on DioException catch (error) {
      throw JobsFailure(
        error.response?.statusCode == 404
            ? 'Lowongan tidak ditemukan atau sudah dihapus.'
            : 'Lowongan belum dapat dimuat. Periksa koneksi dan coba kembali.',
      );
    }
  }

  Future<JobPage> list(Map<String, String> filters, {String? cursor}) async =>
      JobPage.fromJson(
        await _get('/jobs', {
          ...filters,
          'limit': 20,
          if (cursor != null) 'cursor': cursor,
        }),
      );
  Future<JapanJob> detail(String id) async =>
      JapanJob.fromJson(await _get('/jobs/${Uri.encodeComponent(id)}'));
  Future<JobMeta> meta() async => JobMeta.fromJson(await _get('/jobs/meta'));
  Future<List<Map<String, dynamic>>> sources() async =>
      ((await _get('/jobs/sources'))['items'] as List)
          .cast<Map<String, dynamic>>();
}

class JobsFailure implements Exception {
  const JobsFailure(this.message);
  final String message;
  @override
  String toString() => message;
}
