package com.nsi.app.ui.login

import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.models.UserInfo
import com.nsi.app.theme.AppColors

@Composable
fun LoginScreen(navController: NavController, viewModel: AppViewModel) {
    var agreed by remember { mutableStateOf(false) }
    var loading by remember { mutableStateOf(false) }
    var showPrivacy by remember { mutableStateOf(false) }
    var showTerms by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(AppColors.BackgroundLight, AppColors.Background)
                )
            )
            .padding(horizontal = 40.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Spacer(modifier = Modifier.weight(1f))

        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Box(
                modifier = Modifier
                    .size(160.dp)
                    .clip(RoundedCornerShape(40.dp))
                    .background(
                        Brush.linearGradient(
                            colors = listOf(AppColors.Primary, AppColors.PrimaryLight)
                        )
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Text("社保", fontSize = 48.sp, fontWeight = FontWeight.Bold, color = AppColors.White)
            }
            Spacer(modifier = Modifier.height(16.dp))
            Text("AI社保智筹", fontSize = 36.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary)
            Spacer(modifier = Modifier.height(8.dp))
            Text("AI驱动的社保规划助手", fontSize = 18.sp, color = AppColors.TextSecondary)
        }

        Spacer(modifier = Modifier.height(60.dp))

        Surface(
            shape = RoundedCornerShape(16.dp),
            color = AppColors.White,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Row(
                modifier = Modifier.padding(16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Checkbox(checked = agreed, onCheckedChange = { agreed = it })
                Text("我已阅读并同意", fontSize = 14.sp, color = AppColors.TextSecondary)
                TextButton(onClick = { showPrivacy = true }) {
                    Text("《隐私政策》", fontSize = 14.sp, color = AppColors.Primary)
                }
                Text("和", fontSize = 14.sp, color = AppColors.TextSecondary)
                TextButton(onClick = { showTerms = true }) {
                    Text("《用户协议》", fontSize = 14.sp, color = AppColors.Primary)
                }
            }
        }

        Spacer(modifier = Modifier.height(32.dp))

        Button(
            onClick = {
                loading = true
                viewModel.setUserInfo(UserInfo())
                val ctx = LocalContext.current
                ctx.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
                    .edit().putString("auth_token", "session_token").apply()
                loading = false
                navController.navigate(Routes.HOME) {
                    popUpTo(Routes.LOGIN) { inclusive = true }
                }
            },
            enabled = agreed,
            modifier = Modifier.fillMaxWidth().height(52.dp),
            shape = RoundedCornerShape(48.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = if (agreed) AppColors.Primary else AppColors.TextMuted
            ),
        ) {
            Text(if (loading) "登录中..." else "微信一键登录", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
        }

        Spacer(modifier = Modifier.weight(1f))

        Text("登录即表示您同意上述条款", fontSize = 13.sp, color = AppColors.TextMuted)
        Spacer(modifier = Modifier.height(40.dp))
    }

    if (showPrivacy) {
        AlertDialog(
            onDismissRequest = { showPrivacy = false },
            title = { Text("隐私政策") },
            text = { Text("我们收集您的信息用于社保规划服务...") },
            confirmButton = { TextButton(onClick = { showPrivacy = false }) { Text("确定") } },
        )
    }
    if (showTerms) {
        AlertDialog(
            onDismissRequest = { showTerms = false },
            title = { Text("用户协议") },
            text = { Text("欢迎使用AI社保智筹...") },
            confirmButton = { TextButton(onClick = { showTerms = false }) { Text("确定") } },
        )
    }
}
