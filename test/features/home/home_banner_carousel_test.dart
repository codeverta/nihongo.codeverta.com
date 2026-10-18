import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/home/presentation/screens/dashboard_screen.dart';
import 'package:nihongo_master/features/home/presentation/widgets/home_banner_carousel.dart';
import 'package:nihongo_master/shared/presentation/widgets/auth_gate.dart';

const banners = [
  HomeBanner(
    id: 'learn',
    title: 'Materi Belajar',
    description: 'Kosakata, kanji, grammar dan reading.',
    imageUrl: '',
    action: 'learn',
    buttonLabel: 'Buka materi',
  ),
  HomeBanner(
    id: 'ssw',
    title: 'SSW / Tokutei Ginou',
    description: 'Siapkan keterampilan dan ujianmu.',
    imageUrl: '',
    action: 'ssw',
    buttonLabel: 'Buka SSW',
  ),
  HomeBanner(
    id: 'jobs',
    title: 'Kerja di Jepang',
    description: 'Temukan peluang sesuai kualifikasimu.',
    imageUrl: '',
    action: 'jobs',
    buttonLabel: 'Buka lowongan',
  ),
];

Widget carousel({
  bool active = true,
  double textScale = 1,
  bool reduceMotion = false,
  List<HomeBanner> items = banners,
  ValueChanged<HomeBanner>? onOpen,
}) => MaterialApp(
  home: MediaQuery(
    data: MediaQueryData(
      textScaler: TextScaler.linear(textScale),
      disableAnimations: reduceMotion,
    ),
    child: Scaffold(
      body: SingleChildScrollView(
        child: Center(
          child: SizedBox(
            width: 320,
            child: HomeBannerCarousel(
              banners: items,
              active: active,
              onOpen: onOpen ?? (_) {},
            ),
          ),
        ),
      ),
    ),
  ),
);

void main() {
  testWidgets('auto advances after five seconds; swipe pauses until play', (
    tester,
  ) async {
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpWidget(carousel());
    await tester.pumpAndSettle();
    final view = tester.widget<PageView>(
      find.byKey(const Key('home-banner-pages')),
    );
    expect(view.controller!.page, 0);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    await tester.pump(const Duration(seconds: 6));
    expect(view.controller!.page, 0);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump(const Duration(seconds: 5));
    await tester.pumpAndSettle();
    expect(view.controller!.page, 1);
    await tester.drag(
      find.byKey(const Key('home-banner-pages')),
      const Offset(-300, 0),
    );
    await tester.pumpAndSettle();
    expect(view.controller!.page, 2);
    expect(find.byTooltip('Putar banner'), findsOneWidget);
    await tester.pump(const Duration(seconds: 6));
    await tester.pumpAndSettle();
    expect(view.controller!.page, 2);
    await tester.tap(find.byKey(const Key('home-banner-toggle')));
    await tester.pump(const Duration(seconds: 5));
    await tester.pumpAndSettle();
    expect(view.controller!.page, 0);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('inactive home and reduced motion never auto advance', (
    tester,
  ) async {
    await tester.pumpWidget(carousel(active: false));
    await tester.pumpAndSettle();
    var view = tester.widget<PageView>(
      find.byKey(const Key('home-banner-pages')),
    );
    await tester.pump(const Duration(seconds: 6));
    expect(view.controller!.page, 0);
    await tester.pumpWidget(carousel(reduceMotion: true));
    await tester.pumpAndSettle();
    view = tester.widget<PageView>(find.byKey(const Key('home-banner-pages')));
    await tester.pump(const Duration(seconds: 6));
    expect(view.controller!.page, 0);
    expect(find.byKey(const Key('home-banner-toggle')), findsNothing);
    await tester.tap(find.byKey(const Key('home-banner-dot-2')));
    await tester.pumpAndSettle();
    expect(view.controller!.page, 2);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('each CTA opens its own destination and large text fits', (
    tester,
  ) async {
    final opened = <String>[];
    await tester.pumpWidget(
      carousel(
        active: false,
        textScale: 2,
        onOpen: (b) => opened.add(b.action),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    for (var i = 0; i < banners.length; i++) {
      await tester.ensureVisible(find.byKey(Key('home-banner-dot-$i')));
      await tester.tap(find.byKey(Key('home-banner-dot-$i')));
      await tester.pumpAndSettle();
      final button = find.widgetWithText(FilledButton, banners[i].buttonLabel);
      await tester.ensureVisible(button);
      await tester.tap(button);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    }
    expect(opened, ['learn', 'ssw', 'jobs']);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('empty and expanded backend banner lists remain usable', (
    tester,
  ) async {
    await tester.pumpWidget(carousel(active: false, items: const []));
    expect(find.byKey(const Key('home-banner-carousel')), findsNothing);
    final many = List.generate(
      12,
      (i) => HomeBanner(
        id: 'banner-$i',
        title: 'Materi $i',
        description: '',
        imageUrl: '',
        action: 'learn',
        buttonLabel: 'Buka',
      ),
    );
    await tester.pumpWidget(carousel(active: false, items: many));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.ensureVisible(find.byKey(const Key('home-banner-dot-11')));
    await tester.tap(find.byKey(const Key('home-banner-dot-11')));
    await tester.pumpAndSettle();
    expect(tester.widget<PageView>(find.byType(PageView)).controller!.page, 11);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('dashboard sends each banner to its actual destination', (
    tester,
  ) async {
    var learnOpened = false;
    final router = GoRouter(
      routes: [
        GoRoute(
          path: '/premium',
          builder: (context, state) =>
              const Scaffold(body: Text('Premium destination')),
        ),
        GoRoute(
          path: '/',
          builder: (context, state) => AuthGate(
            authorize: () async => true,
            child: Scaffold(
              body: DashboardScreen(
                active: false,
                openLearn: () => learnOpened = true,
                openExam: () {},
                openProfile: () {},
              ),
            ),
          ),
        ),
        GoRoute(
          path: '/jobs',
          builder: (context, state) =>
              const Scaffold(body: Text('Job destination')),
        ),
        GoRoute(
          path: '/ssw',
          builder: (context, state) =>
              const Scaffold(body: Text('SSW destination')),
        ),
      ],
    );
    addTearDown(router.dispose);
    const data = HomeDashboard(
      isAuthenticated: true,
      dailyGoal: HomeDailyGoal(
        minutesLearned: 0,
        targetMinutes: 20,
        streakDays: 0,
      ),
      banners: banners,
      categories: [],
      tracks: [],
      exams: [],
      plans: [
        SubscriptionPlan(
          id: 'lifetime',
          name: 'Lifetime Access',
          billingType: 'lifetime',
          priceIdr: 149000,
          features: ['Semua materi'],
          recommended: true,
        ),
      ],
      achievements: [],
    );
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authControllerProvider.overrideWith(
            (ref) => AuthController.test(const AuthenticationState.guest()),
          ),
          homeDashboardProvider.overrideWith((ref) async => data),
        ],
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Lihat paket'));
    await tester.pumpAndSettle();
    expect(find.text('Premium destination'), findsOneWidget);
    router.pop();
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('home-banner-dot-3')));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Buka lowongan'));
    await tester.pumpAndSettle();
    expect(find.text('Job destination'), findsOneWidget);
    router.pop();
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('home-banner-dot-2')));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Buka SSW'));
    await tester.pumpAndSettle();
    expect(find.text('SSW destination'), findsOneWidget);
    router.pop();
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('home-banner-dot-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Buka materi'));
    await tester.pumpAndSettle();
    expect(learnOpened, isTrue);
    await tester.pumpWidget(const SizedBox.shrink());
  });
}
