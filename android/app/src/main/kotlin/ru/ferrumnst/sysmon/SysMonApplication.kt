package ru.ferrumnst.sysmon

import android.app.Application
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.AppPreferencesStore
import ru.ferrumnst.sysmon.data.session.OfflineCacheStore
import ru.ferrumnst.sysmon.data.session.SessionStore
import ru.ferrumnst.sysmon.data.websocket.LiveWebSocketClient
import ru.ferrumnst.sysmon.notifications.NotificationHelper
import ru.ferrumnst.sysmon.update.AppUpdateManager
import ru.ferrumnst.sysmon.update.AppUpdateWorker
import ru.ferrumnst.sysmon.widget.WidgetUpdateWorker
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class SysMonApplication : Application() {
    private val appScope = CoroutineScope(SupervisorJob())

    lateinit var repository: HubRepository
        private set

    lateinit var appUpdateManager: AppUpdateManager
        private set

    lateinit var appPreferencesStore: AppPreferencesStore
        private set

    override fun onCreate() {
        super.onCreate()
        NotificationHelper.ensureChannel(this)
        val sessionStore = SessionStore(this)
        appPreferencesStore = AppPreferencesStore(this)
        val offlineCache = OfflineCacheStore(this)
        val liveClient = LiveWebSocketClient(appScope)
        repository = HubRepository(sessionStore, offlineCache, liveClient)
        appUpdateManager = AppUpdateManager(this)
        WidgetUpdateWorker.schedule(this)
        AppUpdateWorker.schedule(this)
        appScope.launch {
            WidgetUpdateWorker.refreshNow(this@SysMonApplication)
            appUpdateManager.checkForUpdate()
        }
    }
}
