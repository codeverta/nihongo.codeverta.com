// Run with: flutter test scripts/export_splash_test.dart
// Exports the actual Flutter splash artwork, without system bars or spinner.
import 'dart:io';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/branding/app_branding.dart';

void main() {
  testWidgets('export default splash artwork', (tester) async {
    tester.view.physicalSize = const Size(1080, 1920);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final key = GlobalKey();
    await tester.runAsync(() async {
      final bytes = await File('C:/Windows/Fonts/segoeui.ttf').readAsBytes();
      final loader = FontLoader('SplashExport')
        ..addFont(Future.value(ByteData.sublistView(bytes)));
      await loader.load();
    });
    await tester.pumpWidget(
      MaterialApp(
        theme: ThemeData(fontFamily: 'SplashExport'),
        home: RepaintBoundary(
          key: key,
          child: const Material(
            color: Color(0xFFFAFAFA),
            child: AppSplashArtwork(branding: AppBranding()),
          ),
        ),
      ),
    );
    await tester.runAsync(
      () => precacheImage(
        const AssetImage('assets/images/app-icon-small.webp'),
        key.currentContext!,
      ),
    );
    await tester.pumpAndSettle();
    final boundary =
        key.currentContext!.findRenderObject()! as RenderRepaintBoundary;
    await tester.runAsync(() async {
      final image = await boundary.toImage(pixelRatio: 3);
      final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
      await File('output/play-store/splash-default.png')
          .writeAsBytes(bytes!.buffer.asUint8List());
      image.dispose();
    });
  });
}
