package ru.ferrumnst.sysmon.ui.util

import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import ru.ferrumnst.sysmon.data.models.AgentInfo
import ru.ferrumnst.sysmon.data.models.SystemMemoryInfo
import ru.ferrumnst.sysmon.data.models.SystemSwapInfo
import java.util.Locale

fun formatGb(value: Double): String {
    return if (value >= 10) {
        value.toInt().toString()
    } else {
        String.format(Locale.getDefault(), "%.1f", value)
    }
}

fun formatGbRange(usedGb: Double?, totalGb: Double?): String? {
    if (usedGb == null || totalGb == null) return null
    return "${formatGb(usedGb)} / ${formatGb(totalGb)} ГБ"
}

fun formatPercent(value: Double?): String? {
    if (value == null) return null
    return if (value >= 10) {
        "${value.toInt()}%"
    } else {
        String.format(Locale.getDefault(), "%.1f%%", value)
    }
}

fun formatResourceLine(
    usedGb: Double?,
    totalGb: Double?,
    percent: Double?,
): String? {
    val range = formatGbRange(usedGb, totalGb)
    val pct = formatPercent(percent)
    return when {
        range != null && pct != null -> "$range · $pct"
        range != null -> range
        pct != null -> pct
        else -> null
    }
}

fun SystemMemoryInfo.displayLine(): String? = formatResourceLine(usedGb, totalGb, percent)

fun SystemMemoryInfo.freeLine(): String? {
    val free = availableGb ?: freeGb
    return free?.let { "свободно ${formatGb(it)} ГБ" }
}

fun SystemSwapInfo.displayLine(): String? = formatResourceLine(usedGb, totalGb, percent)

fun AgentInfo.memoryUsedGb(): Double? {
    val memory = system as? JsonObject ?: return null
    return memory["memory"]?.jsonObject?.get("used_gb")?.jsonPrimitive?.doubleOrNull
}

fun AgentInfo.memoryTotalGb(): Double? {
    val memory = system as? JsonObject ?: return null
    return memory["memory"]?.jsonObject?.get("total_gb")?.jsonPrimitive?.doubleOrNull
}

fun AgentInfo.ramPercent(): Double? {
    val memory = system as? JsonObject ?: return null
    return memory["memory"]?.jsonObject?.get("percent")?.jsonPrimitive?.doubleOrNull
}

fun AgentInfo.cpuPercent(): Double? {
    val systemJson = system as? JsonObject ?: return null
    return systemJson["cpu"]?.jsonObject?.get("percent")?.jsonPrimitive?.doubleOrNull
}

fun AgentInfo.memoryDisplayLine(): String? {
    return formatResourceLine(memoryUsedGb(), memoryTotalGb(), ramPercent())
}
