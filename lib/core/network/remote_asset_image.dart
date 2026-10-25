import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/offline/learning_asset.dart';

/// Resolve backend-relative image URLs against the configured API origin.
String resolveAssetImageUrl(String apiBaseUrl, String url, String asset) {
  final path = url.isNotEmpty ? url : (asset.isEmpty ? '' : '/$asset');
  if (path.isEmpty) return '';
  return Uri.parse(apiBaseUrl).resolve(path).toString();
}

/// Build an API-origin URL for retrying assets that failed on a configured CDN.
String resolveAssetImageFallbackUrl(String apiBaseUrl, String asset) =>
    resolveAssetImageUrl(apiBaseUrl, '', asset);

class RemoteAssetImage extends ConsumerWidget {
  const RemoteAssetImage({
    required this.asset,
    required this.url,
    this.errorBuilder,
    this.loadingBuilder,
    this.fit = BoxFit.fill,
    this.alignment = Alignment.center,
    this.persistOffline = false,
    super.key,
  });

  final String asset;
  final String url;
  final WidgetBuilder? errorBuilder;
  final WidgetBuilder? loadingBuilder;
  final BoxFit fit;
  final Alignment alignment;
  final bool persistOffline;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final apiBaseUrl = ref.watch(apiBaseUrlProvider);
    final imageUrl = resolveAssetImageUrl(apiBaseUrl, url, asset);
    final fallbackUrl = resolveAssetImageFallbackUrl(apiBaseUrl, asset);
    Widget placeholder() =>
        errorBuilder?.call(context) ??
        ColoredBox(
          color: Theme.of(context).colorScheme.surfaceContainerLow,
          child: const Center(child: Icon(Icons.image_outlined)),
        );
    if (imageUrl.isEmpty) return placeholder();
    if (persistOffline) {
      final saved = ref.watch(learningAssetProvider(imageUrl));
      return saved.when(
        data: (bytes) {
          if (bytes == null) {
            if (fallbackUrl != imageUrl && fallbackUrl.isNotEmpty) {
              final fallback = ref
                  .watch(learningAssetProvider(fallbackUrl))
                  .valueOrNull;
              if (fallback != null) {
                return Image.memory(
                  fallback,
                  fit: fit,
                  alignment: alignment,
                  excludeFromSemantics: true,
                );
              }
            }
            return placeholder();
          }
          return Image.memory(
            bytes,
            fit: fit,
            alignment: alignment,
            excludeFromSemantics: true,
            errorBuilder: (_, __, ___) => placeholder(),
          );
        },
        error: (_, __) => placeholder(),
        loading: () => loadingBuilder?.call(context) ?? placeholder(),
      );
    }
    return Image.network(
      imageUrl,
      fit: fit,
      alignment: alignment,
      cacheWidth: 2172,
      excludeFromSemantics: true,
      errorBuilder: (_, __, ___) {
        if (fallbackUrl.isNotEmpty && fallbackUrl != imageUrl) {
          return Image.network(
            fallbackUrl,
            fit: fit,
            alignment: alignment,
            cacheWidth: 2172,
            excludeFromSemantics: true,
            errorBuilder: (_, __, ___) => placeholder(),
            loadingBuilder: (_, child, progress) => progress == null
                ? child
                : loadingBuilder?.call(context) ?? placeholder(),
          );
        }
        return placeholder();
      },
      loadingBuilder: (_, child, progress) => progress == null
          ? child
          : loadingBuilder?.call(context) ?? placeholder(),
    );
  }
}
