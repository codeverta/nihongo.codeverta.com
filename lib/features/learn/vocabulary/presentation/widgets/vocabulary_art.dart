import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
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

  static final _bundledTheme = RegExp(
    r'^assets/images/vocabulary/n5-theme-(0[1-9]|1[0-9]|20)\.webp$',
  );

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
      label: description == 'Ilustrasi kosakata'
          ? description.localized(context)
          : description,
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
                persistOffline: true,
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

  Widget _fallbackSheet(BuildContext context, {required IconData icon}) {
    if (_bundledTheme.hasMatch(asset)) {
      // Keep the same 4 x 2 sheet and crop while remote artwork loads or fails.
      // These public N5 illustrations are bundled, so no network is required.
      return Image.asset(
        'backend/$asset',
        fit: BoxFit.fill,
        excludeFromSemantics: true,
        errorBuilder: (_, __, ___) => _placeholderSheet(context, icon: icon),
      );
    }
    return _placeholderSheet(context, icon: icon);
  }

  Widget _placeholderSheet(BuildContext context, {required IconData icon}) =>
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
