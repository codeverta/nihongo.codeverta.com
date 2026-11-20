import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';

/// Marks a "finished studying" result screen. While it is mounted, a
/// full-screen ad may load in the background; when the user leaves it (back,
/// "done", or moving on) the ad is shown if this kind is switched on and the
/// interval since the last full-screen ad has passed. Nothing is ever shown
/// while studying or when content opens (see docs/ads-xp-spec.md).
///
/// [completionKey] names the finished item (for example
/// `grammar:grammar-n3-01`) so an item the user paid for with a rewarded ad
/// gets no full-screen ad.
class CompletionAdBoundary extends ConsumerStatefulWidget {
  const CompletionAdBoundary({
    required this.kind,
    this.completionKey,
    this.child = const SizedBox.shrink(),
    super.key,
  });

  final CompletionKind kind;
  final String? completionKey;
  final Widget child;

  @override
  ConsumerState<CompletionAdBoundary> createState() =>
      _CompletionAdBoundaryState();
}

class _CompletionAdBoundaryState extends ConsumerState<CompletionAdBoundary> {
  late final InterstitialController _controller;

  @override
  void initState() {
    super.initState();
    // Captured now: ref cannot be used in dispose.
    _controller = ref.read(interstitialControllerProvider);
    final kind = widget.kind;
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => _controller.prepareAfterCompletion(kind),
    );
  }

  @override
  void dispose() {
    final controller = _controller;
    final kind = widget.kind;
    final key = widget.completionKey;
    // After the frame, so the ad never covers a half-built route.
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => unawaited(controller.showAfterCompletion(kind, key: key)),
    );
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
