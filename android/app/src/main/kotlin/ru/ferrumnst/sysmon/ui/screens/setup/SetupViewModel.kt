package ru.ferrumnst.sysmon.ui.screens.setup

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import ru.ferrumnst.sysmon.data.repository.HubRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class SetupUiState(
    val hubUrl: String = "",
    val isLoading: Boolean = false,
    val error: String? = null,
    val pairingName: String? = null,
    val pairingKey: String? = null,
)

class SetupViewModel(
    private val repository: HubRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(SetupUiState())
    val uiState: StateFlow<SetupUiState> = _uiState.asStateFlow()

    fun onUrlChange(url: String) {
        _uiState.update {
            it.copy(hubUrl = url, error = null, pairingName = null, pairingKey = null)
        }
    }

    fun applyScannedPayload(raw: String, onAutoConnect: (() -> Unit)? = null) {
        val scan = parseHubPairingScan(raw)
        _uiState.update {
            it.copy(
                hubUrl = scan.hubUrl,
                error = null,
                pairingName = scan.hubName,
                pairingKey = scan.hubKey,
            )
        }
        if (scan.includesCredentials && onAutoConnect != null) {
            save(onAutoConnect)
        }
    }

    fun save(onSuccess: () -> Unit) {
        val state = _uiState.value
        val url = state.hubUrl.trim()
        if (url.isBlank()) {
            _uiState.update { it.copy(error = "Укажите URL хаба") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                repository.saveHubUrl(url)
                val session = repository.getSessionOnce()
                val status = repository.checkAuth(session)
                val name = state.pairingName?.trim().orEmpty()
                val key = state.pairingKey.orEmpty()
                if (status.authRequired && name.isNotBlank() && key.isNotBlank()) {
                    repository.login(session, name, key).getOrThrow()
                }
            }.onSuccess {
                _uiState.update {
                    it.copy(isLoading = false, pairingName = null, pairingKey = null)
                }
                onSuccess()
            }.onFailure { error ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = error.message ?: "Не удалось подключиться к хабу",
                    )
                }
            }
        }
    }

    class Factory(
        private val repository: HubRepository,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return SetupViewModel(repository) as T
        }
    }
}
