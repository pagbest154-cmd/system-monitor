package ru.ferrumnst.sysmon.ui.screens.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.DashboardPanel
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.SensorInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.util.HubErrors
import ru.ferrumnst.sysmon.ui.util.chartPointsFromHistory
import ru.ferrumnst.sysmon.ui.util.isEffectivelyLive
import ru.ferrumnst.sysmon.ui.util.resolvePanelSensorIds
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.drop
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class PanelHistory(
    val panelId: String,
    val histories: Map<String, List<MetricReading>> = emptyMap(),
)

data class DashboardUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val error: String? = null,
    val mode: String = "hub",
    val selectedAgentId: String = "",
    val selectedAgentName: String = "",
    val agents: List<Pair<String, String>> = emptyList(),
    val sensors: List<SensorInfo> = emptyList(),
    val panels: List<DashboardPanel> = emptyList(),
    val panelHistories: Map<String, PanelHistory> = emptyMap(),
    val period: String = "1h",
    val liveConnected: Boolean = false,
    val selectedAgentOnline: Boolean = true,
    val isLive: Boolean = false,
    val lastUpdateTs: Double = 0.0,
    val liveReadings: Map<String, MetricReading> = emptyMap(),
    val liveBuffers: Map<String, List<MetricReading>> = emptyMap(),
    val systemInfo: AgentSystemInfo? = null,
)

