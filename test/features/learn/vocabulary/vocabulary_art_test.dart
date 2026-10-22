import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/widgets/vocabulary_art.dart';

void main() {
  testWidgets('failed vocabulary image keeps the cropped panel visible', (
    tester,
  ) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: Scaffold(
            body: Center(
              child: SizedBox(
                width: 88,
                height: 88,
                child: VocabularyArt(
                  asset: 'assets/images/vocabulary/n5-theme-01.webp',
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
}
