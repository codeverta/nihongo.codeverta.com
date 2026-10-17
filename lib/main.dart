import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/app_update/app_update_controller.dart';
import 'package:nihongo_master/core/app_update/app_update_gate.dart';
import 'package:nihongo_master/core/branding/app_branding.dart';
import 'package:nihongo_master/core/notifications/notification_lifecycle.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/core/theme/app_theme.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';
import 'package:shared_preferences/shared_preferences.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final preferences = await SharedPreferences.getInstance();

  runApp(
    ProviderScope(
      overrides: [sharedPreferencesProvider.overrideWithValue(preferences)],
      child: const NihongoMasterApp(),
    ),
  );
}

class NihongoMasterApp extends ConsumerWidget {
  const NihongoMasterApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    ref.watch(billingControllerProvider);

    return MaterialApp.router(
      title:
          ref.watch(appBrandingProvider).valueOrNull?.appName ??
          'Nihongo Master',
      builder: (context, child) => NotificationLifecycle(
        child: AppUpdateGate(child: BrandStartup(child: child!)),
      ),
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      routeInformationProvider: router.routeInformationProvider,
      routeInformationParser: router.routeInformationParser,
      routerDelegate: router.routerDelegate,
      backButtonDispatcher: ref.watch(appUpdateBackDispatcherProvider),
    );
  }
}
