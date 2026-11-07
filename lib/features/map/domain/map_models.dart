import 'dart:math' as math;
import 'dart:ui';

/// Spot categories in display order (labels are localized in the UI).
const mapCategories = [
  'culture',
  'shrine',
  'nature',
  'food',
  'city',
  'history',
];

class LocalizedText {
  const LocalizedText({
    required this.text,
    required this.sourceLang,
    required this.manual,
  });
  factory LocalizedText.fromJson(Object? json) {
    final map = json is Map<String, dynamic> ? json : const <String, dynamic>{};
    return LocalizedText(
      text: '${map['text'] ?? ''}'.trim(),
      sourceLang: '${map['sourceLang'] ?? ''}',
      manual: map['manual'] == true,
    );
  }
  final String text;
  final String sourceLang;
  final bool manual;
  bool get isEmpty => text.isEmpty;
}

class MapPhrase {
  const MapPhrase({
    required this.japanese,
    required this.reading,
    required this.meaningId,
    required this.meaningEn,
  });
  factory MapPhrase.fromJson(Map<String, dynamic> json) => MapPhrase(
    japanese: '${json['japanese'] ?? ''}',
    reading: '${json['reading'] ?? ''}',
    meaningId: '${json['meaningId'] ?? ''}',
    meaningEn: '${json['meaningEn'] ?? ''}',
  );
  final String japanese;
  final String reading;
  final String meaningId;
  final String meaningEn;

  /// English users only see phrases that have an English meaning.
  String? meaningFor(String lang) => lang == 'en'
      ? (meaningEn.isEmpty ? null : meaningEn)
      : (meaningId.isEmpty ? null : meaningId);
}

class MapImage {
  const MapImage({
    required this.url,
    required this.author,
    required this.license,
    required this.licenseUrl,
    required this.sourceUrl,
  });
  final String url;
  final String author;
  final String license;
  final String licenseUrl;
  final String sourceUrl;
}

class MapSource {
  const MapSource({
    required this.kind,
    required this.lang,
    required this.url,
    required this.license,
  });
  final String kind;
  final String lang;
  final String url;
  final String license;
}

class MapSpot {
  const MapSpot({
    required this.id,
    required this.category,
    required this.prefecture,
    required this.lat,
    required this.lng,
    required this.name,
    required this.nameJa,
    required this.reading,
    required this.summary,
    required this.tips,
    required this.phrases,
    required this.image,
    required this.sources,
  });

  factory MapSpot.fromJson(Map<String, dynamic> json) {
    Map<String, dynamic> object(Object? value) =>
        value is Map<String, dynamic> ? value : const {};
    final names = object(json['name']);
    final summary = object(json['summary']);
    final tips = object(json['tips']);
    final image = json['image'];
    return MapSpot(
      id: '${json['id'] ?? ''}',
      category: '${json['category'] ?? ''}',
      prefecture: '${json['prefecture'] ?? ''}',
      lat: (json['lat'] as num?)?.toDouble() ?? 0,
      lng: (json['lng'] as num?)?.toDouble() ?? 0,
      name: {
        'id': '${names['id'] ?? ''}',
        'en': '${names['en'] ?? ''}',
      },
      nameJa: '${json['nameJa'] ?? ''}',
      reading: '${json['reading'] ?? ''}',
      summary: {
        for (final lang in const ['id', 'en'])
          lang: LocalizedText.fromJson(summary[lang]),
      },
      tips: {
        for (final lang in const ['id', 'en'])
          lang: LocalizedText.fromJson(tips[lang]),
      },
      phrases: [
        for (final item in (json['phrases'] as List?) ?? const [])
          if (item is Map<String, dynamic>) MapPhrase.fromJson(item),
      ],
      image: image is Map<String, dynamic> && '${image['url'] ?? ''}'.isNotEmpty
          ? MapImage(
              url: '${image['url']}',
              author: '${image['author'] ?? ''}',
              license: '${image['license'] ?? ''}',
              licenseUrl: '${image['licenseUrl'] ?? ''}',
              sourceUrl: '${image['sourceUrl'] ?? ''}',
            )
          : null,
      sources: [
        for (final item in (json['sources'] as List?) ?? const [])
          if (item is Map<String, dynamic>)
            MapSource(
              kind: '${item['kind'] ?? ''}',
              lang: '${item['lang'] ?? ''}',
              url: '${item['url'] ?? ''}',
              license: '${item['license'] ?? ''}',
            ),
      ],
    );
  }

  final String id;
  final String category;
  final String prefecture;
  final double lat;
  final double lng;
  final Map<String, String> name;
  final String nameJa;
  final String reading;
  final Map<String, LocalizedText> summary;
  final Map<String, LocalizedText> tips;
  final List<MapPhrase> phrases;
  final MapImage? image;
  final List<MapSource> sources;

  String nameFor(String lang) {
    final value = name[lang == 'en' ? 'en' : 'id'] ?? '';
    return value.isNotEmpty ? value : (name['en'] ?? name['id'] ?? nameJa);
  }

  LocalizedText summaryFor(String lang) =>
      summary[lang == 'en' ? 'en' : 'id'] ??
      const LocalizedText(text: '', sourceLang: '', manual: false);

  LocalizedText tipsFor(String lang) =>
      tips[lang == 'en' ? 'en' : 'id'] ??
      const LocalizedText(text: '', sourceLang: '', manual: false);

  List<MapPhrase> phrasesFor(String lang) =>
      phrases.where((p) => p.meaningFor(lang) != null).toList();
}

