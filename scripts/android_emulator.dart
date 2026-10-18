import 'dart:convert';
import 'dart:io';

/// Selects a connected Android device or boots an AVD when none is connected.
/// An explicit non-Android target (for example `-d chrome`) is left alone.
Future<List<String>> prepareAndroidDeviceArgs({
  required Directory projectRoot,
  required List<String> flutterArgs,
}) async {
  final requestedId = _requestedDeviceId(flutterArgs);
  final devices = await _flutterMachineList(
    ['devices', '--machine'],
    projectRoot,
    'daftar perangkat Flutter',
  );
  final androidDevices = devices.where(_isAndroidDevice).toList();
  Map<String, dynamic>? requestedDevice;
  if (requestedId != null) {
    for (final device in devices) {
      if (device['id'] == requestedId) {
        requestedDevice = device;
        break;
      }
    }
  }

  if (requestedDevice != null && !_isAndroidDevice(requestedDevice)) {
    return flutterArgs;
  }
  if (requestedDevice != null) return flutterArgs;
  if (requestedId == null && androidDevices.isNotEmpty) {
    return _setDeviceId(flutterArgs, '${androidDevices.first['id']}');
  }

  final emulators = await _flutterMachineList(
    ['emulators', '--machine'],
    projectRoot,
    'daftar Android emulator',
  );
  final selectedEmulator = _selectEmulator(emulators, requestedId);
  if (selectedEmulator == null) {
    if (requestedId != null && !_looksLikeAndroidDeviceId(requestedId)) {
      return flutterArgs;
    }
    throw const ProcessException(
      'flutter',
      [],
      'Tidak ada AVD Android. Buat emulator terlebih dahulu di Android Studio Device Manager.',
    );
  }

  final sdkRoot = _findAndroidSdk(projectRoot);
  await _stopOfflineEmulators(sdkRoot);
  stdout.writeln(
    'Emulator Android belum aktif. Menyalakan ${selectedEmulator['id']} ...',
  );
  await _launchEmulator(selectedEmulator['id'] as String, sdkRoot, projectRoot);

  final deadline = DateTime.now().add(const Duration(minutes: 3));
  while (DateTime.now().isBefore(deadline)) {
    await Future<void>.delayed(const Duration(seconds: 2));
    final currentDevices = await _flutterMachineList(
      ['devices', '--machine'],
      projectRoot,
      'daftar perangkat Flutter',
    );
    final currentAndroidDevices = currentDevices.where(_isAndroidDevice);
    if (currentAndroidDevices.isNotEmpty) {
      final device = currentAndroidDevices.firstWhere(
        (candidate) => candidate['id'] == requestedId,
        orElse: () => currentAndroidDevices.first,
      );
      stdout.writeln('Emulator Android siap: ${device['id']}');
      if (requestedId == null || requestedId == selectedEmulator['id']) {
        return _setDeviceId(flutterArgs, '${device['id']}');
      }
      if (requestedId.startsWith('emulator-') && device['id'] != requestedId) {
        return _setDeviceId(flutterArgs, '${device['id']}');
      }
      return flutterArgs;
    }
  }

  throw ProcessException(
    'emulator',
    ['-avd', '${selectedEmulator['id']}'],
    'Emulator belum siap setelah 3 menit. Coba Cold Boot Now dari Android Studio Device Manager.',
  );
}

Map<String, dynamic>? _selectEmulator(
  List<Map<String, dynamic>> emulators,
  String? requestedId,
) {
  if (emulators.isEmpty) return null;
  if (requestedId != null) {
    for (final emulator in emulators) {
      if (emulator['id'] == requestedId) return emulator;
    }
    if (requestedId.startsWith('emulator-') && emulators.length == 1) {
      return emulators.first;
    }
    if (requestedId.startsWith('emulator-')) {
      throw ProcessException(
        'flutter',
        ['emulators', '--machine'],
        'ID $requestedId belum aktif dan ada beberapa AVD. Jalankan dengan ID AVD dari `flutter emulators`.',
      );
    }
    return null;
  }
  return emulators.first;
}

Future<List<Map<String, dynamic>>> _flutterMachineList(
  List<String> args,
  Directory workingDirectory,
  String description,
) async {
  final result = await Process.run(
    'flutter',
    args,
    workingDirectory: workingDirectory.path,
    runInShell: Platform.isWindows,
  );
  if (result.exitCode != 0) {
    throw ProcessException(
      'flutter',
      args,
      'Tidak bisa membaca $description: ${result.stderr}'.trim(),
      result.exitCode,
    );
  }
  try {
    final decoded = jsonDecode('${result.stdout}') as List<dynamic>;
    return decoded.whereType<Map<String, dynamic>>().toList();
  } on FormatException catch (error) {
    throw ProcessException(
      'flutter',
      args,
      'Output $description tidak valid: $error',
      result.exitCode,
    );
  }
}

