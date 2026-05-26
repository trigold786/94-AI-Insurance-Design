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
    "310000" to "上海", "110000" to "北京",
    "440300" to "深圳", "440100" to "广州",
    "330100" to "杭州",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CityPickerScreen(navController: NavController, viewModel: AppViewModel) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("选择城市") })
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
                            Text("✓", color = AppColors.Primary, fontSize = 20.sp)
                        }
                    }
                }
                HorizontalDivider(thickness = 0.5.dp, color = AppColors.StepInactive)
            }
        }
    }
}
