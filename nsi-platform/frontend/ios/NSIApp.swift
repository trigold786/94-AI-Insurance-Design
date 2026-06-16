import SwiftUI

@main
struct NSIApp: App {
    @StateObject private var appState = AppState()

    var body: some Scene {
        WindowGroup {
            NavigationStack {
                if UserDefaults.standard.string(forKey: "auth_token") != nil {
                    HomeView()
                        .navigationDestination(for: String.self) { route in
                            switch route {
                            case "profile": ProfileFormView()
                            case "loading": LoadingView()
                            case "preview": PreviewView()
                            case "compliance": ComplianceView()
                            case "rights": RightsView()
                            case "settings": SettingsView()
                            case "sandbox": SandboxView()
                            case "advisor": AdvisorView()
                            case "feedback": FeedbackView()
                            default: EmptyView()
                            }
                        }
                } else {
                    LoginView()
                }
            }
            .environmentObject(appState)
        }
    }
}
