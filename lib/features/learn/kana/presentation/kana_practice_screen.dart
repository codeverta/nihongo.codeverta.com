import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/learn/kana/data/kana_repository.dart';
import 'package:nihongo_master/features/learn/kana/domain/kana_models.dart';
import 'package:nihongo_master/features/learn/kana/domain/stroke_matcher.dart';
import 'package:nihongo_master/features/learn/kana/presentation/kana_detail_screen.dart';
import 'package:nihongo_master/features/learn/kana/presentation/widgets/kana_stroke_painter.dart';

/// The three practice stages, from most to least help.
enum KanaPracticeStage { trace, hint, memory }

/// Latihan menulis terpandu: jiplak → petunjuk → dari ingatan.
class KanaPracticeScreen extends ConsumerWidget {
  const KanaPracticeScreen({required this.char, super.key});

  final String char;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final kana = KanaChart.of(scriptOf(char)).find(char);
    final strokes = ref.watch(kanaStrokesProvider);
    return Scaffold(
      appBar: AppBar(
        title: Text('${'Latihan menulis'.localized(context)} $char'),
      ),
      body: switch ((kana, strokes)) {
        (final Kana kana, AsyncData(:final value))
            when value[kana.char] != null =>
          KanaPracticeView(
            key: ValueKey(kana.char),
            kana: kana,
            strokes: value[kana.char]!,
          ),
        (_, AsyncLoading()) => const Center(child: CircularProgressIndicator()),
        _ => Center(
          child: Text(
            'Data urutan goresan belum dapat dimuat.'.localized(context),
          ),
        ),
      },
    );
  }
}

class KanaPracticeView extends ConsumerStatefulWidget {
  const KanaPracticeView({
    required this.kana,
    required this.strokes,
    this.matcher = const StrokeMatcher(),
    super.key,
  });

  final Kana kana;
  final List<Stroke> strokes;
  final StrokeMatcher matcher;

  @override
  ConsumerState<KanaPracticeView> createState() => _KanaPracticeViewState();
}

class _KanaPracticeViewState extends ConsumerState<KanaPracticeView> {
  static const _missesBeforeHelp = 3;

  KanaPracticeStage _stage = KanaPracticeStage.trace;
  int _strokeIndex = 0;
  int _misses = 0;
  bool _helped = false;
  String? _message;
  bool _messageIsError = false;
  Stroke? _live;
  int? _pointer;
  double _side = 1;

  bool get _finished => _strokeIndex >= widget.strokes.length;
  bool get _showHelpStroke => _misses >= _missesBeforeHelp;

  void _restart(KanaPracticeStage stage) => setState(() {
    _stage = stage;
    _strokeIndex = 0;
    _misses = 0;
    _helped = false;
    _message = null;
    _live = null;
  });

  Offset _unit(Offset local) => Offset(
    (local.dx / _side).clamp(0.0, 1.0),
    (local.dy / _side).clamp(0.0, 1.0),
  );

  void _onEnd() {
    final drawn = _live;
    if (drawn == null || _finished) return;
    final verdict = widget.matcher.check(drawn, widget.strokes, _strokeIndex);
    setState(() {
      _live = null;
      if (verdict == StrokeVerdict.correct) {
        _strokeIndex++;
        _misses = 0;
        _messageIsError = false;
        _message = _finished ? null : 'Bagus!'.localized(context);
        if (_finished) _complete();
        return;
      }
      _misses++;
      _messageIsError = true;
      if (_showHelpStroke) {
        if (_stage == KanaPracticeStage.memory) _helped = true;
        _message = 'Lihat contoh goresan berwarna, lalu tiru.'.localized(
          context,
        );
        return;
      }
      _message = switch (verdict) {
        StrokeVerdict.reversed =>
          'Arahnya terbalik. Mulai dari ujung yang lain.',
        StrokeVerdict.wrongOrder =>
          'Urutannya belum tepat. Goresan itu nanti dulu.',
        StrokeVerdict.wrongStart => 'Titik awalnya belum tepat.',
        _ => 'Bentuknya belum pas. Coba lagi pelan-pelan.',
      }.localized(context);
    });
  }

  void _complete() {
    if (_stage == KanaPracticeStage.memory && !_helped) {
      ref.read(kanaProgressProvider.notifier).markPracticed(widget.kana.char);
    }
  }

  String _stageTitle(KanaPracticeStage stage) => switch (stage) {
    KanaPracticeStage.trace => 'Tahap 1: Jiplak',
    KanaPracticeStage.hint => 'Tahap 2: Dengan petunjuk',
    KanaPracticeStage.memory => 'Tahap 3: Dari ingatan',
  }.localized(context);

