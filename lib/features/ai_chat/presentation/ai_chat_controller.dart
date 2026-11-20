import 'dart:async';
import 'dart:math' as math;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/ads/ads_config.dart';
import 'package:nihongo_master/core/ads/interstitial_ads.dart';
import 'package:nihongo_master/core/ads/rewarded_ads.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/features/ai_chat/data/ai_chat_repository.dart';

typedef AiChatKey = ({AiLessonType type, String lessonId});

/// A server decision the UI must present (sheet or dialog).
enum AiNotice {
  /// Free quota used; offer an opt-in rewarded ad.
  adOffer,

  /// Nothing more today for a free user (iOS, ads off, or ad cap reached).
  dailyLimit,

  /// A Lifetime user reached the daily limit.
  premiumLimit,

  /// A Lifetime user used the daily messages; offer an optional rewarded ad
  /// for [AiChatState.messagesPerAd] more.
  lifetimeAdOffer,
}

class AiChatState {
  const AiChatState({
    this.messages = const [],
    this.streaming,
    this.loadingHistory = false,
    this.sending = false,
    this.notice,
    this.error,
    this.retryable = false,
    this.messagesPerAd = 0,
  });

  final List<AiChatMessage> messages;

  /// The assistant answer currently being streamed, if any.
  final String? streaming;
  final bool loadingHistory;
  final bool sending;
  final AiNotice? notice;
  final String? error;

  /// Whether the last failed request can be retried as is.
  final bool retryable;

  /// Messages one ad adds, for [AiNotice.lifetimeAdOffer].
  final int messagesPerAd;

  AiChatState copyWith({
    List<AiChatMessage>? messages,
    String? Function()? streaming,
    bool? loadingHistory,
    bool? sending,
    AiNotice? Function()? notice,
    String? Function()? error,
    bool? retryable,
    int? messagesPerAd,
  }) => AiChatState(
    messages: messages ?? this.messages,
    streaming: streaming != null ? streaming() : this.streaming,
    loadingHistory: loadingHistory ?? this.loadingHistory,
    sending: sending ?? this.sending,
    notice: notice != null ? notice() : this.notice,
    error: error != null ? error() : this.error,
    retryable: retryable ?? this.retryable,
    messagesPerAd: messagesPerAd ?? this.messagesPerAd,
  );
}

typedef _Request = ({AiChip? chip, String text});

class AiChatController extends StateNotifier<AiChatState> {
  AiChatController(this._ref, this.key) : super(const AiChatState()) {
    unawaited(loadHistory());
  }

  final Ref _ref;
  final AiChatKey key;
  _Request? _pending;
  StreamSubscription<AiStreamEvent>? _subscription;
  Timer? _reveal;

  AiChatRepository get _repository => _ref.read(aiChatRepositoryProvider);
  AnalyticsService get _analytics => _ref.read(analyticsServiceProvider);

  Future<void> loadHistory() async {
    state = state.copyWith(loadingHistory: true);
    try {
      final messages = await _repository.history(key.type, key.lessonId);
      if (!mounted) return;
      state = state.copyWith(messages: messages, loadingHistory: false);
    } on Object {
      // History is a convenience; an empty conversation is still usable.
      if (mounted) state = state.copyWith(loadingHistory: false);
    }
  }

  /// Sends a quick action or a free-text question.
  Future<void> send({AiChip? chip, String text = ''}) async {
    final trimmed = text.trim();
    if (state.sending) return;
    if ((chip == null || chip.needsText) && trimmed.isEmpty) return;
    _pending = (
      chip: chip,
      text: chip == null || chip.needsText ? trimmed : '',
    );
    _analytics.logEvent('ai_chat_sent', {
      'chip': chip?.wire ?? 'free_text',
      'lesson_type': key.type.name,
      'lesson_id': key.lessonId,
    });
    await _run(_pending!);
  }

  Future<void> retry() async {
    final pending = _pending;
    if (pending == null || state.sending) return;
    await _run(pending);
  }