bool _isAndroidDevice(Map<String, dynamic> device) =>
    '${device['targetPlatform']}'.startsWith('android-');

bool _looksLikeAndroidDeviceId(String id) =>
    id.startsWith('emulator-') ||
    id.startsWith('Pixel_') ||
    id.startsWith('pixel_');

String? _requestedDeviceId(List<String> args) {
  for (var index = 0; index < args.length; index++) {
    final argument = args[index];
    if ((argument == '-d' || argument == '--device-id') &&
        index + 1 < args.length) {
      return args[index + 1];
    }
    if (argument.startsWith('--device-id=')) {
      return argument.substring('--device-id='.length);
    }
  }
  return null;
}

List<String> _setDeviceId(List<String> args, String deviceId) {
  final result = List<String>.of(args);
  for (var index = 0; index < result.length; index++) {
    if ((result[index] == '-d' || result[index] == '--device-id') &&
        index + 1 < result.length) {
      result[index + 1] = deviceId;
      return result;
    }
    if (result[index].startsWith('--device-id=')) {
      result[index] = '--device-id=$deviceId';
      return result;
    }
  }
  result.addAll(['--device-id', deviceId]);
  return result;
}

String? _findAndroidSdk(Directory projectRoot) {
  final candidates = <String?>[
    Platform.environment['ANDROID_SDK_ROOT'],
    Platform.environment['ANDROID_HOME'],
    _sdkFromLocalProperties(projectRoot),
    if (Platform.isWindows && Platform.environment['LOCALAPPDATA'] != null)
      '${Platform.environment['LOCALAPPDATA']}\\Android\\Sdk',
    if (Platform.isMacOS && Platform.environment['HOME'] != null)
      '${Platform.environment['HOME']}/Library/Android/sdk',
    if (Platform.isLinux && Platform.environment['HOME'] != null)
      '${Platform.environment['HOME']}/Android/Sdk',
  ];
  for (final candidate in candidates) {
    if (candidate == null || candidate.isEmpty) continue;
    final root = Directory(candidate);
    if (root.existsSync()) return root.path;
  }
  return null;
}

String? _sdkFromLocalProperties(Directory projectRoot) {
  final file = File(
    '${projectRoot.path}${Platform.pathSeparator}android'
    '${Platform.pathSeparator}local.properties',
  );
  if (!file.existsSync()) return null;
  for (final line in file.readAsLinesSync()) {
    if (!line.startsWith('sdk.dir=')) continue;
    return line.substring('sdk.dir='.length).replaceAll(r'\\', r'\');
  }
  return null;
}

String? _sdkTool(String? sdkRoot, String directory, String name) {
  if (sdkRoot == null) return null;
  final executable = Platform.isWindows ? '$name.exe' : name;
  final file = File(
    '$sdkRoot${Platform.pathSeparator}$directory'
    '${Platform.pathSeparator}$executable',
  );
  return file.existsSync() ? file.path : null;
}

Future<void> _stopOfflineEmulators(String? sdkRoot) async {
  final adb = _sdkTool(sdkRoot, 'platform-tools', 'adb');
  if (adb == null) return;
  final result = await Process.run(adb, ['devices']);
  if (result.exitCode != 0) return;
  final offlineIds = LineSplitter.split('${result.stdout}')
      .map((line) => RegExp(r'^(emulator-\d+)\s+offline\b').firstMatch(line))
      .whereType<RegExpMatch>()
      .map((match) => match.group(1)!)
      .toList();
  for (final id in offlineIds) {
    stdout.writeln('Menghentikan emulator Android yang macet ($id) ...');
    await Process.run(adb, ['-s', id, 'emu', 'kill']);
  }
}

Future<void> _launchEmulator(
  String emulatorId,
  String? sdkRoot,
  Directory projectRoot,
) async {
  final emulator = _sdkTool(sdkRoot, 'emulator', 'emulator');
  if (emulator != null) {
    try {
      await Process.start(
        emulator,
        ['-avd', emulatorId, '-no-snapshot-load'],
        workingDirectory: projectRoot.path,
        mode: ProcessStartMode.detached,
      );
      return;
    } on ProcessException {
      // Let Flutter resolve the SDK and launch the AVD as a fallback.
    }
  }

  final result = await Process.run(
    'flutter',
    ['emulators', '--launch', emulatorId],
    workingDirectory: projectRoot.path,
    runInShell: Platform.isWindows,
  );
  if (result.exitCode != 0) {
    throw ProcessException(
      'flutter',
      ['emulators', '--launch', emulatorId],
      '${result.stderr}'.trim(),
      result.exitCode,
    );
  }
}
