package ru.ferrumnst.sysmon.ui.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.ui.theme.SysMonColors
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import kotlin.math.abs
import kotlin.math.roundToInt

data class ChartSeries(
    val label: String,
    val points: List<MetricReading>,
    val color: Color,
    val unit: String? = null,
)

private data class PlottedSeries(
    val label: String,
    val color: Color,
    val unit: String?,
    val points: List<MetricReading>,
)

private data class ChartSelection(
    val ts: Double,
    val crossX: Float,
    val entries: List<SelectionEntry>,
)

private data class SelectionEntry(
    val label: String,
    val color: Color,
    val value: Double,
    val unit: String?,
    val x: Float,
    val y: Float,
)

private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm:ss")
    .withZone(ZoneId.systemDefault())

@Composable
fun MultiLineMetricChart(
    series: List<ChartSeries>,
    modifier: Modifier = Modifier,
    height: Int = 200,
) {
    val plotted = remember(series) {
        series.mapNotNull { item ->
            val points = item.points.filter { it.value != null && !it.value!!.isNaN() && it.ts != null }
            if (points.isEmpty()) null else {
                PlottedSeries(
                    label = item.label,
                    color = item.color,
                    unit = item.unit,
                    points = points,
                )
            }
        }
    }

    if (plotted.isEmpty()) {
        Text(
            text = "Нет данных за выбранный период",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = modifier.padding(top = 12.dp),
        )
        return
    }

    val allValues = plotted.flatMap { it.points.mapNotNull { it.value } }
    val minValue = allValues.minOrNull() ?: 0.0
    val maxValue = allValues.maxOrNull() ?: 1.0
    val valueRange = (maxValue - minValue).takeIf { it > 0.0 } ?: 1.0

    val minTs = plotted.minOf { it.points.minOf { point -> point.ts!! } }
    val maxTs = plotted.maxOf { it.points.maxOf { point -> point.ts!! } }
    val timeRange = (maxTs - minTs).takeIf { it > 0.0 } ?: 1.0

    var chartWidthPx by remember { mutableStateOf(0f) }
    var selection by remember(plotted) { mutableStateOf<ChartSelection?>(null) }

    val gridColor = MaterialTheme.colorScheme.outline.copy(alpha = 0.25f)
    val crosshairColor = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.35f)
    val density = LocalDensity.current

    Column(modifier = modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(height.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(
                modifier = Modifier.width(40.dp),
                verticalArrangement = Arrangement.SpaceBetween,
                horizontalAlignment = Alignment.End,
            ) {
                Text(
                    text = formatAxisValue(maxValue),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                Spacer(modifier = Modifier.weight(1f))
                Text(
                    text = formatAxisValue(minValue),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }

            Box(
                modifier = Modifier
                    .weight(1f)
                    .height(height.dp),
            ) {
                Canvas(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(height.dp)
                        .onSizeChanged { chartWidthPx = it.width.toFloat() }
                        .pointerInput(plotted, minTs, timeRange) {
                            fun updateSelection(x: Float, width: Float, chartHeight: Float) {
                                if (width <= 0f) return
                                val chartLeft = 8f
                                val chartWidth = width - 16f
                                val fraction = ((x - chartLeft) / chartWidth).coerceIn(0f, 1f)
                                val ts = minTs + fraction * timeRange
                                selection = buildSelection(
                                    plotted = plotted,
                                    ts = ts,
                                    minTs = minTs,
                                    timeRange = timeRange,
                                    minValue = minValue,
                                    valueRange = valueRange,
                                    chartLeft = chartLeft,
                                    chartWidth = chartWidth,
                                    chartTop = 8f,
                                    chartBottom = chartHeight - 8f,
                                )
                            }

                            detectTapGestures { offset ->
                                updateSelection(offset.x, size.width.toFloat(), size.height.toFloat())
                            }
                            detectDragGestures(
                                onDragStart = { offset ->
                                    updateSelection(offset.x, size.width.toFloat(), size.height.toFloat())
                                },
                                onDrag = { change, _ ->
                                    change.consume()
                                    updateSelection(change.position.x, size.width.toFloat(), size.height.toFloat())
                                },
                            )
                        },
                ) {
                    val chartLeft = 8f
                    val chartRight = size.width - 8f
                    val chartTop = 8f
                    val chartBottom = size.height - 8f
                    val chartWidth = chartRight - chartLeft
                    val chartHeight = chartBottom - chartTop

                    repeat(5) { index ->
                        val y = chartTop + chartHeight * index / 4f
                        drawLine(
                            color = gridColor,
                            start = Offset(chartLeft, y),
                            end = Offset(chartRight, y),
                            strokeWidth = 1f,
                            pathEffect = PathEffect.dashPathEffect(floatArrayOf(6f, 8f)),
                        )
                    }

                    plotted.forEach { item ->
                        val linePath = Path()
                        val areaPath = Path()
                        var started = false

                        item.points.forEach { point ->
                            val value = point.value ?: return@forEach
                            val ts = point.ts ?: return@forEach
                            val x = chartLeft + ((ts - minTs) / timeRange * chartWidth).toFloat()
                            val y = chartBottom - ((value - minValue) / valueRange * chartHeight).toFloat()
                            if (!started) {
                                linePath.moveTo(x, y)
                                areaPath.moveTo(x, chartBottom)
                                areaPath.lineTo(x, y)
                                started = true
                            } else {
                                linePath.lineTo(x, y)
                                areaPath.lineTo(x, y)
                            }
                        }

                        if (!started) return@forEach

                        areaPath.lineTo(chartLeft + chartWidth, chartBottom)
                        areaPath.close()

                        drawPath(
                            path = areaPath,
                            brush = Brush.verticalGradient(
                                colors = listOf(item.color.copy(alpha = 0.22f), Color.Transparent),
                                startY = chartTop,
                                endY = chartBottom,
                            ),
                        )
                        drawPath(
                            path = linePath,
                            color = item.color,
                            style = Stroke(width = 2.5f, cap = StrokeCap.Round),
                        )
                    }

                    selection?.let { current ->
                        val crossX = current.crossX
                        drawLine(
                            color = crosshairColor,
                            start = Offset(crossX, chartTop),
                            end = Offset(crossX, chartBottom),
                            strokeWidth = 1.5f,
                        )
                        current.entries.forEach { entry ->
                            drawCircle(
                                color = entry.color,
                                radius = 5f,
                                center = Offset(entry.x, entry.y),
                            )
                            drawCircle(
                                color = Color.White,
                                radius = 2f,
                                center = Offset(entry.x, entry.y),
                            )
                        }
                    }
                }

                selection?.let { current ->
                    val tooltipWidthPx = with(density) { 168.dp.toPx() }
                    val xOffsetPx = (current.crossX - tooltipWidthPx / 2f)
                        .coerceIn(0f, (chartWidthPx - tooltipWidthPx).coerceAtLeast(0f))

                    ChartTooltip(
                        selection = current,
                        modifier = Modifier.offset { IntOffset(xOffsetPx.roundToInt(), 0) },
                    )
                }
            }
        }

        Text(
            text = "Проведите по графику для просмотра значений",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
            modifier = Modifier.padding(top = 4.dp, start = 40.dp),
        )

        if (plotted.size > 1) {
            ChartLegend(plotted.map { ChartSeries(it.label, it.points, it.color, it.unit) })
        }
    }
}

@Composable
private fun ChartTooltip(
    selection: ChartSelection,
    modifier: Modifier = Modifier,
) {
    Surface(
        modifier = modifier.width(168.dp),
        shape = RoundedCornerShape(10.dp),
        color = MaterialTheme.colorScheme.surfaceContainerHigh,
        shadowElevation = 4.dp,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 10.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(
                text = formatChartTime(selection.ts),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            selection.entries.forEach { entry ->
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .background(entry.color, CircleShape),
                    )
                    Text(
                        text = "${entry.label}: ${formatChartValue(entry.value, entry.unit)}",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurface,
                    )
                }
            }
        }
    }
}

