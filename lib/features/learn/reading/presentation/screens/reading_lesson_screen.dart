import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/learn/reading/domain/reading_models.dart';
import 'package:nihongo_master/features/learn/reading/presentation/controllers/reading_controller.dart';
import 'package:nihongo_master/features/learn/reading/presentation/widgets/reading_illustration.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

class ReadingLessonScreen extends ConsumerWidget {
  const ReadingLessonScreen({required this.lessonId, super.key});
  final String lessonId;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = readingLessonProvider(lessonId);
    final lesson = ref.watch(provider);
    final controller = ref.read(provider.notifier);
    return Scaffold(
      appBar: AppBar(
        title: Text(
          lesson.valueOrNull?.lesson.summary.title ?? 'Cerita Reading',
        ),
      ),
      body: lesson.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, s) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(e.toString(), textAlign: TextAlign.center),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: controller.load,
                  child: const Text('Muat ulang'),
                ),
              ],
            ),
          ),
        ),
        data: (s) => switch (s.stage) {
          ReadingStage.study => _Study(state: s, controller: controller),
          ReadingStage.quiz => _Quiz(state: s, controller: controller),
          ReadingStage.result => _Result(
            result: s.result!,
            retry: controller.startQuiz,
            study: controller.studyAgain,
            finish: () => context.pop(),
            busy: s.busy,
            error: s.error,
          ),
        },
      ),
    );
  }
}

class _Study extends ConsumerStatefulWidget {
  const _Study({required this.state, required this.controller});
  final ReadingState state;
  final ReadingController controller;
  @override
  ConsumerState<_Study> createState() => _StudyState();
}

