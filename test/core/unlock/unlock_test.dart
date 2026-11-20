import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/unlock/locked_content.dart';
import 'package:nihongo_master/core/unlock/unlock_flow.dart';
import 'package:nihongo_master/core/unlock/unlock_repository.dart';
import 'package:nihongo_master/core/unlock/unlock_sheet.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

const _target = UnlockTarget('grammar', 'grammar-n3-01');

DioException _error(int status, Map<String, dynamic> body) {
  final request = RequestOptions(path: '/x');
  return DioException(
    requestOptions: request,
    response: Response(requestOptions: request, statusCode: status, data: body),
    type: DioExceptionType.badResponse,
  );
}

class _FakeRepository extends UnlockRepository {
  _FakeRepository() : super(Dio());

  bool alreadyUnlocked = false;
  List<bool> claims = [true];
  final calls = <String>[];

  @override
  Future<UnlockIntent> createIntent(UnlockTarget target) async {
    calls.add('intent:${target.key}');
    return UnlockIntent(
      alreadyUnlocked: alreadyUnlocked,
      nonce: 'nonce-1',
      userId: 'user-1',
    );
  }

  @override
  Future<bool> claim(String nonce) async {
    calls.add('claim:$nonce');
    return claims.length > 1 ? claims.removeAt(0) : claims.first;
  }

  @override
  Future<void> fallback(UnlockTarget target) async {
    calls.add('fallback:${target.key}');
  }
}

class _FakeShower implements RewardedAdShower {
  _FakeShower(this.outcomes);
  final List<RewardedAdOutcome> outcomes;
  final shown = <({String unit, String? userId, String? customData})>[];

  @override
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  }) async {
    shown.add((unit: unitId, userId: userId, customData: customData));
    return outcomes.length > 1 ? outcomes.removeAt(0) : outcomes.first;
  }
}

