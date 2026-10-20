import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

class NewsToken {
  const NewsToken({
    required this.start,
    required this.end,
    required this.surface,
    required this.base,
    required this.reading,
    required this.readingSource,
  });
  factory NewsToken.fromJson(Map<String, dynamic> json) => NewsToken(
    start: json['start'] as int,
    end: json['end'] as int,
    surface: json['surface'] as String,
    base: json['base'] as String,
    reading: json['reading'] as String? ?? '',
    readingSource: json['readingSource'] as String? ?? 'unknown',
  );
  final int start;
  final int end;
  final String surface;
  final String base;
  final String reading;
  final String readingSource;
}

class NewsParagraph {
  const NewsParagraph({
    required this.text,
    required this.tokens,
    this.kind = 'text',
  });
  factory NewsParagraph.fromJson(Map<String, dynamic> json) => NewsParagraph(
    text: json['text'] as String,
    kind: json['kind'] as String? ?? 'text',
    tokens: newsMaps(json['tokens']).map(NewsToken.fromJson).toList(),
  );
  final String text;
  final String kind;
  final List<NewsToken> tokens;
}

class NewsArticle {
  NewsArticle(this.json);
  final Map<String, dynamic> json;
  String get id => json['id'] as String;
  String get dateLabel =>
      json['dateType'] == 'source' ? 'Tanggal sumber' : 'Terbit';
  String get title => json['title'] as String;
  String get summary => json['summary'] as String? ?? '';
  String get source => json['sourceName'] as String;
  String get kind => json['kind'] as String;
  String get kindLabel =>
      kind == 'easy' ? 'Info Jepang · Easy' : 'Berita · Reguler';
  String get url => json['url'] as String;
  String get license => json['license'] as String;
  String get licenseUrl => json['licenseUrl'] as String;
  String get changes => json['changes'] as String? ?? '';
  String get imageSourceUrl => json['imageSourceUrl'] as String? ?? '';
  String get imageLicenseUrl => json['imageLicenseUrl'] as String? ?? '';
  String get imageUrl => json['imageUrl'] as String? ?? '';
  String get imageCredit => json['imageCredit'] as String? ?? '';
  String get imageLicense => json['imageLicense'] as String? ?? '';
  DateTime get publishedAt => DateTime.parse(json['publishedAt'] as String);
  DateTime get sourceDate {
    final original = json['originalDate'] as String?;
    if (original != null && original.isNotEmpty) {
      return DateTime.parse('${original}T00:00:00Z');
    }
    final japan = publishedAt.toUtc().add(const Duration(hours: 9));
    return DateTime.utc(japan.year, japan.month, japan.day);
  }

  DateTime get fetchedAt => DateTime.parse(json['fetchedAt'] as String);
  List<NewsToken> get titleTokens =>
      newsMaps(json['titleTokens']).map(NewsToken.fromJson).toList();
  List<NewsParagraph> get paragraphs =>
      newsMaps(json['paragraphs']).map(NewsParagraph.fromJson).toList();
  List<Map<String, dynamic>> get credits => newsMaps(json['credits']);
  List<Map<String, dynamic>> get references => newsMaps(json['references']);
}

List<Map<String, dynamic>> newsMaps(Object? json) =>
    json is List ? json.whereType<Map<String, dynamic>>().toList() : [];

class NewsPage {
  const NewsPage({
    required this.items,
    required this.total,
    required this.page,
    required this.sources,
  });
  factory NewsPage.fromJson(Map<String, dynamic> json) => NewsPage(
    items: newsMaps(json['items']).map(NewsArticle.new).toList(),
    total: json['total'] as int,
    page: json['page'] as int,
    sources: newsMaps(json['sources']),
  );
  final List<NewsArticle> items;
  final List<Map<String, dynamic>> sources;
  final int total;
  final int page;
}

class NewsRepository {
  NewsRepository(this.dio);
  final Dio dio;
  Future<NewsPage> list({
    int page = 1,
    int limit = 20,
    String source = '',
  }) async {
    final res = await dio.get<Map<String, dynamic>>(
      '/news',
      queryParameters: {'page': page, 'limit': limit, 'source': source},
    );
    return NewsPage.fromJson(res.data!['data'] as Map<String, dynamic>);
  }

  Future<NewsArticle> detail(String id) async {
    final res = await dio.get<Map<String, dynamic>>(
      '/news/${Uri.encodeComponent(id)}',
    );
    return NewsArticle(res.data!['data'] as Map<String, dynamic>);
  }
}

final newsRepositoryProvider = Provider(
  (ref) => NewsRepository(ref.watch(dioProvider)),
);
final newsPreviewProvider = FutureProvider.autoDispose(
  (ref) => ref.watch(newsRepositoryProvider).list(limit: 3),
);
final newsPageProvider = FutureProvider.autoDispose
    .family<NewsPage, ({int page, String source})>(
      (ref, filter) => ref
          .watch(newsRepositoryProvider)
          .list(page: filter.page, source: filter.source),
    );
final newsArticleProvider = FutureProvider.autoDispose
    .family<NewsArticle, String>(
      (ref, id) => ref.watch(newsRepositoryProvider).detail(id),
    );
