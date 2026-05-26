import SwiftUI
import NSIAPI

struct PlanDetailView: View {
    @EnvironmentObject var appState: AppState
    let planId: String

    @State private var schemes: [AppState.Scheme] = []
    @State private var currentSchemeIndex = 0

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                Text("方案详情")
                    .font(.system(size: 28, weight: .bold))

                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(Array(schemes.enumerated()), id: \.offset) { index, scheme in
                            schemeCard(scheme: scheme, isActive: index == currentSchemeIndex)
                                .onTapGesture { currentSchemeIndex = index }
                        }
                    }
                }

                if schemes.indices.contains(currentSchemeIndex) {
                    let s = schemes[currentSchemeIndex]
                    VStack(spacing: 16) {
                        detailRow(label: "缴费基数", value: "\(s.baseSalary)元")
                        detailRow(label: "月缴费", value: "\(Int(s.monthlyCost))元")
                        detailRow(label: "企业年补贴", value: "\(Int(s.annualSubsidy))元/年", highlight: true)
                        detailRow(label: "预计养老金", value: "\(Int(s.projectedPension))元/月", highlight: true)
                    }
                    .padding()
                    .background(Color.white)
                    .cornerRadius(16)
                }

                VStack(spacing: 12) {
                    Button("查看完整报告") {
                        if let url = URL(string: "\(AppConstants.apiBaseURL)/v1/plans/report?plan_id=\(planId)") {
                            var req = URLRequest(url: url)
                            req.setValue(appState.userInfo?.nickName ?? "default", forHTTPHeaderField: "x-user-id")
                            UIApplication.shared.open(url)
                        }
                    }
                    .buttonStyle(PrimaryButtonStyle())

                    Button("合规检查清单") {
                    }
                    .buttonStyle(SecondaryButtonStyle())
                }
            }
            .padding()
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("方案详情")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            loadSchemes()
        }
    }

    private func schemeCard(scheme: AppState.Scheme, isActive: Bool) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(scheme.name)
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
            Text("预计养老金")
                .font(.system(size: 13))
                .foregroundColor(.gray)
            Text("\(Int(scheme.projectedPension))元/月")
                .font(.system(size: 24, weight: .bold))
                .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
            Text("月缴费: \(Int(scheme.monthlyCost))元")
                .font(.system(size: 13))
                .foregroundColor(.gray)
        }
        .padding()
        .frame(width: 200)
        .background(Color.white)
        .cornerRadius(16)
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(isActive ? Color(red: 0.1, green: 0.34, blue: 0.86) : Color.clear, lineWidth: 2)
        )
    }

    private func detailRow(label: String, value: String, highlight: Bool = false) -> some View {
        HStack {
            Text(label)
                .font(.system(size: 16))
                .foregroundColor(.gray)
            Spacer()
            Text(value)
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(highlight ? Color(red: 0, green: 0.6, blue: 0.4) : .primary)
        }
    }

    private func loadSchemes() {
        if let result = appState.planResult {
            schemes = result.recommendedSchemes
            return
        }
        let userID = appState.userInfo?.nickName ?? "default"
        guard let client = try? NSIClient(baseURL: AppConstants.apiBaseURL, userID: userID) else { return }
        Task {
            let plan = try? await client.getPlanDetail(planID: planId)
            if let s = plan?.recommendedSchemes {
                schemes = s.map { ss in
                    AppState.Scheme(
                        name: ss.name ?? "",
                        baseSalary: ss.baseSalary ?? 0,
                        monthlyCost: ss.monthlyCost ?? 0,
                        projectedPension: ss.projectedPension ?? 0,
                        annualSubsidy: ss.annualSubsidy ?? 0
                    )
                }
            }
        }
    }
}
