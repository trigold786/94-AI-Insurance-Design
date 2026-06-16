import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var appState: AppState

    @State private var fontScale = "medium"
    @State private var defaultTab = "home"
    @State private var notificationsEnabled = true
    @State private var loading = false
    @State private var errorMessage: String?
    @State private var isInitializing = true
    @State private var loggedOut = false

    private let primaryColor = Color(red: 0.1, green: 0.34, blue: 0.86)

    private let fontOptions: [(value: String, label: String)] = [
        ("small", "小"), ("medium", "标准"), ("large", "大")
    ]
    private let tabOptions: [(value: String, label: String)] = [
        ("home", "首页"), ("profile", "信息采集"), ("sandbox", "模拟器"),
        ("compliance", "合规认定"), ("rights", "权益监测")
    ]

    var body: some View {
        Form {
            Section(header: Text("显示")) {
                Picker("字体大小", selection: $fontScale) {
                    ForEach(fontOptions, id: \.value) { Text($0.label).tag($0.value) }
                }
                .onChange(of: fontScale) { _ in if !isInitializing { saveSettings() } }
            }

            Section(header: Text("偏好设置")) {
                Picker("默认页面", selection: $defaultTab) {
                    ForEach(tabOptions, id: \.value) { Text($0.label).tag($0.value) }
                }
                .onChange(of: defaultTab) { _ in if !isInitializing { saveSettings() } }

                Toggle("推送通知", isOn: $notificationsEnabled)
                    .onChange(of: notificationsEnabled) { _ in if !isInitializing { saveSettings() } }
            }

            Section {
                Button(role: .destructive) { logout() } label: {
                    HStack { Spacer(); Text("退出登录").bold(); Spacer() }
                }
            }

            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage)
                        .foregroundColor(.red)
                        .font(.system(size: 13))
                }
            }
        }
        .navigationTitle("设置")
        .navigationBarTitleDisplayMode(.inline)
        .overlay {
            if loading {
                ZStack {
                    Color.black.opacity(0.05).ignoresSafeArea()
                    ProgressView()
                        .scaleEffect(1.2)
                        .tint(primaryColor)
                }
                .transition(.opacity)
            }
        }
        .animation(.easeInOut(duration: 0.2), value: loading)
        .fullScreenCover(isPresented: $loggedOut) {
            NavigationStack { LoginView().environmentObject(appState) }
        }
        .task { await loadSettings() }
    }

    private func logout() {
        UserDefaults.standard.removeObject(forKey: "auth_token")
        appState.userInfo = nil
        loggedOut = true
    }

    @MainActor
    private func loadSettings() async {
        loading = true
        errorMessage = nil
        do {
            let resp: SettingsData? = try await SettingsAPI.request("/v1/settings", method: "GET")
            if let f = resp?.fontScale { fontScale = f }
            if let t = resp?.defaultTab { defaultTab = t }
            if let n = resp?.notificationsEnabled { notificationsEnabled = n }
        } catch {
            errorMessage = error.localizedDescription
        }
        loading = false
        isInitializing = false
    }

    private func saveSettings() {
        let body: [String: Any] = [
            "font_scale": fontScale,
            "default_tab": defaultTab,
            "notifications_enabled": notificationsEnabled
        ]
        Task { @MainActor in
            do {
                let _: SettingsData? = try await SettingsAPI.request("/v1/settings", method: "POST", body: body)
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }
}

private struct SettingsData: Decodable {
    let fontScale: String?
    let defaultTab: String?
    let notificationsEnabled: Bool?
}

private struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String?
    let data: T?
}

private enum SettingsAPI {
    static func request<T: Decodable>(_ path: String, method: String, body: [String: Any]? = nil) async throws -> T? {
        guard let token = UserDefaults.standard.string(forKey: "auth_token") else {
            throw NSError(domain: "NSI", code: 401, userInfo: [NSLocalizedDescriptionKey: "请先登录"])
        }
        guard let url = URL(string: AppConstants.apiBaseURL + path) else {
            throw NSError(domain: "NSI", code: 2, userInfo: [NSLocalizedDescriptionKey: "无效地址"])
        }
        var req = URLRequest(url: url)
        req.httpMethod = method
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        req.timeoutInterval = 30
        if let body = body {
            req.httpBody = try JSONSerialization.data(withJSONObject: body)
        }
        let (data, response) = try await URLSession.shared.data(for: req)
        guard let http = response as? HTTPURLResponse else {
            throw NSError(domain: "NSI", code: 3, userInfo: [NSLocalizedDescriptionKey: "无效响应"])
        }
        if http.statusCode == 401 {
            throw NSError(domain: "NSI", code: 401, userInfo: [NSLocalizedDescriptionKey: "请先登录"])
        }
        if http.statusCode >= 400 {
            let msg = String(data: data, encoding: .utf8) ?? "请求失败"
            throw NSError(domain: "NSI", code: http.statusCode, userInfo: [NSLocalizedDescriptionKey: msg])
        }
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let wrapper = try decoder.decode(ApiResponse<T>.self, from: data)
        if wrapper.code != 0 {
            throw NSError(domain: "NSI", code: wrapper.code, userInfo: [NSLocalizedDescriptionKey: wrapper.msg ?? "服务异常"])
        }
        return wrapper.data
    }
}
