package ru.ferrumnst.sysmon.ui.util

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LiveConnectionTest {
    @Test
    fun offlineWhenWebSocketDisconnected() {
        assertFalse(isEffectivelyLive(wsConnected = false, lastUpdateTs = 100.0, agentOnline = true, nowSec = 150.0))
    }

    @Test
    fun offlineWhenAgentOffline() {
        assertFalse(isEffectivelyLive(wsConnected = true, lastUpdateTs = 100.0, agentOnline = false, nowSec = 150.0))
    }

    @Test
    fun offlineWhenNoDataYet() {
        assertFalse(isEffectivelyLive(wsConnected = true, lastUpdateTs = 0.0, agentOnline = true, nowSec = 150.0))
    }

    @Test
    fun offlineWhenDataStale() {
        assertFalse(isEffectivelyLive(wsConnected = true, lastUpdateTs = 100.0, agentOnline = true, nowSec = 281.0))
    }

    @Test
    fun liveWhenConnectedWithFreshData() {
        assertTrue(isEffectivelyLive(wsConnected = true, lastUpdateTs = 200.0, agentOnline = true, nowSec = 250.0))
    }
}
