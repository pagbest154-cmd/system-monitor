package ru.ferrumnst.sysmon.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Memory
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.data.models.GpuInfo
import ru.ferrumnst.sysmon.ui.theme.SysMonColors

@Composable
fun GpuSection(
    gpus: List<GpuInfo>,
    liveGpuTemp: Double? = null,
    modifier: Modifier = Modifier,
) {
    if (gpus.isEmpty()) return

    DashboardCard(
        title = "Видеокарты",
        subtitle = "${gpus.size}",
        modifier = modifier,
    ) {
        gpus.forEachIndexed { index, gpu ->
            GpuCard(
                gpu = gpu,
                liveGpuTemp = if (index == 0) liveGpuTemp else null,
            )
            if (index < gpus.lastIndex) {
                androidx.compose.material3.HorizontalDivider(
                    modifier = Modifier.padding(vertical = 10.dp),
                    color = MaterialTheme.colorScheme.outline.copy(alpha = 0.2f),
                )
            }
        }
    }
}

@Composable
private fun GpuCard(
    gpu: GpuInfo,
    liveGpuTemp: Double?,
) {
    val temperature = liveGpuTemp ?: gpu.temperatureC
    val memoryPercent = gpu.memoryPercent
    val progress by animateFloatAsState(
        targetValue = ((memoryPercent ?: 0.0) / 100.0).coerceIn(0.0, 1.0).toFloat(),
        label = "gpuVramProgress",
    )

    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            androidx.compose.material3.Icon(
                imageVector = Icons.Default.Memory,
                contentDescription = null,
                tint = SysMonColors.Accent,
            )
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = gpu.name ?: "GPU",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurface,
                )
                gpu.videoProcessor?.takeIf { it.isNotBlank() }?.let { processor ->
                    Text(
                        text = processor,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            }
            temperature?.let { temp ->
                AnimatedMetricValue(
                    text = "${formatGpuNumber(temp)}°C",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                )
            }
        }

        if (gpu.memoryUsedGb != null && gpu.memoryTotalGb != null) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    text = "VRAM",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                Text(
                    text = buildString {
                        append("${formatGpuNumber(gpu.memoryUsedGb)} / ${formatGpuNumber(gpu.memoryTotalGb)} ГБ")
                        memoryPercent?.let { append(" · ${formatGpuNumber(it)}%") }
                    },
                    style = MaterialTheme.typography.bodySmall,
                    fontWeight = FontWeight.Medium,
                )
            }
            if (memoryPercent != null) {
                LinearProgressIndicator(
                    progress = { progress },
                    modifier = Modifier.fillMaxWidth(),
                    color = SysMonColors.ChartPurple,
                    trackColor = MaterialTheme.colorScheme.surfaceVariant,
                )
            }
        }

        val meta = buildList {
            gpu.vendor?.takeIf { it.isNotBlank() }?.let { add(it) }
            gpu.driverVersion?.takeIf { it.isNotBlank() }?.let { add("Драйвер $it") }
            gpu.utilizationPercent?.let { add("Загрузка ${formatGpuNumber(it)}%") }
            gpu.cudaVersion?.takeIf { it.isNotBlank() }?.let { add("CUDA $it") }
            gpu.resolution?.takeIf { it.isNotBlank() }?.let { add(gpu.resolution) }
        }
        if (meta.isNotEmpty()) {
            Text(
                text = meta.joinToString(" · "),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

private fun formatGpuNumber(value: Double): String {
    return if (value % 1.0 == 0.0) {
        value.toLong().toString()
    } else {
        String.format("%.1f", value)
    }
}
