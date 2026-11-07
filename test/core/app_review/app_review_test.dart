import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/app_review/app_review_policy.dart';
import 'package:nihongo_master/core/app_review/app_review_service.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:shared_preferences/shared_preferences.dart';

class FakeStoreReview implements StoreReviewClient {
  bool available = true;
  int requests = 0;
  String? openedWith;
  bool opened = false;
  @override
  Future<bool> isAvailable() async => available;
  @override
  Future<void> requestReview() async {
    requests++;
  }
  @override
  Future<void> openStoreListing({String? appStoreId}) async {
    opened = true;
    openedWith = appStoreId;
  }
}

void main() {
  group('AppReviewPolicy', () {
    const policy = AppReviewPolicy();
    final now = DateTime.utc(2026, 11, 5);

    test('needs 3 lessons and 3 active days', () {
      expect(
        policy.isEligible(completedLessons: 3, activeDays: 3, now: now),
        isTrue,
      );
      expect(
        policy.isEligible(completedLessons: 2, activeDays: 5, now: now),
        isFalse,
      );
      expect(
        policy.isEligible(completedLessons: 5, activeDays: 2, now: now),
        isFalse,
      );
    });

    test('waits 30 days after the last request', () {
      bool eligible(DateTime last) => policy.isEligible(
        completedLessons: 3,
        activeDays: 3,
        now: now,
        lastRequestedAt: last,
      );
      expect(eligible(now.subtract(const Duration(days: 29))), isFalse);
      expect(eligible(now.subtract(const Duration(days: 30))), isTrue);
    });
  });

  group('AppReviewService', () {
    late SharedPreferences preferences;
    late FakeStoreReview store;
    late DateTime now;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      preferences = await SharedPreferences.getInstance();
      store = FakeStoreReview();
      now = DateTime(2026, 11, 1, 9);
    });

    AppReviewService service({bool supported = true}) => AppReviewService(
      preferences,
      store,
      clock: () => now,
      supported: supported,
    );

    Future<void> useOnThreeDaysWithThreeLessons(AppReviewService s) async {
      for (var day = 0; day < 3; day++) {
        now = DateTime(2026, 11, 1 + day, 9);
        await s.recordLessonCompleted('grammar:lesson-$day');
      }
    }

    test('does not ask before the conditions are met', () async {
      final s = service();
      await s.recordLessonCompleted('grammar:a');
      await s.recordLessonCompleted('kanji:b');
      await s.recordLessonCompleted('reading:c');
      expect(s.activeDays, 1);
      expect(await s.maybeRequestReview(), isFalse);
      expect(store.requests, 0);
    });

    test('the same lesson and the same day are counted once', () async {
      final s = service();
      await s.recordLessonCompleted('grammar:a');
      await s.recordLessonCompleted('grammar:a');
      await s.recordActiveDay();
      expect(s.completedLessons, 1);
      expect(s.activeDays, 1);
    });

    test('asks once, then waits 30 days', () async {
      final s = service();
      await useOnThreeDaysWithThreeLessons(s);
      expect(await s.maybeRequestReview(), isTrue);
      expect(store.requests, 1);

      now = now.add(const Duration(days: 10));
      expect(await s.maybeRequestReview(), isFalse);

      now = now.add(const Duration(days: 20));
      expect(await s.maybeRequestReview(), isTrue);
      expect(store.requests, 2);
    });

    test('state survives an app restart', () async {
      await useOnThreeDaysWithThreeLessons(service());
      expect(await service().maybeRequestReview(), isTrue);
      expect(await service().maybeRequestReview(), isFalse);
    });

    test('skips unsupported platforms and unavailable stores', () async {
      await useOnThreeDaysWithThreeLessons(service());
      expect(await service(supported: false).maybeRequestReview(), isFalse);
      store.available = false;
      expect(await service().maybeRequestReview(), isFalse);
      expect(store.requests, 0);
      expect(preferences.getString(AppReviewService.lastRequestedKey), isNull);
    });
  });

  test('store listing is hidden on iOS until the App Store ID is set', () {
    expect(storeListingAvailable(platform: TargetPlatform.android), isTrue);
    expect(
      storeListingAvailable(platform: TargetPlatform.iOS, iosAppStoreId: ''),
      isFalse,
    );
    expect(
      storeListingAvailable(
        platform: TargetPlatform.iOS,
        iosAppStoreId: '1234567890',
      ),
      isTrue,
    );
    expect(
      storeListingAvailable(platform: TargetPlatform.android, isWeb: true),
      isFalse,
    );
    expect(storeListingAvailable(platform: TargetPlatform.windows), isFalse);
  });

  test('rate button copy is translated', () {
    expect(
      const AppLocalizations(Locale('en')).text('Beri rating aplikasi'),
      'Rate this app',
    );
  });
}
