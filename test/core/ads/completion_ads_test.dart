import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/completion_ads.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';

class _RecordingController extends InterstitialController {
  _RecordingController(super.ref);
  final prepared = <CompletionKind>[];
  final shown = <(CompletionKind, String?)>[];

  @override
  void prepareAfterCompletion(CompletionKind kind) => prepared.add(kind);

  @override
  Future<void> showAfterCompletion(CompletionKind kind, {String? key}) async =>
      shown.add((kind, key));
}

void main() {
  testWidgets(
    'a result screen preloads while shown and shows the ad when left',
    (tester) async {
      late _RecordingController controller;
      final onResult = ValueNotifier(true);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            interstitialControllerProvider.overrideWith((ref) {
              controller = _RecordingController(ref);
              ref.onDispose(controller.dispose);
              return controller;
            }),
          ],
          child: MaterialApp(
            home: ValueListenableBuilder<bool>(
              valueListenable: onResult,
              builder: (_, result, __) => result
                  ? const CompletionAdBoundary(
                      kind: CompletionKind.lessons,
                      completionKey: 'grammar:grammar-n5-01',
                      child: Text('Hasil kuis'),
                    )
                  : const Text('Daftar lesson'),
            ),
          ),
        ),
      );
      await tester.pump();
      expect(find.text('Hasil kuis'), findsOneWidget);
      expect(controller.prepared, [CompletionKind.lessons]);
      expect(
        controller.shown,
        isEmpty,
        reason: 'nothing while the score is on screen',
      );

      onResult.value = false;
      await tester.pump();
      await tester.pump();
      expect(controller.shown, [
        (CompletionKind.lessons, 'grammar:grammar-n5-01'),
      ]);
    },
  );
}
