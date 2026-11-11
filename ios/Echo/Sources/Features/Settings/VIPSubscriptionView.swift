#if os(iOS)
import StoreKit
import SwiftUI

/// ECHO VIP — auto-renewable App Store subscription (`com.echo.app.vip.monthly`).
struct VIPSubscriptionView: View {
    @State private var isProcessing = false
    @State private var errorMessage: String?
    @State private var showSuccess = false
    @State private var isActive = VIPSubscriptionStore.isActive
    @State private var priceLabel = VIPSubscriptionStore.monthlyPriceLabel

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
            if let price = await VIPSubscriptionService.loadMonthlyPrice() {
                priceLabel = price
            }
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
            benefitRow("bolt.fill", "2× API rate limits")
            benefitRow("paintbrush.fill", "Premium themes & app icon")
            benefitRow("checkmark.seal.fill", "VIP trust badge on profile")
            benefitRow("bell.badge.fill", "Priority signal delivery")
            benefitRow("heart.fill", "Supports the community treasury")
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
            HStack(alignment: .firstTextBaseline, spacing: 4) {
                Text(priceLabel)
                    .font(.system(size: 36, weight: .bold))
                    .foregroundColor(.echoInk)
                Text("/ month")
                    .font(.system(size: 16))
                    .foregroundColor(.echoInk55)
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
                    isProcessing ? "Contacting the App Store…" : "Subscribe for \(priceLabel)/mo",
                    style: .primary,
                    size: .large,
                    action: { Task { await subscribe() } }
                )
                .disabled(isProcessing)
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
        Text("Payment is charged to your Apple ID. ECHO VIP renews each month until you cancel in Settings → Apple ID → Subscriptions. Free Echo keeps full private messaging. VIP adds convenience and status, not extra security. Rewards in the Rewards tab are earned and are not part of this purchase.")
            .font(.system(size: 12))
            .foregroundColor(.echoInk55)
            .fixedSize(horizontal: false, vertical: true)
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
