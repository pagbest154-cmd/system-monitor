package ru.ferrumnst.sysmon.ui.screens.hosts

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import ru.ferrumnst.sysmon.data.models.AgentConfigEntry
import ru.ferrumnst.sysmon.data.models.NetworkInterfaceInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.util.getNetworkInterfaceOverride
import ru.ferrumnst.sysmon.ui.util.updateAgentNetworkInterface

data class HostNetworkUiState(
    val agentId: String = "",
    val agentName: String = "",
    val isLoading: Boolean = true,
    val isSaving: Boolean = false,
    val error: String? = null,
    val notInConfig: Boolean = false,
    val interfaces: List<NetworkInterfaceInfo> = emptyList(),
    val selectedInterface: String = "",
    val customInterface: String = "",
    val saved: Boolean = false,
)

class HostNetworkViewModel(
    application: Application,
    private val repository: HubRepository,
    private val agentId: String,
    private val agentName: String,
) : AndroidViewModel(application) {
    private var configAgents: List<AgentConfigEntry> = emptyList()

    private val _uiState = MutableStateFlow(
        HostNetworkUiState(agentId = agentId, agentName = agentName),
    )
    val uiState: StateFlow<HostNetworkUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null, saved = false) }
            runCatching {
                val session = repository.getSessionOnce()
                configAgents = repository.getAgentsConfig(session)
                val entry = configAgents.find { it.id == agentId }
                val system = repository.getSystemInfo(session, agentId)
                val ifaces = system.network.filter { !it.name.isNullOrBlank() }
                val current = entry?.let { getNetworkInterfaceOverride(it.overrides) } ?: ""
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        notInConfig = entry == null,
                        interfaces = ifaces,
                        selectedInterface = if (ifaces.any { n -> n.name == current }) current else "",
                        customInterface = current,
                    )
                }
            }.onFailure { err ->
                _uiState.update {
                    it.copy(isLoading = false, error = err.message ?: "Не удалось загрузить")
                }
            }
        }
    }

    fun setCustomInterface(value: String) {
        _uiState.update {
            val match = it.interfaces.any { iface -> iface.name == value }
            it.copy(
                customInterface = value,
                selectedInterface = if (match) value else "",
            )
        }
    }

    fun setSelectedInterface(value: String) {
        _uiState.update {
            it.copy(
                selectedInterface = value,
                customInterface = value,
            )
        }
    }

    fun save() {
        viewModelScope.launch {
            val state = _uiState.value
            if (state.notInConfig) return@launch
            _uiState.update { it.copy(isSaving = true, error = null) }
            runCatching {
                val session = repository.getSessionOnce()
                val iface = state.customInterface.trim().ifBlank { state.selectedInterface.trim() }
                val updated = updateAgentNetworkInterface(configAgents, agentId, iface)
                configAgents = repository.updateAgentsConfig(session, updated)
                _uiState.update { it.copy(isSaving = false, saved = true) }
            }.onFailure { err ->
                _uiState.update {
                    it.copy(isSaving = false, error = err.message ?: "Не удалось сохранить")
                }
            }
        }
    }

    class Factory(
        private val application: Application,
        private val repository: HubRepository,
        private val agentId: String,
        private val agentName: String,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return HostNetworkViewModel(application, repository, agentId, agentName) as T
        }
    }
}
