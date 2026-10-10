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
    final auth = ref.watch(authControllerProvider);
    if (!auth.isAuthenticated) return const SizedBox.shrink();
    final overview = ref.watch(diagnosticOverviewProvider);
    final premium = auth.user?.isPremium ?? false;
    final colors = Theme.of(context).colorScheme;
    return RefreshIndicator(
      onRefresh: () => ref.refresh(diagnosticOverviewProvider.future),
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          const Text(
            'Latihan Ujian',
            style: TextStyle(fontSize: 28, fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 6),
          Text(
            'Pilih level dan mulai latihan sesuai format JLPT.',
            style: TextStyle(color: colors.onSurfaceVariant),
          ),
          const SizedBox(height: 28),
          const _SectionTitle('Simulasi JLPT'),
          const SizedBox(height: 4),
          _ExamRow(
            badge: 'N5',
            title: 'Simulasi N5',
            metadata: '38 soal · 40 menit',
            description: 'Kosakata, kanji & grammar',
            note: 'Tanpa reading & listening',
            onTap: () => context.push('/exams/n5/simulation'),
          ),
          const Divider(height: 1),
          _ExamRow(
            badge: 'N4',
            title: 'Simulasi N4',
            metadata: '57 soal · 80 menit',
            description: 'Kosakata, kanji, grammar & reading',
            note: 'Tanpa listening',
            locked: !premium,
            onTap: () =>
                context.push(premium ? '/exams/n4/simulation' : '/premium'),
          ),
          const SizedBox(height: 20),
          const _SectionTitle('Cek kemampuan awal'),
          const SizedBox(height: 4),
          _ExamRow(
            icon: Icons.insights_outlined,
            title: 'Diagnostic N5',
            description: 'Temukan bagian yang perlu kamu latih.',
            onTap: () => context.push('/exams/n5'),
          ),
          const SizedBox(height: 20),
          const _SectionTitle('Riwayat diagnostic'),
          const SizedBox(height: 12),
          overview.when(
            loading: () => const Padding(
              padding: EdgeInsets.all(20),
              child: Center(child: CircularProgressIndicator()),
            ),
            error: (e, _) => ExamMessage(
              message: diagnosticError(e),
              retry: () => ref.invalidate(diagnosticOverviewProvider),
            ),
            data: (data) {
              final history = diagnosticList(data['history']);
              final activeAttemptId = data['activeAttemptId'] as String;
              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (activeAttemptId.isNotEmpty) ...[
                    Material(
                      color: examInk.withValues(alpha: .06),
                      borderRadius: BorderRadius.circular(12),
                      clipBehavior: Clip.antiAlias,
                      child: ListTile(
                        leading: const Icon(
                          Icons.play_circle_outline,
                          color: examInk,
                        ),
                        title: const Text('Lanjutkan diagnostic N5'),
                        subtitle: const Text('Jawabanmu sudah tersimpan.'),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () =>
                            context.push('/exams/n5/attempt/$activeAttemptId'),
                      ),
                    ),
                    const SizedBox(height: 12),
                  ],
                  if (history.isEmpty)
                    Text(
                      'Belum ada hasil. Riwayat akan muncul setelah kamu menyelesaikan diagnostic.',
                      style: TextStyle(
                        color: colors.onSurfaceVariant,
                        height: 1.5,
                      ),
                    ),
                  for (var i = 0; i < history.length; i++) ...[
                    if (i > 0) const Divider(height: 1),
                    ListTile(
                      contentPadding: EdgeInsets.zero,
                      title: const Text(
                        'Diagnostic N5',
                        style: TextStyle(fontWeight: FontWeight.w600),
                      ),
                      subtitle: Text(
                        '${history[i]['correct']}/${history[i]['total']} benar · ${_date(history[i]['completedAt'] as String)}',
                      ),
                      trailing: Text(
                        '${history[i]['score']}%',
                        style: const TextStyle(
                          color: examGreen,
                          fontSize: 20,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      onTap: () => context
                          .push('/exams/n5/attempt/${history[i]['attemptId']}'),
                    ),
                  ],
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

class _SectionTitle extends StatelessWidget {
  const _SectionTitle(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Text(
        text,
        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
      );
}

class _ExamRow extends StatelessWidget {
  const _ExamRow({
    required this.title,
    required this.description,
    required this.onTap,
    this.badge,
    this.icon,
    this.metadata,
    this.note,
    this.locked = false,
  });

  final String? badge;
  final IconData? icon;
  final String title;
  final String? metadata;
  final String description;
  final String? note;
  final bool locked;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Semantics(
      button: true,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 4),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 48,
                  height: 48,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: examInk.withValues(alpha: .07),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: badge != null
                      ? Text(
                          badge!,
                          style: const TextStyle(
                            color: examInk,
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                          ),
                        )
                      : Icon(icon, color: examInk),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: const TextStyle(
                          fontSize: 17,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      if (metadata != null) ...[
                        const SizedBox(height: 4),
                        Text(metadata!, style: const TextStyle(fontSize: 13)),
                      ],
                      const SizedBox(height: 4),
                      Text(
                        description,
                        style: TextStyle(
                          fontSize: 13,
                          color: colors.onSurfaceVariant,
                        ),
                      ),
                      if (note != null)
                        Text(
                          note!,
                          style: TextStyle(
                            fontSize: 12,
                            color: colors.onSurfaceVariant,
                          ),
                        ),
                      if (locked) ...[
                        const SizedBox(height: 6),
                        Text(
                          'Lifetime Access',
                          style: TextStyle(
                            color: colors.primary,
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Icon(
                    locked ? Icons.lock_outline : Icons.chevron_right,
                    size: 20,
                    color: colors.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
