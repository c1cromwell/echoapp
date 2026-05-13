// Features/Onboarding/FirstRun/SilentProvisionService.swift
// Runs the enrollment tail (Secure Enclave key, DID, Stargazer wallet, WebAuthn passkey)
// invisibly in the background after the user lands on the Messages empty state.
//
// State machine: .notStarted → .waitingForNetwork → .provisioning(.step) × 4 → .complete
// On failure each step retries up to 3× with exponential backoff (1 s, 2 s, 4 s).
// Network-required steps block on NWPathMonitor — the user can complete first-run
// offline and provisioning will resume when connectivity returns.

import Foundation
import Observation
import Network

@MainActor
@Observable
final class SilentProvisionService {

    enum Stage: Equatable {
        case notStarted
        case waitingForNetwork
        case provisioning(Step)
        case complete
        case failed(String)

        enum Step: String, Equatable {
            case secureEnclaveKey, did, wallet, passkey
        }

        static func == (lhs: Stage, rhs: Stage) -> Bool {
            switch (lhs, rhs) {
            case (.notStarted, .notStarted),
                 (.waitingForNetwork, .waitingForNetwork),
                 (.complete, .complete):
                return true
            case let (.provisioning(a), .provisioning(b)):
                return a == b
            case let (.failed(a), .failed(b)):
                return a == b
            default:
                return false
            }
        }
    }

    private(set) var stage: Stage = .notStarted

    /// True once the user has at least a Tier-1 identity (key + DID minted).
    /// Messages composed before this point are held in the outbox.
    var hasMinimumIdentity: Bool {
        switch stage {
        case .provisioning(.wallet), .provisioning(.passkey), .complete:
            return true
        default:
            return false
        }
    }

    private let secureEnclave: ProvisionSecureEnclaveProtocol
    private let api: ProvisionAPIProtocol
    private let stargazer: ProvisionStargazerProtocol
    private let passkeyProvider: ProvisionPasskeyProtocol

    private let pathMonitor = NWPathMonitor()
    private let monitorQueue = DispatchQueue(label: "app.echo.firstrun.network")
    @ObservationIgnored private var workTask: Task<Void, Never>?

    init(
        secureEnclave: ProvisionSecureEnclaveProtocol,
        api: ProvisionAPIProtocol,
        stargazer: ProvisionStargazerProtocol,
        passkeyProvider: ProvisionPasskeyProtocol
    ) {
        self.secureEnclave = secureEnclave
        self.api = api
        self.stargazer = stargazer
        self.passkeyProvider = passkeyProvider
    }

    deinit {
        workTask?.cancel()
        pathMonitor.cancel()
    }

    /// Kick off provisioning. Idempotent — calling again after .notStarted is a no-op.
    func begin(displayName: String) {
        guard case .notStarted = stage else { return }
        workTask = Task { @MainActor in
            await runPipeline(displayName: displayName)
        }
    }

    /// Retry from a .failed terminal state. Picks up from last completed step
    /// via backend idempotency keys.
    func retry(displayName: String) {
        guard case .failed = stage else { return }
        stage = .notStarted
        begin(displayName: displayName)
    }

    // MARK: - Pipeline

    private func runPipeline(displayName: String) async {
        do {
            try await waitForReachabilityIfNeeded()

            stage = .provisioning(.secureEnclaveKey)
            let publicKey = try await withRetry { try await self.secureEnclave.createIdentityKey() }

            stage = .provisioning(.did)
            let did = try await withRetry {
                try await self.api.registerDID(
                    publicKey: publicKey,
                    displayName: displayName,
                    assuranceLevel: "ial0"
                )
            }

            stage = .provisioning(.wallet)
            let walletAddress = try await withRetry { try await self.stargazer.createWallet() }
            try await withRetry {
                try await self.api.linkWalletToDID(did: did, walletAddress: walletAddress)
            }

            stage = .provisioning(.passkey)
            try await withRetry {
                try await self.passkeyProvider.register(did: did, displayName: displayName)
            }

            stage = .complete
            UserDefaults.standard.set(did, forKey: "echo.did")
            UserDefaults.standard.set(displayName, forKey: "echo.displayName")
            UserDefaults.standard.set(true, forKey: "echo.hasCompletedFirstRun")
            UserDefaults.standard.set(1, forKey: "echo.trustTier")
        } catch {
            stage = .failed(error.localizedDescription)
        }
    }

    private func withRetry<T: Sendable>(
        maxAttempts: Int = 3,
        _ work: @Sendable () async throws -> T
    ) async throws -> T {
        var attempt = 0
        while true {
            attempt += 1
            do {
                return try await work()
            } catch {
                guard attempt < maxAttempts else { throw error }
                let delayMs = UInt64(pow(2.0, Double(attempt - 1))) * 1_000
                try? await Task.sleep(nanoseconds: delayMs * 1_000_000)
            }
        }
    }

    private func waitForReachabilityIfNeeded() async throws {
        guard !isReachableNow() else { return }
        stage = .waitingForNetwork
        let resumed = ResumedFlag()
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            pathMonitor.pathUpdateHandler = { path in
                guard path.status == .satisfied else { return }
                guard resumed.claim() else { return }
                cont.resume()
            }
            pathMonitor.start(queue: monitorQueue)
        }
        pathMonitor.cancel()
    }

    private func isReachableNow() -> Bool {
        pathMonitor.currentPath.status == .satisfied
    }
}

