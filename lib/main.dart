import 'dart:async';
import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:nihongo_master/core/app_review/app_review_service.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/app_update/app_update_controller.dart';
import 'package:nihongo_master/core/app_update/app_update_gate.dart';
import 'package:nihongo_master/core/branding/app_branding.dart';
import 'package:nihongo_master/core/diagnostics/error_reporter.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/notifications/notification_lifecycle.dart';
import 'package:nihongo_master/core/offline/offline_lifecycle.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/settings/theme_mode_preference.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  runZonedGuarded(
    () async {
      WidgetsFlutterBinding.ensureInitialized();
      FlutterError.onError = (details) {
        FlutterError.presentError(details);
        unawaited(
          ErrorReporter.capture(
            details.exception,
            details.stack ?? StackTrace.current,
          ),
        );
      };
      PlatformDispatcher.instance.onError = (error, stackTrace) {
        unawaited(ErrorReporter.capture(error, stackTrace, severity: 'fatal'));
        return false;
      };

      await initializeDateFormatting();
      final preferences = await SharedPreferences.getInstance();
      ErrorReporter.initialize(preferences);
      unawaited(
        AppReviewService(
          preferences,
          const InAppReviewClient(),
        ).recordActiveDay(),
      );
      runApp(
        ProviderScope(
          overrides: [sharedPreferencesProvider.overrideWithValue(preferences)],
          child: const NihongoMasterApp(),
        ),
      );
    },
    (error, stackTrace) {
      unawaited(ErrorReporter.capture(error, stackTrace, severity: 'fatal'));
    },
  );
}

class NihongoMasterApp extends ConsumerWidget {
  const NihongoMasterApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    final locale = ref.watch(appPresentationLocaleProvider);
    final supportedLocales = ref.watch(appSupportedLocalesProvider);
    ref
      ..watch(billingControllerProvider)
      ..watch(analyticsBootstrapProvider);

    return MaterialApp.router(
      title:
          ref.watch(appBrandingProvider).valueOrNull?.appName ??
          'Nihongo Master',
      builder: (context, child) => OfflineLifecycle(
        child: NotificationLifecycle(
          child: AppUpdateGate(child: BrandStartup(child: child!)),
        ),
      ),
      debugShowCheckedModeBanner: false,
      locale: locale,
      supportedLocales: supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ref.watch(themeModeProvider),
      routeInformationProvider: router.routeInformationProvider,
      routeInformationParser: router.routeInformationParser,
      routerDelegate: router.routerDelegate,
      backButtonDispatcher: ref.watch(appUpdateBackDispatcherProvider),
    );
  }
}
