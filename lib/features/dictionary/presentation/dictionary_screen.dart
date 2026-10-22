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

  int filterCount(DictionaryState state) =>
      (state.mode == 'auto' ? 0 : 1) +
      (state.jlpt == null ? 0 : 1) +
      (state.common ? 1 : 0);

  Future<void> showFilters(DictionaryState state) async {
    FocusScope.of(context).unfocus();
    final result = await showModalBottomSheet<DictionaryState>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      showDragHandle: true,
      builder: (_) => _DictionaryFilters(initial: state),
    );
    if (result == null || !mounted) return;
    final query = input.text;
    final library = ref.read(dictionaryLibraryProvider.notifier);
    await ref
        .read(dictionaryControllerProvider.notifier)
        .filter(
          result.jlpt,
          common: result.common,
          mode: result.mode,
          text: query,
        );
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
            helperText: 'Cari dengan Jepang, romaji, atau Inggris',
            filled: true,
            fillColor: colors.surface,
            counterText: '',
            prefixIcon: const Icon(Icons.search),
            suffixIcon: ValueListenableBuilder<TextEditingValue>(
              valueListenable: input,
              builder: (context, value, _) => value.text.isEmpty
                  ? const SizedBox.shrink()
                  : IconButton(
                      tooltip: 'Bersihkan pencarian',
                      icon: const Icon(Icons.close),
                      onPressed: () => unawaited(search('')),
                    ),
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
          const SizedBox(height: 12),
          Row(
            children: [
              OutlinedButton.icon(
                style: OutlinedButton.styleFrom(
                  minimumSize: const Size(0, 44),
                  padding: const EdgeInsets.symmetric(horizontal: 14),
                ),
                onPressed: () => unawaited(showFilters(state)),
                icon: const Icon(Icons.tune_rounded, size: 18),
                label: Text(
                  'Filter${filterCount(state) == 0 ? '' : ' · ${filterCount(state)}'}',
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  [
                    switch (state.mode) {
                      'japanese' => 'Jepang / romaji',
                      'english' => 'Inggris',
                      _ => 'Bahasa otomatis',
                    },
                    if (state.jlpt != null) state.jlpt!,
                    if (state.common) 'Kata umum',
                  ].join(' · '),
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: colors.onSurfaceVariant,
                  ),
                ),
              ),
            ],
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
          const SizedBox(height: 8),
          Text(
            'Mau cari kata apa hari ini?',
            style: Theme.of(
              context,
            ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 4),
          Text(
            'Coba salah satu kata untuk mulai menjelajah.',
            style: TextStyle(color: colors.onSurfaceVariant, fontSize: 13),
          ),
          const SizedBox(height: 14),
          LayoutBuilder(
            builder: (context, constraints) {
              final width = constraints.maxWidth < 260
                  ? constraints.maxWidth
                  : (constraints.maxWidth - 12) / 2;
              return Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  for (final example in const [
                    ('猫', 'ねこ · cat'),
                    ('食べる', 'たべる · to eat'),
                    ('学校', 'がっこう · school'),
                    ('ありがとう', 'thank you'),
                  ])
                    SizedBox(
                      width: width,
                      child: Material(
                        color: colors.surface,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                          side: BorderSide(color: colors.outlineVariant),
                        ),
                        clipBehavior: Clip.antiAlias,
                        child: InkWell(
                          onTap: () => unawaited(search(example.$1)),
                          child: Padding(
                            padding: const EdgeInsets.all(14),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  example.$1,
                                  style: TextStyle(
                                    fontSize: 20,
                                    fontWeight: FontWeight.w700,
                                    color: colors.secondary,
                                  ),
                                ),
                                const SizedBox(height: 6),
                                Text(
                                  example.$2,
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: colors.onSurfaceVariant,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                ],
              );
            },
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

class _DictionaryFilters extends StatefulWidget {
  const _DictionaryFilters({required this.initial});
  final DictionaryState initial;

  @override
  State<_DictionaryFilters> createState() => _DictionaryFiltersState();
}

class _DictionaryFiltersState extends State<_DictionaryFilters> {
  late String mode = widget.initial.mode;
  late String? jlpt = widget.initial.jlpt;
  late bool common = widget.initial.common;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      top: false,
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(24, 0, 24, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    'Filter pencarian',
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                IconButton(
                  tooltip: 'Tutup filter',
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text('Bahasa pencarian', style: theme.textTheme.titleSmall),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final item in const {
                  'auto': 'Auto',
                  'japanese': 'Jepang',
                  'english': 'Inggris',
                }.entries)
                  ChoiceChip(
                    label: Text(item.value),
                    selected: mode == item.key,
                    onSelected: (_) => setState(() => mode = item.key),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Pilih Jepang untuk romaji, atau Inggris untuk mencari arti.',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 24),
            DropdownButtonFormField<String>(
              isExpanded: true,
              initialValue: jlpt ?? 'all',
              decoration: const InputDecoration(
                labelText: 'Level JLPT',
                border: OutlineInputBorder(),
              ),
              items: [
                const DropdownMenuItem(
                  value: 'all',
                  child: Text('Semua level'),
                ),
                for (final level in ['N5', 'N4', 'N3', 'N2', 'N1'])
                  DropdownMenuItem(value: level, child: Text(level)),
              ],
              onChanged: (value) =>
                  setState(() => jlpt = value == 'all' ? null : value),
              // Recreate the field when Reset changes its initial selection.
              key: ValueKey(jlpt),
            ),
            const SizedBox(height: 16),
            SwitchListTile.adaptive(
              contentPadding: EdgeInsets.zero,
              title: const Text('Kata umum saja'),
              subtitle: const Text('Tampilkan kata yang sering digunakan.'),
              value: common,
              onChanged: (value) => setState(() => common = value),
            ),
            const SizedBox(height: 20),
            Wrap(
              alignment: WrapAlignment.spaceBetween,
              spacing: 12,
              runSpacing: 12,
              children: [
                TextButton(
                  onPressed: () => setState(() {
                    mode = 'auto';
                    jlpt = null;
                    common = false;
                  }),
                  child: const Text('Reset filter'),
                ),
                FilledButton.icon(
                  onPressed: () => Navigator.pop(
                    context,
                    DictionaryState(mode: mode, jlpt: jlpt, common: common),
                  ),
                  icon: const Icon(Icons.check_rounded, size: 18),
                  label: const Text('Terapkan filter'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
