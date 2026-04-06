import Foundation

// MARK: - Auth State Machine

enum AuthState: Equatable {
    case unauthenticated
    case phoneEntry
    case otpVerification(phone: String, verificationId: String)
    case passkeySetup(tempToken: String)
    case profileSetup
    case trustIntro
    case authenticated(user: AuthUserProfile)
    case locked(reason: LockReason, retryAfter: Date?)
    case recovery(method: RecoveryMethod)

    static func == (lhs: AuthState, rhs: AuthState) -> Bool {
        switch (lhs, rhs) {
        case (.unauthenticated, .unauthenticated): return true
        case (.phoneEntry, .phoneEntry): return true
        case (.otpVerification(let lp, let lv), .otpVerification(let rp, let rv)):
            return lp == rp && lv == rv
        case (.passkeySetup(let lt), .passkeySetup(let rt)):
            return lt == rt
        case (.profileSetup, .profileSetup): return true
        case (.trustIntro, .trustIntro): return true
        case (.authenticated, .authenticated): return true
        case (.locked(let lr, let ld), .locked(let rr, let rd)):
            return lr == rr && ld == rd
        case (.recovery(let lm), .recovery(let rm)):
            return lm == rm
        default: return false
        }
    }
}

// MARK: - Lock Reason

enum LockReason: Equatable {
    case tooManyAttempts
    case suspiciousActivity
    case accountSuspended
}

// MARK: - Recovery Method

enum RecoveryMethod: String, CaseIterable {
    case recoveryPhrase = "recovery_phrase"
    case trustedContacts = "trusted_contacts"
    case phoneReverification = "phone"
}

// MARK: - Auth Events

enum AuthEvent {
    case phoneSubmitted(phone: String, verificationId: String)
    case otpVerified(tempToken: String)
    case passkeyCreated
    case profileCompleted
    case trustIntroDismissed
    case loginSucceeded(AuthUserProfile)
    case sessionExpired
    case loggedOut
    case accountLocked(LockReason, retryAfter: Date?)
    case recoveryInitiated(RecoveryMethod)
    case recoveryCompleted
}

// MARK: - Step-Up Actions

enum StepUpAction: String, Hashable {
    case revokeDevice = "revoke_device"
    case changePhone = "change_phone"
    case sendLargePayment = "send_large_payment"
    case exportRecoveryPhrase = "export_recovery_phrase"
    case deleteAccount = "delete_account"

    var displayTitle: String {
        switch self {
        case .revokeDevice: return "Remove Device"
        case .changePhone: return "Change Phone Number"
        case .sendLargePayment: return "Large Payment"
        case .exportRecoveryPhrase: return "Export Recovery Phrase"
        case .deleteAccount: return "Delete Account"
        }
    }

    var requiresPasskeyAndOTP: Bool {
        switch self {
        case .exportRecoveryPhrase, .deleteAccount: return true
        default: return false
        }
    }
}

// MARK: - AuthUserProfile (lightweight view model for auth state)

struct AuthUserProfile: Equatable, Codable {
    let id: String
    let did: String
    let displayName: String?
    let username: String?
    let trustScore: Int
    let trustTier: Int

    static func == (lhs: AuthUserProfile, rhs: AuthUserProfile) -> Bool {
        lhs.id == rhs.id
    }
}
