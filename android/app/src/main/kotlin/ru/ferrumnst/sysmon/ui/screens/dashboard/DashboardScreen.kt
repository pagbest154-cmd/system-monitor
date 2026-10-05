package ru.ferrumnst.sysmon.ui.screens.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Memory
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Speed
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.MenuAnchorType
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.DashboardPanel
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.SensorInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.components.AnimatedMetricValue
import ru.ferrumnst.sysmon.ui.components.ChartSeries
import ru.ferrumnst.sysmon.ui.components.DashboardCard
import ru.ferrumnst.sysmon.ui.components.DiskUsageBar
import ru.ferrumnst.sysmon.ui.components.GpuSection
import ru.ferrumnst.sysmon.ui.components.MetricGaugeCard
import ru.ferrumnst.sysmon.ui.components.MultiLineMetricChart
import ru.ferrumnst.sysmon.ui.components.QuickStatCard
import ru.ferrumnst.sysmon.ui.components.StatusBadge
import ru.ferrumnst.sysmon.ui.components.SysMonBottomBarClearance
import ru.ferrumnst.sysmon.ui.components.SystemResourcesCard
import ru.ferrumnst.sysmon.ui.components.chartColorForIndex
import ru.ferrumnst.sysmon.ui.theme.SysMonColors
import ru.ferrumnst.sysmon.ui.util.chartPointsFromHistory
import ru.ferrumnst.sysmon.ui.util.formatGbRange
import ru.ferrumnst.sysmon.ui.util.formatTimestamp
import ru.ferrumnst.sysmon.ui.util.formatValue
import ru.ferrumnst.sysmon.ui.util.readingForSensor
import ru.ferrumnst.sysmon.ui.components.RaidPanelCard
import ru.ferrumnst.sysmon.ui.util.resolvePanelSensorIds

private val periods = listOf(
    "1h" to "1ч",
    "6h" to "6ч",
    "1d" to "1д",
    "1w" to "1н",
)

private val quickStatSensors = setOf("cpu_percent", "ram_used")

private fun isQuickStatsDuplicatePanel(
    panel: DashboardPanel,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
): Boolean {
    val sensorIds = resolvePanelSensorIds(panel, sensors, liveReadings)
    return sensorIds.isNotEmpty() && sensorIds.all { it in quickStatSensors }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardScreen(
    repository: HubRepository,
) {
    val vm: DashboardViewModel = viewModel(factory = DashboardViewModel.Factory(repository))
    val state by vm.uiState.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        DashboardHeader(
            state = state,
            onRefresh = { vm.refresh(showFullLoading = false) },
            onAgentSelected = vm::selectAgent,
            onPeriodSelected = vm::setPeriod,
        )

        PullToRefreshBox(
            isRefreshing = state.isRefreshing,
            onRefresh = { vm.refresh(showFullLoading = false) },
            modifier = Modifier.fillMaxSize(),
        ) {
            when {
                state.isLoading -> {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                    }
                }
                state.error != null -> {
                    Text(
                        text = state.error!!,
                        modifier = Modifier.padding(16.dp),
                        color = MaterialTheme.colorScheme.error,
                    )
                }
                state.mode == "hub" && state.agents.isEmpty() -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(16.dp),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = "Нет зарегистрированных хостов",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
                else -> {
                    LazyColumn(
                        contentPadding = PaddingValues(
                            start = 16.dp,
                            end = 16.dp,
                            top = 8.dp,
                            bottom = SysMonBottomBarClearance,
                        ),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                    item {
                        QuickStatsRow(
                            sensors = state.sensors,
                            liveReadings = state.liveReadings,
                            systemInfo = state.systemInfo,
                        )
                    }

                    state.systemInfo?.let { systemInfo ->
                        item {
                            SystemResourcesCard(systemInfo = systemInfo)
                        }
                    }

                    state.systemInfo?.gpus?.takeIf { it.isNotEmpty() }?.let { gpus ->
                        item {
                            GpuSection(
                                gpus = gpus,
                                liveGpuTemp = state.liveReadings["gpu_temp"]?.value
                                    ?: state.sensors.find { it.id == "gpu_temp" }?.current?.value,
                            )
                        }
                    }

                    val gaugePanels = state.panels.filter { panel ->
                        panel.type == "gauge" &&
                            resolvePanelSensorIds(panel, state.sensors, state.liveReadings)
                                .none { it in quickStatSensors }
                    }
                    if (gaugePanels.isNotEmpty()) {
                        item {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                gaugePanels.take(2).forEach { panel ->
                                    PanelCard(
                                        panel = panel,
                                        period = state.period,
                                        sensors = state.sensors,
                                        liveReadings = state.liveReadings,
                                        histories = state.panelHistories[panel.id]?.histories ?: emptyMap(),
                                        liveBuffers = state.liveBuffers,
                                        modifier = Modifier.weight(1f),
                                    )
                                }
                            }
                        }
                    }

                    items(
                        state.panels.filter { panel ->
                            (panel.type == "line" || panel.type == "bar" || panel.type == "raid") &&
                                !isQuickStatsDuplicatePanel(panel, state.sensors, state.liveReadings)
                        },
                    ) { panel ->
                        PanelCard(
                            panel = panel,
                            period = state.period,
                            sensors = state.sensors,
                            liveReadings = state.liveReadings,
                            histories = state.panelHistories[panel.id]?.histories ?: emptyMap(),
                            liveBuffers = state.liveBuffers,
                        )
                    }

                    item {
                        SensorsList(
                            sensors = state.sensors,
                            liveReadings = state.liveReadings,
                            systemInfo = state.systemInfo,
                        )
                    }

                        item { Spacer(modifier = Modifier.height(8.dp)) }
                    }
                }
            }
        }
    }
}

