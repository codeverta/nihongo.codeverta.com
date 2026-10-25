import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Link availability is only a hint; successful/failed API requests also update
/// [backendReachableProvider] because Wi-Fi may have no internet connection.
final connectivityProvider = StreamProvider<List<ConnectivityResult>>((
  ref,
) async* {
  final connectivity = Connectivity();
  try {
    yield await connectivity.checkConnectivity();
    yield* connectivity.onConnectivityChanged;
  } on Object {
    // An unavailable platform plugin must not prevent startup or network use.
    yield const [];
  }
});

final backendReachableProvider = StateProvider<bool?>((ref) => null);

final offlineProvider = Provider<bool>((ref) {
  final links = ref.watch(connectivityProvider).valueOrNull;
  return (links?.contains(ConnectivityResult.none) ?? false) ||
      ref.watch(backendReachableProvider) == false;
});

final unmeteredConnectionProvider = Provider<bool>((ref) {
  final links = ref.watch(connectivityProvider).valueOrNull ?? const [];
  return links.contains(ConnectivityResult.wifi) ||
      links.contains(ConnectivityResult.ethernet);
});
