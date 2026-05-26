package com.nsi.app.ui.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.models.PlanInput
import com.nsi.app.models.ProfileInput
import com.nsi.app.theme.AppColors
import kotlinx.coroutines.launch
import com.nsi.sdk.NSIClient
import kotlinx.serialization.json.Json
import kotlinx.serialization.encodeToString

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    val scope = rememberCoroutineScope()

    var currentStep by remember { mutableIntStateOf(1) }
    var age by remember { mutableStateOf("") }
    var gender by remember { mutableStateOf("male") }
    var household by remember { mutableStateOf("") }
    var employment by remember { mutableStateOf("flexible") }
    var years by remember { mutableStateOf("") }
    var hasChildren by remember { mutableStateOf(false) }
    var budget by remember { mutableStateOf("") }
    var balance by remember { mutableStateOf("") }
    var submitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf("") }

    val genders = listOf("male" to "男", "female" to "女")
    val employmentStatuses = listOf(
        "employed" to "企业就业", "flexible" to "灵活就业",
        "self_employed" to "自雇", "unemployed" to "失业",
    )

    @Composable
    fun StepCircle(number: Int, active: Boolean) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .clip(CircleShape)
                .background(if (active) AppColors.Primary else AppColors.StepInactive),
            contentAlignment = Alignment.Center,
        ) {
            Text("$number", color = if (active) AppColors.White else AppColors.TextMuted, fontWeight = FontWeight.SemiBold)
        }
    }

    @Composable
    fun StepLine(active: Boolean) {
        Box(
            modifier = Modifier
                .width(40.dp)
                .height(3.dp)
                .background(if (active) AppColors.Primary else AppColors.StepInactive)
        )
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background)
            .padding(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.Center,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            StepCircle(1, currentStep >= 1)
            StepLine(currentStep >= 2)
            StepCircle(2, currentStep >= 2)
            StepLine(currentStep >= 3)
            StepCircle(3, currentStep >= 3)
        }

        Spacer(modifier = Modifier.height(24.dp))

        when (currentStep) {
            1 -> {
                Text("基本信息", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("性别")
                    Spacer(modifier = Modifier.width(16.dp))
                    genders.forEach { (value, label) ->
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            RadioButton(selected = value == gender, onClick = { gender = value })
                            Text(label)
                        }
                    }
                }
                OutlinedTextField(value = age, onValueChange = { age = it }, label = { Text("年龄(16-70)") }, singleLine = true)
                OutlinedTextField(value = household, onValueChange = { household = it }, label = { Text("户籍地") }, singleLine = true)
                Spacer(modifier = Modifier.height(20.dp))
                Button(onClick = { currentStep = 2 }, modifier = Modifier.fillMaxWidth().height(48.dp), shape = RoundedCornerShape(48.dp)) {
                    Text("下一步")
                }
            }
            2 -> {
                Text("就业信息", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                var expanded by remember { mutableStateOf(false) }
                ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
                    OutlinedTextField(
                        value = employmentStatuses.find { it.first == employment }?.second ?: "请选择",
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("就业状态") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
                        modifier = Modifier.menuAnchor(),
                    )
                    ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                        employmentStatuses.forEach { (value, label) ->
                            DropdownMenuItem(
                                text = { Text(label) },
                                onClick = { employment = value; expanded = false },
                            )
                        }
                    }
                }
                OutlinedTextField(value = years, onValueChange = { years = it }, label = { Text("累计社保年限(年)") }, singleLine = true)
                Spacer(modifier = Modifier.height(20.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    OutlinedButton(onClick = { currentStep = 1 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("上一步")
                    }
                    Button(onClick = { currentStep = 3 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("下一步")
                    }
                }
            }
            3 -> {
                Text("补充信息", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("子女情况")
                    Spacer(modifier = Modifier.width(16.dp))
                    Switch(checked = hasChildren, onCheckedChange = { hasChildren = it })
                }
                OutlinedTextField(value = budget, onValueChange = { budget = it }, label = { Text("月预算（元）") }, singleLine = true)
                OutlinedTextField(value = balance, onValueChange = { balance = it }, label = { Text("当前账户余额（元）") }, singleLine = true)
                if (errorMessage.isNotEmpty()) {
                    Text(errorMessage, color = AppColors.Error, fontSize = 14.sp)
                }
                Spacer(modifier = Modifier.height(20.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    OutlinedButton(onClick = { currentStep = 2 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("上一步")
                    }
                    Button(
                        onClick = {
                            val ageInt = age.toIntOrNull()
                            if (ageInt == null || ageInt !in 16..70) {
                                errorMessage = "年龄必须在16-70之间"
                                return@Button
                            }
                            val budgetDouble = budget.toDoubleOrNull()
                            if (budgetDouble == null || budgetDouble <= 0) {
                                errorMessage = "请输入有效月预算"
                                return@Button
                            }
                            errorMessage = ""
                            submitting = true

                            val profileData = ProfileInput(
                                age = ageInt,
                                gender = gender,
                                householdRegionCode = uiState.currentCityCode,
                                currentResidenceCode = uiState.currentCityCode,
                                employmentStatus = employment,
                                socialSecurityYears = years.toIntOrNull() ?: 0,
                                hasChildren = hasChildren,
                            )
                            viewModel.setProfileData(profileData)
                            viewModel.setPlanInput(PlanInput(
                                age = ageInt,
                                gender = gender,
                                employment = employment,
                                contributionYears = years.toIntOrNull() ?: 0,
                                currentBalance = balance.toDoubleOrNull() ?: 0.0,
                                monthlyBudget = budgetDouble,
                                localAvgSalary = budgetDouble * 2,
                            ))

                            scope.launch {
                                try {
                                    val client = NSIClient("http://127.0.0.1:39401", uiState.userInfo.nickName)
                                    val json = Json { ignoreUnknownKeys = true }
                                    client.updateProfile(json.encodeToString(profileData))
                                    navController.navigate(Routes.LOADING) {
                                        popUpTo(Routes.PROFILE) { inclusive = true }
                                    }
                                } catch (e: Exception) {
                                    errorMessage = e.message ?: "提交失败"
                                    submitting = false
                                }
                            }
                        },
                        enabled = !submitting,
                        modifier = Modifier.weight(1f).height(48.dp),
                        shape = RoundedCornerShape(48.dp),
                    ) {
                        Text(if (submitting) "提交中..." else "生成方案")
                    }
                }
            }
        }
    }
}
