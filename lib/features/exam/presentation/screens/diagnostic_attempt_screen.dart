import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/screens/diagnostic_result_view.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

class DiagnosticAttemptScreen extends ConsumerStatefulWidget {
  const DiagnosticAttemptScreen({required this.attemptId, super.key});
  final String attemptId;
  @override
  ConsumerState<DiagnosticAttemptScreen> createState() => _AttemptState();
}

class _AttemptState extends ConsumerState<DiagnosticAttemptScreen> {
  Json? session;
  Json? result;
  String? error;
  bool loading = true;
  bool busy = false;
  bool allowPop = false;
  bool expiryAttempted = false;
  int index = 0;
  int secondsAtLoad = 0;
  final watch = Stopwatch();
  Timer? timer;
  final marked = <String>{};
  final scroll = ScrollController();
  List<Json> get questions => diagnosticList(session!['questions']);
  Map<String, dynamic> get answers =>
      session!['answers'] as Map<String, dynamic>;
  int get remaining => (secondsAtLoad - watch.elapsed.inSeconds).clamp(0, 1800);
  DiagnosticRepository get repository => ref.read(diagnosticRepositoryProvider);
  @override
  void initState() {
    super.initState();
    unawaited(load());
  }

  Future<void> load() async {
    timer?.cancel();
    setState(() {
      loading = true;
      error = null;
    });
    try {
      final data = await repository.session(widget.attemptId);
      if (!mounted) return;
      secondsAtLoad = DateTime.parse(data['expiresAt'] as String)
          .difference(DateTime.parse(data['serverTime'] as String))
          .inSeconds;
      watch
        ..reset()
        ..start();
      setState(() {
        session = data;
        result = data['result'] as Json?;
        loading = false;
      });
      if (result == null) {
        timer = Timer.periodic(const Duration(seconds: 1), (_) {
          if (!mounted) return;
          setState(() {});
          if (remaining == 0 && !busy && !expiryAttempted) {
            expiryAttempted = true;
            unawaited(finish());
          }
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          loading = false;
          error = diagnosticError(e);
        });
      }
    }
  }

  Future<void> select(String question, String option) async {
    if (busy || remaining == 0) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      await repository.answer(widget.attemptId, question, option);
      if (mounted) setState(() => answers[question] = option);
    } catch (e) {
      if (mounted) {
        setState(
          () => error =
              '${diagnosticError(e)} Pilihan belum tersimpan; pilih ulang.',
        );
      }
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  Future<void> finish() async {
    if (busy || result != null) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final data = await repository.complete(widget.attemptId);
      if (!mounted) return;
      timer?.cancel();
      setState(() => result = data);
      ref
        ..invalidate(diagnosticOverviewProvider)
        ..invalidate(gamificationProvider)
        ..invalidate(homeDashboardProvider);
    } catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  Future<void> confirmFinish() async {
    final empty = questions.length - answers.length;
    final yes = await showDialog<bool>(
      context: context,
      builder: (c) => AlertDialog(
        title: const Text('Selesaikan diagnostic?'),
        content: Text(
          empty > 0
              ? 'Masih ada $empty soal kosong. Soal kosong dihitung salah. Kamu bisa memeriksanya sebelum mengirim.'
              : 'Semua soal sudah dijawab. Setelah dikirim, jawaban tidak dapat diubah.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(c, false),
            child: const Text('Periksa lagi'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(c, true),
            child: const Text('Kirim jawaban'),
          ),
        ],
      ),
    );
    if ((yes ?? false) && mounted) await finish();
  }

  void jump(int value) {
    setState(() => index = value);
    if (scroll.hasClients) scroll.jumpTo(0);
  }

  Future<void> numbers() async {
    final selected = await showModalBottomSheet<int>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (c) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const ExamHeading('Peta soal'),
              Text(
                '${answers.length}/${questions.length} dijawab • ${marked.length} ditandai',
              ),
              const SizedBox(height: 10),
              const Text('Indigo: dijawab  ·  Bendera: ditandai'),
              const SizedBox(height: 20),
              Wrap(
                spacing: 8,
                runSpacing: 10,
                children: [
                  for (var i = 0; i < questions.length; i++)
                    SizedBox(
                      width: 52,
                      height: 52,
                      child: FilledButton(
                        onPressed: () => Navigator.pop(c, i),
                        style: FilledButton.styleFrom(
                          padding: EdgeInsets.zero,
                          backgroundColor: answers
                                  .containsKey(questions[i]['id'])
                              ? examInk
                              : Theme.of(c).colorScheme.surfaceContainerHighest,
                          foregroundColor:
                              answers.containsKey(questions[i]['id'])
                                  ? Colors.white
                                  : examInk,
                          side: i == index
                              ? BorderSide(
                                  color: Theme.of(c).colorScheme.primary,
                                  width: 2,
                                )
                              : BorderSide.none,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(14),
                          ),
                        ),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              '${i + 1}',
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            if (marked.contains(questions[i]['id']))
                              const Icon(
                                Icons.flag_rounded,
                                size: 12,
                              ),
                          ],
                        ),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: () => Navigator.pop(
                    c,
                    questions.indexWhere(
                      (q) => !answers.containsKey(q['id']),
                    ),
                  ),
                  child: const Text('Ke soal belum dijawab'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
    if (selected != null && selected >= 0 && mounted) jump(selected);
  }

  Future<void> leave() async {
    if (busy) return;
    final yes = await showDialog<bool>(
      context: context,
      builder: (c) => AlertDialog(
        title: const Text('Keluar dari sesi?'),
        content: const Text(
          'Jawaban yang sudah tersimpan bisa dilanjutkan dari halaman Ujian. Timer tetap berjalan. Tanda soal tidak disimpan saat keluar.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(c, false),
            child: const Text('Tetap di sini'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(c, true),
            child: const Text('Keluar'),
          ),
        ],
      ),
    );
    if ((yes ?? false) && mounted) {
      setState(() => allowPop = true);
      ref
        ..invalidate(diagnosticOverviewProvider)
        ..invalidate(gamificationProvider)
        ..invalidate(homeDashboardProvider);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) context.pop();
      });
    }
  }

