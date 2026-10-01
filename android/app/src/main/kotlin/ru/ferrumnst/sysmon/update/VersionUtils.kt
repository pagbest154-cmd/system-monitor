package ru.ferrumnst.sysmon.update

object VersionUtils {
    fun normalize(version: String): String {
        val trimmed = version.trim().removePrefix("v")
        val dash = trimmed.indexOf('-')
        return if (dash >= 0) trimmed.substring(0, dash) else trimmed
    }

    fun versionKey(version: String): List<Int> {
        return normalize(version)
            .split('.')
            .map { part -> part.toIntOrNull() ?: 0 }
    }

    fun isNewer(latest: String, current: String): Boolean {
        val latestParts = versionKey(latest)
        val currentParts = versionKey(current)
        val maxLen = maxOf(latestParts.size, currentParts.size)
        for (index in 0 until maxLen) {
            val latestValue = latestParts.getOrElse(index) { 0 }
            val currentValue = currentParts.getOrElse(index) { 0 }
            if (latestValue > currentValue) return true
            if (latestValue < currentValue) return false
        }
        return false
    }

    fun apkAssetName(version: String): String {
        val label = version.trim().removePrefix("v")
        return "sysmon-$label.apk"
    }
}
