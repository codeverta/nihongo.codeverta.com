import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/online_exam_gate.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/simulation_text.dart';

class SimulationIntroScreen extends ConsumerStatefulWidget {
  const SimulationIntroScreen({super.key, this.level = 'N5'});
  final String level;

  @override
  ConsumerState<SimulationIntroScreen> createState() => _IntroState();
}

class _IntroState extends ConsumerState<SimulationIntroScreen> {
  bool busy = false;
  String? error;
  bool get n4 => jlptBaseLevel(widget.level) == 'N4';
  String get paperLabel => jlptPaperLabel(widget.level);
  bool get n3 => jlptBaseLevel(widget.level) == 'N3';
  bool get nat => widget.level.startsWith('NAT');
  int get natPaper =>
      widget.level == 'NAT' ? 1 : int.tryParse(widget.level.substring(4)) ?? 1;
  AutoDisposeFutureProvider<Json> get overviewProvider => widget.level != 'N5'
      ? simulationLevelOverviewProvider(widget.level)
      : simulationOverviewProvider;

  Future<void> start() async {
    if (!requireOnlineExam(context, ref)) return;
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
        nat
            ? 'NAT 5Q Exam $natPaper'
            : 'JLPT ${jlptBaseLevel(widget.level)}${jlptPaperNumber(widget.level) > 1 ? ' ${'Paket'.localized(context)} $paperLabel' : ''} Simulation',
      ),
    ),
    body: ref
        .watch(overviewProvider)
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (e, _) => Padding(
            padding: const EdgeInsets.all(20),
            child: ExamMessage(
              message: diagnosticError(e).localized(context),
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
                        title: 'Duduk. Fokus.\nKerjakan seperti ujian.'
                            .localized(context),
                        subtitle: nat
                            ? 'Exam $natPaper · ${data['questionCount']} soal · ${data['durationMinutes']} menit\nKosakata, kanji, grammar, dan reading NAT 5Q. Listening belum termasuk.'
                                  .localized(context)
                            : n3
                            ? '${'Paket'.localized(context)} $paperLabel · ${'74 soal · 100 menit'.localized(context)}\n${'Kosakata, kanji, grammar, dan reading N3. Listening dilewati.'.localized(context)}'
                            : n4
                            ? '${'Paket'.localized(context)} $paperLabel · ${'57 soal · 80 menit'.localized(context)}\n${'Kosakata, kanji, grammar, dan reading N4. Listening dilewati.'.localized(context)}'
                            : '${'Paket'.localized(context)} $paperLabel · ${'38 soal · 40 menit'.localized(context)}\n${'Kosakata, kanji, dan grammar N5. Dokkai dan listening dilewati.'.localized(context)}',
                      ),
                      const SizedBox(height: 24),
                      ExamHeading('Dua bagian, dua timer'.localized(context)),
                      ExamPanel(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              (nat && natPaper > 1
                                      ? '1. 文字・語彙 — 21 soal / 20 menit'
                                      : n3
                                      ? '1. 文字・語彙 — 35 soal / 30 menit'
                                      : n4
                                      ? '1. 文字・語彙 — 28 soal / 25 menit'
                                      : '1. 文字・語彙 — 21 soal / 20 menit')
                                  .localized(context),
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              nat && natPaper > 1
                                  ? 'Kosakata, kanji, dan ungkapan dalam konteks'
                                        .localized(context)
                                  : n3
                                  ? '8 bacaan kanji · 6 penulisan · 11 konteks · 5 padanan makna · 5 penggunaan kata'
                                        .localized(context)
                                  : n4
                                  ? '7 bacaan kanji · 5 penulisan · 8 konteks · 4 padanan makna · 4 penggunaan kata'
                                        .localized(context)
                                  : '7 bacaan kanji · 5 penulisan · 6 konteks · 3 padanan makna'
                                        .localized(context),
                              style: const TextStyle(height: 1.6),
                            ),
                            const Divider(height: 28),
                            Text(
                              (nat
                                      ? '2. 文法・読解 — 22 soal / 40 menit'
                                      : n3
                                      ? '2. 文法・読解 — 39 soal / 70 menit'
                                      : n4
                                      ? '2. 文法・読解 — 29 soal / 55 menit'
                                      : '2. 文法 — 17 soal / 20 menit')
                                  .localized(context),
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              nat
                                  ? natPaper > 1
                                        ? 'Grammar, pemahaman teks, dan pencarian informasi'
                                              .localized(context)
                                        : '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 5 reading (2 pendek, 2 sedang, 1 pencarian informasi)'
                                              .localized(context)
                                  : n3
                                  ? '13 pilihan grammar · 5 susunan kalimat ★ · 5 grammar teks · 16 reading (4 pendek, 6 sedang, 4 panjang, 2 pencarian informasi)'
                                        .localized(context)
                                  : n4
                                  ? '13 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 8 reading (3 pendek, 3 sedang, 2 pencarian informasi)'
                                        .localized(context)
                                  : '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks'
                                        .localized(context),
                              style: const TextStyle(height: 1.6),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                      ExamHeading('Aturan simulasi'.localized(context)),
                      ExamPanel(
                        child: Text(
                          '• Instruksi dan soal dalam bahasa Jepang.\n'
                                  '• Pilih satu dari empat jawaban; boleh mengubahnya selama bagian aktif.\n'
                                  '• Gunakan peta soal dan tanda ragu untuk pemeriksaan.\n'
                                  '• Setelah dikirim atau waktu habis, bagian dikunci. Bagian berikutnya langsung dimulai.\n'
                                  '• Timer tetap berjalan saat aplikasi ditutup.\n'
                                  '• Jawaban tersimpan di server setelah koneksi berhasil. Jawaban kosong dihitung salah.\n'
                                  '• Kunci dan pembahasan terbuka setelah kedua bagian selesai.'
                              .localized(context),
                          style: const TextStyle(height: 1.8),
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        nat
                            ? 'Soal orisinal untuk latihan NAT-TEST 5Q. Dua sesi tertulis: kosakata/kanji 21 soal selama 20 menit, grammar 17 soal dan reading 5 soal selama 40 menit. Listening belum termasuk. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi NAT-TEST.'
                                  .localized(context)
                            : n3
                            ? 'Latihan orisinal dengan jenis soal N3. Waktu kosakata 30 menit dan grammar + reading 70 menit. Listening dilewati. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.'
                                  .localized(context)
                            : n4
                            ? 'Struktur mengikuti jenis soal N4 dan perkiraan jumlah setelah revisi 2020. Waktu kosakata 25 menit dan grammar + reading 55 menit. Listening dilewati. Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.'
                                  .localized(context)
                            : 'Struktur mengikuti jenis soal N5 dan perkiraan jumlah resmi setelah revisi 2020. '
                                      'Grammar teks termasuk bunpou, bukan tes pemahaman dokkai. '
                                      'Waktu grammar disesuaikan menjadi 20 menit karena dokkai dilewati; waktu resmi grammar + dokkai adalah 40 menit. '
                                      'Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.'
                                  .localized(context),
                        style: const TextStyle(fontSize: 12, height: 1.6),
                      ),
                      if (history.isNotEmpty) ...[
                        const SizedBox(height: 24),
                        ExamHeading('Riwayat simulasi'.localized(context)),
                        for (final r in history)
                          ListTile(
                            contentPadding: EdgeInsets.zero,
                            title: Text(
                              'Paket 01 · ${r['correct']}/${r['total']} benar'
                                  .localized(context),
                            ),
                            subtitle: Text(
                              _date(context, r['completedAt'] as String),
                            ),
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
                              error!.localized(context),
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
                                ? 'Lanjutkan simulasi'.localized(context)
                                : history.isEmpty
                                ? 'Mulai simulasi'.localized(context)
                                : 'Ulangi paket 01'.localized(context),
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

  String _date(BuildContext context, String raw) => regionalDate(
    DateTime.parse(raw),
    localeName: appRegionLocaleName(context),
  );
}
