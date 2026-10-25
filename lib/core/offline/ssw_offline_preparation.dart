import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

class SswOfflinePreparation extends ConsumerStatefulWidget {
  const SswOfflinePreparation({required this.sectorId, super.key});
  final String sectorId;
  @override
  ConsumerState<SswOfflinePreparation> createState() => _PreparationState();
}

class _PreparationState extends ConsumerState<SswOfflinePreparation> {
  String? _prepared;
  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authControllerProvider).user;
    final offline = ref.watch(offlineProvider);
    final progress = ref.watch(learningDownloadProgressProvider).valueOrNull;
    final signature = '${user?.id}:${widget.sectorId}';
    if (offline) _prepared = null;
    if ((user?.isPremium ?? false) && !offline && _prepared != signature) {
      _prepared = signature;
      final sectorId = widget.sectorId;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        unawaited(() async {
          try {
            final downloads = await ref.read(learningDownloadsProvider.future);
            if (mounted) await downloads.preloadSswSector(sectorId: sectorId);
          } on Object {
            // The curriculum itself is bundled and remains available.
          }
        }());
      });
    }
    final thisSector = progress?.courseKey == 'ssw:${widget.sectorId}';
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: Text(
        thisSector && progress!.isPreparing
            ? 'Materi tersedia offline · menyiapkan latihan…'
            : thisSector &&
                  progress!.total > 0 &&
                  progress.failed == 0 &&
                  progress.remaining == 0
            ? 'Materi dan latihan siap dipakai offline.'
            : 'Materi tersedia di perangkat. Simulasi ujian memerlukan internet.',
        style: Theme.of(context).textTheme.bodySmall,
      ),
    );
  }
}
