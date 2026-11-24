// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// Echo identity records. Every identifier is a 32-byte hash computed by the gateway
/// (internal/besu/ids.go): no DID, username, key or credential id is ever stored as text, and no
/// field here could hold one. Only listed signers (the gateway's key) may write.
///
/// What a hash does and does not hide: a DID is a long random public key, so its hash reveals
/// nothing to someone who does not already know the DID. A username is short and guessable, so
/// its hash can be found by trying names. Username ownership is meant to be checkable by anyone
/// who knows the name; do not treat it as secret.
contract EchoIdentity {
    address public owner;
    mapping(address => bool) public signers;

    /// subject => device key hash => block it was registered in (0 = not registered).
    mapping(bytes32 => mapping(bytes32 => uint64)) public deviceKeyAt;

    /// subject => its latest trust-tier commitment, H(tier || nonce).
    mapping(bytes32 => bytes32) public trustTierOf;
    /// commitment => block it was anchored in. A commitment is anchored at most once.
    mapping(bytes32 => uint64) public trustTierAt;

    /// username hash => subject that owns it. First registration wins.
    mapping(bytes32 => bytes32) public usernameOwner;

    struct Credential {
        bytes32 subject;
        bytes32 issuer;
        bytes32 kind;
        uint64 blockNumber;
    }
    /// credential id hash => issuance record (blockNumber 0 = not issued).
    mapping(bytes32 => Credential) public credentials;

    struct StatusList {
        uint64 sequence;
        bytes32 listHash;
        uint64 blockNumber;
    }
    /// issuer => its current revocation list, by hash. The list itself is served off-chain and
    /// checked against listHash. sequence only ever increases.
    mapping(bytes32 => StatusList) public statusLists;

    /// kind => digest => block. For records that are a single commitment (group metadata,
    /// governance outcomes): the gateway hashes the record and anchors the digest.
    mapping(bytes32 => mapping(bytes32 => uint64)) public attestedAt;

    event DeviceKeyRegistered(bytes32 indexed subject, bytes32 indexed keyHash);
    event TrustTierAnchored(bytes32 indexed subject, bytes32 indexed commitment);
    event UsernameRegistered(bytes32 indexed usernameHash, bytes32 indexed subject);
    event CredentialIssued(bytes32 indexed credentialHash, bytes32 indexed subject, bytes32 indexed issuer, bytes32 kind);
    event StatusListPublished(bytes32 indexed issuer, uint64 sequence, bytes32 listHash);
    event Attested(bytes32 indexed kind, bytes32 indexed digest);
    event SignerSet(address indexed signer, bool allowed);

    error NotOwner();
    error NotAuthorisedSigner(address signer);
    error EmptyValue();
    error AlreadyRecorded(bytes32 value);
    error UsernameTaken(bytes32 usernameHash);
    error StaleSequence(uint64 current, uint64 offered);

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

    function registerDeviceKey(bytes32 subject, bytes32 keyHash) external onlySigner {
        if (subject == bytes32(0) || keyHash == bytes32(0)) revert EmptyValue();
        if (deviceKeyAt[subject][keyHash] != 0) revert AlreadyRecorded(keyHash);
        deviceKeyAt[subject][keyHash] = uint64(block.number);
        emit DeviceKeyRegistered(subject, keyHash);
    }

    function anchorTrustTier(bytes32 subject, bytes32 commitment) external onlySigner {
        if (subject == bytes32(0) || commitment == bytes32(0)) revert EmptyValue();
        if (trustTierAt[commitment] != 0) revert AlreadyRecorded(commitment);
        trustTierAt[commitment] = uint64(block.number);
        trustTierOf[subject] = commitment;
        emit TrustTierAnchored(subject, commitment);
    }

    /// Registers a username to a subject. Registering a name the subject already owns does
    /// nothing; a name owned by someone else is refused.
    function registerUsername(bytes32 usernameHash, bytes32 subject) external onlySigner {
        if (usernameHash == bytes32(0) || subject == bytes32(0)) revert EmptyValue();
        bytes32 current = usernameOwner[usernameHash];
        if (current == subject) return;
        if (current != bytes32(0)) revert UsernameTaken(usernameHash);
        usernameOwner[usernameHash] = subject;
        emit UsernameRegistered(usernameHash, subject);
    }

    function issueCredential(bytes32 credentialHash, bytes32 subject, bytes32 issuer, bytes32 kind) external onlySigner {
        if (credentialHash == bytes32(0) || subject == bytes32(0) || issuer == bytes32(0) || kind == bytes32(0)) revert EmptyValue();
        if (credentials[credentialHash].blockNumber != 0) revert AlreadyRecorded(credentialHash);
        credentials[credentialHash] = Credential(subject, issuer, kind, uint64(block.number));
        emit CredentialIssued(credentialHash, subject, issuer, kind);
    }

    /// Publishes an issuer's revocation list. The sequence must be greater than the last one, so
    /// an older list can never replace a newer one.
    function publishStatusList(bytes32 issuer, uint64 sequence, bytes32 listHash) external onlySigner {
        if (issuer == bytes32(0) || listHash == bytes32(0) || sequence == 0) revert EmptyValue();
        StatusList storage current = statusLists[issuer];
        if (sequence <= current.sequence) revert StaleSequence(current.sequence, sequence);
        statusLists[issuer] = StatusList(sequence, listHash, uint64(block.number));
        emit StatusListPublished(issuer, sequence, listHash);
    }

    function attest(bytes32 kind, bytes32 digest) external onlySigner {
        if (kind == bytes32(0) || digest == bytes32(0)) revert EmptyValue();
        if (attestedAt[kind][digest] != 0) revert AlreadyRecorded(digest);
        attestedAt[kind][digest] = uint64(block.number);
        emit Attested(kind, digest);
    }
}
