import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

/// A Japanese example line: text, reading and meaning.
class LevelListExample {
  const LevelListExample({
    required this.japanese,
    required this.reading,
    required this.meaning,
  });

  factory LevelListExample.fromJson(Map<String, dynamic> json) =>
      LevelListExample(
        japanese: json['japanese'] as String? ?? '',
        reading: json['reading'] as String? ?? '',
        meaning: json['meaning'] as String? ?? '',
      );

  final String japanese;
  final String reading;
  final String meaning;
}

/// One grammar point of a level, with the lesson it belongs to.
class GrammarListItem {
  const GrammarListItem({
    required this.lessonId,
    required this.lessonTitle,
    required this.lessonNumber,
    required this.pattern,
    required this.meaning,
    required this.formation,
    required this.explanation,
    required this.commonMistake,
    required this.examples,
  });

  final String lessonId;
  final String lessonTitle;
  final int lessonNumber;
  final String pattern;
  final String meaning;
  final String formation;
  final String explanation;
  final String commonMistake;
  final List<LevelListExample> examples;

  bool matches(String query) {
    final q = query.trim().toLowerCase();
    if (q.isEmpty) return true;
    return '$pattern $meaning $lessonTitle'.toLowerCase().contains(q);
  }
}

/// One kanji of a level, with the lesson it belongs to.
class KanjiListItem {
  const KanjiListItem({
    required this.lessonId,
    required this.lessonTitle,
    required this.lessonNumber,
    required this.character,
    required this.meaning,
    required this.onReading,
    required this.kunReading,
    required this.strokes,
    required this.mnemonic,
    required this.note,
    required this.examples,
  });

  final String lessonId;
  final String lessonTitle;
  final int lessonNumber;
  final String character;
  final String meaning;
  final String onReading;
  final String kunReading;
  final int strokes;
  final String mnemonic;
  final String note;
  final List<LevelListExample> examples;

  bool matches(String query) {
    final q = query.trim().toLowerCase();
    if (q.isEmpty) return true;
    return '$character $meaning $onReading $kunReading'.toLowerCase().contains(
      q,
    );
  }
}

enum LevelListKind {
  grammar('grammar', 'points', 'Grammar'),
  kanji('kanji', 'characters', 'Kanji'),

  /// A whole simulation paper (JLPT, JFT, NAT, SSW) with its answer key.
  simulation('simulation', '', 'Simulasi');

  const LevelListKind(this.path, this.listSegment, this.fileLabel);

  final String path;
  final String listSegment;
  final String fileLabel;

  String fileName(String level) => 'NihongoMaster-$fileLabel-$level.pdf';

  String pdfPath(String level) => this == simulation
      ? '/simulation-pdf/${Uri.encodeComponent(level)}'
      : '/$path/levels/$level/pdf';
}

/// A refusal from the PDF endpoint. `premium_required` means Lifetime only.
class LevelPdfException implements Exception {
  const LevelPdfException(this.code, this.message);

  final String code;
  final String message;

  static const premiumRequired = 'premium_required';
  static const network = 'network';

  @override
  String toString() => message;
}

List<LevelListExample> _examples(Object? value) => [
  if (value is List)
    for (final item in value)
      if (item is Map<String, dynamic>) LevelListExample.fromJson(item),
];

String _text(Map<String, dynamic> json, String key) =>
    json[key] as String? ?? '';

class LevelListRepository {
  LevelListRepository(this._dio);

  final Dio _dio;

  Future<List<Map<String, dynamic>>> _lessons(
    LevelListKind kind,
    String level,
  ) async {
    final response = await _dio.get<Map<String, dynamic>>(
      '/${kind.path}/levels/$level/${kind.listSegment}',
    );
    final data = response.data?['data'];
    final lessons = data is Map<String, dynamic> ? data['lessons'] : null;
    return [
      if (lessons is List)
        for (final lesson in lessons)
          if (lesson is Map<String, dynamic>) lesson,
    ];
  }

  Future<List<GrammarListItem>> grammar(String level) async {
    final lessons = await _lessons(LevelListKind.grammar, level);
    return [
      for (final (index, lesson) in lessons.indexed)
        for (final point in (lesson['points'] as List?) ?? const [])
          if (point is Map<String, dynamic>)
            GrammarListItem(
              lessonId: _text(lesson, 'id'),
              lessonTitle: _text(lesson, 'title'),
              lessonNumber: index + 1,
              pattern: _text(point, 'pattern'),
              meaning: _text(point, 'meaning'),
              formation: _text(point, 'formation'),
              explanation: _text(point, 'explanation'),
              commonMistake: _text(point, 'commonMistake'),
              examples: _examples(point['examples']),
            ),
    ];
  }

  Future<List<KanjiListItem>> kanji(String level) async {
    final lessons = await _lessons(LevelListKind.kanji, level);
    return [
      for (final (index, lesson) in lessons.indexed)
        for (final character in (lesson['characters'] as List?) ?? const [])
          if (character is Map<String, dynamic>)
            KanjiListItem(
              lessonId: _text(lesson, 'id'),
              lessonTitle: _text(lesson, 'title'),
              lessonNumber: index + 1,
              character: _text(character, 'character'),
              meaning: _text(character, 'meaning'),
              onReading: _text(character, 'onReading'),
              kunReading: _text(character, 'kunReading'),
              strokes: character['strokes'] as int? ?? 0,
              mnemonic: _text(character, 'mnemonic'),
              note: _text(character, 'note'),
              examples: _examples(character['examples']),
            ),
    ];
  }

  /// Downloads the Lifetime-only PDF for a level.
  Future<Uint8List> pdf(LevelListKind kind, String level) async {
    try {
      final response = await _dio.get<List<int>>(
        kind.pdfPath(level),
        options: Options(
          responseType: ResponseType.bytes,
          receiveTimeout: const Duration(seconds: 60),
          headers: {'Accept': 'application/pdf'},
        ),
      );
      return Uint8List.fromList(response.data ?? const []);
    } on DioException catch (error) {
      final status = error.response?.statusCode;
      if (status == 402) {
        throw const LevelPdfException(
          LevelPdfException.premiumRequired,
          'Download PDF khusus pengguna Lifetime Access.',
        );
      }
      if (error.response == null) {
        throw const LevelPdfException(
          LevelPdfException.network,
          'Butuh koneksi internet',
        );
      }
      throw const LevelPdfException(
        'pdf_failed',
        'PDF belum dapat dibuat. Coba lagi.',
      );
    }
  }
}

final levelListRepositoryProvider = Provider<LevelListRepository>(
  (ref) => LevelListRepository(ref.watch(dioProvider)),
);

final grammarLevelListProvider = FutureProvider.autoDispose
    .family<List<GrammarListItem>, String>(
      (ref, level) => ref.watch(levelListRepositoryProvider).grammar(level),
    );

final kanjiLevelListProvider = FutureProvider.autoDispose
    .family<List<KanjiListItem>, String>(
      (ref, level) => ref.watch(levelListRepositoryProvider).kanji(level),
    );
