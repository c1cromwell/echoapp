// SecurityTests/SecureEnclaveSecurityTests.swift
//
// Clean, isolated tests for WO-208 / WO-211 / WO-223 / WO-224.
// Kept in a separate SPM test target so pre-existing EchoTests compilation
// issues do not prevent these from running.

import XCTest
import CryptoKit
@testable import Echo

// MARK: - WO-223: X25519 Key Agreement

final class X25519KeyAgreementTests: XCTestCase {

    func testPerformKeyAgreement_outputIs32Bytes() throws {
        let manager = SecureEnclaveManager()
        let theirKey = Curve25519.KeyAgreement.PrivateKey()
        let key = try manager.performKeyAgreement(
            theirPublicKeyData: theirKey.publicKey.rawRepresentation,
            contextInfo: "test-session"
        )
        let keyBytes = key.withUnsafeBytes { Data($0) }
        XCTAssertEqual(keyBytes.count, 32, "Session key must be 32 bytes (AES-256)")
    }

    func testPerformKeyAgreement_differentContextProducesDifferentKey() throws {
        let manager = SecureEnclaveManager()
        let theirKey = Curve25519.KeyAgreement.PrivateKey()
        let raw = theirKey.publicKey.rawRepresentation

        let k1 = try manager.performKeyAgreement(theirPublicKeyData: raw, contextInfo: "ctx-A")
        let k2 = try manager.performKeyAgreement(theirPublicKeyData: raw, contextInfo: "ctx-B")
        let b1 = k1.withUnsafeBytes { Data($0) }
        let b2 = k2.withUnsafeBytes { Data($0) }
        XCTAssertNotEqual(b1, b2, "Different contexts must produce different keys")
    }

    func testPerformKeyAgreement_wrongLengthThrows() {
        let manager = SecureEnclaveManager()
        // X25519 public keys must be exactly 32 bytes; 31 bytes is invalid
        let badKey = Data(repeating: 0x01, count: 31)
        XCTAssertThrowsError(
            try manager.performKeyAgreement(theirPublicKeyData: badKey, contextInfo: "x"),
            "Wrong-length key must throw"
        )
    }
}

// MARK: - WO-224: Storage Key Derivation

final class StorageKeyDerivationTests: XCTestCase {

    func testDeriveStorageKey_is32Bytes() {
        let manager = SecureEnclaveManager()
        let key = manager.deriveStorageKey(keyId: "echo-identity-key")
        let bytes = key.withUnsafeBytes { Data($0) }
        XCTAssertEqual(bytes.count, 32)
    }

    func testDeriveStorageKey_deterministicWithinMonth() {
        let manager = SecureEnclaveManager()
        let k1 = manager.deriveStorageKey(keyId: "echo-identity-key")
        let k2 = manager.deriveStorageKey(keyId: "echo-identity-key")
        let b1 = k1.withUnsafeBytes { Data($0) }
        let b2 = k2.withUnsafeBytes { Data($0) }
        XCTAssertEqual(b1, b2, "Storage key must be deterministic within a calendar month")
    }

    func testDeriveStorageKey_differsByKeyId() {
        let manager = SecureEnclaveManager()
        let k1 = manager.deriveStorageKey(keyId: "key-A")
        let k2 = manager.deriveStorageKey(keyId: "key-B")
        let b1 = k1.withUnsafeBytes { Data($0) }
        let b2 = k2.withUnsafeBytes { Data($0) }
        XCTAssertNotEqual(b1, b2, "Storage keys must differ when key ID differs")
    }
}

// MARK: - WO-211: Biometric Lockout

final class BiometricLockoutTests: XCTestCase {

    override func setUp() {
        super.setUp()
        UserDefaults.standard.removeObject(forKey: "com.echo.biometric.failureCount")
        UserDefaults.standard.removeObject(forKey: "com.echo.biometric.lockoutUntil")
    }

    func testInitialState_isAllowed() {
        let state = SecureEnclaveManager().currentLockState()
        XCTAssertFalse(state.isLocked)
        if case .allowed = state { } else {
            XCTFail("Expected .allowed, got \(state)")
        }
    }

