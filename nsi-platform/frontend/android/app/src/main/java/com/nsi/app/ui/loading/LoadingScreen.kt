package com.nsi.app.ui.loading

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.models.PlanResult
import com.nsi.app.models.Scheme
import com.nsi.app.theme.AppColors
import com.nsi.sdk.NSIClient
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.encodeToString

@Composable
fun LoadingScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    var progressText by remember { mutableStateOf("正在分析您的社保情况...") }
    val scope = rememberCoroutineScope()

    LaunchedEffect(Unit) {
        val input = uiState.planInput ?: return@LaunchedEffect
        progressText = "正在计算最优方案..."

        scope.launch {
            try {
                val client = NSIClient("http://127.0.0.1:39401", uiState.userInfo.nickName)
                val json = Json { ignoreUnknownKeys = true }
                val plan = client.generatePlan(json.encodeToString(input))
                viewModel.setPlanResult(PlanResult(
                    planId = plan.planId ?: "",
                    recommendedSchemes = (plan.recommendedSchemes ?: emptyList()).map { s ->
                        Scheme(
                            name = s.name ?: "",
                            baseSalary = s.baseSalary ?: 0,
                            monthlyCost = s.monthlyCost ?: 0.0,
                            projectedPension = s.projectedPension ?: 0.0,
                            annualSubsidy = s.annualSubsidy ?: 0.0,
                        )
                    }
                ))
                delay(800)
                navController.navigate(Routes.PREVIEW) {
                    popUpTo(Routes.LOADING) { inclusive = true }
                }
            } catch (_: Exception) {
                progressText = "生成失败，请重试"
                delay(2000)
            }
        }
    }

    Column(
        modifier = Modifier.fillMaxSize().background(AppColors.Background),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        CircularProgressIndicator(color = AppColors.Primary, strokeWidth = 6.dp, modifier = Modifier.size(64.dp))
        Spacer(modifier = Modifier.height(24.dp))
        Text("AI 正在为您匹配最优政策...", fontSize = 18.sp)
        Spacer(modifier = Modifier.height(16.dp))
        Text(progressText, fontSize = 14.sp, color = AppColors.TextSecondary)
    }
}
