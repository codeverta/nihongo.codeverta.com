import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_timezone/flutter_timezone.dart';

/// Returns the device's IANA timezone identifier, never a localized display
/// name or a short abbreviation such as PST.
final deviceTimeZoneIdentifierProvider = FutureProvider<String?>((ref) async {
  try {
    final timezone = await FlutterTimezone.getLocalTimezone();
    final identifier = timezone.identifier.trim();
    if (isIanaTimezoneIdentifier(identifier)) return identifier;
  } on Object {
    // Timezone is optional metadata. A failed platform lookup must not block
    // sign-in or any other API request.
  }
  return null;
});

bool isIanaTimezoneIdentifier(String value) {
  if (value == 'UTC') return true;
  if (value.isEmpty || value.contains(RegExp(r'\s'))) return false;
  if (value.startsWith('Etc/')) return value.length > 4;
  return value.contains('/') &&
      value
          .split('/')
          .every((part) => part.isNotEmpty && part != '.' && part != '..');
}
