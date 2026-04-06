// Tests/WalletTests.swift
// Comprehensive wallet, evidence, and design system tests

import XCTest
@testable import Echo

// MARK: - Wallet Types Tests

final class WalletTypesTests: XCTestCase {

    func testBalanceInfo_staked() {
        let balance = BalanceInfo(total: 1000, available: 600)
        XCTAssertEqual(balance.staked, 400)
    }

    func testStakingTier_allCases() {
        XCTAssertEqual(StakingTier.allCases.count, 4)
        XCTAssertEqual(StakingTier.bronze.durationDays, 30)
        XCTAssertEqual(StakingTier.silver.durationDays, 90)
        XCTAssertEqual(StakingTier.gold.durationDays, 180)
        XCTAssertEqual(StakingTier.platinum.durationDays, 365)
    }

    func testStakingTier_apr() {
        XCTAssertEqual(StakingTier.bronze.apr, 5.0)
        XCTAssertEqual(StakingTier.silver.apr, 8.0)
        XCTAssertEqual(StakingTier.gold.apr, 12.0)
        XCTAssertEqual(StakingTier.platinum.apr, 15.0)
    }

    func testStakingTier_displayName() {
        XCTAssertEqual(StakingTier.bronze.displayName, "Bronze")
        XCTAssertEqual(StakingTier.platinum.displayName, "Platinum")
    }

    func testTokenLockPosition_isLocked() {
        let locked = TokenLockPosition(
            id: "1", amount: 100, tier: "gold",
            lockedUntil: Date().addingTimeInterval(86400),
            vestingType: nil, originalAmount: 100,
            cliffDate: nil, cliffCompleted: false,
            vestedAmount: 0, withdrawableAmount: 0,
            nextUnlockDate: nil, nextUnlockAmount: 0,
            delegatedTo: nil
        )
        XCTAssertTrue(locked.isLocked)

        let unlocked = TokenLockPosition(
            id: "2", amount: 100, tier: "bronze",
            lockedUntil: Date().addingTimeInterval(-86400),
            vestingType: nil, originalAmount: 100,
            cliffDate: nil, cliffCompleted: false,
            vestedAmount: 0, withdrawableAmount: 0,
            nextUnlockDate: nil, nextUnlockAmount: 0,
            delegatedTo: nil
        )
        XCTAssertFalse(unlocked.isLocked)
    }

    func testTokenLockPosition_isFounderVesting() {
        let founderLock = TokenLockPosition(
            id: "1", amount: 60000000, tier: "platinum",
            lockedUntil: Date().addingTimeInterval(86400 * 365),
            vestingType: "founder", originalAmount: 60000000,
            cliffDate: nil, cliffCompleted: false,
            vestedAmount: 0, withdrawableAmount: 0,
            nextUnlockDate: nil, nextUnlockAmount: 0,
            delegatedTo: nil
        )
        XCTAssertTrue(founderLock.isFounderVesting)

        let userLock = TokenLockPosition(
            id: "2", amount: 100, tier: "gold",
            lockedUntil: Date(), vestingType: nil,
            originalAmount: 100, cliffDate: nil,
            cliffCompleted: false, vestedAmount: 0,
            withdrawableAmount: 0, nextUnlockDate: nil,
            nextUnlockAmount: 0, delegatedTo: nil
        )
        XCTAssertFalse(userLock.isFounderVesting)
    }

    func testRewardCapEntry_progress() {
        let entry = RewardCapEntry(earned: 50, cap: 100)
        XCTAssertEqual(entry.progress, 0.5, accuracy: 0.01)
        XCTAssertEqual(entry.remaining, 50)

        let full = RewardCapEntry(earned: 100, cap: 100)
        XCTAssertEqual(full.progress, 1.0, accuracy: 0.01)
        XCTAssertEqual(full.remaining, 0)

        let zeroCap = RewardCapEntry(earned: 0, cap: 0)
        XCTAssertEqual(zeroCap.progress, 0)
    }

    func testStargazerError_descriptions() {
        XCTAssertNotNil(StargazerError.notInitialized.errorDescription)
        XCTAssertNotNil(StargazerError.transactionFailed("test").errorDescription)
        XCTAssertNotNil(StargazerError.insufficientBalance.errorDescription)
        XCTAssertNotNil(StargazerError.walletCreationFailed.errorDescription)
    }

