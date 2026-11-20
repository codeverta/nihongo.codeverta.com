import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:nihongo_master/core/unlock/locked_content.dart';
import 'package:nihongo_master/core/unlock/unlock_sheet.dart';
import 'package:nihongo_master/features/jobs/data/jobs_repository.dart';

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
    if (widget.job.applyLocked) {
      await _applyThroughServer(user);
      return;
    }
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
        _message(
          'Tautan lamaran belum dapat dibuka. Coba lagi.'.localized(context),
        );
      }
    } on Object {
      if (mounted) {
        _message(
          'Akses premium belum dapat diverifikasi. Periksa koneksi lalu coba lagi.'
              .localized(context),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  /// While the jobs gate is on the link is not in the public listing: the
  /// server gives it to Lifetime accounts and to accounts that opened this
  /// job with a rewarded ad.
  Future<void> _applyThroughServer(AuthenticatedUser? user) async {
    if (user == null) {
      await context.push<bool>('/login');
      return;
    }
    final repository = ref.read(jobsRepositoryProvider);
    setState(() => _busy = true);
    try {
      Uri? uri;
      try {
        uri = await repository.applyLink(widget.job.id);
      } on LockedContentException catch (locked) {
        final target = locked.target;
        if (target == null) {
          _subscribe();
          return;
        }
        if (!mounted) return;
        setState(() => _busy = false);
        final opened = await showUnlockSheet(
          context,
          target,
          title: 'Buka tautan lamaran'.localized(context),
          returnTo: _destination,
        );
        if (!opened || !mounted) return;
        setState(() => _busy = true);
        uri = await repository.applyLink(widget.job.id);
      }
      if (!mounted || ref.read(authControllerProvider).user?.id != user.id) {
        return;
      }
      final launched =
          uri != null && await ref.read(jobApplicationLauncherProvider)(uri);
      if (!launched && mounted) {
        _message(
          'Tautan lamaran belum dapat dibuka. Coba lagi.'.localized(context),
        );
      }
    } on Object {
      if (mounted) {
        _message(
          'Tautan lamaran belum dapat dimuat. Periksa koneksi lalu coba lagi.'
              .localized(context),
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
            (widget.job.applyLocked ||
                jobExternalUri(widget.job.applyUrl) != null) &&
            !_busy
        ? _apply
        : null,
    icon: _busy
        ? const SizedBox.square(
            dimension: 18,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        : const Icon(Icons.workspace_premium_rounded),
    label: Text(
      _busy
          ? 'Memeriksa akses…'.localized(context)
          : 'Lihat & lamar di situs sumber'.localized(context),
    ),
  );
}
