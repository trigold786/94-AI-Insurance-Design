import SwiftUI

struct FeedbackView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var category = "feature"
    @State private var content = ""
    @State private var contact = ""
    @State private var submitting = false
    @State private var errorMessage: String?
    @State private var submitted = false

    private let primaryColor = Color(red: 0.1, green: 0.34, blue: 0.86)

    private let categories: [(value: String, label: String)] = [
        ("feature", "功能建议"), ("bug", "问题报告"), ("other", "其他")
    ]

    var body: some View {
        Form {
            Section(header: Text("反馈类型")) {
                Picker("分类", selection: $category) {
                    ForEach(categories, id: \.value) { Text($0.label).tag($0.value) }
                }
            }

            Section(header: Text("内容")) {
                ZStack(alignment: .topLeading) {
                    TextEditor(text: $content)
                        .frame(minHeight: 140)
                    if content.isEmpty {
                        Text("请描述您的建议或遇到的问题...")
                            .font(.system(size: 16))
                            .foregroundColor(Color.gray.opacity(0.5))
                            .padding(.top, 8)
                            .padding(.leading, 4)
                            .allowsHitTesting(false)
                    }
                }
            }

            Section(header: Text("联系方式（选填）")) {
                TextField("手机号 / 邮箱", text: $contact)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)
            }

            if let errorMessage = errorMessage {
                Section {
                    Text(errorMessage).foregroundColor(.red).font(.system(size: 13))
                }
            }

            Section {
                Button(action: submit) {
                    HStack {
                        Spacer()
                        Text(submitting ? "提交中..." : "提交反馈")
                            .font(.system(size: 17, weight: .semibold))
                            .foregroundColor(.white)
                        Spacer()
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 6)
                }
                .listRowBackground(
                    LinearGradient(colors: [primaryColor, Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing)
                )
                .disabled(content.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || submitting)
            }
        }
        .navigationTitle("意见反馈")
        .navigationBarTitleDisplayMode(.inline)
        .alert("已提交", isPresented: $submitted) {
            Button("好的") { dismiss() }
        } message: {
            Text("感谢您的反馈，我们会尽快处理。")
        }
    }

    private func submit() {
        let body: [String: Any] = [
            "category": category,
            "content": content.trimmingCharacters(in: .whitespacesAndNewlines),
            "contact": contact.trimmingCharacters(in: .whitespaces)
        ]
        submitting = true
        errorMessage = nil
        Task { @MainActor in
            do {
                let _: FeedbackData? = try await FeedbackAPI.request("/v1/feedback", method: "POST", body: body)
                submitting = false
                submitted = true
            } catch {
                submitting = false
                errorMessage = error.localizedDescription
            }
        }
    }
}

private struct FeedbackData: Decodable {}

private struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String?
    let data: T?
}

private enum FeedbackAPI {
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
