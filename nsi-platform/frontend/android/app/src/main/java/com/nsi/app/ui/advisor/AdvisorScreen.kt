package com.nsi.app.ui.advisor

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
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
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

private const val API_BASE = "http://10.0.2.2:39401"

data class ChatMessage(val role: String, val text: String, val time: String)

private suspend fun httpPost(path: String, token: String, body: String): JSONObject = withContext(Dispatchers.IO) {
    val conn = (URL("$API_BASE$path").openConnection() as HttpURLConnection).apply {
        requestMethod = "POST"
        connectTimeout = 15000
        readTimeout = 30000
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
fun AdvisorScreen(navController: NavController, viewModel: AppViewModel) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val token = remember {
        context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
            .getString("auth_token", "") ?: ""
    }
    val listState = rememberLazyListState()

    fun now(): String = SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date())

    var messages by remember {
        mutableStateOf(listOf(ChatMessage("ai", "您好，我是 AI 社保顾问，请问有什么可以帮您？", now())))
    }
    var input by remember { mutableStateOf("") }
    var sending by remember { mutableStateOf(false) }

    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) listState.animateScrollToItem(messages.size - 1)
    }

    fun send() {
        val q = input.trim()
        if (q.isEmpty() || sending) return
        input = ""
        messages = messages + ChatMessage("user", q, now())
        sending = true
        scope.launch {
            try {
                val body = JSONObject().apply {
                    put("question", q)
                    put("context", JSONObject())
                }.toString()
                val resp = httpPost("/v1/advisor/ask", token, body)
                val data = resp.optJSONObject("data") ?: resp
                val answer = data.optString("answer", data.optString("text", "暂无回复"))
                messages = messages + ChatMessage("ai", answer, now())
            } catch (e: Exception) {
                messages = messages + ChatMessage("ai", "出错了: ${e.message}", now())
            }
            sending = false
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background),
    ) {
        Text(
            "AI 顾问",
            fontSize = 20.sp,
            fontWeight = FontWeight.Bold,
            color = AppColors.TextPrimary,
            modifier = Modifier.padding(16.dp),
        )
        LazyColumn(
            modifier = Modifier.weight(1f).fillMaxWidth().padding(horizontal = 12.dp),
            state = listState,
            verticalArrangement = Arrangement.spacedBy(8.dp),
            contentPadding = PaddingValues(vertical = 8.dp),
        ) {
            items(messages.size) { i ->
                MessageBubble(messages[i])
            }
            if (sending) {
                item {
                    Row(modifier = Modifier.padding(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = AppColors.Primary,
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text("AI 正在思考...", fontSize = 14.sp, color = AppColors.TextSecondary)
                    }
                }
            }
        }
        Surface(color = AppColors.White, shadowElevation = 4.dp) {
            Row(
                modifier = Modifier.fillMaxWidth().padding(12.dp).imePadding(),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                OutlinedTextField(
                    value = input,
                    onValueChange = { input = it },
                    modifier = Modifier.weight(1f),
                    placeholder = { Text("请输入您的问题") },
                    shape = RoundedCornerShape(24.dp),
                    maxLines = 4,
                    colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = AppColors.Primary),
                )
                Spacer(modifier = Modifier.width(8.dp))
                Button(
                    onClick = { send() },
                    enabled = !sending && input.isNotBlank(),
                    modifier = Modifier.height(48.dp),
                    shape = RoundedCornerShape(24.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = AppColors.Primary),
                    contentPadding = PaddingValues(horizontal = 20.dp),
                ) {
                    Text("发送", color = AppColors.White, fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}

@Composable
private fun MessageBubble(msg: ChatMessage) {
    val isUser = msg.role == "user"
    val align = if (isUser) Alignment.End else Alignment.Start
    val bg = if (isUser) AppColors.Primary else AppColors.White
    val fg = if (isUser) AppColors.White else AppColors.TextPrimary
    Column(modifier = Modifier.fillMaxWidth(), horizontalAlignment = align) {
        Surface(
            shape = RoundedCornerShape(12.dp),
            color = bg,
            modifier = Modifier.widthIn(max = 280.dp),
        ) {
            Column(modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp)) {
                Text(msg.text, fontSize = 14.sp, color = fg)
                Text(
                    msg.time,
                    fontSize = 10.sp,
                    color = if (isUser) AppColors.White.copy(alpha = 0.7f) else AppColors.TextMuted,
                )
            }
        }
    }
}
