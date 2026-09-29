package ru.ferrumnst.sysmon.widget

import android.appwidget.AppWidgetManager
import android.content.Intent
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.MenuAnchorType
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.glance.appwidget.GlanceAppWidgetManager
import androidx.lifecycle.lifecycleScope
import ru.ferrumnst.sysmon.SysMonApplication
import ru.ferrumnst.sysmon.data.models.AgentInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.ui.theme.SysMonTheme
import kotlinx.coroutines.launch

private const val MAX_WIDGET_SENSORS = 6

private val defaultSensorIds = listOf(
    "cpu_percent",
    "ram_used",
    "gpu_temp",
    "net_rx",
)

class WidgetConfigureActivity : ComponentActivity() {
    private var appWidgetId: Int = AppWidgetManager.INVALID_APPWIDGET_ID

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setResult(RESULT_CANCELED)

        appWidgetId = intent?.extras?.getInt(
            AppWidgetManager.EXTRA_APPWIDGET_ID,
            AppWidgetManager.INVALID_APPWIDGET_ID,
        ) ?: AppWidgetManager.INVALID_APPWIDGET_ID

        if (appWidgetId == AppWidgetManager.INVALID_APPWIDGET_ID) {
            finish()
            return
        }

        val repository = (application as SysMonApplication).repository
        setContent {
            SysMonTheme {
                WidgetConfigureScreen(
                    appWidgetId = appWidgetId,
                    repository = repository,
                    onSave = ::saveWidget,
                    onCancel = { finish() },
                )
            }
        }
    }

    private fun saveWidget(config: WidgetConfig) {
        lifecycleScope.launch {
            val appContext = applicationContext
            WidgetConfigStore.save(appContext, appWidgetId, config)
            val manager = GlanceAppWidgetManager(appContext)
            val glanceId = manager.getGlanceIdBy(appWidgetId)
            WidgetConfigStore.saveForGlance(appContext, glanceId, config)
            MetricWidget().update(appContext, glanceId)
            WidgetUpdateWorker.schedule(appContext)

            val result = Intent().putExtra(AppWidgetManager.EXTRA_APPWIDGET_ID, appWidgetId)
            setResult(RESULT_OK, result)
            finish()
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun WidgetConfigureScreen(
    appWidgetId: Int,
    repository: HubRepository,
    onSave: (WidgetConfig) -> Unit,
    onCancel: () -> Unit,
) {
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }
    var agents by remember { mutableStateOf<List<AgentInfo>>(emptyList()) }
    var sensorOptions by remember { mutableStateOf<List<WidgetSensorOption>>(emptyList()) }
    var selectedAgentId by remember { mutableStateOf("") }
    var selectedSensorIds by remember { mutableStateOf(emptySet<String>()) }
    var agentMenuExpanded by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    val savedConfig = remember(appWidgetId) { WidgetConfigStore.load(context, appWidgetId) }

    suspend fun loadSensorOptions(agentId: String): List<WidgetSensorOption> {
        val session = repository.getSessionOnce()
        val sensors = repository.getSensors(session, agentId)
        val system = repository.getSystemInfo(session, agentId)
        return WidgetSensorCatalog.buildOptions(sensors, system)
    }

    fun pickInitialSensors(available: Set<String>, savedIds: List<String>?): Set<String> {
        val saved = savedIds?.filter { it in available }?.toSet() ?: emptySet()
        if (saved.isNotEmpty()) return saved
        val defaults = defaultSensorIds.filter { it in available }.toSet()
        if (defaults.isNotEmpty()) return defaults
        return available.take(3).toSet()
    }

    LaunchedEffect(appWidgetId) {
        isLoading = true
        error = null
        runCatching {
            val session = repository.getSessionOnce()
            if (!session.isConfigured || !session.hasCredentials) {
                error = "Сначала настройте хаб и войдите в приложении SysMon"
                return@runCatching
            }
            agents = repository.getAgents(session)
            val initialAgent = when {
                savedConfig?.agentId?.isNotBlank() == true &&
                    agents.any { it.id == savedConfig.agentId } -> savedConfig.agentId
                session.selectedAgentId.isNotBlank() &&
                    agents.any { it.id == session.selectedAgentId } -> session.selectedAgentId
                agents.size == 1 -> agents.first().id
                else -> agents.firstOrNull()?.id ?: ""
            }
            selectedAgentId = initialAgent
            if (initialAgent.isNotBlank()) {
                sensorOptions = loadSensorOptions(initialAgent)
                val available = sensorOptions.map { it.id }.toSet()
                selectedSensorIds = pickInitialSensors(available, savedConfig?.sensorIds)
            }
        }.onFailure {
            error = it.message ?: "Не удалось загрузить данные"
        }
        isLoading = false
    }

    fun loadSensorsForAgent(agentId: String, keepSelection: Boolean) {
        scope.launch {
            isLoading = true
            runCatching {
                sensorOptions = loadSensorOptions(agentId)
                val available = sensorOptions.map { it.id }.toSet()
                selectedSensorIds = if (keepSelection) {
                    val picked = selectedSensorIds.intersect(available)
                    if (picked.isNotEmpty()) picked else pickInitialSensors(available, null)
                } else {
                    pickInitialSensors(available, savedConfig?.takeIf { it.agentId == agentId }?.sensorIds)
                }
            }.onFailure {
                error = it.message
            }
            isLoading = false
        }
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        containerColor = MaterialTheme.colorScheme.background,
        bottomBar = {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                OutlinedButton(
                    onClick = onCancel,
                    modifier = Modifier.weight(1f),
                    enabled = !isLoading,
                ) {
                    Text("Отмена")
                }
                Button(
                    onClick = {
                        val agent = agents.find { it.id == selectedAgentId }
                        if (agent == null || selectedSensorIds.isEmpty()) return@Button
                        val orderedSensorIds = buildList {
                            savedConfig?.sensorIds?.forEach { id ->
                                if (id in selectedSensorIds) add(id)
                            }
                            sensorOptions.forEach { option ->
                                if (option.id in selectedSensorIds && option.id !in this) {
                                    add(option.id)
                                }
                            }
                        }.take(MAX_WIDGET_SENSORS)
                        onSave(
                            WidgetConfig(
                                agentId = agent.id,
                                agentName = agent.name ?: agent.id,
                                sensorIds = orderedSensorIds,
                            ),
                        )
                    },
                    enabled = !isLoading && selectedAgentId.isNotBlank() && selectedSensorIds.isNotEmpty(),
                    modifier = Modifier.weight(1f),
                ) {
                    Text("Сохранить")
                }
            }
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .statusBarsPadding()
                .padding(horizontal = 20.dp, vertical = 12.dp),
        ) {
            Text(
                "Виджет показателей",
                style = MaterialTheme.typography.headlineSmall,
                color = MaterialTheme.colorScheme.onSurface,
            )
            Text(
                "Выберите хост и до $MAX_WIDGET_SENSORS датчиков для отображения на рабочем столе.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp, bottom = 16.dp),
            )

            when {
                isLoading && agents.isEmpty() -> {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        CircularProgressIndicator()
                    }
                }
                error != null -> {
                    Text(error!!, color = MaterialTheme.colorScheme.error)
                }
                else -> {
                    val selectedAgent = agents.find { it.id == selectedAgentId }
                    val fieldColors = OutlinedTextFieldDefaults.colors(
                        focusedTextColor = MaterialTheme.colorScheme.onSurface,
                        unfocusedTextColor = MaterialTheme.colorScheme.onSurface,
                        disabledTextColor = MaterialTheme.colorScheme.onSurface,
                        focusedBorderColor = MaterialTheme.colorScheme.primary,
                        unfocusedBorderColor = MaterialTheme.colorScheme.outline,
                    )

                    ExposedDropdownMenuBox(
                        expanded = agentMenuExpanded,
                        onExpandedChange = { agentMenuExpanded = it },
                    ) {
                        OutlinedTextField(
                            value = selectedAgent?.name ?: selectedAgentId.ifBlank { "Выберите хост" },
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("Хост") },
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(agentMenuExpanded) },
                            colors = fieldColors,
                            modifier = Modifier
                                .fillMaxWidth()
                                .menuAnchor(MenuAnchorType.PrimaryNotEditable),
                        )
                        ExposedDropdownMenu(
                            expanded = agentMenuExpanded,
                            onDismissRequest = { agentMenuExpanded = false },
                        ) {
                            agents.forEach { agent ->
                                DropdownMenuItem(
                                    text = { Text(agent.name ?: agent.id) },
                                    onClick = {
                                        agentMenuExpanded = false
                                        if (selectedAgentId != agent.id) {
                                            selectedAgentId = agent.id
                                            loadSensorsForAgent(agent.id, keepSelection = false)
                                        }
                                    },
                                )
                            }
                        }
                    }

                    Text(
                        "Выбрано ${selectedSensorIds.size} из $MAX_WIDGET_SENSORS",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(top = 12.dp, bottom = 8.dp),
                    )

                    if (isLoading) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .weight(1f),
                            contentAlignment = Alignment.Center,
                        ) {
                            CircularProgressIndicator()
                        }
                    } else if (sensorOptions.isEmpty()) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .weight(1f),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                "Нет доступных датчиков для этого хоста",
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    } else {
                        LazyColumn(
                            modifier = Modifier.weight(1f),
                            verticalArrangement = Arrangement.spacedBy(4.dp),
                        ) {
                            items(sensorOptions, key = { it.id }) { option ->
                                val checked = option.id in selectedSensorIds
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 2.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    Checkbox(
                                        checked = checked,
                                        onCheckedChange = { enabled ->
                                            selectedSensorIds = when {
                                                enabled && selectedSensorIds.size >= MAX_WIDGET_SENSORS -> {
                                                    Toast.makeText(
                                                        context,
                                                        "Максимум $MAX_WIDGET_SENSORS показателей",
                                                        Toast.LENGTH_SHORT,
                                                    ).show()
                                                    selectedSensorIds
                                                }
                                                enabled -> selectedSensorIds + option.id
                                                else -> selectedSensorIds - option.id
                                            }
                                        },
                                    )
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(
                                            option.name,
                                            style = MaterialTheme.typography.bodyMedium,
                                            color = MaterialTheme.colorScheme.onSurface,
                                        )
                                        Text(
                                            option.id,
                                            style = MaterialTheme.typography.labelSmall,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                        )
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
