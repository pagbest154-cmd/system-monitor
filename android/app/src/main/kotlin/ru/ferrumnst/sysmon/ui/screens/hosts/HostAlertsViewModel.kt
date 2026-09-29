package ru.ferrumnst.sysmon.ui.screens.hosts

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import ru.ferrumnst.sysmon.data.models.AgentAlertConfig
import ru.ferrumnst.sysmon.data.models.AlertOfflineRule
import ru.ferrumnst.sysmon.data.models.AlertSensorRule
import ru.ferrumnst.sysmon.data.models.NtfyAlertConfig
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.notifications.NtfySubscription
import ru.ferrumnst.sysmon.notifications.NtfySubscriptionManager
import ru.ferrumnst.sysmon.notifications.NtfySubscriptionStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class SensorAlertRow(
    val sensorId: String,
    val name: String,
    val unit: String?,
    val enabled: Boolean,
    val thresholdText: String,
)

data class HostAlertsUiState(
    val agentId: String = "",
    val agentName: String = "",
    val isLoading: Boolean = true,
    val isSaving: Boolean = false,
    val error: String? = null,
    val alertsEnabled: Boolean = false,
    val offlineEnabled: Boolean = false,
    val offlineAfterSec: String = "180",
    val cooldownMinutes: Int = 15,
    val sensorRows: List<SensorAlertRow> = emptyList(),
    val ntfyBaseUrl: String = "",
    val ntfyTopic: String = "",
    val ntfyToken: String = "",
    val ntfySubscribed: Boolean = false,
    val saved: Boolean = false,
)

class HostAlertsViewModel(
    application: Application,
    private val repository: HubRepository,
    private val agentId: String,
    private val agentName: String,
) : AndroidViewModel(application) {
    private val subscriptionStore = NtfySubscriptionStore(application)
    private val _uiState = MutableStateFlow(
        HostAlertsUiState(agentId = agentId, agentName = agentName),
    )
    val uiState: StateFlow<HostAlertsUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, saved = false) }
            runCatching {
                val session = repository.getSessionOnce()
                val resolvedName = repository.getAgents(session)
                    .find { it.id == agentId }
                    ?.name
                    ?.takeIf { it.isNotBlank() }
                    ?: agentName
                val sensors = repository.getSensors(session, agentId).filter { it.enabled }
                val alertsResponse = repository.getAgentAlerts(session, agentId)
                val alerts = alertsResponse.config
                val ruleMap = alerts.sensors.associateBy { it.sensorId }
                val rows = sensors.map { sensor ->
                    val rule = ruleMap[sensor.id]
                    val defaultThreshold = sensor.warnAbove ?: 80.0
                    SensorAlertRow(
                        sensorId = sensor.id,
                        name = sensor.name,
                        unit = sensor.unit,
                        enabled = rule?.enabled == true,
                        thresholdText = formatThreshold(rule?.threshold ?: defaultThreshold),
                    )
                }
                val subscribed = subscriptionStore.getAll().any { it.agentId == agentId }
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        agentName = resolvedName,
                        alertsEnabled = alerts.enabled,
                        offlineEnabled = alerts.offline.enabled,
                        offlineAfterSec = alerts.offline.afterSec.toString(),
                        cooldownMinutes = (alerts.cooldownSec / 60).coerceAtLeast(1),
                        sensorRows = rows,
                        ntfyBaseUrl = alertsResponse.ntfyBaseUrl.ifBlank { "https://ntfy.sh" },
                        ntfyTopic = alerts.ntfy.topic,
                        ntfyToken = alerts.ntfy.token,
                        ntfySubscribed = subscribed,
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isLoading = false, error = error.message ?: "Не удалось загрузить настройки")
                }
            }
        }
    }

    fun setAlertsEnabled(enabled: Boolean) {
        _uiState.update { it.copy(alertsEnabled = enabled) }
    }

    fun setOfflineEnabled(enabled: Boolean) {
        _uiState.update { it.copy(offlineEnabled = enabled) }
    }

    fun setOfflineAfterSec(value: String) {
        _uiState.update { it.copy(offlineAfterSec = value) }
    }

    fun setCooldownMinutes(minutes: Int) {
        _uiState.update { it.copy(cooldownMinutes = minutes) }
    }

    fun setNtfySubscribed(subscribed: Boolean) {
        _uiState.update { it.copy(ntfySubscribed = subscribed) }
    }

    fun setSensorEnabled(sensorId: String, enabled: Boolean) {
        _uiState.update { state ->
            state.copy(
                sensorRows = state.sensorRows.map { row ->
                    if (row.sensorId == sensorId) row.copy(enabled = enabled) else row
                },
            )
        }
    }

    fun setSensorThreshold(sensorId: String, value: String) {
        _uiState.update { state ->
            state.copy(
                sensorRows = state.sensorRows.map { row ->
                    if (row.sensorId == sensorId) row.copy(thresholdText = value) else row
                },
            )
        }
    }

    fun save(onSuccess: () -> Unit) {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, error = null) }
            runCatching {
                val session = repository.getSessionOnce()
                val sensors = state.sensorRows.map { row ->
                    AlertSensorRule(
                        sensorId = row.sensorId,
                        enabled = row.enabled,
                        threshold = row.thresholdText.replace(',', '.').toDoubleOrNull() ?: 0.0,
                    )
                }
                val config = AgentAlertConfig(
                    enabled = state.alertsEnabled,
                    offline = AlertOfflineRule(
                        enabled = state.offlineEnabled,
                        afterSec = state.offlineAfterSec.toIntOrNull() ?: 180,
                    ),
                    sensors = sensors,
                    cooldownSec = state.cooldownMinutes * 60,
                    ntfy = NtfyAlertConfig(
                        topic = state.ntfyTopic,
                        token = state.ntfyToken,
                    ),
                )
                val saved = repository.updateAgentAlerts(session, agentId, config)
                val topic = saved.config.ntfy.topic
                val baseUrl = saved.ntfyBaseUrl.ifBlank { state.ntfyBaseUrl }
                _uiState.update {
                    it.copy(
                        isSaving = false,
                        saved = true,
                        ntfyTopic = topic,
                        ntfyBaseUrl = baseUrl,
                    )
                }
                if (state.ntfySubscribed && state.alertsEnabled && topic.isNotBlank()) {
                    val subscription = NtfySubscription(
                        agentId = agentId,
                        baseUrl = baseUrl,
                        topic = topic,
                        token = saved.config.ntfy.token,
                    )
                    NtfySubscriptionManager.subscribe(getApplication(), subscription)
                } else {
                    NtfySubscriptionManager.unsubscribe(getApplication(), agentId)
                }
                onSuccess()
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isSaving = false,
                        error = error.message ?: "Не удалось сохранить",
                    )
                }
            }
        }
    }

    private fun formatThreshold(value: Double): String {
        return if (value % 1.0 == 0.0) value.toInt().toString() else value.toString()
    }

    class Factory(
        private val application: Application,
        private val repository: HubRepository,
        private val agentId: String,
        private val agentName: String,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return HostAlertsViewModel(application, repository, agentId, agentName) as T
        }
    }
}
