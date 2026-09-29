package ru.ferrumnst.sysmon.update

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.provider.Settings
import androidx.core.content.FileProvider
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.File
import java.util.concurrent.TimeUnit

class AppUpdateManager(
    private val context: Context,
) {
    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
    }

    private val httpClient: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(20, TimeUnit.SECONDS)
        .readTimeout(120, TimeUnit.SECONDS)
        .writeTimeout(120, TimeUnit.SECONDS)
        .build()

    private val _updateInfo = MutableStateFlow<AppUpdateInfo?>(null)
    val updateInfo: StateFlow<AppUpdateInfo?> = _updateInfo.asStateFlow()

    fun currentAppVersion(): String {
        val packageInfo = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            context.packageManager.getPackageInfo(
                context.packageName,
                PackageManager.PackageInfoFlags.of(0),
            )
        } else {
            @Suppress("DEPRECATION")
            context.packageManager.getPackageInfo(context.packageName, 0)
        }
        return packageInfo.versionName ?: "0.0.0"
    }

    suspend fun checkForUpdate(force: Boolean = false): AppUpdateInfo {
        if (!force) {
            _updateInfo.value?.let { cached ->
                if (cached.error == null) return cached
            }
        }

        val current = currentAppVersion()
        return withContext(Dispatchers.IO) {
            runCatching {
                val request = Request.Builder()
                    .url("$GITHUB_API/releases/latest")
                    .header("Accept", "application/vnd.github+json")
                    .header("User-Agent", USER_AGENT)
                    .build()

                httpClient.newCall(request).execute().use { response ->
                    if (!response.isSuccessful) {
                        throw IllegalStateException("GitHub API ${response.code}")
                    }
                    val body = response.body?.string() ?: throw IllegalStateException("Пустой ответ GitHub")
                    val release = json.decodeFromString(GhRelease.serializer(), body)
                    if (release.draft || release.prerelease) {
                        AppUpdateInfo(currentVersion = current)
                    } else {
                        val latest = VersionUtils.normalize(release.tagName)
                        val apkName = VersionUtils.apkAssetName(latest)
                        val asset = release.assets.firstOrNull { it.name == apkName }
                        val updateAvailable = asset != null && VersionUtils.isNewer(latest, current)
                        AppUpdateInfo(
                            currentVersion = current,
                            latestVersion = latest,
                            updateAvailable = updateAvailable,
                            downloadUrl = asset?.browserDownloadUrl,
                            releaseUrl = release.htmlUrl,
                            apkFileName = asset?.name,
                        )
                    }
                }
            }.getOrElse { error ->
                AppUpdateInfo(
                    currentVersion = current,
                    error = error.message ?: "Не удалось проверить обновления",
                )
            }.also { info ->
                _updateInfo.value = info
            }
        }
    }

    suspend fun downloadApk(
        downloadUrl: String,
        fileName: String,
        onProgress: (Float) -> Unit,
    ): File = withContext(Dispatchers.IO) {
        val target = File(context.cacheDir, fileName)
        if (target.exists()) {
            target.delete()
        }

        val request = Request.Builder()
            .url(downloadUrl)
            .header("User-Agent", USER_AGENT)
            .build()

        httpClient.newCall(request).execute().use { response ->
            if (!response.isSuccessful) {
                throw IllegalStateException("Ошибка загрузки APK (${response.code})")
            }
            val body = response.body ?: throw IllegalStateException("Пустой ответ при загрузке APK")
            val totalBytes = body.contentLength().coerceAtLeast(0L)
            body.byteStream().use { input ->
                target.outputStream().use { output ->
                    val buffer = ByteArray(DEFAULT_BUFFER_SIZE)
                    var downloaded = 0L
                    while (true) {
                        val read = input.read(buffer)
                        if (read <= 0) break
                        output.write(buffer, 0, read)
                        downloaded += read
                        if (totalBytes > 0L) {
                            onProgress(downloaded.toFloat() / totalBytes.toFloat())
                        }
                    }
                }
            }
        }
        onProgress(1f)
        target
    }

    fun canInstallPackages(): Boolean {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return true
        return context.packageManager.canRequestPackageInstalls()
    }

    fun openInstallPermissionSettings(activity: Activity) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val intent = Intent(
            Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES,
            Uri.parse("package:${context.packageName}"),
        )
        activity.startActivity(intent)
    }

    fun installApk(apkFile: File) {
        val uri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            apkFile,
        )
        val intent = Intent(Intent.ACTION_VIEW).apply {
            setDataAndType(uri, "application/vnd.android.package-archive")
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION or Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        context.startActivity(intent)
    }

    companion object {
        private const val GITHUB_REPO = "pagbest154-cmd/system-monitor"
        private const val GITHUB_API = "https://api.github.com/repos/$GITHUB_REPO"
        private const val USER_AGENT = "SysMon-Android"
    }
}
