package com.nsi.app.ui.preview

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.AppColors

@Composable
fun PreviewScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()

    Column(
        modifier = Modifier.fillMaxSize().background(AppColors.Background).padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Spacer(modifier = Modifier.height(40.dp))
        Text("您的社保方案", fontSize = 28.sp, fontWeight = FontWeight.Bold)
        Text("以下是 AI 为您推荐的社保方案预览", fontSize = 16.sp, color = AppColors.TextSecondary)

        Spacer(modifier = Modifier.height(24.dp))

        Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
            Column(modifier = Modifier.padding(32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text("预计年补贴额", fontSize = 16.sp, color = AppColors.TextSecondary)
                Text("约 ¥XX,XXX/年", fontSize = 32.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary, modifier = Modifier.blur(12.dp))
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
            Column(modifier = Modifier.padding(32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text("月节省额", fontSize = 16.sp, color = AppColors.TextSecondary)
                Text("约 ¥XXX/月", fontSize = 32.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary, modifier = Modifier.blur(12.dp))
            }
        }

        Spacer(modifier = Modifier.height(40.dp))

        Button(
            onClick = {
                val planId = uiState.planResult?.planId
                if (planId != null) {
                    navController.navigate(Routes.plan(planId))
                }
            },
            modifier = Modifier.fillMaxWidth().height(48.dp),
            shape = RoundedCornerShape(48.dp),
            colors = ButtonDefaults.buttonColors(containerColor = AppColors.Primary),
        ) {
            Text("解锁完整报告", fontWeight = FontWeight.SemiBold)
        }

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            "解锁后可查看方案对比、年度现金流和行动清单",
            fontSize = 13.sp,
            color = AppColors.TextMuted,
            textAlign = TextAlign.Center,
        )
    }
}
