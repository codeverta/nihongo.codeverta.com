import 'package:flutter/widgets.dart';
import 'package:intl/intl.dart';

String regionalDate(DateTime value, {required String localeName}) =>
    DateFormat.yMd(localeName).format(value.toLocal());

String regionalDateTime(DateTime value, {required String localeName}) =>
    DateFormat.yMd(localeName).add_Hm().format(value.toLocal());

String regionalNumber(num value, {required String localeName}) =>
    NumberFormat.decimalPattern(localeName).format(value);

String appRegionLocaleName(BuildContext context) =>
    Localizations.localeOf(context).toString();
