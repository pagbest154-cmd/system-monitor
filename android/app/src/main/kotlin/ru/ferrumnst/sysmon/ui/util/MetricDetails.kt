package ru.ferrumnst.sysmon.ui.util

import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import ru.ferrumnst.sysmon.data.models.MetricReading

fun MetricReading.diskVolumeSubtitle(): String? {
    val obj = details?.jsonObject ?: return null
    val used = obj["used_gb"]?.jsonPrimitive?.doubleOrNull
    val total = obj["total_gb"]?.jsonPrimitive?.doubleOrNull
    if (used != null && total != null) {
        return "$used ГБ из $total ГБ"
    }
    return null
}

fun isNetworkThroughputSensor(sensorId: String?, sensorType: String?): Boolean {
    if (sensorType == "system.network_bytes") return true
    return sensorId == "net_rx" || sensorId == "net_tx"
}

const val MAX_NETWORK_CHART_MBPS = 100_000.0
