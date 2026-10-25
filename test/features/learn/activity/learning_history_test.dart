import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/activity/continue_learning_section.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:shared_preferences/shared_preferences.dart';

class FakeHistory extends LearningHistoryRepository {
  FakeHistory() : super(Dio());
  List<LearningActivity> items = [];
  bool offline = false;
  Completer<List<LearningActivity>>? pendingFetch;
  Completer<LearningActivity>? pendingSave;
  final saves = <LearningActivity>[];
  @override
  Future<List<LearningActivity>> fetch() async {
    if (pendingFetch != null) return pendingFetch!.future;
    if (offline) throw StateError('offline');
    return items;
  }

  @override
  Future<LearningActivity> save(LearningActivity item) async {
    if (offline) throw StateError('offline');
    saves.add(item);
    final saved =
        await (pendingSave?.future ?? Future.value(item.copy(pending: false)));
    items = [saved, ...items.where((e) => e.key != item.key)];
    return saved;
  }
}

LearningActivity activity(
  int id, {
  String level = 'N5',
  bool completed = false,
}) => LearningActivity(
  kind: 'grammar',
  lessonId: 'grammar-${level.toLowerCase()}-$id',
  title: 'Materi $id',
  level: level,
  position: 1,
  total: 4,
  accessedAt: DateTime.now().toUtc().subtract(
    Duration(days: 1, minutes: 10 - id),
  ),
  completed: completed,
);

