import 'package:flutter/foundation.dart';

@immutable
class HomeDashboard {
  const HomeDashboard({
    required this.isAuthenticated,
    required this.dailyGoal,
    required this.categories,
    required this.tracks,
    required this.exams,
    required this.plans,
    required this.achievements,
    this.profile,
  });

  factory HomeDashboard.fromJson(Map<String, dynamic> json) {
    return HomeDashboard(
      isAuthenticated: json['isAuthenticated'] as bool? ?? false,
      profile: json['profile'] is Map<String, dynamic>
          ? HomeProfile.fromJson(json['profile'] as Map<String, dynamic>)
          : null,
      dailyGoal: HomeDailyGoal.fromJson(_map(json['dailyGoal'])),
      categories: _list(json['categories'], LearningCategory.fromJson),
      tracks: _list(json['tracks'], LearningTrack.fromJson),
      exams: _list(json['exams'], HomeExam.fromJson),
      plans: _list(json['plans'], SubscriptionPlan.fromJson),
      achievements: _list(json['achievements'], Achievement.fromJson),
    );
  }

  final bool isAuthenticated;
  final HomeProfile? profile;
  final HomeDailyGoal dailyGoal;
  final List<LearningCategory> categories;
  final List<LearningTrack> tracks;
  final List<HomeExam> exams;
  final List<SubscriptionPlan> plans;
  final List<Achievement> achievements;
}

@immutable
class HomeProfile {
  const HomeProfile({
    required this.id,
    required this.displayName,
    required this.email,
    required this.target,
    required this.level,
    required this.xp,
    required this.streakDays,
    required this.tier,
    required this.isLifetime,
    this.photoUrl,
  });

  factory HomeProfile.fromJson(Map<String, dynamic> json) {
    final entitlement = _map(json['entitlement']);
    return HomeProfile(
      id: json['id'] as String? ?? '',
      displayName: json['displayName'] as String? ?? '',
      email: json['email'] as String? ?? '',
      photoUrl: json['photoUrl'] as String?,
      target: json['target'] as String? ?? '',
      level: json['level'] as int? ?? 1,
      xp: json['xp'] as int? ?? 0,
      streakDays: json['streakDays'] as int? ?? 0,
      tier: entitlement['tier'] as String? ?? 'free',
      isLifetime: entitlement['isLifetime'] as bool? ?? false,
    );
  }

  final String id;
  final String displayName;
  final String email;
  final String? photoUrl;
  final String target;
  final int level;
  final int xp;
  final int streakDays;
  final String tier;
  final bool isLifetime;
}

@immutable
class HomeDailyGoal {
  const HomeDailyGoal({
    required this.minutesLearned,
    required this.targetMinutes,
    required this.streakDays,
  });

  factory HomeDailyGoal.fromJson(Map<String, dynamic> json) => HomeDailyGoal(
        minutesLearned: json['minutesLearned'] as int? ?? 0,
        targetMinutes: json['targetMinutes'] as int? ?? 0,
        streakDays: json['streakDays'] as int? ?? 0,
      );

  final int minutesLearned;
  final int targetMinutes;
  final int streakDays;

  double get completion => targetMinutes == 0
      ? 0
      : (minutesLearned / targetMinutes).clamp(0, 1).toDouble();
}

@immutable
class LearningCategory {
  const LearningCategory({
    required this.id,
    required this.name,
    required this.description,
    required this.lessonCount,
    required this.requiredTier,
  });

  factory LearningCategory.fromJson(Map<String, dynamic> json) =>
      LearningCategory(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        lessonCount: json['lessonCount'] as int? ?? 0,
        requiredTier: json['requiredTier'] as String? ?? 'free',
      );

  final String id;
  final String name;
  final String description;
  final int lessonCount;
  final String requiredTier;
  bool get isPremium => requiredTier == 'premium';
}

@immutable
class LearningTrack {
  const LearningTrack({
    required this.id,
    required this.system,
    required this.level,
    required this.title,
    required this.description,
    required this.progress,
    required this.requiredTier,
  });

  factory LearningTrack.fromJson(Map<String, dynamic> json) => LearningTrack(
        id: json['id'] as String? ?? '',
        system: json['system'] as String? ?? '',
        level: json['level'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        progress: json['progress'] as int? ?? 0,
        requiredTier: json['requiredTier'] as String? ?? 'free',
      );

  final String id;
  final String system;
  final String level;
  final String title;
  final String description;
  final int progress;
  final String requiredTier;
  bool get isPremium => requiredTier == 'premium';
}

@immutable
class HomeExam {
  const HomeExam({
    required this.id,
    required this.system,
    required this.level,
    required this.title,
    required this.questionCount,
    required this.durationMinutes,
    required this.requiredTier,
  });

  factory HomeExam.fromJson(Map<String, dynamic> json) => HomeExam(
        id: json['id'] as String? ?? '',
        system: json['system'] as String? ?? '',
        level: json['level'] as String? ?? '',
        title: json['title'] as String? ?? '',
        questionCount: json['questionCount'] as int? ?? 0,
        durationMinutes: json['durationMinutes'] as int? ?? 0,
        requiredTier: json['requiredTier'] as String? ?? 'free',
      );

  final String id;
  final String system;
  final String level;
  final String title;
  final int questionCount;
  final int durationMinutes;
  final String requiredTier;
  bool get isPremium => requiredTier == 'premium';
}

@immutable
class SubscriptionPlan {
  const SubscriptionPlan({
    required this.id,
    required this.name,
    required this.billingType,
    required this.priceIdr,
    required this.features,
    required this.recommended,
  });

  factory SubscriptionPlan.fromJson(Map<String, dynamic> json) =>
      SubscriptionPlan(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        billingType: json['billingType'] as String? ?? '',
        priceIdr: json['priceIdr'] as int? ?? 0,
        features: (json['features'] as List<dynamic>? ?? const [])
            .whereType<String>()
            .toList(growable: false),
        recommended: json['recommended'] as bool? ?? false,
      );

  final String id;
  final String name;
  final String billingType;
  final int priceIdr;
  final List<String> features;
  final bool recommended;
}

@immutable
class Achievement {
  const Achievement({
    required this.id,
    required this.name,
    required this.description,
    required this.icon,
    required this.progress,
    required this.target,
    this.unlockedAt,
  });

  factory Achievement.fromJson(Map<String, dynamic> json) => Achievement(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        description: json['description'] as String? ?? '',
        icon: json['icon'] as String? ?? '',
        progress: json['progress'] as int? ?? 0,
        target: json['target'] as int? ?? 0,
        unlockedAt: json['unlockedAt'] == null
            ? null
            : DateTime.parse(json['unlockedAt'] as String),
      );

  final String id;
  final String name;
  final String description;
  final String icon;
  final int progress;
  final int target;
  final DateTime? unlockedAt;
  bool get isUnlocked => unlockedAt != null;
}

Map<String, dynamic> _map(Object? value) =>
    value is Map<String, dynamic> ? value : const <String, dynamic>{};

List<T> _list<T>(Object? value, T Function(Map<String, dynamic>) fromJson) {
  return (value as List<dynamic>? ?? const [])
      .whereType<Map<String, dynamic>>()
      .map(fromJson)
      .toList(growable: false);
}
