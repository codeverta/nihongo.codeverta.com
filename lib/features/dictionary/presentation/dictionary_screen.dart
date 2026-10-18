import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_controller.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_details.dart';

class DictionaryScreen extends ConsumerStatefulWidget {
  const DictionaryScreen({
    this.embedded = false,
    this.initialQuery = '',
    super.key,
  });
  final bool embedded;
  final String initialQuery;
  @override
  ConsumerState<DictionaryScreen> createState() => _DictionaryScreenState();
}

class _DictionaryScreenState extends ConsumerState<DictionaryScreen> {
  late final TextEditingController input;
  bool savedOnly = false;

  @override
  void initState() {
    super.initState();
    input = TextEditingController(text: widget.initialQuery);
    if (widget.initialQuery.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) unawaited(search(widget.initialQuery));
      });
    }
  }

  @override
  void dispose() {
    input.dispose();
    super.dispose();
  }

  Future<void> search(String query, {String? mode}) async {
    FocusScope.of(context).unfocus();
    input.text = query;
    setState(() => savedOnly = false);
    final controller = ref.read(dictionaryControllerProvider.notifier);
    final library = ref.read(dictionaryLibraryProvider.notifier);
    await controller.search(query, mode: mode);
    if (mounted && query.trim().isNotEmpty && query.runes.length <= 100) {
      await library.remember(query);
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(dictionaryControllerProvider);
    final library = ref.watch(dictionaryLibraryProvider);
    final colors = Theme.of(context).colorScheme;
    final entries = savedOnly ? library.saved : state.entries;
    final body = ListView(
      key: const PageStorageKey('dictionary-scroll'),
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
      keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: colors.primaryContainer,
                borderRadius: BorderRadius.circular(14),
              ),
              child: Text(
                '辞',
                style: TextStyle(
                  fontSize: 28,
                  color: colors.onPrimaryContainer,
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Kamus Jepang',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const Text('Cari kata. Pahami maknanya.'),
                ],
              ),
            ),
            IconButton(
              tooltip: 'Sumber dan bantuan',
              icon: const Icon(Icons.info_outline),
              onPressed: () => showDictionarySources(context),
            ),
          ],
        ),
        const SizedBox(height: 20),
        TextField(
          controller: input,
          textInputAction: TextInputAction.search,
          maxLength: 100,
          onSubmitted: (value) => unawaited(search(value)),
          decoration: InputDecoration(
            hintText: '猫, ねこ, neko, cat…',
            labelText: 'Jepang, romaji, atau Inggris',
            counterText: '',
            prefixIcon: const Icon(Icons.search),
            suffixIcon: IconButton(
              tooltip: 'Bersihkan pencarian',
              icon: const Icon(Icons.close),
              onPressed: () => unawaited(search('')),
            ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
          ),
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(
              child: FilledButton.icon(
                onPressed: () => unawaited(search(input.text)),
                icon: const Icon(Icons.search),
                label: const Text('Cari di kamus'),
              ),
            ),
            const SizedBox(width: 10),
            IconButton.filledTonal(
              tooltip: savedOnly ? 'Kembali ke hasil' : 'Kata tersimpan',
              isSelected: savedOnly,
              onPressed: () => setState(() => savedOnly = !savedOnly),
              icon: const Icon(Icons.bookmark_outline),
              selectedIcon: const Icon(Icons.bookmark),
            ),
          ],
        ),
        if (!savedOnly) ...[
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: [
              for (final mode in const {
                'auto': 'Auto',
                'japanese': 'Jepang',
                'english': 'Inggris',
              }.entries)
                ChoiceChip(
                  label: Text(mode.value),
                  selected: state.mode == mode.key,
                  onSelected: (_) =>
                      unawaited(search(input.text, mode: mode.key)),
                ),
            ],
          ),
          const SizedBox(height: 6),
          const Text(
            'Pilih Jepang untuk romaji, atau Inggris untuk mencari arti.',
          ),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                ChoiceChip(
                  label: const Text('Semua'),
                  selected: state.jlpt == null,
                  onSelected: (_) => unawaited(
                    ref
                        .read(dictionaryControllerProvider.notifier)
                        .filter(null, common: state.common),
                  ),
                ),
                for (final level in ['N5', 'N4', 'N3', 'N2', 'N1']) ...[
                  const SizedBox(width: 6),
                  ChoiceChip(
                    label: Text(level),
                    selected: state.jlpt == level,
                    onSelected: (_) => unawaited(
                      ref
                          .read(dictionaryControllerProvider.notifier)
                          .filter(level, common: state.common),
                    ),
                  ),
                ],
              ],
            ),
          ),
          Align(
            alignment: Alignment.centerLeft,
            child: FilterChip(
              label: const Text('Kata umum saja'),
              selected: state.common,
              onSelected: (value) => unawaited(
                ref
                    .read(dictionaryControllerProvider.notifier)
                    .filter(state.jlpt, common: value),
              ),
            ),
          ),
        ],
        const SizedBox(height: 12),
        Text(
          savedOnly
              ? 'Favorit tersimpan di perangkat • ${entries.length}/60'
              : 'Data Jisho.org • Arti dalam bahasa Inggris • Gratis',
          style: Theme.of(
            context,
          ).textTheme.bodySmall?.copyWith(color: colors.onSurfaceVariant),
        ),
        const SizedBox(height: 16),
        if (!savedOnly && state.query.isEmpty) ...[
          const Text('Mulai dari kata yang ingin kamu pahami.'),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: [
              for (final example in ['猫', '食べる', '学校', 'ありがとう'])
                ActionChip(
                  label: Text(example),
                  onPressed: () => unawaited(search(example)),
                ),
            ],
          ),
          if (library.recent.isNotEmpty) ...[
            const SizedBox(height: 24),
            Row(
              children: [
                const Expanded(child: Text('Pencarian terakhir')),
                TextButton(
                  onPressed: () async {
                    final ok = await ref
                        .read(dictionaryLibraryProvider.notifier)
                        .clearHistory();
                    if (!ok && context.mounted) {
                      dictionaryNotice(context, 'Riwayat belum dapat dihapus.');
                    }
                  },
                  child: const Text('Hapus'),
                ),
              ],
            ),
            Wrap(
              spacing: 8,
              runSpacing: 4,
              children: [
                for (final query in library.recent)
                  ActionChip(
                    label: Text(query),
                    onPressed: () => unawaited(search(query)),
                  ),
              ],
            ),
          ],
        ],
        if (!savedOnly && state.loading)
          const Padding(
            padding: EdgeInsets.all(32),
            child: Center(child: CircularProgressIndicator()),
          ),
        if (!savedOnly && state.error != null) ...[
          Text(state.error!, style: TextStyle(color: colors.error)),
          Wrap(
            spacing: 8,
            children: [
              OutlinedButton.icon(
                onPressed: () => unawaited(
                  state.entries.isEmpty
                      ? search(state.query)
                      : ref.read(dictionaryControllerProvider.notifier).more(),
                ),
                icon: const Icon(Icons.refresh),
                label: const Text('Coba lagi'),
              ),
              TextButton(
                onPressed: () =>
                    openDictionaryUrl(context, jishoSearchUri(state.query)),
                child: const Text('Buka Jisho'),
              ),
            ],
          ),
        ],
        if (savedOnly && entries.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 24),
            child: Text(
              'Belum ada kata tersimpan. Ketuk ikon bookmark pada hasil pencarian.',
            ),
          ),
        if (!savedOnly &&
            !state.loading &&
            state.error == null &&
            state.query.isNotEmpty) ...[
          Text(
            '${entries.length} kata dimuat untuk “${state.query}”',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 10),
          if (state.query.runes.length == 1 &&
              isDictionaryKanji(state.query.runes.first))
            Align(
              alignment: Alignment.centerLeft,
              child: ActionChip(
                label: Text('Detail kanji ${state.query}'),
                onPressed: () => showKanjiDetails(context, state.query),
              ),
            ),
          if (entries.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Text(
                'Kata belum ditemukan. Coba bentuk dasar, kana, romaji, atau kata Inggris; ubah juga filter JLPT.',
              ),
            ),
        ],
        for (final entry in entries)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: DictionaryWordCard(entry: entry),
          ),
        if (!savedOnly && state.hasMore)
          OutlinedButton.icon(
            onPressed: state.loadingMore
                ? null
                : () => unawaited(
                    ref.read(dictionaryControllerProvider.notifier).more(),
                  ),
            icon: state.loadingMore
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.expand_more),
            label: Text(state.loadingMore ? 'Memuat…' : 'Muat kata berikutnya'),
          ),
        const SizedBox(height: 20),
        TextButton(
          onPressed: () => showDictionarySources(context),
          child: const Text('Sumber data & lisensi'),
        ),
      ],
    );
    if (widget.embedded) return body;
    return Scaffold(
      appBar: AppBar(title: const Text('Kamus')),
      body: SafeArea(child: body),
    );
  }
}

class DictionaryWordCard extends StatelessWidget {
  const DictionaryWordCard({required this.entry, super.key});
  final DictionaryEntry entry;
  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.zero,
      child: InkWell(
        borderRadius: BorderRadius.circular(16),
        onTap: () => showDictionaryWord(context, entry),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          entry.reading,
                          style: Theme.of(context).textTheme.bodyLarge
                              ?.copyWith(
                                color: Theme.of(context).colorScheme.primary,
                              ),
                        ),
                        Text(
                          entry.word,
                          style: Theme.of(context).textTheme.headlineMedium
                              ?.copyWith(fontWeight: FontWeight.w800),
                        ),
                      ],
                    ),
                  ),
                  DictionaryBookmark(entry: entry),
                  const Icon(Icons.chevron_right),
                ],
              ),
              const SizedBox(height: 8),
              DictionaryTags(entry: entry),
              if (entry.senses.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  dictionaryStrings(
                    entry.senses.first['parts_of_speech'],
                  ).join(' • '),
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
              Text(entry.summary),
              if (entry.senses.length > 1)
                Text(
                  '+${entry.senses.length - 1} makna lainnya',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
            ],
          ),
        ),
      ),
    );
  }
}
