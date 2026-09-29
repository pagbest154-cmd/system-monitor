package ru.ferrumnst.sysmon

import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import kotlinx.coroutines.runBlocking
import ru.ferrumnst.sysmon.notifications.NotificationHelper
import ru.ferrumnst.sysmon.ui.SysMonApp
import ru.ferrumnst.sysmon.ui.theme.SysMonTheme

class MainActivity : ComponentActivity() {
    private var launchAgentId by mutableStateOf<String?>(null)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        handleLaunchAgent(intent)
        val repository = (application as SysMonApplication).repository
        setContent {
            SysMonTheme {
                SysMonApp(
                    repository = repository,
                    launchAgentId = launchAgentId,
                    onLaunchAgentHandled = { launchAgentId = null },
                )
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleLaunchAgent(intent)
    }

    private fun handleLaunchAgent(intent: Intent?) {
        val agentId = intent.agentIdExtra()
        launchAgentId = agentId
        if (!agentId.isNullOrBlank()) {
            runBlocking {
                (application as SysMonApplication).repository.saveSelectedAgent(agentId)
            }
        }
    }

    private fun Intent?.agentIdExtra(): String? {
        return this?.getStringExtra(NotificationHelper.EXTRA_AGENT_ID)?.takeIf { it.isNotBlank() }
    }
}