@Composable
private fun QuickStatsRow(
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
    systemInfo: AgentSystemInfo?,
) {
    val cpu = sensors.find { it.id == "cpu_percent" }
    val ram = sensors.find { it.id == "ram_used" }
    if (cpu == null && ram == null) return

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        cpu?.let { sensor ->
            val cores = systemInfo?.cpu?.let { cpuInfo ->
                when {
                    cpuInfo.coresPhysical != null && cpuInfo.coresLogical != null ->
                        "${cpuInfo.coresPhysical} физ. / ${cpuInfo.coresLogical} лог."
                    cpuInfo.coresLogical != null -> "${cpuInfo.coresLogical} ядер"
                    else -> null
                }
            }
            QuickStatCard(
                title = "CPU",
                value = (liveReadings[sensor.id] ?: sensor.current)?.value,
                unit = sensor.unit,
                icon = Icons.Default.Speed,
                modifier = Modifier.weight(1f),
                detail = cores,
            )
        }
        ram?.let { sensor ->
            val memory = systemInfo?.memory
            QuickStatCard(
                title = "Память",
                value = (liveReadings[sensor.id] ?: sensor.current)?.value,
                unit = sensor.unit,
                icon = Icons.Default.Memory,
                modifier = Modifier.weight(1f),
                detail = formatGbRange(memory?.usedGb, memory?.totalGb),
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun DashboardHeader(
    state: DashboardUiState,
    onRefresh: () -> Unit,
    onAgentSelected: (String) -> Unit,
    onPeriodSelected: (String) -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .statusBarsPadding()
            .background(MaterialTheme.colorScheme.surface)
            .padding(horizontal = 16.dp, vertical = 10.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = state.selectedAgentName.ifBlank { "Мониторинг" },
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurface,
                )
                Text(
                    text = "Обновлено: ${formatTimestamp(state.lastUpdateTs)}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            StatusBadge(
                status = if (state.isLive) "live" else "offline",
                label = if (state.isLive) "Live" else "Offline",
            )
            IconButton(onClick = onRefresh) {
                Icon(
                    Icons.Default.Refresh,
                    contentDescription = "Обновить",
                    tint = MaterialTheme.colorScheme.onSurface,
                )
            }
        }

        if (state.agents.isNotEmpty()) {
            var expanded by remember { mutableStateOf(false) }
            ExposedDropdownMenuBox(
                expanded = expanded,
                onExpandedChange = { expanded = it },
            ) {
                OutlinedTextField(
                    value = state.selectedAgentName.ifBlank { "Выберите хост" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Хост") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor(MenuAnchorType.PrimaryNotEditable),
                )
                ExposedDropdownMenu(
                    expanded = expanded,
                    onDismissRequest = { expanded = false },
                ) {
                    state.agents.forEach { (id, name) ->
                        DropdownMenuItem(
                            text = { Text(name) },
                            onClick = {
                                expanded = false
                                onAgentSelected(id)
                            },
                        )
                    }
                }
            }
        }

        LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            items(periods) { (value, label) ->
                FilterChip(
                    selected = state.period == value,
                    onClick = { onPeriodSelected(value) },
                    label = { Text(label) },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = SysMonColors.Accent.copy(alpha = 0.2f),
                        selectedLabelColor = SysMonColors.Accent,
                    ),
                )
            }
        }
    }
}

