import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/settings/display_preferences.dart';

class DisplaySettingsScreen extends ConsumerWidget {
  const DisplaySettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final preferences = ref.watch(displayPreferencesProvider);
    final controller = ref.read(displayPreferencesProvider.notifier);

    return Scaffold(
      appBar: AppBar(title: const Text('Pengaturan tampilan')),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            'Atur bantuan membaca di seluruh aplikasi.',
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          const SizedBox(height: 12),
          Card(
            child: Column(
              children: [
                SwitchListTile.adaptive(
                  key: const Key('display-furigana-setting'),
                  title: const Text('Furigana'),
                  subtitle: const Text('Tampilkan cara baca di atas kanji.'),
                  value: preferences.furigana,
                  onChanged: (enabled) =>
                      controller.setFurigana(enabled: enabled),
                ),
                const Divider(height: 1, indent: 16, endIndent: 16),
                SwitchListTile.adaptive(
                  key: const Key('display-translation-setting'),
                  title: const Text('Terjemahan'),
                  subtitle: const Text('Tampilkan arti contoh dan bacaan.'),
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
