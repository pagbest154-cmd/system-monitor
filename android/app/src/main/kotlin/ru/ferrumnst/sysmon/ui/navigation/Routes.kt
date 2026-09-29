package ru.ferrumnst.sysmon.ui.navigation

import android.net.Uri

object Routes {
    const val Setup = "setup"
    const val Login = "login"
    const val Main = "main"
    const val Dashboard = "dashboard"
    const val Hosts = "hosts"
    const val HostAlerts = "host_alerts/{agentId}/{agentName}"
    const val Settings = "settings"

    fun hostAlerts(agentId: String, agentName: String) =
        "host_alerts/${Uri.encode(agentId)}/${Uri.encode(agentName)}"
}
