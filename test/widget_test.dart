import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_banner_carousel.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
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
          newsPreviewProvider.overrideWith(
            (ref) async =>
                const NewsPage(items: [], total: 0, page: 1, sources: []),
          ),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const MainScaffold(),
        ),
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
        overrides: [sharedPreferencesProvider.overrideWithValue(preferences)],
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Kuasai Bahasa Jepang'), findsOneWidget);
  });

  testWidgets('main navigation changes the active tab', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
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
          newsPreviewProvider.overrideWith(
            (ref) async =>
                const NewsPage(items: [], total: 0, page: 1, sources: []),
          ),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const MainScaffold(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Lifetime dari API'), findsOneWidget);
    expect(
      find.ancestor(
        of: find.text('Lifetime dari API'),
        matching: find.byType(HomeBannerCarousel),
      ),
      findsOneWidget,
    );
    expect(find.text('Target harian'), findsNothing);
    final homeScroll = find
        .descendant(
          of: find.byKey(const PageStorageKey('dashboard-scroll')),
          matching: find.byType(Scrollable),
        )
        .first;
    await tester.scrollUntilVisible(
      find.text('Vocabulary API'),
      150,
      scrollable: homeScroll,
    );
    expect(find.text('Vocabulary API'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Kerja di Jepang'),
      150,
      scrollable: homeScroll,
    );
    expect(find.text('Kerja di Jepang'), findsOneWidget);

    await tester.tap(find.text('Belajar'));
    await tester.pumpAndSettle();

    expect(find.text('Pilih level JLPT'), findsOneWidget);
    expect(find.byKey(const Key('open-manga')), findsNothing);
    expect(find.text('Anki · Flashcards'), findsNothing);
    expect(find.text('SSW / Tokutei Ginou'), findsNothing);
    expect(find.text('Kerja di Jepang'), findsNothing);

    await tester.tap(find.text('Ujian'));
    await tester.pumpAndSettle();

    expect(find.text('Latihan Ujian'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('Lifetime Access').first, 150);
    expect(find.text('Lifetime Access'), findsWidgets);

    await tester.tap(find.text('Profil'));
    await tester.pumpAndSettle();

    expect(find.text('Gamification'), findsOneWidget);
  });

  testWidgets('home premium offer is hidden for lifetime subscribers', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({});
    final preferences = await SharedPreferences.getInstance();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(
              const AuthenticationState.authenticated(
                AuthenticatedUser(
                  id: 'lifetime-user',
                  displayName: 'Lifetime User',
                  email: 'lifetime@example.com',
                  isLifetime: true,
                ),
              ),
            ),
          ),
          homeDashboardProvider.overrideWith((ref) async => _dashboard),
          newsPreviewProvider.overrideWith(
            (ref) async =>
                const NewsPage(items: [], total: 0, page: 1, sources: []),
          ),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const MainScaffold(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('NIHONGO PREMIUM'), findsNothing);
    expect(find.text('Lihat paket'), findsNothing);
  });

  testWidgets(
    'verified purchase hides home and profile offers without restart',
    (tester) async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final auth = _PurchaseAuthController();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPreferencesProvider.overrideWithValue(preferences),
            authControllerProvider.overrideWith((ref) => auth),
            homeDashboardProvider.overrideWith((ref) async => _dashboard),
            newsPreviewProvider.overrideWith(
              (ref) async =>
                  const NewsPage(items: [], total: 0, page: 1, sources: []),
            ),
            gamificationProvider.overrideWith((ref) async => _gamification),
          ],
          child: MaterialApp(
            theme: AppTheme.lightTheme,
            home: const MainScaffold(),
          ),
        ),
      );
      await tester.pumpAndSettle();
      var carousel = tester.widget<HomeBannerCarousel>(
        find.byType(HomeBannerCarousel),
      );
      expect(
        carousel.banners.where((banner) => banner.action == 'premium'),
        hasLength(1),
      );
      expect(carousel.banners.first.description, contains('Rp1.499.000'));
      await tester.tap(find.text('Profil'));
      await tester.pumpAndSettle();
      final profileScroll = find
          .descendant(
            of: find.byKey(const PageStorageKey('profile-scroll')),
            matching: find.byType(Scrollable),
          )
          .first;
      await tester.scrollUntilVisible(
        find.text('Lifetime Access'),
        150,
        scrollable: profileScroll,
      );
      expect(find.text('Lifetime Access'), findsOneWidget);

      auth.subscribe();
      await tester.pumpAndSettle();
      expect(find.text('Lifetime Access'), findsNothing);
      await tester.tap(find.text('Beranda'));
      await tester.pumpAndSettle();
      carousel = tester.widget<HomeBannerCarousel>(
        find.byType(HomeBannerCarousel),
      );
      expect(
        carousel.banners.where((banner) => banner.action == 'premium'),
        isEmpty,
      );
      expect(carousel.banners.map((banner) => banner.id), [
        'learn',
        'ssw',
        'jobs',
        'anki',
        'manga-reading',
      ]);
      expect(find.text('Lifetime dari API'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

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
          newsPreviewProvider.overrideWith(
            (ref) async =>
                const NewsPage(items: [], total: 0, page: 1, sources: []),
          ),
        ],
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('home-greeting')), findsOneWidget);
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
  banners: [
    HomeBanner(
      id: 'learn',
      title: 'Materi belajar',
      description: 'Pelajari bahasa Jepang.',
      imageUrl: '',
      action: 'learn',
      buttonLabel: 'Buka materi',
    ),
  ],
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

class _PurchaseAuthController extends AuthController {
  _PurchaseAuthController()
    : super.test(
        const AuthenticationState.authenticated(
          AuthenticatedUser(
            id: 'test-user',
            displayName: 'Test User',
            email: 'test@example.com',
          ),
        ),
      );
  void subscribe() {
    state = const AuthenticationState.authenticated(
      AuthenticatedUser(
        id: 'test-user',
        displayName: 'Test User',
        email: 'test@example.com',
        isLifetime: true,
      ),
    );
  }
}

const _gamification = GamificationSummary(
  xp: 0,
  level: 1,
  levelXp: 0,
  nextLevelXp: 100,
  streakDays: 0,
  bestStreak: 0,
  lessonsCompleted: 0,
  examsCompleted: 0,
  activeDays: 0,
  minutesToday: 0,
  achievements: [],
);
