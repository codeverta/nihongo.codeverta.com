import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

void main() {
  group('alignJapaneseRuby', () {
    test('uses okurigana as an anchor', () {
      expect(alignJapaneseRuby('食べる', 'たべる').map((s) => (s.base, s.ruby)), [
        ('食', 'た'),
        ('べる', ''),
      ]);
    });

    test('aligns multiple kanji groups around written kana', () {
      expect(
        alignJapaneseRuby(
          '私は学校へ行く',
          'わたしはがっこうへいく',
        ).map((s) => (s.base, s.ruby)),
        [
          ('私', 'わたし'),
          ('は', ''),
          ('学校', 'がっこう'),
          ('へ', ''),
          ('行', 'い'),
          ('く', ''),
        ],
      );
    });

    test(
      'uses the last matching kana anchor when kanji reading contains it',
      () {
        expect(
          alignJapaneseRuby(
            '友達と学校へ行きます。',
            'ともだちとがっこうへいきます。',
          ).map((s) => (s.base, s.ruby)),
          [
            ('友達', 'ともだち'),
            ('と', ''),
            ('学校', 'がっこう'),
            ('へ', ''),
            ('行', 'い'),
            ('きます', ''),
            ('。', ''),
          ],
        );
      },
    );

    test('uses the full reading for an all-kanji word', () {
      expect(alignJapaneseRuby('学校', 'がっこう').map((s) => (s.base, s.ruby)), [
        ('学校', 'がっこう'),
      ]);
    });

    test('converts katakana anchors before matching', () {
      expect(alignJapaneseRuby('食ベる', 'たべる').map((s) => (s.base, s.ruby)), [
        ('食', 'た'),
        ('ベる', ''),
      ]);
    });

    test('keeps repeated kanji groups aligned independently', () {
      expect(
        alignJapaneseRuby('私は私です', 'わたしはわたしです').map((s) => (s.base, s.ruby)),
        [('私', 'わたし'), ('は', ''), ('私', 'わたし'), ('です', '')],
      );
    });

    test('treats the kanji iteration mark as part of a kanji word', () {
      expect(alignJapaneseRuby('日々', 'ひび').map((s) => (s.base, s.ruby)), [
        ('日々', 'ひび'),
      ]);
    });

    test(
      'ignores spaces in a sentence reading and keeps punctuation plain',
      () {
        expect(
          alignJapaneseRuby(
            '今日は学校へ行かない。',
            'きょう は がっこう へ いかない 。',
          ).map((s) => (s.base, s.ruby)),
          [
            ('今日', 'きょう'),
            ('は', ''),
            ('学校', 'がっこう'),
            ('へ', ''),
            ('行', 'い'),
            ('かない', ''),
            ('。', ''),
          ],
        );
      },
    );

    test('does not include sentence-final punctuation in ruby', () {
      expect(alignJapaneseRuby('学校。', 'がっこう。').map((s) => (s.base, s.ruby)), [
        ('学校', 'がっこう'),
        ('。', ''),
      ]);
    });
  });
}
