# Phase 2: Onboarding, Identity & Credentials

**Total Work Orders:** 31  
**Status Summary:** 19 Completed, 10 Backlog, 2 Blocked  
**Last synced with Software Factory:** 2026-05-29

---

## Ready (3)

### WO-14: Build Streamlined User Onboarding Flow

**Status:** ✅ Completed (2026-05-26) — re-scoped to `FirstRunCoordinator`; see `docs/WO-14_ONBOARDING_RESCOPE.md`. OIDC4VC wallet path remains WO-100.

**Blueprint:** Decentralized Identity and Authentication

## Summary

Build the iOS onboarding UI flow for credential-based registration — guiding new users through username selection, DID generation (orchestrated via backend), Secure Enclave passkey setup, optional identity verification, and onboarding completion. The flow should complete in under 5 minutes for base access. Integrates with `AuthCoordinator` and calls existing service layers (DID, passkey, credentials).

## In Scope

- `OnboardingCoordinator` managing navigation flow: `UsernameView` → `DIDCreationProgressView` → `PasskeySetupView` → `OptionalVerificationView` → `OnboardingCompleteView`
- `UsernameView`: real-time availability check (`GET /v1/users/check-username`), E.164 or alphanumeric validation, debounced API calls
- `DIDCreationProgressView`: spinner with status messages ("Creating your identity on Cardano…"), 30-second timeout handling with error screen
- `PasskeySetupView`: `ASAuthorizationController` presentation, Secure Enclave key generation, success/failure states
- `OptionalVerificationView`: choose between Apple Digital ID (iOS 17+), document scan, or "Skip for now" → routes to `VerifyIdentityUseCase`
- `OnboardingCompleteView`: display DID, initial trust tier badge, optional "Continue to Full Identity Setup" CTA
- Onboarding state persistence: if interrupted mid-flow, resume from last completed step on re-launch
- Session: after DID + passkey creation, receive auth token from backend, store in Keychain

## Out of Scope

- DID anchoring on Cardano (WO-180, backend)
- Passkey signature validation (WO-1, WO-2)
- Identity verification processing (WO-17)
- Universal Onboarding phone-number flow (WO-203, WO-204)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Onboarding Flow:**
```
Start → UsernameView → DIDCreationProgressView → PasskeySetupView
                                                      ↓
                    OptionalVerificationView (Apple ID / Doc Scan / Skip)
                                                      ↓
                                           OnboardingCompleteView
```

**DID Creation Progress:**
```swift
struct DIDCreationProgressView: View {
    @ObservedObject var viewModel: OnboardingViewModel
    var body: some View {
        VStack {
            ProgressView()
            Text(viewModel.statusMessage)  // "Creating DID on Cardano...", "Linking passkey..."
        }
        .onAppear { Task { await viewModel.createDID() } }
        .onChange(of: viewModel.didCreationFailed) { failed in
            if failed { /* Show error with retry */ }
        }
    }
}
```

**Passkey Setup:**
```swift
struct PasskeySetupView: View {
    var body: some View {
        Button("Set Up Passkey") {
            Task { await viewModel.createPasskey() }
            // Calls: ASAuthorizationController → Secure Enclave → POST /v1/auth/passkey
        }
    }
}
```

## Blueprints

- Decentralized Identity and Authentication — Defines the onboarding flow including username selection, DID generation, passkey creation, optional verification, and 5-minute completion target

---

### WO-100: Implement OIDC4VC Protocol Support for Credential-Based Registration

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Build the OIDC4VC protocol client on iOS that initiates a credential-based registration flow when a user selects "Register with Verifiable Credential." This covers the wallet connection interface, credential selection from the user's existing digital wallet, and credential presentation to the backend OIDC4VC verifier endpoint. The backend's OIDC4VC issuer/verifier implementation is in WO-182.

## In Scope

- OIDC4VC authorization request construction (pre-authorized code flow and authorization code flow with PKCE)
- Wallet connection interface: show supported wallets, deep-link to wallet app, handle credential request callback
- Credential selection UI: list credentials available in connected wallet filtered by type acceptability
- Credential presentation: serialize and submit `VerifiablePresentation` to backend OIDC4VC verifier
- OIDC4VC discovery metadata fetch: `GET /.well-known/openid-credential-issuer`
- Protocol compliance validation: reject non-compliant requests with clear error messages
- Support for multiple wallet providers implementing OIDC4VC standard
- Success path: credential accepted → trigger profile creation and passkey setup continuation

## Out of Scope

- Backend credential verification logic (WO-182)
- Trust registry management (WO-118)
- Profile creation logic (WO-129)
- Passkey creation (WO-1)

## Requirements

Derived from the Streamlined Onboarding with Verifiable Credentials blueprint.

**OIDC4VC Flow (iOS):**
```swift
// Domain/UseCases/Auth/RegisterWithVerifiableCredentialUseCase.swift
struct RegisterWithVerifiableCredentialUseCase {
    func execute() async throws -> RegistrationResult {
        // 1. Fetch issuer metadata (discovery endpoint)
        let metadata = try await backendAPI.fetchOIDC4VCMetadata()

        // 2. Initiate credential request
        let credentialRequest = OIDC4VCCredentialRequest(
            credentialIssuer: metadata.credentialIssuer,
            credentialType: metadata.supportedCredentials.first!.type,
            format: "ldp_vc"
        )

        // 3. Deep-link to wallet: open wallet app with request
        let callback = try await walletConnector.requestCredential(credentialRequest)

        // 4. Submit presentation to backend verifier
        let presentation = VerifiablePresentation(credentials: callback.credentials)
        let result = try await backendAPI.submitPresentation(presentation)

        return .success(userDID: result.userDID, trustTier: result.trustTier)
    }
}
```

