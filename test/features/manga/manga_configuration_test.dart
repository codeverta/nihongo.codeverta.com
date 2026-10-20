import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/manga/data/manga_repository.dart';

void main() {
  test('dedicated Manga client forwards the current account session', () async {
    var token = 'first-session';
    final mainDio = Dio(BaseOptions(baseUrl: 'http://10.0.2.2:8080/v1'));
    mainDio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (request, handler) {
          request.headers['Authorization'] = 'Bearer $token';
          handler.next(request);
        },
      ),
    );
    final container = ProviderContainer(
      overrides: [
        apiBaseUrlProvider.overrideWithValue(mainDio.options.baseUrl),
        dioProvider.overrideWithValue(mainDio),
      ],
    );
    addTearDown(container.dispose);
    final repo = container.read(mangaRepositoryProvider);
    final adapter = _SessionAdapter();
    repo.dio.httpClientAdapter = adapter;
    await repo.list();
    expect(adapter.authorization, 'Bearer first-session');
    expect(adapter.uri.port, 8081);
    token = 'new-session';
    await repo.list();
    expect(adapter.authorization, 'Bearer new-session');
  });
  test('ordinary local debug run reaches the standalone Manga service', () {
    final container = ProviderContainer(
      overrides: [
        apiBaseUrlProvider.overrideWithValue('http://10.0.2.2:8080/v1'),
        dioProvider.overrideWithValue(
          Dio(BaseOptions(baseUrl: 'http://10.0.2.2:8080/v1')),
        ),
      ],
    );
    addTearDown(container.dispose);

    expect(
      container.read(mangaRepositoryProvider).dio.options.baseUrl,
      'http://10.0.2.2:8081/v1',
    );
  });

  test('release and remote backends keep the main API', () {
    for (final debug in [true, false]) {
      expect(
        resolveMangaApiBaseUrl(
          apiBaseUrl: 'https://api.example.com/v1',
          debug: debug,
        ),
        'https://api.example.com/v1',
      );
    }
    expect(
      resolveMangaApiBaseUrl(
        apiBaseUrl: 'http://10.0.2.2:8080/v1',
        debug: false,
      ),
      'http://10.0.2.2:8080/v1',
    );
  });

  test('an explicit Manga endpoint always takes precedence', () {
    expect(
      resolveMangaApiBaseUrl(
        apiBaseUrl: 'http://10.0.2.2:8080/v1',
        debug: true,
        dedicatedBaseUrl: 'https://manga.example.com/v1',
      ),
      'https://manga.example.com/v1',
    );
  });

  test('local development supports desktop and iOS loopback too', () {
    for (final host in ['localhost', '127.0.0.1', '[::1]']) {
      expect(
        resolveMangaApiBaseUrl(apiBaseUrl: 'http://$host:8080/v1', debug: true),
        'http://$host:8081/v1',
      );
    }
  });
}

class _SessionAdapter implements HttpClientAdapter {
  String? authorization;
  late Uri uri;
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    authorization = options.headers['Authorization'] as String?;
    uri = options.uri;
    return ResponseBody.fromString(
      '{"data":{"items":[],"total":0,"page":1}}',
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
