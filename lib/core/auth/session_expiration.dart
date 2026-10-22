import 'package:flutter_riverpod/flutter_riverpod.dart';

const sessionExpiredMessage =
    'Sesi Anda berakhir. Silakan masuk kembali atau lanjut sebagai tamu.';

String? redirectToLoginIfSessionExpired(String currentPath, String? message) {
  if (message == sessionExpiredMessage && currentPath != '/login') {
    return '/login';
  }
  return null;
}

/// Counts invalid-session responses so any active auth controller can react.
final sessionExpirationEventProvider = StateProvider<int>((ref) => 0);
