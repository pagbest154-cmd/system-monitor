package ru.ferrumnst.sysmon.ui.screens.hosts

import android.Manifest
import android.app.Application
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.components.DashboardCard
import ru.ferrumnst.sysmon.ui.theme.SysMonColors

@Composable
fun HostAlertsScreen(
    repository: HubRepository,
    agentId: String,
    agentName: String,
    onBack: () -> Unit,
) {
    val app = LocalContext.current.applicationContext as Application
    val vm: HostAlertsViewModel = viewModel(
        factory = HostAlertsViewModel.Factory(app, repository, agentId, agentName),
    )
    val state by vm.uiState.collectAsState()

    val permissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { }

    fun requestPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .statusBarsPadding()
                .background(MaterialTheme.colorScheme.surface)
                .padding(start = 4.dp, end = 16.dp, top = 8.dp, bottom = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Назад")
            }
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    "Уведомления",
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                )
                Text(
                    state.agentName,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            Icon(
                Icons.Default.Notifications,
                contentDescription = null,
                tint = if (state.alertsEnabled) SysMonColors.Accent else MaterialTheme.colorScheme.onSurfaceVariant,
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
            else -> {
                LazyColumn(
                    modifier = Modifier.weight(1f),
                    contentPadding = PaddingValues(
                        start = 16.dp,
                        end = 16.dp,
                        top = 12.dp,
                        bottom = 12.dp,
                    ),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    item {
                        DashboardCard(title = "Алерты") {
                            SettingSwitchRow(
                                title = "Уведомления для хоста",
                                subtitle = "Push при превышении порогов и offline",
                                checked = state.alertsEnabled,
                                onCheckedChange = {
                                    if (it) requestPermissionIfNeeded()
                                    vm.setAlertsEnabled(it)
                                },
                            )
                        }
                    }

                    if (state.alertsEnabled) {
                        item {
                            DashboardCard(
                                title = "ntfy",
                                subtitle = "Topic — секрет подписки, не публикуйте",
                            ) {
                                InfoValueRow(
                                    label = "Сервер",
                                    value = state.ntfyBaseUrl.ifBlank { "https://ntfy.sh" },
                                )
                                InfoValueRow(
                                    label = "Topic",
                                    value = state.ntfyTopic.ifBlank { "Создаётся при сохранении" },
                                    monospace = true,
                                    copyable = state.ntfyTopic.isNotBlank(),
                                )
                                HorizontalDivider(
                                    modifier = Modifier.padding(vertical = 12.dp),
                                    color = MaterialTheme.colorScheme.outline.copy(alpha = 0.5f),
                                )
                                SettingSwitchRow(
                                    title = "Подписка в приложении",
                                    subtitle = "Фоновое получение через WebSocket",
                                    checked = state.ntfySubscribed,
                                    onCheckedChange = {
                                        if (it) requestPermissionIfNeeded()
                                        vm.setNtfySubscribed(it)
                                    },
                                )
                            }
                        }

                        item {
                            DashboardCard(title = "Хост недоступен") {
                                SettingSwitchRow(
                                    title = "Сообщать об offline",
                                    subtitle = "Если агент не отвечает дольше заданного времени",
                                    checked = state.offlineEnabled,
                                    onCheckedChange = vm::setOfflineEnabled,
                                )
                                if (state.offlineEnabled) {
                                    OutlinedTextField(
                                        value = state.offlineAfterSec,
                                        onValueChange = vm::setOfflineAfterSec,
                                        label = { Text("Секунд без связи") },
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .padding(top = 12.dp),
                                        singleLine = true,
                                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                                    )
                                }
                            }
                        }

                        item {
                            DashboardCard(
                                title = "Пауза между повторами",
                                subtitle = "Не слать один и тот же алерт чаще",
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(top = 8.dp),
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                ) {
                                    listOf(5, 15, 30).forEach { minutes ->
                                        FilterChip(
                                            selected = state.cooldownMinutes == minutes,
                                            onClick = { vm.setCooldownMinutes(minutes) },
                                            label = { Text("$minutes мин") },
                                            colors = FilterChipDefaults.filterChipColors(
                                                selectedContainerColor = SysMonColors.Accent.copy(alpha = 0.2f),
                                                selectedLabelColor = SysMonColors.Accent,
                                            ),
                                        )
                                    }
                                }
                            }
                        }

                        item {
                            DashboardCard(
                                title = "Режим порогов",
                                subtitle = "Свой порог или warn/critical из датчика",
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(top = 8.dp),
                                    horizontalArrangement = Arrangement.SpaceBetween,
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    Text(
                                        if (state.sensorThresholdMode) "Как в датчике" else "Свой порог для push",
                                        style = MaterialTheme.typography.bodyMedium,
                                    )
                                    Switch(
                                        checked = state.sensorThresholdMode,
                                        onCheckedChange = vm::setSensorThresholdMode,
                                    )
                                }
                            }
                        }

                        item {
                            DashboardCard(
                                title = "Датчики",
                                subtitle = if (state.sensorThresholdMode) {
                                    "Предупр. / критич. из настроек датчика"
                                } else {
                                    "Порог срабатывания (≥)"
                                },
                            ) {
                                if (state.sensorRows.isEmpty()) {
                                    Text(
                                        "Нет доступных датчиков",
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                                        modifier = Modifier.padding(top = 8.dp),
                                    )
                                } else {
                                    Column(modifier = Modifier.padding(top = 4.dp)) {
                                        state.sensorRows.forEachIndexed { index, row ->
                                            if (index > 0) {
                                                HorizontalDivider(
                                                    modifier = Modifier.padding(vertical = 10.dp),
                                                    color = MaterialTheme.colorScheme.outline.copy(alpha = 0.4f),
                                                )
                                            }
                                            SensorAlertRowView(
                                                row = row,
                                                sensorMode = state.sensorThresholdMode,
                                                onEnabledChange = { vm.setSensorEnabled(row.sensorId, it) },
                                                onThresholdChange = { vm.setSensorThreshold(row.sensorId, it) },
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }

                    state.error?.let { message ->
                        item {
                            Text(
                                message,
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }
                    }

                }
            }
        }

        Button(
            onClick = {
                requestPermissionIfNeeded()
                vm.save(onBack)
            },
            enabled = !state.isLoading && !state.isSaving,
            modifier = Modifier
                .fillMaxWidth()
                .navigationBarsPadding()
                .padding(horizontal = 16.dp, vertical = 12.dp),
        ) {
            Text(if (state.isSaving) "Сохранение…" else "Сохранить")
        }
    }
}

@Composable
private fun SettingSwitchRow(
    title: String,
    subtitle: String? = null,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    enabled: Boolean = true,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = if (subtitle != null) 0.dp else 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(
            modifier = Modifier
                .weight(1f)
                .padding(end = 12.dp),
        ) {
            Text(
                title,
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.Medium,
            )
            if (subtitle != null) {
                Text(
                    subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(top = 2.dp),
                )
            }
        }
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            enabled = enabled,
        )
    }
}

@Composable
private fun InfoValueRow(
    label: String,
    value: String,
    monospace: Boolean = false,
    copyable: Boolean = false,
) {
    val clipboard = LocalClipboardManager.current
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                label,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Text(
                value,
                style = MaterialTheme.typography.bodyMedium.copy(
                    fontFamily = if (monospace) FontFamily.Monospace else FontFamily.Default,
                ),
                modifier = Modifier.padding(top = 2.dp),
            )
        }
        if (copyable) {
            IconButton(
                onClick = { clipboard.setText(AnnotatedString(value)) },
            ) {
                Icon(
                    Icons.Default.ContentCopy,
                    contentDescription = "Копировать",
                    tint = SysMonColors.Accent,
                )
            }
        }
    }
}

@Composable
private fun SensorAlertRowView(
    row: SensorAlertRow,
    sensorMode: Boolean,
    onEnabledChange: (Boolean) -> Unit,
    onThresholdChange: (String) -> Unit,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                row.name,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
            )
            Text(
                row.sensorId,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
        if (sensorMode) {
            val limits = buildList {
                row.warnAbove?.let { add("≥ $it") }
                row.criticalAbove?.let { add("≥ $it") }
            }.joinToString(" / ").ifBlank { "—" }
            Text(
                text = row.unit?.let { "$limits $it" } ?: limits,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.widthIn(min = 72.dp, max = 120.dp),
            )
        } else {
            OutlinedTextField(
                value = row.thresholdText,
                onValueChange = onThresholdChange,
                modifier = Modifier.widthIn(min = 72.dp, max = 96.dp),
                singleLine = true,
                enabled = row.enabled,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                suffix = {
                    row.unit?.let {
                        Text(it, style = MaterialTheme.typography.labelSmall)
                    }
                },
                shape = RoundedCornerShape(10.dp),
            )
        }
        Switch(
            checked = row.enabled,
            onCheckedChange = onEnabledChange,
        )
    }
}
