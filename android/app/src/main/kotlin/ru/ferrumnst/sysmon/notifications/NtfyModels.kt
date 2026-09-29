package ru.ferrumnst.sysmon.notifications

import kotlinx.serialization.Serializable

@Serializable
data class NtfyMessage(
    val event: String? = null,
    val message: String? = null,
    val title: String? = null,
    val topic: String? = null,
    val priority: Int? = null,
    val headers: Map<String, String>? = null,
)

@Serializable
data class NtfySubscription(
    val agentId: String,
    val baseUrl: String,
    val topic: String,
    val token: String = "",
) {
    val key: String
        get() = "$baseUrl|$topic"

    fun wsUrl(): String {
        val trimmed = baseUrl.trimEnd('/')
        val wsBase = when {
            trimmed.startsWith("https://") -> "wss://${trimmed.removePrefix("https://")}"
            trimmed.startsWith("http://") -> "ws://${trimmed.removePrefix("http://")}"
            else -> "wss://$trimmed"
        }
        return "$wsBase/$topic/ws"
    }
}
