package ru.ferrumnst.sysmon.ui.screens.settings

import android.app.Activity
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.launch
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.AppPreferences
import ru.ferrumnst.sysmon.data.session.AppPreferencesStore
import ru.ferrumnst.sysmon.data.session.HubSession
import ru.ferrumnst.sysmon.ui.components.DashboardCard

@Composable
fun AppSettingsTab(
    state: SettingsUiState,
    session: HubSession,
    appPrefs: AppPreferences,
    appPreferencesStore: AppPreferencesStore,
    repository: HubRepository,
    vm: SettingsViewModel,
    scope: CoroutineScope,
    activity: Activity?,
    onLogout: () -> Unit,
    onResetHub: () -> Unit,
) {
    val context = LocalContext.current

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        DashboardCard(title = "Интерфейс") {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("Тёмная тема")
                Switch(
                    checked = appPrefs.useDarkTheme,
                    onCheckedChange = { enabled ->
                        scope.launch { appPreferencesStore.setUseDarkTheme(enabled) }
                    },
                )
            }
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("Биометрическая блокировка")
                Switch(
                    checked = appPrefs.biometricLockEnabled,
                    onCheckedChange = { enabled ->
                        scope.launch { appPreferencesStore.setBiometricLockEnabled(enabled) }
                    },
                )
            }
            Text(
                "При включении потребуется отпечаток или PIN перед открытием приложения",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 4.dp),
            )
        }

        DashboardCard(title = "Подключение к хабу") {
            Text("URL: ${session.hubUrl.ifBlank { "—" }}", style = MaterialTheme.typography.bodyMedium)
            if (session.hasCredentials) {
                Text("Пользователь: ${session.hubName}", style = MaterialTheme.typography.bodyMedium)
            }
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
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
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
            Text(
                "Если установка пишет «Приложение не установлено» — удалите SysMon и установите APK из релиза заново (разная подпись старой сборки).",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp),
            )
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
    }
}
