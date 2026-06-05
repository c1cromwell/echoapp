#if os(iOS)
import SwiftUI

/// ECHO VIP — $9.99/month capacity and status upgrade (Phase 1 local + vip-verify).
struct VIPSubscriptionView: View {
    @State private var isProcessing = false
    @State private var errorMessage: String?
    @State private var showSuccess = false
    @State private var isActive = VIPSubscriptionStore.isActive

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
        .alert("VIP activated", isPresented: $showSuccess) {
            Button("OK", role: .cancel) {}
        } message: {
            Text("You're on ECHO VIP. Benefits apply on this device; App Store billing ships in a later update.")
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
                    Text(isActive ? "Member" : "Upgrade your experience")
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
                Text(VIPSubscriptionStore.monthlyPriceLabel)
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
                EchoButton("VIP active", style: .secondary, size: .large, action: {})
                    .disabled(true)
            } else {
                EchoButton(
                    isProcessing ? "Activating…" : "Subscribe for \(VIPSubscriptionStore.monthlyPriceLabel)/mo",
                    style: .primary,
                    size: .large,
                    action: { Task { await subscribe() } }
                )
                .disabled(isProcessing)
            }
        }
        .frame(maxWidth: .infinity)
        .padding(20)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    private var legalNote: some View {
        Text("Free tier keeps full E2E messaging and anchoring. VIP adds convenience and status — never security. App Store auto-renew ships in a future build; this TestFlight path activates VIP for development.")
            .font(.system(size: 12))
            .foregroundColor(.echoInk55)
            .fixedSize(horizontal: false, vertical: true)
    }

    private func subscribe() async {
        isProcessing = true
        errorMessage = nil
        defer { isProcessing = false }
        do {
            try await VIPSubscriptionService.activateMonthly()
            isActive = true
            showSuccess = true
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
