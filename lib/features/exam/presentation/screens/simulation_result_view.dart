import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/simulation_text.dart';
import 'package:nihongo_master/features/ssw/presentation/ssw_assessment_screens.dart';

class SimulationResultView extends StatefulWidget {
  const SimulationResultView({
    required this.result,
    super.key,
    this.level = 'N5',
  });
  final String level;
  final Json result;

  @override
  State<SimulationResultView> createState() => _ResultState();
}

class _ResultState extends State<SimulationResultView> {
  bool wrongOnly = false;

  @override
  Widget build(BuildContext context) {
    final result = widget.result;
    final reviews = diagnosticList(result['reviews']);
    final jftExamNumber = widget.level == 'JFT'
        ? 1
        : int.tryParse(widget.level.replaceFirst('JFT-', '')) ?? 1;
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        DiagnosticHero(
          title: 'Simulasi selesai.',
          subtitle:
              '${widget.level.startsWith('SSW-')
                  ? sswExamTitle(widget.level)
                  : widget.level == 'JFT' || widget.level.startsWith('JFT-')
                  ? 'JFT Exam $jftExamNumber · Kosakata, Percakapan & Reading'
                  : widget.level.startsWith('NAT')
                  ? 'NAT 5Q Exam ${widget.level == 'NAT' ? 1 : widget.level.substring(4)} · Kosakata, Grammar & Reading'
                  : 'Paket 01 · ${widget.level} · Kosakata & Grammar${widget.level != 'N5' ? ' & Reading' : ''}'}\n${result['correct']}/${result['total']} benar · ${result['score']}%',
          child: Text(
            '${result['answered']} dijawab · ${(result['total'] as int) - (result['answered'] as int)} kosong\n'
            'Durasi ${(result['durationSeconds'] as int) ~/ 60} menit ${(result['durationSeconds'] as int) % 60} detik',
            style: const TextStyle(color: Colors.white, height: 1.8),
          ),
        ),
        const SizedBox(height: 20),
        Text(
          result['scoringNote'] as String,
          style: const TextStyle(fontSize: 12, height: 1.6),
        ),
        const SizedBox(height: 24),
        const ExamHeading('Hasil per jenis soal'),
        ExamPanel(
          child: Column(
            children: [
              for (final group in diagnosticList(result['sections']))
                Padding(
                  padding: const EdgeInsets.only(bottom: 18),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        group['name'] as String,
                        style: const TextStyle(fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        '${group['correct']}/${group['total']} benar · ${group['score']}%',
                      ),
                      const SizedBox(height: 8),
                      LinearProgressIndicator(
                        value: (group['score'] as int) / 100,
                        color: examInk,
                        minHeight: 6,
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        const ExamHeading('Pembahasan'),
        Align(
          alignment: Alignment.centerLeft,
          child: FilterChip(
            label: const Text('Salah & belum dijawab'),
            selected: wrongOnly,
            onSelected: (value) => setState(() => wrongOnly = value),
          ),
        ),
        for (final review in reviews.where(
          (r) => !wrongOnly || r['correct'] == false,
        ))
          Padding(
            padding: const EdgeInsets.only(top: 10),
            child: ExamPanel(
              child: ExpansionTile(
                tilePadding: EdgeInsets.zero,
                childrenPadding: const EdgeInsets.only(top: 12),
                leading: Icon(
                  review['correct'] == true
                      ? Icons.check_circle
                      : Icons.cancel_outlined,
                  color: review['correct'] == true
                      ? examGreen
                      : Theme.of(context).colorScheme.error,
                ),
                title: Text('Soal ${(review['question'] as Json)['number']}'),
                subtitle: Text((review['question'] as Json)['group'] as String),
                children: [_review(review)],
              ),
            ),
          ),
        const SizedBox(height: 24),
        OutlinedButton.icon(
          onPressed: () => context.pushReplacement(
            '/exams/${widget.level.toLowerCase()}/simulation',
          ),
          icon: const Icon(Icons.refresh_rounded),
          label: const Text('Ulangi paket 01'),
        ),
      ],
    );
  }

  Widget _review(Json review) {
    final question = review['question'] as Json;
    final options = diagnosticList(question['options']);
    String answer(String id) =>
        options
            .where((o) => o['id'] == id)
            .map((o) => o['text'] as String)
            .firstOrNull ??
        'Tidak dijawab';
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if ((question['passage'] as String? ?? '').isNotEmpty) ...[
          Text(
            question['passage'] as String,
            style: const TextStyle(fontSize: 17, height: 1.8),
          ),
          const SizedBox(height: 16),
        ],
        SimulationText(
          question['prompt'] as String,
          style: const TextStyle(
            fontSize: 20,
            height: 1.8,
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 16),
        Text('Jawabanmu: ${answer(review['selectedOptionId'] as String)}'),
        const SizedBox(height: 8),
        Text(
          'Jawaban benar: ${answer(review['correctOptionId'] as String)}',
          style: const TextStyle(color: examGreen, fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 12),
        Text(
          review['explanation'] as String,
          style: const TextStyle(height: 1.7),
        ),
      ],
    );
  }
}
