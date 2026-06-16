package com.nsi.app.ui.plan

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.models.Scheme
import com.nsi.app.theme.AppColors
import com.nsi.sdk.NSIClient

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PlanDetailScreen(navController: NavController, viewModel: AppViewModel, planId: String) {
    val uiState by viewModel.uiState.collectAsState()
    var schemes by remember { mutableStateOf<List<Scheme>>(emptyList()) }
    var currentIndex by remember { mutableIntStateOf(0) }
    var isUnlocked by remember { mutableStateOf(true) }

    LaunchedEffect(planId) {
        try {
            val client = NSIClient("http://127.0.0.1:39401", uiState.userInfo.nickName)
            isUnlocked = client.checkUnlock(planId).unlocked
        } catch (_: Exception) {}
        val cached = uiState.planResult
        if (cached != null) {
            schemes = cached.recommendedSchemes
        } else {
            try {
                val client = NSIClient("http://127.0.0.1:39401", uiState.userInfo.nickName)
                val plan = client.getPlanDetail(planId)
                schemes = (plan.recommendedSchemes ?: emptyList()).map { s ->
                    Scheme(
                        name = s.name ?: "",
                        baseSalary = s.baseSalary ?: 0,
                        monthlyCost = s.monthlyCost ?: 0.0,
                        projectedPension = s.projectedPension ?: 0.0,
                        annualSubsidy = s.annualSubsidy ?: 0.0,
                    )
                }
            } catch (_: Exception) {}
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("方案详情") })
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .background(AppColors.Background)
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
        ) {
            LazyRow(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                items(schemes.size) { index ->
                    val s = schemes[index]
                    val isActive = index == currentIndex
                    Card(
                        modifier = Modifier
                            .width(200.dp)
                            .clickable { currentIndex = index }
                            .then(if (isActive) Modifier.border(2.dp, AppColors.Primary, RoundedCornerShape(16.dp)) else Modifier),
                        shape = RoundedCornerShape(16.dp),
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(s.name, fontWeight = FontWeight.SemiBold, color = AppColors.Primary)
                            Spacer(modifier = Modifier.height(8.dp))
                            Text("预计养老金", fontSize = 13.sp, color = AppColors.TextSecondary)
                            Text("${s.projectedPension.toInt()}元/月", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary)
                            Text("月缴费: ${s.monthlyCost.toInt()}元", fontSize = 13.sp, color = AppColors.TextSecondary)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            if (schemes.isNotEmpty() && currentIndex < schemes.size) {
                val s = schemes[currentIndex]
                Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        detailRow("缴费基数", "${s.baseSalary}元")
                        detailRow("月缴费", "${s.monthlyCost.toInt()}元")
                        detailRow("企业年补贴", "${s.annualSubsidy.toInt()}元/年", highlight = true)
                        detailRow("预计养老金", "${s.projectedPension.toInt()}元/月", highlight = true)
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Surface(
                shape = RoundedCornerShape(12.dp),
                color = AppColors.BackgroundLight,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(
                    "详细行动清单和风险提示请在Web端查看",
                    modifier = Modifier.padding(12.dp),
                    fontSize = 13.sp,
                    color = AppColors.TextSecondary,
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Button(
                onClick = { navController.navigate(Routes.COMPLIANCE) },
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(48.dp),
            ) {
                Text("立即申请补贴")
            }

            Spacer(modifier = Modifier.height(12.dp))

            Button(
                onClick = {},
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(48.dp),
            ) {
                Text("保存为 PDF")
            }

            Spacer(modifier = Modifier.height(12.dp))

            OutlinedButton(
                onClick = {},
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(48.dp),
            ) {
                Text("分享方案")
            }
        }

        if (!isUnlocked) {
            AlertDialog(
                onDismissRequest = { navController.popBackStack() },
                title = { Text("未解锁") },
                text = { Text("请先解锁完整报告") },
                confirmButton = {
                    TextButton(onClick = { navController.popBackStack() }) { Text("返回") }
                },
            )
        }
    }
}

@Composable
private fun detailRow(label: String, value: String, highlight: Boolean = false) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, color = AppColors.TextSecondary)
        Text(value, fontWeight = FontWeight.SemiBold, color = if (highlight) AppColors.Highlight else AppColors.TextPrimary)
    }
    HorizontalDivider(thickness = 0.5.dp, color = AppColors.StepInactive)
}
