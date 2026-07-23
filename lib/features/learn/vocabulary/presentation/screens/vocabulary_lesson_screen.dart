import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';

class VocabularyLessonScreen extends ConsumerWidget {
  const VocabularyLessonScreen({required this.lessonId, super.key});

  final String lessonId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = vocabularyLessonControllerProvider(lessonId);
    final lesson = ref.watch(provider);
    final controller = ref.read(provider.notifier);
    return Scaffold(
      appBar: AppBar(
        title: Text(lesson.valueOrNull?.lesson.title ?? 'Lesson Vocabulary'),
      ),
      body: SafeArea(
        top: false,
        child: lesson.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, stackTrace) => _LessonError(
            message: error.toString(),
            retry: controller.load,
          ),
          data: (state) => switch (state.stage) {
            VocabularyStage.study => _StudyStage(
                state: state,
                previous: controller.previousWord,
                next: controller.nextWord,
                startQuiz: controller.startQuiz,
              ),
            VocabularyStage.quiz => _QuizStage(
                state: state,
                selectOption: controller.selectOption,
                submit: controller.submitAnswer,
                next: controller.continueQuiz,
              ),
            VocabularyStage.result => _ResultStage(
                result: state.result!,
                retry: controller.retry,
                finish: context.pop,
              ),
          },
        ),
      ),
    );
  }
}

class _StudyStage extends StatelessWidget {
  const _StudyStage({
    required this.state,
    required this.previous,
    required this.next,
    required this.startQuiz,
  });

  final VocabularyLessonState state;
  final VoidCallback previous;
  final VoidCallback next;
  final VoidCallback startQuiz;

  @override
  Widget build(BuildContext context) {
    final word = state.lesson.words[state.wordIndex];
    final last = state.wordIndex == state.lesson.words.length - 1;
    final progress = (state.wordIndex + 1) / state.lesson.words.length;
    return Column(
      children: [
        LinearProgressIndicator(value: progress, minHeight: 6),
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 14, 20, 0),
          child: Row(
            children: [
              Text(
                'Pelajari kata ${state.wordIndex + 1}/${state.lesson.words.length}',
                style: const TextStyle(fontWeight: FontWeight.w700),
              ),
              const Spacer(),
              Text('Nilai terbaik ${state.lesson.bestScore}%'),
            ],
          ),
        ),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: _WordCard(word: word),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
          child: Row(
            children: [
              OutlinedButton(
                onPressed: state.wordIndex == 0 ? null : previous,
                child: const Icon(Icons.arrow_back_rounded),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  key: Key(last ? 'start-vocabulary-quiz' : 'next-word'),
                  onPressed: state.submitting
                      ? null
                      : last
                          ? startQuiz
                          : next,
                  icon: state.submitting
                      ? const SizedBox.square(
                          dimension: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Icon(
                          last
                              ? Icons.quiz_rounded
                              : Icons.arrow_forward_rounded,
                        ),
                  label: Text(last ? 'Mulai kuis' : 'Saya paham, lanjut'),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _WordCard extends StatelessWidget {
  const _WordCard({required this.word});

  final VocabularyWord word;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              word.term,
              key: Key('word-${word.id}'),
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 54, fontWeight: FontWeight.w900),
            ),
            const SizedBox(height: 4),
            Text(
              word.reading,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: colors.primary,
                fontSize: 21,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 18),
            Text(
              word.meaning,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
            ),
            const SizedBox(height: 24),
            _InfoBox(
              icon: Icons.chat_bubble_outline_rounded,
              title: word.example,
              body: word.exampleMeaning,
            ),
            const SizedBox(height: 12),
            _InfoBox(
              icon: Icons.lightbulb_outline_rounded,
              title: 'Catatan pemakaian',
              body: word.note,
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoBox extends StatelessWidget {
  const _InfoBox({
    required this.icon,
    required this.title,
    required this.body,
  });

  final IconData icon;
  final String title;
  final String body;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(15),
      decoration: BoxDecoration(
        color: colors.primary.withValues(alpha: 0.06),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: colors.primary),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 4),
                Text(body),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _QuizStage extends StatelessWidget {
  const _QuizStage({
    required this.state,
    required this.selectOption,
    required this.submit,
    required this.next,
  });

  final VocabularyLessonState state;
  final ValueChanged<String> selectOption;
  final VoidCallback submit;
  final VoidCallback next;

  @override
  Widget build(BuildContext context) {
    final attempt = state.attempt!;
    final question = attempt.questions[state.questionIndex];
    final feedback = state.feedback;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (state.questionIndex + 1) / attempt.questions.length,
          minHeight: 6,
        ),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'Pertanyaan ${state.questionIndex + 1} dari ${attempt.questions.length}',
                  style: Theme.of(context).textTheme.labelLarge,
                ),
                const SizedBox(height: 12),
                Text(
                  question.prompt,
                  key: Key('question-${question.id}'),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.w900,
                      ),
                ),
                const SizedBox(height: 22),
                for (final option in question.options) ...[
                  _OptionTile(
                    option: option,
                    selected: state.selectedOptionId == option.id,
                    feedback: feedback,
                    onTap: () => selectOption(option.id),
                  ),
                  const SizedBox(height: 10),
                ],
                if (feedback != null) ...[
                  const SizedBox(height: 10),
                  _FeedbackCard(feedback: feedback),
                ],
              ],
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
          child: SizedBox(
            width: double.infinity,
            child: FilledButton(
              key: Key(feedback == null ? 'check-answer' : 'continue-quiz'),
              onPressed: state.submitting
                  ? null
                  : feedback == null
                      ? state.selectedOptionId == null
                          ? null
                          : submit
                      : next,
              child: state.submitting
                  ? const SizedBox.square(
                      dimension: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(
                      feedback == null
                          ? 'Periksa jawaban'
                          : state.questionIndex == attempt.questions.length - 1
                              ? 'Lihat hasil'
                              : 'Pertanyaan berikutnya',
                    ),
            ),
          ),
        ),
      ],
    );
  }
}

