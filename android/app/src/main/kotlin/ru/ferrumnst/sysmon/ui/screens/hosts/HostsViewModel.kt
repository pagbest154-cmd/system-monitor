package ru.ferrumnst.sysmon.ui.screens.hosts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import ru.ferrumnst.sysmon.data.models.AgentInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.util.HubErrors
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.drop
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class HostsUiState(
    val isLoading: Boolean = true,
    val error: String? = null,
    val agents: List<AgentInfo> = emptyList(),
    val selectedAgentId: String = "",
)

class HostsViewModel(
    private val repository: HubRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(HostsUiState())
    val uiState: StateFlow<HostsUiState> = _uiState.asStateFlow()

    init {
        refresh()
        viewModelScope.launch {
            repository.session
                .map { it.selectedAgentId }
                .distinctUntilChanged()
                .drop(1)
                .collect { selectedId ->
                    _uiState.update { it.copy(selectedAgentId = selectedId) }
                }
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            val session = repository.getSessionOnce()
            runCatching {
                val agents = repository.getAgents(session)
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        agents = agents,
                        selectedAgentId = session.selectedAgentId,
                    )
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = HubErrors.userMessage(error, "Ошибка загрузки хостов"),
                    )
                }
            }
        }
    }

    fun selectAgent(agentId: String) {
        viewModelScope.launch {
            repository.saveSelectedAgent(agentId)
            _uiState.update { it.copy(selectedAgentId = agentId) }
        }
    }

    class Factory(
        private val repository: HubRepository,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return HostsViewModel(repository) as T
        }
    }
}
