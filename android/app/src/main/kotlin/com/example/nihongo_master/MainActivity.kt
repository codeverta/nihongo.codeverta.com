package com.codeverta.nihongomaster

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import android.os.Build

class MainActivity : FlutterActivity() {
    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, "com.codeverta.nihongomaster/app_info")
            .setMethodCallHandler { call, result ->
                if (call.method != "getAppInfo") {
                    result.notImplemented()
                } else {
                    try {
                        @Suppress("DEPRECATION")
                        val info = packageManager.getPackageInfo(packageName, 0)
                        @Suppress("DEPRECATION")
                        val build = if (Build.VERSION.SDK_INT >= 28) info.longVersionCode else info.versionCode.toLong()
                        result.success(mapOf(
                            "platform" to "android",
                            "buildNumber" to build.toString(),
                            "version" to (info.versionName ?: ""),
                            "androidApi" to Build.VERSION.SDK_INT,
                            "osVersion" to Build.VERSION.RELEASE
                        ))
                    } catch (error: Exception) {
                        result.error("APP_INFO_UNAVAILABLE", "Tidak dapat membaca versi aplikasi", null)
                    }
                }
            }
    }
}
