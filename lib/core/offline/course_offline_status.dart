import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Shows the offline note and a manual "save for offline" button. Lessons are
/// no longer prefetched automatically: each lesson is fetched (and cached) only
/// when opened, so entering or refreshing a course costs one list request.
class CourseOfflineStatus extends ConsumerStatefulWidget {
  const CourseOfflineStatus({
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
  ConsumerState<CourseOfflineStatus> createState() =>
      _CourseOfflineStatusState();
}

class _CourseOfflineStatusState extends ConsumerState<CourseOfflineStatus> {
  bool _saving = false;
  bool _pinned = false;

  @override
  void didUpdateWidget(covariant CourseOfflineStatus oldWidget) {
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
    ref.watch(authControllerProvider.select((s) => s.user?.id));
    if (offline) {
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
    return Align(
      alignment: Alignment.centerRight,
      child: TextButton.icon(
        onPressed: _saving || _pinned ? null : _saveCourse,
        icon: Icon(
          _pinned ? Icons.download_done_rounded : Icons.download_outlined,
          size: 18,
        ),
        label: Text(
          _saving
              ? 'Menyimpan…'.localized(context)
              : _pinned
              ? 'Tersimpan'.localized(context)
              : 'Simpan teks'.localized(context),
        ),
      ),
    );
  }
}
