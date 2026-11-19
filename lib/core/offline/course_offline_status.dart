import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// The offline note shown at the top of a course while the device is
/// offline. The "save for offline" action lives in the app bar
/// ([CourseOfflineButton]).
class CourseOfflineStatus extends ConsumerWidget {
  const CourseOfflineStatus({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (!ref.watch(offlineProvider)) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          const Icon(Icons.offline_bolt_outlined, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Offline · buka materi yang sudah disiapkan.'.localized(context),
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        ],
      ),
    );
  }
}

/// App-bar action that saves a whole course for offline use. Lessons are no
/// longer prefetched automatically: each lesson is fetched (and cached) only
/// when opened, so entering or refreshing a course costs one list request.
/// Hidden while offline or before the course has loaded.
class CourseOfflineButton extends ConsumerStatefulWidget {
  const CourseOfflineButton({
    required this.type,
    required this.level,
    required this.lessonIds,
    this.activeLessonId,
    super.key,
  });
  final String type;
  final String level;
  final List<String> lessonIds;
  final String? activeLessonId;

  @override
  ConsumerState<CourseOfflineButton> createState() =>
      _CourseOfflineButtonState();
}

class _CourseOfflineButtonState extends ConsumerState<CourseOfflineButton> {
  bool _saving = false;
  bool _pinned = false;

  @override
  void didUpdateWidget(covariant CourseOfflineButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.type != widget.type || oldWidget.level != widget.level) {
      _pinned = false;
    }
  }

  Future<void> _saveCourse() async {
    if (_saving) return;
    final userId = ref.read(authControllerProvider).user?.id;
    if (userId == null) return;
    final type = widget.type;
    final level = widget.level;
    final ids = List<String>.of(widget.lessonIds);
    final activeLessonId = widget.activeLessonId;
    setState(() => _saving = true);
    try {
      final downloads = await ref.read(learningDownloadsProvider.future);
      if (!mounted) return;
      final result = await downloads.preloadCourse(
        type: type,
        level: level,
        lessonIds: ids,
        activeLessonId: activeLessonId,
      );
      if (!mounted || ref.read(authControllerProvider).user?.id != userId) {
        return;
      }
      final cache = await ref.read(learningCacheProvider.future);
      var saved = 0;
      for (final id in ids) {
        final path = '/$type/lessons/${Uri.encodeComponent(id)}';
        if (await cache.setResponsePinned(
          accountId: userId,
          key: path,
          pinned: true,
        )) {
          saved++;
        }
        await cache.setResponsePinned(
          accountId: userId,
          key: '/offline/practice/$type/${Uri.encodeComponent(id)}',
          pinned: true,
        );
      }
      if (mounted && widget.type == type && widget.level == level) {
        setState(
          () => _pinned =
              saved == ids.length &&
              result.failed == 0 &&
              result.remaining == 0,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              _pinned
                  ? 'Teks course disimpan untuk offline.'.localized(context)
                  : '$saved ${'dari'.localized(context)} ${widget.lessonIds.length} ${'lesson tersimpan.'.localized(context)} ${'Hubungkan internet atau kosongkan ruang untuk melengkapi.'.localized(context)}',
            ),
          ),
        );
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Materi belum selesai disimpan. Periksa koneksi dan ruang penyimpanan.'
                  .localized(context),
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final offline = ref.watch(offlineProvider);
    final signedIn =
        ref.watch(authControllerProvider.select((s) => s.user?.id)) != null;
    if (offline || !signedIn || widget.lessonIds.isEmpty) {
      return const SizedBox.shrink();
    }
    final tooltip = _saving
        ? 'Menyimpan…'
        : _pinned
        ? 'Tersimpan untuk offline'
        : 'Simpan offline';
    return IconButton(
      key: const ValueKey('course-offline-save'),
      tooltip: tooltip.localized(context),
      onPressed: _saving || _pinned ? null : _saveCourse,
      icon: _saving
          ? const SizedBox.square(
              dimension: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : Icon(
              _pinned
                  ? Icons.download_done_rounded
                  : Icons.download_for_offline_outlined,
            ),
    );
  }
}
