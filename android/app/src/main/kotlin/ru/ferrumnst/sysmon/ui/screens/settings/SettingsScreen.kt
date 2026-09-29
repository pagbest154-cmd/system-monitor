package ru.ferrumnst.sysmon.ui.screens.settings

import android.app.Activity
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import android.app.Application
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.SysMonApplication
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.components.DashboardCard
import ru.ferrumnst.sysmon.ui.components.SysMonBottomBarClearance
import kotlinx.coroutines.launch

@Composable
fun SettingsScreen(
    repository: HubRepository,
    onLogout: () -> Unit,
    onResetHub: () -> Unit,
) {
    val app = LocalContext.current.applicationContext as Application
    val sysMonApp = app as SysMonApplication
    val vm: SettingsViewModel = viewModel(
        factory = SettingsViewModel.Factory(app, repository, sysMonApp.appUpdateManager),
    )
    val state by vm.uiState.collectAsState()
    val session by repository.session.collectAsState(initial = ru.ferrumnst.sysmon.data.session.HubSession())
    val scope = rememberCoroutineScope()
    val activity = LocalContext.current as? Activity

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .statusBarsPadding()
            .padding(start = 16.dp, end = 16.dp, top = 12.dp, bottom = 16.dp + SysMonBottomBarClearance),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(
            "Настройки",
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.onBackground,
        )

        if (state.isLoading) {
            CircularProgressIndicator(modifier = Modifier.padding(24.dp))
            return@Column
        }

        state.message?.let {
            Text(it, color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.bodyMedium)
        }
        state.error?.let {
            Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
        }

        DashboardCard(title = "Подключение приложения") {
            Text("URL: ${session.hubUrl.ifBlank { "—" }}", style = MaterialTheme.typography.bodyMedium)
            if (session.hasCredentials) {
                Text("Пользователь: ${session.hubName}", style = MaterialTheme.typography.bodyMedium)
            }
            Text(
                "Версия hub: ${state.hubVersion}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 4.dp),
            )
            if (session.hasCredentials) {
                OutlinedButton(
                    onClick = {
                        scope.launch {
                            repository.logout(session)
                            onLogout()
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                ) {
                    Text("Выйти")
                }
            }
            Button(
                onClick = {
                    scope.launch {
                        repository.clearAll()
                        onResetHub()
                    }
                },
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Сменить хаб")
            }
        }

        DashboardCard(title = "Обновление приложения") {
            Text(
                "Версия приложения: ${state.appVersion}",
                style = MaterialTheme.typography.bodyMedium,
            )
            when {
                state.isCheckingAppUpdate -> {
                    CircularProgressIndicator(modifier = Modifier.padding(top = 8.dp))
                }
                state.appUpdateAvailable && state.latestAppVersion != null -> {
                    Text(
                        "Доступна версия ${state.latestAppVersion}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.padding(top = 4.dp),
                    )
                    if (state.isDownloadingAppUpdate) {
                        val progress = state.appUpdateProgress
                        if (progress != null) {
                            Text(
                                "Загрузка: ${(progress * 100).toInt()}%",
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(top = 8.dp),
                            )
                        } else {
                            CircularProgressIndicator(modifier = Modifier.padding(top = 8.dp))
                        }
                    } else {
                        Button(
                            onClick = {
                                val hostActivity = activity ?: return@Button
                                vm.downloadAndInstall(hostActivity)
                            },
                            enabled = activity != null,
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(top = 8.dp),
                        ) {
                            Text("Обновить до ${state.latestAppVersion}")
                        }
                    }
                }
                else -> {
                    Text(
                        "Установлена последняя версия",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(top = 4.dp),
                    )
                }
            }
            state.appUpdateMessage?.let { message ->
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    message,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
            OutlinedButton(
                onClick = vm::checkAppUpdate,
                enabled = !state.isCheckingAppUpdate && !state.isDownloadingAppUpdate,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
            ) {
                Text("Проверить обновления")
            }
        }

        DashboardCard(title = "Виджеты") {
            Text(
                "Добавьте виджет «Показатели хоста» на рабочий стол Android: долгое нажатие → Виджеты → SysMon.",
                style = MaterialTheme.typography.bodyMedium,
            )
            Text(
                "При добавлении выберите хост и до 6 датчиков (CPU, память, GPU, сеть, диски и др.). Обновление каждые 15 минут.",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 6.dp),
            )
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
            Text(
                "Панелей на dashboard: ${state.panelCount}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp),
            )
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
