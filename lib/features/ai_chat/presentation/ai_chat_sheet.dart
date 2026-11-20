import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/offline/network_status.dart';
import 'package:nihongo_master/features/ai_chat/data/ai_chat_repository.dart';
import 'package:nihongo_master/features/ai_chat/presentation/ai_chat_controller.dart';
import 'package:nihongo_master/features/ai_chat/presentation/ai_markdown.dart';

/// Maximum characters per message; the server enforces the same limit.
const aiMaxMessageLength = 500;

/// App-bar action that opens the tutor for the current lesson.
class AskAiButton extends StatelessWidget {
  const AskAiButton({
    required this.lessonType,
    required this.lessonId,
    super.key,
  });

  final AiLessonType lessonType;
  final String lessonId;

  @override
  Widget build(BuildContext context) {
    return IconButton(
      key: const ValueKey('ask-ai-button'),
      tooltip: 'Tanya AI'.localized(context),
      icon: const Icon(Icons.auto_awesome_outlined),
      onPressed: lessonId.isEmpty
          ? null
          : () => showAiChatSheet(context, (
              type: lessonType,
              lessonId: lessonId,
            )),
    );
  }
}

Future<void> showAiChatSheet(BuildContext context, AiChatKey key) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    showDragHandle: true,
    builder: (context) => FractionallySizedBox(
      heightFactor: 0.92,
      child: AiChatSheet(chatKey: key),
    ),
  );
}

class AiChatSheet extends ConsumerStatefulWidget {
  const AiChatSheet({required this.chatKey, super.key});

  final AiChatKey chatKey;

  @override
  ConsumerState<AiChatSheet> createState() => _AiChatSheetState();
}

class _AiChatSheetState extends ConsumerState<AiChatSheet> {
  final _input = TextEditingController();
  final _scroll = ScrollController();
  bool _correctionMode = false;

  @override
  void dispose() {
    _input.dispose();
    _scroll.dispose();
    super.dispose();
  }

  AiChatController get _controller =>
      ref.read(aiChatControllerProvider(widget.chatKey).notifier);

