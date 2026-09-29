package ru.ferrumnst.sysmon.widget

import android.content.Context
import android.content.Intent
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.LocalContext
import androidx.glance.action.clickable
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver
import androidx.glance.appwidget.LinearProgressIndicator
import androidx.glance.appwidget.SizeMode
import androidx.glance.appwidget.action.actionStartActivity
import androidx.glance.appwidget.cornerRadius
import androidx.glance.appwidget.provideContent
import androidx.glance.background
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.Spacer
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.height
import androidx.glance.layout.padding
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ru.ferrumnst.sysmon.MainActivity
import ru.ferrumnst.sysmon.ui.components.formatChartValue
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

class MetricWidget : GlanceAppWidget() {
    override val sizeMode = SizeMode.Exact

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        val config = WidgetConfigStore.load(context, id)
        val data = WidgetDataLoader.load(context, config)
        provideContent {
            WidgetContent(data)
        }
    }
}

class MetricWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = MetricWidget()

    override fun onDeleted(context: Context, appWidgetIds: IntArray) {
        super.onDeleted(context, appWidgetIds)
        appWidgetIds.forEach { widgetId ->
            WidgetConfigStore.delete(context, widgetId)
        }
    }
}

private fun widgetColor(hex: Long): ColorProvider = ColorProvider(Color(hex))

private val Bg = widgetColor(0xFF1A2332)
private val TextPrimary = widgetColor(0xFFE8EDF4)
private val TextMuted = widgetColor(0xFF8B9CB3)
private val Accent = widgetColor(0xFF3B82F6)

private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm")
    .withZone(ZoneId.systemDefault())

@Composable
private fun WidgetContent(data: WidgetDisplayData) {
    val context = LocalContext.current
    val openApp = actionStartActivity(Intent(context, MainActivity::class.java))

    Column(
        modifier = GlanceModifier
            .fillMaxSize()
            .cornerRadius(20.dp)
            .background(Bg)
            .clickable(openApp)
            .padding(14.dp),
    ) {
        if (data.error != null) {
            Text(
                text = "SysMon",
                style = TextStyle(color = Accent, fontWeight = FontWeight.Medium, fontSize = 14.sp),
            )
            Spacer(GlanceModifier.height(8.dp))
            Text(
                text = data.error,
                style = TextStyle(color = TextMuted, fontSize = 13.sp),
            )
        } else {
            Row(
                modifier = GlanceModifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = data.agentName,
                    style = TextStyle(
                        color = TextPrimary,
                        fontWeight = FontWeight.Medium,
                        fontSize = 14.sp,
                    ),
                    modifier = GlanceModifier.defaultWeight(),
                )
                Text(
                    text = timeFormatter.format(Instant.ofEpochMilli(data.updatedAt)),
                    style = TextStyle(color = TextMuted, fontSize = 11.sp),
                )
            }

            Spacer(GlanceModifier.height(10.dp))

            data.metrics.forEachIndexed { index, metric ->
                MetricRow(metric)
                if (index < data.metrics.lastIndex) {
                    Spacer(GlanceModifier.height(8.dp))
                }
            }
        }
    }
}

@Composable
private fun MetricRow(metric: WidgetMetricItem) {
    val valueText = metric.value?.let { formatChartValue(it, metric.unit) } ?: "—"
    val valueColor = statusColor(metric.status)

    Row(
        modifier = GlanceModifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = metric.name,
            style = TextStyle(color = TextMuted, fontSize = 12.sp),
            modifier = GlanceModifier.defaultWeight(),
        )
        Text(
            text = valueText,
            style = TextStyle(color = valueColor, fontWeight = FontWeight.Medium, fontSize = 13.sp),
        )
    }

    if (metric.unit == "%" && metric.value != null) {
        Spacer(GlanceModifier.height(4.dp))
        LinearProgressIndicator(
            progress = (metric.value / 100.0).coerceIn(0.0, 1.0).toFloat(),
            modifier = GlanceModifier.fillMaxWidth(),
            color = valueColor,
        )
    }
}

private fun statusColor(status: String): ColorProvider {
    return when (status) {
        "critical", "error" -> widgetColor(0xFFEF4444)
        "warning" -> widgetColor(0xFFF59E0B)
        "ok" -> widgetColor(0xFF3B82F6)
        else -> widgetColor(0xFF8B9CB3)
    }
}
