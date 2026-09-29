package ru.ferrumnst.sysmon.data.api

import ru.ferrumnst.sysmon.data.models.AgentAlertConfig
import ru.ferrumnst.sysmon.data.models.AgentAlertsResponse
import ru.ferrumnst.sysmon.data.models.AgentsConfigResponse
import ru.ferrumnst.sysmon.data.models.AgentsConfigUpdateRequest
import ru.ferrumnst.sysmon.data.models.AgentsConfigUpdateResponse
import ru.ferrumnst.sysmon.data.models.AgentsResponse
import ru.ferrumnst.sysmon.data.models.AuthStatus
import ru.ferrumnst.sysmon.data.models.DashboardResponse
import ru.ferrumnst.sysmon.data.models.DashboardUpdateRequest
import ru.ferrumnst.sysmon.data.models.HubConfigResponse
import ru.ferrumnst.sysmon.data.models.HubConfigUpdateRequest
import ru.ferrumnst.sysmon.data.models.HubInfoResponse
import ru.ferrumnst.sysmon.data.models.LoginRequest
import ru.ferrumnst.sysmon.data.models.MetricsHistoryResponse
import ru.ferrumnst.sysmon.data.models.ModeResponse
import ru.ferrumnst.sysmon.data.models.SensorsResponse
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.models.VersionResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.HTTP
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path
import retrofit2.http.Query

interface HubApi {
    @GET("api/auth/status")
    suspend fun authStatus(): AuthStatus

    @POST("api/auth/login")
    suspend fun login(@Body body: LoginRequest): Response<Unit>

    @POST("api/auth/logout")
    suspend fun logout(): Response<Unit>

    @GET("api/mode")
    suspend fun mode(): ModeResponse

    @GET("api/agents")
    suspend fun agents(): AgentsResponse

    @GET("api/sensors")
    suspend fun sensors(@Query("agent") agentId: String? = null): SensorsResponse

    @GET("api/metrics/{sensorId}")
    suspend fun metrics(
        @Path(value = "sensorId", encoded = true) sensorId: String,
        @Query("agent") agentId: String? = null,
        @Query("period") period: String = "1h",
    ): MetricsHistoryResponse

    @GET("api/dashboard")
    suspend fun dashboard(): DashboardResponse

    @GET("api/system")
    suspend fun system(@Query("agent") agentId: String? = null): AgentSystemInfo

    @GET("api/version")
    suspend fun version(): VersionResponse

    @GET("api/hub/info")
    suspend fun hubInfo(): HubInfoResponse

    @GET("api/config/hub")
    suspend fun hubConfig(): HubConfigResponse

    @PUT("api/config/hub")
    suspend fun updateHubConfig(@Body body: HubConfigUpdateRequest): HubConfigResponse

    @GET("api/config/agents")
    suspend fun agentsConfig(): AgentsConfigResponse

    @PUT("api/config/agents")
    suspend fun updateAgentsConfig(@Body body: AgentsConfigUpdateRequest): AgentsConfigUpdateResponse

    @PUT("api/config/dashboard")
    suspend fun updateDashboard(@Body body: DashboardUpdateRequest): Response<Unit>

    @GET("api/alerts/{agentId}")
    suspend fun getAgentAlerts(@Path("agentId") agentId: String): AgentAlertsResponse

    @PUT("api/alerts/{agentId}")
    suspend fun updateAgentAlerts(
        @Path("agentId") agentId: String,
        @Body body: AgentAlertConfig,
    ): AgentAlertsResponse
}
