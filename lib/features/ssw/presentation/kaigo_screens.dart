import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/ssw_offline_preparation.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/ssw/data/kaigo_repository.dart';
import 'package:nihongo_master/features/ssw/domain/kaigo_curriculum.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';
import 'package:url_launcher/url_launcher.dart';

String kaigoLessonPath(String id) =>
    '/ssw/nursing-care/lessons/${Uri.encodeComponent(id)}';
String sswLessonPath(String sector, String id) =>
    '/ssw/$sector/lessons/${Uri.encodeComponent(id)}';

class KaigoCurriculumPanel extends ConsumerWidget {
  const KaigoCurriculumPanel({super.key, this.sectorId = 'nursing-care'});
  final String sectorId;
  String get label => sectorId == 'nursing-care'
      ? 'Kaigo'
      : sswSectorById(sectorId)?.name ?? 'SSW';

  @override
  Widget build(BuildContext context, WidgetRef ref) => ref
      .watch(sswCurriculumProvider(sectorId))
      .when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => _LoadError(
          message: error is SswCurriculumTranslationUnavailable
              ? 'Materi SSW belum tersedia dalam bahasa Inggris.'.localized(
                  context,
                )
              : 'Materi SSW belum dapat dimuat.'.localized(context),
          onRetry: () => ref.invalidate(sswCurriculumProvider(sectorId)),
        ),
        data: (course) {
          final completed = ref.watch(
            sectorId == 'nursing-care'
                ? kaigoProgressProvider
                : sswProgressProvider(sectorId),
          );
          final count = course.lessons
              .where((l) => completed.contains(l.id))
              .length;
          final next = course.lessons.firstWhere(
            (lesson) => !completed.contains(lesson.id),
            orElse: () => course.lessons.first,
          );
          final signedIn = ref.watch(
            authControllerProvider.select((s) => s.user != null),
          );
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Kurikulum $label'.localized(context),
                style: Theme.of(
                  context,
                ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 8),
              Text(
                '${course.modules.length} modul · ${course.lessons.length} lesson · ±${course.minutes ~/ 60} jam belajar'
                    .localized(context),
              ),
              SswOfflinePreparation(sectorId: sectorId),
              if (sectorId == 'agriculture' || sectorId == 'fishery') ...[
                const SizedBox(height: 12),
                Text(
                  sectorId == 'agriculture'
                      ? 'Dua jalur: Tanaman (20 lesson) dan Peternakan (20 lesson). Pilih modul sesuai pekerjaan yang ingin kamu persiapkan. Simulasi tiap jalur tersedia terpisah di tab Tes.'
                            .localized(context)
                      : 'Dua jalur: Perikanan Tangkap (15 lesson) dan Budi Daya (15 lesson). Pilih modul sesuai jalurmu; paket simulasi Tangkap dan Budi Daya berbeda.'
                            .localized(context),
                  style: const TextStyle(
                    height: 1.6,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
              const SizedBox(height: 12),
              Text(
                'Pelajari berurutan dari dasar bidang hingga bahasa Jepang di tempat kerja. Tiap lesson berisi tujuan, penjelasan, langkah kerja, kasus, kosakata, dan cek pemahaman.'
                    .localized(context),
                style: TextStyle(height: 1.6),
              ),
              const SizedBox(height: 12),
              Text(
                'Ritme belajar: 1–2 lesson per hari. Baca materi, ucapkan contoh Jepang, jelaskan kembali kasus dengan kata sendiri, lalu kerjakan cek pemahaman. Ulangi bagian yang belum jelas sebelum lanjut.'
                    .localized(context),
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
                        '$count/${course.lessons.length} lesson selesai'
                            .localized(context),
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
                                  .localized(context)
                            : 'Masuk untuk menyimpan progress setelah aplikasi ditutup.'
                                  .localized(context),
                      ),
                      const SizedBox(height: 12),
                      FilledButton.icon(
                        key: const Key('kaigo-continue'),
                        onPressed: () =>
                            context.push(sswLessonPath(sectorId, next.id)),
                        icon: const Icon(Icons.menu_book_rounded),
                        label: Text(
                          count == 0
                              ? 'Mulai belajar'.localized(context)
                              : count == course.lessons.length
                              ? 'Pelajari kembali'.localized(context)
                              : 'Lanjutkan belajar'.localized(context),
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
                      '${course.modules[index].lessons.where((l) => completed.contains(l.id)).length}/${course.modules[index].lessons.length} selesai'
                          .localized(context),
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
                            '±${lesson.minutes} menit · ${lesson.checks.length} cek pemahaman'
                                .localized(context),
                          ),
                          trailing: const Icon(Icons.chevron_right_rounded),
                          onTap: () =>
                              context.push(sswLessonPath(sectorId, lesson.id)),
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
  const KaigoLessonScreen({
    required this.lessonId,
    super.key,
    this.sectorId = 'nursing-care',
  });
  final String sectorId;
  final String lessonId;

  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    appBar: AppBar(
      title: Text(
        sectorId == 'nursing-care'
            ? 'Lesson Kaigo'
            : 'Lesson ${sswSectorById(sectorId)?.name ?? 'SSW'}',
      ),
    ),
    body: ref
        .watch(sswCurriculumProvider(sectorId))
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (_, __) => _LoadError(
            message: 'Materi SSW belum dapat dimuat.'.localized(context),
            onRetry: () => ref.invalidate(sswCurriculumProvider(sectorId)),
          ),
          data: (course) {
            final lesson = course.lessonById(lessonId);
            if (lesson == null) {
              return Center(
                child: Text('Lesson tidak ditemukan.'.localized(context)),
              );
            }
            return _LessonReader(
              key: ValueKey(lesson.id),
              course: course,
              sectorId: sectorId,
              lesson: lesson,
            );
          },
        ),
  );
}

class _LessonReader extends ConsumerStatefulWidget {
  const _LessonReader({
    required this.course,
    required this.lesson,
    required this.sectorId,
    super.key,
  });
  final KaigoCurriculum course;
  final KaigoLesson lesson;
  final String sectorId;
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
      await ref
          .read(
            widget.sectorId == 'nursing-care'
                ? kaigoProgressProvider.notifier
                : sswProgressProvider(widget.sectorId).notifier,
          )
          .complete(widget.lesson.id);
      if (!mounted) return;
      if (ref.read(authControllerProvider).user?.id != account) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Akun berubah. Buka kembali materi untuk melanjutkan.'.localized(
                context,
              ),
            ),
          ),
        );
        return;
      }
      final lessons = widget.course.lessons;
      final index = lessons.indexWhere((l) => l.id == widget.lesson.id);
      if (index + 1 < lessons.length) {
        context.pushReplacement(
          sswLessonPath(widget.sectorId, lessons[index + 1].id),
        );
      } else {
        context.go('/ssw/${widget.sectorId}');
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Progress belum tersimpan. Silakan coba lagi.'.localized(context),
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final furigana = ref.watch(displayPreferencesProvider).furigana;
    final lesson = widget.lesson;
    final index = widget.course.lessons.indexWhere((l) => l.id == lesson.id);
    final allAnswered = selections.length == lesson.checks.length;
    final correct =
        submitted &&
        List.generate(
          lesson.checks.length,
          (i) => selections[i] == lesson.checks[i].answer,
        ).every((v) => v);
    final completed = ref
        .watch(
          widget.sectorId == 'nursing-care'
              ? kaigoProgressProvider
              : sswProgressProvider(widget.sectorId),
        )
        .contains(lesson.id);
    final colors = Theme.of(context).colorScheme;
    return ListView(
      key: PageStorageKey('kaigo-reader-${lesson.id}'),
      padding: const EdgeInsets.all(20),
      children: [
        Text(
          'LESSON ${index + 1} / ${widget.course.lessons.length} · ±${lesson.minutes} MENIT'
              .localized(context),
          style: TextStyle(color: colors.primary, fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 12),
        Text(
          lesson.title,
          key: const Key('kaigo-lesson-title'),
          style: Theme.of(
            context,
          ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
        ),
        if (completed)
          Padding(
            padding: EdgeInsets.only(top: 12),
            child: Text(
              '✓ Sudah dipelajari'.localized(context),
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
          ),
        FilterChip(
          key: const Key('ssw-furigana-toggle'),
          label: Text('Furigana'.localized(context)),
          selected: furigana,
          showCheckmark: false,
          selectedColor: colors.secondary,
          labelStyle: TextStyle(
            color: furigana ? colors.onSecondary : colors.onSurface,
          ),
          avatar: Icon(
            Icons.translate_rounded,
            size: 18,
            color: furigana ? colors.onSecondary : colors.primary,
          ),
          onSelected: (enabled) => ref
              .read(displayPreferencesProvider.notifier)
              .setFurigana(enabled: enabled),
        ),
        const SizedBox(height: 20),
        _ContentCard(
          title: 'Tujuan belajar'.localized(context),
          children: [
            for (final objective in lesson.objectives) _Bullet(objective),
          ],
        ),
        for (final section in lesson.sections) ...[
          const SizedBox(height: 22),
          Text(
            section.title,
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w700),
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
          title: 'Alur yang perlu diingat'.localized(context),
          children: [
            for (var i = 0; i < lesson.steps.length; i++)
              _Bullet('${i + 1}. ${lesson.steps[i]}'),
          ],
        ),
        const SizedBox(height: 16),
        _ContentCard(
          title: 'Perhatikan'.localized(context),
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
          title: 'Contoh kasus'.localized(context),
          children: [
            Text(lesson.scenario, style: const TextStyle(height: 1.7)),
            const SizedBox(height: 12),
            Text(
              'Respons yang tepat'.localized(context),
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 6),
            Text(lesson.response, style: const TextStyle(height: 1.7)),
          ],
        ),
        const SizedBox(height: 22),
        _ContentCard(
          title: 'Kosakata bidang'.localized(context),
          children: [
            for (final word in lesson.vocabulary)
              _JapaneseLine(word, showFurigana: furigana),
          ],
        ),
        const SizedBox(height: 16),
        _ContentCard(
          title: 'Ungkapan di tempat kerja'.localized(context),
          children: [
            for (final phrase in lesson.dialogue)
              _JapaneseLine(phrase, showFurigana: furigana),
          ],
        ),
        const SizedBox(height: 26),
        Text(
          'Cek pemahaman'.localized(context),
          style: Theme.of(
            context,
          ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 8),
        Text(
          'Latihan dalam lesson, bukan simulasi ujian. Jawab semua pertanyaan, baca pembahasannya, dan coba kembali jika belum tepat.'
              .localized(context),
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
                      backgroundColor: selections[i] == j
                          ? colors.primaryContainer
                          : null,
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
                      ? 'Tepat'.localized(context)
                      : 'Belum tepat'.localized(context),
                  style: TextStyle(
                    fontWeight: FontWeight.w700,
                    color: selections[i] == lesson.checks[i].answer
                        ? colors.primary
                        : colors.error,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Jawaban: ${lesson.checks[i].options[lesson.checks[i].answer]}'
                      .localized(context),
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
            onPressed: allAnswered
                ? () => setState(() => submitted = true)
                : null,
            child: Text('Periksa jawaban'.localized(context)),
          )
        else if (!correct)
          OutlinedButton(
            key: const Key('kaigo-retry'),
            onPressed: () => setState(() {
              submitted = false;
              selections.clear();
            }),
            child: Text('Coba kembali'.localized(context)),
          ),
        if (correct) ...[
          Text(
            'Semua jawaban tepat. Tinjau kembali materi saat diperlukan, lalu lanjutkan ke lesson berikutnya.'
                .localized(context),
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
                  ? 'Selesaikan kurikulum'.localized(context)
                  : 'Simpan & lanjutkan'.localized(context),
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
  const _JapaneseLine(this.word, {required this.showFurigana});
  final KaigoWord word;
  final bool showFurigana;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 16),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (showFurigana)
          JapaneseRubyText(
            text: word.japanese,
            reading: word.reading,
            style: const TextStyle(
              fontSize: 19,
              fontWeight: FontWeight.w700,
              height: 1.7,
            ),
            rubyStyle: TextStyle(
              fontSize: 10,
              color: Theme.of(context).colorScheme.primary,
            ),
          )
        else
          Text(
            word.japanese,
            style: const TextStyle(
              fontSize: 19,
              fontWeight: FontWeight.w700,
              height: 1.7,
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
    'Materi persiapan SSW ditulis untuk Nihongo Master. Keterampilan lapangan memerlukan pelatihan dan supervisi. Ikuti SOP, spesifikasi produk, serta arahan petugas berwenang.'
        .localized(context),
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
    title: Text('Referensi dan cakupan materi'.localized(context)),
    children: [
      Padding(
        padding: EdgeInsets.all(12),
        child: Text(
          'Cakupan disusun berdasarkan referensi di bawah. Materi dan soal aplikasi ditulis mandiri, bukan salinan buku atau bank soal resmi. Pembaruan SSW: 17 Oktober 2026.'
              .localized(context),
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
                SnackBar(
                  content: Text(
                    'Tautan belum dapat dibuka. Coba kembali.'.localized(
                      context,
                    ),
                  ),
                ),
              );
            }
          },
        ),
    ],
  );
}

class _LoadError extends StatelessWidget {
  const _LoadError({required this.onRetry, required this.message});
  final VoidCallback onRetry;
  final String message;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(24),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(message),
        const SizedBox(height: 12),
        OutlinedButton(
          onPressed: onRetry,
          child: Text('Muat ulang'.localized(context)),
        ),
      ],
    ),
  );
}
