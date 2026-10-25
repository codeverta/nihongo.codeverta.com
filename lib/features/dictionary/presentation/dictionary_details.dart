import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/dictionary/data/dictionary_repository.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_controller.dart';
import 'package:url_launcher/url_launcher.dart';

Uri jishoSearchUri(String query) => Uri.https('jisho.org', '/search/$query');

void dictionaryNotice(BuildContext context, String message) {
  ScaffoldMessenger.of(
    context,
  ).showSnackBar(SnackBar(content: Text(message.localized(context))));
}

Future<void> openDictionaryUrl(BuildContext context, Uri uri) async {
  try {
    if (!await launchUrl(uri, mode: LaunchMode.externalApplication) &&
        context.mounted) {
      dictionaryNotice(context, 'Tautan belum dapat dibuka.');
    }
  } on Object {
    if (context.mounted) {
      dictionaryNotice(context, 'Tautan belum dapat dibuka.');
    }
  }
}

class DictionaryBookmark extends ConsumerWidget {
  const DictionaryBookmark({required this.entry, super.key});
  final DictionaryEntry entry;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final saved = ref
        .watch(dictionaryLibraryProvider)
        .saved
        .any((e) => e.id == entry.id);
    return IconButton(
      tooltip:
          (saved ? 'Hapus ${entry.word} dari favorit' : 'Simpan ${entry.word}')
              .localized(context),
      icon: Icon(saved ? Icons.bookmark : Icons.bookmark_border),
      onPressed: () async {
        final ok = await ref
            .read(dictionaryLibraryProvider.notifier)
            .toggle(entry);
        if (!ok && context.mounted) {
          dictionaryNotice(
            context,
            'Gagal menyimpan atau favorit sudah mencapai batas 60 kata.',
          );
        }
      },
    );
  }
}

class DictionaryTags extends StatelessWidget {
  const DictionaryTags({required this.entry, super.key});
  final DictionaryEntry entry;
  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 6,
    runSpacing: 4,
    children: [
      if (entry.common)
        Chip(
          label: Text('Kata umum'.localized(context)),
          visualDensity: VisualDensity.compact,
        ),
      for (final level in entry.jlpt)
        Chip(label: Text('JLPT $level'), visualDensity: VisualDensity.compact),
    ],
  );
}

void _sheet(BuildContext context, Widget child) {
  showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    useSafeArea: true,
    builder: (context) => FractionallySizedBox(heightFactor: .9, child: child),
  );
}

void showDictionaryWord(BuildContext context, DictionaryEntry entry) {
  _sheet(context, _WordDetails(entry: entry));
}

class _WordDetails extends StatelessWidget {
  const _WordDetails({required this.entry});
  final DictionaryEntry entry;
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.fromLTRB(24, 0, 24, 32),
    children: [
      Row(
        children: [
          Expanded(
            child: Text(
              'Detail kata'.localized(context),
              style: Theme.of(context).textTheme.titleLarge,
            ),
          ),
          DictionaryBookmark(entry: entry),
          IconButton(
            tooltip: 'Tutup detail'.localized(context),
            onPressed: () => Navigator.pop(context),
            icon: const Icon(Icons.close),
          ),
        ],
      ),
      SelectableText(
        entry.reading,
        style: Theme.of(context).textTheme.titleLarge,
      ),
      SelectableText(
        entry.word,
        style: Theme.of(context).textTheme.displaySmall,
      ),
      DictionaryTags(entry: entry),
      const SizedBox(height: 20),
      for (var i = 0; i < entry.senses.length; i++)
        _Sense(sense: entry.senses[i], number: i + 1),
      if (entry.forms.length > 1) ...[
        const Divider(),
        Text(
          'Bentuk dan bacaan lain'.localized(context),
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        for (final form in entry.forms.skip(1))
          SelectableText(
            '${form['word'] ?? form['reading'] ?? ''} ${form['word'] != null ? '【${form['reading'] ?? ''}】' : ''}',
          ),
      ],
      if (entry.kanji.isNotEmpty) ...[
        const SizedBox(height: 16),
        Text('Kanji dalam kata ini'.localized(context)),
        Wrap(
          spacing: 8,
          children: [
            for (final character in entry.kanji)
              ActionChip(
                label: Text(character, style: const TextStyle(fontSize: 24)),
                onPressed: () => showKanjiDetails(context, character),
              ),
          ],
        ),
      ],
      const Divider(height: 32),
      Wrap(
        spacing: 8,
        runSpacing: 4,
        children: [
          OutlinedButton.icon(
            onPressed: () async {
              await Clipboard.setData(
                ClipboardData(
                  text:
                      '${entry.word}【${entry.reading}】\n${entry.senses.map((s) => dictionaryStrings(s['english_definitions']).join('; ')).join('\n')}',
                ),
              );
              if (context.mounted) {
                dictionaryNotice(context, 'Kata dan arti disalin.');
              }
            },
            icon: const Icon(Icons.copy),
            label: Text('Salin'.localized(context)),
          ),
          OutlinedButton(
            onPressed: () => openDictionaryUrl(
              context,
              Uri.https('jisho.org', '/word/${entry.id}'),
            ),
            child: Text('Lihat di Jisho'.localized(context)),
          ),
          OutlinedButton(
            onPressed: () => openDictionaryUrl(
              context,
              jishoSearchUri('${entry.word} #sentences'),
            ),
            child: Text('Contoh kalimat di Jisho'.localized(context)),
          ),
        ],
      ),
      const SizedBox(height: 16),
      Text(
        'Sumber: Jisho.org. Makna dan urutan mengikuti sumber; bukan terjemahan otomatis.'
            .localized(context),
      ),
      TextButton(
        onPressed: () => showDictionarySources(context),
        child: Text('Atribusi & lisensi data'.localized(context)),
      ),
    ],
  );
}

