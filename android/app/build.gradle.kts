import java.io.File
import java.io.FileInputStream
import java.util.Properties

plugins {
    id("com.android.application")
    id("com.google.gms.google-services")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

val signingProperties = Properties()
val signingPropertiesFile = rootProject.file("key.properties")
if (signingPropertiesFile.exists()) {
    FileInputStream(signingPropertiesFile).use(signingProperties::load)
}

fun signingValue(environmentName: String, propertyName: String): String? =
    System.getenv(environmentName)?.takeIf(String::isNotBlank)
        ?: signingProperties.getProperty(propertyName)?.takeIf(String::isNotBlank)

val uploadStorePath = signingValue("ANDROID_KEYSTORE_PATH", "storeFile")
val uploadStorePassword = signingValue("ANDROID_KEYSTORE_PASSWORD", "storePassword")
val uploadKeyAlias = signingValue("ANDROID_KEY_ALIAS", "keyAlias")
val uploadKeyPassword = signingValue("ANDROID_KEY_PASSWORD", "keyPassword")
val releaseSigningConfigured =
    uploadStorePath != null &&
        File(uploadStorePath).isFile &&
        uploadStorePassword != null &&
        uploadKeyAlias != null &&
        uploadKeyPassword != null

if (!releaseSigningConfigured) {
    logger.warn(
        "No upload key configured: release builds use the debug certificate and cannot be uploaded to Play Console.",
    )
}

android {
    namespace = "com.codeverta.nihongomaster"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        isCoreLibraryDesugaringEnabled = true
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.codeverta.nihongomaster"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        if (releaseSigningConfigured) {
            create("release") {
                keyAlias = requireNotNull(uploadKeyAlias)
                keyPassword = requireNotNull(uploadKeyPassword)
                storeFile = File(requireNotNull(uploadStorePath))
                storePassword = requireNotNull(uploadStorePassword)
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
            // The GitHub release workflow requires upload-key secrets. The debug
            // fallback is only for local preview builds, never for Play uploads.
            signingConfig = signingConfigs.getByName(
                if (releaseSigningConfigured) "release" else "debug",
            )
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}

// Firebase Flutter plugins manage compatible native Firebase SDK versions.
dependencies {
    coreLibraryDesugaring("com.android.tools:desugar_jdk_libs:2.1.4")
}
