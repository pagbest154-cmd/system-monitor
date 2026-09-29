package ru.ferrumnst.sysmon.data.session

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import ru.ferrumnst.sysmon.data.models.AgentInfo

private val Context.offlineCacheDataStore: DataStore<Preferences> by preferencesDataStore(
    name = "sysmon_offline_cache",
)

class OfflineCacheStore(private val context: Context) {
    private val json = Json { ignoreUnknownKeys = true }

    private object Keys {
        val AGENTS_JSON = stringPreferencesKey("agents_json")
        val CACHED_AT = longPreferencesKey("cached_at")
    }

    suspend fun saveAgents(agents: List<AgentInfo>) {
        context.offlineCacheDataStore.edit { prefs ->
            prefs[Keys.AGENTS_JSON] = json.encodeToString(agents)
            prefs[Keys.CACHED_AT] = System.currentTimeMillis()
        }
    }

    suspend fun getCachedAgents(): List<AgentInfo>? {
        val prefs = context.offlineCacheDataStore.data.first()
        val raw = prefs[Keys.AGENTS_JSON] ?: return null
        return runCatching { json.decodeFromString<List<AgentInfo>>(raw) }.getOrNull()
    }

    suspend fun cachedAtMillis(): Long? {
        return context.offlineCacheDataStore.data.map { it[Keys.CACHED_AT] }.first()
    }

    suspend fun clear() {
        context.offlineCacheDataStore.edit { it.clear() }
    }
}
