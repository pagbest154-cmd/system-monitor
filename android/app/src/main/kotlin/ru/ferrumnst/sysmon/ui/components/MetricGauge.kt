package ru.ferrumnst.sysmon.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.ui.theme.SysMonColors
import ru.ferrumnst.sysmon.ui.util.formatValue

@Composable
fun MetricGaugeCard(
    title: String,
    value: Double?,
    unit: String?,
    modifier: Modifier = Modifier,
    maxValue: Double = 100.0,
) {
    DashboardCard(title = title, modifier = modifier) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            GaugeArc(
                value = value,
                maxValue = maxValue,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(110.dp),
            )
            AnimatedMetricValue(
                text = formatValue(value, unit),
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold,
            )
        }
    }
}

@Composable
private fun GaugeArc(
    value: Double?,
    maxValue: Double,
    modifier: Modifier = Modifier,
) {
    val target = ((value ?: 0.0) / maxValue).coerceIn(0.0, 1.0).toFloat()
    val progress by animateFloatAsState(targetValue = target, label = "gauge")
    val trackColor = MaterialTheme.colorScheme.surfaceVariant
    val progressColor = when {
        progress >= 0.9f -> SysMonColors.Critical
        progress >= 0.75f -> SysMonColors.Warn
        else -> SysMonColors.Accent
    }

    Canvas(modifier = modifier) {
        val stroke = Stroke(width = 14f, cap = StrokeCap.Round)
        val diameter = size.minDimension * 0.85f
        val topLeft = Offset((size.width - diameter) / 2f, (size.height - diameter) / 2f)
        val arcSize = Size(diameter, diameter)

        drawArc(
            color = trackColor,
            startAngle = 135f,
            sweepAngle = 270f,
            useCenter = false,
            topLeft = topLeft,
            size = arcSize,
            style = stroke,
        )
        drawArc(
            color = progressColor,
            startAngle = 135f,
            sweepAngle = 270f * progress,
            useCenter = false,
            topLeft = topLeft,
            size = arcSize,
            style = stroke,
        )
    }
}
