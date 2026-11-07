import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_study_view.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_art.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_ruby_text.dart';

class VocabularyLessonScreen extends ConsumerWidget {
  const VocabularyLessonScreen({required this.lessonId, super.key});

  final String lessonId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = vocabularyLessonControllerProvider(lessonId);
    final lesson = ref.watch(provider);
    final controller = ref.read(provider.notifier);
    final state = lesson.valueOrNull;

    final Widget body;
    if (state != null) {
      body = switch (state.stage) {
        VocabularyStage.study => VocabularyStudyView(
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
          error: state.error,
          busy: state.submitting,
          retry: controller.retry,
          finish: context.pop,
        ),
      };
    } else if (lesson.hasError) {
      body = _LessonError(
        message: lesson.error.toString(),
        retry: controller.load,
      );
    } else {
      body = const Center(child: CircularProgressIndicator());
    }

    return Scaffold(
      key: ValueKey(
        state == null
            ? 'lesson-$lessonId-loading'
            : 'lesson-$lessonId-${state.stage.name}',
      ),
      appBar: AppBar(title: Text(state?.lesson.title ?? 'Lesson Vocabulary')),
      body: body,
    );
  }
}

class _QuizStage extends ConsumerWidget {
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
  Widget build(BuildContext context, WidgetRef ref) {
    final attempt = state.attempt!;
    final question = attempt.questions[state.questionIndex];
    final feedback = state.feedback;
    final furigana = ref.watch(displayPreferencesProvider).furigana;
    final annotateQuestion = furigana && question.kind != 'Cara baca';
    return Column(
      children: [
        LinearProgressIndicator(
          backgroundColor: Theme.of(
            context,
          ).colorScheme.secondary.withValues(alpha: .1),
          color: Theme.of(context).colorScheme.secondary,
          value: (state.questionIndex + 1) / attempt.questions.length,
          minHeight: 6,
        ),
        Expanded(
          child: SingleChildScrollView(
            key: ValueKey(question.id),
            primary: false,
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (question.imageAsset.isNotEmpty)
                  ClipRRect(
                    borderRadius: BorderRadius.circular(22),
                    child: VocabularyArt(
                      asset: question.imageAsset,
                      url: question.imageUrl,
                      panel: question.imagePanel,
                      description: question.imageDescription,
                      aspectRatio: 1.6,
                    ),
                  ),
                const SizedBox(height: 14),
                Text(
                  question.kind.localized(context),
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.secondary,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 18),
                Text(
                  '${'Pertanyaan'.localized(context)} ${state.questionIndex + 1} ${'dari'.localized(context)} ${attempt.questions.length}',
                  style: Theme.of(context).textTheme.labelLarge,
                ),
                const SizedBox(height: 12),
                VocabularyRubyText(
                  text: question.prompt,
                  words: state.lesson.words,
                  enabled: annotateQuestion,
                  key: Key('question-${question.id}'),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 22),
                for (final option in question.options) ...[
                  _OptionTile(
                    option: option,
                    words: state.lesson.words,
                    furigana: annotateQuestion,
                    selected: state.selectedOptionId == option.id,
                    feedback: feedback,
                    onTap: () => selectOption(option.id),
                  ),
                  const SizedBox(height: 10),
                ],
                if (state.error != null)
                  Text(
                    state.error!.localized(context),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                if (feedback != null) ...[
                  const SizedBox(height: 10),
                  _FeedbackCard(
                    feedback: feedback,
                    words: state.lesson.words,
                    furigana: furigana,
                  ),
                ],
              ],
            ),
          ),
        ),
        SafeArea(
          top: false,
          child: Padding(
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
                            ? 'Periksa jawaban'.localized(context)
                            : state.questionIndex ==
                                  attempt.questions.length - 1
                            ? 'Lihat hasil'.localized(context)
                            : 'Pertanyaan berikutnya'.localized(context),
                      ),
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
    required this.words,
    required this.furigana,
    required this.selected,
    required this.feedback,
    required this.onTap,
  });

  final QuizOption option;
  final List<VocabularyWord> words;
  final bool furigana;
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
            border: Border.all(
              color: color,
              width: selected || isCorrect ? 2 : 1,
            ),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            children: [
              Expanded(
                child: VocabularyRubyText(
                  text: option.text,
                  words: words,
                  enabled: furigana,
                  style: const TextStyle(
                    fontWeight: FontWeight.w700,
                    fontSize: 19,
                    height: 1.6,
                  ),
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
  const _FeedbackCard({
    required this.feedback,
    required this.words,
    required this.furigana,
  });

  final AnswerFeedback feedback;
  final List<VocabularyWord> words;
  final bool furigana;

  @override
  Widget build(BuildContext context) {
    final color = feedback.correct
        ? Colors.green
        : Theme.of(context).colorScheme.error;
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
              feedback.correct
                  ? 'Benar!'.localized(context)
                  : 'Belum tepat'.localized(context),
              style: TextStyle(
                color: color,
                fontSize: 18,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 6),
            VocabularyRubyText(
              text: feedback.explanation,
              words: words,
              enabled: furigana,
            ),
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
    required this.error,
    required this.busy,
  });

  final VocabularyResult result;
  final VoidCallback retry;
  final VoidCallback finish;
  final String? error;
  final bool busy;

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
            result.pendingSync
                ? 'Hasil sementara'.localized(context)
                : result.passed
                ? 'Lesson dikuasai!'.localized(context)
                : 'Sedikit lagi!'.localized(context),
            style: Theme.of(
              context,
            ).textTheme.headlineMedium?.copyWith(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 8),
          Text(
            result.pendingSync
                ? 'Jawaban tersimpan di perangkat. Kelulusan, XP, dan achievement menunggu verifikasi saat terhubung kembali.'
                      .localized(context)
                : result.passed
                ? 'Anda sudah memahami kosakata pada lesson ini.'.localized(
                    context,
                  )
                : 'Nilai minimum kelulusan adalah 80%. Pelajari feedback lalu coba kembali.'
                      .localized(context),
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
                  '${result.correctAnswers} ${'dari'.localized(context)} ${result.totalQuestions} ${'jawaban benar'.localized(context)}',
                ),
                const SizedBox(height: 8),
                if (!result.pendingSync)
                  Text(
                    '${'Nilai terbaik:'.localized(context)} ${result.bestScore}% • ${'Progress total:'.localized(context)} ${result.progress}%',
                  )
                else
                  Text('Menunggu sinkronisasi'.localized(context)),
              ],
            ),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            child: FilledButton.icon(
              key: const Key('retry-vocabulary-lesson'),
              onPressed: busy ? null : retry,
              icon: const Icon(Icons.replay_rounded),
              label: Text(
                busy
                    ? 'Menyiapkan kuis…'.localized(context)
                    : 'Ulangi lesson'.localized(context),
              ),
            ),
          ),
          const SizedBox(height: 10),
          if (error != null) ...[
            Text(
              error!.localized(context),
              style: TextStyle(color: colors.error),
            ),
            const SizedBox(height: 10),
          ],
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: finish,
              child: Text('Kembali ke daftar lesson'.localized(context)),
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
            Text(message.localized(context), textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: retry,
              child: Text('Muat ulang'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}
