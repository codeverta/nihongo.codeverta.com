import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/utils/save_file.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/learn/level_list/data/level_list_repository.dart';

/// Delivers a downloaded PDF to the user. Replaceable in tests.
typedef PdfSaver =
    Future<void> Function(BuildContext context, Uint8List bytes, String name);

final levelPdfSaverProvider = Provider<PdfSaver>(
  (ref) =>
      (context, bytes, name) => saveFileToDevice(
        context,
        bytes: bytes,
        name: name,
        mimeType: 'application/pdf',
        extension: 'pdf',
        typeLabel: 'PDF',
      ),
);

/// Whether the signed-in user has Lifetime Access. Falls back to false when
/// the account state is unavailable; the server enforces the real gate.
final levelPdfLifetimeProvider = Provider<bool>((ref) {
  try {
    return ref.watch(
      authControllerProvider.select((s) => s.user?.isLifetime ?? false),
    );
  } on StateError {
    return false;
  }
});

/// "Download PDF" for a level list. Lifetime users get the file; everyone
/// else sees a lock and is pointed to Lifetime Access.
class LevelPdfButton extends ConsumerStatefulWidget {
  const LevelPdfButton({
    required this.kind,
    required this.level,
    this.compact = false,
    super.key,
  });

  final LevelListKind kind;
  final String level;

  /// Icon-only variant for tight spots such as next to an exam's Start button.
  final bool compact;

  @override
  ConsumerState<LevelPdfButton> createState() => _LevelPdfButtonState();
}

class _LevelPdfButtonState extends ConsumerState<LevelPdfButton> {
  bool _busy = false;

  Future<void> _showLifetimePrompt() async {
    final upgrade = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.workspace_premium_rounded),
        title: Text('Download PDF khusus Lifetime Access'.localized(context)),
        content: Text(
          'Simpan daftar lengkap satu level sebagai PDF untuk dibaca kapan saja. Fitur ini termasuk dalam Lifetime Access.'
              .localized(context),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Nanti'.localized(context)),
          ),
          FilledButton(
            key: const ValueKey('level-pdf-upsell'),
            onPressed: () => Navigator.pop(context, true),
            child: Text('Lihat Lifetime'.localized(context)),
          ),
        ],
      ),
    );
    if ((upgrade ?? false) && mounted) unawaited(context.push('/premium'));
  }

  Future<void> _download() async {
    setState(() => _busy = true);
    try {
      final bytes = await ref
          .read(levelListRepositoryProvider)
          .pdf(widget.kind, widget.level);
      if (!mounted) return;
      await ref.read(levelPdfSaverProvider)(
        context,
        bytes,
        widget.kind.fileName(widget.level),
      );
    } on LevelPdfException catch (error) {
      if (!mounted) return;
      if (error.code == LevelPdfException.premiumRequired) {
        setState(() => _busy = false);
        await _showLifetimePrompt();
      } else {
        ScaffoldMessenger.maybeOf(context)?.showSnackBar(
          SnackBar(content: Text(error.message.localized(context))),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final lifetime = ref.watch(levelPdfLifetimeProvider);
    final offline = ref.watch(offlineProvider);
    final label = offline
        ? 'Butuh koneksi internet'
        : _busy
        ? 'Menyiapkan PDF…'
        : 'Download PDF';
    final onPressed = offline || _busy
        ? null
        : lifetime
        ? _download
        : _showLifetimePrompt;
    if (widget.compact) {
      return IconButton.outlined(
        key: const ValueKey('level-pdf-button'),
        tooltip: (offline ? label : 'Download PDF').localized(context),
        onPressed: onPressed,
        icon: _busy
            ? const SizedBox.square(
                dimension: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : Icon(
                lifetime
                    ? Icons.picture_as_pdf_outlined
                    : Icons.lock_outline_rounded,
              ),
      );
    }
    return OutlinedButton.icon(
      key: const ValueKey('level-pdf-button'),
      // The app theme makes outlined buttons full width (Size.fromHeight),
      // which cannot lay out inside the header Row and blanked the list.
      style: OutlinedButton.styleFrom(
        minimumSize: const Size(0, 44),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      ),
      onPressed: offline || _busy
          ? null
          : lifetime
          ? _download
          : _showLifetimePrompt,
      icon: _busy
          ? const SizedBox.square(
              dimension: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : Icon(
              lifetime
                  ? Icons.picture_as_pdf_outlined
                  : Icons.lock_outline_rounded,
            ),
      label: Text(label.localized(context)),
    );
  }
}
