import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/jobs/data/jobs_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/jobs/presentation/jobs_screens.dart';

Map<String, dynamic> _job(String id, {String status = 'active'}) => {
  'id': id,
  'title': 'Care worker $id',
  'company': 'Example employer',
  'category': 'nursing-care',
  'prefecture': '13',
  'location': 'Tokyo',
  'sector': 'nursing-care',
  'employmentType': 'full-time',
  'remote': 'onsite',
  'visa': 'unknown',
  'japaneseLevel': 'N4',
  'salaryMin': 250000,
  'salaryMax': 300000,
  'salaryPeriod': 'month',
  'salaryCurrency': 'JPY',
  'description': 'Support residents with daily activities.',
  'qualifications': 'Japanese communication and relevant training.',
  'sourceName': 'Example source',
  'sourceUrl': 'https://example.com/job',
  'applyUrl': 'https://example.com/apply',
  'lastSeenAt': '2026-10-18T00:00:00Z',
  'status': status,
};

class _Repository extends JobsRepository {
  _Repository() : super(Dio());
  final requests = <Map<String, String>>[];
  Completer<JobPage>? pending;
  bool fail = false;
  @override
  Future<JobPage> list(Map<String, String> filters, {String? cursor}) async {
    requests.add({...filters, if (cursor != null) 'cursor': cursor});
    if (pending != null) return pending!.future;
    if (fail) throw const JobsFailure('Test connection failure');
    return JobPage.fromJson({
      'items': [_job(cursor == null ? 'one' : 'two')],
      'total': 2,
      'hasMore': cursor == null,
      'nextCursor': cursor == null ? 'page2' : null,
    });
  }

  @override
  Future<JobMeta> meta() async => JobMeta.fromJson({
    'categories': [
      {'id': 'nursing-care', 'label': 'Kaigo'},
    ],
    'sectors': [
      {'id': 'nursing-care', 'label': 'Kaigo SSW'},
    ],
    'prefectures': [
      {'id': '13', 'label': 'Tokyo'},
    ],
    'employmentTypes': [
      {'id': 'full-time', 'label': 'Penuh waktu'},
    ],
    'japaneseLevels': [
      {'id': 'N4', 'label': 'N4'},
    ],
    'visaOptions': [
      {'id': 'unknown', 'label': 'Belum dicantumkan'},
    ],
  });

  @override
  Future<JapanJob> detail(String id) async =>
      JapanJob.fromJson(_job(id, status: 'inactive'));
}

