import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class DiagnosticIntroScreen extends ConsumerStatefulWidget {
  const DiagnosticIntroScreen({super.key});
  @override
  ConsumerState<DiagnosticIntroScreen> createState() => _DiagnosticIntroState();
}

class _DiagnosticIntroState extends ConsumerState<DiagnosticIntroScreen> {
  bool busy = false;
  String? error;
  Future<void> start() async {
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id = await ref.read(diagnosticRepositoryProvider).start();
      if (!mounted) return;
      await context.push('/exams/n5/attempt/$id');
      if (mounted) ref.invalidate(diagnosticOverviewProvider);
    } catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final data = ref.watch(diagnosticOverviewProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('JLPT N5 Diagnostic')),
      body: data.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Padding(
          padding: const EdgeInsets.all(20),
          child: ExamMessage(
            message: diagnosticError(e),
            retry: () => ref.invalidate(diagnosticOverviewProvider),
          ),
        ),
        data: (overview) {
          final exam = overview['exam'] as Map;
          final counts = overview['sectionCounts'] as Map;
          final active = (overview['activeAttemptId'] as String).isNotEmpty;
          return Column(
            children: [
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.all(20),
                  children: [
                    const DiagnosticHero(
                      title: 'Kenali titik\nawal belajarmu.',
                      subtitle:
                          'Kerjakan dengan kemampuanmu saat ini. Hasilnya membantu memilih materi belajar berikutnya.',
                    ),
                    const SizedBox(height: 20),
                    ExamPanel(
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceAround,
                        children: [
                          _stat(
                            Icons.quiz_outlined,
                            '${exam['questionCount']}',
                            'soal',
                          ),
                          _stat(
                            Icons.timer_outlined,
                            '${exam['durationMinutes']}',
                            'menit',
                          ),
                          _stat(Icons.lock_open_rounded, 'Gratis', 'akses'),
                        ],
                      ),
                    ),
                    const SizedBox(height: 24),
                    const ExamHeading('Yang akan kamu kerjakan'),
                    for (final name in [
                      'Kosakata',
                      'Kanji',
                      'Grammar',
                      'Reading',
                    ])
                      Padding(
                        padding: const EdgeInsets.only(bottom: 10),
                        child: ExamPanel(
                          child: Row(
                            children: [
                              CircleAvatar(
                                backgroundColor: examInk.withValues(alpha: .08),
                                foregroundColor: examInk,
                                child: Icon(_icon(name)),
                              ),
                              const SizedBox(width: 14),
                              Expanded(
                                child: Text(
                                  name,
                                  style: const TextStyle(
                                    fontSize: 17,
                                    fontWeight: FontWeight.w800,
                                  ),
                                ),
                              ),
                              Text(
                                '${counts[name]} soal',
                                style: const TextStyle(
                                  color: examInk,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    const SizedBox(height: 14),
                    const ExamPanel(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Sebelum mulai',
                            style: TextStyle(
                              fontWeight: FontWeight.w800,
                              fontSize: 17,
                            ),
                          ),
                          SizedBox(height: 10),
                          Text(
                            '• Pilih satu jawaban untuk setiap soal.\n• Kamu boleh kembali dan mengubah jawaban.\n• Tandai soal yang ingin diperiksa lagi.\n• Jawaban disimpan saat koneksi tersedia.\n• Waktu tetap berjalan saat keluar dari sesi.\n• Soal kosong dihitung salah.',
                            style: TextStyle(height: 1.8),
                          ),
                          SizedBox(height: 12),
                          Text(
                            'Latihan buatan Nihongo Master. Hasil berupa persentase jawaban benar, bukan skor atau prediksi kelulusan resmi JLPT.',
                            style: TextStyle(fontSize: 12, height: 1.5),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              SafeArea(
                top: false,
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (error != null)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: Text(
                            error!,
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.error,
                            ),
                          ),
                        ),
                      FilledButton.icon(
                        onPressed: busy ? null : start,
                        icon: busy
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.play_arrow_rounded),
                        label: Text(
                          active ? 'Lanjutkan sesi' : 'Mulai diagnostic',
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _stat(IconData icon, String value, String label) => Column(
        children: [
          Icon(icon, color: examInk),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(fontSize: 23, fontWeight: FontWeight.w900),
          ),
          Text(label),
        ],
      );
  IconData _icon(String name) => switch (name) {
        'Kosakata' => Icons.translate_rounded,
        'Kanji' => Icons.brush_rounded,
        'Grammar' => Icons.account_tree_outlined,
        _ => Icons.auto_stories_outlined
      };
}
