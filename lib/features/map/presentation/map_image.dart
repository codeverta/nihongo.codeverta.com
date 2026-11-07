import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/map/presentation/map_image_store.dart';

/// A spot photo kept on disk after it has been shown once, so it stays
/// visible offline. On the web it falls back to the browser cache.
class MapSpotImage extends ConsumerStatefulWidget {
  const MapSpotImage({required this.url, this.fit = BoxFit.cover, super.key});
  final String url;
  final BoxFit fit;

  @override
  ConsumerState<MapSpotImage> createState() => _MapSpotImageState();
}

class _MapSpotImageState extends ConsumerState<MapSpotImage> {
  Future<Uint8List?>? _bytes;

  String get _absoluteUrl {
    final uri = Uri.parse(widget.url);
    if (uri.hasScheme) return widget.url;
    return Uri.parse(
      ref.read(apiBaseUrlProvider),
    ).resolve(widget.url).toString();
  }

  @override
  void initState() {
    super.initState();
    if (mapImagesUseDisk) _bytes = cachedMapImage(_absoluteUrl);
  }

  @override
  void didUpdateWidget(MapSpotImage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.url != widget.url && mapImagesUseDisk) {
      _bytes = cachedMapImage(_absoluteUrl);
    }
  }

  @override
  Widget build(BuildContext context) {
    final surface = Theme.of(context).colorScheme.surfaceContainerHighest;
    final placeholder = ColoredBox(
      color: surface,
      child: const Center(child: Icon(Icons.image_not_supported_outlined)),
    );
    if (!mapImagesUseDisk) {
      return Image.network(
        _absoluteUrl,
        fit: widget.fit,
        width: double.infinity,
        errorBuilder: (_, __, ___) => placeholder,
      );
    }
    return FutureBuilder<Uint8List?>(
      future: _bytes,
      builder: (context, snapshot) {
        final bytes = snapshot.data;
        if (bytes == null) {
          return snapshot.connectionState == ConnectionState.done
              ? placeholder
              : ColoredBox(color: surface);
        }
        return Image.memory(
          bytes,
          fit: widget.fit,
          width: double.infinity,
          gaplessPlayback: true,
          errorBuilder: (_, __, ___) => placeholder,
        );
      },
    );
  }
}
