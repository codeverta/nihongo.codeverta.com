import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:url_launcher/url_launcher.dart';

String kaigoLessonPath(String id) =>
    '/ssw/nursing-care/lessons/${Uri.encodeComponent(id)}';

class KaigoCurriculumPanel extends ConsumerWidget {
  const KaigoCurriculumPanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => ref
      .watch(kaigoCurriculumProvider)
      .when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) =>
            _LoadError(onRetry: () => ref.invalidate(kaigoCurriculumProvider)),
        data: (course) {
          final completed = ref.watch(kaigoProgressProvider);
          final count =
              course.lessons.where((l) => completed.contains(l.id)).length;
          final next = course.lessons.firstWhere(
            (lesson) => !completed.contains(lesson.id),
            orElse: () => course.lessons.first,
          );
          final signedIn =
              ref.watch(authControllerProvider.select((s) => s.user != null));
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Kurikulum Kaigo',
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 8),
              Text(
                '${course.modules.length} modul · ${course.lessons.length} lesson · ±${course.minutes ~/ 60} jam belajar',
              ),
              const SizedBox(height: 12),
              const Text(
                'Pelajari berurutan dari dasar perawatan hingga bahasa Jepang di tempat kerja. Tiap lesson berisi tujuan, penjelasan, langkah kerja, kasus, kosakata, dan cek pemahaman.',
                style: TextStyle(height: 1.6),
              ),
              const SizedBox(height: 12),
              const Text(
                'Ritme belajar: 1–2 lesson per hari. Baca materi, ucapkan contoh Jepang, jelaskan kembali kasus dengan kata sendiri, lalu kerjakan cek pemahaman. Ulangi bagian yang belum jelas sebelum lanjut.',
                style: TextStyle(height: 1.6),
              ),
              const SizedBox(height: 16),
              Card(
                margin: EdgeInsets.zero,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '$count/${course.lessons.length} lesson selesai',
                        key: const Key('kaigo-progress'),
                        style: const TextStyle(fontWeight: FontWeight.w700),
                      ),
                      const SizedBox(height: 10),
                      LinearProgressIndicator(
                        value: count / course.lessons.length,
                        borderRadius: BorderRadius.circular(8),
                        minHeight: 7,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        signedIn
                            ? 'Progress tersimpan untuk akunmu di perangkat ini.'
                            : 'Masuk untuk menyimpan progress setelah aplikasi ditutup.',
                      ),
                      const SizedBox(height: 12),
                      FilledButton.icon(
                        key: const Key('kaigo-continue'),
                        onPressed: () => context.push(kaigoLessonPath(next.id)),
                        icon: const Icon(Icons.menu_book_rounded),
                        label: Text(
                          count == 0
                              ? 'Mulai belajar'
                              : count == course.lessons.length
                                  ? 'Pelajari kembali'
                                  : 'Lanjutkan belajar',
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              const _StudyNote(),
              const SizedBox(height: 20),
              for (var index = 0; index < course.modules.length; index++) ...[
                Card(
                  margin: const EdgeInsets.only(bottom: 10),
                  clipBehavior: Clip.antiAlias,
                  child: ExpansionTile(
                    key: PageStorageKey(
                      'kaigo-module-${course.modules[index].id}',
                    ),
                    initiallyExpanded: index == 0,
                    title: Text(
                      '${index + 1}. ${course.modules[index].title}',
                      style: const TextStyle(fontWeight: FontWeight.w700),
                    ),
                    subtitle: Text(
                      '${course.modules[index].lessons.where((l) => completed.contains(l.id)).length}/${course.modules[index].lessons.length} selesai',
                    ),
                    childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                    children: [
                      Text(
                        course.modules[index].description,
                        style: const TextStyle(height: 1.5),
                      ),
                      const SizedBox(height: 8),
                      for (final lesson in course.modules[index].lessons)
                        ListTile(
                          key: Key('kaigo-open-${lesson.id}'),
                          contentPadding: EdgeInsets.zero,
                          leading: Icon(
                            completed.contains(lesson.id)
                                ? Icons.check_circle_rounded
                                : Icons.article_outlined,
                            color: Theme.of(context).colorScheme.primary,
                          ),
                          title: Text(lesson.title),
                          subtitle: Text(
                            '±${lesson.minutes} menit · ${lesson.checks.length} cek pemahaman',
                          ),
                          trailing: const Icon(Icons.chevron_right_rounded),
                          onTap: () => context.push(kaigoLessonPath(lesson.id)),
                        ),
                    ],
                  ),
                ),
              ],
              _Sources(course: course),
            ],
          );
        },
      );
}

