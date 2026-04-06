// Tests/GovernanceTests.swift
// Unit tests for governance models, service, and weight calculations

import XCTest
@testable import Echo

// MARK: - Governance Tier Tests

final class GovernanceTierTests: XCTestCase {

    func testTier1ZeroMultiplier() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 1), 0.0)
    }

    func testTier2HalfMultiplier() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 2), 0.5)
    }

    func testTier3StandardMultiplier() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 3), 1.0)
    }

    func testTier4EnhancedMultiplier() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 4), 1.5)
    }

    func testTier5MaxMultiplier() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 5), 2.0)
    }

    func testInvalidTierReturnsZero() {
        XCTAssertEqual(GovernanceTier.multiplier(for: 0), 0.0)
        XCTAssertEqual(GovernanceTier.multiplier(for: 6), 0.0)
    }

    func testCanVoteTier2WithStake() {
        XCTAssertTrue(GovernanceTier.canVote(tier: 2, totalStaked: 1000))
    }

    func testCannotVoteTier1() {
        XCTAssertFalse(GovernanceTier.canVote(tier: 1, totalStaked: 1_000_000))
    }

    func testCannotVoteZeroStake() {
        XCTAssertFalse(GovernanceTier.canVote(tier: 5, totalStaked: 0))
    }

    func testAntiPlutocraticDesign() {
        // Tier 1 whale with 50M ECHO → zero governance power
        XCTAssertFalse(GovernanceTier.canVote(tier: 1, totalStaked: 50_000_000))
        XCTAssertEqual(GovernanceTier.multiplier(for: 1), 0.0)
    }
}

// MARK: - Proposal Model Tests

final class ProposalModelTests: XCTestCase {

    func testProposalTypesRawValues() {
        XCTAssertEqual(ProposalType.protocolUpgrade.rawValue, "protocol_upgrade")
        XCTAssertEqual(ProposalType.treasuryAllocation.rawValue, "treasury_allocation")
        XCTAssertEqual(ProposalType.parameterChange.rawValue, "parameter_change")
        XCTAssertEqual(ProposalType.boardElection.rawValue, "board_election")
    }

    func testProposalTypeDisplayNames() {
        XCTAssertEqual(ProposalType.protocolUpgrade.displayName, "Protocol Upgrade")
        XCTAssertEqual(ProposalType.boardElection.displayName, "Board Election")
    }

    func testThresholdRequiredPercent() {
        XCTAssertEqual(ThresholdType.simpleMajority.requiredPercent, 51)
        XCTAssertEqual(ThresholdType.supermajority67.requiredPercent, 67)
        XCTAssertEqual(ThresholdType.supermajority75.requiredPercent, 75)
    }

    func testProposalStatusValues() {
        XCTAssertEqual(ProposalStatus.active.rawValue, "active")
        XCTAssertEqual(ProposalStatus.passed.rawValue, "passed")
        XCTAssertEqual(ProposalStatus.failed.rawValue, "failed")
        XCTAssertEqual(ProposalStatus.executed.rawValue, "executed")
    }

    func testVoteValueSystemImages() {
        XCTAssertEqual(VoteValue.for.systemImage, "checkmark.circle.fill")
        XCTAssertEqual(VoteValue.against.systemImage, "xmark.circle.fill")
        XCTAssertEqual(VoteValue.abstain.systemImage, "minus.circle.fill")
    }

    func testVoteValueDisplayNames() {
        XCTAssertEqual(VoteValue.for.displayName, "For")
        XCTAssertEqual(VoteValue.against.displayName, "Against")
        XCTAssertEqual(VoteValue.abstain.displayName, "Abstain")
    }
}

// MARK: - Proposal Tally Tests

final class ProposalTallyTests: XCTestCase {

    func testTallyPercentCalculation() {
        let tally = ProposalTally(
            proposalId: "p1",
            forWeight: 7000,
            againstWeight: 2000,
            abstainWeight: 1000,
            totalWeight: 10000,
            forPercent: 70.0,
            voterCount: 10,
            passed: true
        )
        XCTAssertEqual(tally.againstPercent, 20.0)
        XCTAssertEqual(tally.abstainPercent, 10.0)
    }

    func testTallyZeroTotalWeight() {
        let tally = ProposalTally(
            proposalId: "p1",
            forWeight: 0,
            againstWeight: 0,
            abstainWeight: 0,
            totalWeight: 0,
            forPercent: 0,
            voterCount: 0,
            passed: false
        )
        XCTAssertEqual(tally.againstPercent, 0)
        XCTAssertEqual(tally.abstainPercent, 0)
    }
}

// MARK: - Voting Power Tests

final class VotingPowerTests: XCTestCase {