  Future<void> _run(_Request request) async {
    final userMessage = AiChatMessage(
      role: AiRole.user,
      content: request.text,
      chip: request.chip,
    );
    final before = state.messages;
    state = state.copyWith(
      messages: [...before, userMessage],
      streaming: () => '',
      sending: true,
      notice: () => null,
      error: () => null,
      retryable: false,
    );
    final buffer = StringBuffer();
    final done = Completer<void>();
    unawaited(_subscription?.cancel());
    _subscription = _repository
        .send(
          lessonType: key.type,
          lessonId: key.lessonId,
          chip: request.chip,
          message: request.text,
        )
        .listen(
          (event) {
            if (!mounted) return;
            switch (event) {
              case AiDelta(:final text):
                buffer.write(text);
                state = state.copyWith(streaming: buffer.toString);
              case AiDone(:final cached):
                // A cached answer arrives as one delta; replay it gradually
                // so it feels the same as a live answer.
                if (cached) {
                  _revealCached(buffer.toString(), userMessage, done);
                } else {
                  _finish(buffer.toString(), done);
                }
            }
          },
          onError: (Object error) {
            if (!mounted) return;
            _fail(error, before);
            if (!done.isCompleted) done.complete();
          },
          onDone: () {
            if (!done.isCompleted && !(_reveal?.isActive ?? false)) {
              done.complete();
            }
          },
          cancelOnError: true,
        );
    await done.future;
  }

  void _finish(String answer, Completer<void> done) {
    _pending = null;
    state = state.copyWith(
      messages: [
        ...state.messages,
        AiChatMessage(role: AiRole.assistant, content: answer),
      ],
      streaming: () => null,
      sending: false,
    );
    if (!done.isCompleted) done.complete();
  }