class KaigoLessonScreen extends ConsumerWidget {
  const KaigoLessonScreen({required this.lessonId, super.key});
  final String lessonId;

  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
        appBar: AppBar(title: const Text('Lesson Kaigo')),
        body: ref.watch(kaigoCurriculumProvider).when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => _LoadError(
                onRetry: () => ref.invalidate(kaigoCurriculumProvider),
              ),
              data: (course) {
                final lesson = course.lessonById(lessonId);
                if (lesson == null) {
                  return const Center(child: Text('Lesson tidak ditemukan.'));
                }
                return _LessonReader(
                  key: ValueKey(lesson.id),
                  course: course,
                  lesson: lesson,
                );
              },
            ),
      );
}

class _LessonReader extends ConsumerStatefulWidget {
  const _LessonReader({required this.course, required this.lesson, super.key});
  final KaigoCurriculum course;
  final KaigoLesson lesson;
  @override
  ConsumerState<_LessonReader> createState() => _LessonReaderState();
}

class _LessonReaderState extends ConsumerState<_LessonReader> {
  final selections = <int, int>{};
  bool submitted = false;
  bool saving = false;

  Future<void> complete() async {
    setState(() => saving = true);
    final account = ref.read(authControllerProvider).user?.id;
    try {
      await ref.read(kaigoProgressProvider.notifier).complete(widget.lesson.id);
      if (!mounted) return;
      if (ref.read(authControllerProvider).user?.id != account) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content:
                Text('Akun berubah. Buka kembali materi untuk melanjutkan.'),
          ),
        );
        return;
      }
      final lessons = widget.course.lessons;
      final index = lessons.indexWhere((l) => l.id == widget.lesson.id);
      if (index + 1 < lessons.length) {
        context.pushReplacement(kaigoLessonPath(lessons[index + 1].id));
      } else {
        context.go('/ssw/nursing-care');
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Progress belum tersimpan. Silakan coba lagi.'),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final lesson = widget.lesson;
    final index = widget.course.lessons.indexWhere((l) => l.id == lesson.id);
    final allAnswered = selections.length == lesson.checks.length;
    final correct = submitted &&
        List.generate(
          lesson.checks.length,
          (i) => selections[i] == lesson.checks[i].answer,
        ).every((v) => v);
    final completed = ref.watch(kaigoProgressProvider).contains(lesson.id);
    final colors = Theme.of(context).colorScheme;
    return ListView(
      key: PageStorageKey('kaigo-reader-${lesson.id}'),
      padding: const EdgeInsets.all(20),
      children: [
        Text(
          'LESSON ${index + 1} / ${widget.course.lessons.length} · ±${lesson.minutes} MENIT',
          style: TextStyle(color: colors.primary, fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 12),
        Text(
          lesson.title,
          key: const Key('kaigo-lesson-title'),
          style: Theme.of(context)
              .textTheme
              .headlineSmall
              ?.copyWith(fontWeight: FontWeight.w800),
        ),
        if (completed)
          const Padding(
            padding: EdgeInsets.only(top: 12),
            child: Text(
              '✓ Sudah dipelajari',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
          ),
        const SizedBox(height: 20),
        _ContentCard(
          title: 'Tujuan belajar',
          children: [
            for (final objective in lesson.objectives) _Bullet(objective),
          ],
        ),
        for (final section in lesson.sections) ...[
          const SizedBox(height: 22),
          Text(
            section.title,
            style: Theme.of(context)
                .textTheme
                .titleLarge
                ?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 10),
          SelectableText(
            section.body,
            scrollPhysics: const NeverScrollableScrollPhysics(),
            style: const TextStyle(fontSize: 16, height: 1.8),
          ),
        ],
        const SizedBox(height: 22),
        _ContentCard(
          title: 'Alur yang perlu diingat',
          children: [
            for (var i = 0; i < lesson.steps.length; i++)
              _Bullet('${i + 1}. ${lesson.steps[i]}'),
          ],
        ),
        const SizedBox(height: 16),
        _ContentCard(
          title: 'Perhatikan',
          color: colors.tertiaryContainer,
          titleColor: colors.onTertiaryContainer,
          children: [
            Text(
              lesson.caution,
              style: TextStyle(height: 1.6, color: colors.onTertiaryContainer),
            ),
          ],
        ),
        const SizedBox(height: 22),
        _ContentCard(
          title: 'Contoh kasus',
          children: [
            Text(lesson.scenario, style: const TextStyle(height: 1.7)),
            const SizedBox(height: 12),
            const Text(
              'Respons yang tepat',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 6),
            Text(lesson.response, style: const TextStyle(height: 1.7)),
          ],
        ),
        const SizedBox(height: 22),
        _ContentCard(
          title: 'Kosakata bidang',
          children: [
            for (final word in lesson.vocabulary) _JapaneseLine(word),
          ],
        ),
        const SizedBox(height: 16),
        _ContentCard(
          title: 'Ungkapan di tempat kerja',
          children: [
            for (final phrase in lesson.dialogue) _JapaneseLine(phrase),
          ],
        ),
        const SizedBox(height: 26),
        Text(
          'Cek pemahaman',
          style: Theme.of(context)
              .textTheme
              .titleLarge
              ?.copyWith(fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 8),
        const Text(
          'Latihan dalam lesson, bukan simulasi ujian. Jawab semua pertanyaan, baca pembahasannya, dan coba kembali jika belum tepat.',
          style: TextStyle(height: 1.6),
        ),
        const SizedBox(height: 16),
        for (var i = 0; i < lesson.checks.length; i++) ...[
          _ContentCard(
            title: '${i + 1}. ${lesson.checks[i].prompt}',
            children: [
              for (var j = 0; j < lesson.checks[i].options.length; j++)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: OutlinedButton(
                    key: Key('kaigo-answer-$i-$j'),
                    onPressed: submitted || saving
                        ? null
                        : () => setState(() => selections[i] = j),
                    style: OutlinedButton.styleFrom(
                      alignment: Alignment.centerLeft,
                      padding: const EdgeInsets.all(12),
                      backgroundColor:
                          selections[i] == j ? colors.primaryContainer : null,
                      foregroundColor: selections[i] == j
                          ? colors.onPrimaryContainer
                          : colors.onSurface,
                      disabledForegroundColor: colors.onSurface,
                    ),
                    child: Row(
                      children: [
                        Icon(
                          selections[i] == j
                              ? Icons.radio_button_checked
                              : Icons.radio_button_off,
                          size: 20,
                        ),
                        const SizedBox(width: 10),
                        Expanded(child: Text(lesson.checks[i].options[j])),
                      ],
                    ),
                  ),
                ),
              if (submitted) ...[
                Text(
                  selections[i] == lesson.checks[i].answer
                      ? 'Tepat'
                      : 'Belum tepat',
                  style: TextStyle(
                    fontWeight: FontWeight.w700,
                    color: selections[i] == lesson.checks[i].answer
                        ? colors.primary
                        : colors.error,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Jawaban: ${lesson.checks[i].options[lesson.checks[i].answer]}',
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
                const SizedBox(height: 6),
                Text(
                  lesson.checks[i].explanation,
                  style: const TextStyle(height: 1.6),
                ),
              ],
            ],
          ),
          const SizedBox(height: 12),
        ],
        if (!submitted)
          FilledButton(
            key: const Key('kaigo-check'),
            onPressed:
                allAnswered ? () => setState(() => submitted = true) : null,
            child: const Text('Periksa jawaban'),
          )
        else if (!correct)
          OutlinedButton(
            key: const Key('kaigo-retry'),
            onPressed: () => setState(() {
              submitted = false;
              selections.clear();
            }),
            child: const Text('Coba kembali'),
          ),
        if (correct) ...[
          const Text(
            'Semua jawaban tepat. Tinjau kembali materi saat diperlukan, lalu lanjutkan ke lesson berikutnya.',
            style: TextStyle(height: 1.6),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            key: const Key('kaigo-complete'),
            onPressed: saving ? null : complete,
            icon: saving
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.check_rounded),
            label: Text(
              index + 1 == widget.course.lessons.length
                  ? 'Selesaikan kurikulum'
                  : 'Simpan & lanjutkan',
            ),
          ),
        ],
        const SizedBox(height: 16),
        const _StudyNote(),
        const SizedBox(height: 20),
      ],
    );
  }
}

class _ContentCard extends StatelessWidget {
  const _ContentCard({
    required this.title,
    required this.children,
    this.color,
    this.titleColor,
  });
  final String title;
  final List<Widget> children;
  final Color? color;
  final Color? titleColor;
  @override
  Widget build(BuildContext context) => Card(
        margin: EdgeInsets.zero,
        color: color,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                title,
                style: TextStyle(
                  fontWeight: FontWeight.w800,
                  fontSize: 17,
                  height: 1.5,
                  color: titleColor,
                ),
              ),
              const SizedBox(height: 10),
              ...children,
            ],
          ),
        ),
      );
}

