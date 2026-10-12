package ru.ferrumnst.sysmon.ui.util

import ru.ferrumnst.sysmon.data.models.DashboardPanel
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.SensorInfo

private val DEFAULT_PANEL_SENSORS = mapOf(
    "system_chart" to listOf("cpu_percent", "ram_used"),
    "disk_bars" to listOf("auto_disks"),
    "net_chart" to listOf("net_rx", "net_tx"),
    "temp_chart" to listOf("cpu_temp", "gpu_temp"),
    "cpu_gauge" to listOf("cpu_percent"),
    "ram_gauge" to listOf("ram_used"),
    "raid_status" to listOf("auto_mdadm"),
)

fun resolvePanelSensorIds(panel: DashboardPanel): List<String> {
    if (panel.sensors.isNotEmpty()) return panel.sensors
    return DEFAULT_PANEL_SENSORS[panel.id] ?: emptyList()
}

fun resolvePanelSensorIds(
    panel: DashboardPanel,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
): List<String> {
    val configured = resolvePanelSensorIds(panel)
    if (configured.contains("auto_disks")) {
        val diskIds = sensors
            .map { it.id }
            .filter { it.startsWith("disk_") || it.startsWith("disk_auto_") }
        if (diskIds.isNotEmpty()) return diskIds
        return liveReadings.keys.filter { it.startsWith("disk_") || it.startsWith("disk_auto_") }
    }
    if (configured.contains("auto_mdadm")) {
        val raidIds = sensors.map { it.id }.filter { it.startsWith("mdadm_") }
        if (raidIds.isNotEmpty()) return raidIds
        return liveReadings.keys.filter { it.startsWith("mdadm_") }
    }
    return configured
}

fun readingForSensor(
    sensorId: String,
    liveReadings: Map<String, MetricReading>,
    sensors: List<SensorInfo>,
): MetricReading? {
    return liveReadings[sensorId] ?: sensors.find { it.id == sensorId }?.current
}

private val QUICK_STAT_SENSORS = setOf("cpu_percent", "ram_used")
private val TEMP_SENSOR_IDS = setOf("cpu_temp", "gpu_temp")

fun isRedundantQuickStatPanel(
    panel: DashboardPanel,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
): Boolean {
    val configured = resolvePanelSensorIds(panel)
    if (configured.isEmpty()) return false
    if (panel.type.equals("gauge", ignoreCase = true) &&
        configured.all { it in QUICK_STAT_SENSORS }
    ) {
        return true
    }
    return false
}

fun panelShouldDisplay(
    panel: DashboardPanel,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
    histories: Map<String, List<MetricReading>> = emptyMap(),
): Boolean {
    if (isRedundantQuickStatPanel(panel, sensors, liveReadings)) return false

    val sensorIds = resolvePanelSensorIds(panel, sensors, liveReadings)
    when (panel.type.lowercase()) {
        "raid", "bar" -> return sensorIds.isNotEmpty()
        "gauge" -> {
            val id = sensorIds.firstOrNull() ?: return false
            return sensors.any { it.id == id } || readingForSensor(id, liveReadings, sensors)?.value != null
        }
        "line" -> {
            if (panel.id == "temp_chart" || sensorIds.all { it in TEMP_SENSOR_IDS }) {
                return sensorIds.any { id ->
                    sensors.any { it.id == id } ||
                        !histories[id].isNullOrEmpty() ||
                        readingForSensor(id, liveReadings, sensors)?.value != null
                }
            }
            return sensorIds.isNotEmpty() && sensorIds.any { id ->
                sensors.any { it.id == id } || readingForSensor(id, liveReadings, sensors)?.value != null
            }
        }
        else -> return sensorIds.isNotEmpty()
    }
}