    func testVotingPowerCodable() throws {
        let power = VotingPower(
            did: "did:dag:test",
            trustTier: 3,
            multiplier: 1.0,
            totalStaked: 50000,
            weight: 50000,
            canVote: true
        )

        let data = try JSONEncoder().encode(power)
        let decoded = try JSONDecoder().decode(VotingPower.self, from: data)
        XCTAssertEqual(decoded, power)
    }

    func testVotingPowerEquatable() {
        let a = VotingPower(did: "a", trustTier: 3, multiplier: 1.0, totalStaked: 100, weight: 100, canVote: true)
        let b = VotingPower(did: "a", trustTier: 3, multiplier: 1.0, totalStaked: 100, weight: 100, canVote: true)
        XCTAssertEqual(a, b)
    }
}

// MARK: - Governance Error Tests

final class GovernanceErrorTests: XCTestCase {

    func testErrorDescriptions() {
        XCTAssertNotNil(GovernanceError.networkError.errorDescription)
        XCTAssertNotNil(GovernanceError.cannotVote.errorDescription)
        XCTAssertNotNil(GovernanceError.alreadyVoted.errorDescription)
        XCTAssertNotNil(GovernanceError.proposalNotFound.errorDescription)
        XCTAssertNotNil(GovernanceError.proposalExpired.errorDescription)
    }
}

// MARK: - Mock Service Tests

#if DEBUG
final class MockGovernanceServiceTests: XCTestCase {

    func testMockReturnsProposals() async throws {
        let mock = MockGovernanceAPIClient()
        mock.proposals = [
            Proposal(
                id: "p1",
                title: "Test Proposal",
                description: "Test",
                type: .parameterChange,
                threshold: .simpleMajority,
                createdBy: "did:dag:test",
                createdAt: Date(),
                endsAt: Date().addingTimeInterval(86400),
                status: .active,
                tally: nil
            )
        ]

        let service = GovernanceService(apiClient: mock)
        let proposals = try await service.fetchProposals()
        XCTAssertEqual(proposals.count, 1)
        XCTAssertEqual(proposals.first?.title, "Test Proposal")
    }

    func testMockReturnsVotingPower() async throws {
        let mock = MockGovernanceAPIClient()
        mock.votingPower = VotingPower(
            did: "did:dag:test",
            trustTier: 5,
            multiplier: 2.0,
            totalStaked: 100000,
            weight: 200000,
            canVote: true
        )

        let service = GovernanceService(apiClient: mock)
        let power = try await service.fetchVotingPower(did: "did:dag:test")
        XCTAssertEqual(power.weight, 200000)
        XCTAssertTrue(power.canVote)
    }

    func testMockSubmitVote() async throws {
        let mock = MockGovernanceAPIClient()
        mock.voteResult = VoteResult(txHash: "hash123", weight: 50000)

        let service = GovernanceService(apiClient: mock)
        let result = try await service.submitVote(proposalId: "p1", value: .for)
        XCTAssertEqual(result.txHash, "hash123")
    }

    func testMockFailureMode() async {
        let mock = MockGovernanceAPIClient()
        mock.shouldFail = true

        let service = GovernanceService(apiClient: mock)
        do {
            _ = try await service.fetchProposals()
            XCTFail("Expected error")
        } catch {
            XCTAssertTrue(error is GovernanceError)
        }
    }

    func testMockCreateProposal() async throws {
        let mock = MockGovernanceAPIClient()
        let service = GovernanceService(apiClient: mock)

        let proposal = try await service.createProposal(
            title: "Test",
            description: "Test",
            type: .protocolUpgrade,
            threshold: .supermajority67,
            endsAt: Date().addingTimeInterval(604800)
        )
        XCTAssertEqual(proposal.status, .active)
    }
}
#endif

// MARK: - Vote Result Codable Tests

final class VoteResultCodableTests: XCTestCase {

    func testVoteResultEncodeDecode() throws {
        let result = VoteResult(txHash: "abc123", weight: 75000)
        let data = try JSONEncoder().encode(result)
        let decoded = try JSONDecoder().decode(VoteResult.self, from: data)
        XCTAssertEqual(decoded.txHash, "abc123")
        XCTAssertEqual(decoded.weight, 75000)
    }
}

// MARK: - Vote Request Codable Tests

final class VoteRequestCodableTests: XCTestCase {

    func testVoteRequestEncodeDecode() throws {
        let req = VoteRequest(did: "did:dag:user", proposalId: "prop1", value: "for")
        let data = try JSONEncoder().encode(req)
        let decoded = try JSONDecoder().decode(VoteRequest.self, from: data)
        XCTAssertEqual(decoded.did, "did:dag:user")
        XCTAssertEqual(decoded.proposalId, "prop1")
        XCTAssertEqual(decoded.value, "for")
    }
}
