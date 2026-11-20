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
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/jobs/data/jobs_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/jobs/presentation/job_apply_button.dart';

// With the jobs gate on, the public listing has no link (applyLocked) and the
// button asks the server, which wants Lifetime or one rewarded ad.

class _Jobs extends JobsRepository {
  _Jobs({required this.lockedFirst}) : super(Dio());
  bool lockedFirst;
  int calls = 0;

  @override
  Future<Uri?> applyLink(String id) async {
    calls++;
    if (lockedFirst) {
      lockedFirst = false;
      throw const LockedContentException(target: UnlockTarget('job', 'one'));
    }
    return Uri.parse('https://example.com/apply');
  }
}

class _Unlocks extends UnlockRepository {
  _Unlocks() : super(Dio());
  @override
  Future<UnlockIntent> createIntent(UnlockTarget target) async =>
      const UnlockIntent(alreadyUnlocked: false, nonce: 'n', userId: 'buyer');
  @override
  Future<bool> claim(String nonce) async => true;
}

class _Shower implements RewardedAdShower {
  @override
  Future<RewardedAdOutcome> show(
    String unitId, {
    String? userId,
    String? customData,
  }) async => RewardedAdOutcome.watched;
}

Future<({List<Uri> launched, _Jobs jobs})> _pump(
  WidgetTester tester, {
  required AuthenticationState auth,
  required bool lockedFirst,
}) async {
  final launched = <Uri>[];
  final jobs = _Jobs(lockedFirst: lockedFirst);
  final job = JapanJob.fromJson({
    'id': 'one',
    'status': 'active',
    'applyUrl': '',
    'applyLocked': true,
  });
  final router = GoRouter(
    initialLocation: '/jobs/one',
    routes: [
      GoRoute(
        path: '/jobs/one',
        builder: (_, __) => Scaffold(body: JobApplyButton(job: job)),
      ),
      GoRoute(
        path: '/login',
        builder: (_, __) => const Scaffold(body: Text('Login destination')),
      ),
      GoRoute(
        path: '/premium',
        builder: (_, __) => const Scaffold(body: Text('Premium destination')),
      ),
    ],
  );
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authControllerProvider.overrideWith((_) => AuthController.test(auth)),
        jobsRepositoryProvider.overrideWithValue(jobs),
        jobApplicationLauncherProvider.overrideWithValue((uri) async {
          launched.add(uri);
          return true;
        }),
        unlockRewardedPlacementProvider.overrideWith(
          (ref) async => const AdsConfig(
            enabled: true,
            unlockRewarded: AdPlacementConfig(
              enabled: true,
              unitId: testRewardedAdUnitId,
            ),
          ),
        ),
        unlockStatusProvider.overrideWith(
          (ref) async => const UnlockStatus(remaining: 3, dailyCap: 20),
        ),
        unlockFlowProvider.overrideWithValue(
          UnlockFlow(
            repository: _Unlocks(),
            shower: _Shower(),
            useFallbackAfterTestAds: false,
          ),
        ),
      ],
      child: MaterialApp.router(routerConfig: router),
    ),
  );
  await tester.pumpAndSettle();
  return (launched: launched, jobs: jobs);
}

AuthenticationState _user({bool lifetime = false}) =>
    AuthenticationState.authenticated(
      AuthenticatedUser(
        id: 'buyer',
        displayName: 'Buyer',
        email: 'buyer@example.com',
        isLifetime: lifetime,
        isPremium: lifetime,
      ),
    );

void main() {
  testWidgets('a locked job is applied to after watching one ad', (
    tester,
  ) async {
    final result = await _pump(tester, auth: _user(), lockedFirst: true);
    final button = tester.widget<FilledButton>(
      find.byKey(const Key('apply-job')),
    );
    expect(button.onPressed, isNotNull, reason: 'applyLocked keeps it usable');

    await tester.tap(find.byKey(const Key('apply-job')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('unlock-watch-ad')), findsOneWidget);
    expect(result.launched, isEmpty);

    await tester.tap(find.byKey(const Key('unlock-watch-ad')));
    await tester.pumpAndSettle();
    expect(result.jobs.calls, 2);
    expect(result.launched, [Uri.parse('https://example.com/apply')]);
  });

  testWidgets('Lifetime gets the link without an ad', (tester) async {
    final result = await _pump(
      tester,
      auth: _user(lifetime: true),
      lockedFirst: false,
    );
    await tester.tap(find.byKey(const Key('apply-job')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('unlock-watch-ad')), findsNothing);
    expect(result.launched, hasLength(1));
  });

  testWidgets('a guest is asked to sign in first', (tester) async {
    final result = await _pump(
      tester,
      auth: const AuthenticationState.guest(),
      lockedFirst: true,
    );
    await tester.tap(find.byKey(const Key('apply-job')));
    await tester.pumpAndSettle();
    expect(find.text('Login destination'), findsOneWidget);
    expect(result.jobs.calls, 0);
  });
}
