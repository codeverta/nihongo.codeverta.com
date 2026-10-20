import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

void main() {
  testWidgets('ruby keeps kanji aligned with the surrounding reading line', (
    tester,
  ) async {
    tester.view
      ..physicalSize = const Size(390, 300)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    const passage = 'はなさんは毎朝七時に起きます。部屋には青いカーテンがあります。今日はいい天気です。';
    const reading = 'はなさんはまいあさしちじにおきます。へやにはあおいかーてんがあります。きょうはいいてんきです。';
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: Padding(
            padding: EdgeInsets.all(20),
            child: SizedBox(
              width: 350,
              child: JapaneseRubyText(
                key: Key('ruby-passage'),
                text: passage,
                reading: reading,
                style: TextStyle(fontSize: 23, height: 1.9, letterSpacing: .4),
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final paragraph = tester.renderObject<RenderParagraph>(
      find.byKey(const Key('ruby-passage')),
    );
    final origin = paragraph.localToGlobal(Offset.zero);
    final plain = paragraph.text.toPlainText();
    final kanaOffset = plain.indexOf('はなさんは');
    final kanaRect = paragraph
        .getBoxesForSelection(
          TextSelection(baseOffset: kanaOffset, extentOffset: kanaOffset + 5),
        )
        .first
        .toRect();
    final baseParagraph = tester.renderObject<RenderParagraph>(
      find.text('毎朝七時'),
    );
    final kanjiRect = baseParagraph
        .getBoxesForSelection(
          const TextSelection(baseOffset: 0, extentOffset: 4),
        )
        .first
        .toRect()
        .shift(baseParagraph.localToGlobal(Offset.zero) - origin);

    expect(
      (kanjiRect.top - kanaRect.top).abs(),
      lessThan(4),
      reason: 'Kanji and the surrounding kana must share a text baseline.',
    );
    expect(tester.takeException(), isNull);
  });
}
