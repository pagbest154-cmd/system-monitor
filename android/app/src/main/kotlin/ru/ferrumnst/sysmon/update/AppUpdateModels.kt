package ru.ferrumnst.sysmon.update

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class GhRelease(
    @SerialName("tag_name") val tagName: String = "",
    val draft: Boolean = false,
    @SerialName("prerelease") val prerelease: Boolean = false,
    @SerialName("html_url") val htmlUrl: String = "",
    val assets: List<GhAsset> = emptyList(),
)

@Serializable
data class GhAsset(
    val name: String = "",
    @SerialName("browser_download_url") val browserDownloadUrl: String = "",
)

data class AppUpdateInfo(
    val currentVersion: String,
    val latestVersion: String? = null,
    val updateAvailable: Boolean = false,
    val downloadUrl: String? = null,
    val releaseUrl: String? = null,
    val apkFileName: String? = null,
    val error: String? = null,
)
