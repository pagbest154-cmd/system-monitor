package ru.ferrumnst.sysmon.ui.screens.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import android.app.Application
import androidx.lifecycle.AndroidViewModel
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.notifications.NtfySubscriptionManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class LoginUiState(
    val hubName: String = "",
    val hubKey: String = "",
    val hubNameReadOnly: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null,
)

class LoginViewModel(
    application: Application,
    private val repository: HubRepository,
) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    init {
        loadAuthStatus()
    }

    private fun loadAuthStatus() {
        viewModelScope.launch {
            val session = repository.getSessionOnce()
            runCatching {
                repository.checkAuth(session)
            }.onSuccess { status ->
                _uiState.update {
                    it.copy(
                        hubName = status.hubName ?: session.hubName,
                        hubNameReadOnly = !status.hubName.isNullOrBlank(),
                    )
                }
            }
        }
    }

    fun onNameChange(name: String) {
        _uiState.update { it.copy(hubName = name, error = null) }
    }

    fun onKeyChange(key: String) {
        _uiState.update { it.copy(hubKey = key, error = null) }
    }

    fun login(onSuccess: () -> Unit) {
        val state = _uiState.value
        if (state.hubName.isBlank() || state.hubKey.isBlank()) {
            _uiState.update { it.copy(error = "Заполните имя и ключ") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            val session = repository.getSessionOnce()
            repository.login(session, state.hubName, state.hubKey)
                .onSuccess {
                    _uiState.update { it.copy(isLoading = false) }
                    NtfySubscriptionManager.syncAfterLogin(getApplication(), repository)
                    onSuccess()
                }
                .onFailure {
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            error = "Неверное имя хаба или ключ",
                        )
                    }
                }
        }
    }

    class Factory(
        private val application: Application,
        private val repository: HubRepository,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return LoginViewModel(application, repository) as T
        }
    }
}
