package ru.ferrumnst.sysmon.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import retrofit2.HttpException
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.HubSession
import ru.ferrumnst.sysmon.notifications.NtfySubscriptionManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

sealed class AppStartDestination {
    data object Loading : AppStartDestination()
    data object Setup : AppStartDestination()
    data object Login : AppStartDestination()
    data object HubUnavailable : AppStartDestination()
    data object Main : AppStartDestination()
}

class AppViewModel(
    application: Application,
    private val repository: HubRepository,
) : AndroidViewModel(application) {
    val session: StateFlow<HubSession> = repository.session.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5_000),
        initialValue = HubSession(),
    )

    private val _startDestination = MutableStateFlow<AppStartDestination>(AppStartDestination.Loading)
    val startDestination: StateFlow<AppStartDestination> = _startDestination.asStateFlow()

    init {
        resolveStartDestination()
    }

    fun resolveStartDestination() {
        viewModelScope.launch {
            _startDestination.value = AppStartDestination.Loading
            val current = repository.getSessionOnce()
            if (!current.isConfigured) {
                _startDestination.value = AppStartDestination.Setup
                return@launch
            }

            runCatching {
                val status = repository.checkAuth(current)
                val destination = when {
                    !status.authRequired || status.authenticated || current.hasCredentials ->
                        AppStartDestination.Main
                    else -> AppStartDestination.Login
                }
                _startDestination.value = destination
                if (destination == AppStartDestination.Main && current.hasCredentials) {
                    NtfySubscriptionManager.syncAfterLogin(getApplication(), repository)
                }
            }.onFailure { error ->
                _startDestination.value = when {
                    error is HttpException && error.code() == 401 -> AppStartDestination.Login
                    current.isConfigured -> AppStartDestination.HubUnavailable
                    else -> AppStartDestination.Setup
                }
            }
        }
    }

    fun onChangeHub() {
        viewModelScope.launch {
            NtfySubscriptionManager.clearAll(getApplication())
            repository.clearAll()
            _startDestination.value = AppStartDestination.Setup
        }
    }

    fun onSetupComplete() {
        resolveStartDestination()
    }

    fun onLoginComplete() {
        _startDestination.value = AppStartDestination.Main
        NtfySubscriptionManager.syncAfterLogin(getApplication(), repository)
    }

    fun onLogout() {
        viewModelScope.launch {
            NtfySubscriptionManager.clearAll(getApplication())
            val current = repository.getSessionOnce()
            repository.logout(current)
            _startDestination.value = AppStartDestination.Login
        }
    }

    class Factory(
        private val application: Application,
        private val repository: HubRepository,
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            return AppViewModel(application, repository) as T
        }
    }
}
