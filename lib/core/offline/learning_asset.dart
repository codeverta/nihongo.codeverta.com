import 'dart:async';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/download_preferences.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';

/// Only public learning illustrations use this cache. Profile photos and manga
/// pages must not be placed in the shared illustration namespace.
final learningAssetProvider = FutureProvider.autoDispose.family<Uint8List?, String>((
  ref,
  url,
) async {
  final offline = ref.watch(offlineProvider);
  final unmetered = ref.watch(unmeteredConnectionProvider);
  final allowLarge = ref.watch(allowLargeMobileDownloadsProvider);
  var disposed = false;
  ref.onDispose(() => disposed = true);
  PersistentLearningCache? cache;
  CachedAsset? saved;
  try {
    final openedCache = await ref.watch(learningCacheProvider.future);
    cache = openedCache;
    saved = await openedCache.readAsset(userId: guestCacheAccountId, url: url);
  } on Object {
    // Secure storage/cache availability must not block online illustrations.
  }
  if (disposed) return null;
  if (offline) return saved?.bytes;

  final cancel = CancelToken();
  final dio = Dio(
    BaseOptions(
      connectTimeout: const Duration(seconds: 5),
      receiveTimeout: const Duration(seconds: 8),
    ),
  );
  ref.onDispose(() {
    cancel.cancel();
    dio.close(force: true);
  });

  Future<Uint8List?> fetch() async {
    try {
      // A single small illustration is allowed on mobile data. Large images
      // stop early unless Wi-Fi or the explicit mobile-download setting allows.
      final maxBytes = unmetered || allowLarge ? 4 * 1024 * 1024 : 512 * 1024;
      final response = await dio.get<List<int>>(
        url,
        options: Options(responseType: ResponseType.bytes),
        cancelToken: cancel,
        onReceiveProgress: (received, total) {
          if (received > maxBytes || total > maxBytes) {
            cancel.cancel('Menunggu Wi-Fi untuk gambar besar.');
          }
        },
      );
      final raw = response.data;
      if (raw == null || raw.length > maxBytes) return saved?.bytes;
      final bytes = Uint8List.fromList(raw);
      try {
        await cache?.writeAsset(
          userId: guestCacheAccountId,
          url: url,
          bytes: bytes,
          contentType: response.headers.value('content-type') ?? 'image/webp',
        );
      } on Object {
        // Still display the successfully fetched illustration.
      }
      return bytes;
    } on Object {
      return saved?.bytes;
    }
  }

  if (saved != null) {
    if (unmetered &&
        DateTime.now().difference(saved.updatedAt) > learningCacheFreshness) {
      unawaited(fetch());
    }
    return saved.bytes;
  }
  return fetch();
});
