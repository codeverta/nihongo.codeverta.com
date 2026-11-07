import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';

class SimulationRepository extends DiagnosticRepository {
  const SimulationRepository(super.dio, {this.level = 'N5'});
  final String level;

  @override
  Future<Json> overview() => request('/simulation/${level.toLowerCase()}');

  @override
  Future<String> start() async =>
      (await request(
            '/exam-attempts',
            method: 'POST',
            body: {
              'examId': level.startsWith('SSW-')
                  ? '${level.toLowerCase()}-simulation-1'
                  : level == 'JFT'
                  ? 'jft-a2-full-1'
                  : level.startsWith('JFT-')
                  ? 'jft-a2-full-${level.substring(4)}'
                  : level.startsWith('NAT')
                  ? 'nat-5q-full-${level == 'NAT' ? '1' : level.substring(4)}'
                  : RegExp(r'^N[1-5]-\d+$').hasMatch(level)
                  ? 'jlpt-${level.toLowerCase().split('-').first}-simulation-${level.split('-').last}'
                  : 'jlpt-${level.toLowerCase()}-simulation-1',
            },
          ))['id']
          as String;

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

final simulationLevelRepositoryProvider =
    Provider.family<SimulationRepository, String>((ref, level) {
      if (level == 'N5') return ref.watch(simulationRepositoryProvider);
      return SimulationRepository(ref.watch(dioProvider), level: level);
    });
final simulationLevelOverviewProvider = FutureProvider.autoDispose
    .family<Json, String>((ref, level) {
      ref.watch(authControllerProvider.select((s) => s.user?.id));
      return ref.watch(simulationLevelRepositoryProvider(level)).overview();
    });
