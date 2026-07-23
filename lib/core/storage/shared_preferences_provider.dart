import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

const authTokenStorageKey = 'auth_session_token';

/// Di-override pada root [ProviderScope] setelah storage diinisialisasi.
final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw StateError('sharedPreferencesProvider belum diinisialisasi');
});
