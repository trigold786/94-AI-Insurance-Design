import SwiftUI
import NSIAPI
import CoreLocation

struct HomeView: View {
    @EnvironmentObject var appState: AppState
    @State private var policies: [PolicyClaim] = []
    @State private var showCityPicker = false
    @State private var navigateToProfile = false
    @State private var navigateToCompliance = false
    @State private var navigateToRights = false
    @State private var locationDetector = LocationDetector()
    @State private var locationRequested = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                Button(action: { showCityPicker = true }) {
                    HStack(spacing: 6) {
                        Text("📍")
                            .font(.system(size: 18))
                        Text(appState.currentCity)
                            .font(.system(size: 18, weight: .semibold))
                            .foregroundColor(.primary)
                        Text("▼")
                            .font(.system(size: 14))
                            .foregroundColor(.gray)
                    }
                }

                VStack(spacing: 12) {
                    Text("社保规划，智能匹配")
                        .font(.system(size: 26, weight: .bold))
                        .foregroundColor(.white)
                    Text("AI 为您匹配最优社保方案")
                        .font(.system(size: 18))
                        .foregroundColor(.white.opacity(0.8))
                    Button(action: { navigateToProfile = true }) {
                        Text("开始社保筹划")
                            .font(.system(size: 18, weight: .semibold))
                            .foregroundColor(.white)
                            .frame(maxWidth: .infinity)
                            .frame(height: 48)
                            .background(Color.white.opacity(0.2))
                            .cornerRadius(48)
                    }
                }
                .padding(40)
                .frame(maxWidth: .infinity)
                .background(
                    LinearGradient(colors: [Color(red: 0.1, green: 0.34, blue: 0.86), Color(red: 0.23, green: 0.51, blue: 0.96)], startPoint: .topLeading, endPoint: .bottomTrailing)
                )
                .cornerRadius(24)

                Button(action: { navigateToCompliance = true }) {
                    HStack {
                        Text("📋").font(.system(size: 20))
                        Text("合规认定与材料清单")
                            .font(.system(size: 16, weight: .medium))
                            .foregroundColor(.primary)
                        Spacer()
                        Image(systemName: "chevron.right").font(.system(size: 14)).foregroundColor(.gray)
                    }
                    .padding()
                    .background(Color.white)
                    .cornerRadius(16)
                    .shadow(color: .black.opacity(0.04), radius: 4)
                }

                Button(action: { navigateToRights = true }) {
                    HStack {
                        Text("🛡️").font(.system(size: 20))
                        Text("权益监测与风险预警")
                            .font(.system(size: 16, weight: .medium))
                            .foregroundColor(.primary)
                        Spacer()
                        Image(systemName: "chevron.right").font(.system(size: 14)).foregroundColor(.gray)
                    }
                    .padding()
                    .background(Color.white)
                    .cornerRadius(16)
                    .shadow(color: .black.opacity(0.04), radius: 4)
                }

                Text("政策速览")
                    .font(.system(size: 20, weight: .semibold))

                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(policies, id: \.claimID) { policy in
                            VStack(alignment: .leading, spacing: 8) {
                                Text(policy.policyType ?? "-")
                                    .font(.system(size: 16, weight: .semibold))
                                    .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                                Text(policy.subsidyCalcMethod ?? "-")
                                    .font(.system(size: 14))
                                    .foregroundColor(.gray)
                            }
                            .padding()
                            .frame(width: 200)
                            .background(Color.white)
                            .cornerRadius(16)
                            .shadow(color: .black.opacity(0.06), radius: 4)
                        }
                    }
                }
            }
            .padding()
        }
        .background(Color(red: 0.96, green: 0.97, blue: 0.98))
        .navigationBarHidden(true)
        .sheet(isPresented: $showCityPicker) {
            CityPickerView()
        }
        .navigationDestination(isPresented: $navigateToProfile) {
            ProfileFormView()
        }
        .navigationDestination(isPresented: $navigateToCompliance) {
            ComplianceView()
        }
        .navigationDestination(isPresented: $navigateToRights) {
            RightsView()
        }
        .task {
            if !locationRequested {
                locationRequested = true
                requestLocation()
            }
            loadPolicies()
        }
    }

    private func requestLocation() {
        let state = appState
        locationDetector.onLocation = { coord in
            let city = mapLocationToCity(lat: coord.latitude, lng: coord.longitude)
            state.currentCity = city.name
            state.currentCityCode = city.code
        }
        locationDetector.request()
    }

    private func loadPolicies() {
        let userID = appState.userInfo?.nickName ?? "default"
        guard let client = try? NSIClient(baseURL: AppConstants.apiBaseURL, userID: userID) else { return }
        Task {
            let claims = try? await client.queryPolicies(regionCode: appState.currentCityCode)
            policies = claims ?? []
        }
    }
}

struct PolicyClaim: Codable {
    let claimID: String?
    let policyType: String?
    let subsidyCalcMethod: String?
}

class LocationDetector: NSObject, CLLocationManagerDelegate {
    private let manager = CLLocationManager()
    var onLocation: ((CLLocationCoordinate2D) -> Void)?

    override init() {
        super.init()
        manager.delegate = self
        manager.desiredAccuracy = kCLLocationAccuracyKilometer
    }

    func request() {
        manager.requestWhenInUseAuthorization()
        manager.requestLocation()
    }

    func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        if let coord = locations.last?.coordinate {
            onLocation?(coord)
        }
    }

    func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {}
}

private func mapLocationToCity(lat: Double, lng: Double) -> (code: String, name: String) {
    let cities: [(lat: Double, lng: Double, code: String, name: String)] = [
        (31.23, 121.47, "310000", "上海"),
        (39.90, 116.40, "110000", "北京"),
        (22.54, 114.06, "440300", "深圳"),
        (23.13, 113.27, "440100", "广州"),
        (30.27, 120.15, "330100", "杭州"),
    ]
    var best = cities[0]
    var bestDist = Double.greatestFiniteMagnitude
    for c in cities {
        let d = (c.lat - lat) * (c.lat - lat) + (c.lng - lng) * (c.lng - lng)
        if d < bestDist { bestDist = d; best = c }
    }
    return (best.code, best.name)
}
