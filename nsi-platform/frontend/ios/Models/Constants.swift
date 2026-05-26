import Foundation

struct AppConstants {
    static let cities: [(code: String, name: String)] = [
        ("310000", "上海"),
        ("110000", "北京"),
        ("440300", "深圳"),
        ("440100", "广州"),
        ("330100", "杭州"),
    ]

    static let employmentStatuses: [(value: String, label: String)] = [
        ("employed", "企业就业"),
        ("flexible", "灵活就业"),
        ("self_employed", "自雇"),
        ("unemployed", "失业"),
    ]

    static let genderOptions: [(value: String, label: String)] = [
        ("male", "男"),
        ("female", "女"),
    ]

    static let apiBaseURL = "http://127.0.0.1:39401"
}
