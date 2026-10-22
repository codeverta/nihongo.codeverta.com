import 'package:flutter_riverpod/flutter_riverpod.dart';

const sessionExpiredMessage =
    'Sesi Anda berakhir. Silakan masuk kembali atau lanjut sebagai tamu.';

/// Counts invalid-session responses so any active auth controller can react.
final sessionExpirationEventProvider = StateProvider<int>((ref) => 0);
