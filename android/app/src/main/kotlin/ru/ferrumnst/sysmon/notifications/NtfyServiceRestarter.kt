package ru.ferrumnst.sysmon.notifications

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

/** Restarts ntfy foreground listener after reboot or APK update. */
class NtfyServiceRestarter : BroadcastReceiver() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    override fun onReceive(context: Context, intent: Intent?) {
        val action = intent?.action ?: return
        if (
            action != Intent.ACTION_BOOT_COMPLETED &&
            action != Intent.ACTION_MY_PACKAGE_REPLACED
        ) {
            return
        }
        val app = context.applicationContext
        scope.launch {
            val subs = NtfySubscriptionStore(app).getAll()
            if (subs.isNotEmpty()) {
                NtfySubscriptionManager.refreshService(app)
            }
        }
    }
}
