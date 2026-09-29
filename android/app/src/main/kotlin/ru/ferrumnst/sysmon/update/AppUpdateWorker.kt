package ru.ferrumnst.sysmon.update

import android.content.Context
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import ru.ferrumnst.sysmon.SysMonApplication
import ru.ferrumnst.sysmon.notifications.NotificationHelper
import java.util.concurrent.TimeUnit

class AppUpdateWorker(
    appContext: Context,
    params: WorkerParameters,
) : CoroutineWorker(appContext, params) {

    override suspend fun doWork(): Result {
        val app = applicationContext as SysMonApplication
        val info = app.appUpdateManager.checkForUpdate(force = true)
        if (info.updateAvailable && !info.latestVersion.isNullOrBlank()) {
            val prefs = applicationContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            val notified = prefs.getString(KEY_NOTIFIED_VERSION, null)
            if (notified != info.latestVersion) {
                NotificationHelper.showAppUpdateAvailable(
                    applicationContext,
                    info.latestVersion!!,
                )
                prefs.edit().putString(KEY_NOTIFIED_VERSION, info.latestVersion).apply()
            }
        }
        return Result.success()
    }

    companion object {
        private const val WORK_NAME = "sysmon_app_update_check"
        private const val PREFS_NAME = "sysmon_app_update"
        private const val KEY_NOTIFIED_VERSION = "notified_version"

        fun schedule(context: Context) {
            val request = PeriodicWorkRequestBuilder<AppUpdateWorker>(1, TimeUnit.HOURS)
                .setConstraints(
                    Constraints.Builder()
                        .setRequiredNetworkType(NetworkType.CONNECTED)
                        .build(),
                )
                .build()

            WorkManager.getInstance(context).enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.KEEP,
                request,
            )
        }
    }
}
