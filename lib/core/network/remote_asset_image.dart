import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

/// Resolve backend-relative image URLs against the configured API origin.
String resolveAssetImageUrl(String apiBaseUrl, String url, String asset) {
  final path = url.isNotEmpty ? url : (asset.isEmpty ? '' : '/$asset');
  if (path.isEmpty) return '';
  return Uri.parse(apiBaseUrl).resolve(path).toString();
}

class RemoteAssetImage extends ConsumerWidget {
  const RemoteAssetImage({
    required this.asset,
    required this.url,
    this.fit = BoxFit.fill,
    this.alignment = Alignment.center,
    super.key,
  });

  final String asset;
  final String url;
  final BoxFit fit;
  final Alignment alignment;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final imageUrl = resolveAssetImageUrl(
      ref.watch(apiBaseUrlProvider),
      url,
      asset,
    );
    Widget placeholder() => ColoredBox(
          color: Theme.of(context).colorScheme.surfaceContainerLow,
          child: const Center(child: Icon(Icons.image_outlined)),
        );
    if (imageUrl.isEmpty) return placeholder();
    return Image.network(
      imageUrl,
      fit: fit,
      alignment: alignment,
      cacheWidth: 2172,
      excludeFromSemantics: true,
      errorBuilder: (_, __, ___) => placeholder(),
      loadingBuilder: (_, child, progress) =>
          progress == null ? child : placeholder(),
    );
  }
}