class _Sense extends StatelessWidget {
  const _Sense({required this.sense, required this.number});
  final Map<String, dynamic> sense;
  final int number;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 18),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          dictionaryStrings(sense['parts_of_speech']).join(' • '),
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
            color: Theme.of(context).colorScheme.primary,
          ),
        ),
        SelectableText(
          '$number. ${dictionaryStrings(sense['english_definitions']).join('; ')}',
          style: Theme.of(context).textTheme.bodyLarge,
        ),
        for (final key in ['tags', 'info', 'restrictions'])
          if (dictionaryStrings(sense[key]).isNotEmpty)
            Text(
              '${key == 'restrictions' ? 'Berlaku untuk: ' : ''}${dictionaryStrings(sense[key]).join(' • ')}'
                  .localized(context),
              style: Theme.of(context).textTheme.bodySmall,
            ),
        for (final key in ['see_also', 'antonyms'])
          if (dictionaryStrings(sense[key]).isNotEmpty) ...[
            Text(
              (key == 'antonyms' ? 'Lawan kata' : 'Lihat juga').localized(
                context,
              ),
            ),
            Wrap(
              spacing: 6,
              children: [
                for (final related in dictionaryStrings(sense[key]))
                  ActionChip(
                    label: Text(related),
                    onPressed: () =>
                        openDictionaryUrl(context, jishoSearchUri(related)),
                  ),
              ],
            ),
          ],
      ],
    ),
  );
}

void showKanjiDetails(BuildContext context, String character) {
  _sheet(context, _KanjiDetails(character: character));
}