  String _stageHelp(KanaPracticeStage stage) => switch (stage) {
    KanaPracticeStage.trace =>
      'Ikuti huruf samar. Mulai dari titik merah, sesuai nomor goresan.',
    KanaPracticeStage.hint =>
      'Huruf samar disembunyikan. Titik merah menunjukkan awal goresan berikutnya.',
    KanaPracticeStage.memory =>
      'Tulis tanpa bantuan. Setelah 3 kali salah, contoh goresan akan muncul.',
  }.localized(context);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final total = widget.strokes.length;
    return ListView(
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 32),
      children: [
        Wrap(
          spacing: 6,
          runSpacing: 6,
          children: [
            for (final stage in KanaPracticeStage.values)
              ChoiceChip(
                key: Key('kana-stage-${stage.name}'),
                label: Text('${stage.index + 1}'),
                tooltip: _stageTitle(stage),
                selected: _stage == stage,
                onSelected: (_) => _restart(stage),
              ),
          ],
        ),
        const SizedBox(height: 10),
        Text(
          _stageTitle(_stage),
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 2),
        Text(_stageHelp(_stage)),
        const SizedBox(height: 12),
        Semantics(
          label: '${'Area menulis'.localized(context)} ${widget.kana.char}',
          child: KanaSquare(
            child: LayoutBuilder(
              builder: (context, constraints) {
                _side = constraints.maxWidth;
                // Raw pointer events keep taps and tiny strokes (dakuten);
                // the eager recognizer stops the list from scrolling while drawing.
                return Listener(
                  key: const Key('kana-practice-pad'),
                  onPointerDown: (e) {
                    if (_finished || _pointer != null) return;
                    _pointer = e.pointer;
                    setState(() => _live = [_unit(e.localPosition)]);
                  },
                  onPointerMove: (e) {
                    if (e.pointer != _pointer) return;
                    setState(() => _live = [...?_live, _unit(e.localPosition)]);
                  },
                  onPointerUp: (e) {
                    if (e.pointer != _pointer) return;
                    _pointer = null;
                    _onEnd();
                  },
                  onPointerCancel: (e) {
                    if (e.pointer != _pointer) return;
                    _pointer = null;
                    setState(() => _live = null);
                  },
                  child: RawGestureDetector(
                    gestures: {
                      EagerGestureRecognizer:
                          GestureRecognizerFactoryWithHandlers<
                            EagerGestureRecognizer
                          >(EagerGestureRecognizer.new, (_) {}),
                    },
                    child: CustomPaint(
                      size: Size.infinite,
                      painter: KanaStrokePainter(
                        reference: widget.strokes,
                        colors: colors,
                        showGuide: _stage == KanaPracticeStage.trace,
                        doneCount: _strokeIndex,
                        activeIndex: _finished ? null : _strokeIndex,
                        showActiveStroke:
                            _stage == KanaPracticeStage.trace ||
                            _showHelpStroke,
                        showActiveStart:
                            _stage != KanaPracticeStage.memory ||
                            _showHelpStroke,
                        live: _live,
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Expanded(
              child: Text(
                '${'Goresan'.localized(context)} '
                '${(_strokeIndex + 1).clamp(1, total)} / $total',
                style: const TextStyle(fontWeight: FontWeight.w700),
              ),
            ),
            IconButton(
              tooltip: 'Ulangi tahap ini'.localized(context),
              onPressed: () => _restart(_stage),
              icon: const Icon(Icons.refresh_rounded),
            ),
          ],
        ),
        if (_message case final message?)
          Text(
            message,
            key: const Key('kana-practice-message'),
            style: TextStyle(
              color: _messageIsError ? colors.error : colors.primary,
              fontWeight: FontWeight.w700,
            ),
          ),
        if (_finished) ...[
          const SizedBox(height: 12),
          _CompletionPanel(
            kana: widget.kana,
            stage: _stage,
            helped: _helped,
            onNextStage: _stage == KanaPracticeStage.memory
                ? null
                : () => _restart(KanaPracticeStage.values[_stage.index + 1]),
            onRetry: () => _restart(_stage),
          ),
        ],
      ],
    );
  }
}

class _CompletionPanel extends StatelessWidget {
  const _CompletionPanel({
    required this.kana,
    required this.stage,
    required this.helped,
    required this.onNextStage,
    required this.onRetry,
  });

  final Kana kana;
  final KanaPracticeStage stage;
  final bool helped;
  final VoidCallback? onNextStage;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final memory = stage == KanaPracticeStage.memory;
    final next = kanaNeighbours(kana, writableOnly: true).$2;
    final String title;
    if (!memory) {
      title = 'Tahap selesai!';
    } else if (helped) {
      title =
          'Selesai dengan bantuan. Ulangi tanpa bantuan untuk mendapat centang.';
    } else {
      title = 'Hebat! Kamu bisa menulis huruf ini dari ingatan.';
    }
    return Card(
      key: const Key('kana-practice-complete'),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              title.localized(context),
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 12),
            if (onNextStage case final onNext?)
              FilledButton(
                key: const Key('kana-next-stage'),
                onPressed: onNext,
                child: Text('Lanjut ke tahap berikutnya'.localized(context)),
              )
            else ...[
              if (next != null)
                FilledButton.icon(
                  key: const Key('kana-next-char'),
                  onPressed: () => context.pushReplacement(
                    AppRoutes.kanaPracticePath(next.char),
                  ),
                  icon: const Icon(Icons.arrow_forward_rounded),
                  label: Text.rich(
                    TextSpan(
                      children: [
                        TextSpan(text: 'Huruf berikutnya'.localized(context)),
                        TextSpan(
                          text: ' ${next.char}',
                          style: AppTypography.japanese(),
                        ),
                      ],
                    ),
                  ),
                ),
              const SizedBox(height: 8),
              OutlinedButton(
                onPressed: onRetry,
                child: Text('Ulangi'.localized(context)),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
