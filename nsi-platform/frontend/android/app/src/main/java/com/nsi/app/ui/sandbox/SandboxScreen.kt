package com.nsi.app.ui.sandbox

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.withContext
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
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

data class PolicyTrigger(val title: String, val description: String, val severity: String)

data class SandboxResult(
    val monthlyCost: Double = 0.0,
    val pension: Double = 0.0,
    val subsidy: Double = 0.0,
    val net: Double = 0.0,
    val policyTriggers: List<PolicyTrigger> = emptyList(),
    val cashflow: List<Double> = emptyList(),
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SandboxScreen(navController: NavController, viewModel: AppViewModel) {
    val context = LocalContext.current
    val token = remember {
        context.getSharedPreferences("nsi_auth", Context.MODE_PRIVATE)
            .getString("auth_token", "") ?: ""
    }

    val cities = listOf(
        "北京" to "110000", "上海" to "310000", "广州" to "440100",
        "深圳" to "440300", "杭州" to "330100",
    )
    val employments = listOf(
        "employed" to "企业就业", "flexible" to "灵活就业",
        "self_employed" to "自雇", "unemployed" to "失业",
    )

    var city by remember { mutableStateOf("110000") }
    var cityLabel by remember { mutableStateOf("北京") }
    var cityExpanded by remember { mutableStateOf(false) }
    var gender by remember { mutableStateOf("male") }
    var age by remember { mutableStateOf(30) }
    var basePercent by remember { mutableStateOf(100) }
    var paidYears by remember { mutableStateOf(10) }
    var planYears by remember { mutableStateOf(15) }
    var employment by remember { mutableStateOf("flexible") }
    var empExpanded by remember { mutableStateOf(false) }
    var localHukou by remember { mutableStateOf(true) }

    var result by remember { mutableStateOf<SandboxResult?>(null) }
    var loading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf("") }

    LaunchedEffect(city, gender, age, basePercent, paidYears, planYears, employment, localHukou) {
        delay(150)
        loading = true
        error = ""
        try {
            val body = JSONObject().apply {
                put("city_code", city)
                put("gender", gender)
                put("age", age)
                put("base_percent", basePercent)
                put("paid_years", paidYears)
                put("plan_years", planYears)
                put("employment_status", employment)
                put("local_hukou", localHukou)
            }.toString()
            val resp = httpPost("/v1/simulator/calculate", token, body)
            val data = resp.optJSONObject("data") ?: resp

            val triggers = mutableListOf<PolicyTrigger>()
            data.optJSONArray("policy_triggers")?.let { arr ->
                for (i in 0 until arr.length()) {
                    val t = arr.getJSONObject(i)
                    triggers.add(
                        PolicyTrigger(
                            title = t.optString("title"),
                            description = t.optString("description"),
                            severity = t.optString("severity", "info"),
                        ),
                    )
                }
            }

            val cashflow = mutableListOf<Double>()
            data.optJSONArray("cashflow")?.let { arr ->
                for (i in 0 until arr.length()) {
                    cashflow.add(arr.optDouble(i))
                }
            }

            result = SandboxResult(
                monthlyCost = data.optDouble("monthly_cost"),
                pension = data.optDouble("pension"),
                subsidy = data.optDouble("subsidy"),
                net = data.optDouble("net"),
                policyTriggers = triggers,
                cashflow = cashflow,
            )
        } catch (e: Exception) {
            error = e.message ?: "计算失败"
        }
        loading = false
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background)
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
    ) {
        Text("沙盘模拟", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = AppColors.TextPrimary)
        Spacer(modifier = Modifier.height(16.dp))

        Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = AppColors.White)) {
            Column(modifier = Modifier.padding(16.dp)) {
                ExposedDropdownMenuBox(expanded = cityExpanded, onExpandedChange = { cityExpanded = it }) {
                    OutlinedTextField(
                        value = cityLabel,
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("城市") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = cityExpanded) },
                        modifier = Modifier.fillMaxWidth().menuAnchor(),
                    )
                    ExposedDropdownMenu(expanded = cityExpanded, onDismissRequest = { cityExpanded = false }) {
                        cities.forEach { (label, code) ->
                            DropdownMenuItem(
                                text = { Text(label) },
                                onClick = { city = code; cityLabel = label; cityExpanded = false },
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                Text("性别", fontSize = 14.sp, color = AppColors.TextSecondary)
                Row {
                    listOf("male" to "男", "female" to "女").forEach { (v, l) ->
                        FilterChip(
                            selected = gender == v,
                            onClick = { gender = v },
                            label = { Text(l) },
                            colors = FilterChipDefaults.filterChipColors(selectedContainerColor = AppColors.Primary),
                            modifier = Modifier.padding(end = 8.dp),
                        )
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                Text("年龄: $age", fontSize = 14.sp, color = AppColors.TextSecondary)
                Slider(
                    value = age.toFloat(),
                    onValueChange = { age = it.toInt() },
                    valueRange = 16f..70f,
                    steps = 53,
                    colors = SliderDefaults.colors(thumbColor = AppColors.Primary, activeTrackColor = AppColors.Primary),
                )

                Spacer(modifier = Modifier.height(8.dp))

                Text("缴费基数: ${basePercent}%", fontSize = 14.sp, color = AppColors.TextSecondary)
                Slider(
                    value = basePercent.toFloat(),
                    onValueChange = { basePercent = it.toInt() },
                    valueRange = 60f..300f,
                    steps = 23,
                    colors = SliderDefaults.colors(thumbColor = AppColors.Primary, activeTrackColor = AppColors.Primary),
                )

                Spacer(modifier = Modifier.height(8.dp))

                Text("已缴年限: $paidYears 年", fontSize = 14.sp, color = AppColors.TextSecondary)
                Slider(
                    value = paidYears.toFloat(),
                    onValueChange = { paidYears = it.toInt() },
                    valueRange = 0f..35f,
                    steps = 34,
                    colors = SliderDefaults.colors(thumbColor = AppColors.Primary, activeTrackColor = AppColors.Primary),
                )

                Spacer(modifier = Modifier.height(8.dp))

                Text("计划年限: $planYears 年", fontSize = 14.sp, color = AppColors.TextSecondary)
                Slider(
                    value = planYears.toFloat(),
                    onValueChange = { planYears = it.toInt() },
                    valueRange = 0f..35f,
                    steps = 34,
                    colors = SliderDefaults.colors(thumbColor = AppColors.Primary, activeTrackColor = AppColors.Primary),
                )

                Spacer(modifier = Modifier.height(12.dp))

                ExposedDropdownMenuBox(expanded = empExpanded, onExpandedChange = { empExpanded = it }) {
                    OutlinedTextField(
                        value = employments.find { it.first == employment }?.second ?: "灵活就业",
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("就业状态") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = empExpanded) },
                        modifier = Modifier.fillMaxWidth().menuAnchor(),
                    )
                    ExposedDropdownMenu(expanded = empExpanded, onDismissRequest = { empExpanded = false }) {
                        employments.forEach { (v, l) ->
                            DropdownMenuItem(
                                text = { Text(l) },
                                onClick = { employment = v; empExpanded = false },
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("本地户籍", modifier = Modifier.weight(1f), fontSize = 14.sp, color = AppColors.TextSecondary)
                    Switch(
                        checked = localHukou,
                        onCheckedChange = { localHukou = it },
                        colors = SwitchDefaults.colors(checkedTrackColor = AppColors.Primary),
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        if (loading && result == null) {
            Box(modifier = Modifier.fillMaxWidth().padding(32.dp), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppColors.Primary)
            }
        }

        if (error.isNotEmpty()) {
            Text(error, color = AppColors.Error, fontSize = 14.sp)
        }

        result?.let { r ->
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                OutputCard("月缴", "${r.monthlyCost.toInt()}元", AppColors.Primary, Modifier.weight(1f))
                OutputCard("养老金", "${r.pension.toInt()}元", AppColors.Highlight, Modifier.weight(1f))
                OutputCard("补贴", "${r.subsidy.toInt()}元", Color(0xFFF59E0B), Modifier.weight(1f))
                OutputCard("净支出", "${r.net.toInt()}元", AppColors.Error, Modifier.weight(1f))
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (r.cashflow.isNotEmpty()) {
                Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = AppColors.White)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("现金流预测", fontSize = 16.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
                        Spacer(modifier = Modifier.height(12.dp))
                        CashflowChart(r.cashflow, AppColors.Primary)
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (r.policyTriggers.isNotEmpty()) {
                Text("政策触发", fontSize = 16.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
                Spacer(modifier = Modifier.height(8.dp))
                r.policyTriggers.forEach { t ->
                    val color = when (t.severity) {
                        "high" -> AppColors.Error
                        "medium" -> Color(0xFFF59E0B)
                        else -> AppColors.Highlight
                    }
                    Card(
                        shape = RoundedCornerShape(12.dp),
                        colors = CardDefaults.cardColors(containerColor = color.copy(alpha = 0.1f)),
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                    ) {
                        Column(modifier = Modifier.padding(12.dp)) {
                            Text(t.title, fontWeight = FontWeight.SemiBold, color = color, fontSize = 14.sp)
                            if (t.description.isNotEmpty()) {
                                Text(t.description, fontSize = 12.sp, color = AppColors.TextSecondary)
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun OutputCard(label: String, value: String, color: Color, modifier: Modifier = Modifier) {
    Card(
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = AppColors.White),
        modifier = modifier,
    ) {
        Column(modifier = Modifier.padding(12.dp), horizontalAlignment = Alignment.CenterHorizontally) {
            Text(label, fontSize = 12.sp, color = AppColors.TextSecondary)
            Spacer(modifier = Modifier.height(4.dp))
            Text(value, fontSize = 16.sp, fontWeight = FontWeight.Bold, color = color)
        }
    }
}

@Composable
private fun CashflowChart(data: List<Double>, barColor: Color) {
    Canvas(modifier = Modifier.fillMaxWidth().height(140.dp)) {
        if (data.isEmpty()) return@Canvas
        val maxVal = (data.maxOrNull() ?: 1.0).coerceAtLeast(1.0)
        val barSpace = size.width / data.size
        val barWidth = barSpace * 0.6f
        data.forEachIndexed { i, v ->
            val ratio = (v / maxVal).toFloat().coerceIn(0f, 1f)
            val barHeight = size.height * ratio
            drawRect(
                color = barColor,
                topLeft = Offset(
                    x = i * barSpace + (barSpace - barWidth) / 2f,
                    y = size.height - barHeight,
                ),
                size = Size(barWidth, barHeight),
            )
        }
    }
}
