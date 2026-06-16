import SwiftUI

struct ComplianceView: View {
    @EnvironmentObject var appState: AppState
    @State private var checklist: ComplianceData?
    @State private var isLoading = true

    struct ComplianceData: Codable {
        let matchedPolicies: [PolicyCompliance]
        let requiredDocs: [RequiredDoc]
        let eligibleTags: [String]

        enum CodingKeys: String, CodingKey {
            case matchedPolicies = "matched_policies"
            case requiredDocs = "required_docs"
            case eligibleTags = "eligible_tags"
        }
    }

    struct PolicyCompliance: Codable {
        let policyType: String
        let claimID: String
        let subsidyCalcMethod: String
        let conditions: [ComplianceCondition]
        let isEligible: Bool
        let unmetConditions: [String]

        enum CodingKeys: String, CodingKey {
            case policyType = "policy_type"
            case claimID = "claim_id"
            case subsidyCalcMethod = "subsidy_calc_method"
            case conditions, isEligible = "is_eligible"
            case unmetConditions = "unmet_conditions"
        }
    }

    struct ComplianceCondition: Codable, Identifiable {
        var id: String { name }
        let name: String
        let description: String
    }

    struct RequiredDoc: Codable, Identifiable {
        var id: String { name }
        let name: String
        let description: String
        let source: String
        let optional: Bool
    }

    var body: some View {
        ScrollView {
            if isLoading {
                ProgressView().padding(.top, 80)
            } else if let data = checklist {
                VStack(alignment: .leading, spacing: 20) {
                    Text("合规认定与材料清单")
                        .font(.system(size: 24, weight: .bold))

                    if !data.eligibleTags.isEmpty {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("您的可认定身份").font(.system(size: 18, weight: .semibold))
                            LazyVGrid(columns: [GridItem(.adaptive(minimum: 100))], spacing: 8) {
                                ForEach(data.eligibleTags, id: \.self) { tag in
                                    Text(tag)
                                        .font(.system(size: 14))
                                        .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                                        .padding(.horizontal, 12)
                                        .padding(.vertical, 6)
                                        .background(Color(red: 0.93, green: 0.95, blue: 1))
                                        .cornerRadius(16)
                                }
                            }
                        }
                    }

                    Text("匹配政策 (\(data.matchedPolicies.count))")
                        .font(.system(size: 18, weight: .semibold))

                    ForEach(data.matchedPolicies) { policy in
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                Text(policy.policyType).font(.system(size: 16, weight: .semibold)).foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                                Spacer()
                                Text(policy.isEligible ? "可申请" : "条件不足")
                                    .font(.system(size: 12))
                                    .foregroundColor(policy.isEligible ? Color(red: 0, green: 0.6, blue: 0.4) : .red)
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 4)
                                    .background(policy.isEligible ? Color.green.opacity(0.1) : Color.red.opacity(0.1))
                                    .cornerRadius(12)
                            }
                            Text(policy.subsidyCalcMethod).font(.system(size: 14)).foregroundColor(.gray)

                            if !policy.conditions.isEmpty {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text("认定条件").font(.system(size: 14, weight: .semibold))
                                    ForEach(policy.conditions) { cond in
                                        VStack(alignment: .leading) {
                                            Text(cond.name).font(.system(size: 13))
                                            Text(cond.description).font(.system(size: 12)).foregroundColor(.gray)
                                        }
                                    }
                                }
                            }

                            if !policy.unmetConditions.isEmpty {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text("未满足条件").font(.system(size: 14, weight: .semibold)).foregroundColor(.red)
                                    ForEach(policy.unmetConditions, id: \.self) { u in
                                        Text("✗ \(u)").font(.system(size: 13)).foregroundColor(.red)
                                    }
                                }
                            }
                        }
                        .padding()
                        .background(Color.white)
                        .cornerRadius(12)
                        .shadow(color: .black.opacity(0.06), radius: 4)
                    }

                    if !data.requiredDocs.isEmpty {
                        Text("必需材料清单 (\(data.requiredDocs.count))")
                            .font(.system(size: 18, weight: .semibold))

                        ForEach(data.requiredDocs) { doc in
                            VStack(alignment: .leading, spacing: 4) {
                                HStack {
                                    Text(doc.name).font(.system(size: 15, weight: .semibold))
                                    Spacer()
                                    Text("来源: \(doc.source)").font(.system(size: 12)).foregroundColor(.gray)
                                }
                                Text(doc.description).font(.system(size: 13)).foregroundColor(.gray)
                                if doc.optional {
                                    Text("非必需材料").font(.system(size: 12)).foregroundColor(.orange)
                                }
                            }
                            .padding()
                            .background(Color.white)
                            .cornerRadius(12)
                        }
                    }

                    Button("查看办理指南") {
                        if let url = URL(string: "\(AppConstants.apiBaseURL)/v1/guide?city_code=\(appState.currentCityCode)") {
                            UIApplication.shared.open(url)
                        }
                    }
                    .buttonStyle(SecondaryButtonStyle())
                }
                .padding()
            }
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationTitle("合规认定")
        .navigationBarTitleDisplayMode(.inline)
        .task { await loadCompliance() }
    }

    private func loadCompliance() async {
        let cityCode = appState.currentCityCode
        let userID = appState.userInfo?.nickName ?? "default"
        guard let url = URL(string: "\(AppConstants.apiBaseURL)/v1/compliance/checklist?city_code=\(cityCode)") else { return }

        var req = URLRequest(url: url)
        req.setValue(userID, forHTTPHeaderField: "x-user-id")

        guard let (data, _) = try? await URLSession.shared.data(for: req) else { return }
        guard let wrapper = try? JSONDecoder().decode(ResponseWrapper.self, from: data) else { return }
        checklist = wrapper.data
        isLoading = false
    }

    struct ResponseWrapper: Codable {
        let code: Int
        let data: ComplianceData
    }
}
