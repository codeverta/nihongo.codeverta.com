import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/offline/network_status.dart';

bool requireOnlineExam(BuildContext context, WidgetRef ref) {
  if (!ref.read(offlineProvider)) return true;
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      content: Text(
        'Simulasi ujian memerlukan koneksi internet. Materi dan latihan lesson yang sudah disiapkan tetap bisa dipakai offline.'
            .localized(context),
      ),
    ),
  );
  return false;
}
