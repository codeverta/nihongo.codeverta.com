import 'dart:async';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/offline/learning_asset.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_art.dart';

void main() {
  testWidgets('failed vocabulary image keeps the cropped panel visible', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          learningAssetProvider.overrideWith((ref, url) async => null),
        ],
        child: const MaterialApp(
          home: Scaffold(
            body: Center(
              child: SizedBox(
                width: 88,
                height: 88,
                child: VocabularyArt(
                  asset: 'assets/images/vocabulary/missing.webp',
                ),
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(
      find.byIcon(Icons.image_not_supported_outlined).hitTestable(),
      findsOneWidget,
    );
  });

  testWidgets('N5 list artwork remains visible when remote images fail', (
    tester,
  ) async {
    const asset = 'assets/images/vocabulary/n5-theme-01.webp';
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          learningAssetProvider.overrideWith((ref, url) async => null),
        ],
        child: const MaterialApp(
          home: Scaffold(
            body: Center(
              child: SizedBox(
                width: 88,
                height: 88,
                child: VocabularyArt(asset: asset),
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.byType(Image), findsOneWidget);
    final image = tester.widget<Image>(find.byType(Image));
    expect(image.image, isA<AssetImage>());
    expect((image.image as AssetImage).assetName, 'backend/$asset');
    expect(tester.takeException(), isNull);
  });

  testWidgets('all 20 bundled N5 theme sheets decode successfully', (
    tester,
  ) async {
    await tester.runAsync(() async {
      for (var theme = 1; theme <= 20; theme++) {
        final number = theme.toString().padLeft(2, '0');
        final data = await rootBundle.load(
          'backend/assets/images/vocabulary/n5-theme-$number.webp',
        );
        final codec = await ui.instantiateImageCodec(
          data.buffer.asUint8List(data.offsetInBytes, data.lengthInBytes),
        );
        final frame = await codec.getNextFrame();
        expect(frame.image.width, greaterThan(0), reason: 'Theme $number');
        expect(frame.image.height, greaterThan(0), reason: 'Theme $number');
        frame.image.dispose();
        codec.dispose();
      }
    });
  });

  testWidgets('bundled N5 art displays while the remote sheet is loading', (
    tester,
  ) async {
    final download = Completer<Uint8List?>();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          learningAssetProvider.overrideWith((ref, url) => download.future),
        ],
        child: const MaterialApp(
          home: SizedBox(
            width: 88,
            height: 88,
            child: VocabularyArt(
              asset: 'assets/images/vocabulary/n5-theme-20.webp',
              panel: 7,
            ),
          ),
        ),
      ),
    );
    await tester.pump();
    expect(tester.widget<Image>(find.byType(Image)).image, isA<AssetImage>());

    final data = await rootBundle.load(
      'backend/assets/images/vocabulary/n5-theme-20.webp',
    );
    download.complete(data.buffer.asUint8List());
    await tester.pumpAndSettle();
    expect(tester.widget<Image>(find.byType(Image)).image, isA<MemoryImage>());
    expect(tester.takeException(), isNull);
  });
}
