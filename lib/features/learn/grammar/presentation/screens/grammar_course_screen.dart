import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/course_offline_status.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/controllers/grammar_controller.dart';

class GrammarCourseScreen extends ConsumerStatefulWidget {
  const GrammarCourseScreen({this.initialLevel = 'N5', super.key});
  final String initialLevel;
  @override
  ConsumerState<GrammarCourseScreen> createState() =>
      _GrammarCourseScreenState();
}

class _GrammarCourseScreenState extends ConsumerState<GrammarCourseScreen> {
  String query = '';
  String chapter = 'Semua';
  late String level =
      ['N5', 'N4', 'N3', 'N2', 'N1'].contains(widget.initialLevel)
      ? widget.initialLevel
      : 'N5';
  void invalidateCourse() {
    if (level == 'N5') ref.invalidate(grammarCourseProvider);
    ref.invalidate(grammarCourseLevelProvider(level));
  }

  Future<void> open(GrammarLessonSummary lesson) async {
    await context.push(AppRoutes.grammarLessonPath(lesson.id));
    if (mounted) invalidateCourse();
  }

  @override
  Widget build(BuildContext context) {
    final provider = grammarCourseLevelProvider(level);
    final course = ref.watch(provider);
    return Scaffold(
      appBar: AppBar(title: Text('Grammar JLPT $level')),
      body: course.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, s) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.cloud_off_rounded, size: 48),
                const SizedBox(height: 16),
                Text(
                  e.toString().localized(context),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: invalidateCourse,
                  child: Text('Muat ulang'.localized(context)),
                ),
              ],
            ),
          ),
        ),
        data: (data) {
          final colors = Theme.of(context).colorScheme;
          final chapters = [
            'Semua',
            ...data.lessons.map((l) => l.chapter).toSet(),
          ];
          final visible = data.lessons
              .where(
                (l) =>
                    (chapter == 'Semua' || chapter == l.chapter) &&
                    '${l.title} ${l.chapter} ${l.patterns.join(' ')}'
                        .toLowerCase()
                        .contains(query.trim().toLowerCase()),
              )
              .toList();
          final resume = data.lessons.where((l) => !l.completed).firstOrNull;
          return RefreshIndicator(
            onRefresh: () async {
              invalidateCourse();
              await ref.read(provider.future);
            },
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
              children: [
                CourseOfflineStatus(
                  type: 'grammar',
                  level: level,
                  lessonIds: data.lessons.map((l) => l.id).toList(),
                  activeLessonId: resume?.id,
                ),
                Wrap(
                  runSpacing: 4,
                  children: [
                    for (final item in ['N5', 'N4', 'N3', 'N2', 'N1'])
                      Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: ChoiceChip(
                          label: Text(item),
                          selected: item == level,
                          onSelected: (_) {
                            if (const ['N3', 'N2', 'N1'].contains(item) &&
                                !(ref
                                        .read(authControllerProvider)
                                        .user
                                        ?.isPremium ??
                                    false)) {
                              context.push('/premium');
                              return;
                            }
                            setState(() {
                              level = item;
                              query = '';
                              chapter = 'Semua';
                            });
                          },
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 16),
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(24),
                    gradient: LinearGradient(
                      colors: [colors.secondary, colors.primary],
                    ),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Icon(
                        Icons.auto_stories_rounded,
                        color: Colors.white,
                        size: 36,
                      ),
                      const SizedBox(height: 14),
                      Text(
                        level == 'N1'
                            ? 'Kuasai pola tingkat lanjut untuk memahami nuansa dan maksud tersirat'
                                  .localized(context)
                            : level == 'N2'
                            ? 'Kuasai nuansa dan hubungan gagasan yang kompleks'
                                  .localized(context)
                            : level == 'N3'
                            ? 'Pahami nuansa, hubungan, dan maksud kalimat'
                                  .localized(context)
                            : level == 'N4'
                            ? 'Ungkapkan rencana, alasan, dan pendapat'
                                  .localized(context)
                            : 'Bangun kalimat Jepang dari dasar'.localized(
                                context,
                              ),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Text(
                        '${data.lessons.length} lesson • ${data.lessons.fold<int>(0, (n, l) => n + l.pointCount)} pola • ${data.lessons.fold<int>(0, (n, l) => n + l.questionCount)} soal'
                            .localized(context),
                        style: const TextStyle(color: Colors.white),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'Pahami pola, baca contoh, lalu kerjakan kuis. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.'
                            .localized(context),
                        style: const TextStyle(
                          color: Colors.white,
                          height: 1.5,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(18),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          '${data.completedLessons}/${data.lessons.length} lesson selesai • ${data.progress}%'
                              .localized(context),
                          key: const Key('grammar-course-progress'),
                          style: const TextStyle(fontWeight: FontWeight.w800),
                        ),
                        const SizedBox(height: 12),
                        LinearProgressIndicator(
                          value: data.progress / 100,
                          backgroundColor: colors.primary.withValues(
                            alpha: 0.1,
                          ),
                          minHeight: 8,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        const SizedBox(height: 16),
                        if (resume != null)
                          FilledButton.icon(
                            key: const Key('grammar-resume'),
                            onPressed: () => open(resume),
                            icon: const Icon(Icons.play_arrow_rounded),
                            label: Text(
                              data.completedLessons == 0
                                  ? 'Mulai belajar'.localized(context)
                                  : 'Lanjutkan belajar'.localized(context),
                            ),
                          )
                        else
                          Text(
                            'Semua lesson selesai! Anda tetap dapat mengulang materi dan kuis.'
                                .localized(context),
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),
                TextField(
                  key: ValueKey('grammar-search-$level'),
                  onChanged: (v) => setState(() => query = v),
                  decoration: InputDecoration(
                    hintText: 'Cari tema atau pola, misalnya そう'.localized(
                      context,
                    ),
                    prefixIcon: Icon(Icons.search_rounded),
                  ),
                ),
                const SizedBox(height: 14),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      for (final c in chapters)
                        Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: ChoiceChip(
                            label: Text(
                              c == 'Semua' ? c.localized(context) : c,
                            ),
                            selected: chapter == c,
                            onSelected: (_) => setState(() => chapter = c),
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                if (visible.isEmpty)
                  Padding(
                    padding: EdgeInsets.all(24),
                    child: Text(
                      'Tidak ada lesson yang cocok. Coba judul atau bab lain.'
                          .localized(context),
                      textAlign: TextAlign.center,
                    ),
                  ),
                for (final lesson in visible) ...[
                  Card(
                    child: ListTile(
                      key: Key('grammar-lesson-${lesson.id}'),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 10,
                      ),
                      onTap: () => open(lesson),
                      leading: CircleAvatar(
                        backgroundColor: colors.primary.withValues(alpha: 0.1),
                        child: lesson.completed
                            ? Icon(Icons.check_rounded, color: colors.primary)
                            : Text(
                                '${lesson.order}',
                                style: TextStyle(
                                  color: colors.primary,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                      ),
                      title: Text(
                        lesson.title,
                        style: const TextStyle(fontWeight: FontWeight.w700),
                      ),
                      subtitle: Text(
                        '${lesson.chapter}\n${lesson.pointCount} pola • ${lesson.questionCount} soal • Terbaik ${lesson.bestScore}%'
                            .localized(context),
                      ),
                      isThreeLine: true,
                      trailing: const Icon(Icons.chevron_right_rounded),
                    ),
                  ),
                  const SizedBox(height: 10),
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}