class _OptionTile extends StatelessWidget {
  const _OptionTile({
    required this.option,
    required this.selected,
    required this.feedback,
    required this.onTap,
  });

  final QuizOption option;
  final bool selected;
  final AnswerFeedback? feedback;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final isCorrect = feedback?.correctOptionId == option.id;
    final selectedWrong = selected && feedback != null && !isCorrect;
    final color = isCorrect
        ? Colors.green
        : selectedWrong
            ? colors.error
            : selected
                ? colors.primary
                : colors.outlineVariant;
    return Material(
      color: color.withValues(alpha: selected || isCorrect ? 0.1 : 0.025),
      borderRadius: BorderRadius.circular(16),
      child: InkWell(
        key: Key('option-${option.id}'),
        onTap: feedback == null ? onTap : null,
        borderRadius: BorderRadius.circular(16),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            border:
                Border.all(color: color, width: selected || isCorrect ? 2 : 1),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  option.text,
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
              ),
              if (isCorrect)
                const Icon(Icons.check_circle_rounded, color: Colors.green),
              if (selectedWrong)
                Icon(Icons.cancel_rounded, color: colors.error),
            ],
          ),
        ),
      ),
    );
  }
}

class _FeedbackCard extends StatelessWidget {
  const _FeedbackCard({required this.feedback});

  final AnswerFeedback feedback;

  @override
  Widget build(BuildContext context) {
    final color =
        feedback.correct ? Colors.green : Theme.of(context).colorScheme.error;
    return Semantics(
      liveRegion: true,
      child: Container(
        key: Key(feedback.correct ? 'answer-correct' : 'answer-wrong'),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.09),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              feedback.correct ? 'Benar!' : 'Belum tepat',
              style: TextStyle(
                color: color,
                fontSize: 18,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 6),
            Text(feedback.explanation),
          ],
        ),
      ),
    );
  }
}

class _ResultStage extends StatelessWidget {
  const _ResultStage({
    required this.result,
    required this.retry,
    required this.finish,
  });

  final VocabularyResult result;
  final VoidCallback retry;
  final VoidCallback finish;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Icon(
            result.passed ? Icons.emoji_events_rounded : Icons.refresh_rounded,
            color: result.passed ? const Color(0xFFE59E24) : colors.primary,
            size: 88,
          ),
          const SizedBox(height: 18),
          Text(
            result.passed ? 'Lesson dikuasai!' : 'Sedikit lagi!',
            style: Theme.of(context)
                .textTheme
                .headlineMedium
                ?.copyWith(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 8),
          Text(
            result.passed
                ? 'Anda sudah memahami kosakata pada lesson ini.'
                : 'Nilai minimum kelulusan adalah 80%. Pelajari feedback lalu coba kembali.',
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 28),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(22),
            decoration: BoxDecoration(
              color: colors.primary.withValues(alpha: 0.08),
              borderRadius: BorderRadius.circular(22),
            ),
            child: Column(
              children: [
                Text(
                  '${result.score}%',
                  key: const Key('vocabulary-score'),
                  style: TextStyle(
                    color: colors.primary,
                    fontSize: 48,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                Text(
                  '${result.correctAnswers} dari ${result.totalQuestions} jawaban benar',
                ),
                const SizedBox(height: 8),
                Text(
                  'Nilai terbaik: ${result.bestScore}% • Progress total: ${result.progress}%',
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            child: FilledButton.icon(
              key: const Key('retry-vocabulary-lesson'),
              onPressed: retry,
              icon: const Icon(Icons.replay_rounded),
              label: const Text('Ulangi lesson'),
            ),
          ),
          const SizedBox(height: 10),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: finish,
              child: const Text('Kembali ke daftar lesson'),
            ),
          ),
        ],
      ),
    );
  }
}

class _LessonError extends StatelessWidget {
  const _LessonError({required this.message, required this.retry});

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
            const Icon(Icons.error_outline_rounded, size: 52),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton(onPressed: retry, child: const Text('Muat ulang')),
          ],
        ),
      ),
    );
  }
}
