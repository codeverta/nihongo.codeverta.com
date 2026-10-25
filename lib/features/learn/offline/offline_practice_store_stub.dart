import 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';

export 'package:nihongo_master/features/learn/offline/offline_practice_store_base.dart';

class EncryptedOfflinePracticeLocalStore implements OfflinePracticeLocalStore {
  @override
  Future<List<Map<String, dynamic>>> read(String userId) => Future.error(
    UnsupportedError('Offline practice storage is unavailable.'),
  );

  @override
  Future<void> write(String userId, List<Map<String, dynamic>> attempts) =>
      Future.error(
        UnsupportedError('Offline practice storage is unavailable.'),
      );

  @override
  Future<void> clear(String userId) async {}
}
