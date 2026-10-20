import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/manga/data/manga_repository.dart';
import 'package:shared_preferences/shared_preferences.dart';

class _Repository extends MangaRepository {
  _Repository() : super(Dio());
  int calls = 0;
  @override
  Future<MangaCatalog> list({
    int page = 1,
    String q = '',
    String genre = '',
  }) async {
    calls++;
    return const MangaCatalog(items: [], total: 0, page: 1);
  }
}

class _Auth extends AuthController {
  _Auth(super.state) : super.test();
  void revoke() => state = const AuthenticationState.guest();
}

void main() {
  for (final path in ['/manga', '/manga/story', '/manga/chapters/chapter']) {
    testWidgets('Manga deep link $path requires lifetime before fetching', (
      tester,
    ) async {
      SharedPreferences.setMockInitialValues({'is_first_time': false});
      final prefs = await SharedPreferences.getInstance();
      final repo = _Repository();
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(prefs),
          authControllerProvider.overrideWith(
            (_) => AuthController.test(const AuthenticationState.guest()),
          ),
          mangaRepositoryProvider.overrideWithValue(repo),
        ],
      );
      addTearDown(container.dispose);
      final router = container.read(appRouterProvider)..go(path);
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      expect(router.routerDelegate.currentConfiguration.uri.path, '/premium');
      expect(
        router.routerDelegate.currentConfiguration.uri.queryParameters['from'],
        path,
      );
      expect(repo.calls, 0);
      await tester.pumpWidget(const SizedBox.shrink());
    });
  }

  testWidgets('lifetime opens Manga and losing access closes it immediately', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({'is_first_time': false});
    final prefs = await SharedPreferences.getInstance();
    final auth = _Auth(
      const AuthenticationState.authenticated(
        AuthenticatedUser(
          id: 'buyer',
          displayName: 'Buyer',
          email: 'buyer@example.com',
          isPremium: true,
          isLifetime: true,
        ),
      ),
    );
    final repo = _Repository();
    final container = ProviderContainer(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(prefs),
        authControllerProvider.overrideWith((_) => auth),
        mangaRepositoryProvider.overrideWithValue(repo),
      ],
    );
    addTearDown(container.dispose);
    final router = container.read(appRouterProvider)..go('/manga');
    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/manga');
    expect(repo.calls, 1);
    auth.revoke();
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/premium');
    await tester.pumpWidget(const SizedBox.shrink());
  });
}
