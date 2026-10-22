import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class SimulationIntroScreen extends ConsumerStatefulWidget {
  const SimulationIntroScreen({super.key, this.level = 'N5'});
  final String level;

  @override
  ConsumerState<SimulationIntroScreen> createState() => _IntroState();
}

class _IntroState extends ConsumerState<SimulationIntroScreen> {
  bool busy = false;
  String? error;
  bool get n4 => widget.level == 'N4';
  bool get n3 => widget.level == 'N3';
  bool get nat => widget.level.startsWith('NAT');
  int get natPaper =>
      widget.level == 'NAT' ? 1 : int.tryParse(widget.level.substring(4)) ?? 1;
  AutoDisposeFutureProvider<Json> get overviewProvider => widget.level != 'N5'
      ? simulationLevelOverviewProvider(widget.level)
      : simulationOverviewProvider;

  Future<void> start() async {
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id = await ref
          .read(simulationLevelRepositoryProvider(widget.level))
          .start();
      if (mounted) {
        ref.invalidate(overviewProvider);
        await context.push(
          '/exams/${widget.level.toLowerCase()}/simulation/$id',
        );
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: Text(
        nat ? 'NAT 5Q Exam $natPaper' : 'JLPT ${widget.level} Simulation',
      ),
    ),
    body: ref
        .watch(overviewProvider)
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (e, _) => Padding(
            padding: const EdgeInsets.all(20),
            child: ExamMessage(
              message: diagnosticError(e),
              retry: () => ref.invalidate(overviewProvider),
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
                      DiagnosticHero(
                        title: 'Duduk. Fokus.\nKerjakan seperti ujian.',
                        subtitle: nat
                            ? 'Exam $natPaper · ${data['questionCount']} soal · ${data['durationMinutes']} menit\nKosakata, kanji, grammar, dan reading NAT 5Q. Listening belum termasuk.'
                            : n3
                            ? 'Paket 01 · 74 soal · 100 menit\nKosakata, kanji, grammar, dan reading N3. Listening dilewati.'
                            : n4
                            ? 'Paket 01 · 57 soal · 80 menit\nKosakata, kanji, grammar, dan reading N4. Listening dilewati.'
                            : 'Paket 01 · 38 soal · 40 menit\nKosakata, kanji, dan grammar N5. Dokkai dan listening dilewati.',
                      ),
                      const SizedBox(height: 24),
                      const ExamHeading('Dua bagian, dua timer'),
                      ExamPanel(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              nat && natPaper > 1
                                  ? '1. 文字・語彙 — 21 soal / 20 menit'
                                  : n3
                                  ? '1. 文字・語彙 — 35 soal / 30 menit'
                                  : n4
                                  ? '1. 文字・語彙 — 28 soal / 25 menit'
                                  : '1. 文字・語彙 — 21 soal / 20 menit',
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              nat && natPaper > 1
                                  ? 'Kosakata, kanji, dan ungkapan dalam konteks'
                                  : n3
                                  ? '8 bacaan kanji · 6 penulisan · 11 konteks · 5 padanan makna · 5 penggunaan kata'
                                  : n4
                                  ? '7 bacaan kanji · 5 penulisan · 8 konteks · 4 padanan makna · 4 penggunaan kata'
                                  : '7 bacaan kanji · 5 penulisan · 6 konteks · 3 padanan makna',
                              style: const TextStyle(height: 1.6),
                            ),
                            const Divider(height: 28),
                            Text(
                              nat
                                  ? '2. 文法・読解 — 22 soal / 40 menit'
                                  : n3
                                  ? '2. 文法・読解 — 39 soal / 70 menit'
                                  : n4
                                  ? '2. 文法・読解 — 29 soal / 55 menit'
                                  : '2. 文法 — 17 soal / 20 menit',
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              nat
                                  ? natPaper > 1
                                        ? 'Grammar, pemahaman teks, dan pencarian informasi'
                                        : '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 5 reading (2 pendek, 2 sedang, 1 pencarian informasi)'
                                  : n3
                                  ? '13 pilihan grammar · 5 susunan kalimat ★ · 5 grammar teks · 16 reading (4 pendek, 6 sedang, 4 panjang, 2 pencarian informasi)'
                                  : n4
                                  ? '13 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 8 reading (3 pendek, 3 sedang, 2 pencarian informasi)'
                                  : '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks',
                              style: const TextStyle(height: 1.6),
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
                      Text(
                        nat
                            ? 'Soal orisinal untuk latihan NAT-TEST 5Q. Dua sesi tertulis: kosakata/kanji 21 soal selama 20 menit, grammar 17 soal dan reading 5 soal selama 40 menit. Listening belum termasuk. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi NAT-TEST.'
                            : n3
                            ? 'Latihan orisinal dengan jenis soal N3. Waktu kosakata 30 menit dan grammar + reading 70 menit. Listening dilewati. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.'
                            : n4
                            ? 'Struktur mengikuti jenis soal N4 dan perkiraan jumlah setelah revisi 2020. Waktu kosakata 25 menit dan grammar + reading 55 menit. Listening dilewati. Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.'
                            : 'Struktur mengikuti jenis soal N5 dan perkiraan jumlah resmi setelah revisi 2020. '
                                  'Grammar teks termasuk bunpou, bukan tes pemahaman dokkai. '
                                  'Waktu grammar disesuaikan menjadi 20 menit karena dokkai dilewati; waktu resmi grammar + dokkai adalah 40 menit. '
                                  'Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.',
                        style: const TextStyle(fontSize: 12, height: 1.6),
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
                            subtitle: Text(_date(r['completedAt'] as String)),
                            trailing: Text(
                              '${r['score']}%',
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                                color: examGreen,
                              ),
                            ),
                            onTap: () => context.push(
                              '/exams/${widget.level.toLowerCase()}/simulation/${r['attemptId']}',
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
