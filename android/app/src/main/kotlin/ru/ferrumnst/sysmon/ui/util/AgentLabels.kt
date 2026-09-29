package ru.ferrumnst.sysmon.ui.util

import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonPrimitive
import ru.ferrumnst.sysmon.data.models.AgentInfo

fun formatOsName(raw: String?): String {
    val trimmed = raw?.trim() ?: return ""
    if (trimmed.isEmpty()) return ""
    val win = Regex("^Microsoft\\s+Windows\\s+(\\d+(?:\\.\\d+)?)", RegexOption.IGNORE_CASE).find(trimmed)
    if (win != null) return "Windows ${win.groupValues[1]}"
    return trimmed.replaceFirstChar { if (it.isLowerCase()) it.titlecase() else it.toString() }
}

fun platformLabel(platform: String?): String = when (platform?.lowercase()) {
    "linux" -> "Linux"
    "windows" -> "Windows"
    "darwin" -> "macOS"
    else -> platform ?: ""
}

fun AgentInfo.systemOs(): String? {
    val system = system as? JsonObject ?: return null
    return system["os"]?.jsonPrimitive?.content
}

fun AgentInfo.osLabel(): String {
    val fromSystem = formatOsName(systemOs())
    if (fromSystem.isNotEmpty()) return fromSystem
    return platformLabel(platform)
}

fun AgentInfo.versionText(): String? {
    val version = agentVersion?.trim()
    return if (!version.isNullOrEmpty()) "v$version" else null
}

fun AgentInfo.versionAndOsText(): String? {
    val version = versionText()
    val os = osLabel()
    return when {
        version != null && os.isNotEmpty() -> "$version · $os"
        version != null -> version
        os.isNotEmpty() -> os
        else -> null
    }
}
