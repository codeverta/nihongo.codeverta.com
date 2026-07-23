import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/learn/vocabulary/domain/vocabulary_models.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/controllers/vocabulary_controller.dart';
import 'package:nihongo_master/features/learn/vocabulary/presentation/screens/vocabulary_course_screen.dart';

void main() {
  testWidgets('course renders real progress and best score from provider', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          vocabularyLevelsProvider.overrideWith((ref) async => _levels),
        ],
        child: const MaterialApp(home: VocabularyCourseScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Fondasi API'), findsOneWidget);
    expect(find.byKey(const Key('level-progress-N5')), findsOneWidget);
    expect(find.text('64%'), findsOneWidget);
    expect(find.textContaining('Nilai terbaik 80%'), findsOneWidget);
  });
}

const _levels = [
  VocabularyLevel(
    level: 'N5',
    title: 'Fondasi API',
    description: 'Data level dari backend',
    requiredTier: 'free',
    progress: 64,
    lessons: [
      VocabularyLessonSummary(
        id: 'n5-daily',
        order: 1,
        title: 'Kata dasar API',
        description: 'Data lesson dari backend',
        wordCount: 10,
        bestScore: 80,
        completed: true,
      ),
    ],
  ),
];
