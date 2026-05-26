package com.nsi.app

import androidx.compose.runtime.Composable
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
    fun plan(planId: String) = "plan/$planId"
}

@Composable
fun NSIApp(viewModel: AppViewModel) {
    val navController = rememberNavController()
    NavHost(navController = navController, startDestination = Routes.HOME) {
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
    }
}
