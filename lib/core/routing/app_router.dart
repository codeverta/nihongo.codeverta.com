import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/screens/login_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_course_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_lesson_screen.dart';
import 'package:nihongo_master/features/onboarding/presentation/screens/onboarding_screen.dart';
import 'package:nihongo_master/shared/presentation/widgets/main_scaffold.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Nama dan path route dipusatkan agar navigasi tidak menggunakan magic string.
abstract final class AppRoutes {
  static const splashName = 'splash';
  static const splashPath = '/';

  static const onboardingName = 'onboarding';
  static const onboardingPath = '/onboarding';

  static const loginName = 'login';
  static const loginPath = '/login';

  static const homeName = 'home';
  static const homePath = '/home';

  static const vocabularyName = 'vocabulary';
  static const vocabularyPath = '/learn/vocabulary';

  static const vocabularyLessonName = 'vocabulary-lesson';
  static const vocabularyLessonRoute = '/learn/vocabulary/lessons/:lessonId';

  static String vocabularyLessonPath(String lessonId) =>
      '/learn/vocabulary/lessons/${Uri.encodeComponent(lessonId)}';
}

const _firstTimeKey = 'is_first_time';

/// Menyimpan status onboarding dan menyinkronkannya ke SharedPreferences.
class FirstTimeNotifier extends StateNotifier<bool> {
  FirstTimeNotifier(this._preferences)
      : super(_preferences.getBool(_firstTimeKey) ?? true);

  final SharedPreferences _preferences;

  Future<void> completeOnboarding() async {
    state = false;
    await _preferences.setBool(_firstTimeKey, false);
  }

  Future<void> resetOnboarding() async {
    state = true;
    await _preferences.setBool(_firstTimeKey, true);
  }
}

final isFirstTimeProvider =
    StateNotifierProvider<FirstTimeNotifier, bool>((ref) {
  return FirstTimeNotifier(ref.watch(sharedPreferencesProvider));
});

/// Satu instance router untuk seluruh lifecycle aplikasi.
final appRouterProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    initialLocation: AppRoutes.splashPath,
    routes: [
      GoRoute(
        path: AppRoutes.splashPath,
        name: AppRoutes.splashName,
        builder: (context, state) => const _SplashScreen(),
      ),
      GoRoute(
        path: AppRoutes.onboardingPath,
        name: AppRoutes.onboardingName,
        builder: (context, state) => OnboardingScreen(
          onCompleted: () =>
              ref.read(isFirstTimeProvider.notifier).completeOnboarding(),
        ),
      ),
      GoRoute(
        path: AppRoutes.loginPath,
        name: AppRoutes.loginName,
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: AppRoutes.homePath,
        name: AppRoutes.homeName,
        builder: (context, state) => const MainScaffold(),
      ),
      GoRoute(
        path: AppRoutes.vocabularyPath,
        name: AppRoutes.vocabularyName,
        builder: (context, state) => const VocabularyCourseScreen(),
      ),
      GoRoute(
        path: AppRoutes.vocabularyLessonRoute,
        name: AppRoutes.vocabularyLessonName,
        builder: (context, state) => VocabularyLessonScreen(
          lessonId: state.pathParameters['lessonId']!,
        ),
      ),
    ],
    redirect: (context, state) {
      final location = state.matchedLocation;
      final isFirstTime = ref.read(isFirstTimeProvider);
      if (isFirstTime) {
        return location == AppRoutes.onboardingPath
            ? null
            : AppRoutes.onboardingPath;
      }

      if (location == AppRoutes.splashPath ||
          location == AppRoutes.onboardingPath) {
        return AppRoutes.homePath;
      }
      return null;
    },
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Text('Halaman tidak ditemukan: ${state.error}'),
      ),
    ),
  );

  // GoRouter perlu diberi tahu saat state yang dipakai redirect berubah.
  ref
    ..listen<bool>(isFirstTimeProvider, (previous, next) => router.refresh())
    ..onDispose(router.dispose);

  return router;
});

class _SplashScreen extends StatelessWidget {
  const _SplashScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(child: CircularProgressIndicator()),
    );
  }
}
