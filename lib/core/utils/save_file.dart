import 'package:file_selector/file_selector.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart' hide XFile;

/// Hands a generated file to the user: a save dialog on desktop, the system
/// share sheet (Save to Files, Drive, chat apps…) everywhere else.
Future<void> saveFileToDevice(
  BuildContext context, {
  required Uint8List bytes,
  required String name,
  required String mimeType,
  required String extension,
  required String typeLabel,
}) async {
  final file = XFile.fromData(bytes, name: name, mimeType: mimeType);
  if (!kIsWeb &&
      const [
        TargetPlatform.windows,
        TargetPlatform.linux,
        TargetPlatform.macOS,
      ].contains(defaultTargetPlatform)) {
    final result = await getSaveLocation(
      suggestedName: name,
      acceptedTypeGroups: [
        XTypeGroup(label: typeLabel, extensions: [extension]),
      ],
    );
    if (result != null) await file.saveTo(result.path);
    return;
  }
  if (!context.mounted) return;
  final box = context.findRenderObject() as RenderBox?;
  await SharePlus.instance.share(
    ShareParams(
      files: [file],
      fileNameOverrides: [name],
      sharePositionOrigin: box == null
          ? null
          : box.localToGlobal(Offset.zero) & box.size,
    ),
  );
}