  void _scrollToEnd() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.jumpTo(_scroll.position.maxScrollExtent);
      }
    });
  }

  Future<void> _sendText() async {
    final text = _input.text.trim();
    if (text.isEmpty) return;
    final chip = _correctionMode ? AiChip.correct : null;
    _input.clear();
    setState(() => _correctionMode = false);
    await _controller.send(chip: chip, text: text);
  }

  Future<void> _tapChip(AiChip chip) async {
    if (chip.needsText) {
      setState(() => _correctionMode = true);
      return;
    }
    await _controller.send(chip: chip);
  }

  Future<void> _presentNotice(AiNotice notice) async {
    switch (notice) {
      case AiNotice.adOffer:
        final accepted = await showModalBottomSheet<bool>(
          context: context,
          showDragHandle: true,
          builder: (context) => const _AdOfferSheet(),
        );
        if (!mounted) return;
        if (accepted ?? false) {
          await _controller.watchAdAndSend();
        } else {
          _controller.declineAd();
        }
      case AiNotice.dailyLimit:
        _controller
          ..dismissNotice()
          ..logPaywallView();
        final upgrade = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text('Kuota Tanya AI hari ini habis'.localized(context)),
            content: Text(
              'Coba lagi besok, atau buka Lifetime Access untuk lebih banyak pertanyaan setiap hari.'
                  .localized(context),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: Text('Nanti'.localized(context)),
              ),
              FilledButton(
                key: const ValueKey('ai-upsell-lifetime'),
                onPressed: () => Navigator.pop(context, true),
                child: Text('Lihat Lifetime'.localized(context)),
              ),
            ],
          ),
        );
        if (upgrade ?? false) {
          if (!mounted) return;
          unawaited(context.push('/premium'));
        }
      case AiNotice.lifetimeAdOffer:
        final messages = ref
            .read(aiChatControllerProvider(widget.chatKey))
            .messagesPerAd;
        final accepted = await showModalBottomSheet<bool>(
          context: context,
          showDragHandle: true,
          builder: (context) => _AdOfferSheet(lifetimeMessages: messages),
        );
        if (!mounted) return;
        if (accepted ?? false) {
          await _controller.watchAdAndSend(lifetime: true);
        } else {
          _controller.declineAd();
        }
      case AiNotice.premiumLimit:
        _controller.dismissNotice();
        await showDialog<void>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text('Batas harian tercapai'.localized(context)),
            content: Text(
              'Batas harian tercapai, coba lagi besok.'.localized(context),
            ),
            actions: [
              FilledButton(
                onPressed: () => Navigator.pop(context),
                child: Text('OK'.localized(context)),
              ),
            ],
          ),
        );
    }
  }

  Future<void> _confirmClear() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Hapus percakapan?'.localized(context)),
        content: Text(
          'Percakapan di pelajaran ini akan hilang dari aplikasi.'.localized(
            context,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Batal'.localized(context)),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text('Hapus'.localized(context)),
          ),
        ],
      ),
    );
    if (confirmed ?? false) {
      try {
        await _controller.clearConversation();
      } on Object {
        if (mounted) {
          ScaffoldMessenger.maybeOf(context)?.showSnackBar(
            SnackBar(
              content: Text('Gagal menghapus percakapan.'.localized(context)),
            ),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final provider = aiChatControllerProvider(widget.chatKey);
    final state = ref.watch(provider);
    final offline = ref.watch(offlineProvider);
    ref.listen(provider, (previous, next) {
      if (next.messages.length != previous?.messages.length ||
          next.streaming != previous?.streaming) {
        _scrollToEnd();
      }
      final notice = next.notice;
      if (notice != null && notice != previous?.notice) {
        unawaited(_presentNotice(notice));
      }
    });
    final theme = Theme.of(context);
    final canSend = !offline && !state.sending;

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 8, 8),
          child: Row(
            children: [
              Icon(Icons.auto_awesome, color: theme.colorScheme.primary),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  'Tanya AI'.localized(context),
                  style: theme.textTheme.titleLarge,
                ),
              ),
              PopupMenuButton<String>(
                tooltip: 'Opsi'.localized(context),
                enabled: state.messages.isNotEmpty && !state.sending,
                onSelected: (_) => _confirmClear(),
                itemBuilder: (context) => [
                  PopupMenuItem(
                    value: 'clear',
                    child: Text('Hapus percakapan'.localized(context)),
                  ),
                ],
              ),
            ],
          ),
        ),
        if (offline)
          MaterialBanner(
            key: const ValueKey('ai-offline'),
            content: Text('Butuh koneksi internet'.localized(context)),
            leading: const Icon(Icons.wifi_off),
            actions: const [SizedBox.shrink()],
          ),
        Expanded(
          child: _Messages(
            state: state,
            scroll: _scroll,
            onRetry: _controller.retry,
          ),
        ),
        _ChipRow(enabled: canSend, onTap: _tapChip),
        _Composer(
          controller: _input,
          enabled: canSend,
          sending: state.sending,
          correctionMode: _correctionMode,
          onCancelCorrection: () => setState(() => _correctionMode = false),
          onSend: _sendText,
        ),
      ],
    );
  }
}

class _Messages extends StatelessWidget {
  const _Messages({
    required this.state,
    required this.scroll,
    required this.onRetry,
  });

  final AiChatState state;
  final ScrollController scroll;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    if (state.loadingHistory && state.messages.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    final streaming = state.streaming;
    final items = <Widget>[
      for (final message in state.messages) _Bubble(message: message),
      if (streaming != null)
        _Bubble(
          message: AiChatMessage(role: AiRole.assistant, content: streaming),
          typing: streaming.isEmpty,
        ),
      if (state.error != null)
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Column(
            children: [
              Text(
                state.error!.localized(context),
                textAlign: TextAlign.center,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
              if (state.retryable)
                TextButton(
                  onPressed: onRetry,
                  child: Text('Coba lagi'.localized(context)),
                ),
            ],
          ),
        ),
    ];
    if (items.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Text(
            'Tanya apa saja tentang pelajaran ini, atau pilih aksi cepat di bawah.'
                .localized(context),
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodyLarge,
          ),
        ),
      );
    }
    return SelectionArea(
      child: ListView(
        controller: scroll,
        padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
        children: items,
      ),
    );
  }
}

class _Bubble extends StatelessWidget {
  const _Bubble({required this.message, this.typing = false});

