import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_ruby_text.dart';

void main() {
  const words = [
    VocabularyWord(
      id: 'friend',
      term: '友達',
      reading: 'ともだち',
      meaning: 'teman',
      example: '友達と学校へ行きます。',
      exampleMeaning: 'Saya pergi ke sekolah bersama teman.',
      note: '',
      examples: [
        VocabularyExample(
          japanese: '友達と学校へ行きます。',
          reading: 'ともだちとがっこうへいきます。',
          meaning: 'Saya pergi ke sekolah bersama teman.',
        ),
      ],
    ),
  ];

  testWidgets('adds furigana to Japanese quiz prompts and choices', (
    tester,
  ) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              VocabularyRubyText(
                text: 'Apa arti 友達 (ともだち)?',
                words: words,
                enabled: true,
              ),
              VocabularyRubyText(
                text: '友達と学校へ行きます。',
                words: words,
                enabled: true,
              ),
            ],
          ),
        ),
      ),
    );

    expect(find.text('ともだち'), findsWidgets);
    final renderedText = tester
        .widgetList<Text>(find.byType(Text))
        .map((widget) => widget.data ?? widget.textSpan?.toPlainText())
        .whereType<String>()
        .toList();
    expect(renderedText, contains('がっこう'));
    expect(renderedText, contains('Apa arti \uFFFC?'));
  });

  testWidgets('omits furigana when disabled', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: VocabularyRubyText(
            text: 'Apa arti 友達?',
            words: words,
            enabled: false,
          ),
        ),
      ),
    );

    expect(find.text('ともだち'), findsNothing);
    expect(find.text('Apa arti 友達?'), findsOneWidget);
  });
}
