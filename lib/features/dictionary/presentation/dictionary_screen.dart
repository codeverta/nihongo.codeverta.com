import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_controller.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_details.dart';
import 'package:nihongo_master/shared/presentation/widgets/motion.dart';

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
              width: 50,
              height: 50,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: Theme.of(context).brightness == Brightness.dark
                    ? AppTheme.darkPrimary
                    : AppTheme.primaryColor,
                borderRadius: BorderRadius.circular(15),
              ),
              child: Text(
                '辞',
                style: AppTypography.japanese(
                  fontSize: 26,
                  fontWeight: FontWeight.w700,
                  color: Colors.white,
                  height: 1,
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Kamus Jepang'.localized(context),
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.5,
                    ),
                  ),
                  Text('Cari kata. Pahami maknanya.'.localized(context)),
                ],
              ),
            ),
            IconButton(
              tooltip: 'Sumber dan bantuan'.localized(context),
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
            helperText: 'Cari dengan Jepang, romaji, atau Inggris'.localized(
              context,
            ),
            filled: true,
            fillColor: Theme.of(context).brightness == Brightness.dark
                ? AppTheme.darkSurface
                : Colors.white,
            counterText: '',
            prefixIcon: const Icon(Icons.search),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(999),
              borderSide: BorderSide(color: colors.primary, width: 1.6),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(999),
              borderSide: BorderSide(color: colors.outlineVariant),
            ),
            suffixIcon: ValueListenableBuilder<TextEditingValue>(
              valueListenable: input,
              builder: (context, value, _) => AnimatedSwitcher(
                duration: AppMotion.of(context, AppMotion.normal),
                switchInCurve: AppMotion.spring,
                transitionBuilder: (child, animation) => RotationTransition(
                  turns: Tween(begin: -0.25, end: 0.0).animate(animation),
                  child: ScaleTransition(scale: animation, child: child),
                ),
                child: value.text.isEmpty
                    ? const SizedBox.shrink(key: ValueKey('empty'))
                    : IconButton(
                        key: const ValueKey('clear'),
                        tooltip: 'Bersihkan pencarian'.localized(context),
                        icon: const Icon(Icons.close),
                        onPressed: () => unawaited(search('')),
                      ),
              ),
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(999),
            ),
          ),
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(
              child: FilledButton.icon(
                onPressed: () => unawaited(search(input.text)),
                icon: const Icon(Icons.search),
                label: Text('Cari di kamus'.localized(context)),
              ),
            ),
            const SizedBox(width: 10),
            Badge(
              isLabelVisible: library.saved.isNotEmpty && !savedOnly,
              backgroundColor: colors.primary,
              textColor: colors.onPrimary,
              label: AnimatedSwitcher(
                duration: AppMotion.of(context, AppMotion.slow),
                switchInCurve: AppMotion.spring,
                transitionBuilder: (child, animation) =>
                    ScaleTransition(scale: animation, child: child),
                child: Text(
                  '${library.saved.length}',
                  key: ValueKey(library.saved.length),
                ),
              ),
              child: IconButton.filledTonal(
                tooltip: savedOnly
                    ? 'Kembali ke hasil'.localized(context)
                    : 'Kata tersimpan'.localized(context),
                isSelected: savedOnly,
                onPressed: () => setState(() => savedOnly = !savedOnly),
                icon: const Icon(Icons.bookmark_outline),
                selectedIcon: const Icon(Icons.bookmark),
              ),
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
                  'Filter${filterCount(state) == 0 ? '' : ' · ${filterCount(state)}'}'
                      .localized(context),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  [
                    switch (state.mode) {
                      'japanese' => 'Jepang / romaji'.localized(context),
                      'english' => 'Inggris'.localized(context),
                      _ => 'Bahasa otomatis'.localized(context),
                    },
                    if (state.jlpt != null) state.jlpt!,
                    if (state.common) 'Kata umum'.localized(context),
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
                    .localized(context)
              : 'Data Jisho.org • Arti dalam bahasa Inggris • Gratis'.localized(
                  context,
                ),
          style: Theme.of(
            context,
          ).textTheme.bodySmall?.copyWith(color: colors.onSurfaceVariant),
        ),
        const SizedBox(height: 16),
        if (!savedOnly && state.query.isEmpty) ...[
          const SizedBox(height: 8),
          Text(
            'Mau cari kata apa hari ini?'.localized(context),
            style: Theme.of(
              context,
            ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 4),
          Text(
            'Coba salah satu kata untuk mulai menjelajah.'.localized(context),
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
                      child: Pressable(
                        onTap: () => unawaited(search(example.$1)),
                        borderRadius: BorderRadius.circular(16),
                        child: SakuraCard(
                          radius: 16,
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                example.$1,
                                style: AppTypography.japanese(
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
                ],
              );
            },
          ),
          if (library.recent.isNotEmpty) ...[
            const SizedBox(height: 24),
            Row(
              children: [
                Expanded(child: Text('Pencarian terakhir'.localized(context))),
                TextButton(
                  onPressed: () async {
                    final ok = await ref
                        .read(dictionaryLibraryProvider.notifier)
                        .clearHistory();
                    if (!ok && context.mounted) {
                      dictionaryNotice(
                        context,
                        'Riwayat belum dapat dihapus.'.localized(context),
                      );
                    }
                  },
                  child: Text('Hapus'.localized(context)),
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
        if (!savedOnly && state.loading) const _ResultSkeleton(),
        if (!savedOnly && state.error != null) ...[
          Text(
            state.error!.localized(context),
            style: TextStyle(color: colors.error),
          ),
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
                label: Text('Coba lagi'.localized(context)),
              ),
              TextButton(
                onPressed: () =>
                    openDictionaryUrl(context, jishoSearchUri(state.query)),
                child: Text('Buka Jisho'.localized(context)),
              ),
            ],
          ),
        ],
        if (savedOnly && entries.isEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 24),
            child: Text(
              'Belum ada kata tersimpan. Ketuk ikon bookmark pada hasil pencarian.'
                  .localized(context),
            ),
          ),
        if (!savedOnly &&
            !state.loading &&
            state.error == null &&
            state.query.isNotEmpty) ...[
          Text(
            '${entries.length} kata dimuat untuk “${state.query}”'.localized(
              context,
            ),
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 10),
          if (state.query.runes.length == 1 &&
              isDictionaryKanji(state.query.runes.first))
            Align(
              alignment: Alignment.centerLeft,
              child: ActionChip(
                label: Text('Detail kanji ${state.query}'.localized(context)),
                onPressed: () => showKanjiDetails(context, state.query),
              ),
            ),
          if (entries.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16),
              child: Text(
                'Kata belum ditemukan. Coba bentuk dasar, kana, romaji, atau kata Inggris; ubah juga filter JLPT.'
                    .localized(context),
              ),
            ),
        ],
        for (final (index, entry) in entries.indexed)
          Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: StaggeredEntrance(
              key: ValueKey('${savedOnly ? 'saved' : state.query}-${entry.id}'),
              index: index,
              child: DictionaryWordCard(entry: entry),
            ),
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
            label: Text(
              (state.loadingMore ? 'Memuat…' : 'Muat kata berikutnya')
                  .localized(context),
            ),
          ),
        const SizedBox(height: 20),
        TextButton(
          onPressed: () => showDictionarySources(context),
          child: Text('Sumber data & lisensi'.localized(context)),
        ),
      ],
    );
    if (widget.embedded) return body;
    return Scaffold(
      appBar: AppBar(title: Text('Kamus'.localized(context))),
      body: SafeArea(child: body),
    );
  }
}

