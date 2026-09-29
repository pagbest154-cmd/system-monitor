package ru.ferrumnst.sysmon.data.session

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "sysmon_session")

data class HubSession(
    val hubUrl: String = "",
    val hubName: String = "",
    val hubKey: String = "",
    val selectedAgentId: String = "",
) {
    val isConfigured: Boolean
        get() = hubUrl.isNotBlank()

    val hasCredentials: Boolean
        get() = hubName.isNotBlank() && hubKey.isNotBlank()
}

class SessionStore(private val context: Context) {
    private object Keys {
        val HUB_URL = stringPreferencesKey("hub_url")
        val HUB_NAME = stringPreferencesKey("hub_name")
        val HUB_KEY = stringPreferencesKey("hub_key")
        val SELECTED_AGENT = stringPreferencesKey("selected_agent")
    }

    val session: Flow<HubSession> = context.dataStore.data.map { prefs ->
        HubSession(
            hubUrl = prefs[Keys.HUB_URL] ?: "",
            hubName = prefs[Keys.HUB_NAME] ?: "",
            hubKey = prefs[Keys.HUB_KEY] ?: "",
            selectedAgentId = prefs[Keys.SELECTED_AGENT] ?: "",
        )
    }

    suspend fun saveHubUrl(url: String) {
        context.dataStore.edit { it[Keys.HUB_URL] = url.trim().removeSuffix("/") }
    }

    suspend fun saveCredentials(name: String, key: String) {
        context.dataStore.edit {
            it[Keys.HUB_NAME] = name.trim()
            it[Keys.HUB_KEY] = key
        }
    }

    suspend fun saveSelectedAgent(agentId: String) {
        context.dataStore.edit { it[Keys.SELECTED_AGENT] = agentId }
    }

    suspend fun clearCredentials() {
        context.dataStore.edit {
            it.remove(Keys.HUB_NAME)
            it.remove(Keys.HUB_KEY)
        }
    }

    suspend fun clearAll() {
        context.dataStore.edit { it.clear() }
    }
}
