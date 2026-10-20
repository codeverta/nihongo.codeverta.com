import 'package:flutter/material.dart';
import 'package:nihongo_master/core/network/remote_asset_image.dart';

/// A generated 4 x 2 sheet supplies eight independent word illustrations.
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
              child: RemoteAssetImage(
                asset: asset,
                url: url,
                loadingBuilder: (context) =>
                    _fallbackSheet(context, icon: Icons.image_outlined),
                errorBuilder: (context) => _fallbackSheet(
                  context,
                  icon: Icons.image_not_supported_outlined,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _fallbackSheet(BuildContext context, {required IconData icon}) =>
      GridView.count(
        crossAxisCount: 4,
        physics: const NeverScrollableScrollPhysics(),
        padding: EdgeInsets.zero,
        children: [
          for (var i = 0; i < 8; i++)
            ColoredBox(
              color: Theme.of(context).colorScheme.surfaceContainerLow,
              child: Center(
                child: Icon(icon, color: Theme.of(context).colorScheme.outline),
              ),
            ),
        ],
      );
}
