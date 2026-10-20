import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/home/domain/home_banner_content.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

HomeDashboard dashboard({
  bool plan = true,
  List<HomeBanner> banners = const [],
}) => HomeDashboard(
  isAuthenticated: true,
  dailyGoal: const HomeDailyGoal(
    minutesLearned: 0,
    targetMinutes: 20,
    streakDays: 0,
  ),
  banners: banners,
  categories: const [],
  tracks: const [],
  exams: const [],
  achievements: const [],
  plans: plan
      ? const [
          SubscriptionPlan(
            id: 'lifetime',
            name: 'Lifetime Access',
            billingType: 'lifetime',
            priceIdr: 99000,
            features: ['Akses lengkap'],
            recommended: true,
          ),
        ]
      : [],
);

void main() {
  test('free and lifetime accounts always have five distinct destinations', () {
    final free = homeBannersFor(dashboard(), lifetime: false);
    expect(free.map((b) => b.action), [
      'premium',
      'learn',
      'ssw',
      'jobs',
      'manga',
    ]);
    expect(free.first.description, contains('Rp99.000'));
    final paid = homeBannersFor(dashboard(), lifetime: true);
    expect(paid.map((b) => b.action), [
      'learn',
      'ssw',
      'jobs',
      'anki',
      'manga',
    ]);
  });
  test('older APIs missing Manga and pricing still produce five slides', () {
    final items = homeBannersFor(dashboard(plan: false), lifetime: false);
    expect(items, hasLength(5));
    expect(items.last.action, 'manga');
    expect(items.map((b) => b.action).toSet(), hasLength(5));
  });
  test(
    'configured Manga image is preserved and duplicates do not add slides',
    () {
      const manga = HomeBanner(
        id: 'configured-manga',
        title: 'Manga pilihan',
        description: '',
        imageUrl: 'https://our-cos.example/manga.png',
        action: 'manga',
        buttonLabel: 'Baca',
      );
      final items = homeBannersFor(
        dashboard(banners: [manga, manga]),
        lifetime: false,
      );
      expect(items, hasLength(5));
      expect(items.where((b) => b.action == 'manga'), hasLength(1));
      expect(items.last.imageUrl, manga.imageUrl);
    },
  );
}
