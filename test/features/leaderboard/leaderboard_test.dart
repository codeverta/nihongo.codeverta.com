import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/leaderboard/data/leaderboard_repository.dart';
import 'package:nihongo_master/features/leaderboard/presentation/home_leaderboard_widgets.dart';
import 'package:nihongo_master/features/leaderboard/presentation/leaderboard_copy.dart';
import 'package:nihongo_master/features/leaderboard/presentation/leaderboard_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _user = AuthenticatedUser(
  id: 'u1',
  displayName: 'Rina',
  email: 'rina@example.com',
);

LeaderboardPage _page({int count = 50, int participants = 60}) =>
    LeaderboardPage(
      period: 'week',
      weekEndsAt: DateTime.now().toUtc().add(
        const Duration(days: 2, hours: 5, minutes: 30),
      ),
      participants: participants,
      entries: [
        for (var i = 1; i <= count; i++)
          LeaderboardEntry(
            rank: i,
            // Every fifth person has no usable name and shows the tag.
            displayName: i % 5 == 0
                ? ''
                : 'Pelajar dengan nama yang sangat panjang sekali $i',
            tag: '${1000 + i}',
            xp: 2000 - i * 10,
          ),
      ],
    );

MyLeaderboard _me({
  int rank = 58,
  int xp = 120,
  bool hidden = false,
  WeeklyResult? lastWeek,
}) => MyLeaderboard(
  hidden: hidden,
  displayName: 'Rina',
  tag: '4321',
  weekEndsAt: DateTime.now().toUtc().add(const Duration(days: 2)),
  week: LeaderboardStanding(
    xp: xp,
    rank: rank,
    participants: 60,
    gapToNext: rank > 1 ? 20 : 0,
  ),
  all: LeaderboardStanding(
    xp: xp * 3,
    rank: rank,
    participants: 90,
    gapToNext: 5,
  ),
  lastWeek: lastWeek,
);

