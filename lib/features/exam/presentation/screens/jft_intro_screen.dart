import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class JftIntroScreen extends ConsumerStatefulWidget {
  const JftIntroScreen({super.key});

  @override
  ConsumerState<JftIntroScreen> createState() => _JftIntroState();
}

class _JftIntroState extends ConsumerState<JftIntroScreen> {
  bool busy = false;
  String? error;

  Future<void> start() async {
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id =
          await ref.read(simulationLevelRepositoryProvider('JFT')).start();
      if (mounted) {
        await context.push('/exams/jft/exam-1/$id');
        if (mounted) ref.invalidate(simulationLevelOverviewProvider('JFT'));
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('JFT Exam 1')),
        body: ref.watch(simulationLevelOverviewProvider('JFT')).when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Padding(
                padding: const EdgeInsets.all(20),
                child: ExamMessage(
                  message: diagnosticError(e),
                  retry: () =>
                      ref.invalidate(simulationLevelOverviewProvider('JFT')),
                ),
              ),
              data: (data) {
                final active = (data['activeAttemptId'] as String).isNotEmpty;
                return Column(
                  children: [
                    Expanded(
                      child: ListView(
                        padding: const EdgeInsets.all(20),
                        children: [
                          const DiagnosticHero(
                            title:
                                'Bahasa Jepang\nuntuk kehidupan sehari-hari.',
                            subtitle:
                                'Exam 1 · 36 soal · 45 menit\nLatihan tertulis JFT-Basic. Listening belum termasuk.',
                          ),
                          const SizedBox(height: 24),
                          const ExamHeading('Tiga bagian, satu timer'),
                          const ExamPanel(
                            child: Text(
                              '1. 文字と語彙 — 12 soal kosakata dan kanji\n'
                              '2. 会話と表現 — 12 soal grammar dan ungkapan\n'
                              '3. 読解 — 12 soal pesan, pengumuman, dan pencarian informasi',
                              style: TextStyle(height: 1.8),
                            ),
                          ),
                          const SizedBox(height: 20),
                          const ExamHeading('Aturan latihan'),
                          const ExamPanel(
                            child: Text(
                              '• Waktu total 45 menit, dibagi sendiri untuk ketiga bagian.\n'
                              '• Jawaban dapat diubah selama bagian masih aktif. Gunakan tanda ragu dan peta soal.\n'
                              '• Setelah mengirim bagian, kamu tidak dapat kembali. Timer terus berjalan.\n'
                              '• Timer tetap berjalan saat aplikasi ditutup. Jawaban tersimpan setelah dikonfirmasi server.\n'
                              '• Soal kosong dihitung salah. Pembahasan tersedia setelah selesai.',
                              style: TextStyle(height: 1.8),
                            ),
                          ),
                          const SizedBox(height: 16),
                          const Text(
                            'Soal orisinal Nihongo Master untuk latihan kehidupan sehari-hari. '
                            'JFT-Basic resmi mencakup listening dan memiliki waktu 60 menit. '
                            'Paket tertulis ini menggunakan waktu latihan 45 menit. '
                            'Skor adalah persentase benar, bukan skor berskala atau kelulusan resmi JFT-Basic.',
                            style: TextStyle(fontSize: 12, height: 1.6),
                          ),
                          for (final result in diagnosticList(data['history']))
                            ListTile(
                              contentPadding: EdgeInsets.zero,
                              title: Text(
                                "Exam 1 · ${result['correct']}/${result['total']} benar",
                              ),
                              subtitle: Text(
                                (result['completedAt'] as String)
                                    .split('T')
                                    .first,
                              ),
                              trailing: Text("${result['score']}%"),
                              onTap: () => context.push(
                                "/exams/jft/exam-1/${result['attemptId']}",
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
                              Text(
                                error!,
                                style: TextStyle(
                                  color: Theme.of(context).colorScheme.error,
                                ),
                              ),
                            FilledButton.icon(
                              onPressed: busy ? null : start,
                              icon: const Icon(Icons.timer_outlined),
                              label: Text(
                                busy
                                    ? 'Menyiapkan ujian...'
                                    : active
                                        ? 'Lanjutkan Exam 1'
                                        : 'Mulai Exam 1',
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
