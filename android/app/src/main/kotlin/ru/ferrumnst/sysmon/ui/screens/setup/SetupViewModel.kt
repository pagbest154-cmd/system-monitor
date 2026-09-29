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
)

class SetupViewModel(
    private val repository: HubRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(SetupUiState())
    val uiState: StateFlow<SetupUiState> = _uiState.asStateFlow()

    fun onUrlChange(url: String) {
        _uiState.update { it.copy(hubUrl = url, error = null) }
    }

    fun applyScannedUrl(raw: String) {
        val url = raw.trim().removeSuffix("/")
        _uiState.update { it.copy(hubUrl = url, error = null) }
    }

    fun save(onSuccess: () -> Unit) {
        val url = _uiState.value.hubUrl.trim()
        if (url.isBlank()) {
            _uiState.update { it.copy(error = "Укажите URL хаба") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                repository.saveHubUrl(url)
                val session = repository.getSessionOnce()
                repository.checkAuth(session)
            }.onSuccess {
                _uiState.update { it.copy(isLoading = false) }
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
