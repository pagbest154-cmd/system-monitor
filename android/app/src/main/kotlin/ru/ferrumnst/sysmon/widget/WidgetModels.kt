package ru.ferrumnst.sysmon.widget

data class WidgetConfig(
    val agentId: String,
    val agentName: String,
    val sensorIds: List<String>,
)

data class WidgetMetricItem(
    val id: String,
    val name: String,
    val value: Double?,
    val unit: String?,
    val status: String,
)

data class WidgetDisplayData(
    val agentName: String = "",
    val metrics: List<WidgetMetricItem> = emptyList(),
    val updatedAt: Long = 0L,
    val error: String? = null,
) {
    companion object {
        fun error(message: String) = WidgetDisplayData(error = message)
    }
}
