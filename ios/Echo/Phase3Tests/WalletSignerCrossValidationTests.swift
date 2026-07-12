#if os(iOS)
import XCTest
@testable import Echo

/// iOS half of the M0 real-funds de-risk gate. Proves the embedded dag4 signing
/// bundle (echo-wallet.bundle.js, loaded by `StargazerSigner`) agrees byte-for-byte
/// with the Go backend on DAG address derivation and secp256k1 signing.
///
/// The constants below are the canonical vector in
/// `Resources/wallet-sdk/testvector.json`, produced by `gen-vector.js` and
/// independently verified by `internal/wallet/dag_crossvalidation_test.go`. If the
/// bundle or dag4 changes incompatibly, this fails before any real funds move.
final class WalletSignerCrossValidationTests: XCTestCase {
    // Canonical vector (BIP-39 all-zero entropy seed, HD path m/44'/1137'/0'/0).
    private let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
    private let privateKey = "041a5c5c5d5cdc229ee0c668bf86cf48183b47bb9948298456cccfc32757d901"
    private let publicKey = "0453f06ad396d382ff1db457e6d2b608c04be2678bbd12207625e581f1e030c4c4c8f9db9094424eea5f7868b846301a8e8857e2c01583714316a043edd192798b"
    private let address = "DAG35XRkcBSHPqT8h62hDzzJJ7YerwUPgqDzGm2P"
    private let message = "echo-wallet-ownership:did:key:zVECTOR:nonce-0123456789"
    private let signature = "3045022100a6ad1cb5f28ccaede29a0b19768c93682abaf640ecf0900ba926b0a0c98976bd02206c6586785c9d8cf83351ecc9d896a42c0b8665e38048080ed21b67797d8b4ac4"

    /// Guards that echo-wallet.bundle.js is actually in the app target's Copy
    /// Bundle Resources phase (the whole stack is inert without it).
    func testSigningBundleIsBundled() {
        XCTAssertNotNil(
            Bundle.main.url(forResource: "echo-wallet.bundle", withExtension: "js"),
            "echo-wallet.bundle.js must be in the app target's Copy Bundle Resources phase"
        )
    }

    /// deriveAddress(pubKey) must match the vector — the pure pubkey→address path
    /// the Go backend uses to reject unauthorized addresses.
    func testDeriveAddressMatchesVector() async throws {
        let derived = try await StargazerSigner.shared.deriveAddress(publicKey: publicKey)
        XCTAssertEqual(derived, address, "dag4 address derivation diverged from the shared vector")
    }

    /// HD derivation from the mnemonic must reproduce the same account.
    func testImportMnemonicMatchesVector() async throws {
        let account = try await StargazerSigner.shared.importMnemonic(mnemonic)
        XCTAssertEqual(account.publicKey, publicKey, "public key from mnemonic diverged")
        XCTAssertEqual(account.address, address, "address from mnemonic diverged")
    }

    /// Restoring from a raw private key must yield the same public key + address.
    func testAccountFromPrivateKeyMatchesVector() async throws {
        let account = try await StargazerSigner.shared.accountFromPrivateKey(privateKey)
        XCTAssertEqual(account.privateKey, privateKey, "private key round-trip diverged")
        XCTAssertEqual(account.publicKey, publicKey, "public key from private key diverged")
        XCTAssertEqual(account.address, address, "address from private key diverged")
    }

    /// Deterministic (RFC 6979) secp256k1 signing must reproduce the exact
    /// signature the Go verifier accepts for this message.
    func testSignMessageReproducesVectorSignature() async throws {
        let sig = try await StargazerSigner.shared.signMessage(privateKey: privateKey, message: message)
        XCTAssertEqual(sig, signature, "signature diverged from the shared vector (Go verifier would reject)")
    }
}
#endif
