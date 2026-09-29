package ru.ferrumnst.sysmon.data.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
data class AuthStatus(
    @SerialName("auth_required") val authRequired: Boolean = false,
    @SerialName("hub_name") val hubName: String? = null,
    val authenticated: Boolean = false,
)

@Serializable
data class LoginRequest(
    val name: String,
    val key: String,
)

@Serializable
data class ErrorResponse(
    val detail: String? = null,
)

@Serializable
data class ModeResponse(
    val mode: String,
)

@Serializable
data class AgentsResponse(
    val agents: List<AgentInfo> = emptyList(),
)

@Serializable
data class AgentInfo(
    val id: String,
    val name: String? = null,
    val hostname: String? = null,
    val status: String? = null,
    @SerialName("last_seen") val lastSeen: Double? = null,
    @SerialName("agent_version") val agentVersion: String? = null,
    val platform: String? = null,
    @SerialName("update_available") val updateAvailable: Boolean = false,
    val system: JsonElement? = null,
    val sensors: List<SensorMeta> = emptyList(),
)

@Serializable
data class SensorsResponse(
    val settings: SensorSettings? = null,
    @SerialName("agent_id") val agentId: String? = null,
    val sensors: List<SensorInfo> = emptyList(),
)

@Serializable
data class SensorSettings(
    @SerialName("retention_days") val retentionDays: Int? = null,
    @SerialName("default_interval_sec") val defaultIntervalSec: Int? = null,
)

@Serializable
data class SensorInfo(
    val id: String,
    @SerialName("full_id") val fullId: String? = null,
    val name: String,
    val type: String? = null,
    val unit: String? = null,
    val enabled: Boolean = true,
    @SerialName("warn_above") val warnAbove: Double? = null,
    val current: MetricReading? = null,
)

@Serializable
data class SensorMeta(
    val id: String,
    val name: String,
    val type: String? = null,
    val unit: String? = null,
    val enabled: Boolean = true,
)

@Serializable
data class MetricReading(
    @SerialName("sensor_id") val sensorId: String? = null,
    val value: Double? = null,
    val status: String? = null,
    val ts: Double? = null,
)

@Serializable
data class MetricsHistoryResponse(
    @SerialName("sensor_id") val sensorId: String,
    val period: String,
    val points: List<MetricReading> = emptyList(),
)

@Serializable
data class DashboardResponse(
    val dashboard: DashboardConfig? = null,
    val panels: List<DashboardPanel> = emptyList(),
)

@Serializable
data class DashboardConfig(
    val title: String? = null,
    @SerialName("refresh_sec") val refreshSec: Int? = null,
)

@Serializable
data class DashboardPanel(
    val id: String,
    val title: String,
    val type: String,
    val sensors: List<String> = emptyList(),
    val period: String? = null,
    val row: Int? = null,
)

@Serializable
data class LiveMessage(
    val type: String,
    val data: Map<String, MetricReading> = emptyMap(),
    @SerialName("agent_id") val agentId: String? = null,
)

@Serializable
data class SystemCpuInfo(
    val name: String? = null,
    val percent: Double? = null,
    @SerialName("cores_physical") val coresPhysical: Int? = null,
    @SerialName("cores_logical") val coresLogical: Int? = null,
)

@Serializable
data class SystemMemoryInfo(
    @SerialName("used_gb") val usedGb: Double? = null,
    @SerialName("total_gb") val totalGb: Double? = null,
    @SerialName("available_gb") val availableGb: Double? = null,
    @SerialName("free_gb") val freeGb: Double? = null,
    val percent: Double? = null,
)

@Serializable
data class SystemSwapInfo(
    @SerialName("used_gb") val usedGb: Double? = null,
    @SerialName("total_gb") val totalGb: Double? = null,
    val percent: Double? = null,
)

@Serializable
data class AgentSystemInfo(
    val hostname: String? = null,
    val os: String? = null,
    val cpu: SystemCpuInfo? = null,
    val memory: SystemMemoryInfo? = null,
    val swap: SystemSwapInfo? = null,
    val gpus: List<GpuInfo> = emptyList(),
)

@Serializable
data class GpuInfo(
    val name: String? = null,
    val vendor: String? = null,
    @SerialName("driver_version") val driverVersion: String? = null,
    @SerialName("memory_total_gb") val memoryTotalGb: Double? = null,
    @SerialName("memory_used_gb") val memoryUsedGb: Double? = null,
    @SerialName("memory_free_gb") val memoryFreeGb: Double? = null,
    @SerialName("memory_percent") val memoryPercent: Double? = null,
    @SerialName("utilization_percent") val utilizationPercent: Double? = null,
    @SerialName("temperature_c") val temperatureC: Double? = null,
    @SerialName("video_processor") val videoProcessor: String? = null,
    @SerialName("cuda_version") val cudaVersion: String? = null,
    @SerialName("cuda_toolkit_version") val cudaToolkitVersion: String? = null,
    val resolution: String? = null,
)

@Serializable
data class HubConfigResponse(
    val hub: HubConfig = HubConfig(),
)

@Serializable
data class HubConfig(
    val domain: String = "",
    @SerialName("public_url") val publicUrl: String = "",
    @SerialName("use_https") val useHttps: Boolean = true,
    @SerialName("trusted_hosts") val trustedHosts: List<String> = emptyList(),
    @SerialName("public_url_resolved") val publicUrlResolved: String? = null,
)

@Serializable
data class HubConfigUpdateRequest(
    val hub: HubConfig,
)

@Serializable
data class HubInfoResponse(
    val mode: String = "hub",
    val domain: String = "",
    @SerialName("public_url") val publicUrl: String = "",
    @SerialName("use_https") val useHttps: Boolean = true,
)

@Serializable
data class AgentsConfigResponse(
    val agents: List<AgentConfigEntry> = emptyList(),
)

@Serializable
data class AgentConfigEntry(
    val id: String,
    val name: String = "",
    val token: String = "",
)

@Serializable
data class AgentsConfigUpdateRequest(
    val agents: List<AgentConfigEntry>,
)

@Serializable
data class AgentsConfigUpdateResponse(
    val status: String? = null,
    val agents: List<AgentConfigEntry> = emptyList(),
)

@Serializable
data class DashboardUpdateRequest(
    val dashboard: DashboardConfig,
    val panels: List<DashboardPanel> = emptyList(),
)

@Serializable
data class VersionResponse(
    @SerialName("current_version") val currentVersion: String? = null,
    @SerialName("latest_version") val latestVersion: String? = null,
    @SerialName("update_available") val updateAvailable: Boolean = false,
)
