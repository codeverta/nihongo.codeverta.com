import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/learn/level_list/data/level_list_repository.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/level_pdf_button.dart';

/// "Daftar Pola": every grammar point of a level, searchable, with a detail
/// sheet and a Lifetime-only PDF download.
class GrammarLevelList extends ConsumerStatefulWidget {
  const GrammarLevelList({required this.level, this.onLevelChanged, super.key});

  final String level;

  /// When set, N5–N1 chips let the user switch level without leaving the list.
  final ValueChanged<String>? onLevelChanged;

  @override
  ConsumerState<GrammarLevelList> createState() => _GrammarLevelListState();
}

class _GrammarLevelListState extends ConsumerState<GrammarLevelList> {
  String _query = '';

  @override
  void didUpdateWidget(GrammarLevelList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.level != widget.level) _query = '';
  }

  @override
  Widget build(BuildContext context) {
    final provider = grammarLevelListProvider(widget.level);
    final items = ref.watch(provider);
    return items.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, _) =>
          LevelListError(onRetry: () => ref.invalidate(provider)),
      data: (all) {
        final visible = all.where((item) => item.matches(_query)).toList();
        return ListView.builder(
          key: PageStorageKey('grammar-level-list-${widget.level}'),
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
          itemCount: visible.length + 1,
          itemBuilder: (context, index) {
            if (index == 0) {
              return LevelListHeader(
                count: '${all.length} pola grammar ${widget.level}'.localized(
                  context,
                ),
                searchKey: ValueKey('grammar-list-search-${widget.level}'),
                hint: 'Cari pola atau arti'.localized(context),
                onSearch: (value) => setState(() => _query = value),
                pdf: LevelPdfButton(
                  kind: LevelListKind.grammar,
                  level: widget.level,
                ),
                level: widget.level,
                onLevelChanged: widget.onLevelChanged,
                empty: visible.isEmpty,
              );
            }
            final item = visible[index - 1];
            return _GrammarTile(item: item);
          },
        );
      },
    );
  }
}

class _GrammarTile extends StatelessWidget {
  const _GrammarTile({required this.item});

  final GrammarListItem item;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        key: ValueKey('grammar-list-item-${item.lessonId}-${item.pattern}'),
        title: Text(
          item.pattern,
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        subtitle: Text(
          '${item.meaning}\n${'Lesson'.localized(context)} ${item.lessonNumber} · ${item.lessonTitle}',
        ),
        isThreeLine: true,
        trailing: const Icon(Icons.chevron_right_rounded),
        onTap: () => showGrammarItemSheet(context, item),
      ),
    );
  }
}

Future<void> showGrammarItemSheet(BuildContext context, GrammarListItem item) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    showDragHandle: true,
    builder: (context) {
      final theme = Theme.of(context);
      return DraggableScrollableSheet(
        expand: false,
        initialChildSize: 0.7,
        maxChildSize: 0.95,
        builder: (context, scroll) => ListView(
          controller: scroll,
          padding: const EdgeInsets.fromLTRB(24, 0, 24, 24),
          children: [
            SelectableText(
              item.pattern,
              style: theme.textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 4),
            Text(item.meaning, style: theme.textTheme.titleMedium),
            LevelListSection(title: 'Pembentukan', body: item.formation),
            LevelListSection(title: 'Penjelasan', body: item.explanation),
            LevelListSection(title: 'Kesalahan umum', body: item.commonMistake),
            if (item.examples.isNotEmpty) ...[
              const LevelListSection(title: 'Contoh', body: ''),
              for (final example in item.examples)
                LevelListExampleTile(example),
            ],
            const SizedBox(height: 20),
            FilledButton.icon(
              key: const ValueKey('level-list-open-lesson'),
              onPressed: () {
                Navigator.pop(context);
                context.push(AppRoutes.grammarLessonPath(item.lessonId));
              },
              icon: const Icon(Icons.menu_book_rounded),
              label: Text(
                '${'Buka lesson'.localized(context)} ${item.lessonNumber}',
              ),
            ),
          ],
        ),
      );
    },
  );
}

/// Count, search field and PDF button shared by the grammar and kanji lists.
class LevelListHeader extends StatelessWidget {
  const LevelListHeader({
    required this.count,
    required this.searchKey,
    required this.hint,
    required this.onSearch,
    required this.pdf,
    required this.empty,
    this.level,
    this.onLevelChanged,
    super.key,
  });

  final String? level;
  final ValueChanged<String>? onLevelChanged;

  final String count;
  final Key searchKey;
  final String hint;
  final ValueChanged<String> onSearch;
  final Widget pdf;
  final bool empty;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (level != null && onLevelChanged != null) ...[
            Wrap(
              spacing: 8,
              children: [
                for (final item in const ['N5', 'N4', 'N3', 'N2', 'N1'])
                  ChoiceChip(
                    label: Text(item),
                    selected: item == level,
                    onSelected: (_) => onLevelChanged!(item),
                  ),
              ],
            ),
            const SizedBox(height: 12),
          ],
          Row(
            children: [
              Expanded(
                child: Text(
                  count,
                  style: Theme.of(context).textTheme.titleSmall,
                ),
              ),
              pdf,
            ],
          ),
          const SizedBox(height: 12),
          TextField(
            key: searchKey,
            onChanged: onSearch,
            decoration: InputDecoration(
              prefixIcon: const Icon(Icons.search_rounded),
              hintText: hint,
              border: const OutlineInputBorder(
                borderRadius: BorderRadius.all(Radius.circular(14)),
              ),
            ),
          ),
          if (empty)
            Padding(
              padding: const EdgeInsets.only(top: 24),
              child: Text(
                'Tidak ada hasil.'.localized(context),
                textAlign: TextAlign.center,
              ),
            ),
        ],
      ),
    );
  }
}

class LevelListSection extends StatelessWidget {
  const LevelListSection({required this.title, required this.body, super.key});

  final String title;
  final String body;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    if (body.trim().isEmpty && title != 'Contoh') {
      return const SizedBox.shrink();
    }
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title.localized(context),
            style: theme.textTheme.labelLarge?.copyWith(
              color: theme.colorScheme.primary,
            ),
          ),
          if (body.trim().isNotEmpty) ...[
            const SizedBox(height: 4),
            SelectableText(body, style: theme.textTheme.bodyMedium),
          ],
        ],
      ),
    );
  }
}

class LevelListExampleTile extends StatelessWidget {
  const LevelListExampleTile(this.example, {super.key});

  final LevelListExample example;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(top: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SelectableText(example.japanese, style: theme.textTheme.titleMedium),
          if (example.reading.isNotEmpty && example.reading != example.japanese)
            Text(example.reading, style: theme.textTheme.bodySmall),
          if (example.meaning.isNotEmpty)
            Text(example.meaning, style: theme.textTheme.bodyMedium),
        ],
      ),
    );
  }
}

class LevelListError extends StatelessWidget {
  const LevelListError({required this.onRetry, super.key});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off_rounded, size: 48),
            const SizedBox(height: 12),
            Text(
              'Daftar belum dapat dimuat. Periksa koneksi lalu coba lagi.'
                  .localized(context),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: onRetry,
              child: Text('Muat ulang'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}
