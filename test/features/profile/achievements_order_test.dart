import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/profile/presentation/screens/achievements_screen.dart';

Achievement _a(String id, int progress, int target, {DateTime? at}) =>
    Achievement(
      id: id,
      name: 'Name $id',
      description: 'Desc $id',
      icon: 'trophy',
      progress: progress,
      target: target,
      unlockedAt: at,
    );

class _Many extends GamificationRepository {
  _Many(this.items) : super(Dio());
  final List<Achievement> items;

  @override
  Future<GamificationSummary> fetch() async => GamificationSummary(
    xp: 600,
    level: 2,
    levelXp: 100,
    nextLevelXp: 500,
    streakDays: 3,
    bestStreak: 7,
    lessonsCompleted: 12,
    examsCompleted: 1,
    activeDays: 8,
    minutesToday: 4,
    achievements: items,
  );
}

void main() {
  test('earned first (newest on top), then locked closest to done', () {
    final ordered = orderAchievements([
      _a('far', 1, 100),
      _a('old', 5, 5, at: DateTime.utc(2026)),
      _a('near', 9, 10),
      _a('new', 5, 5, at: DateTime.utc(2026, 5)),
      _a('mid', 5, 10),
      _a('mid-b', 50, 100),
      _a('zero-target', 0, 0),
    ]);
    expect(ordered.map((a) => a.id), [
      'new',
      'old',
      'near',
      'mid',
      'mid-b',
      'far',
      'zero-target',
    ]);
  });

  testWidgets('a long list of 52 achievements works on a small phone', (
    tester,
  ) async {
    tester.view
      ..physicalSize = const Size(320, 640)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    final items = [
      for (var i = 0; i < 52; i++)
        _a('a$i', i % 4, 10, at: i == 51 ? DateTime.utc(2026, 11) : null),
    ];
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(
              const AuthenticationState.authenticated(
                AuthenticatedUser(
                  id: 'u',
                  displayName: 'U',
                  email: 'u@example.com',
                ),
              ),
            ),
          ),
          gamificationRepositoryProvider.overrideWithValue(_Many(items)),
        ],
        child: const MaterialApp(
          locale: Locale('en'),
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          home: AchievementsScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    final list = find.byType(Scrollable).first;
    await tester.scrollUntilVisible(
      find.text('Name a51'),
      200,
      scrollable: list,
    );
    expect(find.text('1/52 achievements unlocked'), findsOneWidget);
    // Ordering itself is covered by the unit test above; here only the long list.
    await tester.scrollUntilVisible(
      find.text('Name a3'),
      200,
      scrollable: list,
    );
    expect(tester.takeException(), isNull);
    await tester.scrollUntilVisible(
      find.text('Locked'),
      -300,
      scrollable: list,
    );
    await tester.tap(find.text('Locked'));
    await tester.pumpAndSettle();
    expect(find.text('Name a51'), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
