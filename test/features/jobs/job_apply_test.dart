import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/jobs/presentation/job_apply_button.dart';

void main() {
  for (final scenario in ['guest', 'free', 'paid', 'revoked', 'offline']) {
    testWidgets('$scenario account follows the correct application flow', (
      tester,
    ) async {
      var checks = 0;
      final launched = <Uri>[];
      final paid = ['paid', 'revoked', 'offline'].contains(scenario);
      final auth = scenario == 'guest'
          ? const AuthenticationState.guest()
          : AuthenticationState.authenticated(
              AuthenticatedUser(
                id: 'buyer',
                displayName: 'Buyer',
                email: 'buyer@example.com',
                isPremium: paid,
                isLifetime: paid,
              ),
            );
      final job = JapanJob.fromJson({
        'id': 'one',
        'status': 'active',
        'applyUrl': 'https://example.com/apply',
      });
      String? returnPath;
      final router = GoRouter(
        initialLocation: '/jobs/one',
        routes: [
          GoRoute(
            path: '/jobs/one',
            builder: (_, __) => Scaffold(body: JobApplyButton(job: job)),
          ),
          GoRoute(
            path: '/premium',
            builder: (_, state) {
              returnPath = state.uri.queryParameters['from'];
              return const Scaffold(body: Text('Premium destination'));
            },
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (_) => AuthController.test(auth),
            ),
            jobApplicationAccessProvider.overrideWithValue(() async {
              checks++;
              if (scenario == 'offline') throw Exception('offline');
              return scenario == 'paid';
            }),
            jobApplicationLauncherProvider.overrideWithValue((uri) async {
              launched.add(uri);
              return true;
            }),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('apply-job')));
      await tester.pumpAndSettle();
      expect(checks, paid ? 1 : 0);
      expect(launched.length, scenario == 'paid' ? 1 : 0);
      if (['guest', 'free', 'revoked'].contains(scenario)) {
        expect(find.text('Premium destination'), findsOneWidget);
        expect(returnPath, '/jobs/one');
      }
      if (scenario == 'offline') {
        expect(
          find.textContaining('Akses premium belum dapat diverifikasi'),
          findsOneWidget,
        );
      }
    });
  }
  test('benefits preserve explicit source text without inventing perks', () {
    expect(
      JapanJob.fromJson({
        'benefits': ['Paid leave', 'Commuting allowance'],
      }).benefits,
      'Paid leave\nCommuting allowance',
    );
    expect(
      JapanJob.fromJson({
        'description':
            'Benefits\nHealth insurance\nTraining budget\n\nRequirements\nN2',
      }).benefits,
      'Health insurance\nTraining budget',
    );
    expect(
      JapanJob.fromJson({
        'description': 'We develop employee benefits software.',
      }).benefits,
      isEmpty,
    );
  });
}