class _KanjiDetails extends ConsumerWidget {
  const _KanjiDetails({required this.character});
  final String character;
  @override
  Widget build(BuildContext context, WidgetRef ref) => ListView(
    padding: const EdgeInsets.fromLTRB(24, 0, 24, 32),
    children: [
      Row(
        children: [
          Expanded(
            child: Text(
              'Detail kanji'.localized(context),
              style: Theme.of(context).textTheme.titleLarge,
            ),
          ),
          IconButton(
            tooltip: 'Tutup kanji'.localized(context),
            onPressed: () => Navigator.pop(context),
            icon: const Icon(Icons.close),
          ),
        ],
      ),
      Text(character, style: Theme.of(context).textTheme.displayLarge),
      ref
          .watch(dictionaryKanjiProvider(character))
          .when(
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ),
            ),
            error: (error, stack) => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Detail kanji belum dapat dimuat dari sumber kamus.'
                      .localized(context),
                ),
                OutlinedButton(
                  onPressed: () =>
                      ref.invalidate(dictionaryKanjiProvider(character)),
                  child: Text('Coba lagi'.localized(context)),
                ),
              ],
            ),
            data: (data) => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  dictionaryStrings(data['meanings']).join('; '),
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 16),
                Wrap(
                  spacing: 8,
                  children: [
                    Chip(
                      label: Text(
                        '${data['stroke_count']} goresan'.localized(context),
                      ),
                    ),
                    if (data['grade'] != null)
                      Chip(label: Text('Grade ${data['grade']}')),
                    if (data['jlpt'] != null)
                      Chip(
                        label: Text(
                          'JLPT lama ${data['jlpt']}'.localized(context),
                        ),
                      ),
                  ],
                ),
                for (final pair in [
                  ('Onyomi', 'on_readings'),
                  ('Kunyomi', 'kun_readings'),
                  ('Bacaan nama', 'name_readings'),
                ]) ...[
                  const SizedBox(height: 16),
                  Text(
                    pair.$1.localized(context),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SelectableText(
                    dictionaryStrings(data[pair.$2]).isEmpty
                        ? 'Tidak tersedia'.localized(context)
                        : dictionaryStrings(data[pair.$2]).join('、'),
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ],
                if (data['freq_mainichi_shinbun'] != null)
                  Text(
                    'Peringkat frekuensi koran: ${data['freq_mainichi_shinbun']}'
                        .localized(context),
                  ),
                const SizedBox(height: 16),
                Text(
                  'Label JLPT kanji berasal dari klasifikasi lama KANJIDIC dan berbeda dengan label N1–N5 pada kosakata.'
                      .localized(context),
                ),
              ],
            ),
          ),
      const SizedBox(height: 24),
      OutlinedButton(
        onPressed: () =>
            openDictionaryUrl(context, jishoSearchUri('$character #kanji')),
        child: Text('Radikal & urutan goresan di Jisho'.localized(context)),
      ),
      const SizedBox(height: 12),
      Text('Sumber: kanjiapi.dev / KANJIDIC2 (EDRDG).'.localized(context)),
      TextButton(
        onPressed: () => showDictionarySources(context),
        child: Text('Atribusi & lisensi data'.localized(context)),
      ),
    ],
  );
}

void showDictionarySources(BuildContext context) {
  _sheet(
    context,
    ListView(
      padding: const EdgeInsets.fromLTRB(24, 0, 24, 32),
      children: [
        Row(
          children: [
            Expanded(
              child: Text(
                'Tentang kamus'.localized(context),
                style: const TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            IconButton(
              tooltip: 'Tutup bantuan'.localized(context),
              onPressed: () => Navigator.pop(context),
              icon: const Icon(Icons.close),
            ),
          ],
        ),
        Text(
          'Cari dengan kanji, hiragana, katakana, romaji, atau kata Inggris. Gunakan filter JLPT dan kata umum. Arti disajikan dalam bahasa Inggris sesuai sumber.'
              .localized(context),
        ),
        const SizedBox(height: 16),
        Text(
          'Kata: Jisho.org, menggunakan JMdict/JMnedict dari Electronic Dictionary Research and Development Group (EDRDG), serta DBpedia untuk makna ensiklopedia. Detail kanji: kanjiapi.dev, menggunakan KANJIDIC2. Data EDRDG digunakan sesuai lisensinya (CC BY-SA). DBpedia memiliki ketentuan atribusi dan berbagi serupa.'
              .localized(context),
        ),
        const SizedBox(height: 16),
        Text(
          'Favorit menyimpan salinan entri di perangkat, terpisah untuk setiap akun. Favorit dapat dibaca tanpa internet; pencarian dan detail kanji memerlukan koneksi. Riwayat bisa dihapus. Maksimal 60 favorit dan 20 pencarian terakhir.'
              .localized(context),
        ),
        const SizedBox(height: 16),
        Text(
          'Kamus ini memakai sumber gratis tanpa API key. Ketersediaannya mengikuti layanan sumber. Contoh kalimat lengkap, pencarian radikal, dan diagram goresan dibuka di Jisho. Nihongo Master tidak berafiliasi dengan Jisho.org.'
              .localized(context),
        ),
        for (final link in [
          ('Jisho.org & sumber data', 'https://jisho.org/about'),
          ('EDRDG & lisensi kamus', 'https://www.edrdg.org/edrdg/licence.html'),
          ('kanjiapi.dev', 'https://kanjiapi.dev/'),
          ('DBpedia & lisensi', 'https://www.dbpedia.org/resources/'),
        ])
          TextButton(
            onPressed: () => openDictionaryUrl(context, Uri.parse(link.$2)),
            child: Text(link.$1.localized(context)),
          ),
      ],
    ),
  );
}
