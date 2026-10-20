import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

List<Map<String, dynamic>> _maps(Object? value) =>
    value is List ? value.whereType<Map<String, dynamic>>().toList() : [];

class MangaSeries {
  const MangaSeries({
    required this.id,
    required this.title,
    required this.japaneseTitle,
    required this.summary,
    required this.author,
    required this.coverUrl,
    required this.sourceUrl,
    required this.genres,
    required this.chapterCount,
    this.fetchedAt,
  });
  factory MangaSeries.fromJson(Map<String, dynamic> j) => MangaSeries(
    id: j['id'] as String,
    title: j['title'] as String,
    japaneseTitle: j['japaneseTitle'] as String? ?? '',
    summary: j['summary'] as String? ?? '',
    author: j['author'] as String? ?? '',
    coverUrl: j['coverUrl'] as String? ?? '',
    sourceUrl: j['sourceUrl'] as String,
    genres: (j['genres'] as List<dynamic>? ?? []).whereType<String>().toList(),
    chapterCount: j['chapterCount'] as int? ?? 0,
    fetchedAt: DateTime.tryParse(j['fetchedAt'] as String? ?? ''),
  );
  final String id;
  final String title;
  final String japaneseTitle;
  final String summary;
  final String author;
  final String coverUrl;
  final String sourceUrl;
  final List<String> genres;
  final int chapterCount;
  final DateTime? fetchedAt;
}

class MangaCatalog {
  const MangaCatalog({
    required this.items,
    required this.total,
    required this.page,
  });
  factory MangaCatalog.fromJson(Map<String, dynamic> j) => MangaCatalog(
    items: _maps(j['items']).map(MangaSeries.fromJson).toList(),
    total: j['total'] as int? ?? 0,
    page: j['page'] as int? ?? 1,
  );
  final List<MangaSeries> items;
  final int total;
  final int page;
}

class MangaPage {
  const MangaPage({
    required this.url,
    required this.width,
    required this.height,
  });
  factory MangaPage.fromJson(Map<String, dynamic> j) => MangaPage(
    url: j['url'] as String,
    width: j['width'] as int,
    height: j['height'] as int,
  );
  final String url;
  final int width;
  final int height;
}

class MangaChapter {
  const MangaChapter({
    required this.id,
    required this.seriesId,
    required this.title,
    required this.number,
    required this.sourceUrl,
    required this.status,
    required this.pages,
    this.retryAt,
  });
  factory MangaChapter.fromJson(Map<String, dynamic> j) => MangaChapter(
    id: j['id'] as String,
    seriesId: j['seriesId'] as String,
    title: j['title'] as String,
    number: j['number'] as String,
    sourceUrl: j['sourceUrl'] as String,
    status: j['status'] as String,
    pages: _maps(j['pages']).map(MangaPage.fromJson).toList(),
    retryAt: DateTime.tryParse(j['retryAt'] as String? ?? ''),
  );
  final String id;
  final String seriesId;
  final String title;
  final String number;
  final String sourceUrl;
  final String status;
  final List<MangaPage> pages;
  final DateTime? retryAt;
  bool get ready => status == 'ready' && pages.isNotEmpty;
  bool get pending => status == 'queued' || status == 'processing';
  bool get canPrepare =>
      !pending &&
      !ready &&
      (retryAt == null || !retryAt!.isAfter(DateTime.now()));
}

class MangaDetail {
  const MangaDetail({required this.series, required this.chapters});
  factory MangaDetail.fromJson(Map<String, dynamic> j) => MangaDetail(
    series: MangaSeries.fromJson(j),
    chapters: _maps(j['chapters']).map(MangaChapter.fromJson).toList(),
  );
  final MangaSeries series;
  final List<MangaChapter> chapters;
}

class MangaRepository {
  const MangaRepository(this.dio);
  final Dio dio;
  Future<Map<String, dynamic>> _get(String path) async {
    final res = await dio.get<Map<String, dynamic>>(path);
    return res.data!['data'] as Map<String, dynamic>;
  }

  Future<MangaCatalog> list({
    int page = 1,
    String q = '',
    String genre = '',
  }) async {
    final res = await dio.get<Map<String, dynamic>>(
      '/manga',
      queryParameters: {'page': page, 'limit': 20, 'q': q, 'genre': genre},
    );
    return MangaCatalog.fromJson(res.data!['data'] as Map<String, dynamic>);
  }

  Future<MangaDetail> detail(String id) async =>
      MangaDetail.fromJson(await _get('/manga/${Uri.encodeComponent(id)}'));
  Future<MangaChapter> chapter(String id) async => MangaChapter.fromJson(
    await _get('/manga/chapters/${Uri.encodeComponent(id)}'),
  );
  Future<MangaChapter> prepare(String id) async {
    final res = await dio.post<Map<String, dynamic>>(
      '/manga/chapters/${Uri.encodeComponent(id)}/prepare',
    );
    return MangaChapter.fromJson(res.data!['data'] as Map<String, dynamic>);
  }
}

// Local development runs the Manga service alongside the account API. Keep this
// default in code so a normal flutter run cannot lose a one-off --dart-define.
// Remote and release builds continue to use the main API unless overridden.
String resolveMangaApiBaseUrl({
  required String apiBaseUrl,
  required bool debug,
  String dedicatedBaseUrl = const String.fromEnvironment('MANGA_API_BASE_URL'),
}) {
  if (dedicatedBaseUrl.isNotEmpty) return dedicatedBaseUrl;
  final uri = Uri.tryParse(apiBaseUrl);
  const localHosts = {'10.0.2.2', 'localhost', '127.0.0.1', '::1'};
  if (debug &&
      uri != null &&
      uri.scheme == 'http' &&
      uri.port == 8080 &&
      localHosts.contains(uri.host)) {
    return uri.replace(port: 8081).toString();
  }
  return apiBaseUrl;
}

final mangaRepositoryProvider = Provider((ref) {
  final apiBase = ref.watch(apiBaseUrlProvider);
  final mangaBase = resolveMangaApiBaseUrl(
    apiBaseUrl: apiBase,
    debug: kDebugMode,
  );
  if (mangaBase == apiBase) return MangaRepository(ref.watch(dioProvider));
  final mainDio = ref.watch(dioProvider);
  final dio = Dio(mainDio.options.copyWith(baseUrl: mangaBase));
  // Forward the current session on every request, including after login/logout.
  dio.interceptors.clear();
  dio.interceptors.addAll(mainDio.interceptors);
  ref.onDispose(() => dio.close(force: true));
  return MangaRepository(dio);
});
typedef MangaFilter = ({int page, String q, String genre});
final mangaCatalogProvider = FutureProvider.autoDispose
    .family<MangaCatalog, MangaFilter>(
      (ref, f) => ref
          .watch(mangaRepositoryProvider)
          .list(page: f.page, q: f.q, genre: f.genre),
    );
final mangaDetailProvider = FutureProvider.autoDispose
    .family<MangaDetail, String>(
      (ref, id) => ref.watch(mangaRepositoryProvider).detail(id),
    );
