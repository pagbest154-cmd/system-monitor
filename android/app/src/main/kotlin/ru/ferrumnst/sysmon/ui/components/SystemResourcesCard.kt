package ru.ferrumnst.sysmon.ui.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.SystemMemoryInfo
import ru.ferrumnst.sysmon.data.models.SystemSwapInfo
import ru.ferrumnst.sysmon.ui.theme.SysMonColors
import ru.ferrumnst.sysmon.ui.util.displayLine
import ru.ferrumnst.sysmon.ui.util.freeLine

@Composable
fun SystemResourcesCard(
    systemInfo: AgentSystemInfo,
    modifier: Modifier = Modifier,
) {
    val memory = systemInfo.memory
    val swap = systemInfo.swap?.takeIf { (it.totalGb ?: 0.0) > 0.0 }
    if (memory == null && swap == null) return

    DashboardCard(title = "Память", modifier = modifier) {
        Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
            memory?.let { ResourceUsageRow(label = "RAM", memory = it) }
            swap?.let { SwapUsageRow(swap = it) }
        }
    }
}

@Composable
private fun ResourceUsageRow(
    label: String,
    memory: SystemMemoryInfo,
) {
    val percent = memory.percent ?: 0.0
    val progress by animateFloatAsState(
        targetValue = (percent / 100.0).coerceIn(0.0, 1.0).toFloat(),
        label = "memoryProgress",
    )
    val color = when {
        progress >= 0.9f -> SysMonColors.Critical
        progress >= 0.75f -> SysMonColors.Warn
        else -> SysMonColors.Ok
    }

    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
            )
            Text(
                text = memory.displayLine() ?: "—",
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
        memory.freeLine()?.let { free ->
            Text(
                text = free,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun SwapUsageRow(swap: SystemSwapInfo) {
    val percent = swap.percent ?: 0.0
    val progress by animateFloatAsState(
        targetValue = (percent / 100.0).coerceIn(0.0, 1.0).toFloat(),
        label = "swapProgress",
    )

    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "Swap",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
            )
            Text(
                text = swap.displayLine() ?: "—",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.SemiBold,
            )
        }
        LinearProgressIndicator(
            progress = { progress },
            modifier = Modifier.fillMaxWidth(),
            color = SysMonColors.Accent,
            trackColor = MaterialTheme.colorScheme.surfaceVariant,
        )
    }
}