    func testVestingState_equality() {
        let date = Date()
        let v1 = VestingState(
            role: "Founder", totalAllocated: 60000000,
            vested: 15000000, locked: 45000000,
            withdrawable: 15000000, nextUnlockAmount: 1250000,
            nextUnlockDate: date, cliffDate: date,
            cliffCompleted: true, vestingPercent: 25.0
        )
        let v2 = VestingState(
            role: "Founder", totalAllocated: 60000000,
            vested: 15000000, locked: 45000000,
            withdrawable: 15000000, nextUnlockAmount: 1250000,
            nextUnlockDate: date, cliffDate: date,
            cliffCompleted: true, vestingPercent: 25.0
        )
        XCTAssertEqual(v1, v2)
    }
}

// MARK: - Wallet ViewModel Tests

@MainActor
final class WalletViewModelTests: XCTestCase {

    func testLoadWallet_success() async {
        let api = MockWalletAPIClient()
        let vm = WalletViewModel(api: api)

        await vm.loadWallet()

        XCTAssertNotNil(vm.walletState)
        XCTAssertEqual(vm.walletState?.totalBalance, 1250)
        XCTAssertEqual(vm.walletState?.available, 750)
        XCTAssertNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
    }

    func testLoadWallet_error() async {
        let api = MockWalletAPIClient()
        api.shouldError = true
        let vm = WalletViewModel(api: api)

        await vm.loadWallet()

        XCTAssertNil(vm.walletState)
        XCTAssertNotNil(vm.errorMessage)
    }

    func testStakeEcho_invalidAmount() async {
        let api = MockWalletAPIClient()
        let vm = WalletViewModel(api: api)
        vm.stakeAmount = ""

        await vm.stakeEcho()

        XCTAssertNotNil(vm.errorMessage)
    }

    func testStakeEcho_success() async {
        let api = MockWalletAPIClient()
        let vm = WalletViewModel(api: api)
        vm.stakeAmount = "100"
        vm.selectedTier = .gold

        await vm.stakeEcho()

        XCTAssertEqual(vm.stakeAmount, "")
        XCTAssertFalse(vm.isStaking)
    }

    func testLoadValidators() async {
        let api = MockWalletAPIClient()
        api.validators = [
            ValidatorInfo(
                id: "v1", address: "DAG_v1",
                uptimePercent: 99.5, commissionPercent: 5.0,
                totalDelegated: 1000000, delegatorCount: 50,
                layer: "currency_l1", estimatedAPR: 12.0
            )
        ]
        let vm = WalletViewModel(api: api)

        await vm.loadValidators()

        XCTAssertEqual(vm.validators.count, 1)
        XCTAssertEqual(vm.validators.first?.id, "v1")
    }

    func testDefaultTierIsBronze() {
        let api = MockWalletAPIClient()
        let vm = WalletViewModel(api: api)
        XCTAssertEqual(vm.selectedTier, .bronze)
    }
}

// MARK: - Delivery Status Tests

final class DeliveryStatusTests: XCTestCase {

    func testSortOrder() {
        XCTAssertTrue(DeliveryStatus.sending < DeliveryStatus.sent)
        XCTAssertTrue(DeliveryStatus.sent < DeliveryStatus.delivered)
        XCTAssertTrue(DeliveryStatus.delivered < DeliveryStatus.read)
        XCTAssertTrue(DeliveryStatus.read < DeliveryStatus.anchored)
        XCTAssertTrue(DeliveryStatus.anchored < DeliveryStatus.verified)
        XCTAssertTrue(DeliveryStatus.failed < DeliveryStatus.sending)
    }

    func testHasVerificationURL() {
        XCTAssertTrue(DeliveryStatus.verified.hasVerificationURL)
        XCTAssertFalse(DeliveryStatus.sent.hasVerificationURL)
        XCTAssertFalse(DeliveryStatus.anchored.hasVerificationURL)
    }

    func testAllStatusesHaveIcons() {
        for status in [DeliveryStatus.sending, .sent, .delivered, .read, .failed, .anchored, .verified] {
            XCTAssertFalse(status.icon.isEmpty, "\(status) should have an icon")
            XCTAssertFalse(status.displayLabel.isEmpty, "\(status) should have a display label")
        }
    }

    func testCodable() throws {
        let original = DeliveryStatus.anchored
        let data = try JSONEncoder().encode(original)
        let decoded = try JSONDecoder().decode(DeliveryStatus.self, from: data)
        XCTAssertEqual(decoded, original)
    }
}

// MARK: - Evidence Bridge Tests

