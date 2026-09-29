package ru.ferrumnst.sysmon.widget

import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.GpuInfo
import ru.ferrumnst.sysmon.data.models.SensorInfo

data class WidgetSensorOption(
    val id: String,
    val name: String,
    val unit: String?,
    val group: String? = null,
)

object WidgetSensorCatalog {
    private val gpuMetricPattern = Regex("^gpu_(\\d+)_(util|vram|temp)$")

    fun buildOptions(sensors: List<SensorInfo>, system: AgentSystemInfo): List<WidgetSensorOption> {
        val options = mutableListOf<WidgetSensorOption>()

        sensors.filter { it.enabled }.forEach { sensor ->
            options += WidgetSensorOption(
                id = sensor.id,
                name = sensor.name,
                unit = sensor.unit,
            )
        }

        system.gpus.forEachIndexed { index, gpu ->
            val label = gpu.name?.takeIf { it.isNotBlank() } ?: "GPU ${index + 1}"
            if (gpu.utilizationPercent != null) {
                options += WidgetSensorOption(
                    id = gpuMetricId(index, "util"),
                    name = "$label — загрузка",
                    unit = "%",
                    group = "gpu",
                )
            }
            if (gpu.memoryPercent != null || hasVramData(gpu)) {
                options += WidgetSensorOption(
                    id = gpuMetricId(index, "vram"),
                    name = "$label — VRAM",
                    unit = "%",
                    group = "gpu",
                )
            }
            if (gpu.temperatureC != null) {
                options += WidgetSensorOption(
                    id = gpuMetricId(index, "temp"),
                    name = "$label — температура",
                    unit = "°C",
                    group = "gpu",
                )
            }
        }

        return options
    }

    fun resolveMetric(
        sensorId: String,
        sensors: List<SensorInfo>,
        gpus: List<GpuInfo>,
        fallbackName: String? = null,
    ): WidgetMetricItem? {
        val gpuMatch = gpuMetricPattern.matchEntire(sensorId)
        if (gpuMatch != null) {
            val index = gpuMatch.groupValues[1].toIntOrNull() ?: return null
            val field = gpuMatch.groupValues[2]
            val gpu = gpus.getOrNull(index) ?: return null
            val label = gpu.name?.takeIf { it.isNotBlank() } ?: "GPU ${index + 1}"
            return when (field) {
                "util" -> WidgetMetricItem(
                    id = sensorId,
                    name = "$label — загрузка",
                    value = gpu.utilizationPercent,
                    unit = "%",
                    status = statusForPercent(gpu.utilizationPercent),
                )
                "vram" -> {
                    val percent = gpu.memoryPercent ?: vramPercent(gpu)
                    WidgetMetricItem(
                        id = sensorId,
                        name = "$label — VRAM",
                        value = percent,
                        unit = "%",
                        status = statusForPercent(percent),
                    )
                }
                "temp" -> WidgetMetricItem(
                    id = sensorId,
                    name = "$label — температура",
                    value = gpu.temperatureC,
                    unit = "°C",
                    status = statusForTemp(gpu.temperatureC),
                )
                else -> null
            }
        }

        val sensor = sensors.find { it.id == sensorId }
        if (sensor != null) {
            return WidgetMetricItem(
                id = sensor.id,
                name = sensor.name,
                value = sensor.current?.value,
                unit = sensor.unit,
                status = sensor.current?.status ?: "unknown",
            )
        }

        if (fallbackName != null) {
            return WidgetMetricItem(
                id = sensorId,
                name = fallbackName,
                value = null,
                unit = null,
                status = "unknown",
            )
        }

        return null
    }

    fun displayName(sensorId: String, options: List<WidgetSensorOption>): String? {
        return options.find { it.id == sensorId }?.name
    }

    private fun gpuMetricId(index: Int, field: String) = "gpu_${index}_$field"

    private fun hasVramData(gpu: GpuInfo): Boolean {
        return gpu.memoryUsedGb != null && gpu.memoryTotalGb != null && gpu.memoryTotalGb > 0
    }

    private fun vramPercent(gpu: GpuInfo): Double? {
        val used = gpu.memoryUsedGb
        val total = gpu.memoryTotalGb
        if (used == null || total == null || total <= 0) return null
        return used / total * 100.0
    }

    private fun statusForPercent(value: Double?): String {
        if (value == null) return "unknown"
        return when {
            value >= 95 -> "critical"
            value >= 85 -> "warning"
            else -> "ok"
        }
    }

    private fun statusForTemp(value: Double?): String {
        if (value == null) return "unknown"
        return when {
            value >= 85 -> "critical"
            value >= 75 -> "warning"
            else -> "ok"
        }
    }
}
