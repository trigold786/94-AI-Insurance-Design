import SwiftUI

struct ChatMessage: Identifiable {
    let id = UUID()
    let role: String
    let text: String
    let timestamp: Date
}

struct AdvisorView: View {
    @EnvironmentObject var appState: AppState

    @State private var messages: [ChatMessage] = []
    @State private var input = ""
    @State private var sending = false
    @State private var errorMessage: String?

    private let primaryColor = Color(red: 0.1, green: 0.34, blue: 0.86)

    var body: some View {
        VStack(spacing: 0) {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        if messages.isEmpty {
                            emptyHint
                        }
                        ForEach(messages) { msg in
                            messageBubble(msg).id(msg.id)
                        }
                    }
                    .padding()
                }
                .onChange(of: messages.count) { _ in
                    if let last = messages.last {
                        withAnimation(.easeOut(duration: 0.2)) {
                            proxy.scrollTo(last.id, anchor: .bottom)
                        }
                    }
                }
            }

            inputBar
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("AI 顾问")
        .navigationBarTitleDisplayMode(.inline)
    }

    private var emptyHint: some View {
        VStack(spacing: 12) {
            Text("💬").font(.system(size: 44))
            Text("向 AI 顾问提问社保相关问题")
                .font(.system(size: 15))
                .foregroundColor(.gray)
            Text("例如：灵活就业如何领取养老金？")
                .font(.system(size: 13))
                .foregroundColor(.gray.opacity(0.8))
        }
        .frame(maxWidth: .infinity)
        .padding(.top, 60)
    }

    private func messageBubble(_ msg: ChatMessage) -> some View {
        let isUser = msg.role == "user"
        return HStack {
            if isUser { Spacer(minLength: 40) }
            Text(msg.text)
                .font(.system(size: 15))
                .foregroundColor(isUser ? .white : .primary)
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(isUser ? primaryColor : Color.white)
                .cornerRadius(16)
                .shadow(color: .black.opacity(0.04), radius: 3)
            if !isUser { Spacer(minLength: 40) }
        }
    }

    private var inputBar: some View {
        let canSend = !input.trimmingCharacters(in: .whitespaces).isEmpty && !sending
        return HStack(spacing: 8) {
            TextField("输入您的问题...", text: $input)
                .font(.system(size: 15))
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(Color.white)
                .cornerRadius(20)
                .shadow(color: .black.opacity(0.04), radius: 3)
                .submitLabel(.send)
                .onSubmit { if canSend { send() } }

            Button(action: send) {
                Image(systemName: sending ? "ellipsis" : "paperplane.fill")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(width: 44, height: 44)
                    .background(canSend ? primaryColor : Color.gray)
                    .clipShape(Circle())
            }
            .disabled(!canSend)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color.white)
        .overlay(alignment: .top) {
            Rectangle()
                .fill(Color.gray.opacity(0.15))
                .frame(height: 0.5)
        }
    }

    private func send() {
        let question = input.trimmingCharacters(in: .whitespaces)
        guard !question.isEmpty else { return }
        messages.append(ChatMessage(role: "user", text: question, timestamp: Date()))
        input = ""
        sending = true
        errorMessage = nil

        Task { @MainActor in
            do {
                let body: [String: Any] = ["question": question]
                let resp: AdvisorData? = try await AdvisorAPI.request("/v1/advisor/ask", method: "POST", body: body)
                let answer = resp?.answer?.isEmpty == false
                    ? resp!.answer!
                    : "抱歉，我暂时无法回答这个问题。"
                messages.append(ChatMessage(role: "assistant", text: answer, timestamp: Date()))
            } catch {
                errorMessage = error.localizedDescription
                messages.append(ChatMessage(role: "assistant", text: "请求失败：\(error.localizedDescription)", timestamp: Date()))
            }
            sending = false
        }
    }
}

private struct AdvisorData: Decodable {
    let answer: String?
}

private struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String?
    let data: T?
}

private enum AdvisorAPI {
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
