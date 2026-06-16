package com.nsi.app.ui.compliance

import android.content.Context
import android.content.Intent
import android.net.Uri
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
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ComplianceScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var matchedPolicies by remember { mutableStateOf<List<JsonObject>>(emptyList()) }
    var requiredDocs by remember { mutableStateOf<List<JsonObject>>(emptyList()) }
    var eligibleTags by remember { mutableStateOf<List<String>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        scope.launch {
            try {
                val url = java.net.URI("http://127.0.0.1:39401/v1/compliance/checklist?city_code=${uiState.currentCityCode}").toURL()
                val conn = url.openConnection() as java.net.HttpURLConnection
                conn.setRequestProperty("x-user-id", uiState.userInfo.nickName)
                val json = Json { ignoreUnknownKeys = true }
                val response = json.parseToJsonElement(conn.inputStream.bufferedReader().readText()).jsonObject
                val data = response["data"]?.jsonObject
                if (data != null) {
                    matchedPolicies = data["matched_policies"]?.jsonArray?.map { it.jsonObject } ?: emptyList()
                    requiredDocs = data["required_docs"]?.jsonArray?.map { it.jsonObject } ?: emptyList()
                    eligibleTags = data["eligible_tags"]?.jsonArray?.map { it.jsonPrimitive.content } ?: emptyList()
                }
            } catch (_: Exception) {}
            isLoading = false
        }
    }

    Scaffold(
        topBar = { TopAppBar(title = { Text("合规认定") }) }
    ) { padding ->
        if (isLoading) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppColors.Primary)
            }
        } else {
            Column(
                modifier = Modifier
                    .padding(padding)
                    .fillMaxSize()
                    .background(AppColors.Background)
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
            ) {
                Text("合规认定与材料清单", fontSize = 24.sp, fontWeight = FontWeight.Bold)

                if (eligibleTags.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("您的可认定身份", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(8.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        eligibleTags.forEach { tag ->
                            Surface(
                                shape = RoundedCornerShape(16.dp),
                                color = AppColors.BackgroundLight,
                            ) {
                                Text(tag, modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp), fontSize = 14.sp, color = AppColors.Primary)
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))
                Text("匹配政策 (${matchedPolicies.size})", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)

                matchedPolicies.forEach { policy ->
                    Spacer(modifier = Modifier.height(12.dp))
                    val isEligible = policy["is_eligible"]?.jsonPrimitive?.content == "true"
                    Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text(
                                    policy["policy_type"]?.jsonPrimitive?.content ?: "",
                                    fontWeight = FontWeight.SemiBold,
                                    color = AppColors.Primary,
                                )
                                Spacer(modifier = Modifier.weight(1f))
                                Surface(
                                    shape = RoundedCornerShape(12.dp),
                                    color = if (isEligible) AppColors.Highlight.copy(alpha = 0.1f) else AppColors.Error.copy(alpha = 0.1f),
                                ) {
                                    Text(
                                        if (isEligible) "可申请" else "条件不足",
                                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp),
                                        fontSize = 12.sp,
                                        color = if (isEligible) AppColors.Highlight else AppColors.Error,
                                    )
                                }
                            }
                            Text(
                                policy["subsidy_calc_method"]?.jsonPrimitive?.content ?: "",
                                fontSize = 14.sp,
                                color = AppColors.TextSecondary,
                            )

                            val conditions = policy["conditions"]?.jsonArray
                            if (conditions != null && conditions.isNotEmpty()) {
                                Spacer(modifier = Modifier.height(8.dp))
                                Text("认定条件", fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
                                conditions.forEach { c ->
                                    val obj = c.jsonObject
                                    Text(obj["name"]?.jsonPrimitive?.content ?: "", fontSize = 13.sp)
                                    Text(obj["description"]?.jsonPrimitive?.content ?: "", fontSize = 12.sp, color = AppColors.TextSecondary)
                                }
                            }

                            val unmet = policy["unmet_conditions"]?.jsonArray
                            if (unmet != null && unmet.isNotEmpty()) {
                                Spacer(modifier = Modifier.height(8.dp))
                                Text("未满足条件", fontWeight = FontWeight.SemiBold, fontSize = 14.sp, color = AppColors.Error)
                                unmet.forEach { u ->
                                    Text("✗ ${u.jsonPrimitive.content}", fontSize = 13.sp, color = AppColors.Error)
                                }
                            }
                        }
                    }
                }

                if (requiredDocs.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("必需材料清单 (${requiredDocs.size})", fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
                    requiredDocs.forEach { doc ->
                        Spacer(modifier = Modifier.height(8.dp))
                        Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(12.dp)) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Row {
                                    Text(doc["name"]?.jsonPrimitive?.content ?: "", fontWeight = FontWeight.SemiBold)
                                    Spacer(modifier = Modifier.weight(1f))
                                    Text("来源: ${doc["source"]?.jsonPrimitive?.content ?: ""}", fontSize = 12.sp, color = AppColors.TextMuted)
                                }
                                Text(doc["description"]?.jsonPrimitive?.content ?: "", fontSize = 13.sp, color = AppColors.TextSecondary)
                                if (doc["optional"]?.jsonPrimitive?.content == "true") {
                                    Text("非必需材料", fontSize = 12.sp, color = androidx.compose.ui.graphics.Color(0xFFF59E0B))
                                }
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.height(24.dp))

                OutlinedButton(
                    onClick = {
                        val url = "http://127.0.0.1:39401/v1/guide?city_code=${uiState.currentCityCode}"
                        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                        context.startActivity(intent)
                    },
                    modifier = Modifier.fillMaxWidth().height(48.dp),
                    shape = RoundedCornerShape(48.dp),
                ) {
                    Text("查看办理指南")
                }
            }
        }
    }
}
