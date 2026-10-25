import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

void main() {
  test('English copy includes dictionary, Anki, and learning labels', () {
    const english = AppLocalizations(Locale('en'));

    expect(english.text('Materi'), 'Materials');
    expect(english.text('Tes'), 'Tests');
    expect(
      english.text('Lifetime Access diperlukan'),
      'Lifetime Access is required',
    );
    expect(english.text('Periksa kartu berikutnya'), 'Check the next card');
    expect(english.text('Kembali ke detail deck'), 'Back to deck details');
    expect(english.text('Detail kata'), 'Word details');
    expect(english.text('Kanji dalam kata ini'), 'Kanji in this word');
  });

  test('English account, attribution, and home copy is translated', () {
    const english = AppLocalizations(Locale('en'));

    expect(english.text('Lanjutkan dengan Google'), 'Continue with Google');
    expect(english.text('Lanjut sebagai tamu'), 'Continue as guest');
    expect(
      english.text('Hapus akun permanen?'),
      'Permanently delete your account?',
    );
    expect(
      english.text('Berita & informasi Jepang'),
      'Japanese news & information',
    );
    expect(
      english.text('Belajar Jepang, setiap hari'),
      'Learn Japanese every day',
    );
    expect(
      english.text(
        'Kenali bidang SSW, pelajari materi, dan coba simulasi ujiannya.',
      ),
      'Explore SSW fields, study the material, and try an exam simulation.',
    );
    expect(
      english.text('Masuk untuk melihat profil.'),
      'Sign in to view your profile.',
    );
  });

  test('English interpolation handles achievement and banner templates', () {
    const english = AppLocalizations(Locale('en'));

    expect(
      english.text('Lihat semua achievement (4/12)'),
      'View all achievements (4/12)',
    );
    expect(
      english.text('4/12 achievement terbuka'),
      '4/12 achievements unlocked',
    );
    expect(
      english.text('Materi N3–N1 yang tersedia • Rp99.000'),
      'Available N3–N1 materials · IDR 99,000',
    );
    expect(english.text('更新されたコンテンツ • Rp120.000'), '更新されたコンテンツ · IDR 120,000');
    expect(
      const AppLocalizations(
        Locale('id'),
      ).text('Lihat semua achievement (4/12)'),
      'Lihat semua achievement (4/12)',
    );
  });

  test('English learn and exam copy covers fixed and runtime text', () {
    const english = AppLocalizations(Locale('en'));

    expect(english.text('Arti cerita'), 'Story meaning');
    expect(english.text('Peta soal'), 'Question map');
    expect(english.text('Pilih satu jawaban'), 'Choose one answer');
    expect(english.text('Simulasi N2'), 'N2 simulation');
    expect(
      english.text('1. 文字・語彙 — 21 soal / 20 menit'),
      '1. 文字・語彙 — 21 questions / 20 min',
    );
    expect(english.text('Bagian 2 · 会話と表現'), 'Part 2 · 会話と表現');
    expect(
      english.text('4/12 dijawab · 2 ditandai'),
      '4/12 answered · 2 marked',
    );
    expect(
      english.text(
        'Bagian sebelumnya dikunci. 文法 dimulai; timer total terus berjalan.',
      ),
      'The previous section is locked. 文法 starts; the total timer keeps running.',
    );
    expect(
      english.text('昔の話\n20 menit • 4 adegan • Terbaik 80%'),
      '昔の話\n20 min · 4 scenes · Best 80%',
    );
    expect(
      english.text(
        'JFT Exam 3 · Kosakata, Percakapan & Reading\n30/36 benar · 83%',
      ),
      'JFT Exam 3 · Vocabulary, Conversation & Reading\n30/36 correct · 83%',
    );
    expect(
      english.text(
        'Semua jawaban benar. Kamu bisa membuka seluruh pembahasan.',
      ),
      'All answers are correct. You can view all explanations.',
    );
    expect(english.text('Kurikulum Kaigo'), 'The Kaigo curriculum');
    expect(
      english.text(
        'Exam 2 · 36 soal · 45 menit\nLatihan tertulis orisinal JFT-Basic. Listening belum termasuk.',
      ),
      'JFT-Basic Exam 2 · 36 questions · 45 minutes\nOriginal written practice for JFT-Basic. Listening is not included.',
    );
    expect(
      english.text(
        'Paket 01 · 38 soal · 40 menit\nKosakata, kanji, dan grammar N5. Dokkai dan listening dilewati.',
      ),
      'Set 01 · 38 questions · 40 minutes\nVocabulary, kanji, and grammar for N5. Dokkai and listening are omitted.',
    );
    expect(
      english.text('2 dijawab\n1 soal kosong\n3 menit 5 detik'),
      '2 answered\n1 unanswered\n3 min 5 sec',
    );
    expect(
      english.text('Masih ada 1 soal kosong; soal kosong dihitung salah.\n\n'),
      'There are 1 unanswered questions; unanswered questions count as incorrect.\n\n',
    );
  });

  test('English SSW assessment metadata is translated', () {
    const english = AppLocalizations(Locale('en'));

    expect(english.text('Coba kembali'), 'Try again');
    expect(english.text('Pertanian · Peternakan'), 'Agriculture · Livestock');
    expect(
      english.text('36 soal · 4 modul · operasi dan keselamatan kereta'),
      '36 questions · 4 modules · railway operations and safety',
    );
    expect(
      english.text('45 soal · 60 menit · 5 ranah keterampilan'),
      '45 questions · 60 min · 5 skill areas',
    );
    expect(english.text('Budi Daya'), 'Aquaculture');
  });

  test(
    'English Anki catalog metadata is translated or intentionally preserved',
    () async {
      const english = AppLocalizations(Locale('en'));
      final catalog =
          jsonDecode(
                await File(
                  'backend/internal/anki/resources/catalog.json',
                ).readAsString(),
              )
              as Map<String, dynamic>;
      final preserved = {
        'Kana & Kanji',
        'JLPT',
        'Hentaigana',
        'Hiragana',
        'Yojijukugo',
        'Katakana',
        'Kodansha Kanji Learner’s Course',
        'Kokuji',
      };
      final gaps = <String>{};

      for (final deck in catalog['decks'] as List<dynamic>) {
        for (final field in [
          'title',
          'category',
          'description',
          'instructions',
          'meaningLanguage',
          'unavailableReason',
        ]) {
          final source = (deck as Map<String, dynamic>)[field] as String? ?? '';
          if (source.isEmpty || preserved.contains(source)) continue;
          if (english.text(source) == source) {
            gaps.add('$field: $source');
          }
        }
      }

      expect(gaps, isEmpty, reason: gaps.join('\n'));
    },
  );

  testWidgets('English home copy is translated while Japanese stays intact', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('en'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: Builder(
          builder: (context) => Scaffold(
            body: Column(
              children: [
                Text('Selamat datang!'.localized(context)),
                Text('おはよう, Aiko!'.localized(context)),
                Text('3 pelajaran'.localized(context)),
                Text('ありがとう'),
              ],
            ),
          ),
        ),
      ),
    );

    expect(find.text('Welcome!'), findsOneWidget);
    expect(find.text('Good morning, Aiko!'), findsOneWidget);
    expect(find.text('3 lessons'), findsOneWidget);
    expect(find.text('ありがとう'), findsOneWidget);
  });
}
