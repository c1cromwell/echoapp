#if os(iOS)
// Core/Stargazer/StargazerSigner.swift
// Loads the embedded dag4 wallet bundle (echo-wallet.bundle.js) into a
// JavaScriptCore context and exposes real secp256k1 key generation, DAG address
// derivation, and signing. This is the client half of real-funds custody: the
// private key is supplied per call (held in the Keychain) and never persisted
// inside JS. The Go backend independently verifies addresses/signatures via the
// same dag4 algorithm (see internal/wallet/dag_address.go +
// dag_crossvalidation_test.go, which is the byte-compat gate for this file).

import Foundation
import JavaScriptCore
import Security

actor StargazerSigner {
    enum SignerError: LocalizedError {
        case bundleMissing
        case contextUnavailable
        case globalMissing
        case js(String)
        case badResult

        var errorDescription: String? {
            switch self {
            case .bundleMissing: return "wallet signing bundle not found in app resources"
            case .contextUnavailable: return "could not create JavaScript context"
            case .globalMissing: return "EchoWallet global missing from bundle"
            case .js(let m): return "wallet signing error: \(m)"
            case .badResult: return "unexpected wallet signing result"
            }
        }
    }

    static let shared = StargazerSigner()

    private let bundleResource: String
    private let bundleExtension: String
    private let bundleSourceOverride: String?
    private var context: JSContext?

    init(bundleResource: String = "echo-wallet.bundle", bundleExtension: String = "js") {
        self.bundleResource = bundleResource
        self.bundleExtension = bundleExtension
        self.bundleSourceOverride = nil
    }

    /// Test seam: load the bundle from an explicit source string instead of the
    /// app resources (so the signer can be exercised without app packaging).
    init(bundleSource: String) {
        self.bundleResource = ""
        self.bundleExtension = ""
        self.bundleSourceOverride = bundleSource
    }

    // MARK: - Public API

    /// Fresh BIP-39 mnemonic (DAG wallet seed). Requires secure RNG.
    func generateMnemonic() throws -> String {
        let echo = try wallet()
        guard let v = echo.invokeMethod("generateMnemonic", withArguments: []),
              let s = v.toString(), !s.isEmpty, !v.isUndefined else {
            throw SignerError.badResult
        }
        return s
    }

    /// Derives the DAG account from a mnemonic (HD path m/44'/1137'/0'/0).
    func importMnemonic(_ mnemonic: String) throws -> WalletAccount {
        try account(method: "importMnemonic", args: [mnemonic])
    }

    /// Derives public key + address from a raw private key (e.g. on restore).
    func accountFromPrivateKey(_ privateKey: String) throws -> WalletAccount {
        try account(method: "accountFromPrivateKey", args: [privateKey])
    }

    /// DAG address for a public key hex.
    func deriveAddress(publicKey: String) throws -> String {
        let echo = try wallet()
        guard let v = echo.invokeMethod("deriveAddress", withArguments: [publicKey]),
              let s = v.toString(), !s.isEmpty, !v.isUndefined else {
            throw SignerError.badResult
        }
        return s
    }

    /// secp256k1 ECDSA (DER hex) over SHA-512(message). Used for proof-of-ownership
    /// (signing the server challenge). Backed by an async dag4 Promise.
    func signMessage(privateKey: String, message: String) async throws -> String {
        let echo = try wallet()
        let promise = echo.invokeMethod("signMessage", withArguments: [privateKey, message])
        return try await resolvePromise(promise)
    }

    struct WalletAccount: Sendable {
        let privateKey: String
        let publicKey: String
        let address: String
    }

    /// Signs a Currency-L1 transaction body (TokenLock / DelegatedStake /
    /// WithdrawDelegatedStake) and returns the submittable {value, proofs}.
    func signTransaction(privateKey: String, publicKey: String, body: [String: Any]) async throws -> SignedTransaction {
        let echo = try wallet()
        let promise = echo.invokeMethod("signTransaction", withArguments: [privateKey, publicKey, body])
        let dict = try await resolvePromiseObject(promise)
        guard let value = dict["value"] as? [String: Any],
              let proofs = dict["proofs"] as? [[String: Any]],
              let first = proofs.first,
              let id = first["id"] as? String,
              let signature = first["signature"] as? String else {
            throw SignerError.badResult
        }
        return SignedTransaction(value: value, proofID: id, signature: signature)
    }

    struct SignedTransaction: Sendable {
        let value: [String: Any]
        let proofID: String
        let signature: String
    }

    private static let textCodecPolyfill = """
    if (typeof TextEncoder === 'undefined') {
      globalThis.TextEncoder = function(){};
      globalThis.TextEncoder.prototype.encode = function(str){
        str = String(str); var out = [];
        for (var i=0;i<str.length;i++){
          var c = str.charCodeAt(i);
          if (c < 0x80) out.push(c);
          else if (c < 0x800){ out.push(0xc0|(c>>6), 0x80|(c&0x3f)); }
          else if (c >= 0xd800 && c < 0xdc00){
            var c2 = str.charCodeAt(++i);
            var cp = 0x10000 + ((c & 0x3ff)<<10) + (c2 & 0x3ff);
            out.push(0xf0|(cp>>18), 0x80|((cp>>12)&0x3f), 0x80|((cp>>6)&0x3f), 0x80|(cp&0x3f));
          } else { out.push(0xe0|(c>>12), 0x80|((c>>6)&0x3f), 0x80|(c&0x3f)); }
        }
        return new Uint8Array(out);
      };
    }
    if (typeof TextDecoder === 'undefined') {
      globalThis.TextDecoder = function(){};
      globalThis.TextDecoder.prototype.decode = function(buf){
        var b = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
        var out = '', i = 0;
        while (i < b.length){
          var c = b[i++];
          if (c < 0x80) out += String.fromCharCode(c);
          else if (c < 0xe0) out += String.fromCharCode(((c&0x1f)<<6)|(b[i++]&0x3f));
          else if (c < 0xf0) out += String.fromCharCode(((c&0x0f)<<12)|((b[i++]&0x3f)<<6)|(b[i++]&0x3f));
          else { var cp = (((c&0x07)<<18)|((b[i++]&0x3f)<<12)|((b[i++]&0x3f)<<6)|(b[i++]&0x3f)) - 0x10000;
            out += String.fromCharCode(0xd800+(cp>>10), 0xdc00+(cp&0x3ff)); }
        }
        return out;
      };
    }
    """

    // MARK: - Context setup

    private func wallet() throws -> JSValue {
        let ctx = try ensureContext()
        guard let echo = ctx.objectForKeyedSubscript("EchoWallet"), !echo.isUndefined, !echo.isNull else {
            throw SignerError.globalMissing
        }
        return echo
    }

    private func ensureContext() throws -> JSContext {
        if let context { return context }
        guard let ctx = JSContext() else { throw SignerError.contextUnavailable }
        ctx.exceptionHandler = { _, exception in
            // Logged; the calling method also checks ctx.exception explicitly.
            if let exception { print("StargazerSigner JS exception: \(exception)") }
        }
        injectHostGlobals(ctx)

        let source: String
        if let override = bundleSourceOverride {
            source = override
        } else {
            guard let url = Bundle.main.url(forResource: bundleResource, withExtension: bundleExtension),
                  let loaded = try? String(contentsOf: url, encoding: .utf8) else {
                throw SignerError.bundleMissing
            }
            source = loaded
        }
        ctx.evaluateScript(source)
        if let exc = ctx.exception {
            throw SignerError.js(exc.toString() ?? "bundle load failed")
        }
        context = ctx
        return ctx
    }

    /// Injects the globals bare JavaScriptCore lacks but the bundle expects:
    /// a minimal `process` and a secure `crypto.getRandomValues`
    /// (SecRandomCopyBytes). Mirrors gen-vector.js's sandbox.
    private func injectHostGlobals(_ ctx: JSContext) {
        ctx.evaluateScript(
            "var process = { env: {}, browser: true, version: '', versions: {}, nextTick: function(cb){ setTimeout(cb, 0); } };"
        )
        // Bare JavaScriptCore lacks TextEncoder/TextDecoder, which dag4 and the
        // brotli wasm glue use. Inject minimal UTF-8 implementations. (We do NOT
        // define `self`, which would route @noble through an absent crypto.subtle.)
        ctx.evaluateScript(Self.textCodecPolyfill)

        let fill: @convention(block) (JSValue) -> JSValue = { typedArray in
            let length = Int(typedArray.forProperty("length")?.toInt32() ?? 0)
            if length > 0 {
                var bytes = [UInt8](repeating: 0, count: length)
                _ = SecRandomCopyBytes(kSecRandomDefault, length, &bytes)
                for i in 0..<length {
                    typedArray.setObject(Int(bytes[i]), atIndexedSubscript: i)
                }
            }
            return typedArray
        }
        let cryptoObj = JSValue(newObjectIn: ctx)
        cryptoObj?.setObject(fill, forKeyedSubscript: "getRandomValues" as NSString)
        ctx.setObject(cryptoObj, forKeyedSubscript: "crypto" as NSString)
    }

    // MARK: - Helpers

    private func account(method: String, args: [Any]) throws -> WalletAccount {
        let echo = try wallet()
        guard let v = echo.invokeMethod(method, withArguments: args), !v.isUndefined, !v.isNull,
              let dict = v.toDictionary() as? [String: Any],
              let pub = dict["publicKey"] as? String,
              let addr = dict["address"] as? String else {
            throw SignerError.badResult
        }
        let priv = (dict["privateKey"] as? String) ?? (args.first as? String) ?? ""
        return WalletAccount(privateKey: priv, publicKey: pub, address: addr)
    }

    /// Bridges a JS Promise<string> to async/await via then/catch callbacks.
    private func resolvePromise(_ promise: JSValue?) async throws -> String {
        let value = try await settle(promise)
        return value?.toString() ?? ""
    }

    /// Bridges a JS Promise<object> to a Swift dictionary.
    private func resolvePromiseObject(_ promise: JSValue?) async throws -> [String: Any] {
        let value = try await settle(promise)
        guard let dict = value?.toDictionary() as? [String: Any] else {
            throw SignerError.badResult
        }
        return dict
    }

    private func settle(_ promise: JSValue?) async throws -> JSValue? {
        guard let promise, !promise.isUndefined else { throw SignerError.badResult }
        return try await withCheckedThrowingContinuation { continuation in
            let resumed = ResumeOnce()
            let onFulfilled: @convention(block) (JSValue?) -> Void = { value in
                guard resumed.claim() else { return }
                continuation.resume(returning: value)
            }
            let onRejected: @convention(block) (JSValue?) -> Void = { error in
                guard resumed.claim() else { return }
                continuation.resume(throwing: SignerError.js(error?.toString() ?? "promise rejected"))
            }
            // JavaScriptCore bridges Objective-C blocks to JS functions; pass
            // both handlers to Promise.then(onFulfilled, onRejected).
            promise.invokeMethod("then", withArguments: [onFulfilled, onRejected])
        }
    }
}

/// One-shot latch so a Promise that somehow settles twice can't resume the
/// continuation more than once.
private final class ResumeOnce: @unchecked Sendable {
    private let lock = NSLock()
    private var done = false
    func claim() -> Bool {
        lock.lock(); defer { lock.unlock() }
        if done { return false }
        done = true
        return true
    }
}
#endif
