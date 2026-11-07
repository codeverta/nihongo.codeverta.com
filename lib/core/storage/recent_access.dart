import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

@immutable
class RecentEntry {
  const RecentEntry({required this.lastMs, required this.count});

  final int lastMs;
  final int count;
}

/// Remembers, per list on this device, which items were opened most recently
/// and how often. Used to float recently used items to the top of lists.
class RecentAccess extends StateNotifier<Map<String, RecentEntry>> {
  RecentAccess(this._preferences, this.scope) : super(const {}) {
    state = _read();
  }

  static const _maxEntries = 200;

  final SharedPreferences _preferences;
  final String scope;

  String get _key => 'recent_access_v1_$scope';

  Map<String, RecentEntry> _read() {
    try {
      final raw = _preferences.getString(_key);
      if (raw == null) return const {};
      final decoded = jsonDecode(raw);
      if (decoded is! Map<String, dynamic>) return const {};
      final result = <String, RecentEntry>{};
      decoded.forEach((id, value) {
        if (value is List && value.length == 2) {
          final last = value[0];
          final count = value[1];
          if (last is int && count is int) {
            result[id] = RecentEntry(lastMs: last, count: count);
          }
        }
      });
      return result;
    } on Object {
      return const {};
    }
  }

  void record(String id, {DateTime? now}) {
    if (id.isEmpty) return;
    final next = {...state};
    final previous = next[id];
    next[id] = RecentEntry(
      lastMs: (now ?? DateTime.now()).millisecondsSinceEpoch,
      count: (previous?.count ?? 0) + 1,
    );
    if (next.length > _maxEntries) {
      final oldest = next.entries.toList()
        ..sort((a, b) => a.value.lastMs.compareTo(b.value.lastMs));
      for (final entry in oldest.take(next.length - _maxEntries)) {
        next.remove(entry.key);
      }
    }
    state = next;
    try {
      _preferences.setString(
        _key,
        jsonEncode({
          for (final e in next.entries) e.key: [e.value.lastMs, e.value.count],
        }),
      );
    } on Object {
      // Ordering is a convenience; ignore storage failures.
    }
  }
}

final recentAccessProvider =
    StateNotifierProvider.family<RecentAccess, Map<String, RecentEntry>, String>(
      (ref, scope) => RecentAccess(ref.watch(sharedPreferencesProvider), scope),
    );

/// Negative when [a] should come before [b]: recently opened items first
/// (latest first, then most opened). Returns 0 when neither was opened so the
/// caller's own order decides.
int compareRecent(Map<String, RecentEntry> recent, String a, String b) {
  final ra = recent[a];
  final rb = recent[b];
  if (ra == null && rb == null) return 0;
  if (ra == null) return 1;
  if (rb == null) return -1;
  final byLast = rb.lastMs.compareTo(ra.lastMs);
  return byLast != 0 ? byLast : rb.count.compareTo(ra.count);
}
