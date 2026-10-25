import 'dart:async';
import 'dart:math';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

final ankiStudyProvider = StateNotifierProvider.autoDispose
    .family<AnkiStudyController, AnkiStudyState, String>((ref, id) {
      ref.watch(
        authControllerProvider.select((s) => (s.user?.id, s.user?.isPremium)),
      );
      final controller = AnkiStudyController(
        ref.watch(ankiRepositoryProvider),
        id,
        onSaved: () => ref.invalidate(ankiDeckProvider(id)),
      );
      unawaited(controller.load());
      return controller;
    });

class AnkiStudyState {
  const AnkiStudyState({
    this.session,
    this.position = 0,
    this.revealed = false,
    this.loading = false,
    this.saving = false,
    this.error,
    this.pendingRating,
    this.requestId,
  });
  final AnkiSession? session;
  final int position;
  final bool revealed;
  final bool loading;
  final bool saving;
  final AnkiFailure? error;
  final String? pendingRating;
  final String? requestId;
  AnkiCard? get current => session != null && position < session!.cards.length
      ? session!.cards[position]
      : null;
}

class AnkiStudyController extends StateNotifier<AnkiStudyState> {
  AnkiStudyController(this.repository, this.deckId, {this.onSaved})
    : super(const AnkiStudyState());
  final AnkiRepository repository;
  final String deckId;
  final void Function()? onSaved;
  Future<void> load() async {
    if (state.loading || state.saving) return;
    state = const AnkiStudyState(loading: true);
    try {
      final session = await repository.study(deckId);
      if (mounted) state = AnkiStudyState(session: session);
    } on AnkiFailure catch (error) {
      if (mounted) state = AnkiStudyState(error: error);
    } on Object {
      if (mounted) {
        state = const AnkiStudyState(
          error: AnkiFailure('Materi belum dapat dimuat. Coba kembali.', null),
        );
      }
    }
  }

  void reveal() {
    if (state.current == null || state.saving || state.revealed) return;
    state = AnkiStudyState(
      session: state.session,
      position: state.position,
      revealed: true,
    );
  }

  Future<void> rate(String rating) async {
    if (state.saving || !state.revealed || state.current == null) return;
    if (!const ['again', 'hard', 'good', 'easy'].contains(rating)) return;
    // Preserve the exact operation after a timeout; its server outcome may be unknown.
    final selected = state.pendingRating ?? rating;
    final id = state.requestId ?? _requestId();
    state = AnkiStudyState(
      session: state.session,
      position: state.position,
      revealed: true,
      saving: true,
      pendingRating: selected,
      requestId: id,
    );
    try {
      await repository.review(deckId, state.current!, selected, id);
      if (mounted) onSaved?.call();
      if (mounted) {
        state = AnkiStudyState(
          session: state.session,
          position: state.position + 1,
        );
      }
    } on AnkiFailure catch (error) {
      if (mounted) {
        state = AnkiStudyState(
          session: state.session,
          position: state.position,
          revealed: true,
          error: error,
          pendingRating: selected,
          requestId: id,
        );
      }
    } on Object {
      if (mounted) {
        state = AnkiStudyState(
          session: state.session,
          position: state.position,
          revealed: true,
          error: const AnkiFailure(
            'Review belum terkonfirmasi. Coba simpan kembali.',
            null,
          ),
          pendingRating: selected,
          requestId: id,
        );
      }
    }
  }
}

String _requestId() {
  final random = Random.secure();
  final bytes = List.generate(16, (_) => random.nextInt(256));
  bytes[6] = (bytes[6] & 15) | 64;
  bytes[8] = (bytes[8] & 63) | 128;
  final hex = bytes.map((v) => v.toRadixString(16).padLeft(2, '0')).join();
  return '${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}';
}
