import 'dart:async';
import 'dart:typed_data';

import 'package:audioplayers/audioplayers.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

bool _privateMedia(String url) =>
    RegExp(r'^/anki/decks/user-[a-f0-9-]+/media/[^/]+$').hasMatch(url);
final _ankiMediaBytesProvider = FutureProvider.autoDispose
    .family<({Uint8List bytes, String mime}), String>((ref, path) async {
      if (!_privateMedia(path)) {
        throw const FormatException('Media deck tidak valid.');
      }
      final owner = ref.watch(
        authControllerProvider.select((state) => state.user?.id),
      );
      if (owner == null) throw StateError('Masuk untuk membuka media deck.');
      final cancel = CancelToken();
      ref.onDispose(cancel.cancel);
      final result = await ref
          .watch(dioProvider)
          .get<List<int>>(
            path,
            options: Options(
              responseType: ResponseType.bytes,
              extra: {'offlineAccountId': owner},
            ),
            cancelToken: cancel,
            onReceiveProgress: (received, total) {
              if (received > 10 * 1024 * 1024 || total > 10 * 1024 * 1024) {
                cancel.cancel('Media terlalu besar.');
              }
            },
          );
      if (ref.read(authControllerProvider).user?.id != owner) {
        throw StateError('Akun sudah berubah.');
      }
      return (
        bytes: Uint8List.fromList(result.data!),
        mime: result.headers.value('content-type') ?? 'audio/mpeg',
      );
    });

class AnkiDeckImage extends ConsumerWidget {
  const AnkiDeckImage(this.url, {super.key});
  final String url;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final Widget child;
    if (_privateMedia(url)) {
      child = ref
          .watch(_ankiMediaBytesProvider(url))
          .when(
            data: (media) => Image.memory(
              media.bytes,
              height: 200,
              fit: BoxFit.contain,
              cacheWidth: 700,
              errorBuilder: (_, __, ___) => const _MediaUnavailable(),
            ),
            error: (_, __) => const _MediaUnavailable(),
            loading: () => const SizedBox(
              height: 100,
              child: Center(child: CircularProgressIndicator()),
            ),
          );
    } else if (Uri.tryParse(url)?.scheme == 'https' ||
        Uri.tryParse(url)?.scheme == 'http') {
      child = Image.network(
        url,
        height: 200,
        fit: BoxFit.contain,
        cacheWidth: 700,
        errorBuilder: (_, __, ___) => const _MediaUnavailable(),
      );
    } else {
      child = const _MediaUnavailable();
    }
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: ClipRRect(borderRadius: BorderRadius.circular(14), child: child),
    );
  }
}

class _MediaUnavailable extends StatelessWidget {
  const _MediaUnavailable();
  @override
  Widget build(BuildContext context) => SizedBox(
    height: 100,
    child: Center(child: Text('Media belum dapat dimuat.'.localized(context))),
  );
}

class AnkiDeckAudio extends ConsumerStatefulWidget {
  const AnkiDeckAudio(this.url, {super.key});
  final String url;
  @override
  ConsumerState<AnkiDeckAudio> createState() => _AudioState();
}

class _AudioState extends ConsumerState<AnkiDeckAudio> {
  AudioPlayer? _player;
  StreamSubscription<PlayerState>? _subscription;
  bool _busy = false;
  bool _playing = false;
  @override
  void didUpdateWidget(covariant AnkiDeckAudio oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.url != widget.url) unawaited(_player?.stop());
  }

  @override
  void dispose() {
    unawaited(_subscription?.cancel());
    unawaited(_player?.dispose());
    super.dispose();
  }

  Future<void> _play() async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      final player = _player ??= AudioPlayer();
      _subscription ??= player.onPlayerStateChanged.listen((state) {
        if (mounted) setState(() => _playing = state == PlayerState.playing);
      });
      if (_playing) {
        await player.stop();
        return;
      }
      final url = widget.url;
      final owner = ref.read(authControllerProvider).user?.id;
      if (_privateMedia(url)) {
        final subscription = ref.listenManual(
          _ankiMediaBytesProvider(url),
          (_, __) {},
        );
        late final ({Uint8List bytes, String mime}) media;
        try {
          media = await ref.read(_ankiMediaBytesProvider(url).future);
        } finally {
          subscription.close();
        }
        if (!mounted ||
            widget.url != url ||
            ref.read(authControllerProvider).user?.id != owner) {
          return;
        }
        await player.play(BytesSource(media.bytes, mimeType: media.mime));
      } else if (const ['http', 'https'].contains(Uri.tryParse(url)?.scheme)) {
        await player.play(UrlSource(url));
      } else {
        throw const FormatException('Audio tidak valid.');
      }
    } on Object {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Audio belum dapat diputar. Periksa koneksi lalu coba lagi.'
                  .localized(context),
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(authControllerProvider.select((state) => state.user?.id), (
      previous,
      next,
    ) {
      if (previous != next) unawaited(_player?.stop());
    });
    return OutlinedButton.icon(
      onPressed: _busy ? null : _play,
      icon: Icon(_playing ? Icons.stop_rounded : Icons.volume_up_outlined),
      label: Text(
        (_playing
                ? 'Hentikan audio'
                : _busy
                ? 'Memuat audio…'
                : 'Putar audio')
            .localized(context),
      ),
    );
  }
}
