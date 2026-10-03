import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/branding/app_branding.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test('invalid remote image and color fall back safely', () {
    final brand = AppBranding.fromJson({
      'iconUrl': 'javascript:alert(1)',
      'splashUrl': 'http://example.com/image.png',
      'backgroundColor': 'bad',
    });
    expect(brand.iconUrl, isEmpty);
    expect(brand.splashUrl, isEmpty);
    expect(brand.backgroundColor, const Color(0xFFFAFAFA));
  });

  testWidgets('bundled splash yields to app even without remote branding', (
    tester,
  ) async {
    SharedPreferences.setMockInitialValues({'app_branding_v1': 'corrupt'});
    final preferences = await SharedPreferences.getInstance();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPreferencesProvider.overrideWithValue(preferences),
          appBrandingProvider.overrideWith((ref) async => const AppBranding()),
        ],
        child: const MaterialApp(
          home: BrandStartup(child: Text('Application content')),
        ),
      ),
    );
    expect(find.text('Nihongo Master'), findsOneWidget);
    await tester.pump(const Duration(milliseconds: 2300));
    await tester.pump();
    expect(find.text('Nihongo Master'), findsNothing);
    expect(find.text('Application content'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