  void _revealCached(
    String answer,
    AiChatMessage userMessage,
    Completer<void> done,
  ) {
    _reveal?.cancel();
    const steps = 30;
    final chunk = math.max(1, (answer.length / steps).ceil());
    var shown = 0;
    state = state.copyWith(streaming: () => '');
    _reveal = Timer.periodic(const Duration(milliseconds: 30), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      shown = math.min(answer.length, shown + chunk);
      state = state.copyWith(streaming: () => answer.substring(0, shown));
      if (shown >= answer.length) {
        timer.cancel();
        _finish(answer, done);
      }
    });
  }

  void _fail(Object error, List<AiChatMessage> before) {
    final exception = error is AiChatException
        ? error
        : const AiChatException(
            AiChatException.network,
            'Butuh koneksi internet',
          );
    switch (exception.code) {
      case AiChatException.adRequired:
        _analytics.logEvent('ai_quota_exhausted', {
          'lesson_type': key.type.name,
        });
        state = state.copyWith(
          messages: before,
          streaming: () => null,
          sending: false,
        );
        unawaited(_offerAdOrUpsell());
      case AiChatException.dailyLimit:
        _analytics.logEvent('ai_quota_exhausted', {
          'lesson_type': key.type.name,
        });
        _showLimit(before, AiNotice.dailyLimit);
      case AiChatException.premiumLimit:
        _showLimit(before, AiNotice.premiumLimit);
      case AiChatException.premiumAdOffer:
        state = state.copyWith(
          messages: before,
          streaming: () => null,
          sending: false,
          messagesPerAd: exception.messagesPerAd,
        );
        unawaited(_offerLifetimeAd());
      default:
        _analytics.logEvent('ai_error', {'code': exception.code});
        state = state.copyWith(
          messages: before,
          streaming: () => null,
          sending: false,
          error: () => exception.message,
          retryable:
              exception.code == AiChatException.upstream ||
              exception.code == AiChatException.network,
        );
    }
  }

  void _showLimit(List<AiChatMessage> before, AiNotice notice) {
    state = state.copyWith(
      messages: before,
      streaming: () => null,
      sending: false,
      notice: () => notice,
    );
  }

  /// Without a usable rewarded placement (iOS, ads off, no consent) the
  /// free user is told the day is over and offered Lifetime instead.
  Future<void> _offerAdOrUpsell() async {
    AdPlacementConfig? placement;
    try {
      placement = await _ref.read(aiChatRewardedPlacementProvider.future);
    } on Object {
      placement = null;
    }
    if (!mounted) return;
    if (placement == null) {
      state = state.copyWith(notice: () => AiNotice.dailyLimit);
    } else {
      _analytics.logEvent('ai_ad_offered', {'lesson_type': key.type.name});
      state = state.copyWith(notice: () => AiNotice.adOffer);
    }
  }

  /// A Lifetime account used its daily messages. Without an ad placement
  /// (iOS, ads off, no consent) the day simply ends for it.
  Future<void> _offerLifetimeAd() async {
    AdPlacementConfig? placement;
    try {
      placement = await _ref.read(
        aiChatLifetimeRewardedPlacementProvider.future,
      );
    } on Object {
      placement = null;
    }
    if (!mounted) return;
    if (placement == null) {
      state = state.copyWith(notice: () => AiNotice.premiumLimit);
    } else {
      _analytics.logEvent('ai_ad_offered', {
        'lesson_type': key.type.name,
        'plan': 'lifetime',
      });
      state = state.copyWith(notice: () => AiNotice.lifetimeAdOffer);
    }
  }

  /// The user accepted the opt-in sheet: show the ad, record the credit and
  /// resend the held message. Returns false if nothing was sent.
  Future<bool> watchAdAndSend({bool lifetime = false}) async {
    final pending = _pending;
    if (pending == null) return false;
    final limit = lifetime ? AiNotice.premiumLimit : AiNotice.dailyLimit;
    final placement = await _ref.read(
      lifetime
          ? aiChatLifetimeRewardedPlacementProvider.future
          : aiChatRewardedPlacementProvider.future,
    );
    if (!mounted) return false;
    if (placement == null) {
      state = state.copyWith(notice: () => limit);
      return false;
    }
    state = state.copyWith(notice: () => null);
    final outcome = await _ref
        .read(rewardedAdShowerProvider)
        .show(placement.unitId);
    if (outcome != RewardedAdOutcome.unavailable) {
      // A full-screen ad just played; restart the interstitial count.
      _ref.read(interstitialControllerProvider).markFullScreenAdShown();
    }
    if (!mounted) return false;
    if (outcome == RewardedAdOutcome.dismissed) {
      _analytics.logEvent('ai_ad_declined', {'stage': 'closed_early'});
      return false;
    }
    if (lifetime && outcome == RewardedAdOutcome.unavailable) {
      // Several messages per ad are only paid for a watched ad.
      state = state.copyWith(
        error: () => 'Iklan belum tersedia. Coba lagi sebentar lagi.',
        retryable: true,
      );
      return false;
    }
    _analytics.logEvent('ai_ad_watched', {
      'result': outcome == RewardedAdOutcome.watched
          ? 'watched'
          : 'unavailable',
    });
    try {
      await _repository.grantAdCredit(
        watched: outcome == RewardedAdOutcome.watched,
      );
    } on AiChatException catch (error) {
      if (!mounted) return false;
      if (error.code == AiChatException.dailyLimit) {
        state = state.copyWith(notice: () => limit);
      } else {
        state = state.copyWith(error: () => error.message, retryable: false);
      }
      return false;
    }
    await _run(pending);
    return true;
  }

  void declineAd() {
    _analytics.logEvent('ai_ad_declined', {'stage': 'sheet'});
    dismissNotice();
  }

  void dismissNotice() => state = state.copyWith(notice: () => null);

  void logPaywallView() =>
      _analytics.logEvent('ai_paywall_view', {'lesson_type': key.type.name});

  Future<void> clearConversation() async {
    await _repository.clear(key.type, key.lessonId);
    if (mounted) {
      _pending = null;
      state = const AiChatState();
    }
  }

  @override
  void dispose() {
    _reveal?.cancel();
    unawaited(_subscription?.cancel());
    super.dispose();
  }
}

final aiChatControllerProvider = StateNotifierProvider.autoDispose
    .family<AiChatController, AiChatState, AiChatKey>(AiChatController.new);
