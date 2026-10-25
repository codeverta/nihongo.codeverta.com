import 'dart:typed_data';

import 'package:nihongo_master/core/offline/learning_cache_persistence_base.dart';

LearningCachePersistence createLearningCachePersistence({
  String? directoryPath,
}) => MemoryLearningCachePersistence();

class MemoryLearningCachePersistence implements LearningCachePersistence {
  String? _index;
  final Map<String, Uint8List> _bodies = <String, Uint8List>{};

  @override
  Future<void> clear() async {
    _index = null;
    _bodies.clear();
  }

  @override
  Future<void> deleteBody(String id) async {
    _bodies.remove(id);
  }

  @override
  Future<Uint8List?> readBody(String id) async {
    final value = _bodies[id];
    return value == null ? null : Uint8List.fromList(value);
  }

  @override
  Future<String?> readIndex() async => _index;

  @override
  Future<void> writeBody(String id, Uint8List contents) async {
    _bodies[id] = Uint8List.fromList(contents);
  }

  @override
  Future<void> writeIndex(String contents) async {
    _index = contents;
  }
}
