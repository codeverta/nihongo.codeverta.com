import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Keeps the whole app above the system navigation bar.
///
/// Recent Android versions draw apps edge-to-edge, so content laid out to the
/// bottom of the screen (bottom sheets, footers, map controls, buttons at the
/// end of a list) ends up under the 3-button or gesture navigation bar. This
/// pads every route once, at the top of the tree, with the bottom system
/// inset and paints that strip in the theme's background colour. Because the
/// padding is consumed here, widgets that already use their own `SafeArea`
/// do not get it twice.
///
/// The widget tree never changes shape: when the keyboard opens Flutter
/// reports the bottom padding as zero, and swapping widgets at that moment
/// would rebuild the whole app and drop the text field's focus.
class SystemNavigationInset extends StatelessWidget {
  const SystemNavigationInset({required this.child, super.key});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final background = theme.scaffoldBackgroundColor;
    final dark = theme.brightness == Brightness.dark;
    return AnnotatedRegion<SystemUiOverlayStyle>(
      // Navigation bar buttons stay readable on the strip in both themes.
      value: SystemUiOverlayStyle(
        systemNavigationBarColor: background,
        systemNavigationBarDividerColor: background,
        systemNavigationBarIconBrightness: dark
            ? Brightness.light
            : Brightness.dark,
        systemNavigationBarContrastEnforced: false,
      ),
      child: ColoredBox(
        color: background,
        child: SafeArea(top: false, left: false, right: false, child: child),
      ),
    );
  }
}
