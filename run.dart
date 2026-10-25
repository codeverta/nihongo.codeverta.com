import 'dart:async';
import 'dart:io';

import 'scripts/android_emulator.dart';

const _defaultApiBaseUrl = 'http://10.0.2.2:8080/v1';

Future<void> main(List<String> args) async {
  try {
    final options = _parseArguments(args);
    if (options.help) {
      _printUsage();
      return;
    }

    final projectRoot = File.fromUri(Platform.script).absolute.parent;
    final backendRoot = Directory(
      '${projectRoot.path}${Platform.pathSeparator}backend',
    );
    final env = await _readEnvFile(
      File('${backendRoot.path}${Platform.pathSeparator}.env'),
    );
    final googleServerClientId = _resolveGoogleClientId(env);
    final offlineAccessPublicKey = _resolveOfflineAccessPublicKey(env);
    final apiBaseUrl =
        options.apiBaseUrl ??
        Platform.environment['API_BASE_URL'] ??
        _defaultApiBaseUrl;
    if (Uri.tryParse(apiBaseUrl)?.hasAuthority != true) {
      throw const _LaunchException('API_BASE_URL tidak valid.');
    }
    final flutterDefines = <String>[
      '--dart-define=API_BASE_URL=$apiBaseUrl',
      '--dart-define=GOOGLE_SERVER_CLIENT_ID=$googleServerClientId',
      if (offlineAccessPublicKey.isNotEmpty)
        '--dart-define=OFFLINE_ACCESS_PUBLIC_KEY=$offlineAccessPublicKey',
    ];
    final flutterArgs = <String>[
      'run',
      ...flutterDefines,
      ...options.flutterArgs,
    ];
    final selectedFlutterArgs = options.backendOnly
        ? flutterArgs
        : <String>[
            'run',
            ...flutterDefines,
            ...await prepareAndroidDeviceArgs(
              projectRoot: projectRoot,
              flutterArgs: options.flutterArgs,
            ),
          ];

    Process? backend;
    Directory? backendBuildDirectory;
    try {
      if (!options.skipBackend) {
        final port = _backendPort(env);
        stdout.writeln('Membangun backend Go ...');
        backendBuildDirectory = await Directory.systemTemp.createTemp(
          'nihongo-master-',
        );
        final executableName = Platform.isWindows
            ? 'nihongo-api.exe'
            : 'nihongo-api';
        final backendExecutable = File(
          '${backendBuildDirectory.path}${Platform.pathSeparator}$executableName',
        );
        final build = await Process.start(
          'go',
          ['build', '-o', backendExecutable.path, './cmd/api'],
          workingDirectory: backendRoot.path,
          runInShell: Platform.isWindows,
          mode: ProcessStartMode.inheritStdio,
        );
        final buildExitCode = await build.exitCode;
        if (buildExitCode != 0) {
          throw _LaunchException(
            'Build backend gagal (exit code $buildExitCode).',
          );
        }

        stdout.writeln('Menjalankan backend di http://localhost:$port ...');
        final backendEnvironment = Map<String, String>.of(Platform.environment)
          ..['GOOGLE_CLIENT_ID'] = googleServerClientId;
        backend = await Process.start(
          backendExecutable.path,
          const [],
          workingDirectory: backendRoot.path,
          environment: backendEnvironment,
          mode: ProcessStartMode.inheritStdio,
        );
        await _waitForBackend(backend, port);
        stdout.writeln('Backend siap. Menjalankan Flutter ...');
      }

      if (options.backendOnly) {
        stdout.writeln('Backend aktif. Tekan Ctrl+C untuk berhenti.');
        exitCode = await backend!.exitCode;
        return;
      }

      final flutter = await Process.start(
        'flutter',
        selectedFlutterArgs,
        workingDirectory: projectRoot.path,
        runInShell: Platform.isWindows,
        mode: ProcessStartMode.inheritStdio,
      );
      exitCode = await flutter.exitCode;
    } finally {
      if (backend != null) {
        backend.kill();
        await backend.exitCode.timeout(
          const Duration(seconds: 5),
          onTimeout: () => -1,
        );
      }
      if (backendBuildDirectory != null && backendBuildDirectory.existsSync()) {
        await backendBuildDirectory.delete(recursive: true);
      }
    }
  } on _LaunchException catch (error) {
    stderr.writeln(error.message);
    exitCode = 1;
  } on ProcessException catch (error) {
    stderr
      ..writeln('Tidak bisa menjalankan ${error.executable}.')
      ..writeln(
        error.message.isNotEmpty
            ? error.message
            : 'Pastikan Flutter SDK dan Go sudah terpasang serta ada di PATH.',
      );
    exitCode = 1;
  } catch (error) {
    stderr.writeln('Gagal menjalankan aplikasi: $error');
    exitCode = 1;
  }
}

