import Foundation

public class NSIClient {
    private let baseURL: URL
    private let userID: String
    private let session: URLSession
    private let decoder: JSONDecoder

    public init(baseURL: String, userID: String) throws {
        guard let url = URL(string: baseURL), !baseURL.isEmpty else {
            throw NSError(domain: "NSIAPI", code: 1, userInfo: [NSLocalizedDescriptionKey: "Invalid baseURL"])
        }
        self.baseURL = url
        self.userID = userID
        self.session = URLSession.shared
        self.decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
    }

    public struct UserProfile: Codable {
        public let userID: String?
        public let age: Int?
        public let gender: String?
        public let employmentStatus: String?
        public let socialSecurityYears: Int?
        public let hasChildren: Bool?
    }

    public struct PolicyClaim: Codable, Identifiable {
        public var id: String { claimID ?? policyID ?? UUID().uuidString }
        public let claimID: String?
        public let policyID: String?
        public let policyType: String?
        public let regionCode: String?
        public let subsidyCalcMethod: String?
        public let sourceName: String?
        public let policyURL: String?
    }

    public struct Scheme: Codable, Identifiable {
        public var id: String { name ?? UUID().uuidString }
        public let name: String?
        public let baseSalary: Int?
        public let monthlyCost: Double?
        public let projectedPension: Double?
        public let annualSubsidy: Double?
    }

    public struct PlanSnapshot: Codable {
        public let planID: String?
        public let recommendedSchemes: [Scheme]?
        public let totalCost: Double?
        public let totalSubsidy: Double?
    }

    public struct Alert: Codable, Identifiable {
        public var id: String { alertID ?? UUID().uuidString }
        public let alertID: String?
        public let alertType: String?
        public let severity: String?
        public let title: String?
        public let message: String?
        public let isRead: Bool?
    }

    public struct ComplianceResult: Codable {
        public let matchedPolicies: [PolicyClaim]?
        public let requiredDocs: [String]?
    }

    private func request<T: Codable>(_ path: String, method: String = "GET", body: Data? = nil) async throws -> T {
        guard let url = URL(string: path, relativeTo: baseURL) else {
            throw NSError(domain: "NSIAPI", code: 2, userInfo: [NSLocalizedDescriptionKey: "Invalid path"])
        }
        var req = URLRequest(url: url)
        req.setValue(userID, forHTTPHeaderField: "x-user-id")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpMethod = method
        req.httpBody = body
        req.timeoutInterval = 30

        let (data, response) = try await session.data(for: req)
        guard let http = response as? HTTPURLResponse else {
            throw NSError(domain: "NSIAPI", code: 3, userInfo: [NSLocalizedDescriptionKey: "Invalid response"])
        }
        if http.statusCode == 401 { throw NSError(domain: "NSIAPI", code: 401, userInfo: [NSLocalizedDescriptionKey: "请先登录"]) }
        if http.statusCode == 404 { throw NSError(domain: "NSIAPI", code: 404, userInfo: [NSLocalizedDescriptionKey: "未找到"]) }
        if http.statusCode >= 400 {
            let msg = String(data: data, encoding: .utf8) ?? "请求失败"
            throw NSError(domain: "NSIAPI", code: http.statusCode, userInfo: [NSLocalizedDescriptionKey: msg])
        }

        let wrapper = try decoder.decode(ResponseWrapper<T>.self, from: data)
        if wrapper.code != 0 {
            throw NSError(domain: "NSIAPI", code: wrapper.code, userInfo: [NSLocalizedDescriptionKey: wrapper.msg ?? "服务异常"])
        }
        return wrapper.data
    }

    public func getProfile() async throws -> UserProfile { try await request("/v1/profile") }
    public func updateProfile(json: String) async throws -> UserProfile {
        try await request("/v1/profile", method: "PUT", body: json.data(using: .utf8))
    }
    public func queryPolicies(regionCode: String? = nil) async throws -> [PolicyClaim] {
        var path = "/v1/policies"
        if let r = regionCode { path += "?region_code=\(r)" }
        return try await request(path)
    }
    public func generatePlan(json: String) async throws -> PlanSnapshot {
        try await request("/v1/plans/generate", method: "POST", body: json.data(using: .utf8))
    }
    public func getPlanDetail(planID: String) async throws -> PlanSnapshot {
        try await request("/v1/plans/\(planID)")
    }
    public func getPlanReport(planID: String) async throws -> String {
        let path = "/v1/plans/report?plan_id=\(planID)"
        guard let url = URL(string: path, relativeTo: baseURL) else { return "" }
        var req = URLRequest(url: url)
        req.setValue(userID, forHTTPHeaderField: "x-user-id")
        let (data, _) = try await session.data(for: req)
        return String(data: data, encoding: .utf8) ?? ""
    }
    public func getCompliance(cityCode: String) async throws -> ComplianceResult {
        try await request("/v1/compliance/checklist?city_code=\(cityCode)")
    }
    public func getPaymentStatus() async throws -> [String: Any] {
        try await request("/v1/rights/payment-status")
    }
    public func getAlerts() async throws -> [Alert] {
        try await request("/v1/rights/alerts")
    }
    public func markAlertRead(alertID: String) async throws -> [String: String] {
        try await request("/v1/rights/alerts/read", method: "POST", body: try JSONEncoder().encode(["alert_id": alertID]))
    }

    private struct ResponseWrapper<T: Codable>: Codable {
        let code: Int
        let data: T
        let msg: String?
    }
}
