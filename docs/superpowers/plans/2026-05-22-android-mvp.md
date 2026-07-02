# Android Jetpack Compose MVP ÂÆûÁé∞ËÆ°Âàí

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build complete Android Jetpack Compose MVP with 7 screens (P1-P7) + ViewModel + API integration

**Architecture:** Jetpack Compose with Navigation Compose, MVVM (ViewModel + StateFlow), NSI SDK for backend calls. Navigation flows login ‚Ü?home ‚Ü?city-picker ‚Ü?profile ‚Ü?loading ‚Ü?preview ‚Ü?plan. Data flows via shared ViewModel.

**Tech Stack:** Kotlin 1.9+, Jetpack Compose, Navigation Compose, kotlinx.serialization, java.net.http.HttpClient

---

### Task 1: Project Scaffold + Models + ViewModel

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/MainActivity.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/NSIApp.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/models/AppState.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/theme/Color.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/theme/Theme.kt`
- Create: `frontend/android/app/src/main/res/values/strings.xml`
- Create: `frontend/android/app/src/main/AndroidManifest.xml`

- [ ] **Step 1: Create directories**

```bash
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/models" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/login" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/home" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/citypicker" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/profile" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/loading" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/preview" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/ui/plan" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/java/com/nsi/app/theme" -Force
New-Item -ItemType Directory -Path "nsi-platform/frontend/android/app/src/main/res/values" -Force
```

- [ ] **Step 2: Write AndroidManifest.xml**

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.nsi.app">
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
    <uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
    <application
        android:allowBackup="true"
        android:label="@string/app_name"
        android:supportsRtl="true"
        android:theme="@style/Theme.Material3.DayNight.NoActionBar"
        android:usesCleartextTraffic="true">
        <activity
            android:name=".MainActivity"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
```

- [ ] **Step 3: Write strings.xml**

```xml
<resources>
    <string name="app_name">AIÁ§æ‰øùÊô∫Á≠π</string>
</resources>
```

- [ ] **Step 4: Write Color.kt**

```kotlin
package com.nsi.app.theme

import androidx.compose.ui.graphics.Color

object AppColors {
    val Primary = Color(0xFF1A56DB)
    val PrimaryLight = Color(0xFF3B82F6)
    val Background = Color(0xFFF5F7FA)
    val BackgroundLight = Color(0xFFEEF2FF)
    val White = Color(0xFFFFFFFF)
    val TextPrimary = Color(0xFF333333)
    val TextSecondary = Color(0xFF6B7280)
    val TextMuted = Color(0xFF9CA3AF)
    val Error = Color(0xFFEF4444)
    val Highlight = Color(0xFF059669)
    val Border = Color(0xFFD1D5DB)
    val StepInactive = Color(0xFFE5E7EB)
}
```

- [ ] **Step 5: Write Theme.kt**

```kotlin
package com.nsi.app.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

private val LightColorScheme = lightColorScheme(
    primary = AppColors.Primary,
    onPrimary = AppColors.White,
    background = AppColors.Background,
    surface = AppColors.White,
    error = AppColors.Error,
)

@Composable
fun NSITheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = LightColorScheme,
        content = content,
    )
}
```

- [ ] **Step 6: Write AppState.kt (ViewModel)**

```kotlin
package com.nsi.app.models

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.serialization.Serializable

data class UserInfo(val nickName: String = "default")

@Serializable
data class ProfileInput(
    val age: Int,
    val gender: String,
    val householdRegionCode: String,
    val currentResidenceCode: String,
    val employmentStatus: String,
    val socialSecurityYears: Int,
    val hasChildren: Boolean,
)

@Serializable
data class PlanInput(
    val age: Int,
    val gender: String,
    val employment: String,
    val contributionYears: Int,
    val currentBalance: Double,
    val monthlyBudget: Double,
    val localAvgSalary: Double,
)

data class Scheme(
    val name: String,
    val baseSalary: Int,
    val monthlyCost: Double,
    val projectedPension: Double,
    val annualSubsidy: Double,
)

data class PlanResult(
    val planId: String,
    val recommendedSchemes: List<Scheme>,
)

data class AppUiState(
    val userInfo: UserInfo = UserInfo(),
    val currentCity: String = "‰∏äÊµ∑",
    val currentCityCode: String = "310000",
    val profileData: ProfileInput? = null,
    val planInput: PlanInput? = null,
    val planResult: PlanResult? = null,
)

class AppViewModel : ViewModel() {
    private val _uiState = MutableStateFlow(AppUiState())
    val uiState: StateFlow<AppUiState> = _uiState.asStateFlow()

    fun setUserInfo(info: UserInfo) {
        _uiState.value = _uiState.value.copy(userInfo = info)
    }

    fun setCity(name: String, code: String) {
        _uiState.value = _uiState.value.copy(currentCity = name, currentCityCode = code)
    }

    fun setProfileData(data: ProfileInput) {
        _uiState.value = _uiState.value.copy(profileData = data)
    }

    fun setPlanInput(input: PlanInput) {
        _uiState.value = _uiState.value.copy(planInput = input)
    }

    fun setPlanResult(result: PlanResult) {
        _uiState.value = _uiState.value.copy(planResult = result)
    }
}
```