class DictionaryWordCard extends StatelessWidget {
  const DictionaryWordCard({required this.entry, super.key});
  final DictionaryEntry entry;
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Pressable(
      onTap: () => showDictionaryWord(context, entry),
      scale: 0.98,
      borderRadius: BorderRadius.circular(18),
      child: SakuraCard(
        padding: const EdgeInsets.fromLTRB(16, 14, 8, 14),
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
                        style: AppTypography.japanese(
                          fontSize: 13,
                          fontWeight: FontWeight.w500,
                          color: theme.colorScheme.primary,
                        ),
                      ),
                      Text(
                        entry.word,
                        style: AppTypography.japanese(
                          fontSize: 28,
                          fontWeight: FontWeight.w700,
                          color: theme.colorScheme.onSurface,
                          height: 1.2,
                        ),
                      ),
                    ],
                  ),
                ),
                DictionaryBookmark(entry: entry),
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Icon(
                    Icons.chevron_right,
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            DictionaryTags(entry: entry),
            if (entry.senses.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                dictionaryStrings(
                  entry.senses.first['parts_of_speech'],
                ).join(' • ').toUpperCase(),
                style: theme.textTheme.labelSmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                  letterSpacing: 0.6,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
            const SizedBox(height: 2),
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Text(
                entry.summary,
                style: theme.textTheme.bodyLarge?.copyWith(
                  fontWeight: FontWeight.w600,
                  fontSize: 14.5,
                ),
              ),
            ),
            if (entry.senses.length > 1)
              Text(
                '+${entry.senses.length - 1} makna lainnya'.localized(
                  context,
                ),
                style: theme.textTheme.bodySmall,
              ),
          ],
        ),
      ),
    );
  }
}

