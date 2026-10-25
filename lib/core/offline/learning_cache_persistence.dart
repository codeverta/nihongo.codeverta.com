import 'package:nihongo_master/core/offline/learning_cache_persistence_base.dart';
import 'package:nihongo_master/core/offline/learning_cache_persistence_memory.dart'
    if (dart.library.io) 'learning_cache_persistence_io.dart'
    as platform;

export 'learning_cache_persistence_base.dart';

LearningCachePersistence createLearningCachePersistence({
  String? directoryPath,
}) => platform.createLearningCachePersistence(directoryPath: directoryPath);
