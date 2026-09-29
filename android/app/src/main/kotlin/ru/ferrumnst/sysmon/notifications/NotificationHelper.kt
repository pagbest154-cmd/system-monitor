package ru.ferrumnst.sysmon.notifications

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import ru.ferrumnst.sysmon.MainActivity
import ru.ferrumnst.sysmon.R

object NotificationHelper {
    const val CHANNEL_ID = "sysmon_alerts"
    const val LISTENER_CHANNEL_ID = "sysmon_ntfy_listener_min"
    const val UPDATE_CHANNEL_ID = "sysmon_app_updates"
    private const val APP_UPDATE_NOTIFICATION_ID = 9001

    fun ensureChannel(context: Context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        val alerts = NotificationChannel(
            CHANNEL_ID,
            "Уведомления SysMon",
            NotificationManager.IMPORTANCE_HIGH,
        ).apply {
            description = "Алерты по хостам и датчикам"
        }
        val listener = NotificationChannel(
            LISTENER_CHANNEL_ID,
            "Фоновые алерты",
            NotificationManager.IMPORTANCE_MIN,
        ).apply {
            description = "Поддержка подписки на уведомления в фоне"
            setShowBadge(false)
        }
        val updates = NotificationChannel(
            UPDATE_CHANNEL_ID,
            "Обновления приложения",
            NotificationManager.IMPORTANCE_DEFAULT,
        ).apply {
            description = "Доступна новая версия SysMon"
        }
        manager.createNotificationChannel(alerts)
        manager.createNotificationChannel(listener)
        manager.createNotificationChannel(updates)
    }

    fun buildListenerNotification(context: Context): android.app.Notification {
        ensureChannel(context)
        return NotificationCompat.Builder(context, LISTENER_CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentTitle("Слушаю алерты")
            .setOngoing(true)
            .setSilent(true)
            .setPriority(NotificationCompat.PRIORITY_MIN)
            .setCategory(NotificationCompat.CATEGORY_SERVICE)
            .build()
    }

    fun showAlert(
        context: Context,
        notificationId: Int,
        title: String,
        body: String,
        agentId: String?,
    ) {
        ensureChannel(context)
        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or
                Intent.FLAG_ACTIVITY_CLEAR_TOP or
                Intent.FLAG_ACTIVITY_SINGLE_TOP
            if (!agentId.isNullOrBlank()) {
                putExtra(EXTRA_AGENT_ID, agentId)
            }
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            notificationId,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentTitle(title)
            .setContentText(body)
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .build()
        NotificationManagerCompat.from(context).notify(notificationId, notification)
    }

    fun showAppUpdateAvailable(context: Context, latestVersion: String) {
        ensureChannel(context)
        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or
                Intent.FLAG_ACTIVITY_CLEAR_TOP or
                Intent.FLAG_ACTIVITY_SINGLE_TOP
            putExtra(EXTRA_OPEN_SETTINGS_UPDATE, true)
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            APP_UPDATE_NOTIFICATION_ID,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        val notification = NotificationCompat.Builder(context, UPDATE_CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentTitle("Доступно обновление SysMon")
            .setContentText("Версия $latestVersion — откройте Настройки")
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .build()
        NotificationManagerCompat.from(context).notify(APP_UPDATE_NOTIFICATION_ID, notification)
    }

    const val EXTRA_AGENT_ID = "agent_id"
    const val EXTRA_OPEN_SETTINGS_UPDATE = "open_settings_update"
}
