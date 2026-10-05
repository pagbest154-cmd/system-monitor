package ru.ferrumnst.sysmon.ui.util

import ru.ferrumnst.sysmon.data.models.AgentConfigEntry
import ru.ferrumnst.sysmon.data.models.SensorOverrideEntry

fun getNetworkInterfaceOverride(overrides: List<SensorOverrideEntry>): String {
    for (sensorId in listOf("net_rx", "net_tx")) {
        val iface = overrides.find { it.sensorId == sensorId }?.params?.get("interface")
        if (!iface.isNullOrBlank()) return iface.trim()
    }
    return ""
}

fun applyNetworkInterfaceOverrides(
    overrides: List<SensorOverrideEntry>,
    iface: String,
): List<SensorOverrideEntry> {
    val kept = overrides.filter { it.sensorId != "net_rx" && it.sensorId != "net_tx" }
    val trimmed = iface.trim()
    if (trimmed.isEmpty()) return kept
    return kept + listOf(
        SensorOverrideEntry(sensorId = "net_rx", params = mapOf("interface" to trimmed)),
        SensorOverrideEntry(sensorId = "net_tx", params = mapOf("interface" to trimmed)),
    )
}

fun updateAgentNetworkInterface(
    agents: List<AgentConfigEntry>,
    agentId: String,
    iface: String,
): List<AgentConfigEntry> {
    return agents.map { entry ->
        if (entry.id != agentId) entry
        else entry.copy(overrides = applyNetworkInterfaceOverrides(entry.overrides, iface))
    }
}
