import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/device_timezone.dart';

void main() {
  test('accepts IANA timezone identifiers', () {
    expect(isIanaTimezoneIdentifier('America/Los_Angeles'), isTrue);
    expect(isIanaTimezoneIdentifier('Asia/Jakarta'), isTrue);
    expect(isIanaTimezoneIdentifier('UTC'), isTrue);
    expect(isIanaTimezoneIdentifier('Etc/GMT+7'), isTrue);
  });

  test('rejects short timezone names and malformed identifiers', () {
    expect(isIanaTimezoneIdentifier('PST'), isFalse);
    expect(isIanaTimezoneIdentifier('GMT+7'), isFalse);
    expect(isIanaTimezoneIdentifier('Asia/'), isFalse);
    expect(isIanaTimezoneIdentifier('America/..'), isFalse);
    expect(isIanaTimezoneIdentifier(''), isFalse);
  });
}
