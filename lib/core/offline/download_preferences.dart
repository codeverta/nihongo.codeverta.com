import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const allowLargeMobileDownloadsKey = 'offline_large_mobile_downloads_v1';

final allowLargeMobileDownloadsProvider =
    StateNotifierProvider<DownloadPreferences, bool>((ref) {
      return DownloadPreferences(ref.watch(sharedPreferencesProvider));
    });

class DownloadPreferences extends StateNotifier<bool> {
  DownloadPreferences(this.preferences)
    : super(preferences.getBool(allowLargeMobileDownloadsKey) ?? false);

  final SharedPreferences preferences;

  Future<void> setEnabled({required bool enabled}) async {
    await preferences.setBool(allowLargeMobileDownloadsKey, enabled);
    if (mounted) state = enabled;
  }
}
