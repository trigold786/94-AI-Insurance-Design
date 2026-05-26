import SwiftUI

struct CityPickerView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            List(AppConstants.cities, id: \.code) { city in
                Button(action: {
                    appState.currentCity = city.name
                    appState.currentCityCode = city.code
                    dismiss()
                }) {
                    HStack {
                        Text(city.name)
                            .font(.system(size: 18))
                            .foregroundColor(appState.currentCityCode == city.code ? Color(red: 0.1, green: 0.34, blue: 0.86) : .primary)
                        Spacer()
                        if appState.currentCityCode == city.code {
                            Image(systemName: "checkmark")
                                .foregroundColor(Color(red: 0.1, green: 0.34, blue: 0.86))
                        }
                    }
                    .padding(.vertical, 8)
                }
            }
            .navigationTitle("选择城市")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}