void main() {
  testWidgets('job search and filters use English app labels', (tester) async {
    final repository = _Repository();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [jobsRepositoryProvider.overrideWithValue(repository)],
        child: const MaterialApp(
          locale: Locale('en'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: JobsScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Work in Japan'), findsOneWidget);
    expect(find.text('Care worker one'), findsOneWidget);
    expect(find.byTooltip('Filter jobs'), findsOneWidget);
    await tester.tap(find.byTooltip('Filter jobs'));
    await tester.pumpAndSettle();
    expect(find.text('Find the right fit'), findsOneWidget);
    expect(find.text('Job category'), findsOneWidget);
    expect(find.text('Minimum salary (JPY)'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Show jobs'),
      180,
      scrollable: find
          .descendant(
            of: find.byType(ListView).last,
            matching: find.byType(Scrollable),
          )
          .first,
    );
    expect(find.text('Show jobs'), findsOneWidget);
  });

  testWidgets('late response cannot replace newer search results', (
    tester,
  ) async {
    final pending = Completer<JobPage>();
    final repository = _Repository()..pending = pending;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [jobsRepositoryProvider.overrideWithValue(repository)],
        child: const MaterialApp(home: JobsScreen()),
      ),
    );
    await tester.pump();
    repository.pending = null;
    await tester.enterText(find.byType(TextField), 'new search');
    await tester.pump(const Duration(milliseconds: 500));
    await tester.pumpAndSettle();
    expect(find.text('Care worker one'), findsOneWidget);
    pending.complete(
      JobPage.fromJson({
        'items': [_job('stale')],
        'total': 1,
        'hasMore': false,
      }),
    );
    await tester.pumpAndSettle();
    expect(find.text('Care worker stale'), findsNothing);
    expect(find.text('Care worker one'), findsOneWidget);
  });

  test(
    'unknown fields stay unknown, salary period and external URI validated',
    () {
      final unknown = JapanJob.fromJson({'id': '1', 'title': 'Job'});
      expect(unknown.salary, 'Gaji belum dicantumkan');
      expect(unknown.visa, 'unknown');
      expect(unknown.japaneseLevel, 'unknown');
      expect(unknown.active, isFalse);
      expect(
        JapanJob.fromJson(_job('1')).salary,
        'JPY 250,000–300,000 / bulan',
      );
      for (final link in [
        'http://example.com',
        'javascript:alert(1)',
        'https://user:password@example.com',
      ]) {
        expect(jobExternalUri(link), isNull);
      }
      expect(jobExternalUri('https://example.com/job'), isNotNull);
    },
  );

  testWidgets('debounced search, pagination and inactive detail', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(390, 844);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = _Repository();
    final router = GoRouter(
      initialLocation: '/jobs',
      routes: [
        GoRoute(path: '/jobs', builder: (_, __) => const JobsScreen()),
        GoRoute(
          path: '/jobs/:id',
          builder: (_, state) =>
              JobDetailScreen(id: state.pathParameters['id']!),
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [jobsRepositoryProvider.overrideWithValue(repository)],
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Care worker one'), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'Tokyo');
    await tester.pump(const Duration(milliseconds: 200));
    expect(repository.requests.length, 1);
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pumpAndSettle();
    expect(repository.requests.last['q'], 'Tokyo');
    await tester.scrollUntilVisible(
      find.text('Care worker two'),
      150,
      scrollable: find
          .descendant(
            of: find.byType(ListView).last,
            matching: find.byType(Scrollable),
          )
          .first,
    );
    // Scrolling near the footer triggers the real pagination listener.
    await tester.pumpAndSettle();
    expect(repository.requests.last['cursor'], 'page2');
    expect(find.text('Care worker two'), findsOneWidget);
    await tester.ensureVisible(find.text('Care worker two'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Care worker two'));
    await tester.pumpAndSettle();
    expect(find.text('Detail lowongan'), findsOneWidget);
    expect(find.text('Buka halaman asli'), findsNothing);
    await tester.scrollUntilVisible(
      find.text('Lihat & lamar di situs sumber'),
      200,
      scrollable: find
          .descendant(
            of: find.byType(ListView).last,
            matching: find.byType(Scrollable),
          )
          .first,
    );
    final apply = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Lihat & lamar di situs sumber'),
    );
    expect(apply.onPressed, isNull);
    expect(tester.takeException(), isNull);
  });

  testWidgets('retry and filters fit narrow phone with large text', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(320, 740);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = _Repository()..fail = true;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [jobsRepositoryProvider.overrideWithValue(repository)],
        child: MaterialApp(
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(
              context,
            ).copyWith(textScaler: const TextScaler.linear(1.4)),
            child: child!,
          ),
          home: const JobsScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Test connection failure'), findsOneWidget);
    repository.fail = false;
    await tester.ensureVisible(find.text('Coba lagi'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Coba lagi'));
    await tester.pumpAndSettle();
    expect(find.text('Test connection failure'), findsNothing);
    expect(
      repository.requests.where((r) => !r.containsKey('cursor')).length,
      2,
    );
    await tester.ensureVisible(find.byTooltip('Filter lowongan'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Filter lowongan'));
    await tester.pumpAndSettle();
    expect(find.text('Temukan yang sesuai'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Tampilkan lowongan'),
      200,
      scrollable: find
          .descendant(
            of: find.byType(ListView).last,
            matching: find.byType(Scrollable),
          )
          .first,
    );
    await tester.tap(find.text('Tampilkan lowongan'));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
