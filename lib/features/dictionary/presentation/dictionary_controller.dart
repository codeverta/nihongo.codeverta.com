import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/dictionary/data/dictionary_repository.dart';
import 'package:nihongo_master/features/dictionary/domain/dictionary_models.dart';
import 'package:shared_preferences/shared_preferences.dart';

class DictionaryState {
  const DictionaryState({
    this.query = '',
    this.mode = 'auto',
    this.entries = const [],
    this.loading = false,
    this.loadingMore = false,
    this.hasMore = false,
    this.error,
    this.page = 1,
    this.jlpt,
    this.common = false,
  });
  final String query;
  final String mode;
  final List<DictionaryEntry> entries;
  final bool loading;
  final bool loadingMore;
  final bool hasMore;
  final String? error;
  final int page;
  final String? jlpt;
  final bool common;
}

final dictionaryControllerProvider =
    StateNotifierProvider.autoDispose<DictionaryController, DictionaryState>((
      ref,
    ) {
      return DictionaryController(ref.watch(dictionaryRepositoryProvider));
    });

class DictionaryController extends StateNotifier<DictionaryState> {
  DictionaryController(this.repository) : super(const DictionaryState());
  final DictionaryRepository repository;
  int _version = 0;

  Future<void> search(
    String text, {
    String? jlpt,
    bool? common,
    String? mode,
  }) async {
    final searchMode = mode ?? state.mode;
    final query = text.trim();
    final version = ++_version;
    final filter = jlpt ?? state.jlpt;
    final commonFilter = common ?? state.common;
    state = DictionaryState(
      query: query,
      mode: searchMode,
      jlpt: filter,
      common: commonFilter,
    );
    if (query.isEmpty) return;
    if (query.runes.length > 100) {
      state = DictionaryState(
        query: query,
        mode: searchMode,
        jlpt: filter,
        common: commonFilter,
        error: 'Pencarian maksimal 100 karakter.',
      );
      return;
    }
    state = DictionaryState(
      query: query,
      mode: searchMode,
      jlpt: filter,
      common: commonFilter,
      loading: true,
    );
    try {
      final result = await repository.search(
        query,
        jlpt: filter,
        common: commonFilter,
        mode: searchMode,
      );
      if (!mounted || version != _version) return;
      state = DictionaryState(
        query: query,
        mode: searchMode,
        jlpt: filter,
        common: commonFilter,
        entries: result.entries,
        hasMore: result.hasMore,
      );
    } on Object catch (error) {
      if (!mounted || version != _version) return;
      state = DictionaryState(
        query: query,
        mode: searchMode,
        jlpt: filter,
        common: commonFilter,
        error: error is DioException && error.response?.statusCode == 400
            ? 'Pencarian tidak valid. Coba gunakan kana jika ejaan romaji belum dikenali pada mode Jepang.'
            : 'Kamus belum dapat diakses. Periksa koneksi atau coba lagi sebentar.',
      );
    }
  }

  Future<void> setMode(String mode) => search(state.query, mode: mode);

  Future<void> filter(
    String? jlpt, {
    required bool common,
    String? mode,
    String? text,
  }) async {
    final query = text ?? state.query;
    state = DictionaryState(
      query: query,
      jlpt: jlpt,
      common: common,
      mode: mode ?? state.mode,
    );
    await search(query, common: common);
  }

  Future<void> more() async {
    if (state.loading || state.loadingMore || !state.hasMore) return;
    final old = state;
    final version = _version;
    state = DictionaryState(
      query: old.query,
      mode: old.mode,
      entries: old.entries,
      page: old.page,
      hasMore: true,
      jlpt: old.jlpt,
      common: old.common,
      loadingMore: true,
    );
    try {
      final result = await repository.search(
        old.query,
        page: old.page + 1,
        jlpt: old.jlpt,
        common: old.common,
        mode: old.mode,
      );
      if (!mounted || version != _version) return;
      final entries = [...old.entries];
      final ids = entries.map((e) => e.id).toSet();
      entries.addAll(result.entries.where((e) => ids.add(e.id)));
      state = DictionaryState(
        query: old.query,
        mode: old.mode,
        entries: entries,
        page: old.page + 1,
        hasMore: result.hasMore,
        jlpt: old.jlpt,
        common: old.common,
      );
    } on Object {
      if (!mounted || version != _version) return;
      state = DictionaryState(
        query: old.query,
        mode: old.mode,
        entries: old.entries,
        page: old.page,
        hasMore: old.hasMore,
        jlpt: old.jlpt,
        common: old.common,
        error: 'Halaman berikutnya gagal dimuat. Coba lagi.',
      );
    }
  }
}

class DictionaryLibrary {
  const DictionaryLibrary({this.saved = const [], this.recent = const []});
  final List<DictionaryEntry> saved;
  final List<String> recent;
}

final dictionaryLibraryProvider =
    StateNotifierProvider<DictionaryLibraryController, DictionaryLibrary>((
      ref,
    ) {
      final id = ref.watch(
        authControllerProvider.select((state) => state.user?.id),
      );
      return DictionaryLibraryController(
        ref.watch(sharedPreferencesProvider),
        id ?? 'guest',
      );
    });

class DictionaryLibraryController extends StateNotifier<DictionaryLibrary> {
  DictionaryLibraryController(this.preferences, this.owner)
    : super(const DictionaryLibrary()) {
    try {
      final data =
          jsonDecode(preferences.getString(_key) ?? '{}')
              as Map<String, dynamic>;
      state = DictionaryLibrary(
        saved: (data['saved'] as List? ?? [])
            .whereType<Map<String, dynamic>>()
            .map(DictionaryEntry.new)
            .take(60)
            .toList(),
        recent: dictionaryStrings(data['recent']).take(20).toList(),
      );
    } on Object {
      // Corrupt local data must not prevent dictionary searches.
    }
  }
  final SharedPreferences preferences;
  final String owner;
  String get _key => 'dictionary_library_v1_$owner';
  Future<void> _queue = Future<void>.value();

  Future<bool> toggle(DictionaryEntry entry) => _mutate(() {
    final saved = [...state.saved];
    if (saved.any((e) => e.id == entry.id)) {
      saved.removeWhere((e) => e.id == entry.id);
    } else {
      if (saved.length >= 60) return null;
      saved.insert(0, entry);
    }
    return DictionaryLibrary(saved: saved, recent: state.recent);
  });

  Future<bool> remember(String query) async {
    final normalized = query.trim();
    if (normalized.isEmpty || normalized.runes.length > 100) return false;
    return _mutate(
      () => DictionaryLibrary(
        saved: state.saved,
        recent: [
          normalized,
          ...state.recent.where((q) => q != normalized),
        ].take(20).toList(),
      ),
    );
  }

  Future<bool> clearHistory() =>
      _mutate(() => DictionaryLibrary(saved: state.saved));

  Future<bool> _mutate(DictionaryLibrary? Function() update) {
    final result = _queue.then((_) async {
      if (!mounted) return false;
      final next = update();
      if (next == null) return false;
      return _persist(next);
    });
    _queue = result.then((_) {});
    return result;
  }

  Future<bool> _persist(DictionaryLibrary next) async {
    try {
      final ok = await preferences.setString(
        _key,
        jsonEncode({
          'saved': next.saved.map((e) => e.json).toList(),
          'recent': next.recent,
        }),
      );
      if (ok && mounted) state = next;
      return ok;
    } on Object {
      return false;
    }
  }
}
