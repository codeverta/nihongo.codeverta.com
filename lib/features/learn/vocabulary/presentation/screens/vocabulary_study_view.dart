import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_art.dart';
import 'package:nihongo_master/shared/presentation/widgets/japanese_ruby_text.dart';

class VocabularyStudyView extends ConsumerStatefulWidget {
  const VocabularyStudyView({
    required this.state,
    required this.previous,
    required this.next,
    required this.startQuiz,
    super.key,
  });
  final VocabularyLessonState state;
  final VoidCallback previous;
  final VoidCallback next;
  final VoidCallback startQuiz;
  @override
  ConsumerState<VocabularyStudyView> createState() => _StudyViewState();
}

class _StudyViewState extends ConsumerState<VocabularyStudyView> {
  @override
  Widget build(BuildContext context) {
    final displayPreferences = ref.watch(displayPreferencesProvider);
    final meaning = displayPreferences.translation;
    final furigana = displayPreferences.furigana;
    final s = widget.state;
    final w = s.lesson.words[s.wordIndex];
    final last = s.wordIndex == s.lesson.words.length - 1;
    final colors = Theme.of(context).colorScheme;
    final examples = w.examples.isEmpty
        ? [
            VocabularyExample(
              japanese: w.example,
              reading: '',
              meaning: w.exampleMeaning,
            ),
          ]
        : w.examples;
    return Column(
      children: [
        LinearProgressIndicator(
          value: (s.wordIndex + 1) / s.lesson.words.length,
          minHeight: 5,
          color: colors.secondary,
          backgroundColor: colors.secondary.withValues(alpha: .1),
        ),
        Expanded(
          child: ListView(
            key: ValueKey('vocabulary-word-${s.wordIndex}'),
            padding: const EdgeInsets.all(20),
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      '${'KATA'.localized(context)} ${s.wordIndex + 1} / ${s.lesson.words.length}',
                      key: const Key('lesson-study-counter'),
                      style: TextStyle(
                        color: colors.secondary,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 1.5,
                      ),
                    ),
                  ),
                  Text(
                    'JLPT ${s.lesson.level}',
                    style: TextStyle(
                      color: colors.secondary,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              Container(
                decoration: BoxDecoration(
                  color: colors.surface,
                  borderRadius: BorderRadius.circular(26),
                  border: Border.all(
                    color: colors.outlineVariant.withValues(alpha: .45),
                  ),
                ),
                clipBehavior: Clip.antiAlias,
                child: Column(
                  children: [
                    if (w.imageAsset.isNotEmpty)
                      VocabularyArt(
                        asset: w.imageAsset,
                        url: w.imageUrl,
                        panel: w.imagePanel,
                        description: w.imageDescription,
                        aspectRatio: 1.6,
                      ),
                    Padding(
                      padding: const EdgeInsets.all(22),
                      child: Column(
                        children: [
                          if (furigana)
                            JapaneseRubyText(
                              text: w.term,
                              reading: w.reading,
                              key: Key('word-${w.id}'),
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                fontSize: w.term.length > 5 ? 34 : 44,
                                height: 1.3,
                                fontWeight: FontWeight.w900,
                                color: colors.secondary,
                              ),
                              rubyStyle: TextStyle(
                                fontSize: 17,
                                color: colors.primary,
                                fontWeight: FontWeight.w600,
                              ),
                            )
                          else
                            Text(
                              w.term,
                              key: Key('word-${w.id}'),
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                fontSize: w.term.length > 5 ? 34 : 44,
                                height: 1.3,
                                fontWeight: FontWeight.w900,
                                color: colors.secondary,
                              ),
                            ),
                          const SizedBox(height: 12),
                          Text(
                            w.meaning,
                            key: const Key('revealed-word-meaning'),
                            textAlign: TextAlign.center,
                            style: const TextStyle(
                              fontSize: 21,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'Pakai dalam kalimat'.localized(context),
                style: const TextStyle(
                  fontSize: 21,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Perhatikan cara kata ini digunakan dalam situasi berbeda.'
                    .localized(context),
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  FilterChip(
                    key: const Key('toggle-furigana'),
                    label: Text('Furigana'.localized(context)),
                    selected: furigana,
                    showCheckmark: false,
                    selectedColor: colors.secondary,
                    labelStyle: TextStyle(
                      color: furigana ? colors.onSecondary : colors.onSurface,
                    ),
                    onSelected: (enabled) => ref
                        .read(displayPreferencesProvider.notifier)
                        .setFurigana(enabled: enabled),
                  ),
                  FilterChip(
                    label: Text('Terjemahan'.localized(context)),
                    selected: meaning,
                    showCheckmark: false,
                    selectedColor: colors.secondary,
                    labelStyle: TextStyle(
                      color: meaning ? colors.onSecondary : colors.onSurface,
                    ),
                    onSelected: (enabled) => ref
                        .read(displayPreferencesProvider.notifier)
                        .setTranslation(enabled: enabled),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              for (final (i, e) in examples.indexed)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: colors.secondary.withValues(alpha: .045),
                      borderRadius: BorderRadius.circular(22),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          '${'CONTOH'.localized(context)} ${i + 1}',
                          style: TextStyle(
                            color: colors.secondary,
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 1.5,
                          ),
                        ),
                        const SizedBox(height: 12),
                        if (furigana)
                          JapaneseRubyText(
                            text: e.japanese,
                            reading: e.reading,
                            key: Key('example-japanese-$i'),
                            style: const TextStyle(
                              fontSize: 23,
                              height: 1.8,
                              fontWeight: FontWeight.w600,
                            ),
                          )
                        else
                          Text(
                            e.japanese,
                            key: Key('example-japanese-$i'),
                            style: const TextStyle(
                              fontSize: 23,
                              height: 1.8,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        if (meaning) ...[
                          const SizedBox(height: 12),
                          Text(
                            e.meaning,
                            style: const TextStyle(fontSize: 16, height: 1.7),
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ExpansionTile(
                tilePadding: EdgeInsets.zero,
                leading: Icon(
                  Icons.lightbulb_outline_rounded,
                  color: colors.primary,
                ),
                title: Text(
                  'Catatan pemakaian'.localized(context),
                  style: const TextStyle(fontWeight: FontWeight.w700),
                ),
                children: [
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(w.note, style: const TextStyle(height: 1.7)),
                  ),
                ],
              ),
              if (s.error != null)
                Text(
                  s.error!.localized(context),
                  style: TextStyle(color: colors.error),
                ),
            ],
          ),
        ),
        SafeArea(
          top: false,
          child: Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
            child: Row(
              children: [
                SizedBox(
                  width: 56,
                  height: 56,
                  child: OutlinedButton(
                    key: const Key('previous-word'),
                    onPressed: s.wordIndex == 0 || s.submitting
                        ? null
                        : widget.previous,
                    style: OutlinedButton.styleFrom(
                      padding: EdgeInsets.zero,
                      minimumSize: const Size(0, 56),
                    ),
                    child: const Icon(Icons.arrow_back_rounded),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: FilledButton.icon(
                    key: Key(last ? 'start-vocabulary-quiz' : 'next-word'),
                    onPressed: s.submitting
                        ? null
                        : last
                        ? widget.startQuiz
                        : widget.next,
                    icon: s.submitting
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Icon(
                            last
                                ? Icons.quiz_outlined
                                : Icons.arrow_forward_rounded,
                          ),
                    label: Text(
                      last
                          ? 'Latihan pemahaman'.localized(context)
                          : 'Kata berikutnya'.localized(context),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
