import 'dart:typed_data';

abstract interface class LearningCachePersistence {
  Future<String?> readIndex();
  Future<void> writeIndex(String contents);
  Future<Uint8List?> readBody(String id);
  Future<void> writeBody(String id, Uint8List contents);
  Future<void> deleteBody(String id);
  Future<void> clear();
}