Future<Map<String, String>> _readEnvFile(File file) async {
  if (!file.existsSync()) {
    throw const _LaunchException(
      'File backend/.env tidak ditemukan. Buat file itu dari backend/.env.example lalu isi Google Web Client ID.',
    );
  }

  final values = <String, String>{};
  for (final rawLine in await file.readAsLines()) {
    final line = rawLine.trim();
    if (line.isEmpty || line.startsWith('#')) continue;
    final separator = line.indexOf('=');
    if (separator <= 0) continue;
    final key = line.substring(0, separator).trim();
    if (values.containsKey(key)) continue;
    var value = line.substring(separator + 1).trim();
    if (value.length >= 2 &&
        ((value.startsWith('"') && value.endsWith('"')) ||
            (value.startsWith("'") && value.endsWith("'")))) {
      value = value.substring(1, value.length - 1);
    }
    values[key] = value;
  }
  return values;
}

String _resolveGoogleClientId(Map<String, String> env) {
  String configured(String key) {
    final fromEnvironment = Platform.environment[key];
    if (fromEnvironment != null && fromEnvironment.trim().isNotEmpty) {
      return fromEnvironment;
    }
    return env[key] ?? '';
  }

  final serverClientId = configured('GOOGLE_SERVER_CLIENT_ID').trim();
  final backendClientId = configured('GOOGLE_CLIENT_ID').trim();
  if (serverClientId.isNotEmpty &&
      backendClientId.isNotEmpty &&
      serverClientId != backendClientId) {
    throw const _LaunchException(
      'GOOGLE_SERVER_CLIENT_ID dan GOOGLE_CLIENT_ID harus memakai Web Client ID yang sama di backend/.env.',
    );
  }

  final clientId = serverClientId.isNotEmpty ? serverClientId : backendClientId;
  if (!clientId.endsWith('.apps.googleusercontent.com')) {
    throw const _LaunchException(
      'Google Web Client ID tidak ditemukan atau formatnya salah di backend/.env.',
    );
  }
  return clientId;
}

String _resolveOfflineAccessPublicKey(Map<String, String> env) {
  final fromEnvironment = Platform.environment['OFFLINE_ACCESS_PUBLIC_KEY']
      ?.trim();
  if (fromEnvironment != null && fromEnvironment.isNotEmpty) {
    return fromEnvironment;
  }
  return env['OFFLINE_ACCESS_PUBLIC_KEY']?.trim() ?? '';
}

int _backendPort(Map<String, String> env) {
  final environmentPort = Platform.environment['PORT'];
  final rawPort = environmentPort == null || environmentPort.isEmpty
      ? env['PORT'] ?? '8080'
      : environmentPort;
  final port = int.tryParse(rawPort);
  if (port == null || port < 1 || port > 65535) {
    throw const _LaunchException(
      'PORT di backend/.env harus berupa angka 1–65535.',
    );
  }
  return port;
}

