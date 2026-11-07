import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/ads_controller.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// Adaptive banner. Renders nothing unless ads are on, the user is free, and
/// consent allows ad requests, so the layout never reserves empty space.
class AdBannerSlot extends ConsumerStatefulWidget {
  const AdBannerSlot({super.key});

  @override
  ConsumerState<AdBannerSlot> createState() => _AdBannerSlotState();
}

class _AdBannerSlotState extends ConsumerState<AdBannerSlot> {
  BannerAd? _ad;
  String? _unit;
  bool _loaded = false;
  int _width = 0;

  Future<void> _load(String unit, int width) async {
    if (_unit == unit && _width == width && _ad != null) return;
    _ad?.dispose();
    _unit = unit;
    _width = width;
    _loaded = false;
    final size = await AdSize.getCurrentOrientationAnchoredAdaptiveBannerAdSize(
      width,
    );
    if (!mounted || size == null) return;
    final ad = BannerAd(
      adUnitId: unit,
      size: size,
      request: const AdRequest(),
      listener: BannerAdListener(
        onAdLoaded: (_) {
          if (mounted) setState(() => _loaded = true);
        },
        onAdFailedToLoad: (ad, _) {
          ad.dispose();
          if (mounted) {
            setState(() {
              _ad = null;
              _loaded = false;
            });
          }
        },
      ),
    );
    _ad = ad;
    await ad.load();
  }

  @override
  void dispose() {
    _ad?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final config = ref.watch(adsReadyProvider).valueOrNull;
    final placement = config?.banner;
    if (placement == null || !placement.usable) {
      _ad?.dispose();
      _ad = null;
      _unit = null;
      return const SizedBox.shrink();
    }
    final width = MediaQuery.sizeOf(context).width.truncate();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _load(placement.unitId, width);
    });
    final ad = _ad;
    if (ad == null || !_loaded) return const SizedBox.shrink();
    return Semantics(
      label: 'Iklan'.localized(context),
      child: SizedBox(
        width: ad.size.width.toDouble(),
        height: ad.size.height.toDouble(),
        child: AdWidget(ad: ad),
      ),
    );
  }
}

enum NativeAdPlacement { news, jobs }

/// Native ad rendered with Google's template so it matches the card style.
/// It collapses to nothing until an ad has actually loaded.
class NativeAdSlot extends ConsumerStatefulWidget {
  const NativeAdSlot({required this.placement, super.key});

  final NativeAdPlacement placement;

  @override
  ConsumerState<NativeAdSlot> createState() => _NativeAdSlotState();
}

class _NativeAdSlotState extends ConsumerState<NativeAdSlot> {
  NativeAd? _ad;
  bool _loaded = false;
  bool _requested = false;

  void _load(String unit) {
    if (_requested) return;
    _requested = true;
    final scheme = Theme.of(context).colorScheme;
    final ad = NativeAd(
      adUnitId: unit,
      request: const AdRequest(),
      nativeTemplateStyle: NativeTemplateStyle(
        templateType: TemplateType.medium,
        mainBackgroundColor: scheme.surface,
        cornerRadius: 12,
        callToActionTextStyle: NativeTemplateTextStyle(
          textColor: scheme.onPrimary,
          backgroundColor: scheme.primary,
          style: NativeTemplateFontStyle.bold,
        ),
        primaryTextStyle: NativeTemplateTextStyle(textColor: scheme.onSurface),
        secondaryTextStyle: NativeTemplateTextStyle(
          textColor: scheme.onSurfaceVariant,
        ),
      ),
      listener: NativeAdListener(
        onAdLoaded: (_) {
          if (mounted) setState(() => _loaded = true);
        },
        onAdFailedToLoad: (ad, _) {
          ad.dispose();
          if (mounted) setState(() => _ad = null);
        },
      ),
    );
    _ad = ad;
    ad.load();
  }

  @override
  void dispose() {
    _ad?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final config = ref.watch(adsReadyProvider).valueOrNull;
    final placement = widget.placement == NativeAdPlacement.news
        ? config?.newsNative
        : config?.jobsNative;
    if (placement == null || !placement.usable) {
      return const SizedBox.shrink();
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _load(placement.unitId);
    });
    final ad = _ad;
    if (ad == null || !_loaded) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Semantics(
        label: 'Iklan'.localized(context),
        child: ConstrainedBox(
          constraints: const BoxConstraints(minHeight: 320, maxHeight: 360),
          child: AdWidget(ad: ad),
        ),
      ),
    );
  }
}

/// Returns the native ad spacing for a list, or 0 when ads are not active.
int nativeAdEvery(WidgetRef ref, NativeAdPlacement placement) {
  final config = ref.watch(adsReadyProvider).valueOrNull;
  final p = placement == NativeAdPlacement.news
      ? config?.newsNative
      : config?.jobsNative;
  if (p == null || !p.usable || p.every < 3) return 0;
  return p.every;
}
