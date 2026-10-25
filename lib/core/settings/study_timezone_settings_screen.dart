import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/device_timezone.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

const studyTimezoneSuggestions = <String>[
  'UTC',
  'Asia/Jakarta',
  'Asia/Pontianak',
  'Asia/Makassar',
  'Asia/Jayapura',
  'Asia/Bangkok',
  'Asia/Ho_Chi_Minh',
  'Asia/Singapore',
  'Asia/Kuala_Lumpur',
  'Asia/Manila',
  'Asia/Hong_Kong',
  'Asia/Shanghai',
  'Asia/Taipei',
  'Asia/Seoul',
  'Asia/Tokyo',
  'Asia/Kolkata',
  'Asia/Dubai',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Moscow',
  'Africa/Cairo',
  'Africa/Johannesburg',
  'America/Sao_Paulo',
  'America/Toronto',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Vancouver',
  'Australia/Perth',
  'Australia/Sydney',
  'Pacific/Auckland',
];

final studyTimezoneRepositoryProvider = Provider<StudyTimezoneRepository>(
  (ref) => StudyTimezoneRepository(ref.watch(dioProvider)),
);

class StudyTimezoneRepository {
  const StudyTimezoneRepository(this._dio);

  final Dio _dio;

  Future<void> update(String timezone) async {
    await _dio.put<void>('/me/timezone', data: {'timezone': timezone});
  }
}

class StudyTimezoneSettingsScreen extends ConsumerStatefulWidget {
  const StudyTimezoneSettingsScreen({super.key});

  @override
  ConsumerState<StudyTimezoneSettingsScreen> createState() =>
      _StudyTimezoneSettingsScreenState();
}

class _StudyTimezoneSettingsScreenState
    extends ConsumerState<StudyTimezoneSettingsScreen> {
  String? _selectedTimezone;
  bool _saving = false;

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authControllerProvider).user;
    final gamification = ref.watch(gamificationProvider);
    final deviceTimezone = ref.watch(deviceTimeZoneIdentifierProvider);

    return Scaffold(
      appBar: AppBar(title: Text('Zona waktu belajar'.localized(context))),
      body: gamification.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) => _TimezoneError(
          message: user == null
              ? 'Masuk untuk mengatur zona waktu belajar.'.localized(context)
              : 'Zona waktu belajar belum dapat dimuat.'.localized(context),
          onRetry: () => ref.invalidate(gamificationProvider),
        ),
        data: (summary) {
          final choices = <String>{
            summary.timezone,
            ...studyTimezoneSuggestions,
            if (deviceTimezone.valueOrNull case final timezone?) timezone,
          }.toList()..sort();
          final selected = _selectedTimezone ?? summary.timezone;
          return ListView(
            padding: const EdgeInsets.all(20),
            children: [
              Text(
                'Zona waktu ini menentukan pergantian hari untuk streak dan progress belajar.'
                    .localized(context),
                style: Theme.of(context).textTheme.bodyLarge,
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        'Zona waktu akun'.localized(context),
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 12),
                      DropdownButtonFormField<String>(
                        key: ValueKey(selected),
                        initialValue: selected,
                        isExpanded: true,
                        decoration: InputDecoration(
                          labelText: 'Pilih zona waktu'.localized(context),
                          border: const OutlineInputBorder(),
                        ),
                        items: [
                          for (final timezone in choices)
                            DropdownMenuItem(
                              value: timezone,
                              child: Text(timezone),
                            ),
                        ],
                        onChanged: _saving
                            ? null
                            : (value) {
                                if (value != null) {
                                  setState(() => _selectedTimezone = value);
                                }
                              },
                      ),
                      if (deviceTimezone.valueOrNull case final timezone?) ...[
                        const SizedBox(height: 8),
                        Align(
                          alignment: Alignment.centerLeft,
                          child: TextButton.icon(
                            onPressed: _saving
                                ? null
                                : () => setState(
                                    () => _selectedTimezone = timezone,
                                  ),
                            icon: const Icon(Icons.my_location_rounded),
                            label: Text(
                              'Gunakan zona waktu perangkat ($timezone)'
                                  .localized(context),
                            ),
                          ),
                        ),
                      ],
                      const SizedBox(height: 8),
                      FilledButton(
                        onPressed: _saving || selected == summary.timezone
                            ? null
                            : () => _save(selected),
                        child: _saving
                            ? const SizedBox.square(
                                dimension: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : Text('Simpan zona waktu'.localized(context)),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Perubahan zona waktu hanya berlaku setelah kamu menyimpannya.'
                    .localized(context),
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _save(String timezone) async {
    setState(() => _saving = true);
    try {
      await ref.read(studyTimezoneRepositoryProvider).update(timezone);
      if (!mounted) return;
      ref.invalidate(gamificationProvider);
      setState(() => _selectedTimezone = null);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Zona waktu belajar berhasil diperbarui.'.localized(context),
          ),
        ),
      );
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Zona waktu belajar gagal diperbarui.'.localized(context),
          ),
        ),
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }
}

class _TimezoneError extends StatelessWidget {
  const _TimezoneError({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => Center(
    child: Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 12),
          OutlinedButton(
            onPressed: onRetry,
            child: Text('Coba lagi'.localized(context)),
          ),
        ],
      ),
    ),
  );
}
