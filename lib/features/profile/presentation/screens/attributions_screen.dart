import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';

class AttributionsLink extends StatelessWidget {
  const AttributionsLink({super.key});
  @override
  Widget build(BuildContext context) => ListTile(
    key: const Key('open-attributions'),
    leading: const Icon(Icons.source_outlined, size: 21),
    title: const Text('Sumber data & atribusi', style: TextStyle(fontSize: 13)),
    trailing: const Icon(Icons.chevron_right_rounded, size: 20),
    onTap: () => context.push('/attributions'),
  );
}

class AttributionsScreen extends StatelessWidget {
  const AttributionsScreen({super.key});
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Sumber data & atribusi')),
    body: ListView(
      padding: const EdgeInsets.all(20),
      children: [
        Text(
          'Dibangun bersama banyak sumber',
          style: Theme.of(
            context,
          ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 10),
        const Text(
          'Terima kasih kepada penerbit, pembuat dataset, dan komunitas yang menjadi sumber konten Nihongo Master. Hak cipta tetap milik masing-masing pemilik.',
        ),
        const SizedBox(height: 24),
        const _SourceGroup(
          title: 'Manga Jepang',
          sources: [
            (
              'Rawkuma',
              'Sumber katalog, metadata, dan panel manga berbahasa Jepang. Karya tetap milik penulis dan penerbit masing-masing.',
              'https://rawkuma.net/',
            ),
          ],
        ),
        const _SourceGroup(
          title: 'Berita & informasi Jepang',
          sources: [
            (
              'Global Voices 日本語',
              'Berita Jepang · CC BY 3.0. Furigana ditambahkan oleh Nihongo Master. Kredit penulis, penerjemah, dan gambar tersedia pada setiap artikel.',
              'https://jp.globalvoices.org/',
            ),
            (
              '大阪市 · Pemerintah Kota Osaka',
              'Informasi publik Jepang. Ketentuan penggunaan Osaka City kompatibel dengan CC BY 4.0; furigana ditambahkan oleh Nihongo Master.',
              'https://www.city.osaka.lg.jp/main/site_policy/0000000124.html',
            ),
          ],
        ),
        const _SourceGroup(
          title: 'Kamus Jepang',
          sources: [
            (
              'Jisho.org',
              'Hasil pencarian kata dan bacaan Jepang melalui API Jisho. Atribusi dataset yang mendasari entri mengikuti informasi dari Jisho.',
              'https://jisho.org/about',
            ),
            (
              'Kanji API',
              'Informasi kanji dan bacaannya melalui kanjiapi.dev.',
              'https://kanjiapi.dev/',
            ),
          ],
        ),
        const _SourceGroup(
          title: 'Kosakata JLPT',
          sources: [
            (
              'OpenJLPT',
              'Dataset kosakata · CC BY-SA 4.0. Data diadaptasi menjadi lesson, dikelompokkan menurut tema, dan diterjemahkan ke bahasa Indonesia. Tingkat JLPT merupakan perkiraan komunitas.',
              'https://github.com/evanclan/OpenJLPT',
            ),
          ],
        ),
        const _SourceGroup(
          title: 'Anki & flashcards',
          sources: [
            (
              'Kotoba / mistval',
              'Resource quiz_data untuk deck flashcard. Lisensi repository dan catatan sumber individual dapat dilihat di halaman Anki → Sumber data & lisensi.',
              'https://github.com/mistval/kotoba/tree/master/resources/quiz_data',
            ),
          ],
        ),
        Card(
          child: ListTile(
            leading: const Icon(Icons.work_outline_rounded),
            title: const Text('Sumber lowongan kerja'),
            subtitle: const Text(
              'Daftar portal dan penyedia lowongan. Sumber asli juga dicantumkan pada setiap lowongan.',
            ),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () => context.push('/jobs/sources'),
          ),
        ),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          onPressed: () => showLicensePage(
            context: context,
            applicationName: 'Nihongo Master',
          ),
          icon: const Icon(Icons.code_rounded),
          label: const Text('Lisensi perangkat lunak open source'),
        ),
      ],
    ),
  );
}

class _SourceGroup extends StatelessWidget {
  const _SourceGroup({required this.title, required this.sources});
  final String title;
  final List<(String, String, String)> sources;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 22),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(
            context,
          ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 8),
        for (final (name, description, url) in sources)
          Card(
            child: ListTile(
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 10,
              ),
              title: Text(name),
              subtitle: Padding(
                padding: const EdgeInsets.only(top: 6),
                child: Text(description),
              ),
              trailing: const Icon(Icons.open_in_new_rounded, size: 18),
              onTap: () async {
                var opened = false;
                try {
                  opened = await launchUrl(
                    Uri.parse(url),
                    mode: LaunchMode.externalApplication,
                  );
                } on Object {
                  /* Show the same retry message for platform errors. */
                }
                if (!opened && context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text(
                        'Tautan sumber belum dapat dibuka. Coba lagi.',
                      ),
                    ),
                  );
                }
              },
            ),
          ),
      ],
    ),
  );
}
