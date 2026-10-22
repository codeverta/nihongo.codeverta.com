import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/simulation_result_view.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/simulation_text.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';

class SimulationAttemptScreen extends ConsumerStatefulWidget {
  const SimulationAttemptScreen({
    required this.attemptId,
    super.key,
    this.level = 'N5',
  });
  final String level;
  final String attemptId;

  @override
  ConsumerState<SimulationAttemptScreen> createState() => _AttemptState();
}

class _AttemptState extends ConsumerState<SimulationAttemptScreen>
    with WidgetsBindingObserver {
  Json? session;
  String? error;
  String? notice;
  bool busy = false;
  bool fetching = false;
  bool allowPop = false;
  bool leaveDialogOpen = false;
  int index = 0;
  int millisecondsAtSync = 0;
  int nextRetryAt = 0;
  final clock = Stopwatch();
  final lifetime = Stopwatch();
  final scroll = ScrollController();
  Timer? timer;

  SimulationRepository get repository =>
      ref.read(simulationLevelRepositoryProvider(widget.level));
  bool get isJft => widget.level == 'JFT' || widget.level.startsWith('JFT-');
  int get jftExamNumber => widget.level == 'JFT'
      ? 1
      : int.tryParse(widget.level.substring('JFT-'.length)) ?? 1;
  bool get isNat => widget.level.startsWith('NAT');
  int get natPaper =>
      widget.level == 'NAT' ? 1 : int.tryParse(widget.level.substring(4)) ?? 1;
  bool get isSsw => widget.level.startsWith('SSW-');
  String get sswSectionLabel =>
      diagnosticList(session?['sections'])
          .where((s) => s['id'] == section)
          .map((s) => s['name'] as String)
          .firstOrNull ??
      'Selesai';
  String get jftSectionLabel => switch (section) {
    'vocabulary' => 'Bagian 1 · 文字と語彙',
    'conversation' => 'Bagian 2 · 会話と表現',
    'reading' => 'Bagian 3 · 読解',
    _ => 'Selesai',
  };
  List<Json> get questions => diagnosticList(session!['questions']);
  Map<String, dynamic> get answers => session!['answers'] as Json;
  Set<String> get marked => (session!['marked'] as List).cast<String>().toSet();
  Json? get result => session?['result'] as Json?;
  String get section => session?['activeSection'] as String? ?? '';
  int get remaining =>
      ((millisecondsAtSync - clock.elapsedMilliseconds + 999) ~/ 1000).clamp(
        0,
        86400,
      );

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    lifetime.start();
    unawaited(load());
    timer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted || session == null || result != null) return;
      setState(() {});
      if (remaining == 0 &&
          !busy &&
          !fetching &&
          lifetime.elapsed.inSeconds >= nextRetryAt) {
        nextRetryAt = lifetime.elapsed.inSeconds + 5;
        unawaited(load());
      }
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed &&
        !busy &&
        !fetching &&
        result == null) {
      unawaited(load());
    }
  }

  void applySession(Json data, {int networkMillis = 0}) {
    final previousSection = section;
    final previousResult = result;
    millisecondsAtSync =
        DateTime.parse(data['expiresAt'] as String)
            .difference(DateTime.parse(data['serverTime'] as String))
            .inMilliseconds -
        networkMillis;
    clock
      ..reset()
      ..start();
    setState(() {
      session = data;
      if (previousSection.isNotEmpty && previousSection != section) {
        index = 0;
        notice = isSsw && result == null
            ? 'Bagian sebelumnya dikunci. $sswSectionLabel dimulai dengan timer bagian baru.'
            : isJft && result == null
            ? 'Bagian sebelumnya dikunci. $jftSectionLabel dimulai; timer total terus berjalan.'
            : section == 'grammar'
            ? isNat
                  ? 'Kosakata dikunci. Grammar dan reading dimulai; timer 40 menit berjalan.'
                  : widget.level == 'N3'
                  ? 'Kosakata dikunci. Bagian grammar dan reading sudah dimulai; timer 70 menit berjalan.'
                  : widget.level != 'N5'
                  ? 'Kosakata dikunci. Bagian grammar dan reading sudah dimulai; timer 55 menit berjalan.'
                  : 'Kosakata dikunci. Bagian grammar sudah dimulai; timer 20 menit berjalan.'
            : 'Seluruh bagian selesai. Jawaban tersimpan telah dinilai.';
      }
    });
    if (previousSection != section && scroll.hasClients) scroll.jumpTo(0);
    if (result != null) {
      timer?.cancel();
      if (previousResult == null) {
        ref
          ..invalidate(gamificationProvider)
          ..invalidate(homeDashboardProvider)
          ..invalidate(simulationOverviewProvider)
          ..invalidate(simulationLevelOverviewProvider(widget.level));
      }
    }
  }

  Future<void> load() async {
    if (fetching || busy) return;
    setState(() {
      fetching = true;
      error = null;
    });
    final latency = Stopwatch()..start();
    try {
      final data = await repository.session(widget.attemptId);
      if (mounted) {
        applySession(data, networkMillis: latency.elapsedMilliseconds);
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => fetching = false);
    }
  }

  Future<void> save(Json q, String option, {required bool flag}) async {
    if (busy || fetching || remaining == 0 || result != null) return;
    setState(() {
      busy = true;
      error = null;
    });
    var failed = false;
    try {
      await repository.saveAnswer(
        widget.attemptId,
        q['id'] as String,
        option,
        marked: flag,
      );
      if (!mounted) return;
      setState(() {
        if (option.isEmpty) {
          answers.remove(q['id']);
        } else {
          answers[q['id'] as String] = option;
        }
        final flags = marked;
        if (flag) {
          flags.add(q['id'] as String);
        } else {
          flags.remove(q['id']);
        }
        session!['marked'] = flags.toList();
      });
    } on Object catch (e) {
      failed = true;
      if (mounted) {
        setState(
          () => error =
              '${diagnosticError(e)} Pilihan belum dikonfirmasi tersimpan.',
        );
      }
    } finally {
      if (mounted) setState(() => busy = false);
    }
    if (failed && mounted) {
      final message = error;
      await load();
      if (mounted && result == null) setState(() => error = message);
    }
  }

  Future<void> submit() async {
    if (busy || fetching || result != null) return;
    final expected = section;
    final unanswered = questions.length - answers.length;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (c) => AlertDialog(
        title: Text(
          isSsw
              ? 'Kirim $sswSectionLabel?'
              : isJft
              ? 'Kirim $jftSectionLabel?'
              : expected == 'vocabulary'
              ? 'Kunci bagian kosakata?'
              : widget.level != 'N5'
              ? 'Kirim grammar dan reading?'
              : 'Kirim bagian grammar?',
        ),
        content: Text(
          '${unanswered > 0 ? 'Masih ada $unanswered soal kosong; soal kosong dihitung salah.\n\n' : ''}'
          '${isSsw
              ? (diagnosticList(session!['sections']).last['id'] == expected ? 'Jawaban dikunci dan hasil ditampilkan.' : 'Bagian ini dikunci. Timer bagian berikutnya langsung dimulai.')
              : isJft
              ? expected == 'reading'
                    ? 'Setelah dikirim, jawaban dikunci dan hasil ditampilkan.'
                    : 'Setelah dikirim, kamu tidak bisa kembali. Bagian berikutnya dimulai dengan sisa waktu yang sama.'
              : expected == 'vocabulary'
              ? 'Setelah dikirim, kamu tidak bisa kembali. Timer bagian kedua langsung dimulai.'
              : 'Setelah dikirim, jawaban dikunci dan hasil ditampilkan.'}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(c, false),
            child: const Text('Periksa lagi'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(c, true),
            child: const Text('Kirim bagian'),
          ),
        ],
      ),
    );
    if (!(confirmed ?? false) || !mounted || result != null) return;
    setState(() {
      busy = true;
      error = null;
    });
    final latency = Stopwatch()..start();
    try {
      final data = await repository.submitSection(widget.attemptId, expected);
      if (mounted) {
        applySession(data, networkMillis: latency.elapsedMilliseconds);
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  Future<void> leave() async {
    if (leaveDialogOpen || busy || fetching) return;
    leaveDialogOpen = true;
    final yes = await showDialog<bool>(
      context: context,
      builder: (c) => AlertDialog(
        title: const Text('Keluar dari simulasi?'),
        content: const Text(
          'Timer tetap berjalan. Kamu dapat melanjutkan dari jawaban yang sudah tersimpan melalui halaman simulasi.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(c, false),
            child: const Text('Tetap di ujian'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(c, true),
            child: const Text('Keluar'),
          ),
        ],
      ),
    );
    leaveDialogOpen = false;
    if ((yes ?? false) && mounted) {
      setState(() => allowPop = true);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) context.pop();
      });
    }
  }

  void jump(int value) {
    setState(() => index = value.clamp(0, questions.length - 1));
    if (scroll.hasClients) scroll.jumpTo(0);
  }

  Future<void> map() async {
    final openedSection = section;
    final selected = await showModalBottomSheet<int>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (c) => SafeArea(
        child: SizedBox(
          height: MediaQuery.sizeOf(c).height * .65,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
            children: [
              const ExamHeading('Peta soal bagian aktif'),
              Text(
                '${answers.length}/${questions.length} dijawab · ${marked.length} ditandai',
              ),
              const SizedBox(height: 8),
              const Text('Indigo: terjawab · Bendera: ragu'),
              const SizedBox(height: 20),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  for (var i = 0; i < questions.length; i++)
                    SizedBox(
                      width: 58,
                      height: 58,
                      child: FilledButton(
                        onPressed: () => Navigator.pop(c, i),
                        style: FilledButton.styleFrom(
                          padding: EdgeInsets.zero,
                          backgroundColor:
                              answers.containsKey(questions[i]['id'])
                              ? examInk
                              : Theme.of(c).colorScheme.surfaceContainerHighest,
                          foregroundColor:
                              answers.containsKey(questions[i]['id'])
                              ? Colors.white
                              : examInk,
                        ),
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text('${questions[i]['number']}'),
                            if (marked.contains(questions[i]['id']))
                              const Icon(Icons.flag_rounded, size: 14),
                          ],
                        ),
                      ),
                    ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
    if (selected != null && mounted && openedSection == section) jump(selected);
  }

  @override
  void dispose() {
    timer?.cancel();
    clock.stop();
    lifetime.stop();
    scroll.dispose();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final data = session;
    final q = data == null || result != null
        ? null
        : questions[index.clamp(0, questions.length - 1)];
    final enabled = !busy && !fetching && remaining > 0;
    return PopScope(
      canPop: allowPop || result != null,
      onPopInvokedWithResult: (popped, _) {
        if (!popped) unawaited(leave());
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(
            isSsw
                ? '${result != null ? 'Hasil' : 'Simulasi'} ${sswExamTitle(widget.level)}'
                : isJft
                ? result != null
                      ? 'Hasil JFT Exam $jftExamNumber'
                      : 'JFT Exam $jftExamNumber'
                : isNat
                ? result != null
                      ? 'Hasil NAT 5Q Exam $natPaper'
                      : 'NAT 5Q Exam $natPaper'
                : result != null
                ? 'Hasil simulasi ${widget.level}'
                : 'JLPT ${widget.level} Simulation',
          ),
          actions: [
            if (data != null && result == null)
              Padding(
                padding: const EdgeInsets.only(right: 16),
                child: Center(
                  child: Text(
                    '${(remaining ~/ 60).toString().padLeft(2, '0')}:${(remaining % 60).toString().padLeft(2, '0')}',
                    key: const Key('simulation-timer'),
                    style: TextStyle(
                      fontWeight: FontWeight.w900,
                      color: remaining < 60
                          ? Theme.of(context).colorScheme.error
                          : examInk,
                    ),
                  ),
                ),
              ),
          ],
        ),
        body: data == null
            ? error != null
                  ? Padding(
                      padding: const EdgeInsets.all(20),
                      child: ExamMessage(message: error!, retry: load),
                    )
                  : const Center(child: CircularProgressIndicator())
            : result != null
            ? SimulationResultView(result: result!, level: widget.level)
            : Column(
                children: [
                  if (busy || fetching)
                    const LinearProgressIndicator(minHeight: 2),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(20, 10, 20, 12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                isSsw
                                    ? sswSectionLabel
                                    : isJft
                                    ? jftSectionLabel
                                    : section == 'vocabulary'
                                    ? 'Bagian 1 · 文字・語彙'
                                    : widget.level != 'N5'
                                    ? 'Bagian 2 · 文法・読解'
                                    : 'Bagian 2 · 文法',
                                style: const TextStyle(
                                  color: examInk,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                            ),
                            IconButton(
                              onPressed: busy || fetching ? null : map,
                              icon: const Icon(Icons.grid_view_rounded),
                              tooltip: 'Peta soal',
                            ),
                          ],
                        ),
                        Text(
                          'Soal ${q!['number']} · ${index + 1}/${questions.length} bagian ini · ${answers.length} dijawab',
                        ),
                        const SizedBox(height: 8),
                        LinearProgressIndicator(
                          value: answers.length / questions.length,
                          color: examInk,
                          minHeight: 5,
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: ListView(
                      controller: scroll,
                      padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
                      children: [
                        if (notice != null)
                          _message(notice!, dismissible: true),
                        if (remaining == 0)
                          _message(
                            'Waktu bagian habis. Menyinkronkan bagian berikutnya; jawaban terkunci.',
                          ),
                        if (error != null)
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              _message(error!),
                              TextButton.icon(
                                onPressed: fetching || busy ? null : load,
                                icon: const Icon(Icons.refresh),
                                label: const Text('Periksa jawaban tersimpan'),
                              ),
                            ],
                          ),
                        Text(
                          q['group'] as String,
                          style: const TextStyle(
                            fontWeight: FontWeight.w800,
                            color: examInk,
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          q['instructions'] as String,
                          style: const TextStyle(fontSize: 15, height: 1.8),
                        ),
                        const SizedBox(height: 20),
                        if ((q['passage'] as String? ?? '').isNotEmpty) ...[
                          ExamPanel(
                            child: Text(
                              q['passage'] as String,
                              style: const TextStyle(fontSize: 19, height: 2),
                            ),
                          ),
                          const SizedBox(height: 20),
                        ],
                        SimulationText(
                          q['prompt'] as String,
                          style: const TextStyle(
                            fontSize: 23,
                            height: 1.9,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const SizedBox(height: 20),
                        for (final (i, option) in diagnosticList(
                          q['options'],
                        ).indexed)
                          Padding(
                            padding: const EdgeInsets.only(bottom: 10),
                            child: Semantics(
                              selected: answers[q['id']] == option['id'],
                              button: true,
                              child: OutlinedButton(
                                key: ValueKey(
                                  'simulation-option-${option['id']}',
                                ),
                                onPressed: enabled
                                    ? () => save(
                                        q,
                                        option['id'] as String,
                                        flag: marked.contains(q['id']),
                                      )
                                    : null,
                                style: OutlinedButton.styleFrom(
                                  alignment: Alignment.centerLeft,
                                  padding: const EdgeInsets.all(16),
                                  backgroundColor:
                                      answers[q['id']] == option['id']
                                      ? examInk.withValues(alpha: .08)
                                      : null,
                                  side: BorderSide(
                                    color: answers[q['id']] == option['id']
                                        ? examInk
                                        : Theme.of(
                                            context,
                                          ).colorScheme.outlineVariant,
                                    width: answers[q['id']] == option['id']
                                        ? 2
                                        : 1,
                                  ),
                                ),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '${i + 1}',
                                      style: const TextStyle(
                                        fontSize: 20,
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                    const SizedBox(width: 16),
                                    Expanded(
                                      child: Text(
                                        option['text'] as String,
                                        style: const TextStyle(
                                          fontSize: 20,
                                          height: 1.6,
                                        ),
                                      ),
                                    ),
                                    if (answers[q['id']] == option['id'])
                                      const Icon(
                                        Icons.check_circle_outline,
                                        size: 20,
                                      ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        Row(
                          children: [
                            Expanded(
                              child: TextButton.icon(
                                onPressed: enabled
                                    ? () => save(
                                        q,
                                        answers[q['id']] as String? ?? '',
                                        flag: !marked.contains(q['id']),
                                      )
                                    : null,
                                icon: Icon(
                                  marked.contains(q['id'])
                                      ? Icons.flag_rounded
                                      : Icons.outlined_flag,
                                ),
                                label: Text(
                                  marked.contains(q['id'])
                                      ? 'Ragu ditandai'
                                      : 'Tandai ragu',
                                ),
                              ),
                            ),
                            TextButton(
                              onPressed: enabled && answers.containsKey(q['id'])
                                  ? () => save(
                                      q,
                                      '',
                                      flag: marked.contains(q['id']),
                                    )
                                  : null,
                              child: const Text('Hapus pilihan'),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  SafeArea(
                    top: false,
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 10, 16, 16),
                      child: Row(
                        children: [
                          IconButton(
                            onPressed: index > 0 && !busy && !fetching
                                ? () => jump(index - 1)
                                : null,
                            icon: const Icon(Icons.arrow_back),
                            tooltip: 'Soal sebelumnya',
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: index < questions.length - 1
                                ? FilledButton.icon(
                                    onPressed: busy || fetching
                                        ? null
                                        : () => jump(index + 1),
                                    icon: const Icon(Icons.arrow_forward),
                                    label: const Text('Berikutnya'),
                                  )
                                : FilledButton.icon(
                                    onPressed: enabled ? submit : null,
                                    icon: const Icon(Icons.lock_outline),
                                    label: const Text('Kirim bagian'),
                                  ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _message(String text, {bool dismissible = false}) => Padding(
    padding: const EdgeInsets.only(bottom: 14),
    child: Material(
      color: examInk.withValues(alpha: .06),
      borderRadius: BorderRadius.circular(12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            Expanded(child: Text(text, style: const TextStyle(height: 1.5))),
            if (dismissible)
              IconButton(
                onPressed: () => setState(() => notice = null),
                icon: const Icon(Icons.close),
                tooltip: 'Tutup pemberitahuan',
              ),
          ],
        ),
      ),
    ),
  );
}
