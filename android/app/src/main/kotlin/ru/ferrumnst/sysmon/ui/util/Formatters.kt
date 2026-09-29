package ru.ferrumnst.sysmon.ui.util

import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.concurrent.TimeUnit

fun formatTimestamp(ts: Double?): String {
    if (ts == null || ts <= 0) return "—"
    val millis = if (ts > 1_000_000_000_000) ts.toLong() else (ts * 1000).toLong()
    val formatter = SimpleDateFormat("HH:mm:ss", Locale.getDefault())
    return formatter.format(Date(millis))
}

fun formatUptime(seconds: Long?): String {
    if (seconds == null || seconds <= 0) return "—"
    val days = TimeUnit.SECONDS.toDays(seconds)
    val hours = TimeUnit.SECONDS.toHours(seconds) % 24
    val minutes = TimeUnit.SECONDS.toMinutes(seconds) % 60
    return when {
        days > 0 -> "${days}д ${hours}ч"
        hours > 0 -> "${hours}ч ${minutes}м"
        else -> "${minutes}м"
    }
}

fun formatValue(value: Double?, unit: String?): String {
    if (value == null) return "—"
    val formatted = if (value >= 100) {
        value.toInt().toString()
    } else if (value >= 10) {
        String.format(Locale.getDefault(), "%.1f", value)
    } else {
        String.format(Locale.getDefault(), "%.2f", value)
    }
    return if (unit.isNullOrBlank()) formatted else "$formatted $unit"
}
