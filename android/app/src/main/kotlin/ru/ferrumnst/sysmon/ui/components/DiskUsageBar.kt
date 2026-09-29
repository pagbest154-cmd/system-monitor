package ru.ferrumnst.sysmon.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import ru.ferrumnst.sysmon.ui.util.formatValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.SensorInfo
import ru.ferrumnst.sysmon.ui.theme.SysMonColors

@Composable
fun DiskUsageBar(
    sensor: SensorInfo,
    reading: MetricReading?,
    modifier: Modifier = Modifier,
) {
    val value = reading?.value ?: sensor.current?.value
    val progress by animateFloatAsState(
        targetValue = ((value ?: 0.0) / 100.0).coerceIn(0.0, 1.0).toFloat(),
        label = "diskProgress",
    )
    val color = when {
        progress >= 0.9f -> SysMonColors.Critical
        progress >= 0.75f -> SysMonColors.Warn
        else -> SysMonColors.Accent
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(vertical = 6.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = sensor.name,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f),
            )
            AnimatedMetricValue(
                text = formatValue(value, sensor.unit),
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.SemiBold,
            )
        }
        LinearProgressIndicator(
            progress = { progress },
            modifier = Modifier.fillMaxWidth(),
            color = color,
            trackColor = MaterialTheme.colorScheme.surfaceVariant,
        )
    }
}
