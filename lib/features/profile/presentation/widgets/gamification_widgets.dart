import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/features/home/domain/home_dashboard.dart';
import 'package:nihongo_master/features/profile/data/gamification_repository.dart';

class GamificationStats extends StatelessWidget {
  const GamificationStats({required this.data, super.key});
  final GamificationSummary data;

  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Wrap(
            alignment: WrapAlignment.spaceAround,
            spacing: 20,
            runSpacing: 12,
            children: [
              _stat(
                '🔥',
                '${data.streakDays}',
                'Streak hari'.localized(context),
              ),
              _stat('⚡', '${data.xp}', 'XP'),
              _stat('🏆', '${data.unlockedCount}', 'Badge'.localized(context)),
            ],
          ),
          const SizedBox(height: 20),
          Text(
            'Level ${data.level} · ${data.levelXp}/${data.nextLevelXp} XP'
                .localized(context),
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(value: data.levelXp / data.nextLevelXp),
          const SizedBox(height: 12),
          Text(
            '${data.lessonsCompleted} lesson lulus · ${data.examsCompleted} jenis ujian dikerjakan'
                .localized(context),
          ),
          const SizedBox(height: 4),
          Text(
            'Streak terbaik ${data.bestStreak} hari · ${data.activeDays} hari aktif'
                .localized(context),
          ),
          const SizedBox(height: 12),
          Text(
            '50 XP saat pertama lulus lesson. 100 XP saat pertama mendapat ≥80% pada tiap jenis ujian. Ulangan tidak menggandakan XP. Streak dihitung dari kuis atau ujian yang selesai, zona waktu belajar yang dipilih.'
                .localized(context),
            style: const TextStyle(fontSize: 12, height: 1.6),
          ),
        ],
      ),
    ),
  );

  Widget _stat(String icon, String value, String label) => Column(
    children: [
      Text(icon, style: const TextStyle(fontSize: 22)),
      Text(
        value,
        style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800),
      ),
      Text(label),
    ],
  );
}

class AchievementTile extends StatelessWidget {
  const AchievementTile({required this.achievement, super.key});
  final Achievement achievement;

  @override
  Widget build(BuildContext context) {
    final a = achievement;
    final progress = a.target > 0
        ? (a.progress / a.target).clamp(0.0, 1.0)
        : 0.0;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(
              a.isUnlocked ? Icons.emoji_events_rounded : Icons.lock_outline,
              color: a.isUnlocked
                  ? Theme.of(context).colorScheme.primary
                  : null,
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    a.name.localized(context),
                    style: const TextStyle(fontWeight: FontWeight.w800),
                  ),
                  const SizedBox(height: 4),
                  Text(a.description.localized(context)),
                  const SizedBox(height: 10),
                  LinearProgressIndicator(value: progress),
                  const SizedBox(height: 6),
                  Text(
                    '${a.progress}/${a.target} · ${a.isUnlocked ? 'Terbuka' : 'Terkunci'}'
                        .localized(context),
                  ),
                  if (a.unlockedAt != null)
                    Text(
                      'Diperoleh ${regionalDate(a.unlockedAt!, localeName: Localizations.localeOf(context).toString())}'
                          .localized(context),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class GamificationError extends StatelessWidget {
  const GamificationError({required this.retry, super.key});
  final VoidCallback retry;
  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Text(
            'Progres belum bisa dimuat. Periksa koneksi atau masuk kembali jika sesi berakhir.'
                .localized(context),
          ),
          TextButton(
            onPressed: retry,
            child: Text('Coba lagi'.localized(context)),
          ),
        ],
      ),
    ),
  );
}
