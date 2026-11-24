#if os(iOS)
import StoreKit
import SwiftUI

/// ECHO VIP — auto-renewable App Store subscription (`com.echo.app.vip.monthly`).
struct VIPSubscriptionView: View {
    @State private var isProcessing = false
    @State private var errorMessage: String?
    @State private var showSuccess = false
    @State private var isActive = VIPSubscriptionStore.isActive
    @State private var priceState: VIPPriceDisplay.LoadState = .loading

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                headerCard
                benefitsList
                pricingCard
                legalNote
            }
            .padding(Spacing.lg.rawValue)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .navigationTitle("ECHO VIP")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            VIPSubscriptionService.startTransactionListener()
            priceState = VIPPriceDisplay.loadState(displayPrice: await VIPSubscriptionService.loadMonthlyPrice())
            isActive = VIPSubscriptionStore.isActive
        }
        .alert("VIP activated", isPresented: $showSuccess) {
            Button("OK", role: .cancel) {}
        } message: {
            Text("You're on ECHO VIP. Your benefits are active on this device.")
        }
    }

    private var headerCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 10) {
                Image(systemName: "star.circle.fill")
                    .font(.system(size: 32))
                    .foregroundStyle(Color.echoTrustGreen)
                VStack(alignment: .leading, spacing: 2) {
                    Text("ECHO VIP")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundColor(.echoInk)
                    Text(isActive ? "Member" : "Get premium features")
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk55)
                }
            }
            if isActive, let expires = VIPSubscriptionStore.expiresAt {
                Text("Renews \(expires.formatted(date: .abbreviated, time: .omitted))")
                    .font(.system(size: 13))
                    .foregroundColor(.echoInk55)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(16)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    private var benefitsList: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("What's included")
                .font(.system(size: 13, weight: .semibold))
                .foregroundColor(.echoInk55)
                .textCase(.uppercase)
            // Only list what the app delivers today. Rate-limit boosts, themes /
            // app icons and priority delivery are not implemented yet; listing
            // them on a paid subscription would be selling features that do
            // not exist. Add each row back when its feature ships.
            benefitRow("checkmark.seal.fill", "Violet VIP mark beside your name")
        }
    }

    private func benefitRow(_ icon: String, _ text: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 16))
                .foregroundColor(.echoSignal)
                .frame(width: 24)
            Text(text)
                .font(.system(size: 15))
                .foregroundColor(.echoInk)
        }
    }

    private var pricingCard: some View {
        VStack(spacing: 14) {
            switch priceState {
            case .loaded(let displayPrice):
                HStack(alignment: .firstTextBaseline, spacing: 4) {
                    Text(displayPrice)
                        .font(.system(size: 36, weight: .bold))
                        .foregroundColor(.echoInk)
                    Text("/ month")
                        .font(.system(size: 16))
                        .foregroundColor(.echoInk55)
                }
            case .loading:
                ProgressView()
            case .unavailable:
                Text("ECHO VIP isn't available from the App Store right now. Check your connection and try again.")
                    .font(.system(size: 14))
                    .foregroundColor(.echoInk55)
                    .multilineTextAlignment(.center)
                Button("Try again") {
                    Task { await retryPriceLoad() }
                }
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.echoSignal)
                .frame(minHeight: 44)
            }

            if let errorMessage {
                Text(errorMessage)
                    .font(.system(size: 13))
                    .foregroundColor(.echoAlert)
                    .multilineTextAlignment(.center)
            }

            if isActive {
                EchoButton("Manage subscription", style: .secondary, size: .large, action: {
                    Task { await manage() }
                })
            } else {
                EchoButton(
                    VIPPriceDisplay.subscribeButtonTitle(state: priceState, isProcessing: isProcessing),
                    style: .primary,
                    size: .large,
                    isDisabled: !VIPPriceDisplay.canSubscribe(state: priceState, isProcessing: isProcessing),
                    action: { Task { await subscribe() } }
                )
                .disabled(!VIPPriceDisplay.canSubscribe(state: priceState, isProcessing: isProcessing))
            }

            Button("Restore purchases") {
                Task { await restore() }
            }
            .font(.system(size: 14, weight: .semibold))
            .foregroundColor(.echoSignal)
            .disabled(isProcessing)
        }
        .frame(maxWidth: .infinity)
        .padding(20)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    private var legalNote: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Payment is charged to your Apple ID. ECHO VIP renews each month until you cancel in Settings → Apple ID → Subscriptions. Free Echo keeps full private messaging. VIP adds convenience and status, not extra security. Points in the Rewards tab are earned through activity and are not part of this purchase; VIP does not add to them.")
                .font(.system(size: 12))
                .foregroundColor(.echoInk55)
                .fixedSize(horizontal: false, vertical: true)
            // Auto-renewable subscriptions must link Terms and Privacy from the purchase screen.
            HStack(spacing: 16) {
                Link("Terms of Service", destination: LegalLinks.termsURL)
                Link("Privacy Policy", destination: LegalLinks.privacyURL)
            }
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.echoSignal)
            .frame(minHeight: 44)
        }
    }

    private func subscribe() async {
        isProcessing = true
        errorMessage = nil
        defer { isProcessing = false }
        do {
            try await VIPSubscriptionService.purchaseMonthly()
            isActive = VIPSubscriptionStore.isActive
            if isActive { showSuccess = true }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func retryPriceLoad() async {
        priceState = .loading
        priceState = VIPPriceDisplay.loadState(displayPrice: await VIPSubscriptionService.loadMonthlyPrice())
    }

    private func restore() async {
        isProcessing = true
        errorMessage = nil
        defer { isProcessing = false }
        do {
            let restored = try await VIPSubscriptionService.restore()
            isActive = VIPSubscriptionStore.isActive
            if !restored {
                errorMessage = "No active ECHO VIP subscription was found for this Apple ID."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func manage() async {
        guard let scene = UIApplication.shared.connectedScenes
            .compactMap({ $0 as? UIWindowScene })
            .first else { return }
        try? await AppStore.showManageSubscriptions(in: scene)
    }
}
#endif
