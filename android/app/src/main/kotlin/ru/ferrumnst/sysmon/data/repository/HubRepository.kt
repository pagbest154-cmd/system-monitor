package ru.ferrumnst.sysmon.data.repository

import ru.ferrumnst.sysmon.data.api.HubApi
import ru.ferrumnst.sysmon.data.api.HubClientFactory
import ru.ferrumnst.sysmon.data.api.MetricsFetcher
import ru.ferrumnst.sysmon.data.models.AgentAlertConfig
import ru.ferrumnst.sysmon.data.models.AgentAlertsResponse
import ru.ferrumnst.sysmon.data.models.AgentConfigEntry
import ru.ferrumnst.sysmon.data.models.AgentInfo
import ru.ferrumnst.sysmon.data.models.AgentsResponse
import ru.ferrumnst.sysmon.data.models.AgentsConfigUpdateRequest
import ru.ferrumnst.sysmon.data.models.AuthStatus
import ru.ferrumnst.sysmon.data.models.DashboardConfig
import ru.ferrumnst.sysmon.data.models.DashboardPanel
import ru.ferrumnst.sysmon.data.models.DashboardResponse
import ru.ferrumnst.sysmon.data.models.DashboardUpdateRequest
import ru.ferrumnst.sysmon.data.models.HubConfig
import ru.ferrumnst.sysmon.data.models.HubConfigUpdateRequest
import ru.ferrumnst.sysmon.data.models.HubInfoResponse
import ru.ferrumnst.sysmon.data.models.LoginRequest
import ru.ferrumnst.sysmon.data.models.MetricReading
import ru.ferrumnst.sysmon.data.models.MetricsHistoryResponse
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.SensorInfo
import ru.ferrumnst.sysmon.data.models.SensorsResponse
import ru.ferrumnst.sysmon.data.models.VersionResponse
import ru.ferrumnst.sysmon.data.session.HubSession
import ru.ferrumnst.sysmon.data.session.OfflineCacheStore
import ru.ferrumnst.sysmon.data.session.SessionStore
import ru.ferrumnst.sysmon.data.websocket.LiveWebSocketClient
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.first
import retrofit2.HttpException

