import 'package:flutter/material.dart';

/// A generated 4×2 sheet supplies eight independent word illustrations.
class VocabularyArt extends StatelessWidget {
  const VocabularyArt({
    required this.asset,
    this.url = '',
    this.panel = 0,
    this.description = 'Ilustrasi kosakata',
    this.aspectRatio = 1,
    super.key,
  });
  final String asset;
  final String url;
  final int panel;
  final String description;
  final double aspectRatio;
  @override
  Widget build(BuildContext context) {
    final cell = panel.clamp(0, 7);
    Widget fallback() => Image.asset(
          asset,
          fit: BoxFit.fill,
          cacheWidth: 2172,
          excludeFromSemantics: true,
          errorBuilder: (_, __, ___) => ColoredBox(
            color: Theme.of(context).colorScheme.surfaceContainerLow,
            child: const Center(child: Icon(Icons.image_outlined)),
          ),
        );
    return Semantics(
      image: true,
      label: description,
      child: AspectRatio(
        aspectRatio: aspectRatio,
        child: LayoutBuilder(
          builder: (context, c) => ClipRect(
            child: OverflowBox(
              alignment: Alignment(
                (cell % 4) * 2 / 3 - 1,
                ((cell ~/ 4) * 2 - 1) / (2 - 1 / aspectRatio),
              ),
              minWidth: c.maxWidth * 4,
              maxWidth: c.maxWidth * 4,
              minHeight: c.maxWidth * 2,
              maxHeight: c.maxWidth * 2,
              child: url.isEmpty
                  ? fallback()
                  : Image.network(
                      url,
                      fit: BoxFit.fill,
                      cacheWidth: 2172,
                      excludeFromSemantics: true,
                      errorBuilder: (_, __, ___) => fallback(),
                      loadingBuilder: (_, child, progress) =>
                          progress == null ? child : fallback(),
                    ),
            ),
          ),
        ),
      ),
    );
  }
}
