package ru.ferrumnst.sysmon.ui.navigation

object Routes {
    const val Setup = "setup"
    const val Login = "login"
    const val Main = "main"
    const val Dashboard = "dashboard"
    const val Hosts = "hosts"
    const val HostAlerts = "host_alerts/{agentId}"
    const val Settings = "settings"

    fun hostAlerts(agentId: String) = "host_alerts/$agentId"
}
