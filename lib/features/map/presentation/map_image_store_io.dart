import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';

const mapImagesUseDisk = true;

String _fileName(String url) {
  final segments = Uri.parse(url).pathSegments;
  final last = segments.isEmpty ? '' : segments.last;
  // Backend photos are content-addressed (sha256.jpg), so the name is stable.
  if (RegExp(r'^[0-9a-f]{64}\.jpg$').hasMatch(last)) return last;
  var hash = 0;
  for (final unit in url.codeUnits) {
    hash = (hash * 31 + unit) & 0x7fffffff;
  }
  return 'u$hash-${url.length}.jpg';
}

/// Returns the photo from disk, downloading it once when missing. Returns
/// null when it is neither stored nor downloadable (for example offline).
Future<Uint8List?> cachedMapImage(String url) async {
  try {
    final directory = Directory(
      '${(await getApplicationSupportDirectory()).path}/map-images',
    );
    final file = File('${directory.path}/${_fileName(url)}');
    if (file.existsSync() && file.lengthSync() > 0) {
      return file.readAsBytes();
    }
    await directory.create(recursive: true);
    final response = await Dio().get<List<int>>(
      url,
      options: Options(
        responseType: ResponseType.bytes,
        receiveTimeout: const Duration(seconds: 20),
      ),
    );
    final bytes = response.data;
    if (bytes == null || bytes.isEmpty) return null;
    final data = Uint8List.fromList(bytes);
    final temp = File('${file.path}.part');
    await temp.writeAsBytes(data, flush: true);
    await temp.rename(file.path);
    return data;
  } on Object {
    return null;
  }
}