    func testFourFailures_stillAllowed() {
        let m = SecureEnclaveManager()
        for _ in 0..<4 { _ = m.recordBiometricFailure() }
        let state = m.currentLockState()
        if case .allowed = state { } else {
            XCTFail("4 failures should still be allowed, got \(state)")
        }
    }

    func testFiveFailures_requiresPasscode() {
        let m = SecureEnclaveManager()
        var last: BiometricLockState = .allowed(failureCount: 0)
        for _ in 0..<5 { last = m.recordBiometricFailure() }
        if case .requiresPasscode = last { } else {
            XCTFail("5th failure should trigger requiresPasscode, got \(last)")
        }
        XCTAssertTrue(last.isLocked)
    }

    func testTenFailures_hardLockout() {
        let m = SecureEnclaveManager()
        var last: BiometricLockState = .allowed(failureCount: 0)
        for _ in 0..<10 { last = m.recordBiometricFailure() }
        if case .hardLocked(let until) = last {
            XCTAssertGreaterThan(until, Date())
            let remaining = last.remainingLockout ?? 0
            XCTAssertGreaterThan(remaining, 0)
            XCTAssertLessThanOrEqual(remaining, 15 * 60 + 2)
        } else {
            XCTFail("10 failures should trigger hardLocked, got \(last)")
        }
    }

    func testSuccess_resetsCounter() {
        let m = SecureEnclaveManager()
        for _ in 0..<5 { _ = m.recordBiometricFailure() }
        m.recordBiometricSuccess()
        let state = m.currentLockState()
        if case .allowed(let count) = state {
            XCTAssertEqual(count, 0)
        } else {
            XCTFail("After success, state should be .allowed, got \(state)")
        }
    }

    func testRemainingLockout_nilForRequiresPasscode() {
        let state = BiometricLockState.requiresPasscode(failureCount: 5)
        XCTAssertNil(state.remainingLockout)
    }
}

// MARK: - WO-208: PrivacySettings Model

final class PrivacySettingsTests: XCTestCase {

    func testDefaultPrivacySettings_allTrue() {
        let settings = PrivacySettings()
        XCTAssertTrue(settings.showLastSeen)
        XCTAssertTrue(settings.showOnlineStatus)
        XCTAssertTrue(settings.showProfilePicture)
        XCTAssertTrue(settings.showStatusMessage)
        XCTAssertTrue(settings.allowCalls)
        XCTAssertTrue(settings.contactDiscoveryOptIn)
        XCTAssertTrue(settings.showEncryptionIndicator)
    }

    func testPrivacySettings_canBeDisabled() {
        var settings = PrivacySettings()
        settings.showLastSeen = false
        settings.showOnlineStatus = false
        XCTAssertFalse(settings.showLastSeen)
        XCTAssertFalse(settings.showOnlineStatus)
        // Other fields unaffected
        XCTAssertTrue(settings.allowCalls)
    }

    func testPrivacySettings_codable() throws {
        var settings = PrivacySettings()
        settings.showLastSeen = false
        settings.allowCalls = false

        let data = try JSONEncoder().encode(settings)
        let decoded = try JSONDecoder().decode(PrivacySettings.self, from: data)

        XCTAssertEqual(decoded.showLastSeen, false)
        XCTAssertEqual(decoded.allowCalls, false)
        XCTAssertEqual(decoded.showEncryptionIndicator, true)
    }
}

// MARK: - WO-223: Background Purge

final class BackgroundPurgeTests: XCTestCase {

    func testPurgeOnBackground_doesNotCrash() async {
        let manager = SecureEnclaveManager()
        await manager.purgeOnBackground()
        // Verify second call is idempotent
        await manager.purgeOnBackground()
    }
}

// MARK: - WO-224: LocalDatabase storage key lifecycle

final class LocalDatabaseStorageKeyTests: XCTestCase {

    func testUnlock_clearsIsLocked() async {
        // Use shared instance; lock first to ensure clean state for this test
        await LocalDatabase.shared.lockStorage()
        let key = SecureEnclaveManager().deriveStorageKey(keyId: "test-key-unlock")
        await LocalDatabase.shared.unlock(storageKey: key)
        let locked = await LocalDatabase.shared.isLocked
        XCTAssertFalse(locked, "After unlock(), isLocked must be false")
        // Restore: lock for other tests
        await LocalDatabase.shared.lockStorage()
    }

