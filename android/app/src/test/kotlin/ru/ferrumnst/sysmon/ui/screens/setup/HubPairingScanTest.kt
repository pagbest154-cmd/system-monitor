package ru.ferrumnst.sysmon.ui.screens.setup

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class HubPairingScanTest {
    @Test
    fun parsePlainUrl() {
        val scan = parseHubPairingScan("https://mon.example.com/")
        assertEquals("https://mon.example.com", scan.hubUrl)
        assertTrue(!scan.includesCredentials)
    }

    @Test
    fun parsePairingUri() {
        val raw =
            "sysmon://pair?url=https%3A%2F%2Fmon.example.com&name=hub1&key=secret-key"
        val scan = parseHubPairingScan(raw)
        assertEquals("https://mon.example.com", scan.hubUrl)
        assertEquals("hub1", scan.hubName)
        assertEquals("secret-key", scan.hubKey)
        assertTrue(scan.includesCredentials)
    }
}
