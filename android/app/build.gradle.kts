plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "ru.ferrumnst.sysmon"
    compileSdk = 35

    defaultConfig {
        applicationId = "ru.ferrumnst.sysmon"
        minSdk = 26
        targetSdk = 35
        val appVersion = providers.gradleProperty("appVersion").orElse("1.0.0").get()
        val appVersionCode = providers.gradleProperty("appVersionCode")
            .map { it.toInt() }
            .orElse(appVersion.substringAfterLast('.').toIntOrNull() ?: 1)
        versionCode = appVersionCode.get()
        versionName = appVersion
    }

    signingConfigs {
        val releaseKeystore = rootProject.file("release.keystore")
        if (releaseKeystore.exists()) {
            create("release") {
                storeFile = releaseKeystore
                storePassword = System.getenv("ANDROID_KEYSTORE_PASSWORD")
                    ?: providers.gradleProperty("RELEASE_STORE_PASSWORD").orNull
                keyAlias = System.getenv("ANDROID_KEY_ALIAS")
                    ?: providers.gradleProperty("RELEASE_KEY_ALIAS").orNull
                keyPassword = System.getenv("ANDROID_KEY_PASSWORD")
                    ?: providers.gradleProperty("RELEASE_KEY_PASSWORD").orNull
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
            signingConfig = signingConfigs.findByName("release")
                ?: signingConfigs.getByName("debug")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.fragment.ktx)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.material.icons.extended)
    implementation(libs.androidx.navigation.compose)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.retrofit)
    implementation(libs.retrofit.kotlinx.serialization)
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.glance.appwidget)
    implementation(libs.glance.material3)
    implementation(libs.androidx.work.runtime.ktx)
    implementation(libs.androidx.biometric)
    implementation(libs.zxing.android.embedded)
    debugImplementation(libs.androidx.ui.tooling)
    testImplementation(libs.junit)
}