Future<void> _pump(
  WidgetTester tester,
  Widget child, {
  AuthenticationState auth = const AuthenticationState.authenticated(_user),
  Cached<LeaderboardPage>? page,
  Exception? pageError,
  Cached<MyLeaderboard>? me,
  Locale locale = const Locale('id'),
  Size size = const Size(390, 844),
  double textScale = 1,
  Map<String, Object> prefs = const {},
}) async {
  tester.view
    ..physicalSize = size
    ..devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  SharedPreferences.setMockInitialValues(prefs);
  final preferences = await SharedPreferences.getInstance();
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(preferences),
        authControllerProvider.overrideWith((_) => AuthController.test(auth)),
        leaderboardPageProvider.overrideWith((ref, period) async {
          if (pageError != null) throw pageError;
          return page ?? Cached(_page(), DateTime.now());
        }),
        myLeaderboardProvider.overrideWith(
          (ref) async => auth.user == null ? null : me,
        ),
      ],
      child: MaterialApp(
        theme: AppTheme.lightTheme,
        locale: locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(
            context,
          ).copyWith(textScaler: TextScaler.linear(textScale)),
          child: child!,
        ),
        home: child,
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('shows podium, top list, countdown and pins own row', (
    tester,
  ) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      me: Cached(_me(), DateTime.now()),
    );
    expect(find.text('Minggu ini'), findsOneWidget);
    expect(find.text('Sepanjang masa'), findsOneWidget);
    expect(find.byKey(const Key('leaderboard-podium')), findsOneWidget);
    expect(find.textContaining('Berakhir dalam 2 hari 5 jam'), findsOneWidget);
    expect(find.text('Pelajar #1005'), findsOneWidget);
    final own = find.byKey(const Key('leaderboard-own-row'));
    expect(own, findsOneWidget);
    expect(
      find.descendant(of: own, matching: find.text('#58 dari 60')),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('own row inside the top 50 is highlighted, not pinned', (
    tester,
  ) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      me: Cached(_me(rank: 7, xp: 1930), DateTime.now()),
    );
    expect(find.byKey(const Key('leaderboard-own-row')), findsNothing);
    final row = tester.widget<Container>(
      find
          .descendant(
            of: find.byKey(const ValueKey('leaderboard-row-7')),
            matching: find.byType(Container),
          )
          .first,
    );
    expect((row.decoration! as BoxDecoration).border, isNotNull);
  });

  testWidgets('hidden users see their position if shown', (tester) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      me: Cached(_me(rank: 7, hidden: true), DateTime.now()),
    );
    expect(find.text('Posisimu jika ditampilkan: #7'), findsOneWidget);
    expect(find.byKey(const Key('leaderboard-hidden-note')), findsOneWidget);
  });

  testWidgets('guests can view the list with a sign-in prompt', (tester) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      auth: const AuthenticationState.guest(),
    );
    expect(find.byKey(const Key('leaderboard-guest-prompt')), findsOneWidget);
    expect(find.text('Masuk untuk ikut peringkat'), findsOneWidget);
    expect(find.byKey(const Key('leaderboard-own-row')), findsNothing);
  });

  testWidgets('offline shows cached data with the update time, no countdown', (
    tester,
  ) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      page: Cached(
        _page(count: 3, participants: 3),
        DateTime(2026, 11, 20, 14, 5),
        fromCache: true,
      ),
      me: Cached(_me(rank: 2), DateTime.now(), fromCache: true),
    );
    expect(find.text('Terakhir diperbarui 14.05'), findsOneWidget);
    expect(find.textContaining('Berakhir dalam'), findsNothing);
  });

  testWidgets('never loaded and offline asks for a connection', (tester) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      pageError: const LeaderboardOffline(),
    );
    expect(find.text('Butuh koneksi internet'), findsOneWidget);
  });

  testWidgets('fits a 320 px phone at 1.5x text, in English', (tester) async {
    await _pump(
      tester,
      const LeaderboardScreen(),
      me: Cached(_me(), DateTime.now()),
      locale: const Locale('en'),
      size: const Size(320, 640),
      textScale: 1.5,
    );
    expect(tester.takeException(), isNull);
    expect(find.text('This week'), findsOneWidget);
    expect(find.text('#58 of 60'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Learner #1005'),
      200,
      scrollable: find.byType(Scrollable).at(1),
    );
    expect(find.text('Learner #1005'), findsOneWidget);
    await tester.drag(find.byType(ListView).first, const Offset(0, -2000));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('All time'));
    await tester.pumpAndSettle();
    expect(find.text('Total XP since joining'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  group('Beranda', () {
    Widget card(Widget child) => Scaffold(
      body: Padding(padding: const EdgeInsets.all(16), child: child),
    );

    testWidgets('rank row shows rank and gap to the next place', (
      tester,
    ) async {
      await _pump(
        tester,
        card(const HomeLeaderboardRankRow()),
        me: Cached(_me(rank: 12), DateTime.now()),
        size: const Size(320, 640),
        textScale: 1.5,
      );
      expect(find.text('#12 minggu ini · 20 XP lagi ke #11'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('zero XP this week invites the first lesson', (tester) async {
      await _pump(
        tester,
        card(const HomeLeaderboardRankRow()),
        me: Cached(_me(rank: 0, xp: 0), DateTime.now()),
      );
      expect(
        find.text(
          'Belum masuk peringkat minggu ini — selesaikan 1 lesson untuk ikut bersaing',
        ),
        findsOneWidget,
      );
    });

    testWidgets('guests get a slim link row', (tester) async {
      await _pump(
        tester,
        card(const HomeLeaderboardGuestRow()),
        auth: const AuthenticationState.guest(),
      );
      expect(find.text('Lihat leaderboard minggu ini'), findsOneWidget);
    });

    testWidgets('last week card shows once and can be dismissed', (
      tester,
    ) async {
      const result = WeeklyResult(weekStart: '2026-11-09', rank: 4, xp: 310);
      await _pump(
        tester,
        card(const HomeLastWeekCard()),
        me: Cached(_me(lastWeek: result), DateTime.now()),
      );
      expect(
        find.text('Minggu lalu kamu finis #4 · 310 XP. Minggu baru dimulai!'),
        findsOneWidget,
      );
      await tester.tap(find.byKey(const Key('home-last-week-dismiss')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('home-last-week-card')), findsNothing);
      final prefs = await SharedPreferences.getInstance();
      expect(prefs.getString('leaderboard_last_week_dismissed'), '2026-11-09');
    });

    testWidgets('a dismissed week stays hidden', (tester) async {
      const result = WeeklyResult(weekStart: '2026-11-09', rank: 4, xp: 310);
      await _pump(
        tester,
        card(const HomeLastWeekCard()),
        me: Cached(_me(lastWeek: result), DateTime.now()),
        prefs: {'leaderboard_last_week_dismissed': '2026-11-09'},
      );
      expect(find.byKey(const Key('home-last-week-card')), findsNothing);
    });
  });

  test('copy covers both languages', () {
    const id = LeaderboardCopy(english: false);
    const en = LeaderboardCopy(english: true);
    expect(id.homeRank(1, 0), '#1 minggu ini · Kamu memimpin!');
    expect(en.homeRank(3, 15), '#3 this week · 15 XP more to reach #2');
    expect(id.endsIn(const Duration(minutes: 35)), 'Berakhir dalam 35 menit');
    expect(
      en.endsIn(const Duration(days: 1, hours: 1)),
      'Ends in 1 day 1 hour',
    );
    expect(en.endsIn(const Duration(seconds: -5)), 'Ends in 0 minutes');
    expect(id.name('', '0042'), 'Pelajar #0042');
    expect(en.lastWeek(1, 900), contains('Last week you finished #1'));
  });

  group('repository cache', () {
    Future<LeaderboardRepository> repository(
      Map<String, Object> saved, {
      required bool online,
    }) async {
      SharedPreferences.setMockInitialValues(saved);
      final prefs = await SharedPreferences.getInstance();
      final dio = Dio(BaseOptions(baseUrl: 'https://api.test/v1'))
        ..interceptors.add(
          InterceptorsWrapper(
            onRequest: (options, handler) {
              if (!online) {
                handler.reject(
                  DioException(
                    requestOptions: options,
                    type: DioExceptionType.connectionError,
                  ),
                );
                return;
              }
              handler.resolve(
                Response(
                  requestOptions: options,
                  statusCode: 200,
                  data: {'data': _page(count: 2, participants: 2).toJson()},
                ),
              );
            },
          ),
        );
      return LeaderboardRepository(dio, prefs);
    }

    test('saves a successful response and serves it offline', () async {
      final online = await repository({}, online: true);
      final fresh = await online.fetchPage(LeaderboardPeriod.week);
      expect(fresh.fromCache, isFalse);
      final saved = online.prefs.getString(
        LeaderboardRepository.pageCacheKey(LeaderboardPeriod.week),
      )!;
      final offline = await repository({
        LeaderboardRepository.pageCacheKey(LeaderboardPeriod.week): saved,
      }, online: false);
      final cached = await offline.fetchPage(LeaderboardPeriod.week);
      expect(cached.fromCache, isTrue);
      expect(cached.data.entries.length, 2);
      expect(jsonDecode(saved), containsPair('at', isA<String>()));
    });

    test('offline with no cache reports LeaderboardOffline', () async {
      final offline = await repository({}, online: false);
      expect(
        () => offline.fetchPage(LeaderboardPeriod.all),
        throwsA(isA<LeaderboardOffline>()),
      );
    });
  });
}