@Composable
private fun ChartLegend(series: List<ChartSeries>) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = 8.dp, start = 40.dp),
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        series.forEach { item ->
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Box(
                    modifier = Modifier
                        .size(8.dp)
                        .background(item.color, CircleShape),
                )
                Text(
                    text = item.label,
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }
    }
}

private fun buildSelection(
    plotted: List<PlottedSeries>,
    ts: Double,
    minTs: Double,
    timeRange: Double,
    minValue: Double,
    valueRange: Double,
    chartLeft: Float,
    chartWidth: Float,
    chartTop: Float,
    chartBottom: Float,
): ChartSelection {
    val crossX = chartLeft + ((ts - minTs) / timeRange * chartWidth).toFloat()
    val entries = plotted.mapNotNull { series ->
        val point = series.points.minByOrNull { abs((it.ts ?: 0.0) - ts) } ?: return@mapNotNull null
        val value = point.value ?: return@mapNotNull null
        val pointTs = point.ts ?: return@mapNotNull null
        val x = chartLeft + ((pointTs - minTs) / timeRange * chartWidth).toFloat()
        val y = chartBottom - ((value - minValue) / valueRange * (chartBottom - chartTop)).toFloat()
        SelectionEntry(
            label = series.label,
            color = series.color,
            value = value,
            unit = series.unit,
            x = x,
            y = y,
        )
    }
    val displayTs = plotted
        .flatMap { it.points }
        .minByOrNull { abs((it.ts ?: 0.0) - ts) }
        ?.ts ?: ts
    return ChartSelection(ts = displayTs, crossX = crossX, entries = entries)
}

private fun formatChartTime(ts: Double): String {
    return timeFormatter.format(Instant.ofEpochMilli((ts * 1000).toLong()))
}

private fun formatAxisValue(value: Double): String {
    return when {
        abs(value) >= 100 -> value.roundToInt().toString()
        abs(value) >= 10 -> String.format("%.0f", value)
        abs(value) >= 1 -> String.format("%.1f", value)
        else -> String.format("%.2f", value)
    }
}

fun formatChartValue(value: Double, unit: String?): String {
    val formatted = when {
        unit == "%" -> String.format("%.1f", value)
        abs(value) < 1 -> String.format("%.3f", value)
        abs(value) < 10 -> String.format("%.2f", value)
        abs(value) < 100 -> String.format("%.1f", value)
        else -> value.roundToInt().toString()
    }
    return if (!unit.isNullOrBlank()) "$formatted $unit" else formatted
}

fun chartColorForIndex(index: Int): Color {
    val palette = listOf(
        SysMonColors.ChartBlue,
        SysMonColors.ChartGreen,
        SysMonColors.ChartOrange,
        SysMonColors.ChartPurple,
    )
    return palette[index % palette.size]
}
