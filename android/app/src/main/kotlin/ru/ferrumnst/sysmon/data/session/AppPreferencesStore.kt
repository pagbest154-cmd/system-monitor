package ru.ferrumnst.sysmon.data.session

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.appPreferencesDataStore: DataStore<Preferences> by preferencesDataStore(
    name = "sysmon_app_prefs",
)

data class AppPreferences(
    val useDarkTheme: Boolean = true,
    val biometricLockEnabled: Boolean = false,
)

class AppPreferencesStore(private val context: Context) {
    private object Keys {
        val USE_DARK_THEME = booleanPreferencesKey("use_dark_theme")
        val BIOMETRIC_LOCK = booleanPreferencesKey("biometric_lock")
    }

    val preferences: Flow<AppPreferences> = context.appPreferencesDataStore.data.map { prefs ->
        AppPreferences(
            useDarkTheme = prefs[Keys.USE_DARK_THEME] ?: true,
            biometricLockEnabled = prefs[Keys.BIOMETRIC_LOCK] ?: false,
        )
    }

    suspend fun setUseDarkTheme(enabled: Boolean) {
        context.appPreferencesDataStore.edit { it[Keys.USE_DARK_THEME] = enabled }
    }

    suspend fun setBiometricLockEnabled(enabled: Boolean) {
        context.appPreferencesDataStore.edit { it[Keys.BIOMETRIC_LOCK] = enabled }
    }
}