void main() {
  late SharedPreferences preferences;
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    preferences = await SharedPreferences.getInstance();
  });
  test(
    'offline positions persist per account and synchronize after reconnect',
    () async {
      final repository = FakeHistory()..offline = true;
      final first = LearningHistoryController(preferences, repository, 'a');
      await first.open(
        kind: 'grammar',
        lessonId: 'grammar-n5-1',
        title: 'Materi 1',
        total: 4,
        completed: false,
      );
      first.position('grammar', 'grammar-n5-1', 2);
      await Future<void>.delayed(Duration.zero);
      first.dispose();
      final restored = LearningHistoryController(preferences, repository, 'a');
      final other = LearningHistoryController(preferences, repository, 'b');
      addTearDown(restored.dispose);
      addTearDown(other.dispose);
      expect(restored.state.single.position, 2);
      expect(other.state, isEmpty);
      repository.offline = false;
      await restored.refresh();
      expect(repository.saves.single.position, 2);
      expect(restored.state.single.pending, isFalse);
      final index = await restored.open(
        kind: 'grammar',
        lessonId: 'grammar-n5-1',
        title: 'Materi 1',
        total: 4,
        completed: false,
      );
      expect(index, 2);
      expect(restored.state.length, 1);
    },
  );
  test(
    'opening a lesson does not wait for a stalled history request',
    () async {
      final repository = FakeHistory()..pendingFetch = Completer();
      final controller = LearningHistoryController(
        preferences,
        repository,
        'a',
      );
      addTearDown(controller.dispose);

      final position = await controller
          .open(
            kind: 'grammar',
            lessonId: 'grammar-n5-1',
            title: 'Materi 1',
            total: 4,
            completed: false,
          )
          .timeout(const Duration(milliseconds: 100));
      expect(position, 0);
      expect(controller.state.single.lessonId, 'grammar-n5-1');

      final refreshing = controller.refresh();
      repository.pendingFetch!.complete([]);
      await refreshing;
    },
  );
  test(
    'newer remote position wins; smaller updated material clamps local index',
    () async {
      final repository = FakeHistory()..items = [activity(1)];
      final controller = LearningHistoryController(
        preferences,
        repository,
        'a',
      );
      addTearDown(controller.dispose);
      await controller.refresh();
      expect(controller.state.single.position, 1);
      repository.offline = true;
      controller.position('grammar', 'grammar-n5-1', 3);
      repository
        ..items = [
          LearningActivity(
            kind: 'grammar',
            lessonId: 'grammar-n5-1',
            title: 'Updated',
            level: 'N5',
            position: 0,
            total: 2,
            accessedAt: DateTime.utc(2020),
          ),
        ]
        ..offline = false;
      await controller.refresh();
      expect(repository.saves.single.position, 1);
      expect(controller.state.single.title, 'Updated');
      repository.items = [
        controller.state.single.copy(
          position: 0,
          accessedAt: DateTime.now().add(const Duration(minutes: 1)),
        ),
      ];
      await controller.refresh();
      expect(controller.state.single.position, 0);
    },
  );
  test(
    'late server acknowledgement does not overwrite a newer local position',
    () async {
      final repository = FakeHistory()..items = [activity(1)];
      final controller = LearningHistoryController(
        preferences,
        repository,
        'a',
      );
      addTearDown(controller.dispose);
      await controller.refresh();
      controller.position('grammar', 'grammar-n5-1', 2);
      repository.pendingSave = Completer();
      final syncing = controller.refresh();
      await Future<void>.delayed(Duration.zero);
      controller.position('grammar', 'grammar-n5-1', 3);
      repository.pendingSave!.complete(
        repository.saves.single.copy(pending: false),
      );
      await syncing;
      expect(controller.state.single.position, 3);
      expect(controller.state.single.pending, isTrue);
    },
  );
  test(
    'disposed lesson does not replace valid history; completed review starts at zero',
    () async {
      final repository = FakeHistory()..items = [activity(1, completed: true)];
      final controller = LearningHistoryController(
        preferences,
        repository,
        'a',
      );
      addTearDown(controller.dispose);
      await controller.refresh();
      await controller.open(
        kind: 'grammar',
        lessonId: 'grammar-n5-2',
        title: 'Abandoned',
        total: 4,
        completed: false,
        isActive: () => false,
      );
      expect(controller.state.single.lessonId, 'grammar-n5-1');
      expect(
        await controller.open(
          kind: 'grammar',
          lessonId: 'grammar-n5-1',
          title: 'Materi 1',
          total: 4,
          completed: true,
        ),
        0,
      );
      expect(controller.state.single.completed, isTrue);
    },
  );
  testWidgets(
    'home section shows one primary and three recent lessons and navigates directly',
    (tester) async {
      tester.view.physicalSize = const Size(360, 800);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = FakeHistory()..items = List.generate(5, activity);
      final history = LearningHistoryController(preferences, repository, 'a');
      await history.refresh();
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) => const Scaffold(
              body: SingleChildScrollView(
                padding: EdgeInsets.all(20),
                child: ContinueLearningSection(),
              ),
            ),
          ),
          GoRoute(
            path: '/learn/grammar/lessons/:id',
            builder: (_, s) =>
                Scaffold(body: Text('Opened ${s.pathParameters['id']}')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            learningHistoryProvider.overrideWith((ref) => history),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                const AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'a',
                    displayName: 'User',
                    email: 'a@example.com',
                  ),
                ),
              ),
            ),
          ],
          child: MaterialApp.router(routerConfig: router),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Materi 4'), findsOneWidget);
      expect(find.text('Materi 0'), findsNothing);
      expect(find.byType(ListTile), findsNWidgets(3));
      await tester.tap(find.widgetWithText(FilledButton, 'Lanjutkan belajar'));
      await tester.pumpAndSettle();
      expect(find.text('Opened grammar-n5-4'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'empty history is hidden; completed and locked states remain accessible with large text',
    (tester) async {
      tester.view.physicalSize = const Size(320, 740);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = FakeHistory();
      final history = LearningHistoryController(preferences, repository, 'a');
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (_, __) => const Scaffold(
              body: SingleChildScrollView(
                padding: EdgeInsets.all(20),
                child: ContinueLearningSection(),
              ),
            ),
          ),
          GoRoute(
            path: '/premium',
            builder: (_, __) =>
                const Scaffold(body: Text('Lifetime destination')),
          ),
        ],
      );
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            learningHistoryProvider.overrideWith((ref) => history),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                const AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'a',
                    displayName: 'User',
                    email: 'a@example.com',
                  ),
                ),
              ),
            ),
          ],
          child: MaterialApp.router(
            routerConfig: router,
            builder: (context, child) => MediaQuery(
              data: MediaQuery.of(
                context,
              ).copyWith(textScaler: const TextScaler.linear(1.5)),
              child: child!,
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Lanjutkan belajar'), findsNothing);
      repository.items = [activity(1, completed: true)];
      await history.refresh();
      await tester.pumpAndSettle();
      expect(find.text('Pelajari lagi'), findsOneWidget);
      repository.items = [activity(2, level: 'N3')];
      await history.refresh();
      await tester.pumpAndSettle();
      expect(find.textContaining('Akses Lifetime diperlukan'), findsOneWidget);
      await tester.ensureVisible(find.text('Buka akses belajar'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Buka akses belajar'));
      await tester.pumpAndSettle();
      expect(find.text('Lifetime destination'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
}
