package ru.ferrumnst.sysmon.data.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AlertOfflineRule(
    val enabled: Boolean = false,
    @SerialName("after_sec") val afterSec: Int = 180,
)

@Serializable
data class AlertSensorRule(
    @SerialName("sensor_id") val sensorId: String,
    val enabled: Boolean = false,
    val threshold: Double = 0.0,
)

@Serializable
data class NtfyAlertConfig(
    val topic: String = "",
    val token: String = "",
)

@Serializable
data class AgentAlertConfig(
    val enabled: Boolean = false,
    @SerialName("threshold_mode") val thresholdMode: String = "manual",
    val offline: AlertOfflineRule = AlertOfflineRule(),
    val sensors: List<AlertSensorRule> = emptyList(),
    @SerialName("cooldown_sec") val cooldownSec: Int = 900,
    @SerialName("notify_recovery") val notifyRecovery: Boolean = false,
    val ntfy: NtfyAlertConfig = NtfyAlertConfig(),
)

@Serializable
data class AgentAlertsResponse(
    @SerialName("agent_id") val agentId: String,
    val config: AgentAlertConfig,
    @SerialName("ntfy_base_url") val ntfyBaseUrl: String = "",
)
