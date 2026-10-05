package ru.ferrumnst.sysmon.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.SensorInfo

private fun JsonElement?.detailString(key: String): String? {
    val obj = this?.jsonObject ?: return null
    val el = obj[key] ?: return null
    if (el is JsonNull) return null
    return el.jsonPrimitive.contentOrNull
}

private fun JsonElement?.detailDouble(key: String): Double? {
    val obj = this?.jsonObject ?: return null
    val el = obj[key] ?: return null
    if (el is JsonNull) return null
    return el.jsonPrimitive.doubleOrNull
}

private fun JsonElement?.detailInt(key: String): Int? {
    return detailDouble(key)?.toInt()
}

private fun raidStateLabel(state: String?): String = when (state) {
    "active" -> "Активен"
    "degraded" -> "Деградация"
    "clean" -> "Чистый"
    "inactive" -> "Неактивен"
    else -> state ?: "—"
}

@Composable
fun RaidPanelCard(
    title: String,
    sensorIds: List<String>,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
    modifier: Modifier = Modifier,
) {
    DashboardCard(title = title, modifier = modifier) {
        if (sensorIds.isEmpty()) {
            Text(
                text = "RAID-массивы не обнаружены",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp),
            )
            return@DashboardCard
        }
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            sensorIds.forEach { sensorId ->
                val meta = sensors.find { it.id == sensorId }
                val reading = liveReadings[sensorId] ?: meta?.current
                val details = reading?.details
                val device = details.detailString("device") ?: meta?.name ?: sensorId
                val level = details.detailString("raid_level") ?: "—"
                val state = raidStateLabel(details.detailString("state"))
                val active = details.detailInt("active_devices")?.toString() ?: "—"
                val failed = details.detailInt("failed_devices")?.toString() ?: "0"
                val check = details.detailDouble("check_progress")
                Column(modifier = Modifier.fillMaxWidth()) {
                    Text(device, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                    Text("Уровень: $level", style = MaterialTheme.typography.bodySmall)
                    Text("Состояние: $state", style = MaterialTheme.typography.bodySmall)
                    Text("Активные: $active · Сбойные: $failed", style = MaterialTheme.typography.bodySmall)
                    if (check != null) {
                        Text("Проверка: ${"%.1f".format(check)}%", style = MaterialTheme.typography.bodySmall)
                    }
                    raidMemberLines(details).forEach { line ->
                        Text(line, style = MaterialTheme.typography.bodySmall)
                    }
                    reading?.status?.let { status ->
                        StatusBadge(status = status, label = status, modifier = Modifier.padding(top = 4.dp))
                    }
                }
            }
        }
    }
}

private val JsonPrimitive.contentOrNull: String?
    get() = if (isString) content else content

private fun raidMemberLines(details: JsonElement?): List<String> {
    val arr = details?.jsonObject?.get("devices") as? JsonArray ?: return emptyList()
    return arr.mapNotNull { el ->
        val obj = el as? JsonObject ?: return@mapNotNull null
        val name = obj["name"]?.jsonPrimitive?.contentOrNull ?: return@mapNotNull null
        val state = obj["state"]?.jsonPrimitive?.contentOrNull ?: "active"
        val label = if (state == "failed") "сбой" else "в строю"
        "· $name ($label)"
    }
}
