import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/notifications/notification_settings_screen.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/auth/presentation/screens/login_screen.dart';
import 'package:nihongo_master/features/billing/presentation/premium_screen.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/diagnostic_attempt_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/diagnostic_intro_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/jft_intro_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_attempt_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_intro_screen.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/screens/grammar_course_screen.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/screens/grammar_lesson_screen.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/screens/kanji_course_screen.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/screens/kanji_lesson_screen.dart';
import 'package:nihongo_master/features/learn/reading/presentation/screens/reading_course_screen.dart';
import 'package:nihongo_master/features/learn/reading/presentation/screens/reading_lesson_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_course_screen.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_lesson_screen.dart';
import 'package:nihongo_master/features/onboarding/presentation/screens/onboarding_screen.dart';
import 'package:nihongo_master/features/profile/presentation/screens/achievements_screen.dart';
import 'package:nihongo_master/features/ssw/presentation/kaigo_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_screens.dart';
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
  static const readingPath = '/learn/reading';
  static String readingCoursePath(String level) =>
      '$readingPath?level=${Uri.encodeQueryComponent(level)}';
  static const readingLessonRoute = '/learn/reading/lessons/:lessonId';
  static String readingLessonPath(String id) =>
      '/learn/reading/lessons/${Uri.encodeComponent(id)}';
  static const kanjiPath = '/learn/kanji';
  static const kanjiLessonRoute = '/learn/kanji/lessons/:lessonId';
  static String kanjiLessonPath(String id) =>
      '/learn/kanji/lessons/${Uri.encodeComponent(id)}';
  static const grammarPath = '/learn/grammar';
  static const grammarLessonRoute = '/learn/grammar/lessons/:lessonId';
  static String grammarLessonPath(String id) =>
      '/learn/grammar/lessons/${Uri.encodeComponent(id)}';
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

final isFirstTimeProvider = StateNotifierProvider<FirstTimeNotifier, bool>((
  ref,
) {
  return FirstTimeNotifier(ref.watch(sharedPreferencesProvider));
});

