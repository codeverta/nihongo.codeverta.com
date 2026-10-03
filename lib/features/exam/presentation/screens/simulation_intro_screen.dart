import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class SimulationIntroScreen extends ConsumerStatefulWidget {
  const SimulationIntroScreen({super.key});

  @override
  ConsumerState<SimulationIntroScreen> createState() => _IntroState();
}

class _IntroState extends ConsumerState<SimulationIntroScreen> {
  bool busy = false;
  String? error;

  Future<void> start() async {
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id = await ref.read(simulationRepositoryProvider).start();
      if (mounted) {
        ref.invalidate(simulationOverviewProvider);
        await context.push('/exams/n5/simulation/$id');
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('JLPT N5 Simulation')),
        body: ref.watch(simulationOverviewProvider).when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Padding(
                padding: const EdgeInsets.all(20),
                child: ExamMessage(
                  message: diagnosticError(e),
                  retry: () => ref.invalidate(simulationOverviewProvider),
                ),
              ),
              data: (data) {
                final active = (data['activeAttemptId'] as String).isNotEmpty;
                final history = diagnosticList(data['history']);
                return Column(
                  children: [
                    Expanded(
                      child: ListView(
                        padding: const EdgeInsets.all(20),
                        children: [
                          const DiagnosticHero(
                            title: 'Duduk. Fokus.\nKerjakan seperti ujian.',
                            subtitle:
                                'Paket 01 · 38 soal · 40 menit\nKosakata, kanji, dan grammar N5. Dokkai dan listening dilewati.',
                          ),
                          const SizedBox(height: 24),
                          const ExamHeading('Dua bagian, dua timer'),
                          const ExamPanel(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  '1. 文字・語彙 — 21 soal / 20 menit',
                                  style: TextStyle(fontWeight: FontWeight.w800),
                                ),
                                SizedBox(height: 8),
                                Text(
                                  '7 bacaan kanji · 5 penulisan · 6 konteks · 3 padanan makna',
                                  style: TextStyle(height: 1.6),
                                ),
                                Divider(height: 28),
                                Text(
                                  '2. 文法 — 17 soal / 20 menit',
                                  style: TextStyle(fontWeight: FontWeight.w800),
                                ),
                                SizedBox(height: 8),
                                Text(
                                  '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks',
                                  style: TextStyle(height: 1.6),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 20),
                          const ExamHeading('Aturan simulasi'),
                          const ExamPanel(
                            child: Text(
                              '• Instruksi dan soal dalam bahasa Jepang.\n'
                              '• Pilih satu dari empat jawaban; boleh mengubahnya selama bagian aktif.\n'
                              '• Gunakan peta soal dan tanda ragu untuk pemeriksaan.\n'
                              '• Setelah dikirim atau waktu habis, bagian dikunci. Bagian berikutnya langsung dimulai.\n'
                              '• Timer tetap berjalan saat aplikasi ditutup.\n'
                              '• Jawaban tersimpan di server setelah koneksi berhasil. Jawaban kosong dihitung salah.\n'
                              '• Kunci dan pembahasan terbuka setelah kedua bagian selesai.',
                              style: TextStyle(height: 1.8),
                            ),
                          ),
                          const SizedBox(height: 16),
                          const Text(
                            'Struktur mengikuti jenis soal N5 dan perkiraan jumlah resmi setelah revisi 2020. '
                            'Grammar teks termasuk bunpou, bukan tes pemahaman dokkai. '
                            'Waktu grammar disesuaikan menjadi 20 menit karena dokkai dilewati; waktu resmi grammar + dokkai adalah 40 menit. '
                            'Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.',
                            style: TextStyle(fontSize: 12, height: 1.6),
                          ),
                          if (history.isNotEmpty) ...[
                            const SizedBox(height: 24),
                            const ExamHeading('Riwayat simulasi'),
                            for (final r in history)
                              ListTile(
                                contentPadding: EdgeInsets.zero,
                                title: Text(
                                  'Paket 01 · ${r['correct']}/${r['total']} benar',
                                ),
                                subtitle:
                                    Text(_date(r['completedAt'] as String)),
                                trailing: Text(
                                  '${r['score']}%',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.w800,
                                    color: examGreen,
                                  ),
                                ),
                                onTap: () => context.push(
                                  '/exams/n5/simulation/${r['attemptId']}',
                                ),
                              ),
                          ],
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
                                  : const Icon(Icons.timer_outlined),
                              label: Text(
                                active
                                    ? 'Lanjutkan simulasi'
                                    : history.isEmpty
                                        ? 'Mulai simulasi'
                                        : 'Ulangi paket 01',
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

  String _date(String raw) {
    final date = DateTime.parse(raw).toLocal();
    return '${date.day}/${date.month}/${date.year}';
  }
}
