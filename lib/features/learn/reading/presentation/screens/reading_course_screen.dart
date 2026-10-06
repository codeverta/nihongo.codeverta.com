import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';
import 'package:nihongo_master/features/learn/reading/presentation/controllers/reading_controller.dart';
import 'package:nihongo_master/features/learn/reading/presentation/widgets/reading_illustration.dart';

class ReadingCourseScreen extends ConsumerStatefulWidget {
  const ReadingCourseScreen({super.key, this.level = 'N5'});
  final String level;
  @override
  ConsumerState<ReadingCourseScreen> createState() =>
      _ReadingCourseScreenState();
}

class _ReadingCourseScreenState extends ConsumerState<ReadingCourseScreen> {
  String query = '';
  String chapter = 'Semua';
  Future<void> open(ReadingLessonSummary lesson) async {
    await context.push(AppRoutes.readingLessonPath(lesson.id));
    if (mounted) ref.invalidate(readingCourseProvider(widget.level));
  }

  @override
  Widget build(BuildContext context) {
    final provider = readingCourseProvider(widget.level);
    final course = ref.watch(provider);
    return Scaffold(
      appBar: AppBar(title: Text('Reading JLPT ${widget.level}')),
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
                Text(e.toString(), textAlign: TextAlign.center),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: () => ref.invalidate(provider),
                  child: const Text('Muat ulang'),
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
                    '${l.title} ${l.japaneseTitle} ${l.chapter} ${l.description}'
                        .toLowerCase()
                        .contains(query.trim().toLowerCase()),
              )
              .toList();
          final resume = data.lessons.where((l) => !l.completed).firstOrNull;
          return RefreshIndicator(
            onRefresh: () => ref.refresh(provider.future),
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
              children: [
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
                      const Text(
                        'Cerita kecil, dunia baru',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Text(
                        '${data.lessons.length} lesson • ${data.lessons.fold<int>(0, (n, l) => n + l.pageCount)} adegan bergambar • ${data.lessons.fold<int>(0, (n, l) => n + l.questionCount)} soal',
                        style: const TextStyle(color: Colors.white),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        widget.level == 'N2'
                            ? 'Latih inferensi, maksud penulis, perbandingan pendapat, dan pencarian informasi. Gunakan bantuan kana dan terjemahan saat diperlukan. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.'
                            : widget.level == 'N3'
                                ? 'Pahami alasan, maksud tokoh, dan kesimpulan dalam cerita Hana. Buka bantuan kana dan terjemahan saat diperlukan. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.'
                                : 'Ikuti Hana dalam cerita bergambar. Buka bantuan kana dan terjemahan saat diperlukan. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.',
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
                          '${data.completedLessons}/${data.lessons.length} cerita selesai • ${data.progress}%',
                          key: const Key(
                            'reading-course-progress',
                          ),
                          style: const TextStyle(
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                        const SizedBox(height: 12),
                        LinearProgressIndicator(
                          value: data.progress / 100,
                          backgroundColor:
                              colors.primary.withValues(alpha: 0.1),
                          minHeight: 8,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        const SizedBox(height: 16),
                        if (resume != null)
                          FilledButton.icon(
                            key: const Key('reading-resume'),
                            onPressed: () => open(resume),
                            icon: const Icon(
                              Icons.play_arrow_rounded,
                            ),
                            label: Text(
                              data.completedLessons == 0
                                  ? 'Buka cerita pertama'
                                  : 'Lanjutkan membaca',
                            ),
                          )
                        else
                          const Text(
                            'Semua cerita selesai! Anda tetap dapat mengulang materi dan kuis.',
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),
                TextField(
                  key: const Key('reading-search'),
                  onChanged: (v) => setState(() => query = v),
                  decoration: const InputDecoration(
                    hintText: 'Cari cerita atau judul Jepang',
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
                            label: Text(c),
                            selected: chapter == c,
                            onSelected: (_) => setState(() => chapter = c),
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                if (visible.isEmpty)
                  const Padding(
                    padding: EdgeInsets.all(24),
                    child: Text(
                      'Cerita tidak ditemukan. Coba kata atau tema lain.',
                      textAlign: TextAlign.center,
                    ),
                  ),
                for (final lesson in visible) ...[
                  Card(
                    clipBehavior: Clip.antiAlias,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        InkWell(
                          onTap: () => open(lesson),
                          child: ReadingIllustration(
                            asset: lesson.imageAsset,
                            url: lesson.imageUrl,
                            panel: 0,
                            description: 'Ilustrasi ${lesson.title}',
                            aspectRatio: 16 / 9,
                          ),
                        ),
                        ListTile(
                          key: Key('reading-lesson-${lesson.id}'),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 10,
                          ),
                          onTap: () => open(lesson),
                          leading: CircleAvatar(
                            backgroundColor:
                                colors.primary.withValues(alpha: 0.1),
                            child: lesson.completed
                                ? Icon(
                                    Icons.check_rounded,
                                    color: colors.primary,
                                  )
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
                            style: const TextStyle(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          subtitle: Text(
                            '${lesson.japaneseTitle}\n${lesson.minutes} menit • ${lesson.pageCount} adegan • Terbaik ${lesson.bestScore}%',
                          ),
                          isThreeLine: true,
                          trailing: const Icon(Icons.chevron_right_rounded),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}
