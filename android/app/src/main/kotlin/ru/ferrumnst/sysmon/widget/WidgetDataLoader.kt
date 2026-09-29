package ru.ferrumnst.sysmon.widget

import android.content.Context
import ru.ferrumnst.sysmon.data.api.HubClientFactory
import ru.ferrumnst.sysmon.data.models.AgentSystemInfo
import ru.ferrumnst.sysmon.data.session.SessionStore
import kotlinx.coroutines.flow.first

object WidgetDataLoader {
    suspend fun load(context: Context, config: WidgetConfig?): WidgetDisplayData {
        if (config == null) {
            return WidgetDisplayData.error("Откройте настройки виджета и выберите хост с датчиками")
        }

        val session = SessionStore(context).session.first()
        if (!session.isConfigured) {
            return WidgetDisplayData.error("Сначала настройте хаб в приложении")
        }
        if (!session.hasCredentials) {
            return WidgetDisplayData.error("Войдите в приложение SysMon")
        }

        return runCatching {
            val api = HubClientFactory.createApi(session)
            val sensors = api.sensors(config.agentId).sensors
            val system = runCatching { api.system(config.agentId) }.getOrElse { AgentSystemInfo() }
            val metrics = config.sensorIds.map { sensorId ->
                WidgetSensorCatalog.resolveMetric(sensorId, sensors, system.gpus)
                    ?: WidgetMetricItem(
                        id = sensorId,
                        name = sensorId,
                        value = null,
                        unit = null,
                        status = "unknown",
                    )
            }

            if (metrics.isEmpty()) {
                WidgetDisplayData.error("Нет данных по выбранным датчикам")
            } else {
                WidgetDisplayData(
                    agentName = config.agentName,
                    metrics = metrics,
                    updatedAt = System.currentTimeMillis(),
                )
            }
        }.getOrElse { error ->
            WidgetDisplayData.error(error.message ?: "Ошибка загрузки")
        }
    }
}