**Wallet Connection:**
```swift
// Presentation/Features/Onboarding/WalletConnectionView.swift
// Shows: supported wallets (any OIDC4VC-compatible digital wallet)
// Deep-links to wallet app via custom URI scheme
// Handles callback URL on return to ECHO app
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines OIDC4VC compliance, wallet connection, credential selection, and presentation flow

---

### WO-132: Build Verifiable Credential Issuance and Wallet Integration System

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the High-Assurance Verifiable Credential issuance system that fires after a successful IDV provider callback. This is the credential issuance path specifically for the In-App Identity Verification feature — issuing a `HighAssuranceCredential` with a 5-year expiry to the user's wallet via OIDC4VC, linked to their DID on Cardano.

## In Scope

- `HighAssuranceCredential` VC generation: W3C VC Data Model 1.0, `verificationLevel: "high_assurance"`, 5-year expiry
- Ed25519Signature2018 signing using ECHO's issuer DID private key
- OIDC4VC credential delivery to user's wallet (pre-authorized code flow)
- Cardano anchoring: store credential reference in user's DID document on Cardano
- Credential issuance within 60 seconds of IDV callback
- Renewal notification: 30 days before 5-year expiry (push notification + email)
- Credential revocation: flip bit in Cardano bit vector registry if verification is invalidated
- Local credential reference storage (no PII — only `{did, credentialType, issuedAt, expiresAt, credentialId}`)

## Out of Scope

- IDV provider integration (WO-120)
- Apple Digital ID (WO-126)
- Trust tier elevation (WO-37, WO-181)
- ECHO token reward distribution (WO-151)
- Streamlined Onboarding credential issuance (different flow, WO-182)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**High-Assurance VC:**
```json
{
  "@context": ["https://www.w3.org/2018/credentials/v1"],
  "type": ["VerifiableCredential", "HighAssuranceCredential"],
  "issuer": "did:prism:cardano:echo-identity-issuer",
  "issuanceDate": "2026-03-20T00:00:00Z",
  "expirationDate": "2031-03-20T00:00:00Z",
  "credentialSubject": {
    "id": "did:prism:cardano:user-did",
    "verificationLevel": "high_assurance",
    "humanityProof": true,
    "verificationMethod": "government_id"
  },
  "proof": {
    "type": "Ed25519Signature2018",
    "created": "2026-03-20T00:00:00Z",
    "verificationMethod": "did:prism:cardano:echo-identity-issuer#key-1",
    "signatureValue": "<ed25519-signature>"
  }
}
```

**Issuance Flow:**
```go
func (s *HighAssuranceVCService) IssueCredential(userDID string, documentType string) error {
    // 1. Build VC with 5-year expiry
    vc := buildHighAssuranceVC(userDID, documentType, time.Now().AddDate(5, 0, 0))
    // 2. Sign with ECHO's issuer DID key (Ed25519)
    signedVC := s.signer.Sign(vc, s.issuerPrivateKey)
    // 3. Deliver to user's wallet via OIDC4VC
    s.oidc4vc.IssueToWallet(userDID, signedVC)
    // 4. Anchor credential reference on Cardano
    s.cardano.AnchorCredentialInDIDDocument(userDID, signedVC.ID)
    // 5. Store local reference (no PII)
    s.db.StoreCredentialReference(userDID, "HighAssuranceCredential", vc.ExpirationDate)
    return nil
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines VC issuance upon successful verification, credential type, 5-year expiry, and wallet delivery
- Decentralized Identity and Authentication — Specifies credential issuance process, Cardano storage, and W3C VC format

---

## Blocked (2)

### WO-180: Implement Decentralized Identifier (DID) Management System

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## ⚠ Blocked — Pending Blueprint Update

**Reason:** The Data Layer and Decentralized Identity blueprints were updated on 2026-04-25 to eliminate Cardano from Phase 1–2. This WO uses Atala PRISM SDK, `did:prism:cardano:` DIDs, and Cardano anchoring — all obsolete.

**Replaced by:**
- **WO-273 — Implement did:key DID Management in Identity Service** — covers DID derivation, no-chain-transaction registration, multi-device key management
- **WO-274 — Implement W3C VC 2.0 Issuance and StatusList2021 Revocation on Constellation Identity Metagraph** — covers VC issuance to Constellation Identity Metagraph

**Do not implement until** this WO's description is rewritten to reflect `did:key` + Constellation Identity Metagraph.

---

## Summary

*(Original content preserved for reference)*

Build the DID management system (Identity Service, port 8001) using Atala PRISM/Veridian infrastructure — generating `did:prism:cardano:<id>` DIDs on user account creation, anchoring DID documents on Cardano, managing multi-device public key registration, and maintaining the backend DID-to-account mapping with Redis caching.

---

### WO-182: Implement Verifiable Credentials Management System

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## ⚠ Blocked — Pending Blueprint Update

**Reason:** The Data Layer and Decentralized Identity blueprints were updated on 2026-04-25 to eliminate Cardano from Phase 1–2. This WO issues VCs with `did:prism:cardano:echo-issuer`, stores credentials on Cardano Plutus UTXOs, and uses bit vector revocation on Cardano — all obsolete.

**Replaced by:**
- **WO-274 — Implement W3C VC 2.0 Issuance and StatusList2021 Revocation on Constellation Identity Metagraph** — covers all VC lifecycle management using Constellation Identity Metagraph and StatusList2021

**Do not implement until** this WO's description is rewritten to reflect Constellation Identity Metagraph VC management.

---

## Summary

*(Original content preserved for reference)*

Build the W3C Verifiable Credentials management system with full OIDC4VC protocol compliance — credential issuance coordinating with IDV providers, storage on Cardano blockchain, credential verification with revocation checking, and multi-format support (JSON-LD, JWT, SD-JWT).

---

## Backlog (21)

### WO-39: Build User Management and Contact System with Privacy Controls

**Blueprint:** Frontend

## Summary

Build the user profile and contact management screens with privacy controls. Contact discovery (PSI, QR code, invite links, username search) is handled by WO-220–222; this work order covers managing contacts once discovered. The combined Frontend blueprint adds four explicit contact use cases with Argon2id phone hashing, a `Contacts/` feature folder, and a `QRContactExchangeUseCase`.

## In Scope

- `User` domain model with avatar, display name, `@username`, bio, trust tier, and verification credentials
- `UserRepository` backed by local SwiftData + backend sync
- Profile view displaying avatar, display name, trust tier badge, verification status, and bio
- Contact list with favorites, search, and filtering
- Contact blocking — blocked users cannot message or view the user's online status; blocking synced to backend (Contacts Service, port 8005)
- Privacy settings view: per-setting controls for `lastSeen`, `onlineStatus`, `profilePicture`, `statusMessage`, `groupInvites`, `callPermissions`
- Per-contact privacy overrides
- Online/offline status indicators in conversation list and profile views
- `TrustBadge` and `VerificationBadge` components
- Contact search with DID-based lookup via backend Contacts Service
- **4 contact use cases in `Domain/UseCases/Contacts/`:**
  - `ContactDiscoveryUseCase` — hashes contacts' phone numbers on-device with Argon2id + per-user salt; sends hashed set to Contacts Service; receives matching DIDs; renders "Contacts on ECHO" list
  - `QRContactExchangeUseCase` — generate QR code containing user DID + public key; scan QR to add contact (zero server involvement for QR generation)
  - `InviteLinkUseCase` — generate referral invite link via `POST /contacts/invite`; track referral chain (max 3 tiers) for 50 ECHO reward
  - `UsernameSearchUseCase` — search public handles via `GET /contacts/search?handle={username}`; returns DID, display name, trust tier, verification status
- **`Contacts/` feature folder** in `Presentation/Features/Contacts/`: contact discovery list, QR scanner, invite link share sheet, username search

## Out of Scope

- PSI OPRF-based backend contact discovery service (WO-220)
- iOS PSI client library integration (WO-221)
- Username index on metagraph, QR infrastructure (WO-222)
- Messaging functionality (WO-28)
- Identity verification flows (WO-17)
- Push notifications (WO-57)
- Persona-specific contact management (Multiple Personas work orders)

## Requirements

From the Frontend blueprint:

**4 Contact Use Cases (canonical, from Frontend):**
```swift
// Domain/UseCases/Contacts/
actor ContactDiscoveryUseCase {
    // Hashes contacts' phone numbers on-device using Argon2id + per-user salt
    // Sends hashed set to /contacts/discover — server never sees raw numbers
    // Returns matching DIDs → "Contacts on ECHO" list
    func discoverContacts() async throws -> [DiscoveredContact]
}

actor QRContactExchangeUseCase {
    func generateQRCode() -> QRCode  // Contains DID + public key; zero server involvement
    func scanQRCode(_ code: QRCode) async throws -> Contact  // Add contact from scan
}

actor InviteLinkUseCase {
    func generateInviteLink() async throws -> URL  // POST /contacts/invite
    // Backend tracks referral chain for 50 ECHO reward (max 3 tiers)
}

actor UsernameSearchUseCase {
    func search(handle: String) async throws -> [ContactSearchResult]
    // GET /contacts/search?handle={username}
    // Returns: DID, displayName, trustTier, verificationStatus
}
```

**User Model:**
```swift
struct User: Identifiable, Codable {
    let id: String          // DID: "did:prism:cardano:..."
    var displayName: String
    var username: String    // "@alice" — public, on metagraph Data L1
    var avatarURL: URL?
    var bio: String?
    var trustTier: Int      // 1–5
    var isVerified: Bool    // Tier 4+
    var isOnline: Bool
    var lastSeen: Date?
    var privacySettings: PrivacySettings
    var credentials: [Credential]
}
```

**Project structure addition:**
```
Presentation/Features/
├── Contacts/               # NEW folder
│   ├── ContactDiscoveryView.swift    # "Contacts on ECHO" list from Argon2id match
│   ├── QRScannerView.swift           # Scan QR to add contact
│   ├── InviteLinkView.swift          # Share invite link
│   └── UsernameSearchView.swift      # Search public handles
```

## Blueprints

- Frontend — Defines 4 contact use cases (`ContactDiscoveryUseCase` with Argon2id, `QRContactExchangeUseCase`, `InviteLinkUseCase`, `UsernameSearchUseCase`), `Contacts/` feature folder, and Argon2id contact discovery hashing
- Dynamic Trust Network and Social Verification — Specifies user profiles, contact blocking, and privacy settings
- Privacy-Preserving Contact Discovery — Defines `@username` as a public Data L1 identifier and alternative contact add methods

---

### WO-109: Build Verifiable Credential Verification Engine with Trust Registry Integration

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Build the backend verifiable credential verification engine for the Streamlined Onboarding flow — verifying cryptographic signatures against the issuer DID public key, checking issuer status in the trust registry, and confirming credential non-revocation via the Cardano bit vector. Used by the OIDC4VC verifier endpoint (WO-182) during credential-based registration.

## In Scope

- Cryptographic signature verification for `Ed25519Signature2018` and `JsonWebSignature2020` proof types
- Issuer DID public key resolution via Cardano (same DID resolution path as user DIDs)
- Trust registry check: verify issuer DID exists in trusted issuer registry and is in `active` status
- Revocation check: query Cardano UTXO bit vector for credential's index; reject if revoked
- Credential expiration check: compare `expirationDate` against current time
- Support for credential formats: JSON-LD, JWT (JSON Web Token), SD-JWT (Selective Disclosure JWT)
- Complete verification within 10 seconds; return structured result: `{valid, reason, issuerDID, credentialType}`
- Verification attempt audit log (timestamp, issuerDID, outcome — no credential content)

## Out of Scope

- Trust registry maintenance (WO-118)
- OIDC4VC presentation protocol (WO-100, WO-182)
- Profile creation after successful verification (WO-129)
- Passkey creation (WO-136)

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Verification Pipeline:**
```go
// pkg/credentials/verifier.go
type CredentialVerifier struct {
    trustRegistry *TrustRegistry
    cardano       *CardanoClient
    redis         *redis.Client
}

func (v *CredentialVerifier) Verify(vc VerifiableCredential) (*VerificationResult, error) {
    // 1. Check expiration
    if vc.ExpirationDate.Before(time.Now()) {
        return &VerificationResult{Valid: false, Reason: "credential_expired"}, nil
    }

    // 2. Verify issuer is in trusted registry
    issuerStatus, err := v.trustRegistry.GetIssuerStatus(vc.Issuer)
    if err != nil || issuerStatus != "active" {
        return &VerificationResult{Valid: false, Reason: "untrusted_issuer"}, nil
    }

    // 3. Check revocation (Cardano UTXO bit vector)
    revoked, err := v.cardano.IsCredentialRevoked(vc.Issuer, vc.CredentialIndex)
    if revoked { return &VerificationResult{Valid: false, Reason: "credential_revoked"}, nil }

    // 4. Verify cryptographic signature
    issuerPublicKey, _ := v.cardano.ResolveDIDPublicKey(vc.Issuer)
    if !verifySignature(vc, issuerPublicKey) {
        return &VerificationResult{Valid: false, Reason: "invalid_signature"}, nil
    }

    return &VerificationResult{Valid: true, CredentialType: vc.Type, IssuerDID: vc.Issuer}, nil
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines credential verification process: signature check, issuer status, revocation check
- Decentralized Identity and Authentication — Specifies W3C VC formats, signature types, and Cardano revocation mechanism

---

### WO-118: Implement Trust Registry Management System for Issuer Verification

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Build the trust registry service that maintains a list of approved verifiable credential issuers — their DID, trust level, credential types they issue, and operational status. The registry is queried by the credential verification engine during OIDC4VC-based onboarding. Registry data is refreshed from upstream trusted sources every 24 hours with admin override capability.

## In Scope

- Trust registry data model: `{issuerDID, name, trustedCredentialTypes[], status, trustLevel, lastVerified, endpoints}`
- Issuer status states: `active`, `suspended`, `revoked`, `pending`
- `GET /v1/trust-registry/issuers/{did}` — real-time issuer status query, response < 2 seconds (Redis cache)
- Automatic 24-hour refresh from upstream trusted registry sources
- Admin API: add/remove issuers with approval workflow, status change with audit trail
- Read-only public access to issuer status for credential verification use
- Audit log of all registry changes, status updates, and access patterns
- Initial registry seed: known government ID issuers, Apple Digital ID, major IDV providers (Prove, Daon, 1Kosmos)

## Out of Scope

- Credential verification logic (WO-109)
- OIDC4VC protocol (WO-100, WO-182)
- User-facing registry browsing UI

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Trust Registry Data Model:**
```go
type TrustedIssuer struct {
    IssuerDID           string            `json:"issuer_did"`
    Name                string            `json:"name"`
    TrustedCredentialTypes []string       `json:"trusted_credential_types"` // ["HighAssuranceCredential", "KYCLiteCredential"]
    Status              IssuerStatus      `json:"status"`    // active, suspended, revoked, pending
    TrustLevel          string            `json:"trust_level"` // "government", "financial_institution", "idv_provider"
    LastVerified        time.Time         `json:"last_verified"`
    StatusHistory       []StatusChange    `json:"status_history,omitempty"`
}

type IssuerStatus string
const (
    IssuerActive    IssuerStatus = "active"
    IssuerSuspended IssuerStatus = "suspended"
    IssuerRevoked   IssuerStatus = "revoked"
    IssuerPending   IssuerStatus = "pending"
)
```

**Registry Query (Cached):**
```go
func (r *TrustRegistryService) GetIssuerStatus(issuerDID string) (IssuerStatus, error) {
    // 1. Redis cache (5-minute TTL for status)
    if cached := r.redis.Get("issuer:" + issuerDID); cached != nil {
        return IssuerStatus(cached), nil
    }
    // 2. Database lookup
    issuer, err := r.db.FindIssuer(issuerDID)
    if err != nil { return "", ErrUnknownIssuer }
    r.redis.Set("issuer:"+issuerDID, string(issuer.Status), 5*time.Minute)
    return issuer.Status, nil
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines trust registry requirements: issuer registry maintenance, status tracking, decentralized architecture, update frequency, and transparency

---

### WO-129: Create Automatic Profile Generation with Verified Data Population

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

After a successful OIDC4VC credential verification, automatically create the user's profile and populate it with verified data extracted from the credential. Assign the appropriate trust tier based on the credential type. This triggers the backend onboarding completion sequence: create account, issue trust tier datum on Cardano, assign trust score, and prompt passkey creation.

## In Scope

- Backend: extract verified attributes from VC `credentialSubject` (no PII stored server-side; reference UUID only)
- Create user account record: DID, username (from credential if available, or generated), trust tier assignment
- Issue trust tier datum on Cardano reflecting the verified credential type
- Trigger trust score initialization via Trust Service (WO-49/WO-181)
- Return initial trust tier and verification badge info to iOS for display
- Profile creation completion within 5 seconds of successful credential verification
- Emit `onboarding_complete` event to Rewards Service for ECHO token reward if applicable
- iOS: `OnboardingCompleteView` displaying DID, trust tier badge, and "Continue to Full Identity Setup" CTA

## Out of Scope

- Profile editing (user-initiated profile management work orders)
- Photo/avatar upload
- Credential verification (WO-109)
- ECHO token reward distribution (WO-151/WO-167)

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Profile Auto-Population:**
```go
// identity/profile_service.go
func (s *ProfileService) CreateFromCredential(verification VerificationResult) (*UserProfile, error) {
    // 1. Create user account (DID from Atala PRISM, username auto-generated or from credential)
    did, err := s.didService.CreateDID(verification.PublicKey)
    
    // 2. Map credential type to trust tier
    tier := tierFromCredentialType(verification.CredentialType)
    // "HighAssuranceCredential" → Tier 4, "KYCLiteCredential" → Tier 3
    
    // 3. Issue trust tier datum on Cardano (~0.3-0.5 ADA from treasury)
    s.cardano.IssueTrustTierDatum(did, tier, credentialExpiry)
    
    // 4. Create minimal profile (no PII stored)
    profile := UserProfile{
        DID:       did,
        TrustTier: tier,
        IsVerified: tier >= 3,
    }
    s.db.CreateProfile(profile)
    
    // 5. Signal for ECHO reward (if High-Assurance or KYC-Lite)
    if tier >= 3 { s.rewardsService.QueueVerificationReward(did, verification.CredentialType) }
    
    return &profile, nil
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines automatic profile creation, verified data population, high trust score assignment, and badge issuance

---

### WO-144: Build Onboarding Analytics and Support System

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Build the onboarding analytics and support backend — tracking completion rates, credential type breakdowns, stage timing, and dropout patterns to optimize the onboarding experience. All analytics are anonymized (no PII, no credential content). Support documentation and FAQ resources are surfaced to users encountering errors during the onboarding flow.

## In Scope

- Onboarding event tracking: `onboarding_started`, `stage_completed`, `stage_failed`, `onboarding_completed`, `credential_type_used`, `dropout_at_stage`
- Completion rate metrics by: credential type, stage, trust tier assigned
- Average time per stage and overall completion time
- Dropout analysis: identify common failure points with error code frequency
- Trust score distribution reporting for newly onboarded users
- Analytics API for internal dashboard (not user-facing)
- Support ticket creation endpoint: `POST /v1/support/tickets` captures error code, stage, device info (no PII)
- In-app support access: FAQ links and documentation surfaced contextually based on current onboarding stage
- Analytics data export in CSV/JSON

## Out of Scope

- Support agent interface (customer support tooling)
- Advanced ML analytics
- External analytics platform integration (Mixpanel, Amplitude, etc.)
- Live chat support

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Analytics Events (Privacy-Safe):**
```go
type OnboardingAnalyticsEvent struct {
    EventType       string    `json:"event_type"`   // No DID/PII
    Stage           string    `json:"stage"`         // "oidc4vc_initiation", "credential_selection", "passkey_setup"
    CredentialType  string    `json:"credential_type,omitempty"` // "HighAssuranceCredential"
    ErrorCode       string    `json:"error_code,omitempty"`      // "untrusted_issuer", "credential_revoked"
    DurationMs      int       `json:"duration_ms,omitempty"`
    Timestamp       time.Time `json:"timestamp"`
    // NEVER: DID, credential content, personal data
}
```

**Onboarding Funnel Metrics:**
```go
type OnboardingFunnelReport struct {
    TotalStarted       int     // Total onboardings initiated
    CompletedBase      int     // Completed with no verification
    CompletedVerified  int     // Completed with VC verification
    DropoutRates       map[string]float64  // Stage → dropout %
    AvgCompletionSecs  float64
    CredentialTypeBreakdown map[string]int  // Type → count
    TrustTierDistribution   map[int]int     // Tier → count
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines analytics requirements including completion rate tracking, credential type analytics, trust score distribution, and support integration

---

### WO-159: Implement Verification Retry System with Error Handling and User Guidance

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the verification retry system — tracking failed verification attempts, enforcing rate limits, providing actionable error guidance, and offering a seamless retry experience. When IDV providers return failures (poor image quality, document mismatch, liveness check failed), users get specific guidance to improve their next attempt.

## In Scope

- Attempt tracking: PostgreSQL record of `{userDID, attemptTime, failureReason, provider, cooldownUntil}`
- Rate limiting: maximum 5 attempts per 24-hour rolling window per DID
- Cooldown enforcement: 1-hour minimum wait between failed attempts
- Error code to user guidance mapping: `image_quality_low` → "Ensure good lighting and the full document is visible"; `liveness_failed` → "Look directly at the camera and remove sunglasses"; `document_mismatch` → "Use the document that matches your selfie"
- Retry UI: `VerificationRetryView` with failure explanation, improvement tips, and "Try Again" CTA
- Attempt exhaustion: after 5 failures in 24 hours, show support escalation path with deep link to support ticket creation
- Cooldown timer display: show "You can retry in Xh Xm" when within cooldown window

## Out of Scope

- Verification processing (WO-104, WO-113, WO-120)
- Manual verification review process
- Account suspension

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Attempt Tracking (Backend):**
```go
type VerificationAttempt struct {
    UserDID        string    `db:"user_did"`
    AttemptTime    time.Time `db:"attempt_time"`
    FailureReason  string    `db:"failure_reason"`  // "image_quality_low", "liveness_failed"
    Provider       string    `db:"provider"`         // "stripe_identity", "sumsub"
    CooldownUntil  time.Time `db:"cooldown_until"`
}

func (s *VerificationService) CanAttempt(userDID string) (bool, time.Duration) {
    attempts := s.db.GetAttempts(userDID, 24*time.Hour)
    if len(attempts) >= 5 { return false, 0 }
    if len(attempts) > 0 && time.Now().Before(attempts[0].CooldownUntil) {
        remaining := time.Until(attempts[0].CooldownUntil)
        return false, remaining
    }
    return true, 0
}
```

**iOS Retry View:**
```swift
struct VerificationRetryView: View {
    let failureReason: String
    let attemptsRemaining: Int
    let cooldownRemaining: TimeInterval?
    var body: some View {
        VStack {
            Text(failureTitle(for: failureReason))
            Text(failureGuidance(for: failureReason))  // Specific improvement tip
            if let cooldown = cooldownRemaining {
                Text("Try again in \(formatDuration(cooldown))")
            } else if attemptsRemaining > 0 {
                Button("Try Again") { /* restart verification */ }
                Text("\(attemptsRemaining) attempt(s) remaining today")
            } else {
                Button("Contact Support") { /* support ticket */ }
            }
        }
    }
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines retry functionality, error message clarity, guidance provision, attempt limits (5/24h), cooldown periods (1h), and support access

---

### WO-169: Implement Onboarding Support System with Documentation and Help Resources

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Build the user-facing in-app support and documentation system for the onboarding flow — contextual error guidance, FAQ resources, and support ticket creation. This is the iOS-side support layer; when users encounter errors during OIDC4VC credential-based onboarding, they get specific guidance with links to relevant documentation and the ability to submit a support ticket with automatic technical context capture.

## In Scope

- `OnboardingHelpView`: accessible from a help icon at any onboarding step (within 2 taps)
- FAQ content organized by topic: credential wallet setup, OIDC4VC process, passkey creation, common error codes
- Contextual error guidance: map error codes (e.g., `untrusted_issuer`, `credential_revoked`, `wallet_connection_failed`) to user-friendly explanations + resolution steps
- Support ticket creation: capture current stage, error code, iOS version, ECHO app version (no PII), submit to backend `POST /v1/support/tickets`
- `OnboardingHelpOverlay`: bottom sheet presentation triggered by error state showing explanation + primary action (retry, get help, skip)
- Troubleshooting guides for: wallet not connecting, credential not accepted, passkey creation failing, biometric not available
- Help resources load within 3 seconds (cached locally, refresh on app launch)

## Out of Scope

- Live chat or real-time support
- Support agent backend tooling
- Multi-language support (English only initially)
- Analytics event tracking (WO-144)

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Error Code to User Guidance Mapping:**
```swift
// Core/Support/OnboardingErrorGuide.swift
struct OnboardingErrorGuide {
    static func guidance(for errorCode: String) -> ErrorGuidance {
        switch errorCode {
        case "untrusted_issuer":
            return ErrorGuidance(
                title: "Credential Not Accepted",
                explanation: "The issuer of your credential is not yet in our trusted registry.",
                actions: [.tryDifferentCredential, .contactSupport]
            )
        case "credential_revoked":
            return ErrorGuidance(
                title: "Credential Has Been Revoked",
                explanation: "Your credential has been revoked by the issuer. Contact your credential issuer.",
                actions: [.contactIssuer, .tryDifferentCredential]
            )
        case "wallet_connection_failed":
            return ErrorGuidance(
                title: "Wallet Connection Failed",
                explanation: "We couldn't connect to your wallet. Make sure your wallet app is installed.",
                actions: [.retryConnection, .viewWalletSetupGuide]
            )
        default:
            return ErrorGuidance(title: "Something Went Wrong", actions: [.retry, .contactSupport])
        }
    }
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines support documentation, FAQ resources, customer support access, troubleshooting guides, and contextual error messaging requirements

---

### WO-187: Implement User Profiles and Contact Management System

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the user profile management system — avatar, display name, username, bio, and status — with full integration with the trust network. Profiles display verification badges based on VC status in the Constellation Identity Metagraph and cached trust scores. Privacy settings control visibility of each profile field. This iOS work order covers the profile editing UI and the backend Contacts Service (port 8005) profile API.

## In Scope

- Profile editing UI: avatar upload/crop, display name, username, bio, status message
- `ProfileViewModel` with `UserRepository` backing (local SwiftData + backend sync)
- Verification badge display: `VerificationBadge` component showing tier-based checkmark or shield
- Trust score visibility: user controls who sees their trust score (`everyone`, `contacts`, `nobody`)
- Last seen and online/offline status display with privacy-controlled visibility
- Privacy settings UI: per-field visibility controls using `PrivacySettings` struct
- Backend Contacts Service (port 8005): `GET /v1/profile/{did}`, `PATCH /v1/profile` endpoints
- Profile data encrypted at rest in SwiftData and on backend PostgreSQL
- Trust tier badge refresh: listen for `trust_tier_changed` WebSocket events, refresh badges

## Out of Scope

- Trust score computation (WO-181)
- Trust tier management (WO-49)
- Contact blocking (WO-190)
- Messaging (WO-28)

## Requirements

Derived from the Dynamic Trust Network and Social Verification blueprint.

**Profile Data Model:**
```swift
struct UserProfile: Codable {
    let did: String               // "did:key:..."
    var displayName: String
    var username: String          // Unique, alphanumeric
    var avatarURL: URL?
    var bio: String?
    var statusMessage: String?
    var trustTier: Int            // 1–5 (from Trust Service)
    var isVerified: Bool          // Tier 4+
    var privacySettings: PrivacySettings
    var lastSeen: Date?           // Only shown per privacy settings
    var isOnline: Bool            // Only shown per privacy settings
}
```

**Backend Profile API (Contacts Service, port 8005):**
```go
// GET /v1/profile/:did  → returns profile with privacy-filtered fields
// PATCH /v1/profile     → update own profile (requires auth)
// Fields returned depend on caller's relationship to target user
// Privacy settings enforced on backend, not just client
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines user profile features, verification badge display, privacy settings, trust score visibility, and contact management
- Frontend — Defines `User` model, `UserRepository`, `TrustBadge`, `VerificationBadge` components

---

### WO-190: Build Contact Blocking and Privacy Control System

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the contact blocking system and granular privacy controls. Blocked users cannot message, call, or see the blocking user's status. Privacy settings control visibility per field (last seen, online status, profile picture, status, group invites, calls). Settings are stored locally and synced to the Contacts Service (port 8005) for backend enforcement.

## In Scope

- Block/unblock API: `POST /v1/contacts/block` and `DELETE /v1/contacts/block/{did}` on Contacts Service (port 8005)
- Blocked user list management UI with search
- Block enforcement: Message Relay (port 8002) rejects messages from blocked DIDs; Trust Service excludes blocked users from contact search
- Silent blocking: blocked users receive no notification
- Privacy settings API: `PATCH /v1/profile/privacy` — persist settings in PostgreSQL
- `PrivacySettings` stored locally in SwiftData + synced to backend
- Per-contact privacy overrides: allow specific contacts to see data despite global setting
- Privacy setting enforcement: backend returns 403 with field omitted when caller is restricted

## Out of Scope

- Trust score privacy controls (handled by trust tier)
- Group management beyond invite permissions
- Disappearing messages or encryption specifics

## Requirements

Derived from the Dynamic Trust Network and Social Verification blueprint.

**Block Enforcement (Backend):**
```go
// When Message Relay receives message from senderDID to recipientDID:
func (s *RelayService) IsBlocked(senderDID, recipientDID string) bool {
    // Check Contacts Service block list (Redis cache, 60s TTL)
    return s.contactsService.IsBlocked(recipientDID, senderDID)
}
// If blocked: return 403, don't relay message, don't notify sender of block
```

**Privacy Settings Model:**
```swift
struct PrivacySettings: Codable {
    var showLastSeen: VisibilityLevel      // .everyone, .contacts, .nobody
    var showOnlineStatus: VisibilityLevel
    var showProfilePicture: VisibilityLevel
    var showStatusMessage: VisibilityLevel
    var allowGroupInvites: VisibilityLevel
    var allowCalls: VisibilityLevel
    var contactOverrides: [String: VisibilityOverride]  // keyed by DID

    enum VisibilityLevel: String, Codable { case everyone, contacts, nobody }
}
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines blocking functionality, blocked user list, privacy settings per field, and per-contact privacy overrides
- Backend — Defines Contacts Service (port 8005) for block list and contact management

---

### WO-199: Implement ISO/IEC 18013-5 mDL Verification for Identity Registration

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication, In-App High-Assurance Identity Verification and Reward

## Summary

Implement ECHO's role as an ISO/IEC 18013-5 mDL verifier — users present their government-issued mobile Driver's License from their device wallet, the backend cryptographically validates the credential's authenticity against the issuing authority's certificate chain, and the result triggers trust tier elevation on Cardano. ECHO is a *verifier* only; mDL credentials are issued by government DMVs, not by ECHO.

## In Scope

- **Device engagement:** support QR code-based device engagement flow (online presentation mode) so the user's mDL wallet app presents the license to ECHO's backend verifier
- **mDL data request:** backend generates a `DeviceRequest` specifying required data elements (data minimization — request only: `family_name`, `given_name`, `birth_date`, `age_over_18`, `issuing_country`, `document_number`); no SSN, no full address
- **MSO signature validation:** parse the user's `DeviceResponse`, extract the `MobileSecurityObject` (MSO), cryptographically verify its signature against the issuing authority's certificate
- **IACA trust store:** maintain a trust store of IACA (Issuing Authority Certificate Authority) root certificates for US states (initial scope); validate certificate chain from IACA root → document signer → MSO
- **Verification result mapping:** successful mDL verification → map to Trust Tier 4 (Verified); submit trust tier datum to Cardano via WO-37; no raw PII stored (only reference UUID + pass/fail + age_over_18 flag per REQ-PRIV-006)
- **Error handling:** return clear user guidance for invalid signatures, expired credentials, revoked certificates, and unsupported issuing authorities
- **Interoperability:** compatible with ISO/IEC 18013-5 compliant holder apps (iOS Wallet, Google Wallet)

## Out of Scope

- mDL issuance — ECHO does not issue government credentials; WO-200 and WO-201 (mDL issuer) have been removed as out of scope for ECHO
- CRL/OCSP management — certificate revocation checking uses standard IACA trust store; full CRL infrastructure is not implemented
- In-person NFC proximity flow (online QR presentation only for Phase 1)
- Android holder apps (iOS-only in Phase 1)

## Requirements

Derived from the In-App High-Assurance Identity Verification and Reward and Decentralized Identity and Authentication blueprints.

**mDL Verification Flow:**
```
1. User initiates mDL verification from profile
2. Backend generates DeviceRequest with required data elements
3. QR code displayed to user; user scans with their mDL holder app
4. mDL holder app presents DeviceResponse (MSO + signed attributes)
5. Backend validates:
   a. IACA certificate chain (root → document signer)
   b. MSO signature using document signer public key
   c. Data elements match request (no extra data exposed)
   d. Credential not expired
6. On success: pass/fail + age_over_18 + issuing_country forwarded to Identity Service
7. Identity Service issues Trust Tier 4 credential on Cardano (WO-37)
8. No PII stored — only reference UUID, pass/fail, age_over_18, document_type
```

**Privacy constraint (from Privacy Architecture blueprint):**
- Backend receives only: pass/fail, confidence, document type, issuing country, age_over_18 flag — no names, DOB, document number, or biometrics (AC-PRIV-006.2)
- IDV result stored on Cardano as opaque reference UUID with credential type and assurance level only (AC-PRIV-006.3)

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines government ID verification flow, data minimization requirements, third-party verifier integration pattern, and trust tier elevation
- Decentralized Identity and Authentication — Defines Trust Tier 4 assignment, credential issuance to Cardano, and the DID-based identity model

---

### WO-202: Build SMS Verification Service with Twilio and Prove Integration

**Type:** Build

**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Build a backend SMS verification service that provides reliable one-time code delivery using Twilio as the primary provider and Prove as a fallback. This service enables phone number verification during the universal onboarding flow.

## In Scope

- Twilio API integration for SMS code generation and delivery
- Prove API integration as fallback provider
- Automatic failover logic when Twilio is unavailable or fails
- Verification code generation (6-digit codes)
- Code expiration management (configurable TTL)
- Rate limiting per phone number to prevent abuse
- Regional provider selection logic
- Retry logic with exponential backoff
- Provider health monitoring and alerting

## Out of Scope

- Phone number storage or persistence (handled by onboarding backend)
- User session management (handled by onboarding backend)
- DID generation or blockchain integration
- Frontend UI components
- Other SMS providers beyond Twilio and Prove

## Requirements

The requirements document for Universal Onboarding and Identity Creation provides the following functional requirements that apply to this work order:

### FR2: SMS Verification

The system sends a one-time verification code via a third-party SMS gateway (Twilio primary, Prove fallback). The user must enter the code to prove ownership of the number.

**Acceptance Criteria:**
- SMS codes must be 6 digits
- Codes must expire after a configurable time period
- System must automatically failover to Prove if Twilio fails
- Verification attempts must be rate-limited per phone number
- All communication must use TLS 1.3

### NFR1: Performance

SMS verification code delivery and verification must complete within 2 seconds on average under normal network conditions.

### NFR4: Reliability

SMS gateway failures trigger an automatic retry with the fallback provider (Prove). The system must guarantee 99.5% successful verification attempts.

### NFR3: Security

All communication uses TLS 1.3 and Kinnami end-to-end encryption.

## Blueprints

- Universal Onboarding and Identity Creation — Defines the SMS verification service architecture, API contracts, and provider failover logic
- Backend — Establishes Go service patterns, API design, error handling, and infrastructure standards

---

### WO-203: Implement Universal Onboarding Backend Orchestration

**Type:** Build

**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Build the backend orchestration service that coordinates the complete universal onboarding flow, including session management, SMS verification, DID creation, passkey linking, and initial trust score assignment. This service acts as the central coordinator for all onboarding operations.

## In Scope

- Onboarding session creation and lifecycle management
- Integration with SMS verification service
- Integration with DID generation service (Cardano/Atala PRISM)
- Passkey public key registration and storage
- Initial trust score assignment (5 points for device-verified users)
- Encrypted phone number storage with optional deletion
- REST API endpoints for the onboarding flow
- Transaction coordination across services
- Error handling and rollback logic
- Audit logging for compliance

## Out of Scope

- SMS provider integration (handled by SMS verification service)
- DID minting on Cardano blockchain (handled by DID service)
- Trust score calculation algorithms (handled by trust score service)
- Frontend UI implementation
- Advanced identity verification flows (separate features)

## Requirements

The requirements document for Universal Onboarding and Identity Creation provides the following functional requirements:

### FR1: Mobile Number Entry

The iOS app must allow a new user to enter a mobile phone number to initiate onboarding.

**Implementation Note:** This work order handles the backend endpoint that receives the phone number.

### FR2: SMS Verification

The system sends a one-time verification code via a third-party SMS gateway (Twilio primary, Prove fallback). The user must enter the code to prove ownership of the number.

**Implementation Note:** This work order orchestrates the SMS verification process.

### FR3: DID Generation

Immediately after successful SMS verification, the backend generates a Decentralized Identifier (DID) on the Cardano blockchain and anchors it to the user's profile.

### FR4: Passkey Creation

The iOS app generates a fresh passkey using the device's Secure Enclave. The public key is sent to the backend and linked to the newly created DID document.

**Implementation Note:** This work order handles receiving and storing the public key.

### FR5: Initial Trust Score Assignment

Upon completion of steps 2-4, the user receives an initial trust score of 5 points (device-verified tier) to unlock basic platform features.

### FR6: Optional Phone Number Decoupling

After DID creation, the user may choose to delete the stored phone number. If deleted, the number is removed from all persistent storage.

### FR7: Transition to Progressive Identity

The onboarding flow must surface a prompt encouraging the user to continue to the "Streamlined Onboarding with Verifiable Credentials" or "In-App High-Assurance Identity Verification" flows for higher trust levels.

**Implementation Note:** Backend returns appropriate flags/data to enable frontend prompts.

### NFR1: Performance

SMS verification code delivery and verification must complete within 2 seconds on average under normal network conditions.

### NFR2: Scalability

The onboarding service must handle 10,000 concurrent sessions without degradation, supporting auto-scaling of backend containers.

### NFR3: Security

All communication uses TLS 1.3 and Kinnami end-to-end encryption. Passkeys never leave the Secure Enclave; only the public key is transmitted.

### NFR5: Privacy

Phone numbers are stored encrypted at rest and are deleted immediately after the user opts to decouple them. No logs retain the raw number.

### NFR6: Auditable Traceability

DID creation and passkey linking events are recorded on the Cardano metagraph with immutable transaction hashes for compliance.

## Blueprints

- Universal Onboarding and Identity Creation — Defines the complete onboarding orchestration flow, API contracts, data models, and service interactions
- Backend — Establishes Go service patterns, REST API design, middleware, encryption, and error handling standards

---

### WO-204: Build Universal Onboarding UI Flow

**Type:** Build

**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Build the complete iOS user interface for the universal onboarding flow, providing a seamless experience from phone number entry through DID creation and passkey setup. This includes all SwiftUI views, view models, and client-side orchestration logic.

## In Scope

- PhoneNumberEntryView with E.164 format validation
- VerificationCodeView with 6-digit input and countdown timer
- DIDCreationProgressView with status indicators and error handling
- PasskeySetupView with Secure Enclave integration
- OnboardingCompleteView showing DID, trust score, and next steps
- PhoneNumberDeletionToggle for optional privacy enhancement
- View models for each screen following MVVM pattern
- API integration with onboarding backend endpoints
- Local passkey generation using Secure Enclave
- State management and navigation flow
- Error handling and user feedback
- Loading states and progress indicators
- Accessibility support (VoiceOver, Dynamic Type)

## Out of Scope

- Backend API implementation
- SMS sending infrastructure
- DID generation on blockchain
- Trust score calculation
- Advanced identity verification flows (separate features)
- Android implementation

## Requirements

The requirements document for Universal Onboarding and Identity Creation provides the following functional requirements:

### FR1: Mobile Number Entry

The iOS app must allow a new user to enter a mobile phone number to initiate onboarding.

### FR2: SMS Verification

The system sends a one-time verification code via a third-party SMS gateway (Twilio primary, Prove fallback). The user must enter the code to prove ownership of the number.

**Implementation Note:** This work order builds the UI for code entry and submission.

### FR3: DID Generation

Immediately after successful SMS verification, the backend generates a Decentralized Identifier (DID) on the Cardano blockchain and anchors it to the user's profile.

**Implementation Note:** This work order displays progress and status to the user.

### FR4: Passkey Creation

The iOS app generates a fresh passkey using the device's Secure Enclave. The public key is sent to the backend and linked to the newly created DID document.

### FR5: Initial Trust Score Assignment

Upon completion of steps 2-4, the user receives an initial trust score of 5 points (device-verified tier) to unlock basic platform features.

**Implementation Note:** This work order displays the trust score to the user.

### FR6: Optional Phone Number Decoupling

After DID creation, the user may choose to delete the stored phone number. If deleted, the number is removed from all persistent storage.

**Implementation Note:** This work order provides the UI toggle for deletion.

### FR7: Transition to Progressive Identity

The onboarding flow must surface a prompt encouraging the user to continue to the "Streamlined Onboarding with Verifiable Credentials" or "In-App High-Assurance Identity Verification" flows for higher trust levels.

### NFR3: Security

All communication uses TLS 1.3 and Kinnami end-to-end encryption. Passkeys never leave the Secure Enclave; only the public key is transmitted.

### NFR5: Privacy

Phone numbers are stored encrypted at rest and are deleted immediately after the user opts to decouple them. No logs retain the raw number.

**Implementation Note:** UI must clearly communicate privacy options to users.

## Blueprints

- Universal Onboarding and Identity Creation — Defines all UI components, user flows, and passkey integration requirements
- Frontend — Establishes SwiftUI architecture, MVVM patterns, view composition, and iOS security best practices

---

### WO-205: Implement Phone Number Management and Optional Deletion

**Type:** Build

**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Implement secure phone number storage with encryption at rest and provide the capability for users to optionally delete their phone number after DID creation. This component ensures privacy-first phone number handling throughout the onboarding lifecycle.

## In Scope

- Encrypted storage for phone numbers using project encryption standards
- Phone number encryption/decryption utilities
- Optional deletion endpoint and logic
- Verification that DID and passkey exist before allowing deletion
- Cascading deletion from all storage locations
- Audit logging for deletion events
- Data retention policy enforcement
- Secure key management for encryption
- Database schema for encrypted phone number storage

## Out of Scope

- SMS verification logic
- DID generation or management
- Trust score calculation
- User session management
- Frontend UI components
- Phone number validation (handled in onboarding orchestration)

## Requirements

The requirements document for Universal Onboarding and Identity Creation provides the following functional requirements:

### FR6: Optional Phone Number Decoupling

After DID creation, the user may choose to delete the stored phone number. If deleted, the number is removed from all persistent storage.

**Acceptance Criteria:**
- Phone number deletion only allowed after DID creation
- Deletion must remove the phone number from all storage locations
- Deletion must be irreversible
- System must continue functioning with DID and passkey after deletion
- Audit log entry must be created for deletion events

### NFR3: Security

All communication uses TLS 1.3 and Kinnami end-to-end encryption. Passkeys never leave the Secure Enclave; only the public key is transmitted.

**Implementation Note:** Phone numbers must be encrypted at rest.

### NFR5: Privacy

Phone numbers are stored encrypted at rest and are deleted immediately after the user opts to decouple them. No logs retain the raw number.

**Acceptance Criteria:**
- Phone numbers encrypted at rest using strong encryption
- No plaintext phone numbers in logs or audit trails
- Deletion executed immediately upon user request
- No backup or archive retains deleted phone numbers

### NFR6: Auditable Traceability

DID creation and passkey linking events are recorded on the Cardano metagraph with immutable transaction hashes for compliance.

**Implementation Note:** Phone number deletion events must also be auditable.

## Blueprints

- Universal Onboarding and Identity Creation — Defines phone number storage requirements, deletion logic, and privacy specifications
- Backend — Establishes encryption patterns, data security standards, and storage architecture
- Data Layer — Defines encrypted data storage patterns and database schema design

---

### WO-220: Build PSI Contact Discovery Backend Service

**Type:** Build

**Blueprint:** Privacy-Preserving Contact Discovery

## Summary

Build the server-side OPRF-based Private Set Intersection service that enables privacy-preserving contact discovery. The server applies its OPRF key to blinded client hashes and returns the evaluated values. It maintains a registry of registered user OPRF-evaluated phone hashes without ever learning what the client queried or which contacts matched.

## In Scope

- OPRF server key management: generate and securely store the server OPRF key (`k_server`); rotate annually
- `POST /v1/contacts/psi/blind-eval` endpoint: accept array of client-blinded hashes → apply server OPRF key (`r' = r × k_server`) → return `r'` values to client
- Registered user OPRF hash registry: on each user registration (if phone number provided), compute `OPRF.Evaluate(H(normalizedPhone), k_server)` and store in PostgreSQL PSI registry
- Registry update: on phone number deletion, remove user's OPRF entry from registry
- Discoverability filtering: only include Tier 3+ users in PSI registry by default; Tier 1–2 users are not discoverable unless they explicitly opt in
- Batch query support: accept up to 5,000 blinded hashes per request
- Privacy guarantee: server never stores the raw blinded query set — computation only, no logging of query content

## Out of Scope

- iOS PSI client (separate work order)
- Username discovery (separate work order)
- Phone number storage or management (handled by Universal Onboarding work orders)

## Requirements

From the Privacy-Preserving Contact Discovery blueprint:

**PSI Protocol (Server Side):**
1. Client sends blinded hashes: `r = H(phone) × k_client`
2. Server applies OPRF key: `r' = r × k_server`
3. Server returns `r'` values to client (server never sees original hashes)
4. Server separately maintains registered user OPRF-evaluated hashes for client comparison

**Server Privacy Guarantee:**
| What the Server Knows | What the Server Does NOT Know |
|---|---|
| A sync request was made from a DID | Which phone numbers were in the user's contacts |
| Total count of blinded hashes in query | Which of the queried hashes matched registered users |
| Timestamp of sync | Which contacts are now connected in-app |

## Blueprints

- Privacy-Preserving Contact Discovery — Defines the OPRF-based PSI protocol, server-side OPRF key application, registered user hash registry, discoverability opt-in defaults, and server privacy guarantees

---

### WO-221: Build PSI Contact Discovery iOS Client

**Type:** Build

**Blueprint:** Privacy-Preserving Contact Discovery

## Summary

Build the iOS PSI contact discovery client using OPRF-based Private Set Intersection (RFC 9497). On permission grant, the app normalizes phone numbers, blinds them locally, queries the PSI backend, unblind results, and resolves matching DIDs — without the server ever learning which contacts were queried. Includes `ContactDiscoverySettings` UI for full user control.

## In Scope

- `ContactDiscoveryService` actor using OPRF client library (RFC 9497 / IETF VOPRF draft)
- iOS Contacts permission prompt with clear explanation: "Find which of your contacts use ECHO without sharing your contact list"
- E.164 phone number normalization from iOS `CNContactStore`
- Client-side OPRF blinding: `r = H(normalizedPhone) × k_client` (random scalar per sync session)
- PSI intersection: unblind server response → compare against server registry → extract matching DIDs
- DID profile resolution: batch `GET /v1/profile/{did}` for matched DIDs → show "Contacts on ECHO" list
- `ContactDiscoverySettings` UI in Settings → Privacy:
  - Toggle: "Allow others to find me by phone number" (default: on for Tier 3+, off for Tier 1–2)
  - Sync frequency: manual, weekly, monthly
  - Last sync timestamp display
  - "Remove my phone from discovery" button
- Manual sync trigger from contacts list
- Permission denied state: show "Add contacts manually by QR code, username, or DID"

## Out of Scope

- PSI backend service (separate work order)
- Username/QR discovery (separate work order)
- Phone number storage/deletion (Universal Onboarding work orders)

## Requirements

From the Privacy-Preserving Contact Discovery blueprint:

**iOS Implementation:**
```swift
actor ContactDiscoveryService {
    func discoverContacts() async throws -> [DiscoveredContact] {
        // 1. Check Contacts permission
        // 2. Fetch + normalize phone numbers (E.164)
        // 3. PSI: OPRF blind → server eval → unblind → intersection
        // 4. Resolve DIDs to profiles from metagraph cache
        // 5. Filter to discoverable users only
    }
}

struct ContactDiscoverySettings {
    var allowDiscoveryByContacts: Bool  // Default: true for Tier 3+
    var lastSyncTimestamp: Date?
    var syncFrequency: SyncFrequency    // manual, weekly, monthly
}
```

## Blueprints

- Privacy-Preserving Contact Discovery — Defines the full iOS PSI client implementation, user opt-in controls, double opt-in model, data flow, and privacy guarantees

---

### WO-222: Implement Username Index, QR Code, and Invitation Link Contact Discovery

**Type:** Build

**Blueprint:** Privacy-Preserving Contact Discovery

## Summary

Implement the three alternative contact discovery methods that require no phone number: public username index on the metagraph Data L1, QR code generation and scanning, and time-limited invitation link generation. These are the primary discovery paths for users who decline phone-based PSI or delete their phone number.

## In Scope

- **Username system**: users set a unique `@username` (alphanumeric, 3–30 chars) → stored in public index on metagraph Data L1 (`{username: "@alice", did: "did:prism:..."}`) → queryable by anyone via `GET /v1/users/lookup?username=@alice`
- Username uniqueness enforcement: Data L1 validator rejects duplicate username registration
- **QR code**: generate QR code from user's DID + display name; display in profile; iOS `AVFoundation` scan → parse DID → add contact
- **Direct DID entry**: `GET /v1/users/lookup?did=did:prism:cardano:...` → fetch profile; add contact button
- **Invitation links**: `POST /v1/invites` → generate time-limited (`expiresIn: 1h|24h|7d`) invitation URL `echo://invite/{token}` → resolves to DID on tap; `maxUses: 1|unlimited`
- QR code display in profile settings and share sheet
- Deep link handling for `echo://invite/{token}` and `echo://user/{did}`

## Out of Scope

- PSI phone-based discovery (separate work orders WO-220, WO-221)
- Phone number storage (Universal Onboarding)
- Contact list management UI (WO-39, WO-187)

## Requirements

From the Privacy-Preserving Contact Discovery blueprint:

**Alternative Discovery Methods (No Phone Number Required):**
- **QR Code**: Every user has a unique QR code in their profile. Others scan to add directly.
- **Username**: Users set a public `@username` stored in a public index on metagraph Data L1 — discoverable by anyone.
- **Direct DID entry**: Advanced users enter a full `did:prism:cardano:...` identifier.
- **Invitation links**: One-time or time-limited links sharing a DID without revealing personal info.

## Blueprints

- Privacy-Preserving Contact Discovery — Defines username index on Data L1, QR code discovery, direct DID entry, and invitation link generation as phone-free contact discovery alternatives

---

### WO-228: Build Privacy Settings Screen, Encryption Indicator, and Account Deletion UI

**Type:** Build

**Blueprint:** Privacy Architecture and Secure Data Handling

## Summary

Build three privacy-focused iOS UI components specified in the Privacy Architecture blueprint: the Privacy Settings screen (last-seen, online status, read receipts, contact discovery controls), the E2E encryption indicator in conversation headers (lock icon showing encryption status and anchoring count), and the multi-step account deletion confirmation flow.

## In Scope

- **Privacy Settings Screen (`PrivacySettingsView`):**
  - Visibility controls: `showLastSeen`, `showOnlineStatus`, `showProfilePicture`, `showStatusMessage`, `allowCalls` — each with `.everyone / .contacts / .nobody` picker
  - Read receipts toggle (on/off globally)
  - Contact discovery opt-in toggle ("Allow others to find me by phone number")
  - Persisted in `PrivacySettings` struct; synced to Contacts Service (`PATCH /v1/profile/privacy`)
  - Accessible via Settings tab → Privacy
- **Conversation Encryption Indicator:**
  - Lock icon (🔒) in conversation header / navigation bar subtitle
  - Tap → `EncryptionInfoSheet` bottom sheet: algorithm name (X25519 + ChaCha20-Poly1305 for 1:1; AES-256-GCM for groups), anchoring status (N messages anchored on-chain), Phase 3 verification status
  - Always present — users should always be able to confirm their conversation is E2E encrypted
- **Account Deletion Flow (`AccountDeletionView`):**
  - Step 1: Warning — "Deleting your account is cryptographically irreversible. Your encryption keys will be destroyed."
  - Step 2: Token disposition — "Burn my ECHO balance" or "Transfer to: [wallet address]"
  - Step 3: Biometric confirmation — "Authenticate to confirm permanent deletion"
  - Step 4: Progress screen — shows deletion steps completing (keys deleted, queue wiped, DID deactivated)
  - Calls `DELETE /v1/account` backend endpoint (WO-218 handles backend logic)

## Out of Scope

- Backend GDPR erasure (WO-218)
- T0–T7 CI enforcement (WO-217)
- Contact blocking UI (WO-190)
- Profile privacy controls on individual contacts (WO-39)

## Requirements

From the Privacy Architecture and Secure Data Handling blueprint:

**Privacy settings screen:** Controls for last-seen visibility, online status, profile picture access, read receipts, and contact discovery opt-in.

**Encryption indicator:** All conversation headers display a lock icon indicating E2E encryption status; tapping shows the encryption spec and commitment anchoring status.

**Account deletion flow:** Multi-step confirmation with explicit warning that deletion is cryptographically irreversible; key destruction confirmation screen.

## Blueprints

- Privacy Architecture and Secure Data Handling — Defines privacy settings screen, encryption indicator, and account deletion UI requirements

---

### WO-234: Implement 24-Word BIP-39 Recovery Phrase Generation and Verification UI

**Type:** Build

**Blueprint:** Secure Enclave Key Management

## Summary

Implement FR8 of the Secure Enclave Key Management blueprint: a 24-word BIP-39 mnemonic recovery phrase displayed once during initial setup, with mandatory 3-word confirmation before proceeding. On a new device, entering the phrase generates a new Secure Enclave key pair and triggers a DID document key rotation on Cardano.

## In Scope

- **BIP-39 mnemonic generation:** Derive 24-word phrase from `{SecureEnclave.publicKey.rawRepresentation + userSalt}` using BIP-39 word list. Never generated on server — computation is entirely local on device
- **`RecoveryPhraseView`:** Display 24 words in numbered grid; no copy-to-clipboard option (security); "I have saved my recovery phrase" toggle required before proceeding; words displayed in SFMono font with large text
- **Mandatory confirmation step:** After display, user must tap 3 randomly selected words in correct order from a scrambled word picker — if wrong, return to display view; 3 attempts max before "come back later" cooldown (10 minutes)
- **Phrase storage:** Never transmitted to any server; never stored on-device after setup; only displayed once. User's responsibility to write down physically
- **New device recovery flow:** Settings → "Set Up New Device" → enter 24 words → validate BIP-39 checksum → generate new Secure Enclave key pair on new device → `POST /identity/devices` with `{newPublicKey, signature(challenge)}` → backend triggers Cardano DID document rotation adding new public key
- **`BiometricLockoutView`:** Countdown timer showing minutes remaining (from FR6 lockout in WO-211), device passcode fallback button, support link, brief explanation

## Out of Scope

- Multi-device QR authorization (WO-211, FR7)
- Full SecureEnclaveManager (WO-223)
- Passkey WebAuthn setup (WO-136)

## Requirements

From the Secure Enclave Key Management blueprint:

**FR8 — Recovery phrase:** During initial setup, the app displays a 24-word BIP-39 mnemonic derived from Secure Enclave public parameters + user passphrase. Displayed exactly once, never stored on any server. On a new device, entering the phrase generates a new Secure Enclave key pair and triggers a DID document key rotation on Cardano.

**API:**
- `POST /identity/devices` — receives new device P-256 public key; authenticated by recovery phrase challenge; backend adds public key to Cardano DID document

## Blueprints

- Secure Enclave Key Management — Defines FR8 recovery phrase generation, `RecoveryPhraseView` display and confirmation requirements, and new device recovery via DID key rotation

---

### WO-287: Implement Session JWT Issuance, Refresh, and Revocation Backend

**Type:** Build

**Blueprint:** Backend, Universal Onboarding and Identity Creation

## Summary

Implement the Identity Service backend for session JWT issuance via challenge-response, token refresh, and per-device session revocation. This is Tier 2 of the three-tier auth model — every API call after onboarding uses a Bearer JWT issued here. Tier 3 per-request DID signatures (high-value ops) are handled by WO-1; registration tokens are handled by WO-203. This WO is the missing bridge that makes normal authenticated API usage work.

## In Scope

- `POST /v1/auth/challenge` — issues a 32-byte cryptographically random nonce tied to the DID, 5-minute expiry, stored in Redis (one-time, replay-prevented)
- `POST /v1/auth/token` — verifies ECDSA P-256 signature over `SHA-256(challenge || did_key || timestamp_seconds)`, confirms challenge not expired/replayed, confirms `device_id` is registered for the DID; issues EdDSA-signed session JWT (2-hour expiry) + opaque refresh token (7-day expiry)
- `POST /v1/auth/refresh` — single-use refresh token rotation; returns new JWT + new refresh token, invalidates the old refresh token
- `DELETE /v1/auth/session` — logout; marks refresh token as revoked in `sessions` table
- `sessions` PostgreSQL table: `user_did`, `device_id`, `refresh_token_hash` (SHA-256), `issued_at`, `expires_at`, `last_used_at`, `revoked_at`
- Redis nonce store for replay prevention (one-time challenge nonces, 5-minute TTL)
- JWT payload: `sub` (did:key), `exp` (+2h), `jti` (unique per-token nonce), `tier` (trust tier at issuance), `device_id`, `scope: "full"`
- JWT signed with backend Ed25519 key; `kid` header enables key rotation without breaking existing tokens
- Per-device revocation: `DELETE /v1/account/devices/:device_id` invalidates all `sessions` rows for that `device_id`
- Account-level revocation on account deletion: invalidate all sessions for `user_did`
- Suspicious activity path: backend may return `401` with `"error": "reauth_required"` to force a new challenge-response cycle

## Out of Scope

- Per-request DID signature validation for high-value operations (WO-1)
- Registration token issuance during onboarding (WO-203)
- Device public key registration (WO-1, WO-203)
- iOS session JWT storage and attachment to requests (WO-6, WO-203)

## Requirements

**NFR8 — Token Security:** Session JWTs expire within 2 hours. Refresh tokens are single-use, 7-day expiry, stored hashed in PostgreSQL. Token revocation propagates within 30 seconds. Challenge nonces are single-use with a 5-minute window to prevent replay attacks.

**Challenge-Response Issuance Flow:**
```
1. Client:  POST /v1/auth/challenge
   Body:    { "did_key": "did:key:z6Mk..." }
   Response: { "challenge": "<32-byte-random-base64url>", "expires_at": "T+5min" }

2. Client:  Sign challenge using Secure Enclave P-256 key (biometric required)
   Input:   SHA-256(challenge || did_key || timestamp_seconds)
   Output:  ECDSA P-256 signature (DER-encoded, base64url)

3. Client:  POST /v1/auth/token
   Body:    { "did_key": "did:key:z6Mk...", "challenge": "<same-32-byte-value>",
              "signature": "<base64url-ecdsa-signature>", "device_id": "<opaque-device-identifier>" }

4. Backend: Decode public key from did:key identifier directly (no chain lookup)
            Verify: SHA-256(challenge || did_key || timestamp) matches signature
            Verify: challenge not expired, not already used (one-time)
            Verify: device_id is registered for this did:key
            Issue:  session JWT + refresh token

5. Response: { "access_token": "<session-jwt>", "refresh_token": "<opaque-32-bytes>",
               "expires_in": 7200, "token_type": "Bearer" }
```

**Session JWT Format:**
```json
// Header
{ "alg": "EdDSA", "typ": "JWT", "kid": "backend-ed25519-2026-v1" }
// Payload
{ "iss": "https://api.echo.app", "sub": "did:key:z6Mk...", "iat": 1714000000,
  "exp": 1714007200, "jti": "unique-nonce-per-token", "tier": 2,
  "device_id": "opaque-device-identifier", "scope": "full" }
```

**Sessions Table:**
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_did TEXT NOT NULL,
    device_id TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,  -- SHA-256 of the opaque token
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,   -- 7 days
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    UNIQUE(refresh_token_hash)
);
CREATE INDEX idx_sessions_user_did ON sessions(user_did) WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_device_id ON sessions(device_id) WHERE revoked_at IS NULL;
```

**Revocation Matrix:**

| Scenario | Method |
|---|---|
| Logout | Invalidate refresh token in `sessions`; JWT expires naturally in ≤2h |
| Device revoked | Invalidate all refresh tokens for `device_id` |
| Account deleted | Invalidate all refresh tokens for `user_did` |
| Suspected compromise | Add `jti` to Redis revocation list (TTL = remaining JWT expiry) + invalidate all refresh tokens |
| Backend key rotation | Old `kid` tokens valid until `exp`; new tokens use new `kid` |

## Blueprints

- Universal Onboarding and Identity Creation — defines the three-tier auth model, challenge-response JWT flow, session table schema, refresh token design, and revocation scenarios
- Backend — specifies Identity Service (port 8001), auth middleware patterns, and error response format

---

### WO-288: Build New Device Login via QR Transfer and Recovery Phrase Flow

**Type:** Build

**Blueprint:** Decentralized Identity and Authentication, Universal Onboarding and Identity Creation

## Summary

Implement the two "sign in on new hardware" paths for returning users: QR-based device transfer (old device generates a QR; new device scans it, old device biometrically confirms, new device key is registered) and recovery phrase login (24-word BIP-39 phrase derives a new Secure Enclave key pair, new did:key registered on the Constellation Identity Metagraph). WO-234 covers phrase display during initial setup; this WO covers using that phrase to recover access on new hardware.

## In Scope

**Backend:**
- `POST /v1/login/link-device/initiate` (authenticated on old device, `X-DID-Signature` required) — generates `link_token` (UUID), `qr_payload` (base64), `expires_at` (T+5min)
- `POST /v1/login/link-device/complete` (from new device, `X-DID-Signature` required) — verifies `link_token` not expired/used, appends `new_public_key` to DID's device key set in `PostgresDIDRegistry`, submits Identity Metagraph VC update, issues full session JWT
- `POST /v1/login/recover` (`X-DID-Signature` required) — accepts `phrase_commitment = H(phrase || nonce)` (proves possession without transmitting the phrase), registers new Secure Enclave key pair against the DID, submits Identity Metagraph VC key rotation, issues session JWT
- Token idempotency: retrying `link-device/complete` with the same `link_token` and same `new_public_key` is safe; different DID returns error (prevents token hijacking)

**iOS:**
- "Sign in on new device" path on the Login / Register screen → QR scanner view
- Settings → Account → "Link New Device" → QR code display with 5-minute countdown timer (screenshot blocked)
- Biometric confirmation prompt on old device when new device scans the QR
- Success flow on new device: "Device linked — message history stays on your old device"
- Recovery phrase entry screen: 24-word input grid, phrase validation, mandatory 3-word challenge before proceeding
- Success flow: "Account recovered — no message history (privacy by design)"

## Out of Scope

- BIP-39 recovery phrase generation and display during initial account creation (WO-234)
- Multi-device registration initiated during fresh onboarding (WO-1)
- Identity re-verification recovery path ("I lost my recovery phrase" for Tier 4 users) — deferred to identity verification WOs
- Session JWT issuance service (WO-287 / separate WO)

## Requirements

**NFR5 — Biometric Lockout:** 5 consecutive failures → device passcode fallback. 10 total failures within 1 hour → 15-minute lockout.

**NFR4 — Recovery Phrase Security:** Screenshot disabled on recovery phrase entry screen. Screen does not appear in app switcher preview. Phrase never transmitted to or stored by any server.

**QR Transfer — What Transfers:**
```
✅ DID identity
✅ New device did:key registered on Constellation Identity Metagraph
✅ Group keys re-encrypted for new device (async, post-link)
❌ Message history (stays on old device — privacy by design)

Requires confirmation on the existing trusted device.
Prevents account takeover even if new device is immediately stolen.
```

**API Specification:**
```
POST /v1/login/link-device/initiate     (authenticated on OLD device; X-DID-Signature required)
Response: { "link_token": "uuid", "qr_payload": "base64", "expires_at": "T+5min" }

POST /v1/login/link-device/complete     (from NEW device; X-DID-Signature required)
Body:     { "link_token": "uuid", "new_public_key": "base58-p256", "device_name": "iPhone 16" }
Response: { "did": "did:key:z6Mk...", "trust_tier": N, "vc_id": "uuid" }

POST /v1/login/recover                  (X-DID-Signature required)
Body:     { "public_key": "base58-p256", "phrase_commitment": "H(phrase||nonce)", "device_name": "..." }
Response: { "did": "did:key:z6Mk...", "trust_tier": N }
Note: phrase_commitment proves possession without transmitting the phrase itself
```

**Account Device Management (also in scope):**
```
GET    /v1/account/devices              → List all linked devices (name, linked_at, last_used)
DELETE /v1/account/devices/:device_id  → Revoke a device (X-DID-Signature required);
                                          removes from PostgresDIDRegistry + invalidates all sessions
```

## Blueprints

- Universal Onboarding and Identity Creation — defines QR transfer flow, recovery phrase login, link-device endpoints, phrase_commitment design, and device management API
- Decentralized Identity and Authentication — specifies Identity Metagraph VC update for new device key registration and key rotation

---

## Echo Passport — Verifiable Credential Wallet (4th Product) — Wave 0 + Wave A (6)

> New product. Full plan: `docs/ECHO_PASSPORT_PLAN.md`. ADRs: `0003` (custody), `0004` (recovery).
> All Backlog. Echo Passport is the holder-side wallet + selective-disclosure + presentation
> layer over credentials Echo already issues — **not** a PII store. Hybrid custody (L1 credentials
> client-encrypted + synced; L2 raw artifacts device-only by default; L3 on-chain hashes only).

### Wave 0 — Recovery & Guardian Mini-Design ✅ (design complete)

> **Note:** Software Factory **WO-292** is already taken (Phase 1 Glacial first-run onboarding).
> Wave 0 is tracked here and in **ADR 0004** (Accepted 2026-05-29), not as a separate SF ticket.

**Status:** ✅ Design complete · **Type:** Design spike (docs only) · **Output:** ADR `0004` + WO-296 spec
**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Design-only wave preceding any recovery code. Social/threshold recovery is the make-or-break for
a non-custodial vault's UX; this WO settles the scheme before implementation so WO-296 builds
against a reviewed design.

## In Scope

- Threshold scheme decision: Shamir M-of-N split of the Passport recovery secret.
- Guardian model: user devices + optional trusted contacts + optional Comply org (institutional guardian).
- Share distribution, rotation, and revocation UX; guardian-acceptance-as-signed-VC format.
- Collusion / coercion / lost-share threat model; how it layers on the existing 24-word BIP-39 + SMS recovery.
- Output `docs/adr/0004-echo-passport-recovery-model.md` and the WO-296 implementation spec.

## Out of Scope

- Any code. Guardian bonds / on-chain incentives (deferred to WO-303, Phase 5).

**Acceptance Criteria:**
- ADR 0004 merged with a chosen M-of-N scheme and guardian taxonomy.
- Threat model enumerates collusion, coercion, and total-share-loss paths with mitigations.
- **No design option permits a server-held decryption key** — verified against the T0–T7 invariant.

### WO-293: Echo Passport — Holder Data Model + Aggregation API (Wave A)

**Status:** 🔄 In Progress · **Depends:** WO-274 (VC 2.0), WO-272 (Identity L1), Wave 0 (ADR 0004)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Backend holder model and aggregation API: list/query/present credentials a user holds across
issuers. Wraps the existing issuance/verification pipeline rather than re-implementing it.

## In Scope

- New `pkg/passport/` package: holder model (`CredentialRef` = opaque UUID → issuer DID, type, hash, status index; **no PII**).
- `GET /v1/passport/credentials` (list/get), `POST /v1/passport/present` (begin/accept — reuse `pkg/credentials/oidc4vc/`).
- Migration `passport_credential_ref`.
- Reuse `pkg/credentials/{models,verifier,statuslist_l1,revocation}.go` for verification + revocation status.

## Out of Scope

- Client-side encryption / sync (WO-294). Selective disclosure derivation (WO-295). Metagraph state (WO-298, Phase 4).

**Acceptance Criteria:**
- Holder can enumerate credentials with live StatusList2021 status.
- Presentation flow reuses OIDC4VCI; no new verification code path.
- T0–T7 CI green: no PII in `passport_credential_ref` or API payloads.

### WO-294: Echo Passport — Client-Encrypted Credential Sync (Wave A)

**Status:** 📋 Backlog · **Depends:** WO-293, WO-33 (Pinata/Storj)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Make credentials available on web / glasses / agent without the phone, by syncing
**client-encrypted** credential blobs. Server stores ciphertext + CID only.

## In Scope

- Generalize the encrypted-blob client out of `internal/logging/ipfs_clients.go` into `pkg/storage/encblob/`.
- Key hierarchy (extend `internal/crypto/keyderivation.go`): `PassportRootKey` (Secure Enclave + recovery secret) → `CredentialSyncKey = HKDF(PassportRootKey, "echo-passport-credential-sync")` → AES-256-GCM.
- `POST/GET /v1/passport/sync` (push/pull ciphertext); migration `passport_sync_blob`.
- Update `docs/data-classification.md` to add the client-encrypted-sync case (T2) so CI rules don't false-positive.

## Out of Scope

- Layer-2 raw-artifact backup (opt-in; later). Recovery (WO-296).

**Acceptance Criteria:**
- Server can never decrypt synced blobs (no key material server-side) — verified by test fixtures that *should* trip Semgrep.
- A second device with the recovery secret can pull + decrypt the credential set.
- `data-classification.md` updated and CI green.

### WO-295: Echo Passport — Selective Disclosure via SD-JWT (Wave A)

**Status:** 📋 Backlog · **Depends:** WO-293
**Blueprint:** Decentralized Identity and Authentication

## Summary

Present the minimum claim ("over 21", "verified resident") instead of the whole credential or the
source document. Completes the SD-JWT path already stubbed in `pkg/credentials/formats.go`.

## In Scope

- `pkg/passport/disclosure/`: full SD-JWT per-claim disclosure derivation + verifier-side check.
- Wire into `POST /v1/passport/present` so a verifier requests a claim subset.

## Out of Scope

- BBS+ / ZK predicate proofs (optional, Wave D / WO-308). Raw-artifact presentation (never — present the VC).

**Acceptance Criteria:**
- A presentation discloses only requested claims; undisclosed claims are cryptographically withheld.
- Verifier validates the SD-JWT against issuer signature + StatusList2021.

### WO-296: Echo Passport — Social-Threshold (Shamir) Recovery (Wave A)

**Status:** 📋 Backlog · **Depends:** WO-274 (VC 2.0), WO-272 (Identity L1), Wave 0 (ADR 0004)
**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Implement the recovery model from ADR 0004: Shamir M-of-N split of the recovery secret across
user devices + optional guardians, layered on the existing BIP-39 + SMS flow.

## In Scope

- `pkg/passport/recovery/`: Shamir split/combine, share distribution, guardian-acceptance VC, rotation.
- Migration `passport_recovery_share` (holds **share metadata only**, never reconstructable material server-side).
- `POST /v1/passport/recovery/{setup,initiate,complete}`.

## Out of Scope

- Guardian bonds / device-rotation hardening (WO-303, Phase 5).

**Acceptance Criteria:**
- M-of-N reconstruction round-trips; M-1 shares reveal nothing (unit test).
- **No server-held key under any path** (honeypot line); CI green.
- Lost-device recovery restores the synced credential set on a fresh device.

### WO-297: Echo Passport — iOS Passport Module (Wave A)

**Status:** 📋 Backlog · **Depends:** WO-293, WO-294, WO-295, WO-296
**Blueprint:** Decentralized Identity and Authentication

## Summary

iOS surface for the Passport: credential list/detail, present flow (selective disclosure),
recovery setup. SwiftData encrypted at rest (existing AES-256-GCM model); Secure Enclave for
`PassportRootKey`.

## In Scope

- New iOS Passport module beside existing identity/credential code.
- Credential list/detail, present-with-disclosure UI, recovery setup UI.
- Biometric gate reuses the hidden-persona re-auth pattern.

## Out of Scope

- Pay-in-chat UX (WO-300, Phase 4). Agent/glasses surface (Wave D).

**Acceptance Criteria:**
- `swift build` library + security-test targets pass (hard gates).
- Present flow shows only disclosed claims; raw artifacts never leave the device.

---
