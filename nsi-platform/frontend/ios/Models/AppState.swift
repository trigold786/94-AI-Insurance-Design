import Foundation
import CoreLocation

@MainActor
class AppState: ObservableObject {
    @Published var userInfo: UserInfo? = UserInfo(nickName: "default")
    @Published var currentCity: String = "上海"
    @Published var currentCityCode: String = "310000"
    @Published var planResult: PlanResult?
    @Published var profileData: ProfileInput?
    @Published var planInput: PlanInput?

    struct UserInfo {
        let nickName: String
    }

    struct ProfileInput: Codable {
        let age: Int
        let gender: String
        let householdRegionCode: String
        let currentResidenceCode: String
        let employmentStatus: String
        let socialSecurityYears: Int
        let hasChildren: Bool
    }

    struct PlanInput: Codable {
        let age: Int
        let gender: String
        let employment: String
        let contributionYears: Int
        let currentBalance: Double
        let monthlyBudget: Double
        let localAvgSalary: Double
    }

    struct PlanResult: Codable {
        let planId: String
        let recommendedSchemes: [Scheme]
    }

    struct Scheme: Codable, Identifiable {
        var id: String { name }
        let name: String
        let baseSalary: Int
        let monthlyCost: Double
        let projectedPension: Double
        let annualSubsidy: Double
    }
}
