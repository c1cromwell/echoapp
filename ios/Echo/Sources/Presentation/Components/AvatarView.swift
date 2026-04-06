import SwiftUI

import SwiftUI

/// Avatar size enumeration
public enum AvatarSize {
    case xs      // 24pt
    case sm      // 32pt
    case md      // 40pt
    case lg      // 56pt
    case xl      // 80pt
    case xxl     // 120pt
    
    var dimension: CGFloat {
        switch self {
        case .xs:
            return 24
        case .sm:
            return 32
        case .md:
            return 40
        case .lg:
            return 56
        case .xl:
            return 80
        case .xxl:
            return 120
        }
    }
    
    var fontSize: CGFloat {
        switch self {
        case .xs, .sm:
            return 10
        case .md:
            return 14
        case .lg:
            return 18
        case .xl:
            return 24
        case .xxl:
            return 36
        }
    }
}

/// Status badge type
public enum AvatarStatus {
    case none
    case online
    case idle
    case offline
    case verified
}

/// ECHO Avatar Component
/// 6 sizes with optional status badge and trust ring
public struct AvatarView: View {
    let imageURL: URL?
    let initials: String?
    let size: AvatarSize
    let status: AvatarStatus
    let showTrustRing: Bool
    let trustLevel: String?
    
    public init(
        imageURL: URL? = nil,
        initials: String? = nil,
        size: AvatarSize = .md,
        status: AvatarStatus = .none,
        showTrustRing: Bool = false,
        trustLevel: String? = nil
    ) {
        self.imageURL = imageURL
        self.initials = initials
        self.size = size
        self.status = status
        self.showTrustRing = showTrustRing
        self.trustLevel = trustLevel
    }
    
    var statusColor: Color {
        switch status {
        case .none:
            return .clear
        case .online:
            return .echoSuccess
        case .idle:
            return .echoWarning
        case .offline:
            return .echoGray400
        case .verified:
            return .echoPrimary
        }
    }
    
    var statusBadgeSize: CGFloat {
        return size.dimension * 0.3
    }
    
    public var body: some View {
        ZStack(alignment: .bottomTrailing) {
            // Trust Ring
            if showTrustRing {
                Circle()
                    .stroke(
                        Color.trustColor(for: trustLevel ?? "newcomer"),
                        lineWidth: 3
                    )
                    .frame(width: size.dimension, height: size.dimension)
            }
            
            // Avatar Circle
            ZStack {
                if let imageURL = imageURL {
                    AsyncImage(url: imageURL) { phase in
                        switch phase {
                        case .empty:
                            ProgressView()
                                .frame(maxWidth: .infinity, maxHeight: .infinity)
                        case .success(let image):
                            image
                                .resizable()
                                .scaledToFill()
                        case .failure:
                            initialsView
                        @unknown default:
                            initialsView
                        }
                    }
                } else {
                    initialsView
                }
            }
            .frame(width: size.dimension, height: size.dimension)
            .clipShape(Circle())
            .background(Circle().fill(Color.echoPrimary.opacity(0.1)))
            
            // Status Badge
            if status != .none {
                Circle()
                    .fill(statusColor)
                    .frame(width: statusBadgeSize, height: statusBadgeSize)
                    .border(Color.white, width: 2)
                    .accessibility(label: Text("Status: \(statusLabel)"))
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibility(label: Text("User avatar"))
        .accessibility(hint: Text(initials ?? "Avatar image"))
    }
    
    private var initialsView: some View {
        ZStack {
            Circle()
                .fill(Color.echoPrimary.opacity(0.2))
            
            if let initials = initials {
                Text(initials)
                    .font(.system(size: size.fontSize, weight: .semibold))
                    .foregroundColor(.echoPrimary)
            }
        }
    }
    
    private var statusLabel: String {
        switch status {
        case .none:
            return "No status"
        case .online:
            return "Online"
        case .idle:
            return "Idle"
        case .offline:
            return "Offline"
        case .verified:
            return "Verified"
        }
    }
}

// MARK: - Preview

#if DEBUG
struct AvatarView_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(initials: "JD", size: .xs)
                AvatarView(initials: "JD", size: .sm)
                AvatarView(initials: "JD", size: .md)
                AvatarView(initials: "JD", size: .lg)
            }
            
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(initials: "JD", size: .xl)
                AvatarView(initials: "JD", size: .xxl)
            }
            
            VStack(spacing: Spacing.md.rawValue) {
                AvatarView(initials: "JD", size: .lg, status: .online)
                AvatarView(initials: "JD", size: .lg, status: .idle)
                AvatarView(initials: "JD", size: .lg, status: .offline)
                AvatarView(initials: "JD", size: .lg, status: .verified)
            }
            
            AvatarView(
                initials: "JD",
                size: .xl,
                status: .online,
                showTrustRing: true,
                trustLevel: "verified"
            )
            
            Spacer()
        }
        .echoSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
