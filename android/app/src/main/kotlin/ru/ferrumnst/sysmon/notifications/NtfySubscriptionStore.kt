package ru.ferrumnst.sysmon.notifications

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

private val Context.ntfyStore: DataStore<Preferences> by preferencesDataStore(name = "sysmon_ntfy")

class NtfySubscriptionStore(private val context: Context) {
    private val json = Json { ignoreUnknownKeys = true }
    private val subscriptionsKey = stringPreferencesKey("subscriptions_json")

    suspend fun getAll(): List<NtfySubscription> {
        val raw = context.ntfyStore.data.map { it[subscriptionsKey] ?: "[]" }.first()
        return runCatching {
            json.decodeFromString<List<NtfySubscription>>(raw)
        }.getOrDefault(emptyList())
    }

    suspend fun upsert(subscription: NtfySubscription) {
        context.ntfyStore.edit { prefs ->
            val current = runCatching {
                json.decodeFromString<List<NtfySubscription>>(prefs[subscriptionsKey] ?: "[]")
            }.getOrDefault(emptyList())
            val updated = current
                .filterNot { it.agentId == subscription.agentId }
                .plus(subscription)
            prefs[subscriptionsKey] = json.encodeToString(updated)
        }
    }

    suspend fun remove(agentId: String) {
        context.ntfyStore.edit { prefs ->
            val current = runCatching {
                json.decodeFromString<List<NtfySubscription>>(prefs[subscriptionsKey] ?: "[]")
            }.getOrDefault(emptyList())
            prefs[subscriptionsKey] = json.encodeToString(current.filterNot { it.agentId == agentId })
        }
    }

    suspend fun clear() {
        context.ntfyStore.edit { it.remove(subscriptionsKey) }
    }
}