final class EvidenceBridgeTests: XCTestCase {

    func testFingerprintMedia() async throws {
        let api = MockEvidenceAPI()
        let bridge = DigitalEvidenceBridge(api: api)

        let data = "test media content".data(using: .utf8)!
        let result = try await bridge.fingerprintMedia(data, messageId: "msg_123")

        XCTAssertEqual(result.eventId, "evt_msg_123")
        XCTAssertTrue(result.verificationUrl.contains("digitalevidence"))
    }

    func testFingerprintMessage() async throws {
        let api = MockEvidenceAPI()
        let bridge = DigitalEvidenceBridge(api: api)

        let result = try await bridge.fingerprintMessage("hello world", messageId: "msg_456")
        XCTAssertEqual(result.eventId, "evt_msg_456")
    }

    func testVerificationURL() async {
        let api = MockEvidenceAPI()
        let bridge = DigitalEvidenceBridge(api: api)

        let url = await bridge.verificationURL(eventId: "evt_123")
        XCTAssertNotNil(url)
        XCTAssertEqual(url?.host, "digitalevidence.constellationnetwork.io")
        XCTAssertTrue(url?.path.contains("evt_123") ?? false)
    }

    func testCheckVerification() async throws {
        let api = MockEvidenceAPI()
        let bridge = DigitalEvidenceBridge(api: api)

        let status = try await bridge.checkVerification(eventId: "evt_test")
        XCTAssertEqual(status.status, "verified")
        XCTAssertEqual(status.eventId, "evt_test")
    }

    func testFingerprintMedia_error() async {
        let api = MockEvidenceAPI()
        api.shouldError = true
        let bridge = DigitalEvidenceBridge(api: api)

        let data = "test".data(using: .utf8)!
        do {
            _ = try await bridge.fingerprintMedia(data, messageId: "msg_err")
            XCTFail("Expected error")
        } catch {
            // Expected
        }
    }
}

// MARK: - Evidence Result Codable Tests

final class EvidenceResultCodableTests: XCTestCase {

    func testEvidenceResult_decodable() throws {
        let json = """
        {
            "event_id": "evt_abc",
            "verification_url": "https://example.com/verify/evt_abc",
            "timestamp": "2026-01-15T10:30:00Z"
        }
        """.data(using: .utf8)!

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        let result = try decoder.decode(EvidenceResult.self, from: json)
        XCTAssertEqual(result.eventId, "evt_abc")
        XCTAssertTrue(result.verificationUrl.contains("evt_abc"))
    }

    func testEvidenceVerificationStatus_decodable() throws {
        let json = """
        {
            "event_id": "evt_xyz",
            "status": "verified",
            "explorer_url": "https://example.com/verify/evt_xyz"
        }
        """.data(using: .utf8)!

        let status = try JSONDecoder().decode(EvidenceVerificationStatus.self, from: json)
        XCTAssertEqual(status.status, "verified")
        XCTAssertEqual(status.eventId, "evt_xyz")
        XCTAssertNotNil(status.explorerUrl)
    }
}

// MARK: - Format ECHO Tests

final class FormatEchoTests: XCTestCase {

    func testFormatEcho_wholeNumber() {
        XCTAssertEqual(formatEcho(1250), "1,250.00")
    }

    func testFormatEcho_decimal() {
        XCTAssertEqual(formatEcho(Decimal(string: "1250.50")!), "1,250.50")
    }

    func testFormatEcho_zero() {
        XCTAssertEqual(formatEcho(0), "0.00")
    }
}

// MARK: - MockWalletAPIClient Tests

final class MockWalletAPIClientTests: XCTestCase {

    func testCreateWallet() async throws {
        let mock = MockWalletAPIClient()
        let info = try await mock.createWallet()
        XCTAssertFalse(info.address.isEmpty)
        XCTAssertFalse(info.publicKey.isEmpty)
    }

    func testGetBalance() async throws {
        let mock = MockWalletAPIClient()
        let balance = try await mock.getBalance()
        XCTAssertEqual(balance.total, 1250)
        XCTAssertEqual(balance.available, 750)
    }

    func testSubmitTokenLock_error() async {
        let mock = MockWalletAPIClient()
        mock.shouldError = true
        do {
            _ = try await mock.submitTokenLock(amount: 100, tier: .gold)
            XCTFail("Expected error")
        } catch {
            XCTAssertTrue(error is StargazerError)
        }
    }
}