- [ ] **Step 7: Write NSIApp.kt (NavHost + navigation)**

```kotlin
package com.nsi.app

import androidx.compose.runtime.Composable
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.nsi.app.models.AppViewModel
import com.nsi.app.ui.login.LoginScreen
import com.nsi.app.ui.home.HomeScreen
import com.nsi.app.ui.citypicker.CityPickerScreen
import com.nsi.app.ui.profile.ProfileScreen
import com.nsi.app.ui.loading.LoadingScreen
import com.nsi.app.ui.preview.PreviewScreen
import com.nsi.app.ui.plan.PlanDetailScreen

object Routes {
    const val LOGIN = "login"
    const val HOME = "home"
    const val CITY_PICKER = "city_picker"
    const val PROFILE = "profile"
    const val LOADING = "loading"
    const val PREVIEW = "preview"
    const val PLAN = "plan/{planId}"
    fun plan(planId: String) = "plan/$planId"
}

@Composable
fun NSIApp(viewModel: AppViewModel) {
    val navController = rememberNavController()
    NavHost(navController = navController, startDestination = Routes.LOGIN) {
        composable(Routes.LOGIN) { LoginScreen(navController, viewModel) }
        composable(Routes.HOME) { HomeScreen(navController, viewModel) }
        composable(Routes.CITY_PICKER) { CityPickerScreen(navController, viewModel) }
        composable(Routes.PROFILE) { ProfileScreen(navController, viewModel) }
        composable(Routes.LOADING) { LoadingScreen(navController, viewModel) }
        composable(Routes.PREVIEW) { PreviewScreen(navController, viewModel) }
        composable(Routes.PLAN) { backStackEntry ->
            val planId = backStackEntry.arguments?.getString("planId") ?: ""
            PlanDetailScreen(navController, viewModel, planId)
        }
    }
}
```

- [ ] **Step 8: Write MainActivity.kt**

```kotlin
package com.nsi.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.lifecycle.viewmodel.compose.viewModel
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.NSITheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            NSITheme {
                val viewModel: AppViewModel = viewModel()
                NSIApp(viewModel)
            }
        }
    }
}
```

---