  final AiChatMessage message;
  final bool typing;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final user = message.role == AiRole.user;
    final chip = message.chip;
    final Widget content;
    if (typing) {
      content = const SizedBox(
        width: 28,
        height: 16,
        child: LinearProgressIndicator(),
      );
    } else if (user) {
      final label = chip?.label.localized(context);
      final text = message.content;
      content = Text(
        label == null ? text : (text.isEmpty ? label : '$label:\n$text'),
        style: TextStyle(color: scheme.onPrimaryContainer),
      );
    } else {
      content = AiMarkdown(message.content);
    }
    return Align(
      alignment: user ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: BoxConstraints(
          maxWidth: MediaQuery.sizeOf(context).width * 0.85,
        ),
        margin: const EdgeInsets.symmetric(vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: user ? scheme.primaryContainer : scheme.surfaceContainerHigh,
          borderRadius: BorderRadius.circular(16),
        ),
        child: content,
      ),
    );
  }
}

class _ChipRow extends StatelessWidget {
  const _ChipRow({required this.enabled, required this.onTap});

  final bool enabled;
  final ValueChanged<AiChip> onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 48,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: AiChip.values.length,
        separatorBuilder: (_, _) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final chip = AiChip.values[index];
          return ActionChip(
            key: ValueKey('ai-chip-${chip.wire}'),
            label: Text(chip.label.localized(context)),
            onPressed: enabled ? () => onTap(chip) : null,
          );
        },
      ),
    );
  }
}

class _Composer extends StatelessWidget {
  const _Composer({
    required this.controller,
    required this.enabled,
    required this.sending,
    required this.correctionMode,
    required this.onCancelCorrection,
    required this.onSend,
  });

  final TextEditingController controller;
  final bool enabled;
  final bool sending;
  final bool correctionMode;
  final VoidCallback onCancelCorrection;
  final VoidCallback onSend;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(
        12,
        4,
        8,
        12 + MediaQuery.viewInsetsOf(context).bottom,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (correctionMode)
            Padding(
              padding: const EdgeInsets.only(left: 4, bottom: 4),
              child: InputChip(
                label: Text(AiChip.correct.label.localized(context)),
                onDeleted: onCancelCorrection,
              ),
            ),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(
                child: TextField(
                  key: const ValueKey('ai-input'),
                  controller: controller,
                  enabled: enabled,
                  minLines: 1,
                  maxLines: 4,
                  textInputAction: TextInputAction.newline,
                  inputFormatters: [
                    LengthLimitingTextInputFormatter(aiMaxMessageLength),
                  ],
                  decoration: InputDecoration(
                    hintText:
                        (correctionMode
                                ? 'Tulis kalimat Jepang yang mau dikoreksi'
                                : 'Tulis pertanyaanmu…')
                            .localized(context),
                    border: const OutlineInputBorder(
                      borderRadius: BorderRadius.all(Radius.circular(24)),
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 10,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 4),
              IconButton.filled(
                key: const ValueKey('ai-send'),
                tooltip: 'Kirim'.localized(context),
                onPressed: enabled ? onSend : null,
                icon: sending
                    ? const SizedBox.square(
                        dimension: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.send),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _AdOfferSheet extends StatelessWidget {
  const _AdOfferSheet({this.lifetimeMessages = 0});

  /// For a Lifetime account: the messages one ad adds. Zero for free plans.
  final int lifetimeMessages;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 0, 24, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Icon(
              Icons.ondemand_video,
              size: 40,
              color: theme.colorScheme.primary,
            ),
            const SizedBox(height: 12),
            Text(
              lifetimeMessages > 0
                  ? 'Pesan Tanya AI hari ini sudah terpakai'.localized(context)
                  : 'Kuota gratis hari ini sudah habis'.localized(context),
              style: theme.textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              lifetimeMessages > 0
                  ? '${'Tonton iklan singkat untuk menambah'.localized(context)} $lifetimeMessages ${'pesan. Pertanyaan ini langsung dikirim.'.localized(context)}'
                  : 'Tonton iklan singkat untuk mengirim pertanyaan ini.'
                        .localized(context),
              key: const ValueKey('ai-ad-offer-text'),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 20),
            FilledButton(
              key: const ValueKey('ai-ad-watch'),
              onPressed: () => Navigator.pop(context, true),
              child: Text('Tonton'.localized(context)),
            ),
            TextButton(
              key: const ValueKey('ai-ad-cancel'),
              onPressed: () => Navigator.pop(context, false),
              child: Text('Batal'.localized(context)),
            ),
          ],
        ),
      ),
    );
  }
}
