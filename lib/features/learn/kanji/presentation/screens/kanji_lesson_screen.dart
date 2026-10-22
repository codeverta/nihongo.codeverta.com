import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/learn/kanji/domain/kanji_models.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/controllers/kanji_controller.dart';
import 'package:nihongo_master/features/learn/kanji/presentation/screens/kanji_writing_pad.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

class KanjiLessonScreen extends ConsumerWidget {
  const KanjiLessonScreen({required this.lessonId, super.key});
  final String lessonId;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = kanjiLessonProvider(lessonId);
    final lesson = ref.watch(provider);
    final controller = ref.read(provider.notifier);
    return Scaffold(
      appBar: AppBar(
        title: Text(lesson.valueOrNull?.lesson.summary.title ?? 'Lesson Kanji'),
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
          KanjiStage.study => _Study(state: s, controller: controller),
          KanjiStage.quiz => _Quiz(state: s, controller: controller),
          KanjiStage.result => _Result(
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
  final KanjiState state;
  final KanjiController controller;

  @override
  ConsumerState<_Study> createState() => _StudyState();
}

class _StudyState extends ConsumerState<_Study> {
  @override
  Widget build(BuildContext context) {
    final furigana = ref.watch(displayPreferencesProvider).furigana;
    final state = widget.state;
    final controller = widget.controller;
    final point = state.lesson.characters[state.characterIndex];
    final last = state.characterIndex == state.lesson.characters.length - 1;
    final colors = Theme.of(context).colorScheme;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (state.characterIndex + 1) / state.lesson.characters.length,
          backgroundColor: colors.primary.withValues(alpha: 0.1),
          minHeight: 6,
        ),
        Expanded(
          child: ListView(
            key: ValueKey('kanji-point-${state.characterIndex}'),
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                '${state.lesson.summary.chapter} • Kanji ${state.characterIndex + 1}/${state.lesson.characters.length}',
                key: const Key('kanji-point-counter'),
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
                    if (furigana)
                      JapaneseRubyText(
                        text: point.character,
                        reading: _kanjiCardReading(point),
                        key: const Key('kanji-pattern'),
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 80,
                          fontWeight: FontWeight.w800,
                          color: colors.secondary,
                        ),
                        rubyStyle: TextStyle(
                          fontSize: 17,
                          fontWeight: FontWeight.w600,
                          color: colors.primary,
                        ),
                      )
                    else
                      Text(
                        point.character,
                        key: const Key('kanji-pattern'),
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 80,
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
                title: 'Bacaan umum • ${point.strokes} goresan',
                body:
                    'On’yomi: ${point.onReading}\nKun’yomi: ${point.kunReading}',
                icon: Icons.extension_rounded,
              ),
              const SizedBox(height: 16),
              _Section(
                title: 'Petunjuk mengingat',
                body: point.mnemonic,
                icon: Icons.menu_book_rounded,
              ),
              const SizedBox(height: 22),
              FilterChip(
                key: const Key('kanji-furigana-toggle'),
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
                'Contoh kata dan pemakaian',
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
                            key: Key('kanji-example-$index'),
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
                            key: Key('kanji-example-$index'),
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
                title: 'Catatan membaca',
                body: point.note,
                icon: Icons.lightbulb_outline_rounded,
              ),
              const SizedBox(height: 22),
              KanjiWritingPad(
                key: ValueKey(point.character),
                character: point.character,
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
                  key: const Key('kanji-previous'),
                  style: OutlinedButton.styleFrom(padding: EdgeInsets.zero),
                  onPressed: state.busy || state.characterIndex == 0
                      ? null
                      : () => controller.moveCharacter(-1),
                  child: const Icon(Icons.arrow_back_rounded),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  key: Key(last ? 'kanji-start-quiz' : 'kanji-next-point'),
                  style: FilledButton.styleFrom(minimumSize: const Size(0, 52)),
                  onPressed: state.busy
                      ? null
                      : last
                      ? controller.startQuiz
                      : () => controller.moveCharacter(1),
                  icon: state.busy
                      ? const _Spinner()
                      : Icon(
                          last
                              ? Icons.quiz_rounded
                              : Icons.arrow_forward_rounded,
                        ),
                  label: Text(last ? 'Mulai kuis' : 'Kanji berikutnya'),
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
  final KanjiState state;
  final KanjiController controller;
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
                key: const Key('kanji-question-counter'),
                style: TextStyle(
                  color: colors.primary,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 18),
              Text(
                q.prompt,
                key: const Key('kanji-question'),
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
                        key: Key('kanji-option-${option.id}'),
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
                    key: const Key('kanji-feedback'),
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
              key: const Key('kanji-submit'),
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
  final KanjiResult result;
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
          key: const Key('kanji-result'),
          textAlign: TextAlign.center,
          style: const TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 12),
        Text(
          result.passed
              ? 'Progress lesson sudah tersimpan. Lanjutkan ke lesson berikutnya.'
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
                  key: const Key('kanji-score'),
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
          key: const Key('kanji-retry'),
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
          key: const Key('kanji-finish'),
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

String _kanjiCardReading(KanjiCharacter point) {
  for (final example in point.examples) {
    if (example.japanese == point.character && example.reading.isNotEmpty) {
      return example.reading;
    }
  }
  final kun = point.kunReading.split('、').first.split('.').first.trim();
  if (kun.isNotEmpty && kun != '—') return kun;
  return point.onReading.split('、').first.trim();
}
