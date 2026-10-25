import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/settings/display_settings_screen.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('fresh install follows Indonesian and English device locales', (
    tester,
  ) async {
    for (final testCase in [
      (deviceLocale: const Locale('id', 'ID'), expected: const Locale('id')),
      (deviceLocale: const Locale('en', 'GB'), expected: const Locale('en')),
    ]) {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final container = _container(
        preferences,
        deviceLocale: testCase.deviceLocale,
      );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const _LocalizationHarness(),
        ),
      );
      await tester.pumpAndSettle();

      expect(container.read(appLanguageProvider), testCase.expected);
      expect(
        find.text(
          '${testCase.expected.languageCode}_${testCase.deviceLocale.countryCode}',
        ),
        findsOneWidget,
      );
      expect(
        find.text(
          testCase.expected.languageCode == 'en'
              ? 'App language'
              : 'Bahasa aplikasi',
        ),
        findsOneWidget,
      );

      await tester.pumpWidget(const SizedBox.shrink());
      container.dispose();
    }
  });

  testWidgets(
    'unsupported device locale falls back to English for a fresh install',
    (tester) async {
      SharedPreferences.setMockInitialValues({});
      final preferences = await SharedPreferences.getInstance();
      final container = _container(
        preferences,
        deviceLocale: const Locale('fr', 'FR'),
      );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const _LocalizationHarness(),
        ),
      );
      await tester.pumpAndSettle();

      expect(container.read(appLanguageProvider), const Locale('en'));
      expect(find.text('en_FR'), findsOneWidget);
      expect(find.text('App language'), findsOneWidget);
      expect(tester.takeException(), isNull);
      container.dispose();
    },
  );

  testWidgets(
    'existing installation stays Indonesian and manual choice survives restart',
    (tester) async {
      SharedPreferences.setMockInitialValues({'is_first_time': false});
      final preferences = await SharedPreferences.getInstance();
      final existingInstall = _container(
        preferences,
        deviceLocale: const Locale('en'),
      );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: existingInstall,
          child: const _LocalizationHarness(),
        ),
      );
      await tester.pumpAndSettle();
      expect(existingInstall.read(appLanguageProvider), const Locale('id'));
      expect(find.text('Bahasa aplikasi'), findsOneWidget);

      await tester.tap(find.byKey(const Key('language-option-en')));
      await tester.pumpAndSettle();
      expect(existingInstall.read(appLanguageProvider), const Locale('en'));
      expect(find.text('App language'), findsOneWidget);

      await tester.pumpWidget(const SizedBox.shrink());
      existingInstall.dispose();
      final restarted = _container(
        preferences,
        deviceLocale: const Locale('id'),
      );
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: restarted,
          child: const _LocalizationHarness(),
        ),
      );
      await tester.pumpAndSettle();

      expect(restarted.read(appLanguageProvider), const Locale('en'));
      expect(find.text('en'), findsOneWidget);
      expect(find.text('App language'), findsOneWidget);
      expect(find.text('日本語を学びましょう'), findsOneWidget);
      expect(tester.takeException(), isNull);
      restarted.dispose();
    },
  );

  test(
    'English is not selectable unless public release or internal QA enables it',
    () {
      expect(
        EnglishReleaseGate.isEnabled(
          isPublicRelease: false,
          isInternalQa: false,
          isDebugBuild: false,
        ),
        isFalse,
      );
      expect(
        EnglishReleaseGate.isEnabled(
          isPublicRelease: false,
          isInternalQa: true,
          isDebugBuild: true,
        ),
        isTrue,
      );
    },
  );
}

ProviderContainer _container(
  SharedPreferences preferences, {
  required Locale deviceLocale,
}) => ProviderContainer(
  overrides: [
    sharedPreferencesProvider.overrideWithValue(preferences),
    deviceLocaleProvider.overrideWithValue(deviceLocale),
    englishAvailabilityProvider.overrideWithValue(true),
  ],
);

class _LocalizationHarness extends ConsumerWidget {
  const _LocalizationHarness();

  @override
  Widget build(BuildContext context, WidgetRef ref) => MaterialApp(
    locale: ref.watch(appPresentationLocaleProvider),
    supportedLocales: ref.watch(appSupportedLocalesProvider),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    home: Scaffold(
      body: Column(
        children: [
          Builder(
            builder: (context) =>
                Text(Localizations.localeOf(context).toString()),
          ),
          Text('日本語を学びましょう'.localized(context)),
          const Expanded(child: DisplaySettingsScreen()),
        ],
      ),
    ),
  );
}
