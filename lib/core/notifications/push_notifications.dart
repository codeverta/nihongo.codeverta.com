import 'dart:async';
import 'dart:convert';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/storage/shared_preferences_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

const announcementTopic = 'nihongo_announcements';
const notificationRoutes = <String>{
  '/home',
  '/dictionary',
  '/anki',
  '/ssw',
  '/premium',
  '/learn/vocabulary',
  '/learn/grammar',
  '/learn/kanji',
  '/learn/reading',
  '/exams/n5',
};
String safeNotificationRoute(Object? value) =>
    value is String && notificationRoutes.contains(value) ? value : '/home';

@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  // Notification payloads are displayed by Android; do not post a duplicate.
  await Firebase.initializeApp();
}

class PushNotificationState {
  const PushNotificationState({
    this.available = false,
    this.enabled = false,
    this.granted = false,
    this.busy = true,
    this.pendingSync = false,
    this.message,
    this.navigation,
  });
  final bool available;
  final bool enabled;
  final bool granted;
  final bool busy;
  final bool pendingSync;
  final String? message;
  final String? navigation;
}

final pushNotificationsProvider =
    StateNotifierProvider<PushNotificationController, PushNotificationState>((
      ref,
    ) {
      return PushNotificationController(ref.watch(sharedPreferencesProvider));
    });

class PushNotificationController extends StateNotifier<PushNotificationState> {
  PushNotificationController(this._preferences)
    : super(
        PushNotificationState(
          available: supported,
          enabled: _preferences.getBool(_preferenceKey) ?? false,
          busy: false,
        ),
      ) {
    // Keep Dart messaging startup and FCM registration behind explicit opt-in.
    if (state.enabled ||
        (_preferences.getBool('announcement_sync_pending') ?? false)) {
      unawaited(initialize());
    } else if (!supported) {
      _set(
        message:
            'Notifikasi push tersedia untuk Android. Dukungan iPhone/iPad sedang disiapkan.',
      );
    }
  }
  static const _preferenceKey = 'announcement_notifications_enabled';
  final SharedPreferences _preferences;
  late final _local = FlutterLocalNotificationsPlugin();
  late final List<StreamSubscription<dynamic>> _subscriptions = [];
  final Set<String> _foregroundIDs = {};
  late bool _initialized = false;
  Future<void>? _initializing;
  late FirebaseMessaging _messaging;
  static bool get supported =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

  void _set({
    bool? available,
    bool? enabled,
    bool? granted,
    bool? busy,
    bool? pendingSync,
    String? message,
    String? navigation,
  }) {
    if (!mounted) return;
    state = PushNotificationState(
      available: available ?? state.available,
      enabled: enabled ?? state.enabled,
      granted: granted ?? state.granted,
      busy: busy ?? state.busy,
      pendingSync: pendingSync ?? state.pendingSync,
      message: message,
      navigation: navigation ?? state.navigation,
    );
  }

  Future<void> initialize() => _initializing ??= _initialize();
  Future<void> _initialize() async {
    _set(busy: true);
    if (!supported) {
      _set(
        busy: false,
        message:
            'Notifikasi push tersedia untuk Android. Dukungan iPhone/iPad sedang disiapkan.',
      );
      return;
    }
    try {
      await Firebase.initializeApp().timeout(const Duration(seconds: 15));
      if (!mounted) return;
      FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
      _messaging = FirebaseMessaging.instance;
      await _local.initialize(
        settings: const InitializationSettings(
          android: AndroidInitializationSettings('ic_notification'),
        ),
        onDidReceiveNotificationResponse: (response) =>
            _openLocal(response.payload),
      );
      await _local
          .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin
          >()
          ?.createNotificationChannel(
            const AndroidNotificationChannel(
              announcementTopic,
              'Pengumuman Nihongo Master',
              description:
                  'Berita aplikasi dan informasi belajar yang Anda aktifkan.',
            ),
          );
      if (!mounted) return;
      _subscriptions
        ..add(
          FirebaseMessaging.onMessage.listen((m) {
            unawaited(_showForeground(m));
          }),
        )
        ..add(
          FirebaseMessaging.onMessageOpenedApp.listen((m) {
            _set(navigation: safeNotificationRoute(m.data['route']));
          }),
        )
        ..add(
          _messaging.onTokenRefresh.listen((_) {
            // Reapply the public topic after registration/token renewal.
            if (mounted && state.enabled && !state.busy) {
              unawaited(_subscribeAfterRegistration());
            }
          }),
        );
      if (!mounted) return;
      _initialized = true;
      _set(available: true, busy: false);
      final initial = await _messaging.getInitialMessage();
      if (initial != null) {
        _set(navigation: safeNotificationRoute(initial.data['route']));
      }
      final localInitial = await _local.getNotificationAppLaunchDetails();
      if (localInitial?.didNotificationLaunchApp ?? false) {
        _openLocal(localInitial?.notificationResponse?.payload);
      }
      if (mounted) {
        await refresh();
      }
    } on Object {
      _set(
        available: false,
        busy: false,
        message:
            'Notifikasi belum dapat dihubungkan. Aplikasi tetap bisa digunakan. Tutup lalu buka lagi untuk mencoba kembali.',
      );
    }
  }

