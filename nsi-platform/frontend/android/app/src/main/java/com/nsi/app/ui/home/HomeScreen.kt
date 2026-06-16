package com.nsi.app.ui.home

import android.Manifest
import android.content.Context
import android.location.LocationManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
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
import androidx.compose.ui.platform.LocalContext
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
    val context = LocalContext.current
    var locationChecked by remember { mutableStateOf(false) }

    val locationLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            try {
                val lm = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
                val loc = lm.getLastKnownLocation(LocationManager.GPS_PROVIDER)
                    ?: lm.getLastKnownLocation(LocationManager.NETWORK_PROVIDER)
                if (loc != null) {
                    val (city, code) = mapLocationToCity(loc.latitude, loc.longitude)
                    viewModel.setCity(city, code)
                }
            } catch (_: Exception) {}
        }
    }

    LaunchedEffect(Unit) {
        if (!locationChecked) {
            locationChecked = true
            locationLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
        }
    }

    LaunchedEffect(uiState.currentCityCode) {
        scope.launch {
            try {
                val client = NSIClient("http://127.0.0.1:39401", uiState.userInfo.nickName)
                val claims = client.queryPolicies(regionCode = uiState.currentCityCode)
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
        TextButton(onClick = { navController.navigate(Routes.CITY_PICKER) }) {
            Text("📍", fontSize = 18.sp)
            Spacer(modifier = Modifier.width(4.dp))
            Text(uiState.currentCity, fontSize = 18.sp, fontWeight = FontWeight.SemiBold, color = AppColors.TextPrimary)
            Text(" ▼", fontSize = 14.sp, color = AppColors.TextMuted)
        }

        Card(
            modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = AppColors.Primary),
        ) {
            Column(
                modifier = Modifier.padding(40.dp).fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text("社保规划，智能匹配", fontSize = 26.sp, fontWeight = FontWeight.Bold, color = AppColors.White)
                Spacer(modifier = Modifier.height(12.dp))
                Text("AI 为您匹配最优社保方案", fontSize = 18.sp, color = AppColors.White.copy(alpha = 0.8f))
                Spacer(modifier = Modifier.height(24.dp))
                Button(
                    onClick = { navController.navigate(Routes.PROFILE) },
                    modifier = Modifier.fillMaxWidth().height(48.dp),
                    shape = RoundedCornerShape(48.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = AppColors.White.copy(alpha = 0.2f)),
                ) {
                    Text("开始社保筹划", fontWeight = FontWeight.SemiBold)
                }
            }
        }

        // Compliance quick link
        Surface(
            onClick = { navController.navigate(Routes.COMPLIANCE) },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            color = AppColors.White,
        ) {
            Row(
                modifier = Modifier.padding(16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("📋", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(12.dp))
                Text("合规认定与材料清单", fontSize = 16.sp, fontWeight = FontWeight.Medium)
                Spacer(modifier = Modifier.weight(1f))
                Text("›", fontSize = 20.sp, color = AppColors.TextMuted)
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        Surface(
            onClick = { navController.navigate(Routes.RIGHTS) },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            color = AppColors.White,
        ) {
            Row(
                modifier = Modifier.padding(16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text("🛡️", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(12.dp))
                Text("权益监测与风险预警", fontSize = 16.sp, fontWeight = FontWeight.Medium)
                Spacer(modifier = Modifier.weight(1f))
                Text("›", fontSize = 20.sp, color = AppColors.TextMuted)
            }
        }

        Text("政策速览", fontSize = 20.sp, fontWeight = FontWeight.SemiBold)

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

private fun mapLocationToCity(lat: Double, lng: Double): Pair<String, String> {
    val cities = listOf(
        Triple(31.23, 121.47, "上海" to "310000"),
        Triple(39.90, 116.40, "北京" to "110000"),
        Triple(22.54, 114.06, "深圳" to "440300"),
        Triple(23.13, 113.27, "广州" to "440100"),
        Triple(30.27, 120.15, "杭州" to "330100"),
    )
    var best = cities[0]
    var bestDist = Double.MAX_VALUE
    for (c in cities) {
        val d = (c.first - lat) * (c.first - lat) + (c.second - lng) * (c.second - lng)
        if (d < bestDist) { bestDist = d; best = c }
    }
    return best.third
}
