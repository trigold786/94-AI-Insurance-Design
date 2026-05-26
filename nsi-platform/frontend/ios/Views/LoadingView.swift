import SwiftUI
import NSIAPI

struct LoadingView: View {
    @EnvironmentObject var appState: AppState
    @State private var progressText = "正在分析您的社保情况..."
    @State private var navigateToPreview = false

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            ProgressView()
                .scaleEffect(2)
                .tint(Color(red: 0.1, green: 0.34, blue: 0.86))

            Text("AI 正在为您匹配最优政策...")
                .font(.system(size: 18))
                .foregroundColor(.primary)

            Text(progressText)
                .font(.system(size: 14))
                .foregroundColor(.gray)

            Spacer()
        }
        .navigationBarBackButtonHidden(true)
        .navigationDestination(isPresented: $navigateToPreview) {
            PreviewView()
        }
        .task {
            guard let input = appState.planInput else { return }
            progressText = "正在计算最优方案..."

            let userID = appState.userInfo?.nickName ?? "default"
            guard let client = try? NSIClient(baseURL: AppConstants.apiBaseURL, userID: userID) else { return }

            do {
                let jsonData = try JSONEncoder().encode(input)
                let jsonString = String(data: jsonData, encoding: .utf8) ?? ""
                let plan = try await client.generatePlan(json: jsonString)
                appState.planResult = AppState.PlanResult(
                    planId: plan.planID ?? "",
                    recommendedSchemes: (plan.recommendedSchemes ?? []).map { s in
                        AppState.Scheme(
                            name: s.name ?? "",
                            baseSalary: s.baseSalary ?? 0,
                            monthlyCost: s.monthlyCost ?? 0,
                            projectedPension: s.projectedPension ?? 0,
                            annualSubsidy: s.annualSubsidy ?? 0
                        )
                    }
                )
                try? await Task.sleep(nanoseconds: 800_000_000)
                navigateToPreview = true
            } catch {
                progressText = "生成失败，请重试"
                try? await Task.sleep(nanoseconds: 2_000_000_000)
            }
        }
    }
}