Future<void> _waitForBackend(Process backend, int port) async {
  final client = HttpClient()..connectionTimeout = const Duration(seconds: 1);
  final deadline = DateTime.now().add(const Duration(minutes: 2));
  var backendFinished = false;
  var backendExitCode = -1;
  unawaited(
    backend.exitCode.then((code) {
      backendFinished = true;
      backendExitCode = code;
    }),
  );

  try {
    while (DateTime.now().isBefore(deadline)) {
      if (backendFinished) {
        throw _LaunchException(
          'Backend berhenti saat startup (exit code $backendExitCode).',
        );
      }
      try {
        final request = await client
            .getUrl(Uri.parse('http://127.0.0.1:$port/healthz'))
            .timeout(const Duration(seconds: 2));
        final response = await request.close().timeout(
          const Duration(seconds: 2),
        );
        await response.drain<void>();
        if (response.statusCode == HttpStatus.ok) return;
      } on SocketException {
        // Backend may still be compiling or starting.
      } on TimeoutException {
        // Try again until startup deadline.
      }
      await Future<void>.delayed(const Duration(milliseconds: 500));
    }
    throw const _LaunchException(
      'Backend belum siap setelah 2 menit. Periksa pesan error Go di terminal.',
    );
  } finally {
    client.close(force: true);
  }
}

_RunOptions _parseArguments(List<String> args) {
  var help = false;
  var skipBackend = false;
  var backendOnly = false;
  String? apiBaseUrl;
  final flutterArgs = <String>[];

  for (var index = 0; index < args.length; index++) {
    final argument = args[index];
    if (argument == '--help' || argument == '-h') {
      help = true;
    } else if (argument == '--no-backend') {
      skipBackend = true;
    } else if (argument == '--backend-only') {
      backendOnly = true;
    } else if (argument == '--api-base-url') {
      if (index + 1 >= args.length) {
        throw const _LaunchException('Isi URL setelah --api-base-url.');
      }
      apiBaseUrl = args[++index];
    } else if (argument.startsWith('--api-base-url=')) {
      apiBaseUrl = argument.substring('--api-base-url='.length);
    } else {
      flutterArgs.add(argument);
    }
  }

  if (apiBaseUrl != null && Uri.tryParse(apiBaseUrl)?.hasAuthority != true) {
    throw const _LaunchException('URL setelah --api-base-url tidak valid.');
  }
  if (skipBackend && backendOnly) {
    throw const _LaunchException(
      '--no-backend tidak bisa dipakai bersama --backend-only.',
    );
  }
  return _RunOptions(
    help: help,
    skipBackend: skipBackend,
    backendOnly: backendOnly,
    apiBaseUrl: apiBaseUrl,
    flutterArgs: flutterArgs,
  );
}

void _printUsage() {
  stdout.writeln('''
Jalankan dari folder utama project:
  dart run.dart [opsi launcher] [opsi flutter run]

Contoh:
  dart run.dart
  dart run.dart -d emulator-5554
  dart run.dart --api-base-url http://192.168.1.20:8080/v1 -d <DEVICE_ID>

Jika tidak ada perangkat Android aktif, launcher menyalakan AVD pertama.
Pilih AVD tertentu dengan ID dari `flutter emulators`, misalnya:
  dart run.dart -d Pixel_8_API_35

Opsi launcher:
  --api-base-url URL  Alamat backend yang dipakai aplikasi
  --no-backend        Backend sudah berjalan di terminal lain
  --backend-only      Jalankan backend saja
  -d, --device-id     Pilih perangkat atau AVD Android
  -h, --help          Tampilkan bantuan
''');
}

class _RunOptions {
  const _RunOptions({
    required this.help,
    required this.skipBackend,
    required this.backendOnly,
    required this.apiBaseUrl,
    required this.flutterArgs,
  });

  final bool help;
  final bool skipBackend;
  final bool backendOnly;
  final String? apiBaseUrl;
  final List<String> flutterArgs;
}

class _LaunchException implements Exception {
  const _LaunchException(this.message);

  final String message;
}
