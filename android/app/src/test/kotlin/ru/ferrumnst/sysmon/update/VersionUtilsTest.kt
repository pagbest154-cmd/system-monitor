package ru.ferrumnst.sysmon.update

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class VersionUtilsTest {
    @Test
    fun isNewer_comparesSemanticVersions() {
        assertTrue(VersionUtils.isNewer("1.0.44", "1.0.43"))
        assertTrue(VersionUtils.isNewer("1.1.0", "1.0.99"))
        assertFalse(VersionUtils.isNewer("1.0.43", "1.0.43"))
        assertFalse(VersionUtils.isNewer("1.0.42", "1.0.43"))
    }

    @Test
    fun apkAssetName_matchesReleaseNaming() {
        assertTrue(VersionUtils.apkAssetName("v1.0.43") == "sysmon-1.0.43.apk")
        assertTrue(VersionUtils.apkAssetName("v1.0.58-fix") == "sysmon-1.0.58-fix.apk")
    }
}
