import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/manga/presentation/manga_screens.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';

void main() {
  testWidgets(
    'Manga and Anki use distinct bundled illustrations and shared cards',
    (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: Column(children: [MangaEntryCard(), AnkiEntryCard()]),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      final cards = tester
          .widgetList<ExploreFeatureCard>(find.byType(ExploreFeatureCard))
          .toList();
      expect(cards.map((c) => c.asset), [
        'assets/images/explore/manga-reading.webp',
        'assets/images/explore/anki-flashcards.webp',
      ]);
      expect(cards.first.premium, isTrue);
      expect(find.text('Premium'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
}
