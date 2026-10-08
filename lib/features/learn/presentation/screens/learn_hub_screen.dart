import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/routing/app_router.dart';

class LearnHubScreen extends StatefulWidget {
  const LearnHubScreen({super.key});

  @override
  State<LearnHubScreen> createState() => _LearnHubScreenState();
}

class _LearnHubScreenState extends State<LearnHubScreen> {
  int _selectedLevel = 0;

  static const _levels = ['N5', 'N4', 'N3', 'N2', 'N1'];
  static const _subjects = [
    (
      'Vocabulary',
      'Kosakata, contoh, kuis & progress',
      Icons.translate_rounded
    ),
    (
      'Grammar',
      'Pola, contoh kalimat, kuis & progress',
      Icons.menu_book_rounded
    ),
    ('Kanji', 'Arti, bacaan, contoh kata & kuis', Icons.brush_rounded),
    (
      'Reading',
      'Cerita bergambar, kosakata & kuis',
      Icons.auto_stories_rounded
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return CustomScrollView(
      key: const PageStorageKey('learn-scroll'),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
          sliver: SliverList.list(
            children: [
              Text(
                'Belajar',
                style: Theme.of(context)
                    .textTheme
                    .headlineMedium
                    ?.copyWith(fontWeight: FontWeight.w900),
              ),
              const SizedBox(height: 8),
              const Text(
                'Pilih level target, lalu mulai dari materi yang sudah tersedia.',
              ),
              const SizedBox(height: 24),
              Text(
                'Pilih level JLPT',
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 12),
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  children: List.generate(_levels.length, (index) {
                    final selected = _selectedLevel == index;
                    return Padding(
                      padding: const EdgeInsets.only(right: 10),
                      child: ChoiceChip(
                        selected: selected,
                        onSelected: (value) {
                          if (value) setState(() => _selectedLevel = index);
                        },
                        label: Text(_levels[index]),
                        labelStyle: TextStyle(
                          fontWeight: FontWeight.w800,
                          color: selected ? colors.onPrimary : null,
                        ),
                        selectedColor: colors.primary,
                        showCheckmark: false,
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                      ),
                    );
                  }),
                ),
              ),
              const SizedBox(height: 24),
              _VocabularyCallout(
                level: _levels[_selectedLevel],
                onPressed: () => context.push(
                  '${AppRoutes.vocabularyPath}?level=${_levels[_selectedLevel]}',
                ),
              ),
              const SizedBox(height: 28),
              Text(
                'Materi ${_levels[_selectedLevel]}',
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 14),
              for (final (index, subject) in _subjects.indexed)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: _SubjectTile(
                    title: subject.$1,
                    subtitle: subject.$2,
                    icon: subject.$3,
                    onTap: index == 0
                        ? () => context.push(
                              '${AppRoutes.vocabularyPath}?level=${_levels[_selectedLevel]}',
                            )
                        : index == 1 && _selectedLevel <= 2
                            ? () => context.push(
                                  '${AppRoutes.grammarPath}?level=${_levels[_selectedLevel]}',
                                )
                            : index == 2 && _selectedLevel <= 2
                                ? () => context.push(
                                      '${AppRoutes.kanjiPath}?level=${_levels[_selectedLevel]}',
                                    )
                                : index == 3 && _selectedLevel <= 3
                                    ? () => context.push(
                                          AppRoutes.readingCoursePath(
                                            _levels[_selectedLevel],
                                          ),
                                        )
                                    : null,
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}

class _VocabularyCallout extends StatelessWidget {
  const _VocabularyCallout({required this.level, required this.onPressed});

  final String level;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: colors.secondary,
        borderRadius: BorderRadius.circular(22),
      ),
      child: Row(
        children: [
          Container(
            width: 58,
            height: 58,
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.14),
              borderRadius: BorderRadius.circular(17),
            ),
            child: const Icon(
              Icons.translate_rounded,
              color: Colors.white,
              size: 30,
            ),
          ),
          const SizedBox(width: 15),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Vocabulary JLPT $level',
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w800,
                    fontSize: 16,
                  ),
                ),
                const SizedBox(height: 5),
                Text(
                  'Pelajari kata, contoh, lalu uji pemahamanmu',
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.76),
                  ),
                ),
              ],
            ),
          ),
          IconButton.filled(
            key: const Key('open-vocabulary-course'),
            onPressed: onPressed,
            style: IconButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: colors.secondary,
            ),
            icon: const Icon(Icons.arrow_forward_rounded),
          ),
        ],
      ),
    );
  }
}

class _SubjectTile extends StatelessWidget {
  const _SubjectTile({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.onTap,
  });

  final String title;
  final String subtitle;
  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Card(
      child: ListTile(
        onTap: onTap,
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        leading: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: colors.primary.withValues(alpha: 0.09),
            borderRadius: BorderRadius.circular(14),
          ),
          child: Icon(icon, color: colors.primary),
        ),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w700)),
        subtitle: Text(subtitle),
        trailing: onTap == null
            ? const Chip(label: Text('Segera'))
            : const Icon(Icons.chevron_right_rounded),
      ),
    );
  }
}
