import SwiftUI

struct LoginView: View {
    @EnvironmentObject var appState: AppState
    @State private var agreed = false
    @State private var loading = false
    @State private var showPrivacy = false
    @State private var showTerms = false
    @State private var navigateToHome = false
    @State private var errorMessage: String?

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            VStack(spacing: 16) {
                ZStack {
                    RoundedRectangle(cornerRadius: 40)
                        .fill(LinearGradient(colors: [Color(red: 0.1, green: 0.34, blue: 0.86), Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing))
                        .frame(width: 160, height: 160)
                    Text("社保")
                        .font(.system(size: 48, weight: .bold))
                        .foregroundColor(.white)
                }

                Text("AI社保智筹")
                    .font(.system(size: 36, weight: .bold))
                    .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))

                Text("AI驱动的社保规划助手")
                    .font(.system(size: 18))
                    .foregroundColor(.gray)
            }

            Spacer().frame(height: 60)

            HStack(spacing: 8) {
                Toggle("", isOn: $agreed)
                    .labelsHidden()
                    .scaleEffect(0.8)
                Text("我已阅读并同意")
                    .font(.system(size: 14))
                    .foregroundColor(.gray)
                Button("《隐私政策》") { showPrivacy = true }
                    .font(.system(size: 14))
                    .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                Text("和")
                    .font(.system(size: 14))
                    .foregroundColor(.gray)
                Button("《用户协议》") { showTerms = true }
                    .font(.system(size: 14))
                    .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
            }
            .padding()
            .background(Color.white)
            .cornerRadius(16)
            .padding(.horizontal)

            Spacer().frame(height: 32)

            Button(action: onLogin) {
                ZStack {
                    RoundedRectangle(cornerRadius: 48)
                        .fill(LinearGradient(colors: [Color(red: 0.1, green: 0.34, blue: 0.86), Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing))
                        .frame(height: 52)
                    Text(loading ? "登录中..." : "微信一键登录")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.white)
                }
            }
            .disabled(!agreed)
            .opacity(agreed ? 1 : 0.5)
            .padding(.horizontal)

            Spacer()

            Text("登录即表示您同意上述条款")
                .font(.system(size: 13))
                .foregroundColor(.gray)
                .padding(.bottom, 40)
        }
        .padding(.horizontal, 40)
        .background(
            LinearGradient(colors: [Color(red: 0.93, green: 0.95, blue: 1), Color(red: 0.96, green: 0.97, blue: 0.98)], startPoint: .top, endPoint: .bottom)
                .ignoresSafeArea()
        )
        .alert("隐私政策", isPresented: $showPrivacy) {
            Button("确定", role: .cancel) {}
        } message: {
            Text("我们收集您的信息用于社保规划服务...")
        }
        .alert("用户协议", isPresented: $showTerms) {
            Button("确定", role: .cancel) {}
        } message: {
            Text("欢迎使用AI社保智筹...")
        }
        .alert("登录失败", isPresented: Binding(get: { errorMessage != nil }, set: { _ in errorMessage = nil })) {
            Button("确定", role: .cancel) { errorMessage = nil }
        } message: {
            Text(errorMessage ?? "")
        }
        .navigationDestination(isPresented: $navigateToHome) {
            HomeView()
        }
    }

    private func onLogin() {
        guard agreed else { return }
        loading = true
        let userID = appState.userInfo?.nickName ?? "default"
        guard let client = try? NSIClient(baseURL: AppConstants.apiBaseURL, userID: userID) else {
            loading = false
            return
        }
        Task {
            do {
                let token = try await client.getToken(userID: userID)
                UserDefaults.standard.set(token, forKey: "auth_token")
                loading = false
                navigateToHome = true
            } catch {
                errorMessage = (error as NSError).localizedDescription
                loading = false
            }
        }
    }
}
