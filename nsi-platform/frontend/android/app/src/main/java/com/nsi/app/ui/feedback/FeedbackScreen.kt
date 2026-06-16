package com.nsi.app.ui.feedback

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
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.AppColors
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

private const val API_BASE = "http://10.0.2.2:39401"

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
fun FeedbackScreen(navController: NavController, viewModel: AppViewModel) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val token = remember {
        context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
            .getString("auth_token", "") ?: ""
    }

    val categories = listOf("功能建议", "问题反馈", "政策咨询", "其他")
    var category by remember { mutableStateOf(categories[0]) }
    var expanded by remember { mutableStateOf(false) }
    var content by remember { mutableStateOf("") }
    var contact by remember { mutableStateOf("") }
    var submitting by remember { mutableStateOf(false) }
    var message by remember { mutableStateOf("") }

    fun submit() {
        if (content.isBlank()) {
            message = "请填写反馈内容"
            return
        }
        submitting = true
        scope.launch {
            try {
                val body = JSONObject().apply {
                    put("category", category)
                    put("content", content)
                    put("contact", contact)
                }.toString()
                httpPost("/v1/feedback", token, body)
                message = "提交成功，感谢您的反馈！"
                content = ""
                contact = ""
            } catch (e: Exception) {
                message = e.message ?: "提交失败"
            }
            submitting = false
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background)
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
    ) {
        Text("意见反馈", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = AppColors.TextPrimary)
        Spacer(modifier = Modifier.height(16.dp))

        Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = AppColors.White)) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("反馈类型", fontSize = 14.sp, color = AppColors.TextSecondary)
                Spacer(modifier = Modifier.height(8.dp))
                ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
                    OutlinedTextField(
                        value = category,
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("选择类型") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
                        modifier = Modifier.fillMaxWidth().menuAnchor(),
                    )
                    ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                        categories.forEach { c ->
                            DropdownMenuItem(
                                text = { Text(c) },
                                onClick = { category = c; expanded = false },
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                Text("反馈内容", fontSize = 14.sp, color = AppColors.TextSecondary)
                Spacer(modifier = Modifier.height(8.dp))
                OutlinedTextField(
                    value = content,
                    onValueChange = { content = it },
                    modifier = Modifier.fillMaxWidth().height(120.dp),
                    placeholder = { Text("请详细描述您的建议或问题...") },
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = AppColors.Primary),
                )

                Spacer(modifier = Modifier.height(16.dp))

                Text("联系方式（选填）", fontSize = 14.sp, color = AppColors.TextSecondary)
                Spacer(modifier = Modifier.height(8.dp))
                OutlinedTextField(
                    value = contact,
                    onValueChange = { contact = it },
                    modifier = Modifier.fillMaxWidth(),
                    placeholder = { Text("手机号/邮箱") },
                    singleLine = true,
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = AppColors.Primary),
                )
            }
        }

        if (message.isNotEmpty()) {
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                message,
                fontSize = 14.sp,
                color = if (message.contains("成功")) AppColors.Highlight else AppColors.Error,
            )
        }

        Spacer(modifier = Modifier.height(20.dp))

        Button(
            onClick = { submit() },
            enabled = !submitting,
            modifier = Modifier.fillMaxWidth().height(50.dp),
            shape = RoundedCornerShape(48.dp),
            colors = ButtonDefaults.buttonColors(containerColor = AppColors.Primary),
        ) {
            Text(if (submitting) "提交中..." else "提交反馈", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
        }
    }
}