class DashboardViewModel(
    private val repository: HubRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(DashboardUiState())
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    private var historyJob: Job? = null
    private val livePointBuffer = mutableMapOf<String, MutableList<MetricReading>>()

    init {
        viewModelScope.launch {
            combine(
                repository.liveConnected,
                repository.lastUpdateTs,
                repository.liveReadings,
            ) { connected, ts, readings ->
                Triple(connected, ts, readings)
            }.collect { (connected, ts, readings) ->
                val hadReadings = _uiState.value.liveReadings.isNotEmpty()
                readings.forEach { (sensorId, reading) ->
                    if (reading.value == null || reading.value.isNaN()) return@forEach
                    val buffer = livePointBuffer.getOrPut(sensorId) { mutableListOf() }
                    val last = buffer.lastOrNull()
                    if (last?.ts != reading.ts || last?.value != reading.value) {
                        buffer.add(reading)
                        if (buffer.size > 240) {
                            buffer.removeAt(0)
                        }
                    }
                }
                val current = _uiState.value
                val isLive = isEffectivelyLive(connected, ts, current.selectedAgentOnline)
                _uiState.update {
                    it.copy(
                        liveConnected = connected,
                        lastUpdateTs = ts,
                        isLive = isLive,
                        liveReadings = readings,
                        liveBuffers = livePointBuffer.mapValues { entry -> entry.value.toList() },
                    )
                }
                if (!hadReadings && readings.isNotEmpty() && _uiState.value.panelHistories.isEmpty()) {
                    loadPanelHistories()
                }
            }
        }

        viewModelScope.launch {
            repository.session
                .map { it.selectedAgentId }
                .distinctUntilChanged()
                .drop(1)
                .collect { selectedId ->
                    if (selectedId != _uiState.value.selectedAgentId) {
                        refresh(showFullLoading = false)
                    }
                }
        }

        viewModelScope.launch {
            while (isActive) {
                delay(1_000)
                val state = _uiState.value
                val isLive = isEffectivelyLive(
                    state.liveConnected,
                    state.lastUpdateTs,
                    state.selectedAgentOnline,
                )
                if (isLive != state.isLive) {
                    _uiState.update { it.copy(isLive = isLive) }
                }
            }
        }

        viewModelScope.launch {
            while (isActive) {
                delay(30_000)
                refreshSelectedAgentOnline()
            }
        }

        refresh()
    }

    private suspend fun refreshSelectedAgentOnline() {
        val state = _uiState.value
        if (state.mode != "hub" || state.selectedAgentId.isBlank()) return
        runCatching {
            val session = repository.getSessionOnce()
            val agents = repository.getAgents(session)
            val online = agents.find { it.id == state.selectedAgentId }?.status == "online"
            val isLive = isEffectivelyLive(state.liveConnected, state.lastUpdateTs, online)
            _uiState.update { it.copy(selectedAgentOnline = online, isLive = isLive) }
        }
    }

    fun refresh(showFullLoading: Boolean = true) {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isLoading = showFullLoading && it.agents.isEmpty(),
                    isRefreshing = !showFullLoading || it.agents.isNotEmpty(),
                    error = null,
                )
            }
            val session = repository.getSessionOnce()
            runCatching {
                val mode = repository.getMode(session)
                val agents = repository.getAgents(session)
                val agentOptions = agents.map { it.id to (it.name ?: it.id) }
                val selectedId = when {
                    session.selectedAgentId.isNotBlank() &&
                        agents.any { it.id == session.selectedAgentId } ->
                        session.selectedAgentId
                    agents.size == 1 -> agents.first().id
                    else -> session.selectedAgentId
                }
                val selectedName = agents.find { it.id == selectedId }?.name ?: selectedId
                val selectedAgentOnline = when {
                    mode != "hub" -> true
                    selectedId.isBlank() -> false
                    else -> agents.find { it.id == selectedId }?.status == "online"
                }
                if (selectedId != _uiState.value.selectedAgentId) {
                    livePointBuffer.clear()
                }
                if (selectedId.isNotBlank() && selectedId != session.selectedAgentId) {
                    repository.saveSelectedAgent(selectedId)
                }

                val sensors = if (mode == "hub" && selectedId.isBlank()) {
                    emptyList()
                } else {
                    repository.getSensors(session, selectedId.ifBlank { null })
                }
                val panels = repository.getDashboardPanels(session)
                val systemInfo = if (mode == "hub" && selectedId.isNotBlank()) {
                    repository.getSystemInfo(session, selectedId)
                } else if (mode != "hub") {
                    repository.getSystemInfo(session, null)
                } else {
                    null
                }

                val liveState = _uiState.value
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        isRefreshing = false,
                        mode = mode,
                        agents = agentOptions,
                        selectedAgentId = selectedId,
                        selectedAgentName = selectedName,
                        selectedAgentOnline = selectedAgentOnline,
                        isLive = isEffectivelyLive(
                            liveState.liveConnected,
                            liveState.lastUpdateTs,
                            selectedAgentOnline,
                        ),
                        sensors = sensors,
                        panels = panels.sortedBy { it.row ?: 0 },
                        systemInfo = systemInfo,
                    )
                }

                if (mode == "hub" && selectedId.isNotBlank()) {
                    repository.connectLive(session, selectedId)
                } else if (mode != "hub") {
                    repository.connectLive(session, null)
                } else {
                    repository.disconnectLive()
                }

                loadPanelHistories()
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        isRefreshing = false,
                        error = HubErrors.userMessage(error, "Ошибка загрузки"),
                    )
                }
            }
        }
    }

    fun selectAgent(agentId: String) {
        viewModelScope.launch {
            repository.saveSelectedAgent(agentId)
            refresh(showFullLoading = false)
        }
    }

    fun setPeriod(period: String) {
        if (period == _uiState.value.period) return
        _uiState.update { it.copy(period = period) }
        loadPanelHistories(period)
    }

    private fun loadPanelHistories(periodOverride: String? = null) {
        historyJob?.cancel()
        historyJob = viewModelScope.launch {
            val state = _uiState.value
            if (state.mode == "hub" && state.selectedAgentId.isBlank()) return@launch

            val period = periodOverride ?: state.period
            val session = repository.getSessionOnce()
            val agentId = state.selectedAgentId.ifBlank { null }
            val histories = mutableMapOf<String, PanelHistory>()

            state.panels.filter { it.type.equals("line", ignoreCase = true) }.forEach { panel ->
                val sensorHistories = mutableMapOf<String, List<MetricReading>>()
                val sensorIds = resolvePanelSensorIds(panel, state.sensors, state.liveReadings)
                sensorIds.forEach { sensorId ->
                    if (sensorId == "auto_disks") return@forEach
                    val live = state.liveReadings[sensorId]
                        ?: state.sensors.find { it.id == sensorId }?.current
                    val apiPoints = runCatching {
                        repository.getMetricsHistory(
                            session = session,
                            sensorId = sensorId,
                            agentId = agentId,
                            period = period,
                        ).points
                    }.getOrElse { emptyList() }
                    sensorHistories[sensorId] = chartPointsFromHistory(
                        history = apiPoints,
                        live = live,
                        buffered = livePointBuffer[sensorId] ?: emptyList(),
                    )
                }
                histories[panel.id] = PanelHistory(panel.id, sensorHistories)
            }

            _uiState.update { it.copy(panelHistories = histories, period = period) }
        }
    }

    override fun onCleared() {
        repository.disconnectLive()
        super.onCleared()
    }

    class Factory(
        private val repository: HubRepository,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return DashboardViewModel(repository) as T
        }
    }
}