class HubRepository(
    private val sessionStore: SessionStore,
    private val offlineCache: OfflineCacheStore,
    private val liveClient: LiveWebSocketClient,
) {
    val session: Flow<HubSession> = sessionStore.session

    val liveReadings: StateFlow<Map<String, MetricReading>> = liveClient.readings
    val liveConnected: StateFlow<Boolean> = liveClient.connected
    val lastUpdateTs: StateFlow<Double> = liveClient.lastUpdateTs

    private var api: HubApi? = null
    private var cachedSession: HubSession? = null

    suspend fun getSessionOnce(): HubSession = sessionStore.session.first()

    private fun apiFor(session: HubSession): HubApi {
        if (api == null || cachedSession != session) {
            api = HubClientFactory.createApi(session)
            cachedSession = session
        }
        return api!!
    }

    suspend fun saveHubUrl(url: String) {
        invalidateClient()
        sessionStore.saveHubUrl(url)
    }

    suspend fun saveCredentials(name: String, key: String) {
        invalidateClient()
        sessionStore.saveCredentials(name, key)
    }

    suspend fun saveSelectedAgent(agentId: String) {
        sessionStore.saveSelectedAgent(agentId)
    }

    suspend fun clearAll() {
        liveClient.disconnect()
        invalidateClient()
        offlineCache.clear()
        sessionStore.clearAll()
    }

    suspend fun clearCredentials() {
        liveClient.disconnect()
        invalidateClient()
        sessionStore.clearCredentials()
    }

    private fun invalidateClient() {
        api = null
        cachedSession = null
    }

    suspend fun checkAuth(session: HubSession): AuthStatus {
        return apiFor(session).authStatus()
    }

    suspend fun login(session: HubSession, name: String, key: String): Result<Unit> {
        return runCatching {
            val response = apiFor(session).login(LoginRequest(name, key))
            if (!response.isSuccessful) {
                throw HttpException(response)
            }
            sessionStore.saveCredentials(name, key)
            invalidateClient()
        }
    }

    suspend fun logout(session: HubSession) {
        runCatching { apiFor(session).logout() }
        liveClient.disconnect()
        sessionStore.clearCredentials()
        invalidateClient()
    }

    suspend fun getMode(session: HubSession): String {
        return apiFor(session).mode().mode
    }

    suspend fun getAgentsResponse(session: HubSession): AgentsResponse {
        val response = apiFor(session).agents()
        offlineCache.saveAgents(response.agents)
        return response
    }

    suspend fun getAgents(session: HubSession): List<AgentInfo> {
        return getAgentsResponse(session).agents
    }

    suspend fun getCachedAgents(): List<AgentInfo>? = offlineCache.getCachedAgents()

    suspend fun getCachedAgentsAtMillis(): Long? = offlineCache.cachedAtMillis()

    suspend fun deleteAgent(session: HubSession, agentId: String) {
        val response = apiFor(session).deleteAgent(agentId)
        if (!response.isSuccessful) {
            throw HttpException(response)
        }
        val cached = offlineCache.getCachedAgents()
        if (cached != null) {
            offlineCache.saveAgents(cached.filter { it.id != agentId })
        }
    }

    suspend fun getSensors(session: HubSession, agentId: String?): List<SensorInfo> {
        return apiFor(session).sensors(agentId).sensors
    }

    suspend fun getSystemInfo(session: HubSession, agentId: String?): AgentSystemInfo {
        return runCatching {
            apiFor(session).system(agentId)
        }.getOrElse { AgentSystemInfo() }
    }

    suspend fun getDashboard(session: HubSession): DashboardResponse {
        return apiFor(session).dashboard()
    }

    suspend fun getDashboardPanels(session: HubSession): List<DashboardPanel> {
        return getDashboard(session).panels
    }

    suspend fun getVersion(session: HubSession): VersionResponse {
        return apiFor(session).version()
    }

    suspend fun getHubInfo(session: HubSession): HubInfoResponse {
        return apiFor(session).hubInfo()
    }

    suspend fun getHubConfig(session: HubSession): HubConfig {
        return apiFor(session).hubConfig().hub
    }

    suspend fun updateHubConfig(session: HubSession, config: HubConfig): HubConfig {
        return apiFor(session).updateHubConfig(HubConfigUpdateRequest(config)).hub
    }

    suspend fun getAgentsConfig(session: HubSession): List<AgentConfigEntry> {
        return apiFor(session).agentsConfig().agents
    }

    suspend fun updateAgentsConfig(session: HubSession, agents: List<AgentConfigEntry>): List<AgentConfigEntry> {
        return apiFor(session).updateAgentsConfig(AgentsConfigUpdateRequest(agents)).agents
    }

    suspend fun updateDashboard(
        session: HubSession,
        dashboard: DashboardConfig,
        panels: List<DashboardPanel>,
    ) {
        val response = apiFor(session).updateDashboard(DashboardUpdateRequest(dashboard, panels))
        if (!response.isSuccessful) {
            throw HttpException(response)
        }
    }

    suspend fun getSensorsResponse(session: HubSession, agentId: String? = null): SensorsResponse {
        return apiFor(session).sensors(agentId)
    }

    suspend fun getMetricsHistory(
        session: HubSession,
        sensorId: String,
        agentId: String?,
        period: String,
    ): MetricsHistoryResponse {
        return MetricsFetcher.fetchHistory(session, sensorId, agentId, period)
    }

    fun connectLive(session: HubSession, agentId: String?) {
        liveClient.connect(session, agentId)
    }

    fun disconnectLive() {
        liveClient.disconnect()
    }

    suspend fun getAgentAlerts(session: HubSession, agentId: String): AgentAlertsResponse {
        return apiFor(session).getAgentAlerts(agentId)
    }

    suspend fun updateAgentAlerts(
        session: HubSession,
        agentId: String,
        config: AgentAlertConfig,
    ): AgentAlertsResponse {
        return apiFor(session).updateAgentAlerts(agentId, config)
    }
}
