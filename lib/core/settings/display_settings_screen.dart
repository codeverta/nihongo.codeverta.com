import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_language.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';

class DisplaySettingsScreen extends ConsumerWidget {
  const DisplaySettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final preferences = ref.watch(displayPreferencesProvider);
    final controller = ref.read(displayPreferencesProvider.notifier);
    final locale = ref.watch(appLanguageProvider);
    final languageController = ref.read(appLanguageProvider.notifier);
    final englishAvailable = ref.watch(englishAvailabilityProvider);

    return Scaffold(
      appBar: AppBar(title: Text('Pengaturan tampilan'.localized(context))),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            'Atur bantuan membaca di seluruh aplikasi.'.localized(context),
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          const SizedBox(height: 12),
          Card(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
                  child: Text(
                    'Bahasa aplikasi'.localized(context),
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Text(
                    'Bahasa yang dipilih hanya disimpan di perangkat ini.'
                        .localized(context),
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ),
                RadioGroup<String>(
                  groupValue: locale.languageCode,
                  onChanged: (value) {
                    if (value != null) {
                      languageController.selectLanguage(value);
                    }
                  },
                  child: Column(
                    children: [
                      RadioListTile<String>(
                        key: const Key('language-option-id'),
                        value: 'id',
                        title: Text('Bahasa Indonesia'.localized(context)),
                      ),
                      if (englishAvailable)
                        const RadioListTile<String>(
                          key: Key('language-option-en'),
                          value: 'en',
                          title: Text('English'),
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: ListTile(
              leading: const Icon(Icons.schedule_rounded),
              title: Text('Zona waktu belajar'.localized(context)),
              subtitle: Text(
                'Zona waktu ini menentukan pergantian hari untuk streak dan progress belajar.'
                    .localized(context),
              ),
              trailing: const Icon(Icons.chevron_right_rounded),
              onTap: () => context.push('/settings/timezone'),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: Column(
              children: [
                SwitchListTile.adaptive(
                  key: const Key('display-furigana-setting'),
                  title: Text('Furigana'.localized(context)),
                  subtitle: Text(
                    'Tampilkan cara baca di atas kanji.'.localized(context),
                  ),
                  value: preferences.furigana,
                  onChanged: (enabled) =>
                      controller.setFurigana(enabled: enabled),
                ),
                const Divider(height: 1, indent: 16, endIndent: 16),
                SwitchListTile.adaptive(
                  key: const Key('display-translation-setting'),
                  title: Text('Terjemahan'.localized(context)),
                  subtitle: Text(
                    'Tampilkan arti contoh dan bacaan.'.localized(context),
                  ),
                  value: preferences.translation,
                  onChanged: (enabled) =>
                      controller.setTranslation(enabled: enabled),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
