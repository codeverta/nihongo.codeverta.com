import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/gamification/double_xp_button.dart';

/// What one finished activity earned, as the server reports it with the
/// result (docs/ads-xp-spec.md). [completionId] names the activity for a
/// later "double XP" claim.
@immutable
class XpGain {
  const XpGain({
    required this.awarded,
    required this.total,
    required this.streak,
    this.completionId = '',
    this.kind = '',
  });

  /// Null when the response carries no XP (an older server, or an offline
  /// result that has not been verified yet).
  static XpGain? fromJson(Object? json) {
    if (json is! Map) return null;
    int number(Object? value) => value is num ? value.toInt() : 0;
    return XpGain(
      awarded: number(json['awarded']),
      total: number(json['total']),
      streak: number(json['streak']),
      completionId: '${json['completionId'] ?? ''}',
      kind: '${json['kind'] ?? ''}',
    );
  }

  final int awarded;
  final int total;
  final int streak;
  final String completionId;

  /// The activity group for double XP: lessons, exams or practice.
  final String kind;
}

/// Records a finished Kaigo or SSW curriculum lesson for XP and the streak.
/// Returns null when it could not be recorded (offline, signed out); the
/// lesson's local progress is kept either way.
Future<XpGain?> recordPracticeLesson(
  Dio dio,
  String sectorId,
  String lessonId,
) async {
  try {
    final response = await dio.post<Map<String, dynamic>>(
      '/me/practice-lessons/${Uri.encodeComponent(sectorId)}/${Uri.encodeComponent(lessonId)}/complete',
    );
    final data = response.data?['data'];
    return data is Map ? XpGain.fromJson(data['xp']) : null;
  } on Object catch (error) {
    debugPrint('Practice lesson XP not recorded: $error');
    return null;
  }
}

/// "+N XP · N-day streak" on a result screen. Draws nothing without a gain.
class XpResultBadge extends StatelessWidget {
  const XpResultBadge({required this.gain, super.key});

  /// Builds the badge from the raw `xp` field of a JSON result.
  factory XpResultBadge.json(Object? json, {Key? key}) =>
      XpResultBadge(gain: XpGain.fromJson(json), key: key);

  final XpGain? gain;

  @override
  Widget build(BuildContext context) {
    final gain = this.gain;
    if (gain == null || (gain.awarded <= 0 && gain.streak <= 0)) {
      return const SizedBox.shrink();
    }
    final colors = Theme.of(context).colorScheme;
    Widget pill(IconData icon, String text, Key key) => Container(
      key: key,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: colors.primaryContainer,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 18, color: colors.onPrimaryContainer),
          const SizedBox(width: 6),
          Text(
            text,
            style: TextStyle(
              fontWeight: FontWeight.w800,
              color: colors.onPrimaryContainer,
            ),
          ),
        ],
      ),
    );
    final pills = Wrap(
      alignment: WrapAlignment.center,
      spacing: 8,
      runSpacing: 8,
      children: [
        if (gain.awarded > 0)
          pill(
            Icons.bolt_rounded,
            '+${gain.awarded} XP',
            const Key('xp-awarded'),
          ),
        if (gain.streak > 0)
          pill(
            Icons.local_fire_department_rounded,
            '${gain.streak} ${'hari beruntun'.localized(context)}',
            const Key('xp-streak'),
          ),
      ],
    );
    if (gain.awarded <= 0) return pills;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        pills,
        const SizedBox(height: 10),
        DoubleXpButton(gain: gain),
      ],
    );
  }
}
