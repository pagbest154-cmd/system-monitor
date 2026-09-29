package ru.ferrumnst.sysmon.ui.util

import ru.ferrumnst.sysmon.data.models.MetricReading

fun chartPointsFromHistory(
    history: List<MetricReading>,
    live: MetricReading?,
    buffered: List<MetricReading> = emptyList(),
): List<MetricReading> {
    val validHistory = history.filter { it.value != null && !it.value.isNaN() }
    if (validHistory.isNotEmpty()) return validHistory

    val validBuffered = buffered.filter { it.value != null && !it.value.isNaN() }
    if (validBuffered.isNotEmpty()) return validBuffered

    val reading = live ?: return emptyList()
    val value = reading.value ?: return emptyList()
    if (value.isNaN()) return emptyList()

    val ts = reading.ts ?: (System.currentTimeMillis() / 1000.0)
    return listOf(
        reading.copy(value = value, ts = ts),
    )
}
