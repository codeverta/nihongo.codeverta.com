import 'dart:io';
import 'dart:typed_data';

import 'package:nihongo_master/core/offline/learning_cache_persistence_base.dart';
import 'package:path_provider/path_provider.dart';

LearningCachePersistence createLearningCachePersistence({
  String? directoryPath,
}) => FileLearningCachePersistence(directoryPath: directoryPath);

class FileLearningCachePersistence implements LearningCachePersistence {
  FileLearningCachePersistence({this.directoryPath});

  final String? directoryPath;
  Future<Directory>? _directory;

  Future<Directory> get _root => _directory ??= _resolveDirectory();

  Future<Directory> _resolveDirectory() async {
    final path = directoryPath;
    final directory = path == null
        ? Directory(
            '${(await getApplicationSupportDirectory()).path}${Platform.pathSeparator}learning_cache',
          )
        : Directory(path);
    await directory.create(recursive: true);
    return directory;
  }

  Future<File> _indexFile() async =>
      File('${(await _root).path}${Platform.pathSeparator}index.json');

  Future<File> _bodyFile(String id) async =>
      File('${(await _root).path}${Platform.pathSeparator}$id.cache');

  @override
  Future<void> clear() async {
    final root = await _root;
    if (root.existsSync()) await root.delete(recursive: true);
    await root.create(recursive: true);
  }

  @override
  Future<void> deleteBody(String id) async {
    final file = await _bodyFile(id);
    if (file.existsSync()) await file.delete();
  }

  @override
  Future<Uint8List?> readBody(String id) async {
    final file = await _bodyFile(id);
    if (!file.existsSync()) return null;
    return file.readAsBytes();
  }

  @override
  Future<String?> readIndex() async {
    final file = await _indexFile();
    if (!file.existsSync()) return null;
    return file.readAsString();
  }

  @override
  Future<void> writeBody(String id, Uint8List contents) async {
    final file = await _bodyFile(id);
    final temporary = File('${file.path}.tmp');
    await temporary.writeAsBytes(contents, flush: true);
    if (file.existsSync()) await file.delete();
    await temporary.rename(file.path);
  }

  @override
  Future<void> writeIndex(String contents) async {
    final file = await _indexFile();
    final temporary = File('${file.path}.tmp');
    await temporary.writeAsString(contents, flush: true);
    if (file.existsSync()) await file.delete();
    await temporary.rename(file.path);
  }
}
