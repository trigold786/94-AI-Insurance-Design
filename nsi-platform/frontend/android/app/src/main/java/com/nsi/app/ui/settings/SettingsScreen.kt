package com.nsi.app.ui.settings

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.AppColors
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

private const val API_BASE = "http://10.0.2.2:39401"

private suspend fun httpGet(path: String, token: String): JSONObject = withContext(Dispatchers.IO) {
    val conn = (URL("$API_BASE$path").openConnection() as HttpURLConnection).apply {
        requestMethod = "GET"
        connectTimeout = 15000
        readTimeout = 15000
        setRequestProperty("Content-Type", "application/json")
        if (token.isNotEmpty()) setRequestProperty("Authorization", "Bearer $token")
    }
    try {
        val code = conn.responseCode
        val text = (if (code in 200..299) conn.inputStream else conn.errorStream)
            ?.bufferedReader()?.use { it.readText() } ?: ""
        if (code !in 200..299) throw RuntimeException("请求失败($code)")
        JSONObject(text)
    } finally {
        conn.disconnect()
    }
}

private suspend fun httpPost(path: String, token: String, body: String): JSONObject = withContext(Dispatchers.IO) {
    val conn = (URL("$API_BASE$path").openConnection() as HttpURLConnection).apply {
        requestMethod = "POST"
        connectTimeout = 15000
        readTimeout = 15000
        doOutput = true
        setRequestProperty("Content-Type", "application/json")
        if (token.isNotEmpty()) setRequestProperty("Authorization", "Bearer $token")
        outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
    }
    try {
        val code = conn.responseCode
        val text = (if (code in 200..299) conn.inputStream else conn.errorStream)
            ?.bufferedReader()?.use { it.readText() } ?: ""
        if (code !in 200..299) throw RuntimeException("请求失败($code)")
        JSONObject(text)
    } finally {
        conn.disconnect()
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(navController: NavController, viewModel: AppViewModel) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val token = remember {
        context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
            .getString("auth_token", "") ?: ""
    }

    var fontScale by remember { mutableStateOf("medium") }
    var defaultTab by remember { mutableStateOf("home") }
    var notifications by remember { mutableStateOf(true) }
    var loading by remember { mutableStateOf(true) }
    var saving by remember { mutableStateOf(false) }
    var message by remember { mutableStateOf("") }
    var tabExpanded by remember { mutableStateOf(false) }

    val fontOptions = listOf("small" to "小", "medium" to "标准", "large" to "大")
    val tabOptions = listOf(
        "home" to "首页", "profile" to "画像", "sandbox" to "沙盘",
        "compliance" to "合规", "rights" to "权益",
    )

    LaunchedEffect(Unit) {
        try {
            val resp = httpGet("/v1/settings", token)
            val data = resp.optJSONObject("data") ?: resp
            fontScale = data.optString("font_scale", "medium")
            defaultTab = data.optString("default_tab", "home")
            notifications = data.optBoolean("notifications", true)
        } catch (e: Exception) {
            message = e.message ?: "加载失败"
        }
        loading = false
    }

    fun save() {
        saving = true
        scope.launch {
            try {
                val body = JSONObject().apply {
                    put("font_scale", fontScale)
                    put("default_tab", defaultTab)
                    put("notifications", notifications)
                }.toString()
                httpPost("/v1/settings", token, body)
                message = "保存成功"
            } catch (e: Exception) {
                message = e.message ?: "保存失败"
            }
            saving = false
        }
    }

    fun logout() {
        context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE).edit().clear().apply()
        navController.navigate(Routes.LOGIN) {
            popUpTo(0) { inclusive = true }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background)
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
    ) {
        Text("设置", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = AppColors.TextPrimary)
        Spacer(modifier = Modifier.height(16.dp))

        if (loading) {
            Box(modifier = Modifier.fillMaxWidth().padding(32.dp), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppColors.Primary)
            }
        } else {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = AppColors.White),
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("字体大小", fontSize = 16.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
                    Spacer(modifier = Modifier.height(8.dp))
                    fontOptions.forEach { (value, label) ->
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            RadioButton(
                                selected = value == fontScale,
                                onClick = { fontScale = value },
                                colors = RadioButtonDefaults.colors(selectedColor = AppColors.Primary),
                            )
                            Text(label, fontSize = 15.sp, color = AppColors.TextPrimary)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = AppColors.White),
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("默认页签", fontSize = 16.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
                    Spacer(modifier = Modifier.height(8.dp))
                    ExposedDropdownMenuBox(expanded = tabExpanded, onExpandedChange = { tabExpanded = it }) {
                        OutlinedTextField(
                            value = tabOptions.find { it.first == defaultTab }?.second ?: "首页",
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("选择默认页签") },
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = tabExpanded) },
                            modifier = Modifier.fillMaxWidth().menuAnchor(),
                            colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = AppColors.Primary),
                        )
                        ExposedDropdownMenu(expanded = tabExpanded, onDismissRequest = { tabExpanded = false }) {
                            tabOptions.forEach { (value, label) ->
                                DropdownMenuItem(
                                    text = { Text(label) },
                                    onClick = { defaultTab = value; tabExpanded = false },
                                )
                            }
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = AppColors.White),
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        "消息通知",
                        fontSize = 16.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = AppColors.TextPrimary,
                        modifier = Modifier.weight(1f),
                    )
                    Switch(
                        checked = notifications,
                        onCheckedChange = { notifications = it },
                        colors = SwitchDefaults.colors(
                            checkedThumbColor = AppColors.White,
                            checkedTrackColor = AppColors.Primary,
                        ),
                    )
                }
            }

            if (message.isNotEmpty()) {
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    message,
                    fontSize = 14.sp,
                    color = if (message == "保存成功") AppColors.Highlight else AppColors.Error,
                )
            }

            Spacer(modifier = Modifier.height(20.dp))

            Button(
                onClick = { save() },
                enabled = !saving,
                modifier = Modifier.fillMaxWidth().height(50.dp),
                shape = RoundedCornerShape(48.dp),
                colors = ButtonDefaults.buttonColors(containerColor = AppColors.Primary),
            ) {
                Text(if (saving) "保存中..." else "保存设置", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
            }

            Spacer(modifier = Modifier.height(12.dp))

            OutlinedButton(
                onClick = { logout() },
                modifier = Modifier.fillMaxWidth().height(50.dp),
                shape = RoundedCornerShape(48.dp),
                colors = ButtonDefaults.outlinedButtonColors(contentColor = AppColors.Error),
            ) {
                Text("退出登录", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
            }
        }
    }
}
