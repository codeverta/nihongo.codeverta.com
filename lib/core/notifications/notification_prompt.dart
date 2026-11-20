import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/analytics/analytics_service.dart';
import 'package:nihongo_master/core/notifications/push_notifications.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const _answeredKey = 'notification_prompt_answered_v1';

/// Whether the person has already been asked about notifications (in
/// onboarding or on Beranda). Asked once, never nagged: any answer, including
/// "Nanti saja", counts. Turning notifications on later stays in Settings.
class NotificationPromptNotifier extends StateNotifier<bool> {
  NotificationPromptNotifier(this._preferences)
    : super(_preferences.getBool(_answeredKey) ?? false);

  final SharedPreferences _preferences;

  Future<void> markAnswered() async {
    state = true;
    await _preferences.setBool(_answeredKey, true);
  }
}

final notificationPromptAnsweredProvider =
    StateNotifierProvider<NotificationPromptNotifier, bool>(
      (ref) => NotificationPromptNotifier(ref.watch(sharedPreferencesProvider)),
    );

/// True when a prompt makes sense: push is supported, notifications are off,
/// and the person was not asked yet.
final shouldAskNotificationsProvider = Provider<bool>((ref) {
  final push = ref.watch(pushNotificationsProvider);
  final answered = ref.watch(notificationPromptAnsweredProvider);
  return push.available && !push.enabled && !answered;
});

/// `ref.read` of either a widget `ref` or a provider `ref`.
typedef ProviderReader = T Function<T>(ProviderListenable<T> provider);

/// Where the prompt was shown, for the anonymous opt-in rate event.
enum NotificationPromptPlace { onboarding, home }

/// Asks for permission and subscribes to announcements. Always records that the
/// person was asked, and logs `notification_prompt` (enabled / denied / failed)
/// when analytics is on. Never throws, so onboarding can never be blocked.
Future<PushToggleResult> enableNotificationsFromPrompt(
  ProviderReader read,
  NotificationPromptPlace place,
) async {
  PushToggleResult result;
  try {
    result = await read(
      pushNotificationsProvider.notifier,
    ).setEnabled(enabled: true);
  } on Object {
    result = PushToggleResult.failed;
  }
  // A failed attempt (for example offline) is not an answer: ask again later.
  if (result != PushToggleResult.failed) {
    await read(notificationPromptAnsweredProvider.notifier).markAnswered();
  }
  read(analyticsServiceProvider).logEvent('notification_prompt', {
    'place': place.name,
    'result': switch (result) {
      PushToggleResult.on => 'enabled',
      PushToggleResult.denied => 'denied',
      _ => 'failed',
    },
  });
  return result;
}

/// "Nanti saja": records the answer and logs `skipped`. No system dialog.
Future<void> skipNotificationPrompt(
  ProviderReader read,
  NotificationPromptPlace place,
) async {
  await read(notificationPromptAnsweredProvider.notifier).markAnswered();
  read(
    analyticsServiceProvider,
  ).logEvent('notification_prompt', {'place': place.name, 'result': 'skipped'});
}
