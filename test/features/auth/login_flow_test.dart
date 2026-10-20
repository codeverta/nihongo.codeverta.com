import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/main.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('email login redirects to Beranda after success', (tester) async {
    final harness = await _createContainer();
    final container = harness.container;
    final repository = harness.repository;
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Belajar'));
    await tester.pumpAndSettle();
    expect(find.text('Masuk untuk mulai belajar'), findsOneWidget);

    await tester.enterText(
      find.byType(TextFormField).at(0),
      'user@example.com',
    );
    await tester.enterText(find.byType(TextFormField).at(1), 'password123');
    await tester.tap(find.widgetWithText(ElevatedButton, 'Masuk'));
    await tester.pumpAndSettle();

    expect(repository.email, 'user@example.com');
    expect(repository.password, 'password123');
    expect(container.read(authControllerProvider).isAuthenticated, isTrue);
    expect(find.text('Masuk untuk mulai belajar'), findsNothing);
    expect(find.byKey(const Key('home-greeting')), findsOneWidget);
  });

  testWidgets('failed email login stays on login and shows the error', (
    tester,
  ) async {
    final harness = await _createContainer(
      failure: const AuthException('Email atau password salah.'),
    );
    final container = harness.container;
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const NihongoMasterApp(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Belajar'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byType(TextFormField).at(0),
      'user@example.com',
    );
    await tester.enterText(find.byType(TextFormField).at(1), 'wrongpass');
    await tester.tap(find.widgetWithText(ElevatedButton, 'Masuk'));
    await tester.pumpAndSettle();

    expect(find.text('Masuk untuk mulai belajar'), findsOneWidget);
    expect(find.text('Email atau password salah.'), findsOneWidget);
    expect(container.read(authControllerProvider).isAuthenticated, isFalse);
  });
}

Future<({ProviderContainer container, _FakeAuthRepository repository})>
_createContainer({AuthException? failure}) async {
  SharedPreferences.setMockInitialValues({'is_first_time': false});
  final preferences = await SharedPreferences.getInstance();
  final repository = _FakeAuthRepository(preferences, failure: failure);
  final authController = AuthController(repository, null);
  await authController.restoreSession();

  return (
    container: ProviderContainer(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(preferences),
        authControllerProvider.overrideWith((ref) => authController),
        homeDashboardProvider.overrideWith((ref) async => _dashboard),
        newsPreviewProvider.overrideWith(
          (ref) async =>
              const NewsPage(items: [], total: 0, page: 1, sources: []),
        ),
      ],
    ),
    repository: repository,
  );
}

class _FakeAuthRepository extends AuthRepository {
  _FakeAuthRepository(SharedPreferences preferences, {this.failure})
    : super(Dio(), preferences);

  final AuthException? failure;
  String? email;
  String? password;

  @override
  bool get hasStoredSession => false;

  @override
  Future<AuthenticatedUser> loginWithEmail(
    String email,
    String password,
  ) async {
    this.email = email;
    this.password = password;
    final error = failure;
    if (error != null) throw error;
    return const AuthenticatedUser(
      id: 'test-user',
      displayName: 'Test User',
      email: 'user@example.com',
    );
  }
}

const _dashboard = HomeDashboard(
  isAuthenticated: true,
  profile: HomeProfile(
    id: 'test-user',
    displayName: 'Test User',
    email: 'user@example.com',
    target: 'JLPT N3',
    level: 3,
    xp: 0,
    streakDays: 0,
    tier: 'free',
    isLifetime: false,
  ),
  dailyGoal: HomeDailyGoal(minutesLearned: 0, targetMinutes: 20, streakDays: 0),
  categories: [],
  tracks: [],
  exams: [],
  plans: [],
  achievements: [],
);
