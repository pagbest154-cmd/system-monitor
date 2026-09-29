package ru.ferrumnst.sysmon.ui.screens.settings

import android.app.Activity
import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import ru.ferrumnst.sysmon.data.models.AgentConfigEntry
import ru.ferrumnst.sysmon.data.models.DashboardConfig
import ru.ferrumnst.sysmon.data.models.DashboardPanel
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.update.AppUpdateManager
import ru.ferrumnst.sysmon.ui.util.HubErrors
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class SettingsUiState(
    val isLoading: Boolean = true,
    val isSaving: Boolean = false,
    val error: String? = null,
    val message: String? = null,
    val appVersion: String = "",
    val appUpdateAvailable: Boolean = false,
    val latestAppVersion: String? = null,
    val isCheckingAppUpdate: Boolean = false,
    val isDownloadingAppUpdate: Boolean = false,
    val appUpdateProgress: Float? = null,
    val appUpdateMessage: String? = null,
    val mode: String = "hub",
    val hubVersion: String = "",
    val retentionDays: Int = 31,
    val defaultIntervalSec: Int = 5,
    val dashboardTitle: String = "",
    val dashboardRefreshSec: Int = 15,
    val panelCount: Int = 0,
    val hubDomain: String = "",
    val hubPublicUrl: String = "",
    val hubUseHttps: Boolean = true,
    val hubPublicUrlResolved: String = "",
    val agents: List<AgentConfigEntry> = emptyList(),
    val panels: List<DashboardPanel> = emptyList(),
)