void main() {
  group('UnlockTarget', () {
    test('reads the item a 402 response offers to unlock', () {
      final error = _error(402, {
        'error': {
          'code': 'premium_required',
          'message': 'Terkunci',
          'unlockable': true,
          'unlock': {'kind': 'news', 'itemId': 'a1'},
        },
      });
      expect(UnlockTarget.fromError(error), const UnlockTarget('news', 'a1'));
      expect(UnlockTarget.fromError(error)!.group, 'news');
      final locked = LockedContentException.from(error)!;
      expect(locked.target?.key, 'news:a1');
      expect(locked.message, 'Terkunci');
    });

    test('a Lifetime-only lock has no target, other errors are not locks', () {
      final lifetimeOnly = _error(402, {
        'error': {'code': 'premium_required', 'message': 'Lifetime'},
      });
      expect(UnlockTarget.fromError(lifetimeOnly), isNull);
      expect(LockedContentException.from(lifetimeOnly)!.target, isNull);
      expect(LockedContentException.from(_error(500, {})), isNull);
      expect(LockedContentException.from(Exception('x')), isNull);
      expect(
        isLoginRequired(
          _error(401, {
            'error': {'code': 'login_required'},
          }),
        ),
        isTrue,
      );
      expect(const UnlockTarget('job', 'j').group, 'jobs');
      expect(const UnlockTarget('kanji', 'k').group, 'lessons');
    });
  });

  group('UnlockFlow', () {
    test(
      'passes the intent to AdMob server-side verification and claims',
      () async {
        final repository = _FakeRepository();
        final shower = _FakeShower([RewardedAdOutcome.watched]);
        final flow = UnlockFlow(
          repository: repository,
          shower: shower,
          useFallbackAfterTestAds: false,
        );
        final attempt = await flow.watchAd(_target, 'unit');
        expect(attempt.outcome, UnlockOutcome.unlocked);
        expect(shower.shown.single.userId, 'user-1');
        expect(shower.shown.single.customData, 'nonce-1');
        expect(repository.calls, [
          'intent:grammar:grammar-n3-01',
          'claim:nonce-1',
        ]);
      },
    );

    test('keeps the nonce when Google has not confirmed yet', () async {
      final repository = _FakeRepository()..claims = [false];
      final flow = UnlockFlow(
        repository: repository,
        shower: _FakeShower([RewardedAdOutcome.watched]),
        useFallbackAfterTestAds: false,
      );
      final attempt = await flow.watchAd(_target, 'unit');
      expect(attempt.outcome, UnlockOutcome.pending);
      expect(attempt.nonce, 'nonce-1');
      expect(
        repository.calls.where((c) => c.startsWith('claim')),
        hasLength(3),
      );
      repository.claims = [true];
      expect((await flow.claim('nonce-1')).outcome, UnlockOutcome.unlocked);
    });

    test('a closed or missing ad does not unlock', () async {
      for (final (outcome, expected) in [
        (RewardedAdOutcome.dismissed, UnlockOutcome.cancelled),
        (RewardedAdOutcome.unavailable, UnlockOutcome.adUnavailable),
      ]) {
        final repository = _FakeRepository();
        final flow = UnlockFlow(
          repository: repository,
          shower: _FakeShower([outcome]),
          useFallbackAfterTestAds: false,
        );
        expect((await flow.watchAd(_target, 'unit')).outcome, expected);
        expect(repository.calls.any((c) => c.startsWith('claim')), isFalse);
      }
    });

    test('an open item skips the ad; test ads use the fallback', () async {
      final repository = _FakeRepository()..alreadyUnlocked = true;
      final shower = _FakeShower([RewardedAdOutcome.watched]);
      final flow = UnlockFlow(
        repository: repository,
        shower: shower,
        useFallbackAfterTestAds: false,
      );
      expect(
        (await flow.watchAd(_target, 'unit')).outcome,
        UnlockOutcome.unlocked,
      );
      expect(shower.shown, isEmpty);

      final debugRepository = _FakeRepository();
      final debugFlow = UnlockFlow(
        repository: debugRepository,
        shower: _FakeShower([RewardedAdOutcome.watched]),
        useFallbackAfterTestAds: true,
      );
      expect(
        (await debugFlow.watchAd(_target, 'unit')).outcome,
        UnlockOutcome.unlocked,
      );
      expect(debugRepository.calls.last, 'fallback:grammar:grammar-n3-01');
    });
  });

  group('UnlockSheet', () {
    Future<List<bool>> pumpSheet(
      WidgetTester tester, {
      required _FakeShower shower,
      required _FakeRepository repository,
      bool adsOn = true,
      UnlockStatus status = const UnlockStatus(
        remaining: 5,
        dailyCap: 20,
        fallbackRemaining: 2,
      ),
    }) async {
      final results = <bool>[];
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) => Scaffold(
              body: Builder(
                builder: (context) => TextButton(
                  onPressed: () async =>
                      results.add(await showUnlockSheet(context, _target)),
                  child: const Text('open'),
                ),
              ),
            ),
          ),
          GoRoute(
            path: '/premium',
            builder: (_, __) =>
                const Scaffold(body: Text('Premium destination')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            unlockRewardedPlacementProvider.overrideWith(
              (ref) async => adsOn
                  ? const AdsConfig(
                      enabled: true,
                      unlockRewarded: AdPlacementConfig(
                        enabled: true,
                        unitId: testRewardedAdUnitId,
                      ),
                    )
                  : null,
            ),
            unlockStatusProvider.overrideWith((ref) async => status),
            unlockFlowProvider.overrideWithValue(
              UnlockFlow(
                repository: repository,
                shower: shower,
                useFallbackAfterTestAds: false,
              ),
            ),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();
      return results;
    }

    testWidgets('watching the ad opens the item', (tester) async {
      final results = await pumpSheet(
        tester,
        shower: _FakeShower([RewardedAdOutcome.watched]),
        repository: _FakeRepository(),
      );
      expect(find.text('Sisa buka gratis hari ini: 5'), findsOneWidget);
      await tester.tap(find.byKey(const Key('unlock-watch-ad')));
      await tester.pumpAndSettle();
      expect(results, [true]);
      expect(find.byType(UnlockSheet), findsNothing);
    });

    testWidgets('after two failed ads the limited no-ad option appears', (
      tester,
    ) async {
      final repository = _FakeRepository();
      final results = await pumpSheet(
        tester,
        shower: _FakeShower([RewardedAdOutcome.unavailable]),
        repository: repository,
      );
      expect(find.byKey(const Key('unlock-fallback')), findsNothing);
      for (var i = 0; i < 2; i++) {
        await tester.tap(find.byKey(const Key('unlock-watch-ad')));
        await tester.pumpAndSettle();
      }
      expect(find.byKey(const Key('unlock-message')), findsOneWidget);
      await tester.tap(find.byKey(const Key('unlock-fallback')));
      await tester.pumpAndSettle();
      expect(results, [true]);
      expect(repository.calls.last, 'fallback:grammar:grammar-n3-01');
    });

    testWidgets('without ads only Lifetime is offered', (tester) async {
      final results = await pumpSheet(
        tester,
        shower: _FakeShower([RewardedAdOutcome.watched]),
        repository: _FakeRepository(),
        adsOn: false,
      );
      expect(find.byKey(const Key('unlock-watch-ad')), findsNothing);
      await tester.tap(find.byKey(const Key('unlock-lifetime')));
      await tester.pumpAndSettle();
      expect(find.text('Premium destination'), findsOneWidget);
      expect(results, [false]);
    });

    testWidgets('an ad the server switched off for this kind is not offered', (
      tester,
    ) async {
      await pumpSheet(
        tester,
        shower: _FakeShower([RewardedAdOutcome.watched]),
        repository: _FakeRepository(),
        status: const UnlockStatus(offered: {'lessons': false}),
      );
      expect(find.byKey(const Key('unlock-watch-ad')), findsNothing);
      expect(find.byKey(const Key('unlock-lifetime')), findsOneWidget);
    });
  });

  group('UnlockLockIcon', () {
    Future<void> pumpIcon(
      WidgetTester tester, {
      required UnlockStatus status,
      bool lifetime = false,
      bool lockable = true,
    }) async {
      await tester.pumpWidget(
        ProviderScope(
          key: UniqueKey(),
          overrides: [
            unlockStatusProvider.overrideWith((ref) async => status),
            authControllerProvider.overrideWith(
              (_) => AuthController.test(
                AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'u',
                    displayName: 'U',
                    email: 'u@example.com',
                    isLifetime: lifetime,
                    isPremium: lifetime,
                  ),
                ),
              ),
            ),
          ],
          child: MaterialApp(
            home: Scaffold(
              body: UnlockLockIcon(target: _target, lockable: lockable),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
    }

    final lock = find.byKey(const Key('locked-grammar:grammar-n3-01'));

    testWidgets(
      'draws a lock only while the gate is on and the item is closed',
      (tester) async {
        await pumpIcon(
          tester,
          status: const UnlockStatus(gates: {'lessons': true}),
        );
        expect(lock, findsOneWidget);
        await pumpIcon(
          tester,
          status: const UnlockStatus(gates: {'lessons': false}),
        );
        expect(lock, findsNothing, reason: 'open content never shows a lock');
        await pumpIcon(
          tester,
          status: const UnlockStatus(
            gates: {'lessons': true},
            items: {'grammar:grammar-n3-01'},
          ),
        );
        expect(lock, findsNothing, reason: 'an unlocked item is open');
        await pumpIcon(
          tester,
          status: const UnlockStatus(gates: {'lessons': true}),
          lifetime: true,
        );
        expect(lock, findsNothing, reason: 'Lifetime opens everything');
        await pumpIcon(
          tester,
          status: const UnlockStatus(gates: {'lessons': true}),
          lockable: false,
        );
        expect(lock, findsNothing, reason: 'N5 and N4 are free');
      },
    );
  });
}
