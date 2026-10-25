abstract interface class OfflinePracticeLocalStore {
  Future<List<Map<String, dynamic>>> read(String userId);
  Future<void> write(String userId, List<Map<String, dynamic>> attempts);
  Future<void> clear(String userId);
}
