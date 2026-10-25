import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/home/domain/home_banner_content.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

void main() {
  test('English home offer does not present the Indonesian-only price', () {
    final banner = homeBannersFor(
      _dashboard(),
      lifetime: false,
      english: true,
    ).first;

    expect(banner.id, 'lifetime-offer');
    expect(
      banner.description,
      'Your app store shows the local price at checkout.',
    );
    expect(banner.description, isNot(contains('Rp')));
  });

  test('Indonesian home offer keeps its configured price', () {
    final banner = homeBannersFor(_dashboard(), lifetime: false).first;

    expect(banner.description, 'All lessons • Rp1.499.000');
  });
}

HomeDashboard _dashboard() => const HomeDashboard(
  isAuthenticated: false,
  dailyGoal: HomeDailyGoal(minutesLearned: 0, targetMinutes: 20, streakDays: 0),
  categories: [],
  tracks: [],
  exams: [],
  plans: [
    SubscriptionPlan(
      id: 'lifetime',
      name: 'Lifetime Access',
      billingType: 'lifetime',
      priceIdr: 1499000,
      features: ['All lessons'],
      recommended: true,
    ),
  ],
  achievements: [],
);
