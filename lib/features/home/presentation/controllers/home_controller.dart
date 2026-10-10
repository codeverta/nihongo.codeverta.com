import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/data/home_repository.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

final homeRepositoryProvider = Provider<HomeRepository>((ref) {
  return HomeRepository(ref.watch(dioProvider));
});

final homeDashboardProvider = FutureProvider.autoDispose<HomeDashboard>((ref) {
  ref.watch(authControllerProvider.select((state) => state.user?.id));
  return ref.watch(homeRepositoryProvider).fetchDashboard();
});
