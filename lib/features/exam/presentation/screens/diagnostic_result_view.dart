import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class DiagnosticResultView extends StatefulWidget {
  const DiagnosticResultView({required this.result, super.key});
  final Json result;
  @override
  State<DiagnosticResultView> createState() => _ResultState();
}

class _ResultState extends State<DiagnosticResultView> {
  bool onlyWrong = false;
  @override
  Widget build(BuildContext context) {
    final r = widget.result;
    final sections = diagnosticList(r['sections']);
    final focus = [...sections]
      ..sort((a, b) => (a['score'] as int).compareTo(b['score'] as int));
    final reviews = diagnosticList(r['reviews']);
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        DiagnosticHero(
          title: (r['score'] as int) >= 80
              ? 'Dasarmu sudah kuat.'.localized(context)
              : 'Langkah berikutnya\nlebih jelas sekarang.'.localized(context),
          subtitle:
              '${r['correct']} dari ${r['total']} soal benar. Gunakan hasil ini untuk menentukan fokus belajar.'
                  .localized(context),
          child: Row(
            children: [
              SizedBox(
                width: 110,
                height: 110,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    SizedBox.expand(
                      child: CircularProgressIndicator(
                        value: (r['score'] as int) / 100,
                        strokeWidth: 8,
                        color: const Color(0xFF9EE5CF),
                        backgroundColor: Colors.white.withValues(alpha: .14),
                      ),
                    ),
                    Text(
                      '${r['score']}%',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 29,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 20),
              Expanded(
                child: Text(
                  '${r['answered']} dijawab\n${(r['total'] as int) - (r['answered'] as int)} soal kosong\n${(r['durationSeconds'] as int) ~/ 60} menit ${(r['durationSeconds'] as int) % 60} detik'
                      .localized(context),
                  style: const TextStyle(color: Colors.white, height: 1.8),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        ExamHeading('Peta kemampuanmu'.localized(context)),
        ExamPanel(
          child: Column(
            children: [
              for (final s in sections)
                Padding(
                  padding: const EdgeInsets.only(bottom: 18),
                  child: Column(
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              s['name'] as String,
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ),
                          Text(
                            '${s['correct']}/${s['total']} benar  •  ${s['score']}%'
                                .localized(context),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      LinearProgressIndicator(
                        value: (s['score'] as int) / 100,
                        color: (s['score'] as int) >= 80 ? examGreen : examInk,
                        backgroundColor: examInk.withValues(alpha: .1),
                        minHeight: 8,
                        borderRadius: BorderRadius.circular(99),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        ExamHeading('Fokus belajar berikutnya'.localized(context)),
        for (final s in focus.take(2))
          Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: ExamPanel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${s['name']} N5',
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _advice(context, s['name'] as String),
                    style: const TextStyle(height: 1.6),
                  ),
                  const SizedBox(height: 10),
                  TextButton.icon(
                    onPressed: () => context.push(_route(s['name'] as String)),
                    icon: const Icon(Icons.arrow_forward_rounded),
                    label: Text('Buka materi belajar'.localized(context)),
                  ),
                ],
              ),
            ),
          ),
        const SizedBox(height: 14),
        Text(
          'Persentase dihitung dari jawaban benar. Diagnostic ini bukan penilaian kelulusan resmi JLPT.'
              .localized(context),
          style: const TextStyle(fontSize: 12, height: 1.6),
        ),
        const SizedBox(height: 24),
        Row(
          children: [
            Expanded(
              child: ExamHeading('Pembahasan jawaban'.localized(context)),
            ),
            FilterChip(
              label: Text('Perlu ditinjau'.localized(context)),
              selected: onlyWrong,
              onSelected: (v) => setState(() => onlyWrong = v),
            ),
          ],
        ),
        if (onlyWrong && reviews.every((q) => q['correct'] == true))
          ExamPanel(
            child: Text(
              'Semua jawaban benar. Kamu bisa membuka seluruh pembahasan.'
                  .localized(context),
            ),
          ),
        for (final (i, review) in reviews.indexed)
          if (!onlyWrong || review['correct'] != true)
            Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: ExamPanel(
                child: ExpansionTile(
                  tilePadding: EdgeInsets.zero,
                  childrenPadding: EdgeInsets.zero,
                  leading: Icon(
                    review['correct'] == true
                        ? Icons.check_circle_rounded
                        : Icons.error_outline_rounded,
                    color: review['correct'] == true
                        ? examGreen
                        : Theme.of(context).colorScheme.error,
                  ),
                  title: Text(
                    '${'Soal'.localized(context)} ${i + 1} • ${(review['question'] as Map)['section']}',
                    style: const TextStyle(fontWeight: FontWeight.w800),
                  ),
                  subtitle: Text(
                    review['selectedOptionId'] == ''
                        ? 'Belum dijawab'.localized(context)
                        : review['correct'] == true
                        ? 'Jawaban benar'.localized(context)
                        : 'Perlu ditinjau'.localized(context),
                  ),
                  children: [_review(context, review)],
                ),
              ),
            ),
        const SizedBox(height: 16),
        OutlinedButton.icon(
          onPressed: () => context.pushReplacement('/exams/n5'),
          icon: const Icon(Icons.refresh_rounded),
          label: Text('Kerjakan ulang diagnostic'.localized(context)),
        ),
      ],
    );
  }

  Widget _review(BuildContext context, Json review) {
    final q = review['question'] as Map;
    final options = diagnosticList(q['options']);
    String text(String id) =>
        options
            .where((o) => o['id'] == id)
            .map((o) => o['text'] as String)
            .firstOrNull ??
        'Tidak dijawab'.localized(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if ((q['passage'] as String? ?? '').isNotEmpty) ...[
          Text(
            q['passage'] as String,
            style: const TextStyle(height: 1.8, fontSize: 17),
          ),
          const SizedBox(height: 16),
        ],
        Text(
          q['prompt'] as String,
          style: const TextStyle(
            fontSize: 19,
            height: 1.8,
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 12),
        Text(
          '${'Jawabanmu:'.localized(context)} ${text(review['selectedOptionId'] as String)}',
        ),
        const SizedBox(height: 8),
        Text(
          '${'Jawaban benar:'.localized(context)} ${text(review['correctOptionId'] as String)}',
          style: const TextStyle(color: examGreen, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 12),
        Text(
          review['explanation'] as String,
          style: const TextStyle(height: 1.7),
        ),
        const SizedBox(height: 12),
      ],
    );
  }

  String _route(String name) => switch (name) {
    'Kosakata' => '/learn/vocabulary',
    'Kanji' => '/learn/kanji',
    'Grammar' => '/learn/grammar',
    _ => '/learn/reading?level=N5',
  };
  String _advice(BuildContext context, String name) => switch (name) {
    'Kosakata' =>
      'Ulangi kata sehari-hari dan pasangan kata yang sering tertukar. Perhatikan konteks kalimat.'
          .localized(context),
    'Kanji' =>
      'Latih arti dan cara baca kanji dasar, lalu gunakan contoh kata untuk mengingatnya.'
          .localized(context),
    'Grammar' =>
      'Tinjau partikel, perubahan bentuk kata kerja, dan pola kalimat dasar.'
          .localized(context),
    _ =>
      'Baca cerita pendek, lalu cari informasi waktu, tempat, alasan, dan tujuan dalam teks.'
          .localized(context),
  };
}
