import SwiftUI
import NSIAPI

struct ProfileFormView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var currentStep = 1
    @State private var age = ""
    @State private var gender = "male"
    @State private var household = ""
    @State private var employment = "flexible"
    @State private var years = ""
    @State private var hasChildren = false
    @State private var budget = ""
    @State private var balance = ""
    @State private var submitting = false
    @State private var errorMessage = ""
    @State private var navigateToLoading = false

    var body: some View {
        VStack(spacing: 24) {
            HStack(spacing: 8) {
                stepView(number: 1, active: currentStep >= 1)
                stepLine(active: currentStep >= 2)
                stepView(number: 2, active: currentStep >= 2)
                stepLine(active: currentStep >= 3)
                stepView(number: 3, active: currentStep >= 3)
            }

            switch currentStep {
            case 1: step1View
            case 2: step2View
            case 3: step3View
            default: EmptyView()
            }

            if !errorMessage.isEmpty {
                Text(errorMessage)
                    .foregroundColor(.red)
                    .font(.system(size: 14))
            }

            Spacer()
        }
        .padding()
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("信息采集")
        .navigationBarTitleDisplayMode(.inline)
        .navigationDestination(isPresented: $navigateToLoading) {
            LoadingView()
        }
    }

    @ViewBuilder
    private var step1View: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("基本信息").font(.system(size: 24, weight: .bold))
            Picker("性别", selection: $gender) {
                ForEach(AppConstants.genderOptions, id: \.value) { opt in
                    Text(opt.label).tag(opt.value)
                }
            }
            .pickerStyle(.menu)
            HStack {
                Text("年龄")
                Spacer()
                TextField("请输入年龄(16-70)", text: $age)
                    .keyboardType(.numberPad)
                    .multilineTextAlignment(.trailing)
            }
            HStack {
                Text("户籍地")
                Spacer()
                TextField("请选择", text: $household)
                    .multilineTextAlignment(.trailing)
            }
            Button("下一步") { currentStep = 2 }
                .buttonStyle(PrimaryButtonStyle())
        }
    }

    @ViewBuilder
    private var step2View: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("就业信息").font(.system(size: 24, weight: .bold))
            Picker("就业状态", selection: $employment) {
                ForEach(AppConstants.employmentStatuses, id: \.value) { opt in
                    Text(opt.label).tag(opt.value)
                }
            }
            .pickerStyle(.menu)
            HStack {
                Text("累计社保年限")
                Spacer()
                TextField("请输入年限(年)", text: $years)
                    .keyboardType(.numberPad)
                    .multilineTextAlignment(.trailing)
            }
            HStack(spacing: 16) {
                Button("上一步") { currentStep = 1 }
                    .buttonStyle(SecondaryButtonStyle())
                Button("下一步") { currentStep = 3 }
                    .buttonStyle(PrimaryButtonStyle())
            }
        }
    }

    @ViewBuilder
    private var step3View: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("补充信息").font(.system(size: 24, weight: .bold))
            Toggle("子女情况", isOn: $hasChildren)
            HStack {
                Text("月预算（元）")
                Spacer()
                TextField("请输入月缴费预算", text: $budget)
                    .keyboardType(.decimalPad)
                    .multilineTextAlignment(.trailing)
            }
            HStack {
                Text("当前账户余额（元）")
                Spacer()
                TextField("请输入当前账户余额", text: $balance)
                    .keyboardType(.decimalPad)
                    .multilineTextAlignment(.trailing)
            }
            HStack(spacing: 16) {
                Button("上一步") { currentStep = 2 }
                    .buttonStyle(SecondaryButtonStyle())
                Button(submitting ? "提交中..." : "生成方案") { onSubmit() }
                    .buttonStyle(PrimaryButtonStyle())
                    .disabled(submitting)
            }
        }
    }

    private func stepView(number: Int, active: Bool) -> some View {
        ZStack {
            Circle()
                .fill(active ? Color(red: 0.1, green: 0.34, blue: 0.86) : Color.gray.opacity(0.3))
                .frame(width: 36, height: 36)
            Text("\(number)")
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(active ? .white : .gray)
        }
    }

    private func stepLine(active: Bool) -> some View {
        Rectangle()
            .fill(active ? Color(red: 0.1, green: 0.34, blue: 0.86) : Color.gray.opacity(0.3))
            .frame(width: 40, height: 3)
    }

    private func onSubmit() {
        guard let ageInt = Int(age), (16...70).contains(ageInt) else {
            errorMessage = "年龄必须在16-70之间"
            return
        }
        guard let budgetDouble = Double(budget), budgetDouble > 0 else {
            errorMessage = "请输入有效月预算"
            return
        }
        submitting = true
        errorMessage = ""

        let profileData = AppState.ProfileInput(
            age: ageInt,
            gender: gender,
            householdRegionCode: appState.currentCityCode,
            currentResidenceCode: appState.currentCityCode,
            employmentStatus: employment,
            socialSecurityYears: Int(years) ?? 0,
            hasChildren: hasChildren
        )
        appState.profileData = profileData
        appState.planInput = AppState.PlanInput(
            age: ageInt,
            gender: gender,
            employment: employment,
            contributionYears: Int(years) ?? 0,
            currentBalance: Double(balance) ?? 0,
            monthlyBudget: budgetDouble,
            localAvgSalary: budgetDouble * 2
        )

        let userID = appState.userInfo?.nickName ?? "default"
        guard let client = try? NSIClient(baseURL: AppConstants.apiBaseURL, userID: userID) else {
            submitting = false
            errorMessage = "客户端初始化失败"
            return
        }

        Task {
            do {
                let jsonData = try JSONEncoder().encode(profileData)
                let jsonString = String(data: jsonData, encoding: .utf8) ?? ""
                try await client.updateProfile(json: jsonString)
                navigateToLoading = true
            } catch {
                errorMessage = error.localizedDescription
                submitting = false
            }
        }
    }
}

struct PrimaryButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 18, weight: .semibold))
            .foregroundColor(.white)
            .frame(maxWidth: .infinity)
            .frame(height: 48)
            .background(
                LinearGradient(colors: [Color(red: 0.1, green: 0.34, blue: 0.86), Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing)
            )
            .cornerRadius(48)
    }
}

struct SecondaryButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 18, weight: .semibold))
            .foregroundColor(.primary)
            .frame(maxWidth: .infinity)
            .frame(height: 48)
            .background(Color.white)
            .cornerRadius(48)
            .overlay(RoundedRectangle(cornerRadius: 48).stroke(Color.gray.opacity(0.5)))
    }
}
