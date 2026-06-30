#if os(iOS)
// Core/Stargazer/WalletKeyStore.swift
// Custodies the user-held dag4 wallet key. Generates a BIP-39 mnemonic once via
// the embedded signer, persists it in the Keychain (device-only), and derives
// the account on demand. The backend never receives the private key — only the
// public key (bound at link time) and per-action signatures.

import Foundation

actor WalletKeyStore {
    static let shared = WalletKeyStore()

    static let mnemonicKey = "echo.wallet.seed"
    static let publicKeyKey = "echo.wallet.publicKey"
    /// Cleared once the user confirms they have backed up the recovery phrase;
    /// gates real-funds actions (backup-before-real-funds).
    static let backupRequiredKey = "echo.wallet.backupRequired"

    private let signer: StargazerSigner
    private let keychain: KeychainManager
    private var cached: StargazerSigner.WalletAccount?

    init(signer: StargazerSigner = .shared, keychain: KeychainManager = .shared) {
        self.signer = signer
        self.keychain = keychain
    }

    /// Returns the wallet account, generating + persisting a new key on first use.
    func ensureWallet() async throws -> StargazerSigner.WalletAccount {
        if let cached { return cached }
        if let mnemonic = try? await keychain.retrieve(key: Self.mnemonicKey, as: String.self),
           !mnemonic.isEmpty {
            let acct = try await signer.importMnemonic(mnemonic)
            cached = acct
            return acct
        }
        let mnemonic = try await signer.generateMnemonic()
        let acct = try await signer.importMnemonic(mnemonic)
        try await persist(mnemonic: mnemonic, account: acct)
        // New wallet: phrase is not yet backed up.
        UserDefaults.standard.set(true, forKey: Self.backupRequiredKey)
        cached = acct
        return acct
    }

    /// The existing account, or nil if not yet provisioned.
    func currentAccount() async throws -> StargazerSigner.WalletAccount? {
        if let cached { return cached }
        guard let mnemonic = try? await keychain.retrieve(key: Self.mnemonicKey, as: String.self),
              !mnemonic.isEmpty else {
            return nil
        }
        let acct = try await signer.importMnemonic(mnemonic)
        cached = acct
        return acct
    }

    /// Reveals the recovery phrase. Callers MUST gate this behind device auth.
    func exportMnemonic() async throws -> String {
        guard let mnemonic = try? await keychain.retrieve(key: Self.mnemonicKey, as: String.self),
              !mnemonic.isEmpty else {
            throw StargazerSigner.SignerError.badResult
        }
        return mnemonic
    }

    /// Restores a wallet from a user-supplied phrase, replacing local key state.
    func restore(mnemonic: String) async throws -> StargazerSigner.WalletAccount {
        let normalized = mnemonic.trimmingCharacters(in: .whitespacesAndNewlines)
        let acct = try await signer.importMnemonic(normalized) // throws on invalid phrase
        try await persist(mnemonic: normalized, account: acct)
        markBackedUp() // a restored phrase is, by definition, already known to the user
        cached = acct
        return acct
    }

    /// Signs a proof-of-ownership challenge with the wallet key.
    func signChallenge(_ challenge: String) async throws -> (publicKey: String, signature: String) {
        let acct = try await ensureWallet()
        let sig = try await signer.signMessage(privateKey: acct.privateKey, message: challenge)
        return (acct.publicKey, sig)
    }

    // MARK: - Backup gate

    nonisolated func backupRequired() -> Bool {
        UserDefaults.standard.bool(forKey: Self.backupRequiredKey)
    }

    nonisolated func markBackedUp() {
        UserDefaults.standard.set(false, forKey: Self.backupRequiredKey)
    }

    // MARK: - Private

    private func persist(mnemonic: String, account: StargazerSigner.WalletAccount) async throws {
        try await keychain.store(key: Self.mnemonicKey, value: mnemonic)
        try await keychain.store(key: Self.publicKeyKey, value: account.publicKey)
        try await keychain.store(key: WalletKeychain.dagAddressKey, value: account.address)
    }
}
#endif
