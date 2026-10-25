import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_downloads.dart';
import 'package:nihongo_master/core/offline/network_status.dart';

final learningCacheProvider = FutureProvider<PersistentLearningCache>((ref) {
  return PersistentLearningCache.open();
});

final learningDownloadsProvider = FutureProvider<LearningDownloads>((
  ref,
) async {
  final dio = ref.watch(dioProvider);
  final access = ref.watch(offlineAccessStoreProvider);
  final cache = await ref.watch(learningCacheProvider.future);
  final downloads = LearningDownloads(
    dio: dio,
    cache: cache,
    currentUserId: () async => access.userId,
    canReadPremium: access.canReadPremium,
    canStartDownload: () async => !ref.read(offlineProvider),
  );
  ref.onDispose(downloads.dispose);
  return downloads;
});

final learningDownloadProgressProvider =
    StreamProvider<LearningDownloadProgress>((ref) async* {
      final downloads = await ref.watch(learningDownloadsProvider.future);
      yield downloads.state;
      yield* downloads.progress;
    });

final learningCacheStatsProvider = StreamProvider<LearningCacheStats>((
  ref,
) async* {
  final cache = await ref.watch(learningCacheProvider.future);
  yield await cache.stats();
  yield* cache.statsChanges;
});
