import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/app_update/app_update_controller.dart';
import 'package:nihongo_master/core/app_update/app_update_policy.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:url_launcher/url_launcher.dart';

class AppUpdateGate extends ConsumerStatefulWidget {
  const AppUpdateGate({required this.child, super.key});
  final Widget child;

  @override
  ConsumerState<AppUpdateGate> createState() => _AppUpdateGateState();
}

class _AppUpdateGateState extends ConsumerState<AppUpdateGate>
    with WidgetsBindingObserver {
  Timer? _timer;
  bool _foreground = true;
  bool _openingStore = false;
  String? _storeError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        unawaited(ref.read(appUpdateControllerProvider.notifier).check());
      }
    });
    _timer = Timer.periodic(const Duration(minutes: 5), (_) {
      if (_foreground && mounted) {
        unawaited(ref.read(appUpdateControllerProvider.notifier).check());
      }
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _foreground = state == AppLifecycleState.resumed;
    if (_foreground) {
      unawaited(ref.read(appUpdateControllerProvider.notifier).check());
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  Future<void> _openStore(AppUpdateState state) async {
    if (_openingStore) return;
    setState(() {
      _openingStore = true;
      _storeError = null;
    });
    try {
      var opened = false;
      if (state.device!.platform == 'android') {
        try {
          opened = await launchUrl(
            Uri.parse('market://details?id=com.codeverta.nihongomaster'),
            mode: LaunchMode.externalApplication,
          ).timeout(const Duration(seconds: 8));
        } on Object {
          // Devices without the native Play Store still get the HTTPS listing.
        }
      }
      if (!opened) {
        opened = await launchUrl(
          Uri.parse(state.policy!.storeUrl),
          mode: LaunchMode.externalApplication,
        ).timeout(const Duration(seconds: 8));
      }
      if (!opened) throw StateError('Store unavailable');
    } on Object {
      if (mounted) {
        setState(
          () => _storeError =
              'Store tidak dapat dibuka. Coba lagi atau hubungi rabih@codeverta.com.',
        );
      }
    } finally {
      if (mounted) setState(() => _openingStore = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(appUpdateControllerProvider);
    // Cached lessons must remain usable while an online version check is still
    // pending. A confirmed mandatory-update dialog continues to block access.
    final obscured = state.showDialog;
    return Stack(
      fit: StackFit.expand,
      children: [
        ExcludeFocus(
          excluding: obscured,
          child: ExcludeSemantics(
            excluding: obscured,
            child: AbsorbPointer(absorbing: obscured, child: widget.child),
          ),
        ),
        if (state.showDialog) ...[
          const ModalBarrier(dismissible: false, color: Colors.black54),
          SafeArea(
            child: Center(
              child: SingleChildScrollView(
                child: _dialog(context, state),
              ),
            ),
          ),
        ],
      ],
    );
  }

  Widget _dialog(BuildContext context, AppUpdateState state) {
    final device = state.device!;
    final policy = state.policy!;
    final needsUpdate = state.requirement == AppUpdateRequirement.update;
    final canContinue =
        state.requirement == AppUpdateRequirement.unsupportedAllowed;
    final store = device.platform == 'android' ? 'Play Store' : 'App Store';
    const androidVersions = {
      21: '5.0',
      22: '5.1',
      23: '6',
      24: '7',
      25: '7.1',
      26: '8',
      27: '8.1',
      28: '9',
      29: '10',
      30: '11',
      31: '12',
      32: '12L',
      33: '13',
      34: '14',
      35: '15',
      36: '16',
    };
    final androidMinimum = androidVersions[policy.minimumAndroidApi];
    final requirement = device.platform == 'android'
        ? androidMinimum == null
            ? 'versi Android yang lebih baru'.localized(context)
            : 'Android $androidMinimum ${'atau lebih baru'.localized(context)}'
        : 'iOS ${policy.minimumIOSVersion} ${'atau lebih baru'.localized(context)}';
    final instructions = device.platform == 'android'
        ? 'Buka Pengaturan HP > Sistem > Pembaruan sistem. Nama menu dapat berbeda pada setiap merek.'
              .localized(context)
        : 'Buka Pengaturan > Umum > Pembaruan Perangkat Lunak.'.localized(
            context,
          );
    return AlertDialog(
      title: Text(
        needsUpdate
            ? 'Perbarui Nihongo Master'.localized(context)
            : 'Perangkat belum mendukung versi baru'.localized(context),
      ),
      icon: Icon(needsUpdate ? Icons.system_update : Icons.phonelink_erase),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              needsUpdate
                  ? '${'Versi aplikasi ini sudah tidak didukung.'.localized(context)} ${'Perbarui melalui'.localized(context)} $store ${'untuk melanjutkan.'.localized(context)}'
                  : '${'Versi aplikasi yang diwajibkan membutuhkan'.localized(context)} $requirement. ${'Perangkat Anda menggunakan'.localized(context)} ${device.platform == 'android' ? 'Android' : 'iOS'} ${device.osVersion}.',
            ),
            const SizedBox(height: 12),
            if (needsUpdate)
              Text(
                '${'Versi terpasang'.localized(context)}: ${device.version}\n${'Versi terbaru'.localized(context)}: ${policy.latestVersion}',
              )
            else ...[
              Text(instructions),
              const SizedBox(height: 12),
              Text(
                canContinue
                    ? 'Anda masih dapat menggunakan versi yang terpasang. Perbarui OS sebelum memasang versi aplikasi baru.'
                          .localized(context)
                    : 'Jika pembaruan OS sudah tidak tersedia, gunakan perangkat lain yang kompatibel. Untuk bantuan: rabih@codeverta.com.'
                          .localized(context),
              ),
            ],
            if (policy.message.isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(policy.message),
            ],
            if (needsUpdate)
              Padding(
                padding: const EdgeInsets.only(top: 12),
                child: Text(
                  'Jika Update tidak muncul di store, periksa akun store dan ketersediaan rilis untuk perangkat Anda. Hubungi rabih@codeverta.com jika perangkat dinyatakan tidak kompatibel.'
                      .localized(context),
                ),
              ),
            if (state.error != null || _storeError != null) ...[
              const SizedBox(height: 12),
              Text(
                (_storeError ?? state.error!).localized(context),
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: state.checking
              ? null
              : () => ref.read(appUpdateControllerProvider.notifier).check(),
          child: Text(
            state.checking
                ? 'Memeriksa…'.localized(context)
                // 'Periksa lagi' is already mapped to "Review again" for exams.
                : Localizations.localeOf(context).languageCode == 'en'
                ? 'Check again'
                : 'Periksa lagi',
          ),
        ),
        if (canContinue)
          FilledButton(
            onPressed: () => ref
                .read(appUpdateControllerProvider.notifier)
                .continueOnLegacyDevice(),
            child: Text('Lanjutkan versi ini'.localized(context)),
          ),
        if (needsUpdate)
          FilledButton(
            onPressed: _openingStore ? null : () => _openStore(state),
            child: Text(
              _openingStore
                  ? 'Membuka store…'.localized(context)
                  : 'Update sekarang'.localized(context),
            ),
          ),
      ],
    );
  }
}
