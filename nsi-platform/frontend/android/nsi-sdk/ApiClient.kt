package com.nsi.sdk

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

/**
 * Bearer-token API client for the NSI backend.
 * Follows the same networking pattern as the existing [NSIClient] in this module
 * (java.net.http.HttpClient + kotlinx.serialization + ResponseWrapper unwrapping),
 * but authenticates with `Authorization: Bearer <token>` and exposes the
 * SMS / settings / simulator / advisor / compliance / feedback endpoints.
 */
class ApiClient(
    private val baseURL: String = DEFAULT_BASE_URL,
    private val tokenProvider: () -> String?,
) {
    private val client = HttpClient.newHttpClient()
    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    // ---- SMS auth ----
    suspend fun sendSms(phone: String): Boolean = callCode("POST", "/v1/auth/sms/send", """{"phone":"$phone"}""")

    suspend fun verifySms(phone: String, code: String): SmsVerifyResult = call("POST", "/v1/auth/sms/verify", """{"phone":"$phone","code":"$code"}""")

    // ---- Profile ----
    suspend fun getProfile(): ProfileResponse = call("/v1/profile")
    suspend fun updateProfile(body: String): ProfileResponse = call("/v1/profile", "PUT", body)

    // ---- Settings ----
    suspend fun getSettings(): SettingsResponse = call("/v1/settings")
    suspend fun updateSettings(body: String): SettingsResponse = call("/v1/settings", "POST", body)

    // ---- Simulator ----
    suspend fun calculate(body: String): SimulatorResult = call("/v1/simulator/calculate", "POST", body)

    // ---- Advisor ----
    suspend fun askAdvisor(question: String, contextJson: String): AdvisorResult {
        val body = buildJsonObject {
            put("question", question)
            put("context", json.parseToJsonElement(contextJson))
        }
        return call("/v1/advisor/ask", "POST", json.encodeToString(JsonObject.serializer(), body))
    }

    // ---- Compliance ----
    suspend fun getComplianceChecklist(cityCode: String? = null): ComplianceChecklist {
        val q = if (cityCode != null) "?city_code=$cityCode" else ""
        return call("/v1/compliance/checklist$q")
    }

    // ---- Feedback ----
    suspend fun submitFeedback(body: String): Boolean = callCode("POST", "/v1/feedback", body)

    // ---- internals ----
    private inline fun <reified T> call(path: String, method: String = "GET", body: String? = null): T {
        val resp = withIO { send(method, path, body) }
        val wrapper = json.decodeFromString(ApiResponse.serializer(serializer()), resp)
        if (wrapper.code != 0) throw RuntimeException(wrapper.msg ?: "API error")
        return wrapper.data ?: throw RuntimeException("Empty response data")
    }

    private fun callCode(method: String, path: String, body: String?): Boolean {
        val resp = withIO { send(method, path, body) }
        val wrapper = json.decodeFromString(ApiResponse.serializer(JsonPrimitive.serializer()), resp)
        return wrapper.code == 0
    }

    private inline fun <T> withIO(block: () -> T): T = kotlinx.coroutines.runBlocking(Dispatchers.IO) { block() }

    private fun send(method: String, path: String, body: String?): String {
        val token = tokenProvider()
        val req = HttpRequest.newBuilder().uri(URI.create("$baseURL$path"))
            .header("Content-Type", "application/json")
        if (!token.isNullOrEmpty()) req.header("Authorization", "Bearer $token")
        when (method) {
            "GET" -> req.GET()
            "PUT" -> req.PUT(HttpRequest.BodyPublishers.ofString(body ?: ""))
            "POST" -> req.POST(HttpRequest.BodyPublishers.ofString(body ?: ""))
        }
        val resp = client.send(req.build(), HttpResponse.BodyHandlers.ofString())
        if (resp.statusCode() == 401) throw RuntimeException("请先登录")
        if (resp.statusCode() >= 400) throw RuntimeException("请求失败(${resp.statusCode()})")
        return resp.body()
    }

    @Serializable
    data class ApiResponse<T>(val code: Int, val data: T? = null, val msg: String? = null)

    companion object {
        const val DEFAULT_BASE_URL = "http://127.0.0.1:39401"
    }
}

@Serializable
data class SmsVerifyResult(
    @SerialName("user_id") val userId: String? = null,
    val token: String? = null,
)

@Serializable
data class ProfileResponse(
    val age: Int? = null,
    val gender: String? = null,
    @SerialName("is_local_hukou") val isLocalHukou: Boolean? = null,
    @SerialName("monthly_income") val monthlyIncome: Double? = null,
    @SerialName("child_age_range") val childAgeRange: String? = null,
    @SerialName("has_elderly_dependents") val hasElderlyDependents: Boolean? = null,
    @SerialName("employment_status") val employmentStatus: String? = null,
    @SerialName("social_security_years") val socialSecurityYears: Int? = null,
)

@Serializable
data class SettingsResponse(
    @SerialName("font_scale") val fontScale: String? = null,
    @SerialName("default_tab") val defaultTab: String? = null,
    @SerialName("notifications_on") val notificationsOn: Boolean? = null,
)

@Serializable
data class SimulatorResult(
    @SerialName("monthly_cost") val monthlyCost: Double? = null,
    @SerialName("monthly_pension") val monthlyPension: Double? = null,
    @SerialName("annual_subsidy") val annualSubsidy: Double? = null,
    @SerialName("net_monthly") val netMonthly: Double? = null,
    @SerialName("policy_triggers") val policyTriggers: List<PolicyTrigger>? = null,
    @SerialName("annual_cashflow") val annualCashflow: List<CashflowPoint>? = null,
)

@Serializable
data class PolicyTrigger(
    val level: String? = null,
    val title: String? = null,
    val message: String? = null,
)

@Serializable
data class CashflowPoint(
    val year: Int? = null,
    val amount: Double? = null,
)

@Serializable
data class AdvisorResult(
    val answer: String? = null,
)

@Serializable
data class ComplianceChecklist(
    @SerialName("matched_policies") val matchedPolicies: List<PolicyInfo>? = null,
    @SerialName("eligible_tags") val eligibleTags: List<String>? = null,
    @SerialName("action_checklist") val actionChecklist: List<String>? = null,
)

@Serializable
data class PolicyInfo(
    val title: String? = null,
    val description: String? = null,
    @SerialName("policy_type") val policyType: String? = null,
    @SerialName("subsidy_calc_method") val subsidyCalcMethod: String? = null,
    @SerialName("is_eligible") val isEligible: Boolean? = null,
)
