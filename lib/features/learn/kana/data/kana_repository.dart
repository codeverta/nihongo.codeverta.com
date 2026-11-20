import 'dart:convert';

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/learn/kana/domain/stroke_matcher.dart';

/// Stroke order data derived from KanjiVG (CC BY-SA 3.0).
const kanaStrokesAsset = 'assets/content/kana/kana_strokes.json';

/// Reference strokes per character, in drawing order, in the unit square.
typedef KanaStrokeData = Map<String, List<Stroke>>;

KanaStrokeData parseKanaStrokes(String source) {
  final json = jsonDecode(source) as Map<String, dynamic>;
  final box = (json['viewBox'] as num).toDouble();
  final strokes = json['strokes'] as Map<String, dynamic>;
  return {
    for (final MapEntry(:key, :value) in strokes.entries)
      key: [
        for (final flat in value as List)
          [
            for (var i = 0; i + 1 < (flat as List).length; i += 2)
              Offset(
                (flat[i] as num).toDouble() / box,
                (flat[i + 1] as num).toDouble() / box,
              ),
          ],
      ],
  };
}

final kanaStrokesProvider = FutureProvider<KanaStrokeData>((ref) async {
  final source = await rootBundle.loadString(kanaStrokesAsset);
  return parseKanaStrokes(source);
});

const _practicedKey = 'kana_practiced_v1';

/// Characters the learner has written from memory without help.
class KanaProgressNotifier extends StateNotifier<Set<String>> {
  KanaProgressNotifier(super.initial, this._write);

  final Future<void> Function(Set<String>) _write;

  Future<void> markPracticed(String char) async {
    if (state.contains(char)) return;
    state = {...state, char};
    await _write(state);
  }
}

final kanaProgressProvider =
    StateNotifierProvider<KanaProgressNotifier, Set<String>>((ref) {
      final preferences = ref.watch(sharedPreferencesProvider);
      return KanaProgressNotifier({
        ...?preferences.getStringList(_practicedKey),
      }, (chars) => preferences.setStringList(_practicedKey, chars.toList()));
    });
