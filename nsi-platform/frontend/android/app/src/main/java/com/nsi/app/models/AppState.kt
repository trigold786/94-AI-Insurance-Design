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
    val userInfo: UserInfo = UserInfo("default"),
    val currentCity: String = "上海",
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
