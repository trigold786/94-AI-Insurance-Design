package com.nsi.app.ui.rights

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.AppColors
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RightsScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    val scope = rememberCoroutineScope()
    var summary by remember { mutableStateOf<JsonObject?>(null) }
    var alerts by remember { mutableStateOf<List<JsonObject>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        scope.launch {
            try {
                val baseUrl = "http://127.0.0.1:39401"
                val json = Json { ignoreUnknownKeys = true }

                val conn1 = java.net.URI("$baseUrl/v1/rights/payment-status").toURL().openConnection() as java.net.HttpURLConnection
                conn1.setRequestProperty("x-user-id", uiState.userInfo.nickName)
                val resp1 = json.parseToJsonElement(conn1.inputStream.bufferedReader().readText()).jsonObject
                summary = resp1["data"]?.jsonObject

                val conn2 = java.net.URI("$baseUrl/v1/rights/alerts?unread_only=false").toURL().openConnection() as java.net.HttpURLConnection
                conn2.setRequestProperty("x-user-id", uiState.userInfo.nickName)
                val resp2 = json.parseToJsonElement(conn2.inputStream.bufferedReader().readText()).jsonObject
                alerts = resp2["data"]?.jsonArray?.map { it.jsonObject } ?: emptyList()
            } catch (_: Exception) {}
            isLoading = false
        }
    }

    @Composable
    fun SummaryCard(value: Int, label: String, color: androidx.compose.ui.graphics.Color) {
        Column(
            modifier = Modifier.weight(1f).padding(4.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Surface(shape = RoundedCornerShape(16.dp), color = AppColors.White) {
                Column(modifier = Modifier.padding(16.dp).fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("$value", fontSize = 32.sp, fontWeight = FontWeight.Bold, color = color)
                    Text(label, fontSize = 14.sp, color = AppColors.TextSecondary)
                }
            }
        }
    }

    Scaffold(
        topBar = { TopAppBar(title = { Text("权益监测") }) }
    ) { padding ->
        if (isLoading) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppColors.Primary)
            }
        } else {
            Column(
                modifier = Modifier.padding(padding).fillMaxSize().background(AppColors.Background)
                    .verticalScroll(rememberScrollState()).padding(16.dp),
            ) {
                Row {
                    SummaryCard(summary?.get("paid_count")?.jsonPrimitive?.content?.toIntOrNull() ?: 0, "已缴纳", AppColors.Highlight)
                    SummaryCard(summary?.get("pending_count")?.jsonPrimitive?.content?.toIntOrNull() ?: 0, "待缴纳", androidx.compose.ui.graphics.Color(0xFFF59E0B))
                    SummaryCard(summary?.get("missed_count")?.jsonPrimitive?.content?.toIntOrNull() ?: 0, "已断缴", AppColors.Error)
                }

                Spacer(modifier = Modifier.height(16.dp))
                Text("预警通知 (${alerts.size})", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)

                if (alerts.isEmpty()) {
                    Text("暂无预警", fontSize = 14.sp, color = AppColors.TextMuted)
                }
                alerts.forEach { alert ->
                    Spacer(modifier = Modifier.height(8.dp))
                    val severity = alert["severity"]?.jsonPrimitive?.content ?: "low"
                    val severityText = when (severity) { "high" -> "高危"; "medium" -> "中等"; else -> "低危" }
                    val severityColor = when (severity) { "high" -> AppColors.Error; "medium" -> androidx.compose.ui.graphics.Color(0xFFF59E0B); else -> AppColors.PrimaryLight }

                    Surface(shape = RoundedCornerShape(12.dp), color = AppColors.White) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Surface(shape = RoundedCornerShape(8.dp), color = severityColor.copy(alpha = 0.1f)) {
                                    Text(severityText, modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp), fontSize = 11.sp, color = severityColor)
                                }
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(if (alert["alert_type"]?.jsonPrimitive?.content == "disconnection_risk") "断缴风险" else "政策变更", fontSize = 12.sp, color = AppColors.TextSecondary)
                                if (alert["is_read"]?.jsonPrimitive?.content == "true") {
                                    Spacer(modifier = Modifier.weight(1f))
                                    Text("已读", fontSize = 11.sp, color = AppColors.TextMuted)
                                }
                            }
                            Text(alert["title"]?.jsonPrimitive?.content ?: "", fontWeight = FontWeight.SemiBold, fontSize = 16.sp)
                            Text(alert["message"]?.jsonPrimitive?.content ?: "", fontSize = 13.sp, color = AppColors.TextSecondary)
                        }
                    }
                }

                val records = summary?.get("recent_records")?.jsonArray
                if (records != null && records.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("近期缴费记录 (${records.size})", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
                    records.forEach { r ->
                        val obj = r.jsonObject
                        val status = obj["status"]?.jsonPrimitive?.content ?: ""
                        val statusText = when (status) { "paid" -> "已缴"; "pending" -> "待缴"; else -> "断缴" }
                        val statusColor = when (status) { "paid" -> AppColors.Highlight; "pending" -> androidx.compose.ui.graphics.Color(0xFFF59E0B); else -> AppColors.Error }

                        Spacer(modifier = Modifier.height(6.dp))
                        Surface(shape = RoundedCornerShape(8.dp), color = AppColors.White) {
                            Row(
                                modifier = Modifier.padding(12.dp).fillMaxWidth(),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Text(obj["month"]?.jsonPrimitive?.content ?: "", fontSize = 14.sp, modifier = Modifier.width(70.dp))
                                Text(obj["policy_type"]?.jsonPrimitive?.content ?: "", fontSize = 13.sp, color = AppColors.TextSecondary, modifier = Modifier.weight(1f))
                                Text("${obj["amount"]?.jsonPrimitive?.content ?: "0"}元", fontSize = 13.sp)
                                Spacer(modifier = Modifier.width(8.dp))
                                Surface(shape = RoundedCornerShape(8.dp), color = statusColor.copy(alpha = 0.1f)) {
                                    Text(statusText, modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp), fontSize = 11.sp, color = statusColor)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
