import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/app_update/app_update_policy.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

class AppUpdateState {
  const AppUpdateState({
    this.device,
    this.policy,
    this.requirement = AppUpdateRequirement.supported,
    this.initialized = false,
    this.checking = false,
    this.dismissed = false,
    this.error,
  });

  final AppDeviceInfo? device;
  final AppUpdatePolicy? policy;
  final AppUpdateRequirement requirement;
  final bool initialized;
  final bool checking;
  final bool dismissed;
  final String? error;

  bool get showDialog =>
      requirement != AppUpdateRequirement.supported && !dismissed;
  bool get interceptBack => !initialized || showDialog;
}

final appUpdateControllerProvider =
    StateNotifierProvider<AppUpdateController, AppUpdateState>((ref) {
  final dio = ref.read(dioProvider);
  final mobile = !kIsWeb &&
      {TargetPlatform.android, TargetPlatform.iOS}
          .contains(defaultTargetPlatform);
  return AppUpdateController(
    deviceInfo: () async {
      if (!mobile) return null;
      const channel = MethodChannel('com.codeverta.nihongomaster/app_info');
      final info = await channel.invokeMapMethod<String, dynamic>('getAppInfo');
      return AppDeviceInfo.fromJson(info!);
    },
    fetchPolicy: () async {
      final token = CancelToken();
      final timer = Timer(const Duration(seconds: 4), token.cancel);
      try {
        final response = await dio.get<Map<String, dynamic>>(
          '/app-update-policy',
          cancelToken: token,
          options: Options(headers: {'Cache-Control': 'no-cache'}),
        );
        return response.data!['data'] as Map<String, dynamic>;
      } finally {
        timer.cancel();
      }
    },
  );
});

class AppUpdateController extends StateNotifier<AppUpdateState> {
  AppUpdateController({required this.deviceInfo, required this.fetchPolicy})
      : super(const AppUpdateState());

  final Future<AppDeviceInfo?> Function() deviceInfo;
  final Future<Map<String, dynamic>> Function() fetchPolicy;
  String? _dismissedPolicy;
  String? _policyKey;

  Future<void> check() async {
    if (state.checking) return;
    state = _copy(checking: true);
    try {
      final device = await deviceInfo().timeout(const Duration(seconds: 2));
      if (!mounted) return;
      if (device == null) {
        state = const AppUpdateState(initialized: true);
        return;
      }
      final data = await fetchPolicy().timeout(const Duration(seconds: 5));
      if (!mounted) return;
      final json = data[device.platform] as Map<String, dynamic>;
      final policy = AppUpdatePolicy.fromJson(json, device.platform);
      final requirement = policy.evaluate(device);
      _policyKey = jsonEncode(json);
      state = AppUpdateState(
        initialized: true,
        device: device,
        policy: policy,
        requirement: requirement,
        dismissed: requirement == AppUpdateRequirement.unsupportedAllowed &&
            _dismissedPolicy == _policyKey,
      );
    } on Object {
      if (!mounted) return;
      // Cold-start outages fail open. An already confirmed requirement remains
      // visible until the server supplies a valid replacement or rollback.
      state = _copy(
        initialized: true,
        checking: false,
        error:
            'Belum dapat memeriksa versi terbaru. Periksa koneksi lalu coba lagi.',
      );
    }
  }

  void continueOnLegacyDevice() {
    if (state.requirement != AppUpdateRequirement.unsupportedAllowed) return;
    _dismissedPolicy = _policyKey;
    state = _copy(dismissed: true);
  }

  AppUpdateState _copy({
    bool? initialized,
    bool? checking,
    bool? dismissed,
    String? error,
  }) =>
      AppUpdateState(
        device: state.device,
        policy: state.policy,
        requirement: state.requirement,
        initialized: initialized ?? state.initialized,
        checking: checking ?? state.checking,
        dismissed: dismissed ?? state.dismissed,
        error: error,
      );
}

class AppUpdateBackDispatcher extends RootBackButtonDispatcher {
  AppUpdateBackDispatcher(this.shouldIntercept);
  final bool Function() shouldIntercept;

  @override
  Future<bool> didPopRoute() async =>
      shouldIntercept() || await super.didPopRoute();
}

final appUpdateBackDispatcherProvider =
    Provider<AppUpdateBackDispatcher>((ref) {
  return AppUpdateBackDispatcher(
    () => ref.read(appUpdateControllerProvider).interceptBack,
  );
});
