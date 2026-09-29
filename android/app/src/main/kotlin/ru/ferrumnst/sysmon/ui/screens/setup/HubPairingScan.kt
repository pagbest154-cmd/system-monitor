package ru.ferrumnst.sysmon.ui.screens.setup

import android.net.Uri

data class HubPairingScan(
    val hubUrl: String,
    val hubName: String? = null,
    val hubKey: String? = null,
) {
    val includesCredentials: Boolean
        get() = !hubName.isNullOrBlank() && !hubKey.isNullOrBlank()
}

fun parseHubPairingScan(raw: String): HubPairingScan {
    val trimmed = raw.trim()
    if (trimmed.startsWith("sysmon://", ignoreCase = true)) {
        val uri = Uri.parse(trimmed)
        if (uri.scheme.equals("sysmon", ignoreCase = true) &&
            uri.host.equals("pair", ignoreCase = true)
        ) {
            val url = uri.getQueryParameter("url")?.trim()?.removeSuffix("/").orEmpty()
            if (url.isNotBlank()) {
                val name = uri.getQueryParameter("name")?.trim().orEmpty()
                val key = uri.getQueryParameter("key").orEmpty()
                return HubPairingScan(
                    hubUrl = url,
                    hubName = name.takeIf { it.isNotBlank() },
                    hubKey = key.takeIf { it.isNotBlank() },
                )
            }
        }
    }
    return HubPairingScan(hubUrl = trimmed.removeSuffix("/"))
}
