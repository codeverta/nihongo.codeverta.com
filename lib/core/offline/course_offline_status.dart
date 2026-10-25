import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/core/offline/offline_providers.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Starts preparation after the first frame; it never delays entering a course.
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
  String? _prepared;
  bool _saving = false;
  bool _pinned = false;

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
                  ? 'Teks course disimpan untuk offline.'
                  : '$saved dari ${widget.lessonIds.length} lesson tersimpan. Hubungkan internet atau kosongkan ruang untuk melengkapi.',
            ),
          ),
        );
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Materi belum selesai disimpan. Periksa koneksi dan ruang penyimpanan.',
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _prepare() {
    final user = ref.read(authControllerProvider).user?.id;
    final signature =
        '$user:${widget.type}:${widget.level}:${widget.activeLessonId}:${widget.lessonIds.join(',')}';
    if (_prepared == signature || ref.read(offlineProvider)) return;
    _pinned = false;
    _prepared = signature;
    final type = widget.type;
    final level = widget.level;
    final activeLessonId = widget.activeLessonId;
    final ids = List<String>.of(widget.lessonIds);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(() async {
        try {
          final downloads = await ref.read(learningDownloadsProvider.future);
          if (!mounted) return;
          await downloads.preloadCourse(
            type: type,
            level: level,
            lessonIds: ids,
            activeLessonId: activeLessonId,
          );
        } on Object {
          // Opening a course must remain usable if preparation fails.
        }
      }());
    });
  }

  @override
  Widget build(BuildContext context) {
    final offline = ref.watch(offlineProvider);
    ref.watch(authControllerProvider.select((s) => s.user?.id));
    if (offline) _prepared = null;
    final progress = ref.watch(learningDownloadProgressProvider).valueOrNull;
    _prepare();
    final thisCourse = progress?.courseKey == '${widget.type}:${widget.level}';
    final preparing = thisCourse && progress!.isPreparing;
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Icon(
            offline ? Icons.offline_bolt_outlined : Icons.download_done_rounded,
            size: 18,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              offline
                  ? 'Offline · buka materi yang sudah disiapkan.'
                  : preparing
                  ? 'Menyiapkan teks materi · ${progress.completed}/${progress.total}'
                  : thisCourse &&
                        (progress!.failed > 0 || progress.remaining > 0)
                  ? 'Sebagian materi tersimpan. Materi lainnya belum selesai disiapkan.'
                  : thisCourse && progress!.completed > 0
                  ? 'Teks materi siap dipakai offline.'
                  : 'Teks disiapkan otomatis untuk belajar berikutnya.',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
          if (preparing)
            const SizedBox.square(
              dimension: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          TextButton(
            onPressed: _saving || _pinned ? null : _saveCourse,
            child: Text(
              _saving
                  ? 'Menyimpan…'
                  : _pinned
                  ? 'Tersimpan'
                  : 'Simpan teks',
            ),
          ),
        ],
      ),
    );
  }
}