class _Bullet extends StatelessWidget {
  const _Bullet(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Text('• $text', style: const TextStyle(height: 1.6)),
      );
}

class _JapaneseLine extends StatelessWidget {
  const _JapaneseLine(this.word);
  final KaigoWord word;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SelectableText(
              word.japanese,
              scrollPhysics: const NeverScrollableScrollPhysics(),
              style: const TextStyle(
                fontSize: 19,
                fontWeight: FontWeight.w700,
                height: 1.7,
              ),
            ),
            Text(
              word.reading,
              style: TextStyle(
                color: Theme.of(context).colorScheme.primary,
                height: 1.6,
              ),
            ),
            const SizedBox(height: 3),
            Text(word.meaning, style: const TextStyle(height: 1.6)),
          ],
        ),
      );
}

class _StudyNote extends StatelessWidget {
  const _StudyNote();
  @override
  Widget build(BuildContext context) => Text(
        'Materi belajar persiapan Kaigo, ditulis untuk Nihongo Master. Praktik bantuan tubuh memerlukan pelatihan dan supervisi. Di tempat kerja, ikuti rencana individual, SOP, dan arahan petugas berwenang.',
        style: TextStyle(
          height: 1.6,
          fontSize: 13,
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      );
}

class _Sources extends StatelessWidget {
  const _Sources({required this.course});
  final KaigoCurriculum course;
  @override
  Widget build(BuildContext context) => ExpansionTile(
        title: const Text('Referensi dan cakupan materi'),
        children: [
          const Padding(
            padding: EdgeInsets.all(12),
            child: Text(
              'Cakupan mengacu pada dasar Kaigo, tubuh dan mental, komunikasi, dukungan kehidupan, serta bahasa Jepang Kaigo. Materi dan soal di aplikasi ditulis mandiri, bukan salinan teks atau bank soal resmi. Ditinjau 11 Oktober 2026.',
              style: TextStyle(height: 1.6),
            ),
          ),
          for (final source in course.sources)
            ListTile(
              title: Text(source.title),
              trailing: const Icon(Icons.open_in_new_rounded, size: 18),
              onTap: () async {
                var opened = false;
                try {
                  opened = await launchUrl(
                    Uri.parse(source.url),
                    mode: LaunchMode.externalApplication,
                  );
                } on Object {
                  opened = false;
                }
                if (!opened && context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Tautan belum dapat dibuka. Coba kembali.'),
                    ),
                  );
                }
              },
            ),
        ],
      );
}

class _LoadError extends StatelessWidget {
  const _LoadError({required this.onRetry});
  final VoidCallback onRetry;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Materi Kaigo belum dapat dimuat.'),
            const SizedBox(height: 12),
            OutlinedButton(onPressed: onRetry, child: const Text('Muat ulang')),
          ],
        ),
      );
}
