package ru.ferrumnst.sysmon.ui.screens.hubunavailable

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import ru.ferrumnst.sysmon.data.models.AgentInfo
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.HubSession
import ru.ferrumnst.sysmon.ui.util.formatTimestamp
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

@Composable
fun HubUnavailableScreen(
    repository: HubRepository,
    onRetry: () -> Unit,
    onChangeHub: () -> Unit,
) {
    val session by repository.session.collectAsState(initial = HubSession())
    var cachedAgents by remember { mutableStateOf<List<AgentInfo>?>(null) }
    var cachedAt by remember { mutableStateOf<Long?>(null) }

    LaunchedEffect(Unit) {
        cachedAgents = repository.getCachedAgents()
        cachedAt = repository.getCachedAgentsAtMillis()
    }

    Scaffold { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState())
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = "Хаб недоступен",
                style = MaterialTheme.typography.headlineMedium,
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Не удалось подключиться к хабу",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            if (session.hubUrl.isNotBlank()) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = session.hubUrl,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            cachedAt?.let { at ->
                val formatted = SimpleDateFormat("dd.MM.yyyy HH:mm", Locale.getDefault()).format(Date(at))
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "Снимок от $formatted",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            if (!cachedAgents.isNullOrEmpty()) {
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = "Последние известные хосты",
                    style = MaterialTheme.typography.titleSmall,
                    modifier = Modifier.fillMaxWidth(),
                )
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    cachedAgents!!.forEach { agent ->
                        Card(
                            colors = CardDefaults.cardColors(
                                containerColor = MaterialTheme.colorScheme.surfaceVariant,
                            ),
                        ) {
                            Column(modifier = Modifier.padding(12.dp)) {
                                Text(
                                    agent.name ?: agent.id,
                                    fontWeight = FontWeight.SemiBold,
                                )
                                Text(
                                    agent.status ?: "unknown",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                )
                                agent.lastSeen?.let {
                                    Text(
                                        formatTimestamp(it),
                                        style = MaterialTheme.typography.labelSmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    )
                                }
                            }
                        }
                    }
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Button(
                onClick = onRetry,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Повторить")
            }
            Spacer(modifier = Modifier.height(12.dp))
            OutlinedButton(
                onClick = onChangeHub,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Сменить хаб")
            }
        }
    }
}
