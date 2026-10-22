import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/learn/grammar/domain/grammar_models.dart';
import 'package:nihongo_master/features/learn/grammar/presentation/controllers/grammar_controller.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

class GrammarLessonScreen extends ConsumerWidget {
  const GrammarLessonScreen({required this.lessonId, super.key});
  final String lessonId;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = grammarLessonProvider(lessonId);
    final lesson = ref.watch(provider);
    final controller = ref.read(provider.notifier);
    return Scaffold(
      appBar: AppBar(
        title: Text(
          lesson.valueOrNull?.lesson.summary.title ?? 'Lesson Grammar',
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
          GrammarStage.study => _Study(state: s, controller: controller),
          GrammarStage.quiz => _Quiz(state: s, controller: controller),
          GrammarStage.result => _Result(
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
  final GrammarState state;
  final GrammarController controller;

  @override
  ConsumerState<_Study> createState() => _StudyState();
}

class _StudyState extends ConsumerState<_Study> {
  @override
  Widget build(BuildContext context) {
    final furigana = ref.watch(displayPreferencesProvider).furigana;
    final state = widget.state;
    final controller = widget.controller;
    final point = state.lesson.points[state.pointIndex];
    final last = state.pointIndex == state.lesson.points.length - 1;
    final colors = Theme.of(context).colorScheme;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (state.pointIndex + 1) / state.lesson.points.length,
          minHeight: 6,
        ),
        Expanded(
          child: ListView(
            key: ValueKey('grammar-point-${state.pointIndex}'),
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                'JLPT ${state.lesson.summary.level} • ${state.lesson.summary.chapter} • Pola ${state.pointIndex + 1}/${state.lesson.points.length}',
                key: const Key('grammar-point-counter'),
                style: TextStyle(
                  color: colors.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(22),
                decoration: BoxDecoration(
                  color: colors.secondary.withValues(alpha: 0.08),
                  borderRadius: BorderRadius.circular(22),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      point.pattern,
                      key: const Key('grammar-pattern'),
                      style: TextStyle(
                        fontSize: 27,
                        fontWeight: FontWeight.w800,
                        color: colors.secondary,
                      ),
                    ),
                    const SizedBox(height: 10),
                    Text(
                      point.meaning,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: 'Cara membentuk',
                body: point.formation,
                icon: Icons.extension_rounded,
              ),
              const SizedBox(height: 16),
              _Section(
                title: 'Penjelasan',
                body: point.explanation,
                icon: Icons.menu_book_rounded,
              ),
              const SizedBox(height: 22),
              FilterChip(
                key: const Key('grammar-furigana-toggle'),
                label: const Text('Furigana'),
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
              const SizedBox(height: 12),
              const Text(
                'Contoh kalimat',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 12),
              for (final (index, example) in point.examples.indexed) ...[
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(18),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (furigana)
                          JapaneseRubyText(
                            text: example.japanese,
                            reading: example.reading,
                            key: Key(
                              'grammar-example-${state.pointIndex}-$index',
                            ),
                            style: const TextStyle(
                              fontSize: 22,
                              fontWeight: FontWeight.w700,
                              height: 1.5,
                            ),
                            rubyStyle: TextStyle(
                              color: colors.primary,
                              fontSize: 11,
                            ),
                          )
                        else
                          Text(
                            example.japanese,
                            key: Key(
                              'grammar-example-${state.pointIndex}-$index',
                            ),
                            style: const TextStyle(
                              fontSize: 22,
                              fontWeight: FontWeight.w700,
                              height: 1.5,
                            ),
                          ),
                        const SizedBox(height: 10),
                        Text(example.meaning),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 12),
              ],
              _Section(
                title: 'Hindari kesalahan ini',
                body: point.commonMistake,
                icon: Icons.lightbulb_outline_rounded,
              ),
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
                  key: const Key('grammar-previous'),
                  style: OutlinedButton.styleFrom(padding: EdgeInsets.zero),
                  onPressed: state.busy || state.pointIndex == 0
                      ? null
                      : () => controller.movePoint(-1),
                  child: const Icon(Icons.arrow_back_rounded),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  key: Key(last ? 'grammar-start-quiz' : 'grammar-next-point'),
                  style: FilledButton.styleFrom(minimumSize: const Size(0, 52)),
                  onPressed: state.busy
                      ? null
                      : last
                      ? controller.startQuiz
                      : () => controller.movePoint(1),
                  icon: state.busy
                      ? const _Spinner()
                      : Icon(
                          last
                              ? Icons.quiz_rounded
                              : Icons.arrow_forward_rounded,
                        ),
                  label: Text(last ? 'Mulai kuis' : 'Pola berikutnya'),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _Section extends StatelessWidget {
  const _Section({required this.title, required this.body, required this.icon});
  final String title;
  final String body;
  final IconData icon;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Row(
        children: [
          Icon(icon, size: 20, color: Theme.of(context).colorScheme.primary),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              title,
              style: const TextStyle(fontWeight: FontWeight.w800),
            ),
          ),
        ],
      ),
      const SizedBox(height: 8),
      Text(body, style: const TextStyle(height: 1.65)),
    ],
  );
}

class _Quiz extends StatelessWidget {
  const _Quiz({required this.state, required this.controller});
  final GrammarState state;
  final GrammarController controller;
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
          minHeight: 6,
        ),
        Expanded(
          child: ListView(
            key: ValueKey(q.id),
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                'Pertanyaan ${state.questionIndex + 1}/${attempt.questions.length}',
                key: const Key('grammar-question-counter'),
                style: TextStyle(
                  color: colors.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 18),
              Text(
                q.prompt,
                key: const Key('grammar-question'),
                style: const TextStyle(
                  fontSize: 23,
                  fontWeight: FontWeight.w800,
                  height: 1.5,
                ),
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
                        key: Key('grammar-option-${option.id}'),
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
                    key: const Key('grammar-feedback'),
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
              key: const Key('grammar-submit'),
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
  final GrammarResult result;
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
          key: const Key('grammar-result'),
          textAlign: TextAlign.center,
          style: const TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 12),
        Text(
          result.passed
              ? 'Progress lesson sudah tersimpan. Lanjutkan ke pola berikutnya.'
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
                  key: const Key('grammar-score'),
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
          key: const Key('grammar-retry'),
          onPressed: busy ? null : retry,
          icon: busy ? const _Spinner() : const Icon(Icons.refresh_rounded),
          label: const Text('Ulangi kuis'),
        ),
        const SizedBox(height: 12),
        OutlinedButton(
          onPressed: busy ? null : study,
          child: const Text('Pelajari materi lagi'),
        ),
        const SizedBox(height: 12),
        TextButton(
          key: const Key('grammar-finish'),
          onPressed: busy ? null : finish,
          child: const Text('Kembali ke daftar lesson'),
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
