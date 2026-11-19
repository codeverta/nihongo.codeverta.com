import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/core/widgets/skeleton.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/home/presentation/screens/dashboard_screen.dart';
import 'package:nihongo_master/features/news/data/news_repository.dart';
import 'package:nihongo_master/features/news/presentation/news_screens.dart';
import 'package:shared_preferences/shared_preferences.dart';

Future<void> _pumpLoadingDashboard(
  WidgetTester tester, {
  bool disableAnimations = false,
}) async {
  tester.view
    ..physicalSize = const Size(390, 844)
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  // Never completes: the dashboard stays in its loading state.
  final pending = Completer<HomeDashboard>();
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(prefs),
        authControllerProvider.overrideWith(
          (ref) => AuthController.test(const AuthenticationState.guest()),
        ),
        homeDashboardProvider.overrideWith((ref) => pending.future),
        newsPreviewProvider.overrideWith((ref) => Completer<NewsPage>().future),
      ],
      child: MediaQuery(
        data: MediaQueryData(
          size: const Size(390, 844),
          disableAnimations: disableAnimations,
        ),
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: Scaffold(
            body: DashboardScreen(
              active: true,
              openLearn: () {},
              openExam: () {},
              openProfile: () {},
            ),
          ),
        ),
      ),
    ),
  );
  await tester.pump(const Duration(milliseconds: 300));
}

void main() {
  testWidgets('a slow dashboard shows a page-shaped skeleton, not a spinner', (
    tester,
  ) async {
    await _pumpLoadingDashboard(tester);
    expect(find.byKey(const Key('home-loading')), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsNothing);
    expect(find.byType(SkeletonShimmer), findsWidgets);
    // Greeting, banner, progress, 8 quick-menu tiles and 3 news rows.
    expect(find.byType(SkeletonBox), findsAtLeastNWidgets(20));
    expect(find.byKey(const Key('news-preview-skeleton')), findsOneWidget);
    // The shimmer keeps animating while loading.
    await tester.pump(const Duration(milliseconds: 400));
    expect(find.byType(ShaderMask), findsWidgets);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('the shimmer stays still when animations are turned off', (
    tester,
  ) async {
    await _pumpLoadingDashboard(tester, disableAnimations: true);
    expect(find.byType(SkeletonBox), findsWidgets);
    expect(find.byType(ShaderMask), findsNothing);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('the news preview shows article placeholders while loading', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          newsPreviewProvider.overrideWith(
            (ref) => Completer<NewsPage>().future,
          ),
        ],
        child: MaterialApp(
          theme: AppTheme.lightTheme,
          home: const Scaffold(
            body: SingleChildScrollView(child: NewsPreview()),
          ),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.byKey(const Key('news-preview-skeleton')), findsOneWidget);
    expect(find.byType(LinearProgressIndicator), findsNothing);
    await tester.pumpWidget(const SizedBox.shrink());
  });
}
