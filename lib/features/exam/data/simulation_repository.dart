import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';

class SimulationRepository extends DiagnosticRepository {
  const SimulationRepository(super.dio);

  @override
  Future<Json> overview() => request('/simulation/n5');

  @override
  Future<String> start() async => (await request(
        '/exam-attempts',
        method: 'POST',
        body: {'examId': 'jlpt-n5-simulation-1'},
      ))['id'] as String;

  @override
  Future<Json> session(String id) =>
      request('/simulation/attempts/${Uri.encodeComponent(id)}');

  Future<void> saveAnswer(
    String id,
    String question,
    String option, {
    required bool marked,
  }) async {
    await request(
      '/simulation/attempts/${Uri.encodeComponent(id)}/answers',
      method: 'PUT',
      body: {'questionId': question, 'optionId': option, 'marked': marked},
    );
  }

  Future<Json> submitSection(String id, String section) => request(
        '/simulation/attempts/${Uri.encodeComponent(id)}/submit-section',
        method: 'POST',
        body: {'section': section},
      );
}

final simulationRepositoryProvider = Provider(
  (ref) => SimulationRepository(ref.watch(dioProvider)),
);

final simulationOverviewProvider = FutureProvider.autoDispose<Json>((ref) {
  ref.watch(authControllerProvider.select((s) => s.user?.id));
  return ref.watch(simulationRepositoryProvider).overview();
});
