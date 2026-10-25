import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/activity/learning_history.dart';

class ContinueLearningSection extends ConsumerWidget {
  const ContinueLearningSection({super.key});

  Future<void> _open(
    BuildContext context,
    WidgetRef ref,
    LearningActivity item,
  ) async {
    final account = ref.read(authControllerProvider).user?.id;
    await ref.read(learningHistoryProvider.notifier).refresh();
    if (!context.mounted ||
        ref.read(authControllerProvider).user?.id != account) {
      return;
    }
    final selected =
        ref
            .read(learningHistoryProvider)
            .where((entry) => entry.key == item.key)
            .firstOrNull ??
        item;
    final premium = ref.read(authControllerProvider).user?.isPremium ?? false;
    final path = selected.available
        ? '${selected.lessonPath}?level=${selected.level}'
        : selected.coursePath;
    if (!selected.available) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Materi ini sudah tidak tersedia. Pilih materi lain di kursusnya.'
                .localized(context),
          ),
        ),
      );
    }
    await context.push<void>(
      selected.isLocked(premium: premium)
          ? '/premium?from=${Uri.encodeQueryComponent(path)}'
          : path,
    );
    if (context.mounted) {
      await ref.read(learningHistoryProvider.notifier).refresh();
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(authControllerProvider).user;
    if (user == null) return const SizedBox.shrink();
    final items = ref.watch(learningHistoryProvider).take(4).toList();
    if (items.isEmpty) return const SizedBox.shrink();
    final latest = items.first;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final locked = latest.isLocked(premium: user.isPremium);
    return Padding(
      padding: const EdgeInsets.only(bottom: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Lanjutkan belajar'.localized(context),
            style: theme.textTheme.titleLarge?.copyWith(
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'Kembali ke materi yang terakhir kamu buka.'.localized(context),
            style: theme.textTheme.bodySmall,
          ),
          const SizedBox(height: 12),
          DecoratedBox(
            decoration: BoxDecoration(
              color: colors.secondary.withValues(alpha: .05),
              border: Border.all(
                color: colors.secondary.withValues(alpha: .12),
              ),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        _icon(latest.kind),
                        size: 20,
                        color: colors.secondary,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          '${latest.category.localized(context)} · JLPT ${latest.level}',
                          style: theme.textTheme.labelLarge?.copyWith(
                            color: colors.secondary,
                          ),
                        ),
                      ),
                      if (locked) const Icon(Icons.lock_outline, size: 18),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Text(
                    latest.title.localized(context),
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    _detail(context, latest, locked),
                    style: theme.textTheme.bodySmall,
                  ),
                  const SizedBox(height: 12),
                  if (latest.available) ...[
                    ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        value: latest.completed
                            ? 1
                            : (latest.position + 1) / latest.total,
                        minHeight: 4,
                        color: colors.secondary,
                        backgroundColor: colors.secondary.withValues(alpha: .1),
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],
                  FilledButton.icon(
                    onPressed: () => _open(context, ref, latest),
                    style: FilledButton.styleFrom(
                      backgroundColor: colors.secondary,
                    ),
                    icon: Icon(
                      locked ? Icons.lock_outline : Icons.play_arrow_rounded,
                      size: 20,
                    ),
                    label: Text(
                      !latest.available
                          ? 'Buka kursus'.localized(context)
                          : locked
                          ? 'Buka akses belajar'.localized(context)
                          : latest.completed
                          ? 'Pelajari lagi'.localized(context)
                          : 'Lanjutkan belajar'.localized(context),
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (items.length > 1) ...[
            const SizedBox(height: 16),
            Text(
              'Sebelumnya kamu belajar'.localized(context),
              style: theme.textTheme.labelLarge,
            ),
            const SizedBox(height: 4),
            for (final item in items.skip(1))
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Icon(_icon(item.kind), color: colors.secondary),
                title: Text(
                  item.title.localized(context),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                  ),
                ),
                subtitle: Text(
                  '${item.category.localized(context)} · ${item.level}\n${_detail(context, item, item.isLocked(premium: user.isPremium))}',
                  style: theme.textTheme.bodySmall,
                ),
                trailing: Icon(
                  item.isLocked(premium: user.isPremium)
                      ? Icons.lock_outline
                      : Icons.chevron_right_rounded,
                  size: 20,
                ),
                onTap: () => _open(context, ref, item),
              ),
          ],
        ],
      ),
    );
  }

  String _detail(BuildContext context, LearningActivity item, bool locked) {
    final status = !item.available
        ? 'Materi tidak tersedia'.localized(context)
        : locked
        ? 'Akses Lifetime diperlukan'.localized(context)
        : item.completed
        ? 'Selesai'.localized(context)
        : '${item.unit.localized(context)} ${item.position + 1} dari ${item.total}'
              .localized(context);
    final at = item.accessedAt.toLocal();
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final day = DateTime(at.year, at.month, at.day);
    final difference = today.difference(day).inDays;
    final date = difference == 0
        ? 'Hari ini'.localized(context)
        : difference == 1
        ? 'Kemarin'.localized(context)
        : '${at.day}/${at.month}/${at.year}';
    return '$status · $date';
  }
}

IconData _icon(String kind) => switch (kind) {
  'vocabulary' => Icons.translate_rounded,
  'grammar' => Icons.menu_book_outlined,
  'kanji' => Icons.brush_outlined,
  _ => Icons.auto_stories_outlined,
};
