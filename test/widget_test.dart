import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/main.dart';
import 'package:nihongo_master/shared/presentation/widgets/main_scaffold.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('guest can open the free dictionary tab', (tester) async {
    SharedPreferences.setMockInitialValues({'is_first_time': false});
    final preferences = await SharedPreferences.getInstance();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          homeDashboardProvider.overrideWith((ref) async => _dashboard),
        ],
        child:
            MaterialApp(theme: AppTheme.lightTheme, home: const MainScaffold()),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Kamus'));
    await tester.pumpAndSettle();
    expect(find.text('Kamus Jepang'), findsOneWidget);
    expect(find.text('Cari di kamus'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
  testWidgets('first launch redirects to onboarding', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
        ],
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text('Kuasai Bahasa Jepang'),
      findsOneWidget,
    );
  });

  testWidgets('main navigation changes the active tab', (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(
              const AuthenticationState.authenticated(
                AuthenticatedUser(
                  id: 'test-user',
                  displayName: 'Test User',
                  email: 'test@example.com',
                ),
              ),
            ),
          ),
          homeDashboardProvider.overrideWith((ref) async => _dashboard),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const MainScaffold(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Mulai belajar'), findsOneWidget);
    expect(find.text('Vocabulary API'), findsOneWidget);
    expect(find.text('Lifetime dari API'), findsOneWidget);

    await tester.tap(find.text('Belajar'));
    await tester.pumpAndSettle();

    expect(find.text('Pilih level JLPT'), findsOneWidget);

    await tester.tap(find.text('Ujian'));
    await tester.pumpAndSettle();

    expect(find.text('Latihan Ujian'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Lifetime Access').first,
      150,
    );
    expect(find.text('Lifetime Access'), findsWidgets);

    await tester.tap(find.text('Profil'));
    await tester.pumpAndSettle();

    expect(find.text('Gamification'), findsOneWidget);
  });

  testWidgets('guest is redirected to login after tapping a feature', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({'is_first_time': false});
    final preferences = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          homeDashboardProvider.overrideWith((ref) async => _dashboard),
        ],
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Mulai belajar'), findsOneWidget);
    await tester.tap(find.text('Belajar'));
    await tester.pumpAndSettle();

    expect(find.text('Masuk untuk mulai belajar'), findsOneWidget);
  });
}

const _dashboard = HomeDashboard(
  isAuthenticated: true,
  profile: HomeProfile(
    id: 'test-user',
    displayName: 'Test User',
    email: 'test@example.com',
    target: 'JLPT N3',
    level: 3,
    xp: 1200,
    streakDays: 7,
    tier: 'free',
    isLifetime: false,
  ),
  dailyGoal: HomeDailyGoal(
    minutesLearned: 12,
    targetMinutes: 20,
    streakDays: 7,
  ),
  categories: [
    LearningCategory(
      id: 'vocabulary',
      name: 'Vocabulary API',
      description: 'Kosakata',
      lessonCount: 240,
      requiredTier: 'free',
    ),
  ],
  tracks: [
    LearningTrack(
      id: 'jlpt-n5',
      system: 'JLPT',
      level: 'N5',
      title: 'JLPT N5 Foundation',
      description: 'Fondasi',
      progress: 40,
      requiredTier: 'free',
    ),
  ],
  exams: [
    HomeExam(
      id: 'exam-1',
      system: 'JLPT',
      level: 'N5',
      title: 'Exam dari API',
      questionCount: 25,
      durationMinutes: 30,
      requiredTier: 'free',
    ),
  ],
  plans: [
    SubscriptionPlan(
      id: 'lifetime',
      name: 'Lifetime dari API',
      billingType: 'lifetime',
      priceIdr: 1499000,
      features: ['Semua materi'],
      recommended: true,
    ),
  ],
  achievements: [
    Achievement(
      id: 'first',
      name: 'First Step API',
      description: 'Selesaikan pelajaran',
      icon: 'footprints',
      progress: 0,
      target: 1,
    ),
  ],
);
