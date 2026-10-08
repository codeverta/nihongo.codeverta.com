import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/presentation/widgets/diagnostic_widgets.dart';

class ExamHubScreen extends ConsumerWidget {
  const ExamHubScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (!ref.watch(authControllerProvider).isAuthenticated) {
      return const SizedBox.shrink();
    }
    final overview = ref.watch(diagnosticOverviewProvider);
    return RefreshIndicator(
      onRefresh: () => ref.refresh(diagnosticOverviewProvider.future),
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          const Text(
            'Latihan Ujian',
            style: TextStyle(fontSize: 30, fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 6),
          const Text('Kenali kemampuanmu. Tentukan langkah berikutnya.'),
          const SizedBox(height: 24),
          DiagnosticHero(
            title: 'JLPT N5\nSimulation',
            subtitle:
                'Paket 01 · 38 soal · 40 menit\nKosakata, kanji, dan grammar. Dua bagian dengan timer dan jawaban yang dikunci setelah selesai.',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Dokkai & listening dilewati',
                  style: TextStyle(color: Colors.white),
                ),
                const SizedBox(height: 14),
                FilledButton.icon(
                  onPressed: () => context.push('/exams/n5/simulation'),
                  style: FilledButton.styleFrom(
                    backgroundColor: Colors.white,
                    foregroundColor: examInk,
                  ),
                  icon: const Icon(Icons.timer_outlined),
                  label: const Text('Buka simulasi N5'),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          DiagnosticHero(
            title: 'JLPT N4\nSimulation',
            subtitle:
                'Paket 01 · 57 soal · 80 menit\nKosakata dan kanji 25 menit, grammar dan reading 55 menit. Lengkap dengan pembahasan dan riwayat hasil.',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Listening dilewati',
                  style: TextStyle(color: Colors.white),
                ),
                const SizedBox(height: 14),
                FilledButton.icon(
                  onPressed: () => context.push('/exams/n4/simulation'),
                  style: FilledButton.styleFrom(
                    backgroundColor: Colors.white,
                    foregroundColor: examInk,
                  ),
                  icon: const Icon(Icons.timer_outlined),
                  label: const Text('Buka simulasi N4'),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          const ExamHeading('Cek awal dengan diagnostic'),
          DiagnosticHero(
            title: 'Seberapa kuat\ndasar bahasa Jepangmu?',
            subtitle:
                'Mulai dari diagnostic N5. Temukan bagian yang sudah kamu kuasai dan yang masih perlu dilatih.',
            child: FilledButton.icon(
              onPressed: () => context.push('/exams/n5'),
              style: FilledButton.styleFrom(
                backgroundColor: Colors.white,
                foregroundColor: examInk,
              ),
              icon: const Icon(Icons.arrow_forward_rounded),
              label: const Text('Buka diagnostic N5'),
            ),
          ),
          const SizedBox(height: 24),
          const ExamHeading('Satu tes, empat kemampuan'),
          const Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              Chip(
                avatar: Icon(Icons.translate_rounded, size: 18),
                label: Text('Kosakata'),
              ),
              Chip(
                avatar: Icon(Icons.brush_rounded, size: 18),
                label: Text('Kanji'),
              ),
              Chip(
                avatar: Icon(Icons.account_tree_outlined, size: 18),
                label: Text('Grammar'),
              ),
              Chip(
                avatar: Icon(Icons.auto_stories_outlined, size: 18),
                label: Text('Reading'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          const ExamHeading('Perjalanan diagnostic'),
          overview.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => ExamMessage(
              message: diagnosticError(e),
              retry: () => ref.invalidate(diagnosticOverviewProvider),
            ),
            data: (data) {
              final history = diagnosticList(data['history']);
              return Column(
                children: [
                  if ((data['activeAttemptId'] as String).isNotEmpty) ...[
                    ListTile(
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20),
                      ),
                      tileColor: examInk.withValues(alpha: .06),
                      leading: const Icon(
                        Icons.play_circle_outline,
                        color: examInk,
                      ),
                      title: const Text('Ada sesi yang belum selesai'),
                      subtitle: const Text('Lanjutkan dari jawaban tersimpan.'),
                      trailing: const Icon(Icons.chevron_right),
                      onTap: () => context.push(
                        '/exams/n5/attempt/${data['activeAttemptId']}',
                      ),
                    ),
                    const SizedBox(height: 12),
                  ],
                  if (history.isEmpty)
                    const ExamPanel(
                      child: Column(
                        children: [
                          Icon(
                            Icons.insights_rounded,
                            size: 38,
                            color: examGreen,
                          ),
                          SizedBox(height: 10),
                          Text(
                            'Langkah pertamamu dimulai di sini',
                            style: TextStyle(fontWeight: FontWeight.w800),
                          ),
                          SizedBox(height: 6),
                          Text(
                            'Hasil dan perkembanganmu akan tampil setelah menyelesaikan diagnostic.',
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),
                  for (final result in history)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 10),
                      child: ExamPanel(
                        child: ListTile(
                          contentPadding: EdgeInsets.zero,
                          leading: const CircleAvatar(
                            backgroundColor: examInk,
                            foregroundColor: Colors.white,
                            child: Text('N5'),
                          ),
                          title: const Text(
                            'JLPT N5 Diagnostic',
                            style: TextStyle(
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          subtitle: Text(
                            '${result['correct']}/${result['total']} benar • ${_date(result['completedAt'] as String)}',
                          ),
                          trailing: Text(
                            '${result['score']}%',
                            style: const TextStyle(
                              color: examGreen,
                              fontSize: 23,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          onTap: () => context.push(
                            '/exams/n5/attempt/${result['attemptId']}',
                          ),
                        ),
                      ),
                    ),
                ],
              );
            },
          ),
        ],
      ),
    );
  }

  String _date(String value) {
    final d = DateTime.parse(value).toLocal();
    return '${d.day}/${d.month}/${d.year}';
  }
}
