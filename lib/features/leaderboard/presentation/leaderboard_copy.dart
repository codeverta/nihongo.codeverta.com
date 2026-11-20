import 'package:flutter/widgets.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';

/// Leaderboard copy that contains numbers. Static strings use `.localized()`;
/// these build both languages directly because they are interpolated.
class LeaderboardCopy {
  const LeaderboardCopy({required this.english});

  LeaderboardCopy.of(BuildContext context)
    : english = AppLocalizations.of(context).locale.languageCode == 'en';

  final bool english;

  /// "Pelajar #1234" when a person has no usable display name.
  String name(String displayName, String tag) {
    final trimmed = displayName.trim();
    if (trimmed.isNotEmpty) return trimmed;
    return english ? 'Learner #$tag' : 'Pelajar #$tag';
  }

  String rankOf(int rank, int participants) =>
      english ? '#$rank of $participants' : '#$rank dari $participants';

  String xp(int value) => '$value XP';

  /// "Berakhir dalam 2 hari 5 jam".
  String endsIn(Duration remaining) {
    final left = remaining.isNegative ? Duration.zero : remaining;
    final days = left.inDays;
    final hours = left.inHours % 24;
    final minutes = left.inMinutes % 60;
    if (english) {
      String unit(int n, String one) => '$n $one${n == 1 ? '' : 's'}';
      if (days > 0) {
        return 'Ends in ${unit(days, 'day')} ${unit(hours, 'hour')}';
      }
      if (hours > 0) {
        return 'Ends in ${unit(hours, 'hour')} ${unit(minutes, 'minute')}';
      }
      return 'Ends in ${unit(minutes, 'minute')}';
    }
    if (days > 0) return 'Berakhir dalam $days hari $hours jam';
    if (hours > 0) return 'Berakhir dalam $hours jam $minutes menit';
    return 'Berakhir dalam $minutes menit';
  }

  /// "Terakhir diperbarui 14.05".
  String lastUpdated(DateTime at) {
    final local = at.toLocal();
    final h = local.hour.toString().padLeft(2, '0');
    final m = local.minute.toString().padLeft(2, '0');
    return english ? 'Last updated $h:$m' : 'Terakhir diperbarui $h.$m';
  }

  /// Home row: "#12 minggu ini · 20 XP lagi ke #11".
  String homeRank(int rank, int gap) {
    if (rank <= 1) {
      return english
          ? "#1 this week · You're in the lead!"
          : '#1 minggu ini · Kamu memimpin!';
    }
    return english
        ? '#$rank this week · $gap XP more to reach #${rank - 1}'
        : '#$rank minggu ini · $gap XP lagi ke #${rank - 1}';
  }

  String hiddenRank(int rank) => english
      ? 'Your position if shown: #$rank'
      : 'Posisimu jika ditampilkan: #$rank';

  String lastWeek(int rank, int xp) => english
      ? 'Last week you finished #$rank · $xp XP. A new week has started!'
      : 'Minggu lalu kamu finis #$rank · $xp XP. Minggu baru dimulai!';
}
