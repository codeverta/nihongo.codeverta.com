import 'package:flutter/material.dart';

/// Each generated sheet contains three square scenes. The viewport shows one.
class ReadingIllustration extends StatelessWidget {
  const ReadingIllustration({
    required this.asset,
    required this.url,
    required this.panel,
    required this.description,
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
    Widget unavailable() => ColoredBox(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          child: const Center(child: Icon(Icons.image_not_supported_outlined)),
        );
    Widget local() => Image.asset(
          asset,
          fit: BoxFit.fill,
          cacheWidth: 2172,
          excludeFromSemantics: true,
          errorBuilder: (context, error, stack) => unavailable(),
        );
    return Semantics(
      image: true,
      label: description,
      child: AspectRatio(
        aspectRatio: aspectRatio,
        child: LayoutBuilder(
          builder: (context, constraints) => ClipRect(
            child: OverflowBox(
              alignment: Alignment(panel.clamp(0, 2).toDouble() - 1, 0),
              minWidth: constraints.maxWidth * 3,
              maxWidth: constraints.maxWidth * 3,
              minHeight: constraints.maxWidth,
              maxHeight: constraints.maxWidth,
              child: url.isEmpty
                  ? local()
                  : Image.network(
                      url,
                      fit: BoxFit.fill,
                      cacheWidth: 2172,
                      excludeFromSemantics: true,
                      errorBuilder: (context, error, stack) => local(),
                      loadingBuilder: (context, child, progress) =>
                          progress == null ? child : local(),
                    ),
            ),
          ),
        ),
      ),
    );
  }
}
