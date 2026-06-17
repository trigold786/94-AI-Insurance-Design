import SwiftUI

struct RightsView: View {
    @EnvironmentObject var appState: AppState
    @State private var summary: PaymentSummary?
    @State private var alerts: [AlertItem] = []
    @State private var isLoading = true

    struct PaymentSummary: Codable {
        let paidCount: Int
        let pendingCount: Int
        let missedCount: Int
        let recentRecords: [PaymentRecord]

        enum CodingKeys: String, CodingKey {
            case paidCount = "paid_count"
            case pendingCount = "pending_count"
            case missedCount = "missed_count"
            case recentRecords = "recent_records"
        }
    }

    struct PaymentRecord: Codable, Identifiable {
        var id: String { recordID }
        let recordID: String
        let policyType: String
        let month: String
        let amount: Double
        let status: String

        enum CodingKeys: String, CodingKey {
            case recordID = "record_id"
            case policyType = "policy_type"
            case month, amount, status
        }
    }

    struct AlertItem: Codable, Identifiable {
        var id: String { alertID }
        let alertID: String
        let alertType: String
        let severity: String
        let title: String
        let message: String
        let isRead: Bool

        enum CodingKeys: String, CodingKey {
            case alertID = "alert_id"
            case alertType = "alert_type"
            case severity, title, message
            case isRead = "is_read"
        }
    }

    var body: some View {
        ScrollView {
            if isLoading {
                ProgressView().padding(.top, 80)
            } else {
                VStack(alignment: .leading, spacing: 20) {
                    Text("权益监测").font(.system(size: 24, weight: .bold))

                    if let s = summary {
                        HStack(spacing: 16) {
                            summaryCard(value: s.paidCount, label: "已缴纳", color: Color(red: 0, green: 0.6, blue: 0.4))
                            summaryCard(value: s.pendingCount, label: "待缴纳", color: .orange)
                            summaryCard(value: s.missedCount, label: "已断缴", color: .red)
                        }
                    }

                    Text("预警通知 (\(alerts.count))").font(.system(size: 18, weight: .semibold))
                    if alerts.isEmpty {
                        Text("暂无预警").font(.system(size: 14)).foregroundColor(.gray)
                    }
                    ForEach(alerts) { alert in
                        VStack(alignment: .leading, spacing: 6) {
                            HStack {
                                severityBadge(alert.severity)
                                Text(alert.alertType == "disconnection_risk" ? "断缴风险" : "政策变更").font(.system(size: 12)).foregroundColor(.gray)
                                if alert.isRead { Spacer(); Text("已读").font(.system(size: 11)).foregroundColor(.gray) }
                            }
                            Text(alert.title).font(.system(size: 16, weight: .semibold))
                            Text(alert.message).font(.system(size: 13)).foregroundColor(.gray)
                        }
                        .padding()
                        .background(Color.white)
                        .cornerRadius(12)
                    }

                    if let s = summary, !s.recentRecords.isEmpty {
                        Text("近期缴费记录 (\(s.recentRecords.count))").font(.system(size: 18, weight: .semibold))
                        ForEach(s.recentRecords) { record in
                            HStack {
                                Text(record.month).font(.system(size: 14)).frame(width: 70, alignment: .leading)
                                Text(record.policyType).font(.system(size: 13)).foregroundColor(.gray)
                                Spacer()
                                Text("\(Int(record.amount))元").font(.system(size: 13))
                                statusBadge(record.status)
                            }
                            .padding(12)
                            .background(Color.white)
                            .cornerRadius(8)
                        }
                    }
                }
                .padding()
            }
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("权益监测")
        .navigationBarTitleDisplayMode(.inline)
        .task { await loadData() }
    }

    private func summaryCard(value: Int, label: String, color: Color) -> some View {
        VStack(spacing: 8) {
            Text("\(value)").font(.system(size: 32, weight: .bold)).foregroundColor(color)
            Text(label).font(.system(size: 14)).foregroundColor(.gray)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color.white)
        .cornerRadius(16)
    }

    private func severityBadge(_ severity: String) -> some View {
        let text: String
        let bg: Color
        if severity == "high" { text = "高危"; bg = Color.red.opacity(0.1) }
        else if severity == "medium" { text = "中等"; bg = Color.orange.opacity(0.1) }
        else { text = "低危"; bg = Color.blue.opacity(0.1) }
        return Text(text).font(.system(size: 11)).padding(.horizontal, 8).padding(.vertical, 3)
            .background(bg).cornerRadius(8)
    }

    private func statusBadge(_ status: String) -> some View {
        let text: String
        let bg: Color
        let fg: Color
        if status == "paid" { text = "已缴"; bg = Color.green.opacity(0.1); fg = Color(red: 0, green: 0.6, blue: 0.4) }
        else if status == "pending" { text = "待缴"; bg = Color.orange.opacity(0.1); fg = .orange }
        else { text = "断缴"; bg = Color.red.opacity(0.1); fg = .red }
        return Text(text).font(.system(size: 11)).padding(.horizontal, 8).padding(.vertical, 3)
            .background(bg).cornerRadius(8).foregroundColor(fg)
    }

    private func loadData() async {
        let userID = appState.userInfo?.nickName ?? "default"
        guard let base = URL(string: AppConstants.apiBaseURL) else { return }

        async let payment = loadJSON(url: base.appendingPathComponent("/v1/rights/payment-status"), userID: userID) as PaymentSummary?
        async let alertsData = loadJSONArray(url: base.appendingPathComponent("/v1/rights/alerts?unread_only=false"), userID: userID) as [AlertItem]?

        let (p, a) = await (payment, alertsData)
        summary = p
        alerts = a ?? []
        isLoading = false
    }

    private func loadJSON<T: Codable>(url: URL, userID: String) async -> T? {
        var req = URLRequest(url: url)
        let token = UserDefaults.standard.string(forKey: "auth_token") ?? ""
        req.setValue("Bearer " + token, forHTTPHeaderField: "Authorization")
        guard let (data, _) = try? await URLSession.shared.data(for: req) else { return nil }
        guard let wrapper = try? JSONDecoder().decode(ResponseBox<T>.self, from: data) else { return nil }
        return wrapper.data
    }

    private func loadJSONArray<T: Codable>(url: URL, userID: String) async -> [T]? {
        var req = URLRequest(url: url)
        let token = UserDefaults.standard.string(forKey: "auth_token") ?? ""
        req.setValue("Bearer " + token, forHTTPHeaderField: "Authorization")
        guard let (data, _) = try? await URLSession.shared.data(for: req) else { return nil }
        guard let wrapper = try? JSONDecoder().decode(ResponseBox<[T]>.self, from: data) else { return nil }
        return wrapper.data
    }

    struct ResponseBox<T: Codable>: Codable {
        let code: Int
        let data: T
    }
}
