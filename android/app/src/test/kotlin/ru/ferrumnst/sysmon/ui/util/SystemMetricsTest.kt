package ru.ferrumnst.sysmon.ui.util

import org.junit.Assert.assertEquals
import org.junit.Test
import ru.ferrumnst.sysmon.data.models.SystemMemoryInfo

class SystemMetricsTest {
    @Test
    fun formatResourceLine_showsGbAndPercent() {
        val line = formatResourceLine(8.7, 16.0, 54.4)
        assertEquals("8.7 / 16 ГБ · 54%", line)
    }

    @Test
    fun memoryDisplayLine_fromModel() {
        val memory = SystemMemoryInfo(usedGb = 4.0, totalGb = 8.0, percent = 50.0)
        assertEquals("4 / 8 ГБ · 50%", memory.displayLine())
    }
}
