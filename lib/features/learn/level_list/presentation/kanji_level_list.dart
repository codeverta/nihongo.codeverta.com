import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/learn/level_list/data/level_list_repository.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/grammar_level_list.dart';
import 'package:nihongo_master/features/learn/level_list/presentation/level_pdf_button.dart';

/// "Daftar Kanji": every kanji of a level as a searchable grid, with a detail
/// sheet and a Lifetime-only PDF download.
class KanjiLevelList extends ConsumerStatefulWidget {
  const KanjiLevelList({required this.level, super.key});

  final String level;

  @override
  ConsumerState<KanjiLevelList> createState() => _KanjiLevelListState();
}

class _KanjiLevelListState extends ConsumerState<KanjiLevelList> {
  String _query = '';

  @override
  void didUpdateWidget(KanjiLevelList oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.level != widget.level) _query = '';
  }

  @override
  Widget build(BuildContext context) {
    final provider = kanjiLevelListProvider(widget.level);
    final items = ref.watch(provider);
    return items.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, _) =>
          LevelListError(onRetry: () => ref.invalidate(provider)),
      data: (all) {
        final visible = all.where((item) => item.matches(_query)).toList();
        return CustomScrollView(
          key: PageStorageKey('kanji-level-list-${widget.level}'),
          slivers: [
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              sliver: SliverToBoxAdapter(
                child: LevelListHeader(
                  count: '${all.length} kanji ${widget.level}'.localized(
                    context,
                  ),
                  searchKey: ValueKey('kanji-list-search-${widget.level}'),
                  hint: 'Cari kanji, arti, atau bacaan'.localized(context),
                  onSearch: (value) => setState(() => _query = value),
                  pdf: LevelPdfButton(
                    kind: LevelListKind.kanji,
                    level: widget.level,
                  ),
                  empty: visible.isEmpty,
                ),
              ),
            ),
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 32),
              sliver: SliverGrid.builder(
                gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                  maxCrossAxisExtent: 96,
                  mainAxisSpacing: 8,
                  crossAxisSpacing: 8,
                  childAspectRatio: 0.8,
                ),
                itemCount: visible.length,
                itemBuilder: (context, index) =>
                    _KanjiTile(item: visible[index]),
              ),
            ),
          ],
        );
      },
    );
  }
}

class _KanjiTile extends StatelessWidget {
  const _KanjiTile({required this.item});

  final KanjiListItem item;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        key: ValueKey('kanji-list-item-${item.character}'),
        onTap: () => showKanjiItemSheet(context, item),
        child: Padding(
          padding: const EdgeInsets.all(6),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                item.character,
                style: theme.textTheme.headlineSmall?.copyWith(
                  fontSize: 32,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                item.meaning,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
                style: theme.textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

Future<void> showKanjiItemSheet(BuildContext context, KanjiListItem item) {
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
            Center(
              child: SelectableText(
                item.character,
                style: theme.textTheme.displayLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            const SizedBox(height: 4),
            Text(
              item.meaning,
              textAlign: TextAlign.center,
              style: theme.textTheme.titleMedium,
            ),
            LevelListSection(title: 'Bacaan on', body: item.onReading),
            LevelListSection(title: 'Bacaan kun', body: item.kunReading),
            if (item.strokes > 0)
              LevelListSection(
                title: 'Jumlah goresan',
                body: '${item.strokes}',
              ),
            LevelListSection(title: 'Mnemonic', body: item.mnemonic),
            LevelListSection(title: 'Catatan', body: item.note),
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
                context.push(AppRoutes.kanjiLessonPath(item.lessonId));
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
