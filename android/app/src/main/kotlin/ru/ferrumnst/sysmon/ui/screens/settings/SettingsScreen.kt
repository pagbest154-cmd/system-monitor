package ru.ferrumnst.sysmon.ui.screens.settings

import android.app.Activity
import android.app.Application
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import ru.ferrumnst.sysmon.SysMonApplication
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.AppPreferencesStore
import ru.ferrumnst.sysmon.data.session.HubSession
import ru.ferrumnst.sysmon.ui.components.SysMonBottomBarClearance

@Composable
fun SettingsScreen(
    repository: HubRepository,
    appPreferencesStore: AppPreferencesStore,
    initialTab: SettingsTab = SettingsTab.App,
    onLogout: () -> Unit,
    onResetHub: () -> Unit,
) {
    val app = LocalContext.current.applicationContext as Application
    val sysMonApp = app as SysMonApplication
    val vm: SettingsViewModel = viewModel(
        factory = SettingsViewModel.Factory(app, repository, sysMonApp.appUpdateManager),
    )
    val state by vm.uiState.collectAsState()
    val session by repository.session.collectAsState(initial = HubSession())
    val appPrefs by appPreferencesStore.preferences.collectAsState(
        initial = ru.ferrumnst.sysmon.data.session.AppPreferences(),
    )
    val scope = rememberCoroutineScope()
    val activity = LocalContext.current as? Activity

    var selectedTab by remember { mutableStateOf(initialTab) }
    LaunchedEffect(initialTab) {
        selectedTab = initialTab
    }

    val tabs = SettingsTab.entries

    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding(),
    ) {
        Text(
            "Настройки",
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.onBackground,
            modifier = Modifier.padding(start = 16.dp, end = 16.dp, top = 12.dp),
        )

        TabRow(
            selectedTabIndex = tabs.indexOf(selectedTab).coerceAtLeast(0),
            modifier = Modifier.padding(top = 8.dp),
        ) {
            tabs.forEach { tab ->
                Tab(
                    selected = selectedTab == tab,
                    onClick = { selectedTab = tab },
                    text = { Text(tab.title) },
                )
            }
        }

        if (state.isLoading) {
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(bottom = SysMonBottomBarClearance),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator()
            }
        } else {
        Column(
            modifier = Modifier
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(
                    start = 16.dp,
                    end = 16.dp,
                    top = 12.dp,
                    bottom = 16.dp + SysMonBottomBarClearance,
                ),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            state.message?.let {
                Text(it, color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.bodyMedium)
            }
            state.error?.let {
                Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
            }

            when (selectedTab) {
                SettingsTab.App -> AppSettingsTab(
                    state = state,
                    session = session,
                    appPrefs = appPrefs,
                    appPreferencesStore = appPreferencesStore,
                    repository = repository,
                    vm = vm,
                    scope = scope,
                    activity = activity,
                    onLogout = onLogout,
                    onResetHub = onResetHub,
                )
                SettingsTab.Hub -> HubSettingsTab(state = state, vm = vm)
            }
        }
        }
    }
}