/// Placeholder berkilau saat hasil kamus sedang dimuat.
class _ResultSkeleton extends StatefulWidget {
  const _ResultSkeleton();

  @override
  State<_ResultSkeleton> createState() => _ResultSkeletonState();
}

class _ResultSkeletonState extends State<_ResultSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulse = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 900),
  );

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (AppMotion.reduced(context)) {
      _pulse.value = 0.5;
    } else if (!_pulse.isAnimating) {
      _pulse.repeat(reverse: true);
    }
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Semantics(
      label: 'Memuat…'.localized(context),
      child: AnimatedBuilder(
        animation: _pulse,
        builder: (context, _) {
          final color = Color.lerp(
            scheme.surfaceContainerHigh,
            scheme.surfaceContainerHighest,
            _pulse.value,
          )!;
          Widget bar(double width, double height) => Container(
            width: width,
            height: height,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(8),
            ),
          );
          return Column(
            children: [
              for (var i = 0; i < 3; i++)
                Padding(
                  padding: const EdgeInsets.only(bottom: 10),
                  child: SakuraCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        bar(48, 10),
                        const SizedBox(height: 8),
                        bar(90, 24),
                        const SizedBox(height: 10),
                        bar(double.infinity, 12),
                      ],
                    ),
                  ),
                ),
            ],
          );
        },
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
                    'Filter pencarian'.localized(context),
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                IconButton(
                  tooltip: 'Tutup filter'.localized(context),
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              'Bahasa pencarian'.localized(context),
              style: theme.textTheme.titleSmall,
            ),
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
                    label: Text(item.value.localized(context)),
                    selected: mode == item.key,
                    onSelected: (_) => setState(() => mode = item.key),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Pilih Jepang untuk romaji, atau Inggris untuk mencari arti.'
                  .localized(context),
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 24),
            DropdownButtonFormField<String>(
              isExpanded: true,
              initialValue: jlpt ?? 'all',
              decoration: InputDecoration(
                labelText: 'Level JLPT'.localized(context),
                border: const OutlineInputBorder(),
              ),
              items: [
                DropdownMenuItem(
                  value: 'all',
                  child: Text('Semua level'.localized(context)),
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
              title: Text('Kata umum saja'.localized(context)),
              subtitle: Text(
                'Tampilkan kata yang sering digunakan.'.localized(context),
              ),
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
                  child: Text('Reset filter'.localized(context)),
                ),
                FilledButton.icon(
                  onPressed: () => Navigator.pop(
                    context,
                    DictionaryState(mode: mode, jlpt: jlpt, common: common),
                  ),
                  icon: const Icon(Icons.check_rounded, size: 18),
                  label: Text('Terapkan filter'.localized(context)),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
