package ru.ferrumnst.sysmon.data.websocket

import ru.ferrumnst.sysmon.data.api.HubClientFactory
import ru.ferrumnst.sysmon.data.models.LiveMessage
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.session.HubSession
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import okhttp3.Credentials
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

class LiveWebSocketClient(
    private val scope: CoroutineScope,
) {
    private var webSocket: WebSocket? = null
    private var reconnectJob: Job? = null
    private var session: HubSession? = null
    private var agentId: String? = null

    private val _readings = MutableStateFlow<Map<String, MetricReading>>(emptyMap())
    val readings: StateFlow<Map<String, MetricReading>> = _readings.asStateFlow()

    private val _connected = MutableStateFlow(false)
    val connected: StateFlow<Boolean> = _connected.asStateFlow()

    private val _lastUpdateTs = MutableStateFlow(0.0)
    val lastUpdateTs: StateFlow<Double> = _lastUpdateTs.asStateFlow()

    fun connect(session: HubSession, agentId: String?) {
        disconnect()
        this.session = session
        this.agentId = agentId
        openSocket()
    }

    fun disconnect() {
        reconnectJob?.cancel()
        reconnectJob = null
        webSocket?.close(1000, "disconnect")
        webSocket = null
        _connected.value = false
        _readings.value = emptyMap()
        _lastUpdateTs.value = 0.0
    }

    private fun openSocket() {
        val currentSession = session ?: return
        _connected.value = false
        val client = HubClientFactory.createOkHttpClient(currentSession)
        val url = HubClientFactory.webSocketUrl(currentSession, agentId)

        val requestBuilder = Request.Builder().url(url)
        if (currentSession.hasCredentials) {
            requestBuilder.header(
                "Authorization",
                Credentials.basic(currentSession.hubName, currentSession.hubKey),
            )
        }

        webSocket = client.newWebSocket(
            requestBuilder.build(),
            object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    _connected.value = true
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    handleMessage(text)
                }

                override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                    webSocket.close(code, reason)
                }

                override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                    _connected.value = false
                    scheduleReconnect(code)
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    _connected.value = false
                    scheduleReconnect(null)
                }
            },
        )
    }

    private fun handleMessage(text: String) {
        val message = runCatching {
            HubClientFactory.json.decodeFromString(LiveMessage.serializer(), text)
        }.getOrNull() ?: return

        if (message.type != "snapshot" && message.type != "update") return

        val agentKey = agentId ?: message.agentId ?: ""
        val incoming = normalizeSnapshot(message.data, agentKey)
        val merged = _readings.value.toMutableMap()
        incoming.forEach { (id, reading) ->
            if (reading.value != null) {
                merged[id] = reading
            }
        }
        _readings.value = merged

        val maxTs = merged.values.maxOfOrNull { it.ts ?: 0.0 } ?: 0.0
        if (maxTs > 0) {
            _lastUpdateTs.value = maxTs
        }
    }

    private fun normalizeSnapshot(
        data: Map<String, MetricReading>,
        agentId: String,
    ): Map<String, MetricReading> {
        if (agentId.isBlank()) return data
        val prefix = "$agentId:"
        return data.mapKeys { (key, _) ->
            if (key.startsWith(prefix)) key.removePrefix(prefix) else key
        }
    }

    private fun scheduleReconnect(closeCode: Int?) {
        if (closeCode == 1008) return
        if (reconnectJob?.isActive == true) return
        reconnectJob = scope.launch {
            delay(3_000)
            if (isActive && session != null) {
                openSocket()
            }
        }
    }
}
