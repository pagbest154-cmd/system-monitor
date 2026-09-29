package ru.ferrumnst.sysmon.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import android.app.Application
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import android.net.Uri
import ru.ferrumnst.sysmon.data.repository.HubRepository
import ru.ferrumnst.sysmon.data.session.AppPreferencesStore
import ru.ferrumnst.sysmon.ui.security.BiometricGate
import ru.ferrumnst.sysmon.ui.components.SysMonBottomBar
import ru.ferrumnst.sysmon.ui.components.SysMonBottomTabs
import ru.ferrumnst.sysmon.ui.navigation.Routes
import ru.ferrumnst.sysmon.ui.screens.dashboard.DashboardScreen
import ru.ferrumnst.sysmon.ui.screens.hosts.HostAlertsScreen
import ru.ferrumnst.sysmon.ui.screens.hosts.HostsScreen
import ru.ferrumnst.sysmon.ui.screens.hubunavailable.HubUnavailableScreen
import ru.ferrumnst.sysmon.ui.screens.login.LoginScreen
import ru.ferrumnst.sysmon.ui.screens.settings.SettingsScreen
import ru.ferrumnst.sysmon.ui.screens.settings.SettingsTab
import ru.ferrumnst.sysmon.ui.screens.setup.SetupScreen
import ru.ferrumnst.sysmon.ui.viewmodel.AppStartDestination
import ru.ferrumnst.sysmon.ui.viewmodel.AppViewModel

@Composable
fun SysMonApp(
    repository: HubRepository,
    appPreferencesStore: AppPreferencesStore,
    launchAgentId: String? = null,
    openSettingsUpdate: Boolean = false,
    onLaunchAgentHandled: () -> Unit = {},
    onOpenSettingsUpdateHandled: () -> Unit = {},
) {
    val app = LocalContext.current.applicationContext as Application
    val appViewModel: AppViewModel = viewModel(factory = AppViewModel.Factory(app, repository))
    val startDestination by appViewModel.startDestination.collectAsState()

    when (startDestination) {
        AppStartDestination.Loading -> {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator()
            }
        }
        AppStartDestination.Setup -> {
            SetupScreen(
                repository = repository,
                onComplete = appViewModel::onSetupComplete,
            )
        }
        AppStartDestination.Login -> {
            LoginScreen(
                repository = repository,
                onSuccess = appViewModel::onLoginComplete,
            )
        }
        AppStartDestination.HubUnavailable -> {
            HubUnavailableScreen(
                repository = repository,
                onRetry = appViewModel::resolveStartDestination,
                onChangeHub = appViewModel::onChangeHub,
            )
        }
        AppStartDestination.Main -> {
            val prefs by appPreferencesStore.preferences.collectAsState(
                initial = ru.ferrumnst.sysmon.data.session.AppPreferences(),
            )
            val session by repository.session.collectAsState(
                initial = ru.ferrumnst.sysmon.data.session.HubSession(),
            )
            BiometricGate(enabled = prefs.biometricLockEnabled && session.hasCredentials) {
                MainScreen(
                    repository = repository,
                    appPreferencesStore = appPreferencesStore,
                    launchAgentId = launchAgentId,
                    openSettingsUpdate = openSettingsUpdate,
                    onLaunchAgentHandled = onLaunchAgentHandled,
                    onOpenSettingsUpdateHandled = onOpenSettingsUpdateHandled,
                    onLogout = appViewModel::onLogout,
                    onResetHub = appViewModel::resolveStartDestination,
                )
            }
        }
    }
}

@Composable
private fun MainScreen(
    repository: HubRepository,
    appPreferencesStore: AppPreferencesStore,
    launchAgentId: String?,
    openSettingsUpdate: Boolean,
    onLaunchAgentHandled: () -> Unit,
    onOpenSettingsUpdateHandled: () -> Unit,
    onLogout: () -> Unit,
    onResetHub: () -> Unit,
) {
    val navController = rememberNavController()
    val backStack by navController.currentBackStackEntryAsState()
    val currentRoute = backStack?.destination?.route ?: Routes.Dashboard
    var settingsInitialTab by remember { mutableStateOf(SettingsTab.App) }

    LaunchedEffect(launchAgentId) {
        val agentId = launchAgentId?.takeIf { it.isNotBlank() } ?: return@LaunchedEffect
        navController.navigate(Routes.Dashboard) {
            popUpTo(navController.graph.findStartDestination().id) {
                saveState = true
            }
            launchSingleTop = true
            restoreState = false
        }
        onLaunchAgentHandled()
    }

    LaunchedEffect(openSettingsUpdate) {
        if (!openSettingsUpdate) return@LaunchedEffect
        settingsInitialTab = SettingsTab.App
        navController.navigate(Routes.Settings) {
            popUpTo(navController.graph.findStartDestination().id) {
                saveState = true
            }
            launchSingleTop = true
            restoreState = true
        }
        onOpenSettingsUpdateHandled()
    }

    val tabs = listOf(
        SysMonBottomTabs.dashboard,
        SysMonBottomTabs.hosts,
        SysMonBottomTabs.settings,
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
    ) {
        NavHost(
            navController = navController,
            startDestination = Routes.Dashboard,
            modifier = Modifier.fillMaxSize(),
        ) {
            composable(Routes.Dashboard) {
                DashboardScreen(repository = repository)
            }
            composable(Routes.Hosts) {
                HostsScreen(
                    repository = repository,
                    onOpenAlerts = { agentId, agentName ->
                        navController.navigate(Routes.hostAlerts(agentId, agentName))
                    },
                )
            }
            composable(
                route = Routes.HostAlerts,
                arguments = listOf(
                    navArgument("agentId") { type = NavType.StringType },
                    navArgument("agentName") { type = NavType.StringType },
                ),
            ) { entry ->
                val agentId = entry.arguments?.getString("agentId") ?: return@composable
                val agentName = Uri.decode(entry.arguments?.getString("agentName").orEmpty())
                    .ifBlank { agentId }
                HostAlertsScreen(
                    repository = repository,
                    agentId = agentId,
                    agentName = agentName,
                    onBack = { navController.popBackStack() },
                )
            }
            composable(Routes.Settings) {
                SettingsScreen(
                    repository = repository,
                    appPreferencesStore = appPreferencesStore,
                    initialTab = settingsInitialTab,
                    onLogout = onLogout,
                    onResetHub = onResetHub,
                )
            }
        }

        val showBottomBar = !currentRoute.startsWith("host_alerts")
        if (showBottomBar) {
            SysMonBottomBar(
                tabs = tabs,
                selectedRoute = currentRoute,
                onTabSelected = { route ->
                    navController.navigate(route) {
                        popUpTo(navController.graph.findStartDestination().id) {
                            saveState = true
                        }
                        launchSingleTop = true
                        restoreState = true
                    }
                },
                modifier = Modifier.align(Alignment.BottomCenter),
            )
        }
    }
}