@Composable
private fun PanelCard(
    panel: DashboardPanel,
    period: String,
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
    histories: Map<String, List<MetricReading>>,
    liveBuffers: Map<String, List<MetricReading>> = emptyMap(),
    modifier: Modifier = Modifier,
) {
    when (panel.type.lowercase()) {
        "gauge" -> {
            resolvePanelSensorIds(panel, sensors, liveReadings).firstOrNull()?.let { sensorId ->
                val meta = sensors.find { it.id == sensorId }
                val reading = liveReadings[sensorId] ?: meta?.current
                MetricGaugeCard(
                    title = panel.title,
                    value = reading?.value,
                    unit = meta?.unit,
                    modifier = modifier,
                )
            }
        }
        "line" -> {
            val sensorIds = resolvePanelSensorIds(panel, sensors, liveReadings)
            val chartSeries = sensorIds.mapIndexedNotNull { index, sensorId ->
                val meta = sensors.find { it.id == sensorId }
                val live = readingForSensor(sensorId, liveReadings, sensors)
                val points = chartPointsFromHistory(
                    history = histories[sensorId] ?: emptyList(),
                    live = live,
                    buffered = liveBuffers[sensorId] ?: emptyList(),
                    sensorId = sensorId,
                    sensorType = meta?.type,
                )
                if (points.isEmpty()) return@mapIndexedNotNull null
                ChartSeries(
                    label = meta?.name ?: sensorId,
                    color = chartColorForIndex(index),
                    points = points,
                    unit = meta?.unit,
                )
            }
            DashboardCard(
                title = panel.title,
                subtitle = period,
                modifier = modifier,
            ) {
                MultiLineMetricChart(series = chartSeries)
            }
        }
        "bar" -> {
            val sensorIds = resolvePanelSensorIds(panel, sensors, liveReadings)
            val panelSensors = sensorIds.mapNotNull { sensorId ->
                sensors.find { it.id == sensorId }
            }
            DashboardCard(title = panel.title, modifier = modifier) {
                if (panelSensors.isEmpty()) {
                    Text(
                        text = "Нет данных",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(top = 8.dp),
                    )
                } else {
                    panelSensors.forEach { sensor ->
                        DiskUsageBar(
                            sensor = sensor,
                            reading = liveReadings[sensor.id] ?: sensor.current,
                        )
                    }
                }
            }
        }
        "raid" -> {
            val sensorIds = resolvePanelSensorIds(panel, sensors, liveReadings)
            RaidPanelCard(
                title = panel.title,
                sensorIds = sensorIds,
                sensors = sensors,
                liveReadings = liveReadings,
                modifier = modifier,
            )
        }
    }
}

@Composable
private fun SensorsList(
    sensors: List<SensorInfo>,
    liveReadings: Map<String, MetricReading>,
    systemInfo: AgentSystemInfo?,
) {
    if (sensors.isEmpty()) return

    DashboardCard(title = "Датчики", subtitle = "${sensors.size} активных") {
        sensors.forEach { sensor ->
            val reading = liveReadings[sensor.id] ?: sensor.current
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(sensor.name, style = MaterialTheme.typography.bodyMedium)
                    val subtitle = when {
                        sensor.id == "ram_used" -> formatGbRange(
                            systemInfo?.memory?.usedGb,
                            systemInfo?.memory?.totalGb,
                        )
                        sensor.type == "system.mdadm_status" -> {
                            val d = reading?.details
                            val level = d?.jsonObject?.get("raid_level")?.jsonPrimitive?.content
                            val state = d?.jsonObject?.get("state")?.jsonPrimitive?.content
                            listOfNotNull(level, state).joinToString(" · ").ifBlank { sensor.type }
                        }
                        else -> sensor.type ?: sensor.id
                    }
                    Text(
                        subtitle ?: (sensor.type ?: sensor.id),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                    Column(horizontalAlignment = Alignment.End) {
                        AnimatedMetricValue(
                            text = formatValue(reading?.value, sensor.unit),
                            style = MaterialTheme.typography.bodyLarge,
                            fontWeight = FontWeight.SemiBold,
                        )
                    reading?.status?.let { status ->
                        StatusBadge(status = status, label = status)
                    }
                }
            }
        }
    }
}
