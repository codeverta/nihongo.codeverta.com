import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

class GamificationSummary {
  const GamificationSummary({
    required this.xp,
    required this.level,
    required this.levelXp,
    required this.nextLevelXp,
    required this.streakDays,
    required this.bestStreak,
    required this.lessonsCompleted,
    required this.examsCompleted,
    required this.activeDays,
    required this.minutesToday,
    required this.achievements,
  });

  factory GamificationSummary.fromJson(Map<String, dynamic> json) =>
      GamificationSummary(
        xp: json['xp'] as int,
        level: json['level'] as int,
        levelXp: json['levelXp'] as int,
        nextLevelXp: json['nextLevelXp'] as int,
        streakDays: json['streakDays'] as int,
        bestStreak: json['bestStreak'] as int,
        lessonsCompleted: json['lessonsCompleted'] as int,
        examsCompleted: json['examsCompleted'] as int,
        activeDays: json['activeDays'] as int,
        minutesToday: json['minutesToday'] as int,
        achievements: (json['achievements'] as List<dynamic>)
            .map((a) => Achievement.fromJson(a as Map<String, dynamic>))
            .toList(growable: false),
      );

  final int xp;
  final int level;
  final int levelXp;
  final int nextLevelXp;
  final int streakDays;
  final int bestStreak;
  final int lessonsCompleted;
  final int examsCompleted;
  final int activeDays;
  final int minutesToday;
  final List<Achievement> achievements;
  int get unlockedCount => achievements.where((a) => a.isUnlocked).length;
}

class GamificationRepository {
  const GamificationRepository(this.dio);
  final Dio dio;

  Future<GamificationSummary> fetch() async {
    final response = await dio.get<Map<String, dynamic>>('/me/gamification');
    return GamificationSummary.fromJson(
      response.data!['data'] as Map<String, dynamic>,
    );
  }
}

final gamificationRepositoryProvider = Provider(
  (ref) => GamificationRepository(ref.watch(dioProvider)),
);
final gamificationProvider =
    FutureProvider.autoDispose<GamificationSummary>((ref) {
  final userId = ref.watch(authControllerProvider.select((s) => s.user?.id));
  if (userId == null) throw StateError('Silakan masuk untuk melihat progres.');
  return ref.watch(gamificationRepositoryProvider).fetch();
});
