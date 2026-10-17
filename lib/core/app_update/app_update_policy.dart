class AppDeviceInfo {
  const AppDeviceInfo({
    required this.platform,
    required this.buildNumber,
    required this.version,
    required this.osVersion,
    this.androidApi = 0,
  });

  factory AppDeviceInfo.fromJson(Map<String, dynamic> json) {
    final build = int.tryParse(json['buildNumber'] as String? ?? '');
    final platform = json['platform'] as String?;
    if (build == null || build < 1 || !{'android', 'ios'}.contains(platform)) {
      throw const FormatException('Invalid native app version');
    }
    return AppDeviceInfo(
      platform: platform!,
      buildNumber: build,
      version: json['version'] as String,
      osVersion: json['osVersion'] as String,
      androidApi: json['androidApi'] as int,
    );
  }

  final String platform;
  final int buildNumber;
  final String version;
  final String osVersion;
  final int androidApi;
}

enum AppUpdateRequirement {
  supported,
  update,
  unsupportedAllowed,
  unsupportedBlocked
}

class AppUpdatePolicy {
  const AppUpdatePolicy({
    required this.enabled,
    required this.minimumBuild,
    required this.latestBuild,
    required this.latestVersion,
    required this.minimumAndroidApi,
    required this.minimumIOSVersion,
    required this.storeUrl,
    required this.unsupportedDevicePolicy,
    required this.message,
  });

  factory AppUpdatePolicy.fromJson(Map<String, dynamic> json, String platform) {
    final policy = AppUpdatePolicy(
      enabled: json['enabled'] as bool,
      minimumBuild: json['minimumBuild'] as int,
      latestBuild: json['latestBuild'] as int,
      latestVersion: json['latestVersion'] as String,
      minimumAndroidApi: json['minimumAndroidApi'] as int,
      minimumIOSVersion: json['minimumIOSVersion'] as String,
      storeUrl: json['storeUrl'] as String,
      unsupportedDevicePolicy: json['unsupportedDevicePolicy'] as String,
      message: json['message'] as String,
    );
    if (policy.minimumBuild < 0 ||
        policy.latestBuild < policy.minimumBuild ||
        policy.minimumAndroidApi < 0 ||
        !{'allow', 'block'}.contains(policy.unsupportedDevicePolicy) ||
        (policy.minimumIOSVersion.isNotEmpty &&
            !_osPattern.hasMatch(policy.minimumIOSVersion)) ||
        (policy.enabled &&
            (policy.latestBuild < 1 ||
                policy.latestVersion.isEmpty ||
                !policy.validStoreUrl(platform)))) {
      throw const FormatException('Invalid app update policy');
    }
    return policy;
  }

  static final _osPattern = RegExp(r'^\d{1,3}(\.\d{1,3}){0,2}$');

  final bool enabled;
  final int minimumBuild;
  final int latestBuild;
  final String latestVersion;
  final int minimumAndroidApi;
  final String minimumIOSVersion;
  final String storeUrl;
  final String unsupportedDevicePolicy;
  final String message;

  bool validStoreUrl(String platform) {
    final uri = Uri.tryParse(storeUrl);
    if (uri == null ||
        uri.scheme != 'https' ||
        uri.userInfo.isNotEmpty ||
        uri.hasPort ||
        uri.hasFragment) {
      return false;
    }
    if (platform == 'android') {
      return uri.host == 'play.google.com' &&
          uri.path == '/store/apps/details' &&
          uri.queryParametersAll.length == 1 &&
          uri.queryParametersAll['id']?.length == 1 &&
          uri.queryParameters['id'] == 'com.codeverta.nihongomaster';
    }
    return platform == 'ios' &&
        uri.host == 'apps.apple.com' &&
        !uri.hasQuery &&
        RegExp(r'^/(?:[a-z]{2}/)?app/(?:[^/]+/)?id[0-9]+$').hasMatch(uri.path);
  }

  AppUpdateRequirement evaluate(AppDeviceInfo device) {
    if (!enabled || minimumBuild == 0 || device.buildNumber >= minimumBuild) {
      return AppUpdateRequirement.supported;
    }
    final compatible = device.platform == 'android'
        ? device.androidApi >= minimumAndroidApi
        : minimumIOSVersion.isEmpty ||
            _atLeast(device.osVersion, minimumIOSVersion);
    if (compatible) return AppUpdateRequirement.update;
    return unsupportedDevicePolicy == 'allow'
        ? AppUpdateRequirement.unsupportedAllowed
        : AppUpdateRequirement.unsupportedBlocked;
  }

  static bool _atLeast(String current, String minimum) {
    if (!_osPattern.hasMatch(current)) {
      throw const FormatException('Invalid OS version');
    }
    final a = current.split('.').map(int.parse).toList();
    final b = minimum.split('.').map(int.parse).toList();
    for (var i = 0; i < 3; i++) {
      final first = i < a.length ? a[i] : 0;
      final second = i < b.length ? b[i] : 0;
      if (first != second) return first > second;
    }
    return true;
  }
}
