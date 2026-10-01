package ru.ferrumnst.sysmon.notifications

import android.app.Service
import android.content.Intent
import android.os.IBinder
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.TimeUnit
import kotlin.random.Random

class NtfyListenerService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val json = Json { ignoreUnknownKeys = true }
    private val sockets = ConcurrentHashMap<String, WebSocket>()
    private val client = OkHttpClient.Builder()
        .pingInterval(30, TimeUnit.SECONDS)
        .build()

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                closeAll()
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
                return START_NOT_STICKY
            }
            else -> {
                // Must run before any async work — startForegroundService() deadline is ~10s on Android 8+.
                startForeground(
                    FOREGROUND_NOTIFICATION_ID,
                    NotificationHelper.buildListenerNotification(this),
                )
                syncSubscriptions()
            }
        }
        return START_STICKY
    }

    override fun onDestroy() {
        closeAll()
        scope.cancel()
        super.onDestroy()
    }

    private fun syncSubscriptions() {
        scope.launch {
            val subscriptions = NtfySubscriptionStore(this@NtfyListenerService).getAll()
            if (subscriptions.isEmpty()) {
                closeAll()
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
                return@launch
            }
            val activeKeys = subscriptions.map { it.key }.toSet()
            sockets.keys.filterNot { it in activeKeys }.forEach { key ->
                sockets.remove(key)?.close(1000, "removed")
            }
            subscriptions.forEach { subscription ->
                if (sockets.containsKey(subscription.key)) return@forEach
                connect(subscription)
            }
        }
    }

    private fun connect(subscription: NtfySubscription) {
        val requestBuilder = Request.Builder().url(subscription.wsUrl())
        if (subscription.token.isNotBlank()) {
            requestBuilder.header("Authorization", "Bearer ${subscription.token}")
        }
        val socket = client.newWebSocket(
            requestBuilder.build(),
            object : WebSocketListener() {
                override fun onMessage(webSocket: WebSocket, text: String) {
                    handleMessage(text, subscription.agentId)
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    sockets.remove(subscription.key)
                    scheduleReconnect(subscription)
                }

                override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                    sockets.remove(subscription.key)
                }
            },
        )
        sockets[subscription.key] = socket
    }

    private fun scheduleReconnect(subscription: NtfySubscription) {
        scope.launch {
            kotlinx.coroutines.delay(5_000)
            if (!sockets.containsKey(subscription.key)) {
                connect(subscription)
            }
        }
    }

    private fun handleMessage(raw: String, fallbackAgentId: String) {
        val message = runCatching { json.decodeFromString<NtfyMessage>(raw) }.getOrNull() ?: return
        if (message.event != "message") return
        val title = message.title?.takeIf { it.isNotBlank() } ?: message.topic ?: "SysMon"
        val body = message.message?.takeIf { it.isNotBlank() } ?: return
        val agentId = message.headers?.agentId() ?: fallbackAgentId.takeIf { it.isNotBlank() }
        val notificationId = (agentId?.hashCode() ?: title.hashCode()) + Random.nextInt(1000)
        NotificationHelper.showAlert(
            context = this,
            notificationId = notificationId,
            title = title,
            body = body,
            agentId = agentId,
        )
    }

    private fun closeAll() {
        sockets.values.forEach { it.close(1000, "stop") }
        sockets.clear()
    }

    companion object {
        const val ACTION_SYNC = "ru.ferrumnst.sysmon.ntfy.SYNC"
        const val ACTION_STOP = "ru.ferrumnst.sysmon.ntfy.STOP"
        private const val FOREGROUND_NOTIFICATION_ID = 1001
    }
}

private fun Map<String, String>.agentId(): String? {
    return entries.firstOrNull { (key, value) ->
        value.isNotBlank() && (
            key.equals("agent_id", ignoreCase = true) ||
                key.equals("agent-id", ignoreCase = true)
            )
    }?.value
}
