import Foundation

/// The App Store product sold as ECHO VIP. Rewards stay earned; this is the
/// only subscription the messaging app sells.
enum VIPSubscriptionPolicy {
    static let monthlyProductID = "com.echo.app.vip.monthly"
    static let bundleID = "com.echo.app"

    static func accepts(productID: String, bundleID: String) -> Bool {
        productID == monthlyProductID && bundleID == Self.bundleID
    }
}

/// What the VIP screens show for price. The only price ever shown is the
/// App Store's own `Product.displayPrice` (localized, correct currency);
/// until that loads there is no price and nothing to subscribe to.
enum VIPPriceDisplay {
    enum LoadState: Equatable {
        case loading
        case loaded(displayPrice: String)
        case unavailable

        var displayPrice: String? {
            if case .loaded(let price) = self { return price }
            return nil
        }
    }

    static func loadState(displayPrice: String?) -> LoadState {
        guard let price = displayPrice?.trimmingCharacters(in: .whitespacesAndNewlines), !price.isEmpty else {
            return .unavailable
        }
        return .loaded(displayPrice: price)
    }

    /// Trailing value on the Settings "ECHO VIP" row; nil shows the title alone.
    static func settingsRowValue(isActive: Bool, displayPrice: String?) -> String? {
        if isActive { return "Active" }
        guard let price = loadState(displayPrice: displayPrice).displayPrice else { return nil }
        return "\(price)/mo"
    }

    static func subscribeButtonTitle(state: LoadState, isProcessing: Bool) -> String {
        if isProcessing { return "Contacting the App Store…" }
        switch state {
        case .loaded(let price): return "Subscribe for \(price)/mo"
        case .loading: return "Loading price…"
        case .unavailable: return "Subscribe"
        }
    }

    /// Subscribe is offered only once the App Store product (and its price) loaded.
    static func canSubscribe(state: LoadState, isProcessing: Bool) -> Bool {
        state.displayPrice != nil && !isProcessing
    }
}

#if os(iOS)
import StoreKit

/// Local mirror of the StoreKit entitlement. The server record is written
/// only after a verified transaction.
enum VIPSubscriptionStore {
    private static let activeKey = "echo.vip.subscription.active"
    private static let expiresKey = "echo.vip.subscription.expiresAt"

    static var isActive: Bool {
        guard UserDefaults.standard.bool(forKey: activeKey) else { return false }
        if let expires = expiresAt, expires < Date() {
            return false
        }
        return true
    }

    /// The mark drawn beside this account's own name.
    static var mark: MembershipMark? { isActive ? .vip : nil }

    static var expiresAt: Date? {
        UserDefaults.standard.object(forKey: expiresKey) as? Date
    }

    /// The last `Product.displayPrice` the App Store returned in this launch, if any.
    static var loadedDisplayPrice: String?

    /// Settings row value. `displayPrice` is the StoreKit price if it has loaded.
    static func statusLabel(displayPrice: String? = loadedDisplayPrice) -> String? {
        VIPPriceDisplay.settingsRowValue(isActive: isActive, displayPrice: displayPrice)
    }

    static func markActive(expiresAt: Date) {
        UserDefaults.standard.set(true, forKey: activeKey)
        UserDefaults.standard.set(expiresAt, forKey: expiresKey)
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: activeKey)
        UserDefaults.standard.removeObject(forKey: expiresKey)
    }
}

@MainActor
enum VIPSubscriptionService {
    enum ActivationError: LocalizedError {
        case notSignedIn
        case networkFailed
        case productUnavailable
        case unverified
        case pending

        var errorDescription: String? {
            switch self {
            case .notSignedIn: return "Sign in before upgrading to VIP."
            case .networkFailed: return "Could not activate VIP. Check your connection and try again."
            case .productUnavailable: return "ECHO VIP is not available from the App Store right now."
            case .unverified: return "The App Store could not verify this purchase."
            case .pending: return "This purchase is waiting for approval. VIP turns on when it completes."
            }
        }
    }

    private static var listener: Task<Void, Never>?

    static func startTransactionListener() {
        guard listener == nil else { return }
        listener = Task {
            for await update in StoreKit.Transaction.updates {
                guard let transaction = try? verified(update) else { continue }
                try? await confirm(transaction, signedTransaction: update.jwsRepresentation)
                await transaction.finish()
            }
        }
    }

    static func loadMonthlyPrice() async -> String? {
        let products = try? await Product.products(for: [VIPSubscriptionPolicy.monthlyProductID])
        if let price = products?.first?.displayPrice {
            VIPSubscriptionStore.loadedDisplayPrice = price
        }
        return products?.first?.displayPrice
    }

    static func purchaseMonthly() async throws {
        startTransactionListener()
        let products = try await Product.products(for: [VIPSubscriptionPolicy.monthlyProductID])
        guard let product = products.first else { throw ActivationError.productUnavailable }
        switch try await product.purchase() {
        case .success(let result):
            let transaction = try verified(result)
            try await confirm(transaction, signedTransaction: result.jwsRepresentation)
            await transaction.finish()
        case .userCancelled:
            return
        case .pending:
            throw ActivationError.pending
        @unknown default:
            throw ActivationError.networkFailed
        }
    }

    static func restore() async throws -> Bool {
        try await AppStore.sync()
        var restored = false
        for await result in StoreKit.Transaction.currentEntitlements {
            guard let transaction = try? verified(result),
                  transaction.productID == VIPSubscriptionPolicy.monthlyProductID else { continue }
            try await confirm(transaction, signedTransaction: result.jwsRepresentation)
            restored = true
        }
        if !restored {
            VIPSubscriptionStore.clear()
        }
        return restored
    }

    private static func verified<T>(_ result: VerificationResult<T>) throws -> T {
        switch result {
        case .verified(let value):
            return value
        case .unverified:
            throw ActivationError.unverified
        }
    }

    private static func confirm(_ transaction: StoreKit.Transaction, signedTransaction: String) async throws {
        guard VIPSubscriptionPolicy.accepts(
            productID: transaction.productID,
            bundleID: VIPSubscriptionPolicy.bundleID
        ) else {
            throw ActivationError.productUnavailable
        }
        guard let did = await CurrentUserSession.currentDID(), !did.isEmpty else {
            throw ActivationError.notSignedIn
        }
        let expires = transaction.expirationDate
            ?? Calendar.current.date(byAdding: .month, value: 1, to: Date())
            ?? Date()
        let body: [String: Any] = [
            "did": did,
            // A purchase grants the VIP mark. It does not change the trust tier.
            "trust_tier": 1,
            "evidence_type": "vip_subscription",
            "signed_transaction": signedTransaction
        ]
        var req = URLRequest(url: EchoAPIBaseURL.url(path: "/v1/auth/vip-verify"))
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        req.timeoutInterval = 15
        try await PasskeySigningInterceptor().intercept(&req)
        let (_, response) = try await URLSession.shared.data(for: req)
        guard let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode) else {
            throw ActivationError.networkFailed
        }
        VIPSubscriptionStore.markActive(expiresAt: expires)
    }
}
#endif
