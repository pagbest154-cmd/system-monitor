package ru.ferrumnst.sysmon.ui.screens.hosts

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Computer
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.SystemUpdate
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.components.StatusBadge
import ru.ferrumnst.sysmon.ui.components.SysMonBottomBarClearance
import ru.ferrumnst.sysmon.ui.theme.SysMonColors
import ru.ferrumnst.sysmon.ui.util.formatTimestamp
import ru.ferrumnst.sysmon.ui.util.cpuPercent
import ru.ferrumnst.sysmon.ui.util.formatGbRange
import ru.ferrumnst.sysmon.ui.util.memoryTotalGb
import ru.ferrumnst.sysmon.ui.util.memoryUsedGb
import ru.ferrumnst.sysmon.ui.util.ramPercent
import ru.ferrumnst.sysmon.ui.screens.hosts.AgentUpdateHints
import ru.ferrumnst.sysmon.ui.util.versionAndOsText

@Composable
fun HostsScreen(
    repository: HubRepository,
    onOpenAlerts: (agentId: String, agentName: String) -> Unit = { _, _ -> },
) {
    val vm: HostsViewModel = viewModel(factory = HostsViewModel.Factory(repository))
    val state by vm.uiState.collectAsState()
    val context = LocalContext.current
    var agentToDelete by remember { mutableStateOf<ru.ferrumnst.sysmon.data.models.AgentInfo?>(null) }
    var showUpdateHints by remember { mutableStateOf(false) }

    agentToDelete?.let { agent ->
        AlertDialog(
            onDismissRequest = { agentToDelete = null },
            title = { Text("Удалить хост?") },
            text = {
                Text("Будут удалены метрики и конфигурация агента «${agent.name ?: agent.id}» на hub.")
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        vm.deleteAgent(agent.id)
                        agentToDelete = null
                    },
                ) {
                    Text("Удалить")
                }
            },
            dismissButton = {
                TextButton(onClick = { agentToDelete = null }) {
                    Text("Отмена")
                }
            },
        )
    }

    if (showUpdateHints) {
        AgentUpdateDialog(
            hints = state.agentUpdateHints,
            onDismiss = { showUpdateHints = false },
            onCopy = { text ->
                val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                clipboard.setPrimaryClip(ClipData.newPlainText("agent update", text))
            },
        )
    }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .statusBarsPadding()
                .background(MaterialTheme.colorScheme.surface)
                .padding(horizontal = 16.dp, vertical = 10.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                "Хосты",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface,
            )
            IconButton(onClick = vm::refresh) {
                Icon(Icons.Default.Refresh, contentDescription = "Обновить")
            }
        }

        state.message?.let { msg ->
            Text(
                text = msg,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp),
                color = MaterialTheme.colorScheme.primary,
                style = MaterialTheme.typography.bodyMedium,
            )
        }

        when {
            state.isLoading -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator()
                }
            }
            state.error != null -> {
                Text(
                    text = state.error!!,
                    modifier = Modifier.padding(16.dp),
                    color = MaterialTheme.colorScheme.error,
                )
            }
            state.agents.isEmpty() -> {
                Text(
                    text = "Нет зарегистрированных агентов",
                    modifier = Modifier.padding(16.dp),
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            else -> {
                LazyColumn(
                    contentPadding = PaddingValues(
                        start = 16.dp,
                        end = 16.dp,
                        top = 12.dp,
                        bottom = SysMonBottomBarClearance,
                    ),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    items(state.agents, key = { it.id }) { agent ->
                        HostCard(
                            agent = agent,
                            isSelected = agent.id == state.selectedAgentId,
                            isDeleting = state.deletingAgentId == agent.id,
                            onSelect = { vm.selectAgent(agent.id) },
                            onOpenAlerts = {
                                onOpenAlerts(agent.id, agent.name ?: agent.id)
                            },
                            onDelete = { agentToDelete = agent },
                            onShowUpdate = { showUpdateHints = true },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun AgentUpdateDialog(
    hints: AgentUpdateHints,
    onDismiss: () -> Unit,
    onCopy: (String) -> Unit,
) {
    val body = buildString {
        hints.latestVersion?.let { append("Версия: $it\n\n") }
        hints.aptCommand?.let { append("Linux (apt):\n$it\n\n") }
        hints.debUrl?.let { append("deb: $it\n\n") }
        hints.windowsUrl?.let { append("Windows: $it\n\n") }
        hints.releaseUrl?.let { append("Release: $it") }
    }.trim()
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Обновление агента") },
        text = {
            Text(
                if (body.isNotBlank()) body else "Команда обновления недоступна",
                style = MaterialTheme.typography.bodySmall,
            )
        },
        confirmButton = {
            TextButton(
                onClick = {
                    hints.aptCommand?.let(onCopy)
                    onDismiss()
                },
                enabled = !hints.aptCommand.isNullOrBlank(),
            ) {
                Text("Скопировать apt")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Закрыть")
            }
        },
    )
}

@Composable
private fun HostCard(
    agent: ru.ferrumnst.sysmon.data.models.AgentInfo,
    isSelected: Boolean,
    isDeleting: Boolean,
    onSelect: () -> Unit,
    onOpenAlerts: () -> Unit,
    onDelete: () -> Unit,
    onShowUpdate: () -> Unit,
) {
    val cpu = agent.cpuPercent()
    val ram = agent.ramPercent()
    val online = agent.status == "online"

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onSelect),
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected) {
                SysMonColors.Accent.copy(alpha = 0.12f)
            } else {
                MaterialTheme.colorScheme.surface
            },
            contentColor = MaterialTheme.colorScheme.onSurface,
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top,
            ) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.weight(1f),
                ) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(RoundedCornerShape(10.dp))
                            .background(SysMonColors.Accent.copy(alpha = 0.15f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            Icons.Default.Computer,
                            contentDescription = null,
                            tint = SysMonColors.Accent,
                        )
                    }
                    Column {
                        Text(
                            text = agent.name ?: agent.id,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = MaterialTheme.colorScheme.onSurface,
                        )
                        Text(
                            text = agent.hostname ?: agent.id,
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                        agent.versionAndOsText()?.let { versionOs ->
                            Text(
                                text = versionOs,
                                style = MaterialTheme.typography.bodySmall,
                                color = if (agent.updateAvailable) {
                                    SysMonColors.Warn
                                } else {
                                    MaterialTheme.colorScheme.onSurfaceVariant
                                },
                            )
                        }
                    }
                }
                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    if (agent.updateAvailable) {
                        IconButton(onClick = onShowUpdate) {
                            Icon(
                                Icons.Default.SystemUpdate,
                                contentDescription = "Обновить агента",
                                tint = SysMonColors.Warn,
                            )
                        }
                    }
                    IconButton(onClick = onOpenAlerts) {
                        Icon(
                            Icons.Default.Notifications,
                            contentDescription = "Уведомления",
                            tint = SysMonColors.Accent,
                        )
                    }
                    IconButton(onClick = onDelete, enabled = !isDeleting) {
                        Icon(
                            Icons.Default.Delete,
                            contentDescription = "Удалить хост",
                            tint = MaterialTheme.colorScheme.error,
                        )
                    }
                    if (isSelected) {
                        Icon(
                            Icons.Default.Check,
                            contentDescription = "Выбран",
                            tint = SysMonColors.Accent,
                            modifier = Modifier
                                .padding(8.dp)
                                .size(24.dp),
                        )
                    }
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 12.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                StatusBadge(
                    status = if (online) "online" else "offline",
                    label = if (online) "Online" else "Offline",
                )
                Text(
                    text = formatTimestamp(agent.lastSeen),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }

            if (cpu != null || ram != null) {
                Column(
                    modifier = Modifier.padding(top = 12.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    cpu?.let { value ->
                        MetricMiniBar(label = "CPU", value = value)
                    }
                    ram?.let { value ->
                        MetricMiniBar(
                            label = "RAM",
                            value = value,
                            detail = formatGbRange(agent.memoryUsedGb(), agent.memoryTotalGb()),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun MetricMiniBar(label: String, value: Double, detail: String? = null) {
    val progress = (value / 100.0).coerceIn(0.0, 1.0).toFloat()
    val color = when {
        progress >= 0.9f -> SysMonColors.Critical
        progress >= 0.75f -> SysMonColors.Warn
        else -> SysMonColors.Ok
    }
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier.size(width = 36.dp, height = 16.dp),
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            LinearProgressIndicator(
                progress = { progress },
                modifier = Modifier.weight(1f),
                color = color,
                trackColor = MaterialTheme.colorScheme.surfaceVariant,
            )
            Text(
                text = "${value.toInt()}%",
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface,
            )
        }
        detail?.let { line ->
            Text(
                text = line,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(start = 44.dp),
            )
        }
    }
}
