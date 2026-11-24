// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// Spike: Echo's Data L1 rules as a contract (docs/BESU_MIGRATION_PLAN.md).
/// Mirrors metagraph/modules/shared_data Validations: a Merkle root (T5) or a trust commitment (T6)
/// is a 32-byte hash and nothing else. The type system enforces that here: there is no field that
/// could carry an email, a phone number or a DID, so the zero-PII rule needs no pattern matching.
/// Only listed signers (the gateway's key) may write, the equivalent of DATA_AUTHORIZED_SIGNERS.
contract EchoAnchor {
    address public owner;                          // slot 0
    mapping(address => bool) public signers;       // slot 1
    /// root => (block number << 32) | leaf count. Zero means "not anchored". Slot 2: offline
    /// verifiers prove a root by proving this storage entry against a block's state root.
    mapping(bytes32 => uint256) public anchoredAt; // slot 2
    /// commitment => epoch. Slot 3.
    mapping(bytes32 => uint256) public commitmentEpoch;

    event RootAnchored(bytes32 indexed root, uint32 leafCount, address indexed signer);
    event CommitmentAnchored(bytes32 indexed commitment, uint64 epoch, address indexed signer);
    event SignerSet(address indexed signer, bool allowed);

    error NotOwner();
    error NotAuthorisedSigner(address signer);
    error EmptyValue();
    error AlreadyAnchored(bytes32 value);

    modifier onlyOwner() { if (msg.sender != owner) revert NotOwner(); _; }
    modifier onlySigner() { if (!signers[msg.sender]) revert NotAuthorisedSigner(msg.sender); _; }

    constructor(address firstSigner) {
        owner = msg.sender;
        signers[firstSigner] = true;
        emit SignerSet(firstSigner, true);
    }

    function setSigner(address signer, bool allowed) external onlyOwner {
        signers[signer] = allowed;
        emit SignerSet(signer, allowed);
    }

    function anchorRoot(bytes32 root, uint32 leafCount) external onlySigner {
        if (root == bytes32(0) || leafCount == 0) revert EmptyValue();
        if (anchoredAt[root] != 0) revert AlreadyAnchored(root);
        anchoredAt[root] = (block.number << 32) | leafCount;
        emit RootAnchored(root, leafCount, msg.sender);
    }

    function anchorCommitment(bytes32 commitment, uint64 epoch) external onlySigner {
        if (commitment == bytes32(0) || epoch == 0) revert EmptyValue();
        if (commitmentEpoch[commitment] != 0) revert AlreadyAnchored(commitment);
        commitmentEpoch[commitment] = epoch;
        emit CommitmentAnchored(commitment, epoch, msg.sender);
    }
}
