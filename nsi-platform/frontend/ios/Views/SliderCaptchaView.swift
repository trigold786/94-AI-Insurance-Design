import SwiftUI

struct SliderCaptchaView: View {
    var onVerified: () -> Void

    @State private var offset: CGFloat = 0
    @State private var verified = false

    private let knobSize: CGFloat = 52
    private let primaryColor = Color(red: 0.1, green: 0.34, blue: 0.86)

    var body: some View {
        GeometryReader { geo in
            let maxWidth = max(geo.size.width - knobSize, 1)
            ZStack(alignment: .leading) {
                RoundedRectangle(cornerRadius: knobSize / 2)
                    .fill(verified ? Color.green.opacity(0.15) : Color.gray.opacity(0.12))

                captchaLabel(progress: Double(offset / maxWidth))

                ZStack {
                    Circle()
                        .fill(verified ? Color.green : (offset >= maxWidth - 1 ? Color.green : primaryColor))
                    Image(systemName: verified ? "checkmark" : "arrow.right")
                        .font(.system(size: 18, weight: .bold))
                        .foregroundColor(.white)
                }
                .frame(width: knobSize, height: knobSize)
                .shadow(color: .black.opacity(0.15), radius: 3, y: 1)
                .offset(x: offset)
                .gesture(
                    DragGesture(minimumDistance: 0)
                        .onChanged { value in
                            guard !verified else { return }
                            offset = max(0, min(maxWidth, value.translation.width))
                        }
                        .onEnded { _ in
                            guard !verified else { return }
                            if offset >= maxWidth - 2 {
                                withAnimation(.spring()) { offset = maxWidth }
                                verified = true
                                onVerified()
                            } else {
                                withAnimation(.spring()) { offset = 0 }
                            }
                        }
                )
            }
            .frame(height: knobSize)
        }
        .frame(height: knobSize)
    }

    private func captchaLabel(progress: Double) -> some View {
        ZStack {
            if verified {
                Text("✓ 验证成功")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.green)
            } else {
                Text("请拖动滑块验证")
                    .font(.system(size: 14))
                    .foregroundColor(.gray)
                    .opacity(1 - min(max(progress, 0), 1))
            }
        }
        .frame(maxWidth: .infinity)
    }
}
