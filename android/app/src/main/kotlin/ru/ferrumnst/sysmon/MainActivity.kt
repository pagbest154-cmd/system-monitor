package ru.ferrumnst.sysmon

import android.content.Intent
import android.os.Bundle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.fragment.app.FragmentActivity
import kotlinx.coroutines.runBlocking
import ru.ferrumnst.sysmon.notifications.NotificationHelper
import ru.ferrumnst.sysmon.ui.SysMonApp
import ru.ferrumnst.sysmon.ui.theme.SysMonTheme

class MainActivity : FragmentActivity() {
    private var launchAgentId by mutableStateOf<String?>(null)
    private var openSettingsUpdate by mutableStateOf(false)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        handleLaunchExtras(intent)
        val app = application as SysMonApplication
        setContent {
            val prefs by app.appPreferencesStore.preferences.collectAsState(
                initial = ru.ferrumnst.sysmon.data.session.AppPreferences(),
            )
            SysMonTheme(darkTheme = prefs.useDarkTheme) {
                SysMonApp(
                    repository = app.repository,
                    appPreferencesStore = app.appPreferencesStore,
                    launchAgentId = launchAgentId,
                    openSettingsUpdate = openSettingsUpdate,
                    onLaunchAgentHandled = { launchAgentId = null },
                    onOpenSettingsUpdateHandled = { openSettingsUpdate = false },
                )
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleLaunchExtras(intent)
    }

    private fun handleLaunchExtras(intent: Intent?) {
        val agentId = intent?.getStringExtra(NotificationHelper.EXTRA_AGENT_ID)?.takeIf { it.isNotBlank() }
        launchAgentId = agentId
        if (!agentId.isNullOrBlank()) {
            runBlocking {
                (application as SysMonApplication).repository.saveSelectedAgent(agentId)
            }
        }
        if (intent?.getBooleanExtra(NotificationHelper.EXTRA_OPEN_SETTINGS_UPDATE, false) == true) {
            openSettingsUpdate = true
        }
    }
}
