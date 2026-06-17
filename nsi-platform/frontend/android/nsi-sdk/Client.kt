package com.nsi.sdk

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse

class NSIClient(private val baseURL: String, private val userID: String) {
    private val client = HttpClient.newHttpClient()
    private val json = Json { ignoreUnknownKeys = true; isLenient = true }

    @Serializable data class UserProfile(val userId: String? = null, val age: Int? = null, val gender: String? = null, val employmentStatus: String? = null, val socialSecurityYears: Int? = null)
    @Serializable data class PolicyClaim(val claimId: String? = null, val policyId: String? = null, val policyType: String? = null, val regionCode: String? = null, val subsidyCalcMethod: String? = null, val sourceName: String? = null)
    @Serializable data class Scheme(val name: String? = null, val baseSalary: Int? = null, val monthlyCost: Double? = null, val projectedPension: Double? = null, val annualSubsidy: Double? = null)
    @Serializable data class PlanSnapshot(val planId: String? = null, val recommendedSchemes: List<Scheme>? = null, val totalCost: Double? = null, val totalSubsidy: Double? = null)
    @Serializable data class Alert(val alertId: String? = null, val alertType: String? = null, val severity: String? = null, val title: String? = null, val message: String? = null, val isRead: Boolean? = null)
    @Serializable data class ComplianceResult(val matchedPolicies: List<PolicyClaim>? = null)
    @Serializable data class Order(@SerialName("order_id") val orderId: String? = null, val amount: Double? = null, val status: String? = null)
    @Serializable data class UnlockResult(val unlocked: Boolean = false)

    suspend fun getProfile(): UserProfile { return call("/v1/profile") }
    suspend fun updateProfile(body: String): UserProfile { return call("/v1/profile", "PUT", body) }
    suspend fun queryPolicies(regionCode: String? = null): List<PolicyClaim> {
        val q = if (regionCode != null) "?region_code=$regionCode" else ""
        return call("/v1/policies$q")
    }
    suspend fun generatePlan(body: String): PlanSnapshot { return call("/v1/plans/generate", "POST", body) }
    suspend fun getPlanDetail(planID: String): PlanSnapshot { return call("/v1/plans/$planID") }
    suspend fun getCompliance(cityCode: String): ComplianceResult { return call("/v1/compliance/checklist?city_code=$cityCode") }
    suspend fun getPaymentStatus(): String { return send("GET", "/v1/rights/payment-status", null) }
    suspend fun getAlerts(): List<Alert> { return call("/v1/rights/alerts") }
    suspend fun markAlertRead(alertID: String) { call<Unit>("/v1/rights/alerts/read", "POST", """{"alert_id":"$alertID"}""") }
    suspend fun createOrder(planID: String): Order { return call("/v1/orders", "POST", """{"plan_id":"$planID"}""") }
    suspend fun payOrder(orderID: String, paymentMethod: String = "wechat"): Order { return call("/v1/orders/$orderID/pay", "POST", """{"payment_method":"$paymentMethod"}""") }
    suspend fun checkUnlock(planID: String): UnlockResult { return call("/v1/orders/check-unlock?plan_id=$planID") }

    private inline suspend fun <reified T> call(path: String, method: String = "GET", body: String? = null): T {
        val resp = send(method, path, body)
        val wrapper = json.decodeFromString<ResponseWrapper<T>>(resp)
        if (wrapper.code != 0) throw RuntimeException(wrapper.msg ?: "API error")
        return wrapper.data
    }

    private suspend fun send(method: String, path: String, body: String?): String {
        val req = HttpRequest.newBuilder().uri(URI.create("$baseURL$path"))
            .header("x-user-id", userID).header("Content-Type", "application/json")
        when (method) {
            "GET" -> req.GET()
            "PUT" -> req.PUT(HttpRequest.BodyPublishers.ofString(body ?: ""))
            "POST" -> req.POST(HttpRequest.BodyPublishers.ofString(body ?: ""))
        }
        val resp = client.send(req.build(), HttpResponse.BodyHandlers.ofString())
        if (resp.statusCode() == 401) throw RuntimeException("请先登录")
        if (resp.statusCode() >= 400) throw RuntimeException("请求失败(${resp.statusCode()})")
        return resp.body()
    suspend fun getPlanReport(planID: String): String {
        return send("GET", "/v1/plans/report?plan_id=$planID", null)
    }
    suspend fun getGuide(cityCode: String): String {
        return send("GET", "/v1/guide?city_code=$cityCode", null)
    }
    suspend fun saveScenario(name: String, paramsJson: String) {
        send("POST", "/v1/simulator/scenarios", """{"name":"$name","params":$paramsJson}""")
    }
    suspend fun listScenarios(): String {
        return send("GET", "/v1/simulator/scenarios", null)
    }
    suspend fun deleteAccount() {
        send("POST", "/v1/auth/delete-account-v2", """{"confirm":"DELETE"}""")
    }
    suspend fun submitPaymentRecord(policyType: String, month: String, amount: Double, status: String, dueDate: String) {
        send("POST", "/v1/rights/payment-records", """{"policy_type":"$policyType","month":"$month","amount":$amount,"status":"$status","due_date":"$dueDate"}""")
    }
}

@Serializable data class ResponseWrapper<T>(val code: Int, val data: T, val msg: String? = null)