class _StudyState extends ConsumerState<_Study> {
  @override
  Widget build(BuildContext context) {
    final displayPreferences = ref.watch(displayPreferencesProvider);
    final furigana = displayPreferences.furigana;
    final translation = displayPreferences.translation;
    final state = widget.state;
    final controller = widget.controller;
    final page = state.lesson.pages[state.pageIndex];
    final last = state.pageIndex == state.lesson.pages.length - 1;
    final colors = Theme.of(context).colorScheme;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (state.pageIndex + 1) / state.lesson.pages.length,
          minHeight: 5,
          backgroundColor: colors.primary.withValues(alpha: 0.1),
        ),
        Expanded(
          child: ListView(
            key: ValueKey('reading-page-${state.pageIndex}'),
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                '${state.lesson.summary.chapter} • Adegan ${state.pageIndex + 1}/${state.lesson.pages.length}',
                key: const Key('reading-page-counter'),
                style: TextStyle(
                  color: colors.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 12),
              ClipRRect(
                borderRadius: BorderRadius.circular(24),
                child: ReadingIllustration(
                  asset: page.imageAsset,
                  url: page.imageUrl,
                  panel: page.imagePanel,
                  description: page.imageDescription,
                ),
              ),
              const SizedBox(height: 18),
              Text(
                page.title,
                style: const TextStyle(
                  fontSize: 23,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 6,
                children: [
                  FilterChip(
                    key: const Key('reading-furigana-toggle'),
                    label: const Text('Furigana'),
                    selected: furigana,
                    showCheckmark: false,
                    selectedColor: colors.secondary,
                    labelStyle: TextStyle(
                      color: furigana ? colors.onSecondary : colors.onSurface,
                    ),
                    onSelected: (enabled) => ref
                        .read(displayPreferencesProvider.notifier)
                        .setFurigana(enabled: enabled),
                    avatar: Icon(
                      Icons.translate_rounded,
                      size: 18,
                      color: furigana ? colors.onSecondary : colors.primary,
                    ),
                  ),
                  FilterChip(
                    label: const Text('Terjemahan'),
                    selected: translation,
                    showCheckmark: false,
                    selectedColor: colors.secondary,
                    labelStyle: TextStyle(
                      color: translation
                          ? colors.onSecondary
                          : colors.onSurface,
                    ),
                    onSelected: (enabled) => ref
                        .read(displayPreferencesProvider.notifier)
                        .setTranslation(enabled: enabled),
                    avatar: Icon(
                      Icons.language_rounded,
                      size: 18,
                      color: translation ? colors.onSecondary : colors.primary,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: colors.surfaceContainerLow,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (furigana)
                      JapaneseRubyText(
                        text: page.japanese,
                        reading: page.reading,
                        key: const Key('reading-passage'),
                        style: const TextStyle(
                          fontSize: 23,
                          height: 1.9,
                          letterSpacing: 0.4,
                        ),
                      )
                    else
                      Text(
                        page.japanese,
                        key: const Key('reading-passage'),
                        style: const TextStyle(
                          fontSize: 23,
                          height: 1.9,
                          letterSpacing: 0.4,
                        ),
                      ),
                    if (translation) ...[
                      const SizedBox(height: 16),
                      const Text(
                        'Arti cerita',
                        style: TextStyle(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        page.meaning,
                        key: const Key('reading-translation'),
                        style: const TextStyle(fontSize: 16, height: 1.7),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 22),
              const Text(
                'Kata untuk diingat',
                style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18),
              ),
              const SizedBox(height: 12),
              for (final word in page.vocabulary) ...[
                Card(
                  margin: EdgeInsets.zero,
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (furigana)
                          JapaneseRubyText(
                            text: word.japanese,
                            reading: word.reading,
                            style: const TextStyle(
                              fontSize: 22,
                              fontWeight: FontWeight.w800,
                            ),
                            rubyStyle: TextStyle(
                              color: colors.primary,
                              fontSize: 12,
                            ),
                          )
                        else
                          Text(
                            word.japanese,
                            style: const TextStyle(
                              fontSize: 22,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        const SizedBox(height: 8),
                        Text(word.meaning),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 10),
              ],
            ],
          ),
        ),
        _Footer(
          error: state.error,
          child: Row(
            children: [
              SizedBox(
                width: 56,
                child: OutlinedButton(
                  key: const Key('reading-previous'),
                  style: OutlinedButton.styleFrom(padding: EdgeInsets.zero),
                  onPressed: state.busy || state.pageIndex == 0
                      ? null
                      : () => controller.movePage(-1),
                  child: const Icon(Icons.arrow_back_rounded),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  key: Key(last ? 'reading-start-quiz' : 'reading-next-page'),
                  style: FilledButton.styleFrom(minimumSize: const Size(0, 52)),
                  onPressed: state.busy
                      ? null
                      : last
                      ? controller.startQuiz
                      : () => controller.movePage(1),
                  icon: state.busy
                      ? const _Spinner()
                      : Icon(
                          last
                              ? Icons.quiz_rounded
                              : Icons.arrow_forward_rounded,
                        ),
                  label: Text(last ? 'Uji pemahaman' : 'Lanjutkan cerita'),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _Quiz extends StatelessWidget {
  const _Quiz({required this.state, required this.controller});
  final ReadingState state;
  final ReadingController controller;
  @override
  Widget build(BuildContext context) {
    final attempt = state.attempt!;
    final q = attempt.questions[state.questionIndex];
    final feedback = state.feedback;
    final colors = Theme.of(context).colorScheme;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (state.questionIndex + 1) / attempt.questions.length,
          backgroundColor: colors.primary.withValues(alpha: 0.1),
          minHeight: 6,
        ),
        Expanded(
          child: ListView(
            key: ValueKey(q.id),
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                'Pertanyaan ${state.questionIndex + 1}/${attempt.questions.length}',
                key: const Key('reading-question-counter'),
                style: TextStyle(
                  color: colors.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 18),
              Text(
                q.prompt,
                key: const Key('reading-question'),
                style: const TextStyle(
                  fontSize: 23,
                  fontWeight: FontWeight.w800,
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 16),
              ExpansionTile(
                title: const Text('Baca cerita lagi'),
                leading: const Icon(Icons.menu_book_rounded),
                childrenPadding: const EdgeInsets.all(16),
                children: [
                  Text(
                    state.lesson.pages.map((p) => p.japanese).join('\n\n'),
                    style: const TextStyle(fontSize: 18, height: 1.8),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              for (final option in q.options) ...[
                Builder(
                  builder: (context) {
                    final selected = state.selection == option.id;
                    final correct = feedback?.correctOptionId == option.id;
                    final color = correct
                        ? Colors.green
                        : selected && feedback != null
                        ? colors.error
                        : selected
                        ? colors.primary
                        : colors.outlineVariant;
                    return Material(
                      color: color.withValues(
                        alpha: selected || correct ? 0.1 : 0.03,
                      ),
                      borderRadius: BorderRadius.circular(16),
                      child: InkWell(
                        key: Key('reading-option-${option.id}'),
                        borderRadius: BorderRadius.circular(16),
                        onTap: state.busy || feedback != null
                            ? null
                            : () => controller.select(option.id),
                        child: Container(
                          padding: const EdgeInsets.all(17),
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: color,
                              width: selected || correct ? 2 : 1,
                            ),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: Text(
                                  option.text,
                                  style: const TextStyle(
                                    fontSize: 17,
                                    height: 1.4,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 8),
                              Icon(
                                correct
                                    ? Icons.check_circle_rounded
                                    : selected
                                    ? Icons.radio_button_checked
                                    : Icons.radio_button_unchecked,
                                color: color,
                              ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 12),
              ],
              if (feedback != null)
                Semantics(
                  liveRegion: true,
                  child: Container(
                    key: const Key('reading-feedback'),
                    padding: const EdgeInsets.all(18),
                    decoration: BoxDecoration(
                      color: (feedback.correct ? Colors.green : colors.error)
                          .withValues(alpha: 0.08),
                      borderRadius: BorderRadius.circular(18),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          feedback.correct ? 'Benar!' : 'Belum tepat',
                          style: TextStyle(
                            fontWeight: FontWeight.w800,
                            fontSize: 19,
                            color: feedback.correct
                                ? Colors.green
                                : colors.error,
                          ),
                        ),
                        const SizedBox(height: 10),
                        Text(
                          feedback.explanation,
                          style: const TextStyle(height: 1.5),
                        ),
                      ],
                    ),
                  ),
                ),
            ],
          ),
        ),
        _Footer(
          error: state.error,
          child: SizedBox(
            width: double.infinity,
            child: FilledButton(
              key: const Key('reading-submit'),
              style: FilledButton.styleFrom(minimumSize: const Size(0, 52)),
              onPressed: state.busy
                  ? null
                  : feedback != null
                  ? controller.nextQuestion
                  : state.selection == null
                  ? null
                  : controller.submit,
              child: state.busy
                  ? const _Spinner()
                  : Text(
                      feedback == null
                          ? 'Periksa jawaban'
                          : state.questionIndex == attempt.questions.length - 1
                          ? 'Lihat hasil'
                          : 'Soal berikutnya',
                    ),
            ),
          ),
        ),
      ],
    );
  }
}

class _Result extends StatelessWidget {
  const _Result({
    required this.result,
    required this.retry,
    required this.study,
    required this.finish,
    required this.busy,
    this.error,
  });
  final ReadingResult result;
  final VoidCallback retry;
  final VoidCallback study;
  final VoidCallback finish;
  final bool busy;
  final String? error;
  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Icon(
          result.passed ? Icons.emoji_events_rounded : Icons.school_rounded,
          size: 80,
          color: result.passed ? const Color(0xFFE59E24) : colors.primary,
        ),
        const SizedBox(height: 20),
        Text(
          result.passed ? 'Lesson dikuasai!' : 'Terus berlatih!',
          key: const Key('reading-result'),
          textAlign: TextAlign.center,
          style: const TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 12),
        Text(
          result.passed
              ? 'Progres cerita sudah tersimpan. Lanjutkan membaca cerita berikutnya.'
              : 'Nilai minimum adalah 80%. Pelajari penjelasan dan coba kuis kembali.',
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 24),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              children: [
                Text(
                  '${result.score}%',
                  key: const Key('reading-score'),
                  style: TextStyle(
                    fontSize: 48,
                    fontWeight: FontWeight.w900,
                    color: colors.primary,
                  ),
                ),
                Text('${result.correct}/${result.total} jawaban benar'),
                const SizedBox(height: 12),
                Text('Nilai terbaik ${result.bestScore}%'),
                Text('Progress kursus ${result.progress}%'),
              ],
            ),
          ),
        ),
        const SizedBox(height: 24),
        if (error != null) ...[
          Text(error!, style: TextStyle(color: colors.error)),
          const SizedBox(height: 12),
        ],
        FilledButton.icon(
          key: const Key('reading-retry'),
          onPressed: busy ? null : retry,
          icon: busy ? const _Spinner() : const Icon(Icons.refresh_rounded),
          label: const Text('Ulangi kuis'),
        ),
        const SizedBox(height: 12),
        OutlinedButton(
          onPressed: busy ? null : study,
          child: const Text('Baca cerita lagi'),
        ),
        const SizedBox(height: 12),
        TextButton(
          key: const Key('reading-finish'),
          onPressed: busy ? null : finish,
          child: const Text('Kembali ke daftar cerita'),
        ),
      ],
    );
  }
}

class _Footer extends StatelessWidget {
  const _Footer({required this.child, this.error});
  final Widget child;
  final String? error;
  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (error != null) ...[
            Text(
              error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
            const SizedBox(height: 8),
          ],
          child,
        ],
      ),
    ),
  );
}

class _Spinner extends StatelessWidget {
  const _Spinner();
  @override
  Widget build(BuildContext context) => const SizedBox.square(
    dimension: 20,
    child: CircularProgressIndicator(strokeWidth: 2),
  );
}