    func testLockStorage_setsIsLocked() async {
        let key = SecureEnclaveManager().deriveStorageKey(keyId: "test-key-lock")
        await LocalDatabase.shared.unlock(storageKey: key)
        await LocalDatabase.shared.lockStorage()
        let locked = await LocalDatabase.shared.isLocked
        XCTAssertTrue(locked, "After lockStorage(), isLocked must be true")
    }

    func testUnlockLockCycle_isIdempotent() async {
        for _ in 0..<3 {
            let key = SecureEnclaveManager().deriveStorageKey(keyId: "test-cycle")
            await LocalDatabase.shared.unlock(storageKey: key)
            await LocalDatabase.shared.lockStorage()
        }
        let locked = await LocalDatabase.shared.isLocked
        XCTAssertTrue(locked)
    }
}

// MARK: - WO-211: BiometricLockState additional edge cases

final class BiometricLockStateEdgeCaseTests: XCTestCase {

    func testAllowed_isNotLocked() {
        XCTAssertFalse(BiometricLockState.allowed(failureCount: 0).isLocked)
        XCTAssertFalse(BiometricLockState.allowed(failureCount: 4).isLocked)
    }

    func testRequiresPasscode_isLocked() {
        XCTAssertTrue(BiometricLockState.requiresPasscode(failureCount: 5).isLocked)
    }

    func testHardLocked_futureDate_isLocked() {
        let state = BiometricLockState.hardLocked(until: Date().addingTimeInterval(300))
        XCTAssertTrue(state.isLocked)
    }

    func testHardLocked_pastDate_isNotLocked() {
        let state = BiometricLockState.hardLocked(until: Date().addingTimeInterval(-1))
        XCTAssertFalse(state.isLocked, "Expired hard lockout must not be locked")
    }

    func testRemainingLockout_positive_for_future() {
        let state = BiometricLockState.hardLocked(until: Date().addingTimeInterval(300))
        guard let remaining = state.remainingLockout else {
            XCTFail("hardLocked must have a remaining lockout")
            return
        }
        XCTAssertGreaterThan(remaining, 0)
        XCTAssertLessThanOrEqual(remaining, 301)
    }

    func testRemainingLockout_zero_for_expired() {
        let state = BiometricLockState.hardLocked(until: Date().addingTimeInterval(-1))
        XCTAssertEqual(state.remainingLockout, 0)
    }
}

// MARK: - WO-208: EnhancedPrivacySettings new fields

final class EnhancedPrivacySettingsTests: XCTestCase {

    func testDefaultsForNewWO208Fields() {
        let s = EnhancedPrivacySettings()
        XCTAssertTrue(s.showLastSeen)
        XCTAssertTrue(s.showProfilePicture)
        XCTAssertTrue(s.showStatusMessage)
        XCTAssertTrue(s.contactDiscoveryOptIn)
        XCTAssertTrue(s.showEncryptionIndicator)
    }

    func testAllWO208FieldsToggleable() {
        var s = EnhancedPrivacySettings()
        s.showLastSeen = false
        s.showProfilePicture = false
        s.showStatusMessage = false
        s.contactDiscoveryOptIn = false
        s.showEncryptionIndicator = false

        XCTAssertFalse(s.showLastSeen)
        XCTAssertFalse(s.showProfilePicture)
        XCTAssertFalse(s.showStatusMessage)
        XCTAssertFalse(s.contactDiscoveryOptIn)
        XCTAssertFalse(s.showEncryptionIndicator)
    }

    func testCodableRoundTrip() throws {
        var s = EnhancedPrivacySettings()
        s.showLastSeen = false
        s.showEncryptionIndicator = true

        let data = try JSONEncoder().encode(s)
        let decoded = try JSONDecoder().decode(EnhancedPrivacySettings.self, from: data)

        XCTAssertEqual(decoded.showLastSeen, false)
        XCTAssertEqual(decoded.showEncryptionIndicator, true)
        XCTAssertEqual(decoded.contactDiscoveryOptIn, true)
    }
}