class SettingsViewModel(
    application: Application,
    private val repository: HubRepository,
    private val appUpdateManager: AppUpdateManager,
) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow(
        SettingsUiState(appVersion = appUpdateManager.currentAppVersion()),
    )
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    init {
        load()
        viewModelScope.launch {
            appUpdateManager.updateInfo.collect { info ->
                if (info == null) return@collect
                _uiState.update { state ->
                    state.copy(
                        appVersion = info.currentVersion,
                        appUpdateAvailable = info.updateAvailable,
                        latestAppVersion = info.latestVersion,
                        appUpdateMessage = info.error,
                    )
                }
            }
        }
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            val session = repository.getSessionOnce()
            runCatching {
                val mode = repository.getMode(session)
                val versionInfo = repository.getVersion(session)
                val version = versionInfo.currentVersion ?: "—"
                val sensors = repository.getSensorsResponse(session)
                val dashboard = repository.getDashboard(session)

                var hubDomain = ""
                var hubPublicUrl = ""
                var hubUseHttps = true
                var hubResolved = ""
                var agents = emptyList<AgentConfigEntry>()

                if (mode == "hub") {
                    val hubConfig = repository.getHubConfig(session)
                    hubDomain = hubConfig.domain
                    hubPublicUrl = hubConfig.publicUrl
                    hubUseHttps = hubConfig.useHttps
                    hubResolved = hubConfig.publicUrlResolved ?: repository.getHubInfo(session).publicUrl
                    agents = repository.getAgentsConfig(session)
                }

                _uiState.update {
                    it.copy(
                        isLoading = false,
                        mode = mode,
                        hubVersion = version,
                        retentionDays = sensors.settings?.retentionDays ?: 31,
                        defaultIntervalSec = sensors.settings?.defaultIntervalSec ?: 5,
                        dashboardTitle = dashboard.dashboard?.title ?: "Мониторинг системы",
                        dashboardRefreshSec = dashboard.dashboard?.refreshSec ?: 15,
                        panelCount = dashboard.panels.size,
                        panels = dashboard.panels,
                        hubDomain = hubDomain,
                        hubPublicUrl = hubPublicUrl,
                        hubUseHttps = hubUseHttps,
                        hubPublicUrlResolved = hubResolved,
                        agents = agents,
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = HubErrors.userMessage(error, "Ошибка загрузки настроек"),
                    )
                }
            }
        }
    }

    fun onDashboardTitleChange(value: String) {
        _uiState.update { it.copy(dashboardTitle = value, message = null) }
    }

    fun onDashboardRefreshChange(value: String) {
        value.toIntOrNull()?.let { sec ->
            _uiState.update { it.copy(dashboardRefreshSec = sec, message = null) }
        }
    }

    fun onHubDomainChange(value: String) {
        _uiState.update { it.copy(hubDomain = value, message = null) }
    }

    fun onHubPublicUrlChange(value: String) {
        _uiState.update { it.copy(hubPublicUrl = value, message = null) }
    }

    fun onHubUseHttpsChange(value: Boolean) {
        _uiState.update { it.copy(hubUseHttps = value, message = null) }
    }

    fun updateAgent(index: Int, id: String? = null, name: String? = null) {
        _uiState.update { state ->
            val updated = state.agents.toMutableList()
            if (index !in updated.indices) return@update state
            val current = updated[index]
            updated[index] = current.copy(
                id = id ?: current.id,
                name = name ?: current.name,
            )
            state.copy(agents = updated, message = null)
        }
    }

    fun addAgent() {
        _uiState.update { state ->
            state.copy(
                agents = state.agents + AgentConfigEntry(
                    id = "agent_${System.currentTimeMillis()}",
                    name = "Новый агент",
                    token = "",
                ),
                message = null,
            )
        }
    }

    fun removeAgent(index: Int) {
        _uiState.update { state ->
            if (index !in state.agents.indices) return@update state
            state.copy(
                agents = state.agents.filterIndexed { i, _ -> i != index },
                message = null,
            )
        }
    }

    fun saveDashboard() {
        viewModelScope.launch {
            val state = _uiState.value
            _uiState.update { it.copy(isSaving = true, error = null, message = null) }
            val session = repository.getSessionOnce()
            runCatching {
                repository.updateDashboard(
                    session = session,
                    dashboard = DashboardConfig(
                        title = state.dashboardTitle.trim(),
                        refreshSec = state.dashboardRefreshSec,
                    ),
                    panels = state.panels,
                )
                _uiState.update { it.copy(isSaving = false, message = "Сохранено") }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isSaving = false, error = HubErrors.userMessage(error, "Ошибка сохранения"))
                }
            }
        }
    }

    fun saveHubConfig() {
        viewModelScope.launch {
            val state = _uiState.value
            _uiState.update { it.copy(isSaving = true, error = null, message = null) }
            val session = repository.getSessionOnce()
            runCatching {
                val saved = repository.updateHubConfig(
                    session = session,
                    config = ru.ferrumnst.sysmon.data.models.HubConfig(
                        domain = state.hubDomain.trim(),
                        publicUrl = state.hubPublicUrl.trim(),
                        useHttps = state.hubUseHttps,
                        trustedHosts = listOf("*"),
                    ),
                )
                _uiState.update {
                    it.copy(
                        isSaving = false,
                        message = "Сохранено",
                        hubPublicUrl = saved.publicUrl,
                        hubPublicUrlResolved = saved.publicUrlResolved ?: it.hubPublicUrlResolved,
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isSaving = false, error = HubErrors.userMessage(error, "Ошибка сохранения"))
                }
            }
        }
    }

    fun saveAgents() {
        viewModelScope.launch {
            val state = _uiState.value
            _uiState.update { it.copy(isSaving = true, error = null, message = null) }
            val session = repository.getSessionOnce()
            runCatching {
                val saved = repository.updateAgentsConfig(session, state.agents)
                _uiState.update {
                    it.copy(isSaving = false, message = "Сохранено", agents = saved)
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isSaving = false, error = HubErrors.userMessage(error, "Ошибка сохранения"))
                }
            }
        }
    }

    fun clearMessage() {
        _uiState.update { it.copy(message = null, error = null) }
    }

    fun checkAppUpdate() {
        viewModelScope.launch {
            _uiState.update { it.copy(isCheckingAppUpdate = true, appUpdateMessage = null) }
            val info = appUpdateManager.checkForUpdate(force = true)
            _uiState.update {
                it.copy(
                    isCheckingAppUpdate = false,
                    appVersion = info.currentVersion,
                    appUpdateAvailable = info.updateAvailable,
                    latestAppVersion = info.latestVersion,
                    appUpdateMessage = info.error,
                )
            }
        }
    }

    fun downloadAndInstall(activity: Activity) {
        val info = appUpdateManager.updateInfo.value
        val downloadUrl = info?.downloadUrl
        val fileName = info?.apkFileName
        if (downloadUrl.isNullOrBlank() || fileName.isNullOrBlank()) return

        if (!appUpdateManager.canInstallPackages()) {
            appUpdateManager.openInstallPermissionSettings(activity)
            _uiState.update {
                it.copy(appUpdateMessage = "Разрешите установку из этого источника и повторите")
            }
            return
        }

        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isDownloadingAppUpdate = true,
                    appUpdateProgress = 0f,
                    appUpdateMessage = null,
                )
            }
            runCatching {
                val apkFile = appUpdateManager.downloadApk(downloadUrl, fileName) { progress ->
                    _uiState.update { state -> state.copy(appUpdateProgress = progress) }
                }
                appUpdateManager.installApk(apkFile)
            }.onFailure { error ->
                _uiState.update {
                    it.copy(appUpdateMessage = error.message ?: "Ошибка загрузки обновления")
                }
            }
            _uiState.update {
                it.copy(isDownloadingAppUpdate = false, appUpdateProgress = null)
            }
        }
    }

    class Factory(
        private val application: Application,
        private val repository: HubRepository,
        private val appUpdateManager: AppUpdateManager,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return SettingsViewModel(application, repository, appUpdateManager) as T
        }
    }
}
