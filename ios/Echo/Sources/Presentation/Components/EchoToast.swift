#if os(iOS)
import SwiftUI

// MARK: - Toast Style

enum EchoToastStyle {
    case success
    case error
    case info
    case copied

    var icon: String {
        switch self {
        case .success: return "checkmark.circle.fill"
        case .error:   return "exclamationmark.circle.fill"
        case .info:    return "info.circle.fill"
        case .copied:  return "doc.on.doc.fill"
        }
    }

    var backgroundColor: Color {
        switch self {
        case .success: return Color.echoTrustGreen
        case .error:   return Color.echoAlert
        case .info:    return Color.echoSignal
        case .copied:  return Color.echoInk
        }
    }
}

// MARK: - Toast Item

struct ToastItem: Identifiable, Equatable {
    let id = UUID()
    let message: String
    let style: EchoToastStyle
    let duration: TimeInterval

    static func == (lhs: ToastItem, rhs: ToastItem) -> Bool {
        lhs.id == rhs.id
    }
}

// MARK: - Toast Manager

@Observable
final class ToastManager {
    static let shared = ToastManager()

    private(set) var current: ToastItem?
    private var dismissTask: Task<Void, Never>?

    private init() {}

    @MainActor
    func show(_ message: String, style: EchoToastStyle = .info, duration: TimeInterval = 2.0) {
        dismissTask?.cancel()
        withAnimation(.spring(response: 0.35, dampingFraction: 0.85)) {
            current = ToastItem(message: message, style: style, duration: duration)
        }
        dismissTask = Task { @MainActor in
            try? await Task.sleep(for: .seconds(duration))
            guard !Task.isCancelled else { return }
            withAnimation(.spring(response: 0.3, dampingFraction: 0.9)) {
                current = nil
            }
        }
    }

    @MainActor
    func dismiss() {
        dismissTask?.cancel()
        withAnimation(.spring(response: 0.3, dampingFraction: 0.9)) {
            current = nil
        }
    }
}

// MARK: - Toast View

struct EchoToastView: View {
    let item: ToastItem

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: item.style.icon)
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(.white)

            Text(item.message)
                .font(.system(size: 14, weight: .medium))
                .foregroundColor(.white)
                .lineLimit(2)

            Spacer(minLength: 0)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(item.style.backgroundColor)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .glacialShadow(radius: 12, opacity: 0.15)
        .padding(.horizontal, 16)
    }
}

// MARK: - Toast Overlay Modifier

struct EchoToastOverlay: ViewModifier {
    @State private var manager = ToastManager.shared

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .top) {
                if let toast = manager.current {
                    EchoToastView(item: toast)
                        .padding(.top, 8)
                        .transition(.move(edge: .top).combined(with: .opacity))
                        .onTapGesture {
                            manager.dismiss()
                        }
                        .zIndex(999)
                }
            }
    }
}

extension View {
    func echoToastOverlay() -> some View {
        modifier(EchoToastOverlay())
    }
}

#endif