class MapData {
  const MapData({required this.spots, this.fromCache = false});
  final List<MapSpot> spots;
  final bool fromCache;
}

class JobCounts {
  const JobCounts({required this.counts, this.fromCache = false});
  final Map<String, int> counts;
  final bool fromCache;
  int operator [](String prefecture) => counts[prefecture] ?? 0;
}

class PrefectureShape {
  const PrefectureShape({
    required this.code,
    required this.name,
    required this.nameJa,
    required this.capitalName,
    required this.capitalLat,
    required this.capitalLng,
    required this.polygons,
  });
  final String code;
  final String name;
  final String nameJa;
  final String capitalName;
  final double capitalLat;
  final double capitalLng;

  /// Rings of [longitude, latitude] points.
  final List<List<Offset>> polygons;
}

/// Prefecture outlines from assets/map/japan_prefectures.json (generated by
/// scripts/build_japan_map.py) plus the projection used to draw them.
class JapanGeometry {
  JapanGeometry({
    required this.prefectures,
    required this.insetCode,
    required this.insetOffset,
  }) {
    var minX = double.infinity, minY = double.infinity;
    var maxX = -double.infinity, maxY = -double.infinity;
    for (final prefecture in prefectures) {
      for (final ring in prefecture.polygons) {
        for (final point in ring) {
          final p = _project(prefecture.code, point.dx, point.dy);
          minX = math.min(minX, p.dx);
          maxX = math.max(maxX, p.dx);
          minY = math.min(minY, p.dy);
          maxY = math.max(maxY, p.dy);
        }
      }
    }
    _bounds = Rect.fromLTRB(minX, minY, maxX, maxY);
    var insetBounds = Rect.zero;
    for (final prefecture in prefectures.where((p) => p.code == insetCode)) {
      for (final ring in prefecture.polygons) {
        for (final point in ring) {
          final p = _project(prefecture.code, point.dx, point.dy);
          final dot = Rect.fromLTWH(p.dx, p.dy, 0, 0);
          insetBounds = insetBounds == Rect.zero
              ? dot
              : insetBounds.expandToInclude(dot);
        }
      }
    }
    _insetBounds = insetBounds;
  }

  factory JapanGeometry.fromJson(Map<String, dynamic> json) {
    final inset = json['inset'] as Map<String, dynamic>? ?? const {};
    return JapanGeometry(
      insetCode: '${inset['code'] ?? ''}',
      insetOffset: Offset(
        (inset['dLon'] as num?)?.toDouble() ?? 0,
        (inset['dLat'] as num?)?.toDouble() ?? 0,
      ),
      prefectures: [
        for (final item in json['prefectures'] as List)
          PrefectureShape(
            code: '${item['code']}',
            name: '${item['name']}',
            nameJa: '${item['nameJa']}',
            capitalName: '${item['capital']['name']}',
            capitalLat: (item['capital']['lat'] as num).toDouble(),
            capitalLng: (item['capital']['lng'] as num).toDouble(),
            polygons: [
              for (final ring in item['polygons'] as List)
                [
                  for (final point in ring as List)
                    Offset(
                      ((point as List)[0] as num).toDouble(),
                      (point[1] as num).toDouble(),
                    ),
                ],
            ],
          ),
      ],
    );
  }

  final List<PrefectureShape> prefectures;
  final String insetCode;

  /// Longitude/latitude shift applied to the inset prefecture (Okinawa).
  final Offset insetOffset;
  late final Rect _bounds;
  late final Rect _insetBounds;

  static final double _xScale = math.cos(36 * math.pi / 180);

  /// Equirectangular projection (longitude scaled at 36°N) in map units,
  /// with y growing downward. Okinawa is shifted into its inset.
  Offset _project(String prefecture, double lng, double lat) {
    if (prefecture == insetCode) {
      lng += insetOffset.dx;
      lat += insetOffset.dy;
    }
    return Offset(lng * _xScale, -lat);
  }

  /// Width / height of the whole drawing.
  double get aspectRatio => _bounds.width / _bounds.height;

  /// Converts a coordinate to a point inside a canvas of [size].
  Offset toCanvas(String prefecture, double lat, double lng, Size size) {
    final p = _project(prefecture, lng, lat);
    final scale = size.width / _bounds.width;
    return Offset(
      (p.dx - _bounds.left) * scale,
      (p.dy - _bounds.top) * scale,
    );
  }

  /// Frame drawn around the inset, in canvas coordinates.
  Rect insetFrame(Size size) {
    if (_insetBounds == Rect.zero) return Rect.zero;
    final scale = size.width / _bounds.width;
    return Rect.fromLTRB(
      (_insetBounds.left - _bounds.left) * scale,
      (_insetBounds.top - _bounds.top) * scale,
      (_insetBounds.right - _bounds.left) * scale,
      (_insetBounds.bottom - _bounds.top) * scale,
    ).inflate(10);
  }

  Map<String, Path> paths(Size size) => {
    for (final prefecture in prefectures)
      prefecture.code: Path()
        ..fillType = PathFillType.nonZero
        ..addPolygons(prefecture, size, this),
  };

  PrefectureShape? byCode(String code) {
    for (final prefecture in prefectures) {
      if (prefecture.code == code) return prefecture;
    }
    return null;
  }
}

extension on Path {
  void addPolygons(PrefectureShape p, Size size, JapanGeometry geometry) {
    for (final ring in p.polygons) {
      addPolygon([
        for (final point in ring)
          geometry.toCanvas(p.code, point.dy, point.dx, size),
      ], true);
    }
  }
}