/// Satu instance router untuk seluruh lifecycle aplikasi.
final appRouterProvider = Provider<GoRouter>((ref) {
  final router = GoRouter(
    initialLocation: AppRoutes.splashPath,
    routes: [
      GoRoute(
        path: '/notifications',
        builder: (context, state) => const NotificationSettingsScreen(),
      ),
      GoRoute(
        path: '/anki',
        builder: (context, state) => const AnkiDeckListScreen(),
      ),
      GoRoute(
        path: '/anki/decks/:deckId',
        builder: (context, state) =>
            AnkiDeckDetailScreen(deckId: state.pathParameters['deckId']!),
      ),
      GoRoute(
        path: '/anki/decks/:deckId/study',
        builder: (context, state) =>
            AnkiStudyScreen(deckId: state.pathParameters['deckId']!),
      ),
      GoRoute(
        path: '/dictionary',
        builder: (context, state) => DictionaryScreen(
          initialQuery: state.uri.queryParameters['q'] ?? '',
        ),
      ),
      GoRoute(
        path: '/ssw',
        builder: (context, state) =>
            SswHubScreen(exams: state.uri.queryParameters['tab'] == 'test'),
      ),
      GoRoute(
        path: '/ssw/:sectorId/lessons/:lessonId',
        builder: (context, state) => KaigoLessonScreen(
          lessonId: state.pathParameters['lessonId']!,
          sectorId: state.pathParameters['sectorId']!,
        ),
      ),
      GoRoute(
        path: '/ssw/:sectorId/practice',
        builder: (context, state) =>
            SswPracticeScreen(sector: state.pathParameters['sectorId']!),
      ),
      for (final level in ['SSW-KAIGO', 'SSW-KAIGO-JP', 'SSW-CLEANING']) ...[
        GoRoute(
          path: '/exams/${level.toLowerCase()}/simulation',
          builder: (context, state) => SswSimulationIntroScreen(level: level),
        ),
        GoRoute(
          path: '/exams/${level.toLowerCase()}/simulation/:attemptId',
          builder: (context, state) => SimulationAttemptScreen(
            level: level,
            attemptId: state.pathParameters['attemptId']!,
          ),
        ),
      ],
      GoRoute(
        path: '/ssw/:sectorId',
        builder: (context, state) => SswSectorScreen(
          sectorId: state.pathParameters['sectorId']!,
          exams: state.uri.queryParameters['tab'] == 'test',
        ),
      ),
      GoRoute(
        path: '/premium',
        builder: (context, state) => const PremiumScreen(),
      ),
      GoRoute(
        path: '/achievements',
        builder: (context, state) => const AchievementsScreen(),
      ),
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
        builder: (context, state) => VocabularyCourseScreen(
          initialLevel: state.uri.queryParameters['level'] ?? 'N5',
        ),
      ),
      GoRoute(
        path: '/exams/n5',
        builder: (context, state) => const DiagnosticIntroScreen(),
      ),
      GoRoute(
        path: '/exams/nat/simulation',
        builder: (context, state) => const SimulationIntroScreen(level: 'NAT'),
      ),
      GoRoute(
        path: '/exams/nat/simulation/:attemptId',
        builder: (context, state) => SimulationAttemptScreen(
          level: 'NAT',
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: '/exams/jft/exam-1',
        builder: (context, state) => const JftIntroScreen(),
      ),
      GoRoute(
        path: '/exams/jft/exam-1/:attemptId',
        builder: (context, state) => SimulationAttemptScreen(
          level: 'JFT',
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: '/exams/n5/simulation',
        builder: (context, state) => const SimulationIntroScreen(),
      ),
      GoRoute(
        path: '/exams/n5/simulation/:attemptId',
        builder: (context, state) => SimulationAttemptScreen(
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: '/exams/n4/simulation',
        builder: (context, state) => const SimulationIntroScreen(level: 'N4'),
      ),
      GoRoute(
        path: '/exams/n4/simulation/:attemptId',
        builder: (context, state) => SimulationAttemptScreen(
          level: 'N4',
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: '/exams/n3/simulation',
        builder: (context, state) => const SimulationIntroScreen(level: 'N3'),
      ),
      GoRoute(
        path: '/exams/n3/simulation/:attemptId',
        builder: (context, state) => SimulationAttemptScreen(
          level: 'N3',
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: '/exams/n5/attempt/:attemptId',
        builder: (context, state) => DiagnosticAttemptScreen(
          attemptId: state.pathParameters['attemptId']!,
        ),
      ),
      GoRoute(
        path: AppRoutes.readingPath,
        builder: (context, state) => ReadingCourseScreen(
          level: state.uri.queryParameters['level'] ?? 'N5',
        ),
      ),
      GoRoute(
        path: AppRoutes.readingLessonRoute,
        builder: (context, state) =>
            ReadingLessonScreen(lessonId: state.pathParameters['lessonId']!),
      ),
      GoRoute(
        path: AppRoutes.kanjiPath,
        builder: (context, state) => KanjiCourseScreen(
          level: state.uri.queryParameters['level'] ?? 'N5',
        ),
      ),
      GoRoute(
        path: AppRoutes.kanjiLessonRoute,
        builder: (context, state) =>
            KanjiLessonScreen(lessonId: state.pathParameters['lessonId']!),
      ),
      GoRoute(
        path: AppRoutes.grammarPath,
        builder: (context, state) => GrammarCourseScreen(
          initialLevel: state.uri.queryParameters['level'] ?? 'N5',
        ),
      ),
      GoRoute(
        path: AppRoutes.grammarLessonRoute,
        builder: (context, state) =>
            GrammarLessonScreen(lessonId: state.pathParameters['lessonId']!),
      ),
      GoRoute(
        path: AppRoutes.vocabularyLessonRoute,
        name: AppRoutes.vocabularyLessonName,
        builder: (context, state) =>
            VocabularyLessonScreen(lessonId: state.pathParameters['lessonId']!),
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

      final user = ref.read(authControllerProvider).user;
      if (requiresPremium(state.uri) && !(user?.isPremium ?? false)) {
        return '/premium?from=${Uri.encodeQueryComponent(state.uri.toString())}';
      }
      return null;
    },
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Text('Halaman tidak ditemukan: ${state.error}')),
    ),
  );

  // GoRouter perlu diberi tahu saat state yang dipakai redirect berubah.
  ref
    ..listen<bool>(isFirstTimeProvider, (previous, next) => router.refresh())
    ..listen<bool>(
      authControllerProvider.select((s) => s.user?.isPremium ?? false),
      (previous, next) => router.refresh(),
    )
    ..onDispose(router.dispose);

  return router;
});

bool requiresPremium(Uri uri) {
  final path = uri.path.toLowerCase();
  if (path.startsWith('/ssw/') && path.endsWith('/practice')) return true;
  if (path.startsWith('/exams/') && !path.startsWith('/exams/n5')) return true;
  if (!path.startsWith('/learn/')) return false;
  final level = (uri.queryParameters['level'] ?? '').toUpperCase();
  return const ['N3', 'N2', 'N1'].contains(level) ||
      RegExp(r'(?:-|/|^)n[123](?:-|/|$)').hasMatch(path);
}

class _SplashScreen extends StatelessWidget {
  const _SplashScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
