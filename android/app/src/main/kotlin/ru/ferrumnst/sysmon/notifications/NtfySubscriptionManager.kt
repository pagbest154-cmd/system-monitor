package ru.ferrumnst.sysmon.notifications

import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.content.ContextCompat
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import ru.ferrumnst.sysmon.data.repository.HubRepository

object NtfySubscriptionManager {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    fun syncAfterLogin(context: Context, repository: HubRepository) {
        scope.launch {
            runCatching {
                val session = repository.getSessionOnce()
                if (!session.hasCredentials) return@launch
                val store = NtfySubscriptionStore(context.applicationContext)
                val existing = store.getAll()
                if (existing.isEmpty()) return@launch
                val updated = existing.mapNotNull { sub ->
                    val alerts = repository.getAgentAlerts(session, sub.agentId)
                    val topic = alerts.config.ntfy.topic.trim()
                    if (!alerts.config.enabled || topic.isBlank()) {
                        null
                    } else {
                        NtfySubscription(
                            agentId = sub.agentId,
                            baseUrl = alerts.ntfyBaseUrl.ifBlank { sub.baseUrl },
                            topic = topic,
                            token = alerts.config.ntfy.token,
                        )
                    }
                }
                store.clear()
                updated.forEach { store.upsert(it) }
                refreshService(context.applicationContext)
            }
        }
    }

    fun subscribe(context: Context, subscription: NtfySubscription) {
        scope.launch {
            NtfySubscriptionStore(context.applicationContext).upsert(subscription)
            refreshService(context.applicationContext)
        }
    }

    fun unsubscribe(context: Context, agentId: String) {
        scope.launch {
            NtfySubscriptionStore(context.applicationContext).remove(agentId)
            refreshService(context.applicationContext)
        }
    }

    fun clearAll(context: Context) {
        scope.launch {
            NtfySubscriptionStore(context.applicationContext).clear()
            context.applicationContext.stopService(
                Intent(context.applicationContext, NtfyListenerService::class.java),
            )
        }
    }

    fun refreshService(context: Context) {
        val appContext = context.applicationContext
        val intent = Intent(appContext, NtfyListenerService::class.java).apply {
            action = NtfyListenerService.ACTION_SYNC
        }
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            ContextCompat.startForegroundService(appContext, intent)
        } else {
            appContext.startService(intent)
        }
    }
}
