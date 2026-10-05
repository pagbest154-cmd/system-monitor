package ru.ferrumnst.sysmon.ui.screens.hosts

import android.app.Application
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.data.repository.HubRepository

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HostNetworkScreen(
    repository: HubRepository,
    agentId: String,
    agentName: String,
    onBack: () -> Unit,
) {
    val app = LocalContext.current.applicationContext as Application
    val vm: HostNetworkViewModel = viewModel(
        factory = HostNetworkViewModel.Factory(app, repository, agentId, agentName),
    )
    val state by vm.uiState.collectAsState()

    LaunchedEffect(state.saved) {
        if (state.saved) onBack()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .statusBarsPadding()
                .padding(horizontal = 4.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Назад")
            }
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    "Сетевой трафик",
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                )
                Text(
                    state.agentName,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }

        if (state.isLoading) {
            CircularProgressIndicator(modifier = Modifier.padding(24.dp))
            return
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            state.error?.let { msg ->
                Text(msg, color = MaterialTheme.colorScheme.error)
            }

            if (state.notInConfig) {
                Text(
                    "Хост не найден в agents.yaml на hub. Добавьте его в Настройки → Агенты.",
                    color = MaterialTheme.colorScheme.error,
                )
            } else {
                var expanded by remember { mutableStateOf(false) }
                ExposedDropdownMenuBox(
                    expanded = expanded,
                    onExpandedChange = { expanded = !expanded },
                ) {
                    OutlinedTextField(
                        value = state.selectedInterface.ifBlank { "Все интерфейсы (сумма)" },
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("Интерфейс") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                        modifier = Modifier
                            .menuAnchor()
                            .fillMaxWidth(),
                    )
                    ExposedDropdownMenu(
                        expanded = expanded,
                        onDismissRequest = { expanded = false },
                    ) {
                        DropdownMenuItem(
                            text = { Text("Все интерфейсы (сумма)") },
                            onClick = {
                                vm.setSelectedInterface("")
                                expanded = false
                            },
                        )
                        state.interfaces.forEach { iface ->
                            val name = iface.name ?: return@forEach
                            DropdownMenuItem(
                                text = { Text(name) },
                                onClick = {
                                    vm.setSelectedInterface(name)
                                    expanded = false
                                },
                            )
                        }
                    }
                }

                if (state.interfaces.isEmpty()) {
                    Text(
                        "Список интерфейсов пока недоступен — введите имя вручную.",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }

                OutlinedTextField(
                    value = state.customInterface,
                    onValueChange = { vm.setCustomInterface(it) },
                    label = { Text("Другое имя (вручную)") },
                    placeholder = { Text("eth0, Wi-Fi, …") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                )

                Text(
                    "Применяется к net_rx и net_tx. Агент подхватит после sync конфига с hub.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )

                Button(
                    onClick = vm::save,
                    enabled = !state.isSaving,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    if (state.isSaving) {
                        CircularProgressIndicator(
                            modifier = Modifier.padding(end = 8.dp),
                            strokeWidth = 2.dp,
                        )
                    }
                    Text("Сохранить")
                }
            }
        }
    }
}