  @override
  void dispose() {
    timer?.cancel();
    watch.stop();
    scroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final q = session == null ? null : questions[index];
    final expired = remaining == 0;
    return PopScope(
      canPop: allowPop || result != null || session == null,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) unawaited(leave());
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(result == null ? 'Diagnostic N5' : 'Hasil diagnostic'),
          actions: [
            if (session != null && result == null)
              Padding(
                padding: const EdgeInsets.only(right: 16),
                child: Chip(
                  avatar: Icon(
                    Icons.timer_outlined,
                    size: 18,
                    color: remaining < 300
                        ? Theme.of(context).colorScheme.error
                        : examInk,
                  ),
                  label: Text(
                    '${(remaining ~/ 60).toString().padLeft(2, '0')}:${(remaining % 60).toString().padLeft(2, '0')}',
                    style: const TextStyle(
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
              ),
          ],
        ),
        body: loading
            ? const Center(child: CircularProgressIndicator())
            : result != null
                ? DiagnosticResultView(result: result!)
                : session == null
                    ? Padding(
                        padding: const EdgeInsets.all(20),
                        child: ExamMessage(
                          message: error ?? 'Sesi tidak dapat dimuat.',
                          retry: load,
                        ),
                      )
                    : Column(
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(20, 8, 20, 16),
                            child: Column(
                              children: [
                                Row(
                                  children: [
                                    Text(
                                      '${q!['section']}  /  Soal ${index + 1} dari ${questions.length}',
                                      style: const TextStyle(
                                        color: examInk,
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                    const Spacer(),
                                    IconButton(
                                      onPressed: busy ? null : numbers,
                                      icon: const Icon(Icons.grid_view_rounded),
                                      tooltip: 'Peta soal',
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                LinearProgressIndicator(
                                  value: answers.length / questions.length,
                                  color: examInk,
                                  backgroundColor:
                                      examInk.withValues(alpha: .1),
                                  minHeight: 6,
                                  borderRadius: BorderRadius.circular(99),
                                ),
                              ],
                            ),
                          ),
                          Expanded(
                            child: ListView(
                              controller: scroll,
                              padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
                              children: [
                                if (expired)
                                  const Padding(
                                    padding: EdgeInsets.only(bottom: 12),
                                    child: Text(
                                      'Waktu habis. Jawaban tersimpan akan dinilai.',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                  ),
                                if ((q['passage'] as String? ?? '')
                                    .isNotEmpty) ...[
                                  ExamPanel(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Text(
                                          'BACAAN',
                                          style: TextStyle(
                                            color: examGreen,
                                            letterSpacing: 2,
                                            fontWeight: FontWeight.w800,
                                          ),
                                        ),
                                        const SizedBox(height: 14),
                                        Text(
                                          q['passage'] as String,
                                          style: const TextStyle(
                                            fontSize: 20,
                                            height: 1.9,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(height: 16),
                                ],
                                ExamPanel(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        q['prompt'] as String,
                                        style: const TextStyle(
                                          fontSize: 23,
                                          height: 1.7,
                                          fontWeight: FontWeight.w700,
                                        ),
                                      ),
                                      const SizedBox(height: 16),
                                      Text(
                                        'Pilih satu jawaban',
                                        style: TextStyle(
                                          color: Theme.of(context)
                                              .colorScheme
                                              .onSurfaceVariant,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(height: 18),
                                for (final (i, option)
                                    in diagnosticList(q['options']).indexed)
                                  Padding(
                                    padding: const EdgeInsets.only(bottom: 12),
                                    child: _Option(
                                      letter: String.fromCharCode(65 + i),
                                      text: option['text'] as String,
                                      selected:
                                          answers[q['id']] == option['id'],
                                      onTap: busy || expired
                                          ? null
                                          : () => select(
                                                q['id'] as String,
                                                option['id'] as String,
                                              ),
                                    ),
                                  ),
                                Row(
                                  children: [
                                    TextButton.icon(
                                      onPressed: busy
                                          ? null
                                          : () => setState(() {
                                                final id = q['id'] as String;
                                                if (!marked.add(id)) {
                                                  marked.remove(id);
                                                }
                                              }),
                                      icon: Icon(
                                        marked.contains(q['id'])
                                            ? Icons.flag_rounded
                                            : Icons.outlined_flag_rounded,
                                      ),
                                      label: Text(
                                        marked.contains(q['id'])
                                            ? 'Ditandai untuk ditinjau'
                                            : 'Tandai soal',
                                      ),
                                    ),
                                    const Spacer(),
                                    if (busy)
                                      const SizedBox(
                                        width: 18,
                                        height: 18,
                                        child: CircularProgressIndicator(
                                          strokeWidth: 2,
                                        ),
                                      ),
                                  ],
                                ),
                                if (answers.containsKey(q['id']))
                                  const Text(
                                    'Jawaban tersimpan',
                                    style: TextStyle(
                                      color: examGreen,
                                      fontSize: 12,
                                    ),
                                  ),
                                if (error != null)
                                  Text(
                                    error!,
                                    style: TextStyle(
                                      color:
                                          Theme.of(context).colorScheme.error,
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          SafeArea(
                            top: false,
                            child: Padding(
                              padding: const EdgeInsets.all(20),
                              child: Row(
                                children: [
                                  SizedBox(
                                    width: 56,
                                    height: 56,
                                    child: OutlinedButton(
                                      onPressed: busy || index == 0
                                          ? null
                                          : () => jump(index - 1),
                                      style: OutlinedButton.styleFrom(
                                        padding: EdgeInsets.zero,
                                        minimumSize: const Size(0, 56),
                                      ),
                                      child: const Icon(
                                        Icons.arrow_back_rounded,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: FilledButton(
                                      onPressed: busy
                                          ? null
                                          : expired
                                              ? finish
                                              : index == questions.length - 1
                                                  ? confirmFinish
                                                  : () => jump(index + 1),
                                      child: Text(
                                        expired
                                            ? 'Lihat hasil'
                                            : index == questions.length - 1
                                                ? 'Selesaikan ujian'
                                                : 'Soal berikutnya',
                                      ),
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
}

class _Option extends StatelessWidget {
  const _Option({
    required this.letter,
    required this.text,
    required this.selected,
    required this.onTap,
  });
  final String letter;
  final String text;
  final bool selected;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) => Semantics(
        selected: selected,
        button: true,
        child: Material(
          color: selected
              ? examInk.withValues(alpha: .07)
              : Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(20),
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(20),
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(20),
                border: Border.all(
                  color: selected
                      ? examInk
                      : Theme.of(context).colorScheme.outlineVariant,
                  width: selected ? 2 : 1,
                ),
              ),
              child: Row(
                children: [
                  CircleAvatar(
                    radius: 18,
                    backgroundColor: selected
                        ? examInk
                        : Theme.of(context).colorScheme.surfaceContainerHighest,
                    foregroundColor: selected ? Colors.white : examInk,
                    child: Text(
                      letter,
                      style: const TextStyle(fontWeight: FontWeight.w800),
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Text(
                      text,
                      style: const TextStyle(fontSize: 19, height: 1.5),
                    ),
                  ),
                  if (selected)
                    const Icon(Icons.check_circle_rounded, color: examInk),
                ],
              ),
            ),
          ),
        ),
      );
}