### Task 2: P1 Login Screen

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/login/LoginScreen.kt`

```kotlin
package com.nsi.app.ui.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
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

        // Logo + Title
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
                Text("Á§æ‰øù", fontSize = 48.sp, fontWeight = FontWeight.Bold, color = AppColors.White)
            }
            Spacer(modifier = Modifier.height(16.dp))
            Text("AIÁ§æ‰øùÊô∫Á≠π", fontSize = 36.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary)
            Spacer(modifier = Modifier.height(8.dp))
            Text("AIÈ©±Âä®ÁöÑÁ§æ‰øùËßÑÂàíÂä©Êâ?, fontSize = 18.sp, color = AppColors.TextSecondary)
        }

        Spacer(modifier = Modifier.height(60.dp))

        // Agreement
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
                Text("ÊàëÂ∑≤ÈòÖËØªÂπ∂ÂêåÊÑ?, fontSize = 14.sp, color = AppColors.TextSecondary)
                TextButton(onClick = { showPrivacy = true }) {
                    Text("„ÄäÈöêÁßÅÊîøÁ≠ñ„Ä?, fontSize = 14.sp, color = AppColors.Primary)
                }
                Text("Âí?, fontSize = 14.sp, color = AppColors.TextSecondary)
                TextButton(onClick = { showTerms = true }) {
                    Text("„ÄäÁî®Êà∑ÂçèËÆÆ„Ä?, fontSize = 14.sp, color = AppColors.Primary)
                }
            }
        }

        Spacer(modifier = Modifier.height(32.dp))

        // Login button
        Button(
            onClick = {
                loading = true
                viewModel.setUserInfo(UserInfo())
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
            Text(if (loading) "ÁôªÂΩï‰∏?.." else "ÂæÆ‰ø°‰∏ÄÈîÆÁôªÂΩ?, fontSize = 18.sp, fontWeight = FontWeight.SemiBold)
        }

        Spacer(modifier = Modifier.weight(1f))

        Text("ÁôªÂΩïÂç≥Ë°®Á§∫ÊÇ®ÂêåÊÑè‰∏äËø∞Êù°Ê¨æ", fontSize = 13.sp, color = AppColors.TextMuted)
        Spacer(modifier = Modifier.height(40.dp))
    }

    if (showPrivacy) {
        AlertDialog(
            onDismissRequest = { showPrivacy = false },
            title = { Text("ÈöêÁßÅÊîøÁ≠ñ") },
            text = { Text("Êàë‰ª¨Êî∂ÈõÜÊÇ®ÁöÑ‰ø°ÊÅØÁî®‰∫éÁ§æ‰øùËßÑÂàíÊúçÂä°...") },
            confirmButton = { TextButton(onClick = { showPrivacy = false }) { Text("Á°ÆÂÆö") } },
        )
    }
    if (showTerms) {
        AlertDialog(
            onDismissRequest = { showTerms = false },
            title = { Text("Áî®Êà∑ÂçèËÆÆ") },
            text = { Text("Ê¨¢Ëøé‰ΩøÁî®AIÁ§æ‰øùÊô∫Á≠π...") },
            confirmButton = { TextButton(onClick = { showTerms = false }) { Text("Á°ÆÂÆö") } },
        )
    }
}
```

---

### Task 3: P2 Home Screen + P3 City Picker

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/home/HomeScreen.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/citypicker/CityPickerScreen.kt`

- [ ] **Step 1: Write HomeScreen.kt**

```kotlin
package com.nsi.app.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.theme.AppColors
import kotlinx.coroutines.launch
import com.nsi.sdk.NSIClient
import kotlinx.serialization.Serializable

@Serializable
data class PolicyClaim(
    val claimId: String? = null,
    val policyType: String? = null,
    val subsidyCalcMethod: String? = null,
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    var policies by remember { mutableStateOf<List<PolicyClaim>>(emptyList()) }
    val scope = rememberCoroutineScope()

    LaunchedEffect(uiState.currentCityCode) {
        scope.launch {
            try {
                val client = NSIClient("http://localhost:30001", uiState.userInfo.nickName)
                val claims = client.queryPolicies(
                    regionCode = uiState.currentCityCode
                )
                policies = claims.map { c ->
                    PolicyClaim(claimId = c.claimId, policyType = c.policyId)
                }
            } catch (_: Exception) {}
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.Background)
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
    ) {
        // City bar
        TextButton(onClick = { navController.navigate(Routes.CITY_PICKER) }) {
            Text("üìç", fontSize = 18.sp)
            Spacer(modifier = Modifier.width(4.dp))
            Text(uiState.currentCity, fontSize = 18.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
            Text(" ‚ñ?, fontSize = 14.sp, color = AppColors.TextMuted)
        }

        // Hero
        Card(
            modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = AppColors.Primary),
        ) {
            Column(
                modifier = Modifier.padding(40.dp).fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text("Á§æ‰øùËßÑÂàíÔºåÊô∫ËÉΩÂåπÈÖ?, fontSize = 26.sp, fontWeight = FontWeight.Bold, color = AppColors.White)
                Spacer(modifier = Modifier.height(12.dp))
                Text("AI ‰∏∫ÊÇ®ÂåπÈÖçÊúÄ‰ºòÁ§æ‰øùÊñπÊ°?, fontSize = 18.sp, color = AppColors.White.copy(alpha = 0.8f))
                Spacer(modifier = Modifier.height(24.dp))
                Button(
                    onClick = { navController.navigate(Routes.PROFILE) },
                    modifier = Modifier.fillMaxWidth().height(48.dp),
                    shape = RoundedCornerShape(48.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = AppColors.White.copy(alpha = 0.2f)),
                ) {
                    Text("ÂºÄÂßãÁ§æ‰øùÁ≠πÂà?, fontWeight = FontWeight.SemiBold)
                }
            }
        }

        // Policy scroll
        Text("ÊîøÁ≠ñÈÄüËßà", fontSize = 20.sp, fontWeight = FontWeight.SemiBold)

        LazyRow(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            items(policies.size) { index ->
                val p = policies[index]
                Card(
                    modifier = Modifier.width(200.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(containerColor = AppColors.White),
                    elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(p.policyType ?: "-", fontWeight = FontWeight.SemiBold, color = AppColors.Primary)
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(p.subsidyCalcMethod ?: "-", fontSize = 14.sp, color = AppColors.TextSecondary)
                    }
                }
            }
        }
    }
}
```

- [ ] **Step 2: Write CityPickerScreen.kt**

```kotlin
package com.nsi.app.ui.citypicker

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
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

private val cities = listOf(
    "310000" to "‰∏äÊµ∑", "110000" to "Âåó‰∫¨",
    "440300" to "Ê∑±Âú≥", "440100" to "ÂπøÂ∑û",
    "330100" to "Êù≠Â∑û",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CityPickerScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("ÈÄâÊã©ÂüéÂ∏Ç") })
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding).fillMaxSize()) {
            cities.forEach { (code, name) ->
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable {
                        viewModel.setCity(name, code)
                        navController.popBackStack()
                    },
                    color = AppColors.White,
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 24.dp, vertical = 20.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            name,
                            fontSize = 18.sp,
                            color = if (uiState.currentCityCode == code) AppColors.Primary else AppColors.TextPrimary,
                            fontWeight = if (uiState.currentCityCode == code) FontWeight.SemiBold else FontWeight.Normal,
                        )
                        Spacer(modifier = Modifier.weight(1f))
                        if (uiState.currentCityCode == code) {
                            Text("‚ú?, color = AppColors.Primary, fontSize = 20.sp)
                        }
                    }
                }
                HorizontalDivider(thickness = 0.5.dp, color = AppColors.StepInactive)
            }
        }
    }
}
```

---

### Task 4: P4 Profile Screen (Multi-Step)

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/profile/ProfileScreen.kt`

- [ ] **Step 1: Write ProfileScreen.kt**

```kotlin
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
import androidx.compose.ui.graphics.Color
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

    val genders = listOf("male" to "Áî?, "female" to "Â•?)
    val employmentStatuses = listOf(
        "employed" to "‰ºÅ‰∏öÂ∞±‰∏ö", "flexible" to "ÁÅµÊ¥ªÂ∞±‰∏ö",
        "self_employed" to "Ëá™Èõá", "unemployed" to "Â§±‰∏ö",
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
        // Step bar
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
                Text("Âü∫Êú¨‰ø°ÊÅØ", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                genderOptions(genders, gender) { gender = it }
                OutlinedTextField(value = age, onValueChange = { age = it }, label = { Text("Âπ¥ÈæÑ(16-70)") }, singleLine = true)
                OutlinedTextField(value = household, onValueChange = { household = it }, label = { Text("Êà∑Á±çÂú?) }, singleLine = true)
                Spacer(modifier = Modifier.height(20.dp))
                Button(onClick = { currentStep = 2 }, modifier = Modifier.fillMaxWidth().height(48.dp), shape = RoundedCornerShape(48.dp)) {
                    Text("‰∏ã‰∏ÄÊ≠?)
                }
            }
            2 -> {
                Text("Â∞±‰∏ö‰ø°ÊÅØ", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                employmentOptions(employmentStatuses, employment) { employment = it }
                OutlinedTextField(value = years, onValueChange = { years = it }, label = { Text("Á¥ØËÆ°Á§æ‰øùÂπ¥Èôê(Âπ?") }, singleLine = true)
                Spacer(modifier = Modifier.height(20.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    OutlinedButton(onClick = { currentStep = 1 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("‰∏ä‰∏ÄÊ≠?)
                    }
                    Button(onClick = { currentStep = 3 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("‰∏ã‰∏ÄÊ≠?)
                    }
                }
            }
            3 -> {
                Text("Ë°•ÂÖÖ‰ø°ÊÅØ", fontSize = 24.sp, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(20.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("Â≠êÂ•≥ÊÉÖÂÜµ")
                    Spacer(modifier = Modifier.width(16.dp))
                    Switch(checked = hasChildren, onCheckedChange = { hasChildren = it })
                }
                OutlinedTextField(value = budget, onValueChange = { budget = it }, label = { Text("ÊúàÈ¢ÑÁÆóÔºàÂÖÉÔºâ") }, singleLine = true)
                OutlinedTextField(value = balance, onValueChange = { balance = it }, label = { Text("ÂΩìÂâçË¥¶Êà∑‰ΩôÈ¢ùÔºàÂÖÉÔº?) }, singleLine = true)
                if (errorMessage.isNotEmpty()) {
                    Text(errorMessage, color = AppColors.Error, fontSize = 14.sp)
                }
                Spacer(modifier = Modifier.height(20.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    OutlinedButton(onClick = { currentStep = 2 }, modifier = Modifier.weight(1f).height(48.dp), shape = RoundedCornerShape(48.dp)) {
                        Text("‰∏ä‰∏ÄÊ≠?)
                    }
                    Button(
                        onClick = {
                            val ageInt = age.toIntOrNull()
                            if (ageInt == null || ageInt !in 16..70) {
                                errorMessage = "Âπ¥ÈæÑÂøÖÈ°ªÂú?6-70‰πãÈó¥"
                                return@Button
                            }
                            val budgetDouble = budget.toDoubleOrNull()
                            if (budgetDouble == null || budgetDouble <= 0) {
                                errorMessage = "ËØ∑ËæìÂÖ•ÊúâÊïàÊúàÈ¢ÑÁÆó"
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
                                    val client = NSIClient("http://localhost:30001", uiState.userInfo.nickName)
                                    val json = Json { ignoreUnknownKeys = true }
                                    client.updateProfile(json.encodeToString(profileData))
                                    navController.navigate(Routes.LOADING) {
                                        popUpTo(Routes.PROFILE) { inclusive = true }
                                    }
                                } catch (e: Exception) {
                                    errorMessage = e.message ?: "Êèê‰∫§Â§±Ë¥•"
                                    submitting = false
                                }
                            }
                        },
                        enabled = !submitting,
                        modifier = Modifier.weight(1f).height(48.dp),
                        shape = RoundedCornerShape(48.dp),
                    ) {
                        Text(if (submitting) "Êèê‰∫§‰∏?.." else "ÁîüÊàêÊñπÊ°à")
                    }
                }
            }
        }
    }
}

@Composable
private fun genderOptions(options: List<Pair<String, String>>, selected: String, onSelect: (String) -> Unit) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Text("ÊÄßÂà´")
        Spacer(modifier = Modifier.width(16.dp))
        options.forEach { (value, label) ->
            Row(verticalAlignment = Alignment.CenterVertically) {
                RadioButton(selected = value == selected, onClick = { onSelect(value) })
                Text(label)
            }
        }
    }
}

@Composable
private fun employmentOptions(options: List<Pair<String, String>>, selected: String, onSelect: (String) -> Unit) {
    var expanded by remember { mutableStateOf(false) }
    ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
        OutlinedTextField(
            value = options.find { it.first == selected }?.second ?: "ËØ∑ÈÄâÊã©",
            onValueChange = {},
            readOnly = true,
            label = { Text("Â∞±‰∏öÁä∂ÊÄ?) },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
            modifier = Modifier.menuAnchor(),
        )
        ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            options.forEach { (value, label) ->
                DropdownMenuItem(
                    text = { Text(label) },
                    onClick = { onSelect(value); expanded = false },
                )
            }
        }
    }
}
```

---

### Task 5: P5 Loading + P6 Preview Screens

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/loading/LoadingScreen.kt`
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/preview/PreviewScreen.kt`

- [ ] **Step 1: Write LoadingScreen.kt**

```kotlin
package com.nsi.app.ui.loading

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.nsi.app.Routes
import com.nsi.app.models.AppViewModel
import com.nsi.app.models.PlanResult
import com.nsi.app.models.Scheme
import com.nsi.app.theme.AppColors
import com.nsi.sdk.NSIClient
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
fun LoadingScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    var progressText by remember { mutableStateOf("Ê≠£Âú®ÂàÜÊûêÊÇ®ÁöÑÁ§æ‰øùÊÉÖÂÜµ...") }
    val scope = rememberCoroutineScope()

    LaunchedEffect(Unit) {
        val input = uiState.planInput ?: return@LaunchedEffect
        progressText = "Ê≠£Âú®ËÆ°ÁÆóÊúÄ‰ºòÊñπÊ°?.."

        scope.launch {
            try {
                val client = NSIClient("http://localhost:30001", uiState.userInfo.nickName)
                val json = kotlinx.serialization.json.Json { ignoreUnknownKeys = true }
                val plan = client.generatePlan(json.encodeToString(input))
                viewModel.setPlanResult(PlanResult(
                    planId = plan.planId ?: "",
                    recommendedSchemes = (plan.recommendedSchemes ?: emptyList()).map { s ->
                        Scheme(
                            name = s.name ?: "",
                            baseSalary = s.baseSalary ?: 0,
                            monthlyCost = s.monthlyCost ?: 0.0,
                            projectedPension = s.projectedPension ?: 0.0,
                            annualSubsidy = s.annualSubsidy ?: 0.0,
                        )
                    }
                ))
                delay(800)
                navController.navigate(Routes.PREVIEW) {
                    popUpTo(Routes.LOADING) { inclusive = true }
                }
            } catch (_: Exception) {
                progressText = "ÁîüÊàêÂ§±Ë¥•ÔºåËØ∑ÈáçËØï"
                delay(2000)
            }
        }
    }

    Column(
        modifier = Modifier.fillMaxSize().background(AppColors.Background),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        CircularProgressIndicator(color = AppColors.Primary, strokeWidth = 6.dp, modifier = Modifier.size(64.dp))
        Spacer(modifier = Modifier.height(24.dp))
        Text("AI Ê≠£Âú®‰∏∫ÊÇ®ÂåπÈÖçÊúÄ‰ºòÊîøÁ≠?..", fontSize = 18.sp)
        Spacer(modifier = Modifier.height(16.dp))
        Text(progressText, fontSize = 14.sp, color = AppColors.TextSecondary)
    }
}
```

- [ ] **Step 2: Write PreviewScreen.kt**

```kotlin
package com.nsi.app.ui.preview

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.graphics.Brush
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
        Text("ÊÇ®ÁöÑÁ§æ‰øùÊñπÊ°à", fontSize = 28.sp, fontWeight = FontWeight.Bold)
        Text("‰ª•‰∏ãÊò?AI ‰∏∫ÊÇ®Êé®ËçêÁöÑÁ§æ‰øùÊñπÊ°àÈ¢ÑËß?, fontSize = 16.sp, color = AppColors.TextSecondary)

        Spacer(modifier = Modifier.height(24.dp))

        // Blurred cards
        Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
            Column(modifier = Modifier.padding(32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text("È¢ÑËÆ°Âπ¥Ë°•Ë¥¥È¢ù", fontSize = 16.sp, color = AppColors.TextSecondary)
                Text("Á∫?¬•XX,XXX/Âπ?, fontSize = 32.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary, modifier = Modifier.blur(12.dp))
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
            Column(modifier = Modifier.padding(32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text("ÊúàËäÇÁúÅÈ¢ù", fontSize = 16.sp, color = AppColors.TextSecondary)
                Text("Á∫?¬•XXX/Êú?, fontSize = 32.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary, modifier = Modifier.blur(12.dp))
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
            colors = ButtonDefaults.buttonColors(
                containerColor = AppColors.Primary
            ),
        ) {
            Text("Ëß£ÈîÅÂÆåÊï¥Êä•Âëä", fontWeight = FontWeight.SemiBold)
        }

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            "Ëß£ÈîÅÂêéÂèØÊü•ÁúãÊñπÊ°àÂØπÊØî„ÄÅÂπ¥Â∫¶Áé∞ÈáëÊµÅÂíåË°åÂä®Ê∏ÖÂç?,
            fontSize = 13.sp,
            color = AppColors.TextMuted,
            textAlign = TextAlign.Center,
        )
    }
}
```

---

### Task 6: P7 Plan Detail Screen

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Create: `frontend/android/app/src/main/java/com/nsi/app/ui/plan/PlanDetailScreen.kt`

- [ ] **Step 1: Write PlanDetailScreen.kt**

```kotlin
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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
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

    LaunchedEffect(planId) {
        val cached = uiState.planResult
        if (cached != null) {
            schemes = cached.recommendedSchemes
        } else {
            try {
                val client = NSIClient("http://localhost:30001", uiState.userInfo.nickName)
                val plan = client.getPlanDetail(uiState.userInfo.nickName, planId)
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
            TopAppBar(title = { Text("ÊñπÊ°àËØ¶ÊÉÖ") })
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
            // Scheme cards
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
                            Text("È¢ÑËÆ°ÂÖªËÄÅÈáë", fontSize = 13.sp, color = AppColors.TextSecondary)
                            Text("${s.projectedPension.toInt()}ÂÖ?Êú?, fontSize = 24.sp, fontWeight = FontWeight.Bold, color = AppColors.Primary)
                            Text("ÊúàÁº¥Ë¥? ${s.monthlyCost.toInt()}ÂÖ?, fontSize = 13.sp, color = AppColors.TextSecondary)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            // Detail card
            if (schemes.isNotEmpty() && currentIndex < schemes.size) {
                val s = schemes[currentIndex]
                Card(modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(16.dp)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        detailRow("Áº¥Ë¥πÂü∫Êï∞", "${s.baseSalary}ÂÖ?)
                        detailRow("ÊúàÁº¥Ë¥?, "${s.monthlyCost.toInt()}ÂÖ?)
                        detailRow("‰ºÅ‰∏öÂπ¥Ë°•Ë¥?, "${s.annualSubsidy.toInt()}ÂÖ?Âπ?, highlight = true)
                        detailRow("È¢ÑËÆ°ÂÖªËÄÅÈáë", "${s.projectedPension.toInt()}ÂÖ?Êú?, highlight = true)
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = {},
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(48.dp),
            ) {
                Text("‰øùÂ≠ò‰∏?PDF")
            }

            Spacer(modifier = Modifier.height(12.dp))

            OutlinedButton(
                onClick = {},
                modifier = Modifier.fillMaxWidth().height(48.dp),
                shape = RoundedCornerShape(48.dp),
            ) {
                Text("ÂàÜ‰∫´ÊñπÊ°à")
            }
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
```

---

### Task 7: Update NSIClient SDK

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

**Files:**
- Modify: `frontend/android/nsi-sdk/Client.kt`

- [ ] **Step 1: Add annualSubsidy to Scheme + getPlanDetail method**

```kotlin
    @Serializable
    data class Scheme(val name: String? = null, val baseSalary: Int? = null, val monthlyCost: Double? = null, val projectedPension: Double? = null, val annualSubsidy: Double? = null)

    suspend fun getPlanDetail(userID: String, planID: String): PlanSnapshot {
        val resp = get("/v1/plans/$planID")
        val wrapper = json.decodeFromString<ResponseWrapper<PlanSnapshot>>(resp)
        return wrapper.data
    }
```

Add the `getPlanDetail` method after `generatePlan`.

- [ ] **Step 2: Add ResponseWrapper (if not present)**

Add after the class body:
```kotlin
@Serializable
data class ResponseWrapper<T>(val code: Int, val data: T)
```

---

### Task 8: Final Verification

| **∞Ê±æ∫≈** | V1.0.0 |
| **◊¥Ã¨** | “—…˙–ß |
| **∑¢≤º»’∆⁄** | 2026-06-15 |

- [ ] **Step 1: Verify all files exist**

Run: `Get-ChildItem -Recurse -LiteralPath "nsi-platform/frontend/android" -Filter "*.kt"`

Expected: Client.kt + MainActivity.kt + NSIApp.kt + AppState.kt + Color.kt + Theme.kt + 7 screen files

- [ ] **Step 2: Verify navigation flow**

Trace: LoginScreen ‚Ü?HOME ‚Ü?CityPickerScreen ‚Ü?HOME ‚Ü?PROFILE ‚Ü?LOADING ‚Ü?PREVIEW ‚Ü?PLAN

- [ ] **Step 3: Verify API coverage**

- `updateProfile` ‚Ä?ProfileScreen ‚ú?- `queryPolicies` ‚Ä?HomeScreen ‚ú?- `generatePlan` ‚Ä?LoadingScreen ‚ú?- `getPlanDetail` ‚Ä?PlanDetailScreen ‚ú?(need to add to Client.kt)
