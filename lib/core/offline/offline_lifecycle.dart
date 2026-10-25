import 'dart:async';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/controllers/grammar_controller.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/controllers/kanji_controller.dart';
import 'package:nihongo_master/features/learn/offline/offline_practice.dart';
import 'package:nihongo_master/features/learn/reading/presentation/controllers/reading_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

class OfflineLifecycle extends ConsumerStatefulWidget {
  const OfflineLifecycle({required this.child, super.key});
  final Widget child;

  @override
  ConsumerState<OfflineLifecycle> createState() => _OfflineLifecycleState();
}

class _OfflineLifecycleState extends ConsumerState<OfflineLifecycle>
    with WidgetsBindingObserver {
  Timer? _timer;
  StreamSubscription<OfflinePracticeSyncEvent>? _events;
  bool _syncing = false;
  bool _foreground = true;
  Future<void> _cleanups = Future.value();
  DateTime? _lastBillingRefresh;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _timer = Timer.periodic(
      const Duration(minutes: 1),
      (_) => unawaited(_sync()),
    );
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _events = ref.read(offlinePracticeServiceProvider).syncEvents.listen((
        event,
      ) {
        if (!mounted ||
            ref.read(authControllerProvider).user?.id != event.userId) {
          return;
        }
        ref
          ..invalidate(homeDashboardProvider)
          ..invalidate(gamificationProvider)
          ..invalidate(vocabularyLevelsProvider)
          ..invalidate(grammarCourseProvider)
          ..invalidate(grammarCourseLevelProvider)
          ..invalidate(kanjiCourseProvider)
          ..invalidate(readingCourseProvider);
        unawaited(ref.read(learningHistoryProvider.notifier).refresh());
      });
      unawaited(_sync());
    });
  }

  Future<void> _clearAccount(String id) async {
    ref.read(learningDownloadsProvider).valueOrNull?.cancel();
    final practice = ref.read(offlinePracticeServiceProvider);
    final anki = ref.read(offlineAnkiReviewsProvider);
    final cacheFuture = ref.read(learningCacheProvider.future);
    await Future.wait([
      practice.clearAccount(id).catchError((Object _) {}),
      anki.clearAccount(id).catchError((Object _) {}),
      cacheFuture
          .then((cache) => cache.clearAccount(id))
          .catchError((Object _) {}),
    ]);
    // Each cleanup is attempted independently. A revoked identity prevents
    // access to private records even if filesystem cleanup fails.
  }

  Future<void> _sync() async {
    if (!mounted || !_foreground || _syncing) return;
    _syncing = true;
    try {
      await _cleanups;
      if (!mounted) return;
      final auth = ref.read(authControllerProvider.notifier);
      await auth.refreshOfflineEntitlement();
      if (!mounted) return;
      final user = ref.read(authControllerProvider).user?.id;
      if (user == null) return;
      final links = ref.read(connectivityProvider).valueOrNull;
      if (links?.contains(ConnectivityResult.none) ?? false) return;
      await auth.refreshUser();
      if (!mounted || ref.read(authControllerProvider).user?.id != user) return;
      if (_lastBillingRefresh == null ||
          DateTime.now().difference(_lastBillingRefresh!) >=
              const Duration(minutes: 5)) {
        _lastBillingRefresh = DateTime.now();
        await ref.read(billingControllerProvider.notifier).refresh();
      }
      if (!mounted ||
          ref.read(authControllerProvider).user?.id != user ||
          ref.read(offlineProvider)) {
        return;
      }
      await ref.read(offlinePracticeServiceProvider).flush();
      if (!mounted || ref.read(authControllerProvider).user?.id != user) return;
      await ref.read(offlineAnkiReviewsProvider).flush();
      if (!mounted) return;
      await ref.read(learningHistoryProvider.notifier).refresh();
      if (mounted) ref.invalidate(ankiDeckProvider);
    } on Object {
      // Durable outboxes remain owner-scoped and retry on reconnect/resume.
    } finally {
      _syncing = false;
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _foreground = state == AppLifecycleState.resumed;
    if (state == AppLifecycleState.resumed) unawaited(_sync());
  }

  @override
  Widget build(BuildContext context) {
    ref
      ..listen<String?>(authControllerProvider.select((s) => s.user?.id), (
        old,
        next,
      ) {
        if (old != null && old != next) {
          _lastBillingRefresh = null;
          _cleanups = _cleanups.then((_) => _clearAccount(old));
        }
        if (next != null && next != old) unawaited(_sync());
      })
      ..listen(connectivityProvider, (_, next) {
        final links = next.valueOrNull;
        if (links?.contains(ConnectivityResult.none) ?? false) {
          ref.read(learningDownloadsProvider).valueOrNull?.cancel();
        }
        if (links != null &&
            links.isNotEmpty &&
            !links.contains(ConnectivityResult.none)) {
          ref.read(backendReachableProvider.notifier).state = null;
          _lastBillingRefresh = null;
          unawaited(_sync());
        }
      });
    return widget.child;
  }

  @override
  void dispose() {
    _timer?.cancel();
    unawaited(_events?.cancel());
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }
}
