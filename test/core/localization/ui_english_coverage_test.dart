import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// Guards English coverage: every static string passed to `.localized(context)`
/// must have an English translation, otherwise English users see Indonesian.
void main() {
  test('static localized() strings all have English copy', () {
    const english = AppLocalizations(Locale('en'));
    // Strings that are intentionally identical in both languages.
    const sameInEnglish = {
      'JLPT N5 Diagnostic',
      'NAT-TEST 5Q',
      'JFT-Basic',
      'LESSON',
      'Achievements',
      'N5, Wakayama, hiragana...',
      'Export deck',
      'Import deck',
      'Furigana',
      'Gamification',
      'Lifetime Access',
    };
    final call = RegExp(r"((?:'(?:[^'\\\n]|\\.)*'\s*)+)\.localized\(");
    final piece = RegExp(r"'((?:[^'\\\n]|\\.)*)'");
    final missing = <String>[];

    for (final entity in Directory('lib').listSync(recursive: true)) {
      if (entity is! File || !entity.path.endsWith('.dart')) continue;
      final path = entity.path.replaceAll(r'\', '/');
      if (path.contains('/core/localization/')) continue;
      final source = entity.readAsStringSync();
      for (final match in call.allMatches(source)) {
        final key = piece
            .allMatches(match.group(1)!)
            .map((m) => m.group(1)!)
            .join()
            .replaceAll(r"\'", "'")
            .replaceAll(r'\n', '\n')
            .replaceAll(r'\$', r'$');
        // Skip interpolated strings and fragments cut by `${...}` lexing.
        if (key.contains(r'$') ||
            key.contains('}') ||
            key.contains(']') ||
            key.contains(')')) {
          continue;
        }
        if (!RegExp('[A-Za-z]').hasMatch(key) || sameInEnglish.contains(key)) {
          continue;
        }
        if (english.text(key) == key) missing.add('$path: $key');
      }
    }

    expect(
      missing,
      isEmpty,
      reason: 'Missing English copy:\n${missing.join('\n')}',
    );
  });
}
