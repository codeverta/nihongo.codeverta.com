import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/device_timezone.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/learning_cache.dart';
import 'package:nihongo_master/core/offline/learning_cache_persistence_memory.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test(
    'API requests carry selected language and Google auth carries IANA zone',
    () async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          offlineAccessStoreProvider.overrideWithValue(
            OfflineAccessStore(
              preferences,
              secretStorage: _MemorySecretStorage(),
            ),
          ),
          learningCacheProvider.overrideWith(
            (ref) => PersistentLearningCache.open(
              persistence: MemoryLearningCachePersistence(),
            ),
          ),
          apiBaseUrlProvider.overrideWithValue('https://api.example.com/v1'),
          deviceLocaleProvider.overrideWithValue(const Locale('en')),
          englishAvailabilityProvider.overrideWithValue(true),
          deviceTimeZoneIdentifierProvider.overrideWith(
            (ref) async => 'America/Los_Angeles',
          ),
        ],
      );
      final adapter = _RequestCaptureAdapter();
      final dio = container.read(dioProvider)..httpClientAdapter = adapter;
      addTearDown(() {
        container.dispose();
        dio.close(force: true);
      });

      await dio.get<void>('/catalog');
      await dio.post<void>('/auth/google', data: {'idToken': 'test'});
      await container.read(appLanguageProvider.notifier).selectLanguage('id');
      await dio.get<void>('/catalog');

      expect(_header(adapter.requests[0], 'Accept-Language'), 'en');
      expect(_header(adapter.requests[1], 'Accept-Language'), 'en');
      expect(
        _header(adapter.requests[1], 'X-Device-Timezone'),
        'America/Los_Angeles',
      );
      expect(_header(adapter.requests[2], 'Accept-Language'), 'id');
      expect(_header(adapter.requests[2], 'X-Device-Timezone'), isNull);
    },
  );
}

String? _header(RequestOptions request, String name) {
  for (final entry in request.headers.entries) {
    if (entry.key.toLowerCase() == name.toLowerCase()) {
      return entry.value as String?;
    }
  }
  return null;
}

class _RequestCaptureAdapter implements HttpClientAdapter {
  final requests = <RequestOptions>[];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return ResponseBody.fromString('{}', 200);
  }

  @override
  void close({bool force = false}) {}
}

class _MemorySecretStorage implements OfflineSecretStorage {
  @override
  Future<String?> read(String key) async => null;

  @override
  Future<void> write({required String key, required String value}) async {}

  @override
  Future<void> delete(String key) async {}
}
