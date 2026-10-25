import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

const examInk = Color(0xFF22235B);
const examGreen = Color(0xFF167C67);

class DiagnosticHero extends StatelessWidget {
  const DiagnosticHero({
    required this.title,
    required this.subtitle,
    this.child,
    super.key,
  });
  final String title;
  final String subtitle;
  final Widget? child;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(28),
          gradient: const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [examInk, Color(0xFF49418B)],
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: .15),
                    borderRadius: BorderRadius.circular(99),
                  ),
                  child: const Text(
                    'JLPT  /  N5',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 2,
                    ),
                  ),
                ),
                const Spacer(),
                const Icon(
                  Icons.auto_awesome_rounded,
                  color: Color(0xFFFFD58A),
                  size: 32,
                ),
              ],
            ),
            const SizedBox(height: 24),
            Text(
              title,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 29,
                height: 1.2,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              subtitle,
              style: TextStyle(
                color: Colors.white.withValues(alpha: .85),
                height: 1.6,
                fontSize: 15,
              ),
            ),
            if (child != null) ...[const SizedBox(height: 22), child!],
          ],
        ),
      );
}

class ExamPanel extends StatelessWidget {
  const ExamPanel({required this.child, super.key});
  final Widget child;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(22),
          border:
              Border.all(color: Theme.of(context).colorScheme.outlineVariant),
        ),
        child: child,
      );
}

class ExamHeading extends StatelessWidget {
  const ExamHeading(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: Text(
          text,
          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w800),
        ),
      );
}

class ExamMessage extends StatelessWidget {
  const ExamMessage({required this.message, required this.retry, super.key});
  final String message;
  final VoidCallback retry;
  @override
  Widget build(BuildContext context) => ExamPanel(
        child: Column(
          children: [
            const Icon(Icons.cloud_off_rounded, size: 36),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 12),
            TextButton(onPressed: retry, child: Text('Coba lagi'.localized(context))),
            if (message.contains('Sesi berakhir'))
              FilledButton(
                onPressed: () async {
                  final success = await context.push<bool>('/login');
                  if ((success ?? false) && context.mounted) retry();
                },
                child: Text('Masuk kembali'.localized(context)),
              ),
          ],
        ),
      );
}
