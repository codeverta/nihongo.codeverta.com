import 'package:flutter/foundation.dart';

const configuredApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8080/v1',
);

String resolveApiBaseUrl(String configuredUrl, TargetPlatform platform) {
  final uri = Uri.tryParse(configuredUrl);
  if (uri == null || platform != TargetPlatform.android) return configuredUrl;

  final host = uri.host.toLowerCase();
  if (host != 'localhost' && host != '127.0.0.1' && host != '::1') {
    return configuredUrl;
  }

  // Inside Android Emulator, loopback belongs to the virtual device itself.
  // 10.0.2.2 is the emulator's route to the development host machine.
  return uri.replace(host: '10.0.2.2').toString();
}
