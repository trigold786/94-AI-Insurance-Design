import SwiftUI

struct PreviewView: View {
    @EnvironmentObject var appState: AppState
    @State private var navigateToPlan = false

    var body: some View {
        VStack(spacing: 24) {
            Text("您的社保方案")
                .font(.system(size: 28, weight: .bold))
                .padding(.top, 40)

            Text("以下是 AI 为您推荐的社保方案预览")
                .font(.system(size: 16))
                .foregroundColor(.gray)

            VStack(spacing: 16) {
                previewCard(title: "预计年补贴额", value: "约 ¥XX,XXX/年")
                previewCard(title: "月节省额", value: "约 ¥XXX/月")
            }
            .padding()

            Button(action: { navigateToPlan = true }) {
                Text("解锁完整报告")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: 48)
                    .background(
                        LinearGradient(colors: [Color(red: 0.1, green: 0.34, blue: 0.86), Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing)
                    )
                    .cornerRadius(48)
            }
            .padding(.horizontal)

            Text("解锁后可查看方案对比、年度现金流和行动清单")
                .font(.system(size: 13))
                .foregroundColor(.gray)
                .multilineTextAlignment(.center)
                .padding(.horizontal)

            Spacer()
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationBarBackButtonHidden(true)
        .navigationDestination(isPresented: $navigateToPlan) {
            if let planId = appState.planResult?.planId {
                PlanDetailView(planId: planId)
            }
        }
    }

    private func previewCard(title: String, value: String) -> some View {
        VStack(spacing: 8) {
            Text(title)
                .font(.system(size: 16))
                .foregroundColor(.gray)
            Text(value)
                .font(.system(size: 32, weight: .bold))
                .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                .blur(radius: 8)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color.white)
        .cornerRadius(16)
    }
}
