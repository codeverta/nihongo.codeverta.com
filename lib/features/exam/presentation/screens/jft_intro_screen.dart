import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/online_exam_gate.dart';
import 'package:url_launcher/url_launcher.dart';

class JftIntroScreen extends ConsumerStatefulWidget {
  const JftIntroScreen({this.examNumber = 1, super.key});

  final int examNumber;

  @override
  ConsumerState<JftIntroScreen> createState() => _JftIntroState();
}

class _JftIntroState extends ConsumerState<JftIntroScreen> {
  bool busy = false;
  String? error;

  String get level =>
      widget.examNumber == 1 ? 'JFT' : 'JFT-${widget.examNumber}';
  String get route => '/exams/jft/exam-${widget.examNumber}';

  Future<void> start() async {
    if (!requireOnlineExam(context, ref)) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id = await ref
          .read(simulationLevelRepositoryProvider(level))
          .start();
      if (mounted) {
        await context.push('$route/$id');
        if (mounted) ref.invalidate(simulationLevelOverviewProvider(level));
      }
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  Future<void> openSource(String url) async {
    var opened = false;
    try {
      opened = await launchUrl(
        Uri.parse(url),
        mode: LaunchMode.externalApplication,
      );
    } on Object {
      opened = false;
    }
    if (!opened && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Tautan belum dapat dibuka. Coba lagi.'.localized(context),
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text('JFT Exam ${widget.examNumber}')),
    body: ref
        .watch(simulationLevelOverviewProvider(level))
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (e, _) => Padding(
            padding: const EdgeInsets.all(20),
            child: ExamMessage(
              message: diagnosticError(e).localized(context),
              retry: () =>
                  ref.invalidate(simulationLevelOverviewProvider(level)),
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
                      DiagnosticHero(
                        title: 'Bahasa Jepang\nuntuk kehidupan sehari-hari.'
                            .localized(context),
                        subtitle:
                            'Exam ${widget.examNumber} · 36 soal · 45 menit\nLatihan tertulis orisinal JFT-Basic. Listening belum termasuk.'
                                .localized(context),
                      ),
                      const SizedBox(height: 24),
                      ExamHeading('Tiga bagian, satu timer'.localized(context)),
                      ExamPanel(
                        child: Text(
                          '1. 文字と語彙 — 12 soal kosakata dan kanji\n'
                                  '2. 会話と表現 — 12 soal grammar dan ungkapan\n'
                                  '3. 読解 — 12 soal pesan, pengumuman, dan pencarian informasi'
                              .localized(context),
                          style: const TextStyle(height: 1.8),
                        ),
                      ),
                      const SizedBox(height: 20),
                      ExamHeading('Aturan latihan'.localized(context)),
                      ExamPanel(
                        child: Text(
                          '• Waktu total 45 menit, dibagi sendiri untuk ketiga bagian.\n'
                                  '• Jawaban dapat diubah selama bagian masih aktif. Gunakan tanda ragu dan peta soal.\n'
                                  '• Setelah mengirim bagian, kamu tidak dapat kembali. Timer terus berjalan.\n'
                                  '• Timer tetap berjalan saat aplikasi ditutup. Jawaban tersimpan setelah dikonfirmasi server.\n'
                                  '• Soal kosong dihitung salah. Pembahasan tersedia setelah selesai.'
                              .localized(context),
                          style: const TextStyle(height: 1.8),
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Soal orisinal Nihongo Master untuk latihan kehidupan sehari-hari. '
                                'JFT-Basic resmi memiliki empat bagian, sekitar 50 soal, dan waktu 60 menit. '
                                'Paket ini hanya mencakup tiga bagian tertulis dengan waktu latihan 45 menit. '
                                'Skor adalah persentase benar, bukan skor berskala atau kelulusan resmi JFT-Basic.'
                            .localized(context),
                        style: const TextStyle(fontSize: 12, height: 1.6),
                      ),
                      const SizedBox(height: 20),
                      ExamHeading('Sumber rujukan'.localized(context)),
                      ExamPanel(
                        child: Text(
                          'Format dan cakupan kemampuan mengacu pada panduan publik Japan Foundation. '
                                  'Tema komunikasi sehari-hari merujuk secara umum pada Irodori; semua soal di aplikasi ditulis orisinal.'
                              .localized(context),
                          style: const TextStyle(height: 1.6),
                        ),
                      ),
                      TextButton.icon(
                        onPressed: () => openSource(
                          'https://www.jpf.go.jp/jft-basic/e/about/index.html',
                        ),
                        icon: const Icon(Icons.open_in_new_rounded, size: 18),
                        label: Text(
                          'Struktur resmi JFT-Basic'.localized(context),
                        ),
                      ),
                      TextButton.icon(
                        onPressed: () => openSource(
                          'https://www.irodori.jpf.go.jp/en/faq.html',
                        ),
                        icon: const Icon(Icons.open_in_new_rounded, size: 18),
                        label: Text(
                          'Irodori · panduan persiapan'.localized(context),
                        ),
                      ),
                      for (final result in diagnosticList(data['history']))
                        ListTile(
                          contentPadding: EdgeInsets.zero,
                          title: Text(
                            "Exam ${widget.examNumber} · ${result['correct']}/${result['total']} benar"
                                .localized(context),
                          ),
                          subtitle: Text(
                            (result['completedAt'] as String).split('T').first,
                          ),
                          trailing: Text("${result['score']}%"),
                          onTap: () =>
                              context.push("$route/${result['attemptId']}"),
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
                            error!.localized(context),
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.error,
                            ),
                          ),
                        FilledButton.icon(
                          onPressed: busy ? null : start,
                          icon: const Icon(Icons.timer_outlined),
                          label: Text(
                            busy
                                ? 'Menyiapkan ujian...'.localized(context)
                                : active
                                ? '${'Lanjutkan Exam'.localized(context)} ${widget.examNumber}'
                                : '${'Mulai Exam'.localized(context)} ${widget.examNumber}',
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
