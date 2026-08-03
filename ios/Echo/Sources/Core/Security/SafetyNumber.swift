import Foundation
import CryptoKit

/// Echo safety numbers (Signal Parity Wave S2) — comparable to Signal Safety Numbers.
/// Fingerprint is derived from both parties' `did:key` material (sorted) so either
/// side shows the same number; Glacial-styled compare UI lives in `SafetyNumberCompareView`.
enum SafetyNumber {
    /// 60-digit display number grouped as 12×5 (Signal-like readability).
    static func digits(localDID: String, peerDID: String) -> String {
        let a = localDID.trimmingCharacters(in: .whitespacesAndNewlines)
        let b = peerDID.trimmingCharacters(in: .whitespacesAndNewlines)
        let sorted = [a, b].sorted()
        let material = Data((sorted[0] + "|" + sorted[1]).utf8)
        let digest = SHA256.hash(data: material)
        // Expand to 60 decimal digits via iterative hashing.
        var digits = ""
        var block = Data(digest)
        while digits.count < 60 {
            let next = SHA256.hash(data: block)
            for byte in next {
                digits.append(String(byte % 10))
                if digits.count >= 60 { break }
            }
            block = Data(next)
        }
        return String(digits.prefix(60))
    }

    /// Grouped for display: "12345 67890 …"
    static func formatted(localDID: String, peerDID: String) -> String {
        let raw = digits(localDID: localDID, peerDID: peerDID)
        return stride(from: 0, to: raw.count, by: 5).map { i in
            let start = raw.index(raw.startIndex, offsetBy: i)
            let end = raw.index(start, offsetBy: 5, limitedBy: raw.endIndex) ?? raw.endIndex
            return String(raw[start..<end])
        }.joined(separator: " ")
    }

    /// Short compare for in-chat key-change banners (first 20 digits).
    static func shortFingerprint(localDID: String, peerDID: String) -> String {
        String(digits(localDID: localDID, peerDID: peerDID).prefix(20))
    }

    /// Persist last-seen peer fingerprint; returns true when peer identity material changed.
    static func didPeerKeyChange(conversationId: String, localDID: String, peerDID: String) -> Bool {
        let key = "echo.safety.\(conversationId).fp"
        let current = digits(localDID: localDID, peerDID: peerDID)
        let previous = UserDefaults.standard.string(forKey: key)
        UserDefaults.standard.set(current, forKey: key)
        guard let previous else { return false }
        return previous != current
    }
}

#if os(iOS)
import SwiftUI

/// Glacial-styled safety number compare sheet.
struct SafetyNumberCompareView: View {
    let contactName: String
    let localDID: String
    let peerDID: String
    var onDismiss: () -> Void = {}

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    Text("Verify \(contactName)")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                    Text("Compare this number with your contact out-of-band (in person, voice call, or another channel). If it matches, you share the same identity keys.")
                        .font(.system(size: 14))
                        .foregroundStyle(Color.echoInk55)
                    Text(SafetyNumber.formatted(localDID: localDID, peerDID: peerDID))
                        .font(.echomono(15))
                        .foregroundStyle(Color.echoInk)
                        .padding(16)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(Color.echoPaperDim)
                        .clipShape(RoundedRectangle(cornerRadius: 14))
                        .textSelection(.enabled)
                    Text("ECHO never shows raw private keys. This fingerprint is derived from both of your DIDs.")
                        .font(.caption)
                        .foregroundStyle(Color.echoInk40)
                }
                .padding(20)
            }
            .background(Color.echoPaper.ignoresSafeArea())
            .navigationTitle("Safety number")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done", action: onDismiss)
                }
            }
        }
    }
}
#endif
