import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

class AnkiOfflinePreparation extends ConsumerStatefulWidget {
  const AnkiOfflinePreparation({
    required this.deckId,
    required this.level,
    super.key,
  });
  final String deckId;
  final String level;
  @override
  ConsumerState<AnkiOfflinePreparation> createState() => _PreparationState();
}

class _PreparationState extends ConsumerState<AnkiOfflinePreparation> {
  String? _prepared;
  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authControllerProvider.select((s) => s.user?.id));
    final offline = ref.watch(offlineProvider);
    final progress = ref.watch(learningDownloadProgressProvider).valueOrNull;
    final signature = '$user:${widget.deckId}';
    if (offline) _prepared = null;
    if (user != null && !offline && _prepared != signature) {
      _prepared = signature;
      final deckId = widget.deckId;
      final level = widget.level;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        unawaited(() async {
          try {
            final downloads = await ref.read(learningDownloadsProvider.future);
            if (mounted) {
              await downloads.preloadDeck(deckId: deckId, level: level);
            }
          } on Object {
            // Preparation should not delay reading a deck's explanation.
          }
        }());
      });
    }
    final thisDeck = progress?.courseKey == 'anki:${widget.deckId}';
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Text(
        offline
            ? 'Offline · sesi kartu yang sudah disiapkan tetap bisa dibuka.'
            : thisDeck && progress!.isPreparing
            ? 'Menyiapkan sesi kartu untuk offline…'
            : thisDeck &&
                  progress!.total > 0 &&
                  progress.failed == 0 &&
                  progress.remaining == 0
            ? 'Sesi kartu tersimpan untuk offline.'
            : 'Sesi kartu disiapkan saat terhubung.',
        style: Theme.of(context).textTheme.bodySmall,
      ),
    );
  }
}
