import 'package:intl/date_symbol_data_local.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';

void main() {
  test('dates and numbers use the device region pattern', () async {
    await initializeDateFormatting();
    final date = DateTime.utc(2026, 1, 2);

    expect(regionalDate(date, localeName: 'en_US'), '1/2/2026');
    expect(regionalDate(date, localeName: 'en_GB'), '02/01/2026');
    expect(regionalNumber(12345, localeName: 'en_US'), '12,345');
    expect(regionalNumber(12345, localeName: 'id_ID'), '12.345');
  });
}
