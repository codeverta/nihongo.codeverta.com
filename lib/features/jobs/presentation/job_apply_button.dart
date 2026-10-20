import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:url_launcher/url_launcher.dart';

typedef JobApplicationAccess = Future<bool> Function();
final jobApplicationAccessProvider = Provider<JobApplicationAccess>((ref) {
  final repository = BillingRepository(ref.watch(dioProvider));
  // Revalidate against billing so stale local access cannot open a paid action.
  return () async => (await repository.status()).premium;
});
final jobApplicationLauncherProvider = Provider<Future<bool> Function(Uri)>(
  (ref) =>
      (uri) => launchUrl(uri, mode: LaunchMode.externalApplication),
);

class JobApplyButton extends ConsumerStatefulWidget {
  const JobApplyButton({required this.job, super.key});
  final JapanJob job;
  @override
  ConsumerState<JobApplyButton> createState() => _JobApplyButtonState();
}

class _JobApplyButtonState extends ConsumerState<JobApplyButton> {
  bool _busy = false;
  String get _destination => '/jobs/${Uri.encodeComponent(widget.job.id)}';
  void _subscribe() =>
      context.push('/premium?from=${Uri.encodeQueryComponent(_destination)}');
  Future<void> _apply() async {
    if (_busy) return;
    final user = ref.read(authControllerProvider).user;
    if (!(user?.isPremium ?? false) && !(user?.isLifetime ?? false)) {
      _subscribe();
      return;
    }
    setState(() => _busy = true);
    try {
      final allowed = await ref.read(jobApplicationAccessProvider)();
      if (!mounted) return;
      // A logout/account switch during verification must not launch the link.
      if (ref.read(authControllerProvider).user?.id != user?.id || !allowed) {
        _subscribe();
        return;
      }
      final uri = jobExternalUri(widget.job.applyUrl);
      if (!widget.job.active || uri == null) return;
      final opened = await ref.read(jobApplicationLauncherProvider)(uri);
      if (!opened && mounted) {
        _message('Tautan lamaran belum dapat dibuka. Coba lagi.');
      }
    } on Object {
      if (mounted) {
        _message(
          'Akses premium belum dapat diverifikasi. Periksa koneksi lalu coba lagi.',
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _message(String text) =>
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(text)));
  @override
  Widget build(BuildContext context) => FilledButton.icon(
    key: const Key('apply-job'),
    style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
    onPressed:
        widget.job.active &&
            jobExternalUri(widget.job.applyUrl) != null &&
            !_busy
        ? _apply
        : null,
    icon: _busy
        ? const SizedBox.square(
            dimension: 18,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        : const Icon(Icons.workspace_premium_rounded),
    label: Text(_busy ? 'Memeriksa akses…' : 'Lihat & lamar di situs sumber'),
  );
}
