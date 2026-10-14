import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/exam_hub_screen.dart';

void main() {
  for (final premium in [false, true]) {
    testWidgets('compact hub supports narrow phones and N4 access: $premium',
        (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) =>
                const Scaffold(body: SafeArea(child: ExamHubScreen())),
          ),
          GoRoute(
            path: '/premium',
            builder: (_, __) =>
                const Scaffold(body: Text('Upgrade destination')),
          ),
          GoRoute(
            path: '/exams/n4/simulation',
            builder: (_, __) => const Scaffold(body: Text('N4 destination')),
          ),
          GoRoute(
            path: '/exams/jft/exam-1',
            builder: (_, __) => const Scaffold(body: Text('JFT destination')),
          ),
          GoRoute(
            path: '/exams/nat/simulation',
            builder: (_, __) => const Scaffold(body: Text('NAT destination')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'test',
                    displayName: 'Test',
                    email: 'test@example.com',
                    isPremium: premium,
                  ),
                ),
              ),
            ),
            diagnosticOverviewProvider.overrideWith(
              (ref) async => {'activeAttemptId': '', 'history': <Json>[]},
            ),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      // Both simulation choices are immediately accessible on a small phone.
      expect(find.text('Simulasi N5').hitTestable(), findsOneWidget);
      expect(find.text('Simulasi N4').hitTestable(), findsOneWidget);
      await tester.tap(find.text('NAT Test'));
      await tester.pumpAndSettle();
      expect(find.text('NAT 5Q Exam 1'), findsOneWidget);
      await tester.tap(find.text('NAT 5Q Exam 1'));
      await tester.pumpAndSettle();
      expect(
        find.text(premium ? 'NAT destination' : 'Upgrade destination'),
        findsOneWidget,
      );
      router.pop();
      await tester.pumpAndSettle();
      expect(find.text('Simulasi N5'), findsNothing);
      await tester.tap(find.text('JFT Test'));
      await tester.pumpAndSettle();
      expect(find.text('JFT Exam 1'), findsOneWidget);
      expect(find.text('NAT 5Q Exam 1'), findsNothing);
      await tester.tap(find.text('JFT Exam 1'));
      await tester.pumpAndSettle();
      expect(
        find.text(premium ? 'JFT destination' : 'Upgrade destination'),
        findsOneWidget,
      );
      router.pop();
      await tester.pumpAndSettle();
      await tester.tap(find.text('JLPT'));
      await tester.pumpAndSettle();
      expect(find.text('Simulasi N5'), findsOneWidget);
      await tester.scrollUntilVisible(find.text('Diagnostic N5'), 100);
      await tester.pumpAndSettle();
      expect(find.text('Diagnostic N5').hitTestable(), findsOneWidget);
      expect(tester.takeException(), isNull);
      // Larger accessibility text may scroll but must still lay out safely.
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'test',
                    displayName: 'Test',
                    email: 'test@example.com',
                    isPremium: premium,
                  ),
                ),
              ),
            ),
            diagnosticOverviewProvider.overrideWith(
              (ref) async => {'activeAttemptId': '', 'history': <Json>[]},
            ),
          ],
          child: MaterialApp.router(
            routerConfig: router,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(context)
                  .copyWith(textScaler: const TextScaler.linear(1.5)),
              child: child!,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.scrollUntilVisible(find.text('Simulasi N4'), -100);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Simulasi N4'));
      await tester.pumpAndSettle();
      expect(
        find.text(premium ? 'N4 destination' : 'Upgrade destination'),
        findsOneWidget,
      );
    });
  }
  for (final premium in [false, true]) {
    testWidgets('compact hub supports narrow phones and N3 access: $premium',
        (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) =>
                const Scaffold(body: SafeArea(child: ExamHubScreen())),
          ),
          GoRoute(
            path: '/premium',
            builder: (_, __) =>
                const Scaffold(body: Text('Upgrade destination')),
          ),
          GoRoute(
            path: '/exams/n3/simulation',
            builder: (_, __) => const Scaffold(body: Text('N3 destination')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'test',
                    displayName: 'Test',
                    email: 'test@example.com',
                    isPremium: premium,
                  ),
                ),
              ),
            ),
            diagnosticOverviewProvider.overrideWith(
              (ref) async => {'activeAttemptId': '', 'history': <Json>[]},
            ),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      // Both simulation choices are immediately accessible on a small phone.
      expect(find.text('Simulasi N5').hitTestable(), findsOneWidget);
      await tester.scrollUntilVisible(find.text('Simulasi N3'), 100);
      await tester.pumpAndSettle();
      expect(find.text('Simulasi N3').hitTestable(), findsOneWidget);
      expect(find.byType(Image), findsWidgets);
      await tester.scrollUntilVisible(find.text('Diagnostic N5'), 100);
      await tester.pumpAndSettle();
      expect(find.text('Diagnostic N5').hitTestable(), findsOneWidget);
      expect(tester.takeException(), isNull);
      // Larger accessibility text may scroll but must still lay out safely.
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'test',
                    displayName: 'Test',
                    email: 'test@example.com',
                    isPremium: premium,
                  ),
                ),
              ),
            ),
            diagnosticOverviewProvider.overrideWith(
              (ref) async => {'activeAttemptId': '', 'history': <Json>[]},
            ),
          ],
          child: MaterialApp.router(
            routerConfig: router,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(context)
                  .copyWith(textScaler: const TextScaler.linear(1.5)),
              child: child!,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.scrollUntilVisible(find.text('Simulasi N3'), -100);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Simulasi N3'));
      await tester.pumpAndSettle();
      expect(
        find.text(premium ? 'N3 destination' : 'Upgrade destination'),
        findsOneWidget,
      );
    });
  }
}
