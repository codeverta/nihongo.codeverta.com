import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';

const _cacheKey = 'app_branding_v1';

class AppBranding {
  const AppBranding({
    this.appName = 'Nihongo Master',
    this.tagline = 'Selangkah lebih dekat dengan Jepang',
    this.iconUrl = '',
    this.splashUrl = '',
    this.backgroundColor = const Color(0xFFFAFAFA),
  });

  factory AppBranding.fromJson(Map<String, dynamic> json) {
    String value(String key, String fallback) =>
        json[key] is String ? json[key] as String : fallback;
    String imageUrl(String key) {
      final raw = value(key, '');
      final uri = Uri.tryParse(raw);
      return uri?.scheme == 'https' && (uri?.host.isNotEmpty ?? false)
          ? raw
          : '';
    }

    final hex = value('backgroundColor', '#FAFAFA');
    return AppBranding(
      appName: value('appName', 'Nihongo Master'),
      tagline: value('tagline', 'Selangkah lebih dekat dengan Jepang'),
      iconUrl: imageUrl('iconUrl'),
      splashUrl: imageUrl('splashUrl'),
      backgroundColor: RegExp(r'^#[0-9a-fA-F]{6}$').hasMatch(hex)
          ? Color(int.parse('FF${hex.substring(1)}', radix: 16))
          : const Color(0xFFFAFAFA),
    );
  }

  final String appName;
  final String tagline;
  final String iconUrl;
  final String splashUrl;
  final Color backgroundColor;
}

final cachedBrandingProvider = Provider<AppBranding>((ref) {
  final raw = ref.watch(sharedPreferencesProvider).getString(_cacheKey);
  try {
    if (raw != null) {
      return AppBranding.fromJson(jsonDecode(raw) as Map<String, dynamic>);
    }
  } on Object {
    // A corrupt cache must never prevent startup.
  }
  return const AppBranding();
});

final appBrandingProvider = FutureProvider<AppBranding>((ref) async {
  final fallback = ref.read(cachedBrandingProvider);
  try {
    final response = await ref
        .read(dioProvider)
        .get<Map<String, dynamic>>(
          '/branding',
        )
        .timeout(const Duration(milliseconds: 1800));
    final data = response.data?['data'] as Map<String, dynamic>;
    final branding = AppBranding.fromJson(data);
    await ref
        .read(sharedPreferencesProvider)
        .setString(_cacheKey, jsonEncode(data));
    return branding;
  } on Object {
    return fallback;
  }
});

class BrandIcon extends StatelessWidget {
  const BrandIcon({required this.branding, this.size = 120, super.key});
  final AppBranding branding;
  final double size;

  @override
  Widget build(BuildContext context) {
    Widget fallback() => Image.asset(
          'assets/images/app-icon-small.webp',
          width: size,
          height: size,
          fit: BoxFit.contain,
        );
    if (branding.iconUrl.isEmpty) return fallback();
    return Image.network(
      branding.iconUrl,
      width: size,
      height: size,
      fit: BoxFit.contain,
      errorBuilder: (_, error, stack) => fallback(),
      loadingBuilder: (_, child, progress) =>
          progress == null ? child : fallback(),
    );
  }
}

/// Shows bundled/cached artwork immediately; startup never waits indefinitely
/// for a backend or image server. Router state remains mounted underneath.
class BrandStartup extends ConsumerStatefulWidget {
  const BrandStartup({required this.child, super.key});
  final Widget child;

  @override
  ConsumerState<BrandStartup> createState() => _BrandStartupState();
}

class _BrandStartupState extends ConsumerState<BrandStartup> {
  Timer? _timer;
  bool _complete = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _timer = Timer(const Duration(milliseconds: 2200), () {
        if (mounted) setState(() => _complete = true);
      });
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final branding = ref.watch(appBrandingProvider).valueOrNull ??
        ref.watch<AppBranding>(cachedBrandingProvider);
    return Stack(
      fit: StackFit.expand,
      children: [
        Offstage(offstage: !_complete, child: widget.child),
        if (!_complete)
          Positioned.fill(
            child: Semantics(
              label: '${branding.appName}, membuka aplikasi',
              child: Material(
                color: branding.backgroundColor,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    if (branding.splashUrl.isNotEmpty)
                      Image.network(
                        branding.splashUrl,
                        fit: BoxFit.contain,
                        errorBuilder: (_, error, stack) =>
                            AppSplashArtwork(branding: branding),
                        loadingBuilder: (_, child, progress) => progress == null
                            ? child
                            : AppSplashArtwork(branding: branding),
                      )
                    else
                      AppSplashArtwork(branding: branding),
                    const Positioned(
                      bottom: 60,
                      left: 0,
                      right: 0,
                      child: Center(
                        child: SizedBox(
                          width: 22,
                          height: 22,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Color(0xFF1D2088),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
      ],
    );
  }
}

class AppSplashArtwork extends StatelessWidget {
  const AppSplashArtwork({required this.branding, super.key});
  final AppBranding branding;

  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              BrandIcon(branding: branding, size: 160),
              const SizedBox(height: 24),
              Text(
                branding.appName,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 30,
                  fontWeight: FontWeight.w800,
                  color: Color(0xFF1D2088),
                ),
              ),
              const SizedBox(height: 12),
              Text(
                branding.tagline,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 14, color: Color(0xFF666666)),
              ),
            ],
          ),
        ),
      );
}
