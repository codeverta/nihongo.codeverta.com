import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';

class VocabularyCourseScreen extends ConsumerWidget {
  const VocabularyCourseScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final levels = ref.watch(vocabularyLevelsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Vocabulary JLPT')),
      body: levels.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) => _ErrorState(
          message: error.toString(),
          retry: () => ref.invalidate(vocabularyLevelsProvider),
        ),
        data: (items) => RefreshIndicator(
          onRefresh: () => ref.refresh(vocabularyLevelsProvider.future),
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
            children: [
              const _CourseHeader(),
              const SizedBox(height: 24),
              for (final level in items) ...[
                _LevelSection(
                  level: level,
                  openLesson: (lesson) async {
                    await context.push(
                      AppRoutes.vocabularyLessonPath(lesson.id),
                    );
                    ref.invalidate(vocabularyLevelsProvider);
                  },
                ),
                const SizedBox(height: 18),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _CourseHeader extends StatelessWidget {
  const _CourseHeader();

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(22),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(24),
        gradient: LinearGradient(
          colors: [colors.primary, colors.secondary],
        ),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(Icons.translate_rounded, color: Colors.white, size: 36),
          SizedBox(height: 14),
          Text(
            'Kuasai kosakata dari N5 hingga N1',
            style: TextStyle(
              color: Colors.white,
              fontSize: 22,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
          SizedBox(height: 8),
          Text(
            'Pelajari arti, cara baca, penggunaan, lalu buktikan pemahaman melalui kuis. Nilai terbaik tersimpan otomatis.',
            style: TextStyle(color: Colors.white, height: 1.5),
          ),
        ],
      ),
    );
  }
}

class _LevelSection extends StatelessWidget {
  const _LevelSection({required this.level, required this.openLesson});

  final VocabularyLevel level;
  final ValueChanged<VocabularyLessonSummary> openLesson;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 52,
                  height: 52,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: colors.primary.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text(
                    level.level,
                    style: TextStyle(
                      color: colors.primary,
                      fontSize: 20,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 13),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        level.title,
                        style:
                            Theme.of(context).textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                      ),
                      const SizedBox(height: 3),
                      Text(level.description),
                    ],
                  ),
                ),
                Text(
                  '${level.progress}%',
                  key: Key('level-progress-${level.level}'),
                  style: TextStyle(
                    color: colors.primary,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 15),
            LinearProgressIndicator(
              value: level.progress / 100,
              minHeight: 8,
              borderRadius: BorderRadius.circular(99),
            ),
            const SizedBox(height: 14),
            for (final lesson in level.lessons)
              _LessonTile(lesson: lesson, onTap: () => openLesson(lesson)),
          ],
        ),
      ),
    );
  }
}

class _LessonTile extends StatelessWidget {
  const _LessonTile({required this.lesson, required this.onTap});

  final VocabularyLessonSummary lesson;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return ListTile(
      key: Key('vocabulary-lesson-${lesson.id}'),
      onTap: onTap,
      contentPadding: EdgeInsets.zero,
      leading: CircleAvatar(
        backgroundColor:
            lesson.completed ? colors.primary : colors.surfaceContainerHighest,
        foregroundColor:
            lesson.completed ? colors.onPrimary : colors.onSurfaceVariant,
        child: Icon(
          lesson.completed ? Icons.check_rounded : Icons.play_arrow_rounded,
        ),
      ),
      title: Text(
        '${lesson.order}. ${lesson.title}',
        style: const TextStyle(fontWeight: FontWeight.w700),
      ),
      subtitle: Text(
        '${lesson.wordCount} kata • Nilai terbaik ${lesson.bestScore}%',
      ),
      trailing: const Icon(Icons.chevron_right_rounded),
    );
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({required this.message, required this.retry});

  final String message;
  final VoidCallback retry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off_rounded, size: 52),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: retry,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Coba lagi'),
            ),
          ],
        ),
      ),
    );
  }
}
