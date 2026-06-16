import SwiftUI

struct SandboxView: View {
    @EnvironmentObject var appState: AppState

    @State private var cityIndex = 0
    @State private var gender = "male"
    @State private var age: Double = 30
    @State private var basePercent: Double = 100
    @State private var paidYears: Double = 15
    @State private var planYears: Double = 15
    @State private var employment = "employed"
    @State private var hasLocalHukou = true

    @State private var monthlyCost: Double = 0
    @State private var monthlyPension: Double = 0
    @State private var annualSubsidy: Double = 0
    @State private var netBenefit: Double = 0
    @State private var triggers: [SimTrigger] = []

    @State private var calculating = false
    @State private var errorMessage: String?
    @State private var debounceTask: Task<Void, Never>?
    @State private var didInitialCalc = false

    private let primaryColor = Color(red: 0.1, green: 0.34, blue: 0.86)

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                inputsCard
                outputsRow
                policyBanners
                chartCard
            }
            .padding()
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("社保模拟器")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if !didInitialCalc {
                didInitialCalc = true
                await calculate()
            }
        }
        .onDisappear { debounceTask?.cancel() }
    }

    private var inputsCard: some View {
        VStack(alignment: .leading, spacing: 18) {
            Text("参数设置")
                .font(.system(size: 18, weight: .semibold))
                .foregroundColor(primaryColor)

            HStack {
                Text("城市").frame(width: 72, alignment: .leading).font(.system(size: 14))
                Picker("城市", selection: $cityIndex) {
                    ForEach(0..<AppConstants.cities.count, id: \.self) { i in
                        Text(AppConstants.cities[i].name).tag(i)
                    }
                }
                .pickerStyle(.menu)
                .onChange(of: cityIndex) { _ in scheduleCalculate() }
            }

            HStack {
                Text("性别").frame(width: 72, alignment: .leading).font(.system(size: 14))
                Picker("性别", selection: $gender) {
                    ForEach(AppConstants.genderOptions, id: \.value) { Text($0.label).tag($0.value) }
                }
                .pickerStyle(.segmented)
                .onChange(of: gender) { _ in scheduleCalculate() }
            }

            sliderRow(title: "年龄", value: $age, range: 16...70, step: 1, suffix: "岁")
            sliderRow(title: "缴费基数", value: $basePercent, range: 60...300, step: 10, suffix: "%")
            sliderRow(title: "已缴年限", value: $paidYears, range: 0...35, step: 1, suffix: "年")
            sliderRow(title: "规划年限", value: $planYears, range: 0...35, step: 1, suffix: "年")

            HStack {
                Text("就业").frame(width: 72, alignment: .leading).font(.system(size: 14))
                Picker("就业", selection: $employment) {
                    ForEach(AppConstants.employmentStatuses, id: \.value) { Text($0.label).tag($0.value) }
                }
                .pickerStyle(.menu)
                .onChange(of: employment) { _ in scheduleCalculate() }
            }

            Toggle("本地户籍", isOn: $hasLocalHukou)
                .onChange(of: hasLocalHukou) { _ in scheduleCalculate() }

            if calculating {
                HStack(spacing: 6) {
                    ProgressView().scaleEffect(0.8)
                    Text("计算中...").font(.system(size: 12)).foregroundColor(.gray)
                }
                .padding(.top, 4)
            }
            if let errorMessage = errorMessage {
                Text(errorMessage).foregroundColor(.red).font(.system(size: 12))
            }
        }
        .cardStyle()
    }

    private func sliderRow(title: String, value: Binding<Double>, range: ClosedRange<Double>, step: Double, suffix: String) -> some View {
        VStack(spacing: 6) {
            HStack {
                Text(title).font(.system(size: 14))
                Spacer()
                Text("\(Int(value.wrappedValue))\(suffix)")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(primaryColor)
            }
            Slider(value: value, in: range, step: step)
                .tint(primaryColor)
                .onChange(of: value.wrappedValue) { _ in scheduleCalculate() }
        }
    }

    private var outputsRow: some View {
        HStack(spacing: 10) {
            metricCard(title: "月缴费", value: monthlyCost, color: .orange)
            metricCard(title: "月养老金", value: monthlyPension, color: primaryColor)
            metricCard(title: "年补贴", value: annualSubsidy, color: .green)
            metricCard(title: "净收益", value: netBenefit, color: .purple)
        }
    }

    private func metricCard(title: String, value: Double, color: Color) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.system(size: 12))
                .foregroundColor(.secondary)
            Text(formatMoney(value))
                .font(.system(size: 15, weight: .bold))
                .foregroundColor(color)
                .lineLimit(1)
                .minimumScaleFactor(0.5)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(Color.white)
        .cornerRadius(12)
        .shadow(color: .black.opacity(0.04), radius: 4)
    }

    @ViewBuilder
    private var policyBanners: some View {
        if !triggers.isEmpty {
            VStack(alignment: .leading, spacing: 8) {
                Text("政策触发")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(primaryColor)
                ForEach(triggers) { t in
                    HStack(alignment: .top, spacing: 8) {
                        Text(levelIcon(t.level)).font(.system(size: 16))
                        VStack(alignment: .leading, spacing: 2) {
                            Text(t.title ?? "")
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundColor(.primary)
                            if let d = t.detail {
                                Text(d)
                                    .font(.system(size: 12))
                                    .foregroundColor(.secondary)
                            }
                        }
                        Spacer()
                    }
                    .padding(12)
                    .background(levelColor(t.level).opacity(0.12))
                    .overlay(
                        RoundedRectangle(cornerRadius: 12)
                            .stroke(levelColor(t.level).opacity(0.5), lineWidth: 1)
                    )
                    .cornerRadius(12)
                }
            }
            .cardStyle()
        }
    }

    private var chartCard: some View {
        let items: [ChartItem] = [
            ChartItem(title: "月缴费", value: monthlyCost, color: .orange),
            ChartItem(title: "月养老金", value: monthlyPension, color: primaryColor),
            ChartItem(title: "年补贴", value: annualSubsidy, color: .green),
            ChartItem(title: "净收益", value: netBenefit, color: .purple)
        ]
        let maxValue = max(items.map { abs($0.value) }.max() ?? 1, 1)

        return VStack(alignment: .leading, spacing: 12) {
            Text("收益对比")
                .font(.system(size: 18, weight: .semibold))
                .foregroundColor(primaryColor)
            GeometryReader { geo in
                HStack(alignment: .bottom, spacing: 16) {
                    ForEach(items) { item in
                        VStack(spacing: 6) {
                            Text(formatMoney(item.value))
                                .font(.system(size: 10))
                                .foregroundColor(.gray)
                            Rectangle()
                                .fill(item.color)
                                .frame(
                                    width: max((geo.size.width - 48) / 4, 8),
                                    height: max(4, CGFloat(abs(item.value) / maxValue) * max(geo.size.height - 44, 40))
                                )
                                .cornerRadius(6)
                            Text(item.title)
                                .font(.system(size: 11))
                                .foregroundColor(.secondary)
                                .lineLimit(1)
                        }
                        .frame(maxWidth: .infinity)
                    }
                }
            }
            .frame(height: 180)
        }
        .cardStyle()
    }

    private func scheduleCalculate() {
        debounceTask?.cancel()
        debounceTask = Task { @MainActor in
            try? await Task.sleep(nanoseconds: 150_000_000)
            if Task.isCancelled { return }
            await calculate()
        }
    }

    @MainActor
    private func calculate() async {
        calculating = true
        errorMessage = nil
        let body: [String: Any] = [
            "city_code": AppConstants.cities[cityIndex].code,
            "gender": gender,
            "age": Int(age),
            "base_percent": Int(basePercent),
            "paid_years": Int(paidYears),
            "plan_years": Int(planYears),
            "employment": employment,
            "has_local_hukou": hasLocalHukou
        ]
        do {
            let result: SimulatorResult? = try await SimulatorAPI.request("/v1/simulator/calculate", method: "POST", body: body)
            withAnimation(.easeInOut(duration: 0.25)) {
                monthlyCost = result?.monthlyCost ?? 0
                monthlyPension = result?.monthlyPension ?? 0
                annualSubsidy = result?.annualSubsidy ?? 0
                netBenefit = result?.netBenefit ?? 0
                triggers = result?.policyTriggers ?? []
            }
        } catch {
            errorMessage = error.localizedDescription
        }
        calculating = false
    }

    private func formatMoney(_ v: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.maximumFractionDigits = 1
        return "¥" + (formatter.string(from: NSNumber(value: v)) ?? "0")
    }

    private func levelColor(_ level: String?) -> Color {
        switch level {
        case "warning": return .orange
        case "error": return .red
        case "success": return .green
        default: return primaryColor
        }
    }

    private func levelIcon(_ level: String?) -> String {
        switch level {
        case "warning": return "⚠️"
        case "error": return "⛔"
        case "success": return "✅"
        default: return "ℹ️"
        }
    }
}

private struct SimulatorResult: Decodable {
    let monthlyCost: Double?
    let monthlyPension: Double?
    let annualSubsidy: Double?
    let netBenefit: Double?
    let policyTriggers: [SimTrigger]?
}

private struct SimTrigger: Decodable, Identifiable {
    var id: String { (title ?? "") + (detail ?? "") }
    let title: String?
    let detail: String?
    let level: String?
}

private struct ChartItem: Identifiable {
    var id: String { title }
    let title: String
    let value: Double
    let color: Color
}

private struct ApiResponse<T: Decodable>: Decodable {
    let code: Int
    let msg: String?
    let data: T?
}

private enum SimulatorAPI {
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

struct CardBackground: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(16)
            .background(Color.white)
            .cornerRadius(16)
            .shadow(color: Color.black.opacity(0.05), radius: 6, x: 0, y: 2)
    }
}

extension View {
    func cardStyle() -> some View {
        modifier(CardBackground())
    }
}
