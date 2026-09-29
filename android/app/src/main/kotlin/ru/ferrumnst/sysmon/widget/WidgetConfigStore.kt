package ru.ferrumnst.sysmon.widget

import android.content.Context
import androidx.core.content.edit
import androidx.glance.appwidget.GlanceAppWidgetManager
import androidx.glance.GlanceId

object WidgetConfigStore {
    private const val PREFS = "widget_configs"
    private const val MAX_SENSORS = 6

    fun save(context: Context, widgetId: Int, config: WidgetConfig) {
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit {
            putString(keyAgent(widgetId), config.agentId)
            putString(keyAgentName(widgetId), config.agentName)
            putString(keySensors(widgetId), config.sensorIds.take(MAX_SENSORS).joinToString(","))
        }
    }

    fun saveForGlance(context: Context, glanceId: GlanceId, config: WidgetConfig) {
        val glanceKey = glanceId.toString()
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit {
            putString(keyGlanceAgent(glanceKey), config.agentId)
            putString(keyGlanceAgentName(glanceKey), config.agentName)
            putString(keyGlanceSensors(glanceKey), config.sensorIds.take(MAX_SENSORS).joinToString(","))
        }
    }

    fun load(context: Context, widgetId: Int): WidgetConfig? {
        val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
        val agentId = prefs.getString(keyAgent(widgetId), null) ?: return null
        val agentName = prefs.getString(keyAgentName(widgetId), agentId) ?: agentId
        val sensors = prefs.getString(keySensors(widgetId), null)
            ?.split(",")
            ?.map { it.trim() }
            ?.filter { it.isNotEmpty() }
            ?: emptyList()
        if (sensors.isEmpty()) return null
        return WidgetConfig(agentId, agentName, sensors)
    }

    suspend fun load(context: Context, glanceId: GlanceId): WidgetConfig? {
        loadByGlanceKey(context, glanceId.toString())?.let { return it }
        val widgetId = GlanceAppWidgetManager(context).getAppWidgetId(glanceId)
        return load(context, widgetId)
    }

    private fun loadByGlanceKey(context: Context, glanceKey: String): WidgetConfig? {
        val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
        val agentId = prefs.getString(keyGlanceAgent(glanceKey), null) ?: return null
        val agentName = prefs.getString(keyGlanceAgentName(glanceKey), agentId) ?: agentId
        val sensors = prefs.getString(keyGlanceSensors(glanceKey), null)
            ?.split(",")
            ?.map { it.trim() }
            ?.filter { it.isNotEmpty() }
            ?: emptyList()
        if (sensors.isEmpty()) return null
        return WidgetConfig(agentId, agentName, sensors)
    }

    fun delete(context: Context, widgetId: Int) {
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit {
            remove(keyAgent(widgetId))
            remove(keyAgentName(widgetId))
            remove(keySensors(widgetId))
        }
    }

    private fun keyAgent(widgetId: Int) = "agent_$widgetId"
    private fun keyAgentName(widgetId: Int) = "agent_name_$widgetId"
    private fun keySensors(widgetId: Int) = "sensors_$widgetId"
    private fun keyGlanceAgent(glanceKey: String) = "glance_agent_$glanceKey"
    private fun keyGlanceAgentName(glanceKey: String) = "glance_agent_name_$glanceKey"
    private fun keyGlanceSensors(glanceKey: String) = "glance_sensors_$glanceKey"
}
