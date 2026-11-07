import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

/// Ads are only ever shown to signed-in or guest users on the free plan.
/// Premium and Lifetime accounts never see them.
bool adsAllowedFor({required AuthenticationState auth}) {
  if (auth.isLoading) return false;
  final user = auth.user;
  if (user == null) return true;
  return !(user.isPremium || user.isLifetime || user.isOfflinePremium);
}

bool get adsSupportedPlatform =>
    !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

final adsConfigProvider = FutureProvider<AdsConfig>((ref) async {
  if (!adsSupportedPlatform) return AdsConfig.off;
  try {
    final response = await ref
        .watch(dioProvider)
        .get<Map<String, dynamic>>('/ads-config');
    final config = AdsConfig.fromJson(response.data?['data']);
    return kDebugMode ? config.withTestUnits() : config;
  } on DioException {
    return AdsConfig.off;
  } on Object {
    return AdsConfig.off;
  }
});

/// Initializes the Mobile Ads SDK after consent is resolved, and only when the
/// server switch is on and the user is on the free plan.
class AdsRuntime {
  bool _started = false;
  bool _canRequest = false;

  bool get canRequestAds => _canRequest;

  Future<bool> start() async {
    if (_started) return _canRequest;
    _started = true;
    try {
      final completer = Completer<void>();
      ConsentInformation.instance.requestConsentInfoUpdate(
        ConsentRequestParameters(),
        () async {
          await ConsentForm.loadAndShowConsentFormIfRequired((_) {});
          if (!completer.isCompleted) completer.complete();
        },
        (_) {
          if (!completer.isCompleted) completer.complete();
        },
      );
      await completer.future.timeout(const Duration(seconds: 20));
      _canRequest = await ConsentInformation.instance.canRequestAds();
      if (_canRequest) await MobileAds.instance.initialize();
    } on Object catch (error) {
      debugPrint('Ads unavailable: $error');
      _canRequest = false;
    }
    return _canRequest;
  }

  Future<bool> privacyOptionsRequired() async {
    try {
      return await ConsentInformation.instance
              .getPrivacyOptionsRequirementStatus() ==
          PrivacyOptionsRequirementStatus.required;
    } on Object {
      return false;
    }
  }

  Future<void> showPrivacyOptions() async {
    try {
      await ConsentForm.showPrivacyOptionsForm((_) {});
    } on Object catch (error) {
      debugPrint('Privacy options unavailable: $error');
    }
  }
}

final adsRuntimeProvider = Provider<AdsRuntime>((ref) => AdsRuntime());

/// True when an ad may be requested right now. Watch this from ad widgets.
final adsReadyProvider = FutureProvider<AdsConfig?>((ref) async {
  final auth = ref.watch(authControllerProvider);
  if (!adsAllowedFor(auth: auth)) return null;
  final config = await ref.watch(adsConfigProvider.future);
  if (!config.enabled) return null;
  final ok = await ref.read(adsRuntimeProvider).start();
  return ok ? config : null;
});