  Future<void> _subscribeAfterRegistration() async {
    if (!mounted || state.busy || !state.enabled) return;
    _set(busy: true);
    try {
      await _messaging
          .subscribeToTopic(announcementTopic)
          .timeout(const Duration(seconds: 15));
      _set(pendingSync: false);
    } on Object {
      _set(
        pendingSync: true,
        message: 'Langganan akan dicoba lagi saat aplikasi kembali aktif.',
      );
    } finally {
      if (mounted) {
        _set(
          busy: false,
          pendingSync: state.pendingSync,
          message: state.message,
        );
      }
    }
  }

  Future<void> refresh() async {
    if (!mounted || !_initialized || state.busy) return;
    _set(busy: true);
    try {
      final settings = await _messaging.getNotificationSettings();
      if (!mounted) return;
      final granted =
          settings.authorizationStatus == AuthorizationStatus.authorized ||
          settings.authorizationStatus == AuthorizationStatus.provisional;
      _set(granted: granted);
      if (state.enabled && granted) {
        await _messaging.setAutoInitEnabled(true);
        await _messaging.getToken().timeout(const Duration(seconds: 15));
        await _messaging
            .subscribeToTopic(announcementTopic)
            .timeout(const Duration(seconds: 15));
      } else if (!state.enabled &&
          (_preferences.getBool('announcement_sync_pending') ?? false)) {
        await _messaging.setAutoInitEnabled(false);
        await _messaging
            .unsubscribeFromTopic(announcementTopic)
            .timeout(const Duration(seconds: 15));
        await _messaging.deleteToken().timeout(const Duration(seconds: 15));
        await _messaging.setAutoInitEnabled(false);
        await _preferences.setBool('announcement_sync_pending', false);
      }
      if (!mounted) return;
      _set(
        busy: false,
        pendingSync: false,
        message: state.enabled && !granted
            ? 'Izin notifikasi di Android belum aktif. Buka pengaturan Android untuk mengizinkannya.'
            : null,
      );
    } on Object {
      if (!mounted) return;
      _set(
        busy: false,
        pendingSync: true,
        message: state.enabled
            ? 'Menunggu koneksi untuk mengaktifkan langganan. Akan dicoba lagi saat aplikasi kembali aktif.'
            : 'Menunggu koneksi untuk menghentikan langganan. Anda juga bisa mematikan izin notifikasi melalui pengaturan Android.',
      );
    }
  }

  Future<void> setEnabled({required bool enabled}) async {
    await initialize();
    if (!mounted || !state.available || state.busy) return;
    _set(busy: true);
    try {
      if (enabled) {
        final settings = await _messaging.requestPermission();
        if (settings.authorizationStatus != AuthorizationStatus.authorized &&
            settings.authorizationStatus != AuthorizationStatus.provisional) {
          _set(
            busy: false,
            granted: false,
            message:
                'Izin belum diberikan. Aktifkan melalui pengaturan Android jika Anda ingin menerima pengumuman.',
          );
          return;
        }
      }
      await _preferences.setBool(_preferenceKey, enabled);
      await _preferences.setBool('announcement_sync_pending', true);
      _set(enabled: enabled);
      if (!enabled) {
        try {
          await _local.cancelAll();
        } on Object {
          /* Retry subscription cleanup independently. */
        }
      }
      _set(busy: false);
      await refresh();
    } on Object {
      _set(
        busy: false,
        message: 'Pengaturan belum dapat diperbarui. Silakan coba lagi.',
      );
    }
  }

  Future<void> _showForeground(RemoteMessage m) async {
    if (!mounted ||
        !state.enabled ||
        !state.granted ||
        m.notification == null) {
      return;
    }
    final id = m.messageId ?? m.data['announcementId']?.toString();
    if (id != null && !_foregroundIDs.add(id)) return;
    if (_foregroundIDs.length > 100) {
      _foregroundIDs.remove(_foregroundIDs.first);
    }
    try {
      await _local.show(
        id: (id ?? DateTime.now().toIso8601String()).hashCode & 0x7fffffff,
        title: m.notification!.title,
        body: m.notification!.body,
        notificationDetails: const NotificationDetails(
          android: AndroidNotificationDetails(
            announcementTopic,
            'Pengumuman Nihongo Master',
            icon: 'ic_notification',
          ),
        ),
        payload: jsonEncode({'route': safeNotificationRoute(m.data['route'])}),
      );
    } on Object {
      /* Notification failures must not interrupt learning. */
    }
  }

  void _openLocal(String? payload) {
    try {
      final data = jsonDecode(payload ?? '{}');
      _set(
        navigation: safeNotificationRoute(data is Map ? data['route'] : null),
      );
    } on Object {
      _set(navigation: '/home');
    }
  }

  void consumeNavigation() {
    if (!mounted) return;
    state = PushNotificationState(
      available: state.available,
      enabled: state.enabled,
      granted: state.granted,
      busy: state.busy,
      pendingSync: state.pendingSync,
      message: state.message,
    );
  }

  Future<void> openSystemSettings() async {
    if (!supported) return;
    try {
      await const MethodChannel(
        'com.codeverta.nihongomaster/app_info',
      ).invokeMethod<void>('openNotificationSettings');
    } on Object {
      _set(
        message:
            'Buka Pengaturan Android → Aplikasi → Nihongo Master → Notifikasi.',
      );
    }
  }

  @override
  void dispose() {
    for (final subscription in _subscriptions) {
      unawaited(subscription.cancel());
    }
    super.dispose();
  }
}
