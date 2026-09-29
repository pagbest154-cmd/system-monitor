package ru.ferrumnst.sysmon.ui.screens.settings

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.Button
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.ui.components.DashboardCard

@Composable
fun HubSettingsTab(
    state: SettingsUiState,
    vm: SettingsViewModel,
) {
    val context = LocalContext.current

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        DashboardCard(title = "Сервер hub") {
            Text(
                "Версия hub: ${state.hubVersion}",
                style = MaterialTheme.typography.bodyMedium,
            )
            if (state.hubUpdateAvailable) {
                Text(
                    "Доступно обновление: ${state.hubLatestVersion ?: "—"}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.tertiary,
                    modifier = Modifier.padding(top = 8.dp),
                )
                state.hubUpdateHint?.let { hint ->
                    Text(
                        hint,
                        style = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace),
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(top = 4.dp),
                    )
                }
                OutlinedButton(
                    onClick = {
                        val text = state.hubUpdateHint ?: state.hubReleaseUrl ?: return@OutlinedButton
                        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        clipboard.setPrimaryClip(ClipData.newPlainText("hub update", text))
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                ) {
                    Text("Скопировать команду обновления")
                }
            }
        }

        DashboardCard(title = "Общие") {
            Text("Хранение истории: ${state.retentionDays} дн.", style = MaterialTheme.typography.bodyMedium)
            Text(
                "Интервал по умолчанию: ${state.defaultIntervalSec} с",
                style = MaterialTheme.typography.bodyMedium,
            )
            if (state.mode == "hub") {
                Text(
                    "В hub-режиме датчики настраиваются на агентах",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(top = 4.dp),
                )
            }
        }

        DashboardCard(title = "Панель мониторинга") {
            OutlinedTextField(
                value = state.dashboardTitle,
                onValueChange = vm::onDashboardTitleChange,
                label = { Text("Заголовок") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
            )
            OutlinedTextField(
                value = state.dashboardRefreshSec.toString(),
                onValueChange = vm::onDashboardRefreshChange,
                label = { Text("Интервал обновления (с)") },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                singleLine = true,
            )
            state.panels.forEachIndexed { index, panel ->
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 12.dp),
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("Панель ${index + 1}", style = MaterialTheme.typography.labelMedium)
                        IconButton(onClick = { vm.removePanel(index) }) {
                            Icon(Icons.Default.Delete, contentDescription = "Удалить панель")
                        }
                    }
                    OutlinedTextField(
                        value = panel.title,
                        onValueChange = { vm.updatePanel(index, title = it) },
                        label = { Text("Заголовок") },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                    )
                    OutlinedTextField(
                        value = panel.type,
                        onValueChange = { vm.updatePanel(index, type = it) },
                        label = { Text("Тип") },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = 6.dp),
                        singleLine = true,
                    )
                    OutlinedTextField(
                        value = panel.sensors.joinToString(", "),
                        onValueChange = { raw ->
                            val sensors = raw.split(',').map { it.trim() }.filter { it.isNotEmpty() }
                            vm.updatePanel(index, sensors = sensors)
                        },
                        label = { Text("Датчики (через запятую)") },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = 6.dp),
                        singleLine = false,
                    )
                }
            }
            OutlinedButton(
                onClick = vm::addPanel,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
            ) {
                Icon(Icons.Default.Add, contentDescription = null)
                Text("Добавить панель", modifier = Modifier.padding(start = 8.dp))
            }
            Button(
                onClick = vm::saveDashboard,
                enabled = !state.isSaving,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
            ) {
                Text(if (state.isSaving) "Сохранение..." else "Сохранить панель")
            }
        }

        if (state.mode == "hub") {
            DashboardCard(
                title = "Домен hub",
                subtitle = "URL для агентов: ${state.hubPublicUrlResolved.ifBlank { "—" }}",
            ) {
                OutlinedTextField(
                    value = state.hubDomain,
                    onValueChange = vm::onHubDomainChange,
                    label = { Text("Домен") },
                    placeholder = { Text("monitor.example.com") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                )
                OutlinedTextField(
                    value = state.hubPublicUrl,
                    onValueChange = vm::onHubPublicUrlChange,
                    label = { Text("Публичный URL") },
                    placeholder = { Text("https://monitor.example.com") },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                    singleLine = true,
                )
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("HTTPS")
                    Switch(
                        checked = state.hubUseHttps,
                        onCheckedChange = vm::onHubUseHttpsChange,
                    )
                }
                Button(
                    onClick = vm::saveHubConfig,
                    enabled = !state.isSaving,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                ) {
                    Text(if (state.isSaving) "Сохранение..." else "Сохранить домен")
                }
            }

            DashboardCard(
                title = "Агенты",
                subtitle = "Токены генерируются на hub при сохранении",
            ) {
                state.agents.forEachIndexed { index, agent ->
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 8.dp),
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Text("Агент ${index + 1}", style = MaterialTheme.typography.labelMedium)
                            IconButton(onClick = { vm.removeAgent(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = "Удалить")
                            }
                        }
                        OutlinedTextField(
                            value = agent.id,
                            onValueChange = { vm.updateAgent(index, id = it) },
                            label = { Text("ID") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                        )
                        OutlinedTextField(
                            value = agent.name,
                            onValueChange = { vm.updateAgent(index, name = it) },
                            label = { Text("Имя") },
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(top = 6.dp),
                            singleLine = true,
                        )
                        OutlinedTextField(
                            value = agent.token,
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("Token") },
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(top = 6.dp),
                            singleLine = true,
                            visualTransformation = PasswordVisualTransformation(),
                            textStyle = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace),
                        )
                    }
                }
                OutlinedButton(
                    onClick = vm::addAgent,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Icon(Icons.Default.Add, contentDescription = null)
                    Text("Добавить агента", modifier = Modifier.padding(start = 8.dp))
                }
                Button(
                    onClick = vm::saveAgents,
                    enabled = !state.isSaving && state.agents.isNotEmpty(),
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                ) {
                    Text(if (state.isSaving) "Сохранение..." else "Сохранить агентов")
                }
            }
        }
    }
}
