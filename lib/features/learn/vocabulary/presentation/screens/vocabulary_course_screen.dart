import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/course_offline_status.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_art.dart';
import 'package:url_launcher/url_launcher.dart';

class VocabularyCourseScreen extends ConsumerStatefulWidget {
  const VocabularyCourseScreen({this.initialLevel = 'N5', super.key});
  final String initialLevel;
  @override
  ConsumerState<VocabularyCourseScreen> createState() => _CourseState();
}

class _CourseState extends ConsumerState<VocabularyCourseScreen> {
  late String selected = widget.initialLevel;
  String query = '';
  Future<void> open(VocabularyLessonSummary lesson) async {
    await context.push(AppRoutes.vocabularyLessonPath(lesson.id));
    if (mounted) ref.invalidate(vocabularyLevelsProvider);
  }

  @override
  Widget build(BuildContext context) {
    final data = ref.watch(vocabularyLevelsProvider);
    final colors = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(title: const Text('Vocabulary JLPT')),
      body: data.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.cloud_off_rounded, size: 40),
                const SizedBox(height: 12),
                Text(
                  e.toString().localized(context),
                  textAlign: TextAlign.center,
                ),
                TextButton(
                  onPressed: () => ref.invalidate(vocabularyLevelsProvider),
                  child: Text('Muat ulang'.localized(context)),
                ),
              ],
            ),
          ),
        ),
        data: (levels) {
          if (levels.isEmpty) {
            return Center(
              child: Text('Materi belum tersedia.'.localized(context)),
            );
          }
          final level =
              levels.where((l) => l.level == selected).firstOrNull ??
              levels.first;
          final lessons = level.lessons
              .where(
                (l) => '${l.title} ${l.description}'.toLowerCase().contains(
                  query.trim().toLowerCase(),
                ),
              )
              .toList();
          final resume = level.lessons.where((l) => !l.completed).firstOrNull;
          final count = level.lessons.fold<int>(0, (n, l) => n + l.wordCount);
          return RefreshIndicator(
            onRefresh: () => ref.refresh(vocabularyLevelsProvider.future),
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(20),
              children: [
                CourseOfflineStatus(
                  type: 'vocabulary',
                  level: level.level,
                  lessonIds: level.lessons.map((l) => l.id).toList(),
                  activeLessonId: resume?.id,
                ),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      for (final item in levels)
                        Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: ChoiceChip(
                            label: Text(item.level),
                            selected: level.level == item.level,
                            showCheckmark: false,
                            selectedColor: colors.secondary,
                            labelStyle: TextStyle(
                              fontWeight: FontWeight.w800,
                              color: level.level == item.level
                                  ? colors.onSecondary
                                  : colors.onSurface,
                            ),
                            onSelected: (_) {
                              if (const [
                                    'N3',
                                    'N2',
                                    'N1',
                                  ].contains(item.level) &&
                                  !(ref
                                          .read(authControllerProvider)
                                          .user
                                          ?.isPremium ??
                                      false)) {
                                context.push('/premium');
                                return;
                              }
                              setState(() {
                                selected = item.level;
                                query = '';
                              });
                            },
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 20),
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(28),
                    gradient: LinearGradient(
                      colors: [colors.secondary, const Color(0xFF51428D)],
                    ),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'VOCABULARY / ${level.level}',
                        style: const TextStyle(
                          color: Color(0xFFFFD997),
                          fontWeight: FontWeight.w800,
                          letterSpacing: 2,
                        ),
                      ),
                      const SizedBox(height: 14),
                      Text(
                        level.level == 'N5'
                            ? 'Lihat gambarnya.\nIngat kata-katanya.'.localized(
                                context,
                              )
                            : level.title,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 28,
                          height: 1.25,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        '${level.lessons.length} lesson • $count entri kosakata'
                            .localized(context),
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Text(
                        level.description,
                        style: const TextStyle(
                          color: Colors.white,
                          height: 1.7,
                        ),
                      ),
                      if (level.sourceUrl.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        Align(
                          alignment: Alignment.centerLeft,
                          child: TextButton.icon(
                            key: Key('vocabulary-source-${level.level}'),
                            onPressed: () async {
                              final uri = Uri.tryParse(level.sourceUrl);
                              if (uri == null || !await launchUrl(uri)) {
                                if (!context.mounted) return;
                                ScaffoldMessenger.of(context).showSnackBar(
                                  SnackBar(
                                    content: Text(
                                      'Sumber tidak dapat dibuka.'.localized(
                                        context,
                                      ),
                                    ),
                                  ),
                                );
                              }
                            },
                            style: TextButton.styleFrom(
                              foregroundColor: Colors.white,
                              padding: EdgeInsets.zero,
                            ),
                            icon: const Icon(
                              Icons.open_in_new_rounded,
                              size: 16,
                            ),
                            label: Text(
                              '${'Sumber:'.localized(context)} ${level.sourceName} • ${level.sourceLicense}',
                              style: const TextStyle(fontSize: 12),
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: colors.secondary.withValues(alpha: .05),
                    borderRadius: BorderRadius.circular(22),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        '${level.lessons.where((l) => l.completed).length}/${level.lessons.length} lesson dikuasai'
                            .localized(context),
                        style: const TextStyle(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 10),
                      LinearProgressIndicator(
                        key: Key('level-progress-${level.level}'),
                        value: level.progress / 100,
                        backgroundColor: colors.secondary.withValues(alpha: .1),
                        color: colors.secondary,
                        borderRadius: BorderRadius.circular(99),
                        minHeight: 7,
                      ),
                      const SizedBox(height: 14),
                      Text(
                        'Progres latihan ${level.progress}% • lulus lesson minimal 80%'
                            .localized(context),
                        style: const TextStyle(fontSize: 12),
                      ),
                      if (resume != null) ...[
                        const SizedBox(height: 12),
                        FilledButton.icon(
                          onPressed: () => open(resume),
                          icon: const Icon(Icons.play_arrow_rounded),
                          label: Text(
                            resume.bestScore > 0
                                ? 'Lanjutkan belajar'.localized(context)
                                : 'Mulai belajar'.localized(context),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  'Jelajahi tema ${level.level}'.localized(context),
                  style: const TextStyle(
                    fontSize: 21,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 12),
                TextFormField(
                  key: ValueKey(level.level),
                  initialValue: query,
                  onChanged: (v) => setState(() => query = v),
                  decoration: InputDecoration(
                    hintText: 'Cari tema, misalnya makanan'.localized(context),
                    prefixIcon: Icon(Icons.search_rounded),
                  ),
                ),
                const SizedBox(height: 18),
                if (lessons.isEmpty)
                  Padding(
                    padding: const EdgeInsets.all(20),
                    child: Text(
                      'Tema belum ditemukan. Coba kata lain.'.localized(
                        context,
                      ),
                    ),
                  ),
                for (final lesson in lessons)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 14),
                    child: Material(
                      color: colors.surface,
                      borderRadius: BorderRadius.circular(22),
                      child: InkWell(
                        key: Key('vocabulary-lesson-${lesson.id}'),
                        onTap: () => open(lesson),
                        borderRadius: BorderRadius.circular(22),
                        child: Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            border: Border.all(
                              color: colors.outlineVariant.withValues(
                                alpha: .5,
                              ),
                            ),
                            borderRadius: BorderRadius.circular(22),
                          ),
                          child: Row(
                            children: [
                              SizedBox(
                                width: 88,
                                height: 88,
                                child: ClipRRect(
                                  borderRadius: BorderRadius.circular(16),
                                  child: lesson.imageAsset.isNotEmpty
                                      ? VocabularyArt(
                                          asset: lesson.imageAsset,
                                          url: lesson.imageUrl,
                                          description:
                                              'Ilustrasi tema ${lesson.title}'
                                                  .localized(context),
                                        )
                                      : ColoredBox(
                                          color: colors.secondary.withValues(
                                            alpha: .08,
                                          ),
                                          child: Icon(
                                            Icons.translate_rounded,
                                            color: colors.secondary,
                                            size: 32,
                                          ),
                                        ),
                                ),
                              ),
                              const SizedBox(width: 14),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '${'LESSON'.localized(context)} ${lesson.order.toString().padLeft(2, '0')}',
                                      style: TextStyle(
                                        color: colors.secondary,
                                        fontSize: 10,
                                        fontWeight: FontWeight.w800,
                                        letterSpacing: 1.3,
                                      ),
                                    ),
                                    const SizedBox(height: 5),
                                    Text(
                                      lesson.title,
                                      style: const TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                    const SizedBox(height: 6),
                                    Text(
                                      '${lesson.wordCount} kata • ${lesson.questionCount > 0 ? lesson.questionCount : lesson.wordCount} soal'
                                          .localized(context),
                                      style: const TextStyle(fontSize: 12),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      lesson.completed
                                          ? 'Dikuasai • ${lesson.bestScore}%'
                                                .localized(context)
                                          : lesson.bestScore > 0
                                          ? 'Terbaik ${lesson.bestScore}%'
                                                .localized(context)
                                          : 'Siap dipelajari'.localized(
                                              context,
                                            ),
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: lesson.completed
                                            ? const Color(0xFF167C67)
                                            : colors.onSurfaceVariant,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 6),
                              Icon(
                                lesson.completed
                                    ? Icons.check_circle_rounded
                                    : Icons.chevron_right_rounded,
                                color: lesson.completed
                                    ? const Color(0xFF167C67)
                                    : colors.secondary,
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }
}
