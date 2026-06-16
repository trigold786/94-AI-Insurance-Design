package com.nsi.app

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.nsi.app.models.AppViewModel
import com.nsi.app.ui.login.LoginScreen
import com.nsi.app.ui.home.HomeScreen
import com.nsi.app.ui.citypicker.CityPickerScreen
import com.nsi.app.ui.profile.ProfileScreen
import com.nsi.app.ui.loading.LoadingScreen
import com.nsi.app.ui.preview.PreviewScreen
import com.nsi.app.ui.plan.PlanDetailScreen
import com.nsi.app.ui.compliance.ComplianceScreen
import com.nsi.app.ui.rights.RightsScreen
import com.nsi.app.ui.settings.SettingsScreen
import com.nsi.app.ui.sandbox.SandboxScreen
import com.nsi.app.ui.advisor.AdvisorScreen
import com.nsi.app.ui.feedback.FeedbackScreen

object Routes {
    const val LOGIN = "login"
    const val HOME = "home"
    const val CITY_PICKER = "city_picker"
    const val PROFILE = "profile"
    const val LOADING = "loading"
    const val PREVIEW = "preview"
    const val PLAN = "plan/{planId}"
    const val COMPLIANCE = "compliance"
    const val RIGHTS = "rights"
    const val SETTINGS = "settings"
    const val SANDBOX = "sandbox"
    const val ADVISOR = "advisor"
    const val FEEDBACK = "feedback"
    fun plan(planId: String) = "plan/$planId"
}

@Composable
fun NSIApp(viewModel: AppViewModel) {
    val navController = rememberNavController()
    val context = LocalContext.current
    val prefs = context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
    val startDestination = if (prefs.getString("auth_token", null) != null) Routes.HOME else Routes.LOGIN
    NavHost(navController = navController, startDestination = startDestination) {
        composable(Routes.LOGIN) { LoginScreen(navController, viewModel) }
        composable(Routes.HOME) { HomeScreen(navController, viewModel) }
        composable(Routes.CITY_PICKER) { CityPickerScreen(navController, viewModel) }
        composable(Routes.PROFILE) { ProfileScreen(navController, viewModel) }
        composable(Routes.LOADING) { LoadingScreen(navController, viewModel) }
        composable(Routes.PREVIEW) { PreviewScreen(navController, viewModel) }
        composable(Routes.PLAN) { backStackEntry ->
            val planId = backStackEntry.arguments?.getString("planId") ?: ""
            PlanDetailScreen(navController, viewModel, planId)
        }
        composable(Routes.COMPLIANCE) {
            ComplianceScreen(navController, viewModel)
        }
        composable(Routes.RIGHTS) {
            RightsScreen(navController, viewModel)
        }
        composable(Routes.SETTINGS) {
            SettingsScreen(navController, viewModel)
        }
        composable(Routes.SANDBOX) {
            SandboxScreen(navController, viewModel)
        }
        composable(Routes.ADVISOR) {
            AdvisorScreen(navController, viewModel)
        }
        composable(Routes.FEEDBACK) {
            FeedbackScreen(navController, viewModel)
        }
    }
}
