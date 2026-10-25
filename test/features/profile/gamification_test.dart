import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/profile/presentation/screens/achievements_screen.dart';
import 'package:nihongo_master/features/profile/presentation/screens/profile_screen.dart';
import 'package:nihongo_master/features/profile/presentation/widgets/gamification_widgets.dart';

const user = AuthenticationState.authenticated(
  AuthenticatedUser(
    id: 'learner',
    displayName: 'Learner',
    email: 'learner@example.com',
  ),
);

Map<String, dynamic> summaryJson({int xp = 550}) => {
  'xp': xp,
  'level': 1 + xp ~/ 500,
  'levelXp': xp % 500,
  'nextLevelXp': 500,
  'streakDays': 2,
  'bestStreak': 7,
  'lessonsCompleted': 9,
  'examsCompleted': 1,
  'activeDays': 8,
  'minutesToday': 3,
  'achievements': [
    {
      'id': 'first-lesson',
      'name': 'First Step',
      'description': 'Lulus satu lesson',
      'icon': 'footprints',
      'progress': 1,
      'target': 1,
      'unlockedAt': '2026-10-09T18:00:00Z',
    },
    {
      'id': 'lessons-10',
      'name': 'Pelajar Tekun',
      'description': 'Lulus 10 lesson berbeda',
      'icon': 'book',
      'progress': 9,
      'target': 10,
    },
  ],
};

class FakeGamification extends GamificationRepository {
  FakeGamification() : super(Dio());
  int calls = 0;
  bool fail = false;
  @override
  Future<GamificationSummary> fetch() async {
    calls++;
    if (fail) throw Exception('offline');
    return GamificationSummary.fromJson(summaryJson());
  }
}

class SwitchableAuth extends AuthController {
  SwitchableAuth() : super.test(user);
  void switchUser(AuthenticatedUser next) {
    state = AuthenticationState.authenticated(next);
  }
}

void main() {
  setUpAll(() async => initializeDateFormatting());

  testWidgets('achievement date follows the English device region', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('en', 'US'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: Scaffold(
          body: AchievementTile(
            achievement: Achievement(
              id: 'first-lesson',
              name: 'First Step',
              description: 'Pass one lesson',
              icon: 'footprints',
              progress: 1,
              target: 1,
              unlockedAt: DateTime(2026, 10, 9, 12),
            ),
          ),
        ),
      ),
    );

    expect(find.text('Earned 10/9/2026'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  test('summary refreshes when switching between signed-in accounts', () async {
    final auth = SwitchableAuth();
    final repository = FakeGamification();
    final container = ProviderContainer(
      overrides: [
        authControllerProvider.overrideWith((ref) => auth),
        gamificationRepositoryProvider.overrideWithValue(repository),
      ],
    );
    addTearDown(container.dispose);
    final subscription = container.listen(gamificationProvider, (_, __) {});
    addTearDown(subscription.close);
    await container.read(gamificationProvider.future);
    auth.switchUser(
      const AuthenticatedUser(
        id: 'other',
        displayName: 'Other',
        email: 'other@example.com',
      ),
    );
    await container.read(gamificationProvider.future);
    expect(repository.calls, 2);
  });
  test(
    'repository reads authenticated summary and preserves unlock timestamp',
    () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            expect(options.path, '/me/gamification');
            handler.resolve(
              Response<Map<String, dynamic>>(
                requestOptions: options,
                statusCode: 200,
                data: {'data': summaryJson()},
              ),
            );
          },
        ),
      );
      final result = await GamificationRepository(dio).fetch();
      expect(result.xp, 550);
      expect(result.level, 2);
      expect(result.unlockedCount, 1);
      expect(result.achievements[0].unlockedAt, DateTime.utc(2026, 10, 9, 18));
      expect(result.achievements[1].isUnlocked, isFalse);
    },
  );

  testWidgets(
    'profile shows real stats and opens filtered achievements on phone',
    (tester) async {
      tester.view.physicalSize = const Size(360, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = FakeGamification();
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) => const Scaffold(body: ProfileScreen()),
          ),
          GoRoute(
            path: '/achievements',
            builder: (_, __) => const AchievementsScreen(),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(user),
            ),
            gamificationRepositoryProvider.overrideWithValue(repository),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('550'), findsOneWidget);
      expect(find.text('Level 2 · 50/500 XP'), findsOneWidget);
      expect(find.text('2.480'), findsNothing);
      await tester.scrollUntilVisible(
        find.text('Lihat semua achievement (1/2)'),
        180,
      );
      await tester.ensureVisible(find.text('Lihat semua achievement (1/2)'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Lihat semua achievement (1/2)'));
      await tester.pumpAndSettle();
      expect(find.text('Achievements'), findsOneWidget);
      await tester.scrollUntilVisible(
        find.widgetWithText(ChoiceChip, 'Terkunci'),
        100,
      );
      await tester.ensureVisible(find.widgetWithText(ChoiceChip, 'Terkunci'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ChoiceChip, 'Terkunci'));
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(find.text('Pelajar Tekun'), 100);
      expect(find.text('First Step'), findsNothing);
      expect(find.text('9/10 · Terkunci'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox());
    },
  );

  testWidgets('failed fetch shows retry and does not invent zero progress', (
    tester,
  ) async {
    final repository = FakeGamification()..fail = true;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(user),
          ),
          gamificationRepositoryProvider.overrideWithValue(repository),
        ],
        child: const MaterialApp(home: AchievementsScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.textContaining('Progres belum bisa dimuat'), findsOneWidget);
    expect(find.text('Level 1 · 0/500 XP'), findsNothing);
    repository.fail = false;
    await tester.tap(find.text('Coba lagi'));
    await tester.pumpAndSettle();
    expect(repository.calls, 2);
    expect(find.text('550'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
  });

  testWidgets('guest cannot see another account achievements', (tester) async {
    final repository = FakeGamification();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          gamificationRepositoryProvider.overrideWithValue(repository),
        ],
        child: const MaterialApp(home: AchievementsScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(repository.calls, 0);
    expect(find.text('Masuk untuk melihat achievements'), findsOneWidget);
    await tester.pumpWidget(const SizedBox());
  });
}
