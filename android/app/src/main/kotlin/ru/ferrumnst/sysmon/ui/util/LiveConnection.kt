package ru.ferrumnst.sysmon.ui.util

const val CONNECTION_OFFLINE_SEC = 180.0

fun isEffectivelyLive(
    wsConnected: Boolean,
    lastUpdateTs: Double,
    agentOnline: Boolean,
    nowSec: Double = System.currentTimeMillis() / 1000.0,
): Boolean {
    if (!wsConnected || !agentOnline) return false
    if (lastUpdateTs <= 0) return false
    return (nowSec - lastUpdateTs) < CONNECTION_OFFLINE_SEC
}
