import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/app_update/app_update_controller.dart';
import 'package:nihongo_master/core/app_update/app_update_policy.dart';
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
    final obscured = !state.initialized || state.showDialog;
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
        if (!state.initialized)
          const Material(child: Center(child: CircularProgressIndicator())),
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
            ? 'versi Android yang lebih baru'
            : 'Android $androidMinimum atau lebih baru'
        : 'iOS ${policy.minimumIOSVersion} atau lebih baru';
    final instructions = device.platform == 'android'
        ? 'Buka Pengaturan HP > Sistem > Pembaruan sistem. Nama menu dapat berbeda pada setiap merek.'
        : 'Buka Pengaturan > Umum > Pembaruan Perangkat Lunak.';
    return AlertDialog(
      title: Text(
        needsUpdate
            ? 'Perbarui Nihongo Master'
            : 'Perangkat belum mendukung versi baru',
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
                  ? 'Versi aplikasi ini sudah tidak didukung. Perbarui melalui $store untuk melanjutkan.'
                  : 'Versi aplikasi yang diwajibkan membutuhkan $requirement. Perangkat Anda menggunakan ${device.platform == 'android' ? 'Android' : 'iOS'} ${device.osVersion}.',
            ),
            const SizedBox(height: 12),
            if (needsUpdate)
              Text(
                'Versi terpasang: ${device.version}\nVersi terbaru: ${policy.latestVersion}',
              )
            else ...[
              Text(instructions),
              const SizedBox(height: 12),
              Text(
                canContinue
                    ? 'Anda masih dapat menggunakan versi yang terpasang. Perbarui OS sebelum memasang versi aplikasi baru.'
                    : 'Jika pembaruan OS sudah tidak tersedia, gunakan perangkat lain yang kompatibel. Untuk bantuan: rabih@codeverta.com.',
              ),
            ],
            if (policy.message.isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(policy.message),
            ],
            if (needsUpdate)
              const Padding(
                padding: EdgeInsets.only(top: 12),
                child: Text(
                  'Jika Update tidak muncul di store, periksa akun store dan ketersediaan rilis untuk perangkat Anda. Hubungi rabih@codeverta.com jika perangkat dinyatakan tidak kompatibel.',
                ),
              ),
            if (state.error != null || _storeError != null) ...[
              const SizedBox(height: 12),
              Text(
                _storeError ?? state.error!,
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
          child: Text(state.checking ? 'Memeriksa…' : 'Periksa lagi'),
        ),
        if (canContinue)
          FilledButton(
            onPressed: () => ref
                .read(appUpdateControllerProvider.notifier)
                .continueOnLegacyDevice(),
            child: const Text('Lanjutkan versi ini'),
          ),
        if (needsUpdate)
          FilledButton(
            onPressed: _openingStore ? null : () => _openStore(state),
            child: Text(_openingStore ? 'Membuka store…' : 'Update sekarang'),
          ),
      ],
    );
  }
}
