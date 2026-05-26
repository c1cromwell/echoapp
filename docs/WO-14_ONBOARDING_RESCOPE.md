# WO-14 — Onboarding re-scope (credential path after Phase 1)

**Date:** 2026-05-26  
**Software Factory:** WO-14 marked **completed**; OIDC4VC wallet flow remains **WO-100**.

## Decision

WO-14 originally specified a separate `OnboardingCoordinator` and Cardano-centric DID progress copy. Phase 1 shipped **`FirstRunCoordinator`** (WO-292) with `did:key`, Secure Enclave passkeys, recovery, and optional VIP verification instead.

**WO-14 is fulfilled by `FirstRunCoordinator` + backend deps**, not a parallel coordinator tree.

## View / step mapping

| WO-14 spec name | Implemented as | Path |
|-----------------|----------------|------|
| `OnboardingCoordinator` | `FirstRunCoordinator` | `Features/Onboarding/FirstRun/FirstRunCoordinator.swift` |
| `UsernameView` | `DisplayNameEntryView` | Same folder; polls `GET /v1/users/check-username` |
| `PasskeySetupView` | `BiometricEnrollmentView` (sheet from `OnboardingOptionsView`) | Face ID / passkey enrollment |
| `DIDCreationProgressView` | Silent provision inside `BiometricEnrollmentView` / `SilentProvisionService` | No Cardano spinner — `did:key` register |
| `OptionalVerificationView` | `VIPPathView` + `EnrollmentMethodPicker` (wallet / mDL branches) | Credential path; full OIDC4VC = WO-100 |
| `OnboardingCompleteView` | `RecoverySetupView` completion → `onComplete` callback | App shell |

## Backend dependencies (WO-14)

| Endpoint | Status |
|----------|--------|
| `GET /v1/users/check-username` | ✅ Public route (`username_handlers.go`) |
| `POST /identity/register` | ✅ did:key (`pkg/didkey/`) |
| Session JWT after passkey | ✅ WO-287 refresh/revoke |

## Out of scope for WO-14 (other WOs)

| Item | WO |
|------|-----|
| OIDC4VC wallet registration client | WO-100 |
| PSI contact discovery iOS | WO-221 |
| Universal phone-first onboarding | WO-203 / WO-204 |

## Verification

- iOS: `UsernameValidator` + `UsernameAvailabilityService` + debounced UI in `DisplayNameEntryView`
- Go: `internal/api/username_handlers_test.go`
- Manual: first-run flow → username taken/available states against `make dev`