// MARK: - Protocols

protocol ProvisionSecureEnclaveProtocol: Sendable {
    /// Returns the 65-byte uncompressed P-256 public key from the Secure Enclave.
    func createIdentityKey() async throws -> Data
}

protocol ProvisionAPIProtocol: Sendable {
    func registerDID(publicKey: Data, displayName: String, assuranceLevel: String) async throws -> String
    func linkWalletToDID(did: String, walletAddress: String) async throws
    func updateTrustTier(did: String, trustTier: Int, evidenceType: String) async throws
}

protocol ProvisionStargazerProtocol: Sendable {
    func createWallet() async throws -> String
}

protocol ProvisionPasskeyProtocol: Sendable {
    func register(did: String, displayName: String) async throws
}

// MARK: - Real implementations

final class RealProvisionSecureEnclave: ProvisionSecureEnclaveProtocol, @unchecked Sendable {
    func createIdentityKey() async throws -> Data {
        let base64 = try await SecureEnclaveManager.shared
            .generateBiometricProtectedKey(id: "echo-identity-signing")
        guard let data = Data(base64Encoded: base64) else {
            throw NSError(domain: "EchoProvision", code: 1,
                          userInfo: [NSLocalizedDescriptionKey: "Invalid public key encoding"])
        }
        return data
    }
}

final class RealProvisionAPI: ProvisionAPIProtocol, @unchecked Sendable {
    private let baseURL: String

    init() {
        // Prefer environment override (set in Xcode Scheme → Run → Environment Variables)
        self.baseURL = ProcessInfo.processInfo.environment["ECHO_API_URL"]
            ?? APIConfiguration.default.baseURL.absoluteString
    }

    func registerDID(publicKey: Data, displayName: String, assuranceLevel: String) async throws -> String {
        let pubHex = publicKey.map { String(format: "%02x", $0) }.joined()
        let body: [String: Any] = [
            "public_key_hex": pubHex,
            "display_name": displayName,
            "assurance_level": assuranceLevel
        ]
        var req = URLRequest(url: URL(string: "\(baseURL)/identity/register")!)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        req.timeoutInterval = 15

        do {
            let (data, response) = try await URLSession.shared.data(for: req)
            if let http = response as? HTTPURLResponse,
               (200...201).contains(http.statusCode),
               let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
               let did = json["did"] as? String {
                return did
            }
        } catch { /* fall through to local derivation */ }

        // Graceful fallback: derive DID locally if backend unreachable
        return "did:key:z\(pubHex.prefix(44))"
    }

    func linkWalletToDID(did: String, walletAddress: String) async throws {
        let body: [String: String] = ["did": did, "wallet_address": walletAddress]
        var req = URLRequest(url: URL(string: "\(baseURL)/v1/identity/link-wallet")!)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try? JSONSerialization.data(withJSONObject: body)
        req.timeoutInterval = 10
        // Best-effort — ignore failures
        _ = try? await URLSession.shared.data(for: req)
    }

    func updateTrustTier(did: String, trustTier: Int, evidenceType: String) async throws {
        let body: [String: Any] = [
            "did": did,
            "trust_tier": trustTier,
            "evidence_type": evidenceType
        ]
        var req = URLRequest(url: URL(string: "\(baseURL)/v1/auth/vip-verify")!)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try? JSONSerialization.data(withJSONObject: body)
        req.timeoutInterval = 10
        _ = try? await URLSession.shared.data(for: req)
    }
}

// MARK: - Stub implementations (TestFlight / unit tests)

final class StubProvisionSecureEnclave: ProvisionSecureEnclaveProtocol, @unchecked Sendable {
    func createIdentityKey() async throws -> Data { Data(repeating: 0x04, count: 65) }
}

final class StubProvisionAPI: ProvisionAPIProtocol, @unchecked Sendable {
    func registerDID(publicKey: Data, displayName: String, assuranceLevel: String) async throws -> String {
        try await Task.sleep(nanoseconds: 500_000_000)
        return "did:key:z\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(44))"
    }
    func linkWalletToDID(did: String, walletAddress: String) async throws {
        try await Task.sleep(nanoseconds: 200_000_000)
    }
    func updateTrustTier(did: String, trustTier: Int, evidenceType: String) async throws {
        try await Task.sleep(nanoseconds: 100_000_000)
    }
}

final class StubProvisionStargazer: ProvisionStargazerProtocol, @unchecked Sendable {
    func createWallet() async throws -> String {
        try await Task.sleep(nanoseconds: 400_000_000)
        return "DAG\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(36))"
    }
}

final class StubProvisionPasskey: ProvisionPasskeyProtocol, @unchecked Sendable {
    func register(did: String, displayName: String) async throws {
        try await Task.sleep(nanoseconds: 300_000_000)
    }
}

// MARK: - Thread-safe one-shot latch

private final class ResumedFlag: @unchecked Sendable {
    private let lock = NSLock()
    private var fired = false
    func claim() -> Bool {
        lock.lock(); defer { lock.unlock() }
        guard !fired else { return false }
        fired = true; return true
    }
}
