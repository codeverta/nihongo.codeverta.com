import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/controllers/kanji_controller.dart';

class KanjiCourseScreen extends ConsumerStatefulWidget {
  const KanjiCourseScreen({this.level = 'N5', super.key});
  final String level;
  @override
  ConsumerState<KanjiCourseScreen> createState() => _KanjiCourseScreenState();
}

class _KanjiCourseScreenState extends ConsumerState<KanjiCourseScreen> {
  String query = '';
  String chapter = 'Semua';
  Future<void> open(KanjiLessonSummary lesson) async {
    await context.push(AppRoutes.kanjiLessonPath(lesson.id));
    if (mounted) ref.invalidate(kanjiCourseProvider(widget.level));
  }

  @override
  Widget build(BuildContext context) {
    final provider = kanjiCourseProvider(widget.level);
    final course = ref.watch(provider);
    return Scaffold(
      appBar: AppBar(title: Text('Kanji JLPT ${widget.level}')),
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
                    '${l.title} ${l.chapter}'
                        .toLowerCase()
                        .contains(query.toLowerCase()),
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
                        Icons.translate_rounded,
                        color: Colors.white,
                        size: 36,
                      ),
                      const SizedBox(height: 14),
                      Text(
                        widget.level == 'N3'
                            ? 'Kenali kanji N3 dalam kalimat sehari-hari'
                            : widget.level == 'N4'
                                ? 'Perluas kemampuan kanji N4'
                                : 'Kenali dan baca kanji dasar',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Text(
                        '${data.lessons.length} lesson • ${data.lessons.fold<int>(0, (n, l) => n + l.characterCount)} kanji • ${data.lessons.fold<int>(0, (n, l) => n + l.questionCount)} soal',
                        style: const TextStyle(color: Colors.white),
                      ),
                      const SizedBox(height: 12),
                      const Text(
                        'Pahami kanji, baca contoh, lalu kerjakan kuis. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.',
                        style: TextStyle(
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
                          '${data.completedLessons}/${data.lessons.length} lesson selesai • ${data.progress}%',
                          key: const Key(
                            'kanji-course-progress',
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
                            key: const Key('kanji-resume'),
                            onPressed: () => open(resume),
                            icon: const Icon(
                              Icons.play_arrow_rounded,
                            ),
                            label: Text(
                              data.completedLessons == 0
                                  ? 'Mulai belajar'
                                  : 'Lanjutkan belajar',
                            ),
                          )
                        else
                          const Text(
                            'Semua lesson selesai! Anda tetap dapat mengulang materi dan kuis.',
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),
                TextField(
                  key: const Key('kanji-search'),
                  onChanged: (v) => setState(() => query = v),
                  decoration: const InputDecoration(
                    hintText: 'Cari judul lesson',
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
                      'Tidak ada lesson yang cocok. Coba judul atau bab lain.',
                      textAlign: TextAlign.center,
                    ),
                  ),
                for (final lesson in visible) ...[
                  Card(
                    child: ListTile(
                      key: Key('kanji-lesson-${lesson.id}'),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 10,
                      ),
                      onTap: () => open(lesson),
                      leading: CircleAvatar(
                        backgroundColor: colors.primary.withValues(alpha: 0.1),
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
                        '${lesson.chapter}\n${lesson.characterCount} kanji • ${lesson.questionCount} soal • Terbaik ${lesson.bestScore}%',
                      ),
                      isThreeLine: true,
                      trailing: const Icon(Icons.chevron_right_rounded),
                    ),
                  ),
                  const SizedBox(height: 10),
                ],
                if (widget.level == 'N4' || widget.level == 'N3')
                  Card(
                    child: ExpansionTile(
                      key: Key('kanji-${widget.level.toLowerCase()}-sources'),
                      title: const Text('Tentang materi dan sumber'),
                      childrenPadding: const EdgeInsets.all(16),
                      children: [
                        SelectableText(
                          '${widget.level == 'N3' ? '367 kanji dari daftar rujukan N3, dengan contoh kosakata dan kalimat.' : '166 kanji dari daftar rujukan N4, termasuk 21 kanji pengulangan N5.'} '
                          'Bacaan pada kartu merupakan pilihan bacaan kamus. '
                          'Petunjuk mengingat dibuat sebagai bantuan belajar.\n\n'
                          'Jumlah goresan dan data bacaan menggunakan KANJIDIC oleh James William Breen dan EDRDG, '
                          'melalui kanji-data oleh David Gouveia. '
                          'Materi data ${widget.level} diadaptasi dengan arti, contoh, dan kuis Indonesia; CC BY-SA 4.0.\n'
                          'Sumber: https://github.com/davidluzgouveia/kanji-data\n'
                          'Lisensi kamus: https://www.edrdg.org/edrdg/licence.html\n'
                          'CC BY-SA 4.0: https://creativecommons.org/licenses/by-sa/4.0/\n\n'
                          'JLPT tidak menerbitkan daftar kanji resmi yang tertutup: '
                          'https://www.jlpt.jp/e/faq/',
                        ),
                      ],
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
