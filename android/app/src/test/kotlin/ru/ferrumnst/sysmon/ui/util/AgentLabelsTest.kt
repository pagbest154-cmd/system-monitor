package ru.ferrumnst.sysmon.ui.util

import org.junit.Assert.assertEquals
import org.junit.Test
import ru.ferrumnst.sysmon.data.models.AgentInfo

class AgentLabelsTest {
    @Test
    fun formatOsName_windows() {
        assertEquals("Windows 11", formatOsName("Microsoft Windows 11 Pro"))
    }

    @Test
    fun osLabel_prefersSystemOs() {
        val agent = AgentInfo(
            id = "host1",
            platform = "linux",
            system = kotlinx.serialization.json.Json.parseToJsonElement("""{"os":"ubuntu 24.04"}"""),
        )
        assertEquals("Ubuntu 24.04", agent.osLabel())
    }

    @Test
    fun versionAndOsText_combinesBoth() {
        val agent = AgentInfo(
            id = "host1",
            agentVersion = "1.0.43",
            platform = "linux",
            system = kotlinx.serialization.json.Json.parseToJsonElement("""{"os":"ubuntu 24.04"}"""),
        )
        assertEquals("v1.0.43 · Ubuntu 24.04", agent.versionAndOsText())
    }
}
