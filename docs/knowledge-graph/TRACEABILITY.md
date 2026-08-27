# Echo traceability graph

Generated `2026-08-27T23:47:35Z` from Factory CSV `docs/echo-work-orders-2026-05-08-155805.csv` + `docs/phase-*-work-orders.md` (last SF sync in headers: 2026-08-27 (factory-live.json)).

Canonical Factory: [factory.8090.ai](https://factory.8090.ai) · project `0f477a55-4e26-4500-98f3-a69e4f3381b3`.

Live MCP pull: `docs/knowledge-graph/factory-live.json` (326 WOs).

Rebuild: `make knowledge-graph`. Agents: skill `echo-knowledge-graph`.

## Chain

```text
Requirements section  ==  Blueprint title  (same name in Echo)
        ↓
Work order  (factory.8090.ai URL on every WO node)
        ↓
Code path  (FEATURES overlay + gap audits)
        ↓
2-device E2E scenario  (docs/E2E_TWO_DEVICE.md)
```

## Phase status (markdown export)

| Phase | Total (header) | Last synced | Status summary |
|-------|----------------|-------------|----------------|
| 1 | 26 | 2026-08-27 (factory-live.json) | 26 Completed |
| 2 | 31 | 2026-08-27 (factory-live.json) | 20 Completed, 9 Backlog, 2 Blocked |
| 3 | 37 | 2026-08-27 (factory-live.json) | 24 Completed, 13 Backlog |
| 4 | 32 | 2026-08-27 (factory-live.json) | 19 Completed, 13 Backlog |
| 5 | 65 | 2026-08-27 (factory-live.json) | 15 Completed, 9 In Progress, 11 Ready, 30 Backlog |
| 6 | 32 | 2026-08-27 (factory-live.json) | 5 Completed, 1 In Progress, 26 Backlog |
| 7 | 103 | 2026-08-27 (factory-live.json) | 40 Completed, 1 In Progress, 62 Backlog |

## Work order status (live Factory)

| Status | Count |
|--------|-------|
| completed | 149 |
| in_progress | 11 |
| ready | 11 |
| backlog | 153 |
| blocked | 2 |
| **total** | **326** |

**Nodes in graph.json:** 324 (CSV + phase md + live IP/ready/blocked).

Per-WO status in graph.json overlays factory-live for in_progress / ready / blocked. Completed vs backlog counts above are the live Factory totals.

### In progress now

| WO | Title | Phase |
|----|-------|-------|
| WO-31 | Screen Sharing with Encrypted Content Transmission | 6 |
| WO-236 | Go Backend ZK Verification Service (Midnight) | 5 |
| WO-257 | PQ Cryptography Mode iOS Client | 5 |
| WO-309 | Echo Comply eDiscovery / Matter UI | 7 |
| WO-321 | Sept 1 2026 Echo Messaging Go-Live Epic | 5 |
| WO-322 | Go-Live W1: Corporate start + Week A E2E | 5 |
| WO-330 | Signal Parity Epic Waves S1–S5 | 5 |
| WO-333 | Signal Parity S1 — History sync + cloud backup | 5 |
| WO-335 | Telegram Parity Epic Waves T0–T3 | 5 |
| WO-337 | Telegram Parity T1 — Saved Messages, drafts, forward, search | 5 |
| WO-338 | Telegram Parity T2 — Folders, scheduled HTTP, bots | 5 |

### Ready (go-live / parity queue)

| WO | Title | Phase |
|----|-------|-------|
| WO-323 | Go-Live W2: Week A finish; Week B; App Store P0 | 5 |
| WO-324 | Go-Live W3: Bank + legal URLs; account deletion | 5 |
| WO-325 | Go-Live W4: Internal TestFlight RC1 | 5 |
| WO-326 | Go-Live W6: External TestFlight / Beta Review | 5 |
| WO-327 | Go-Live W5: Internal TF bake + ASC metadata | 5 |
| WO-328 | Go-Live W7: Go/No-Go meeting + buffer | 5 |
| WO-329 | Go-Live Launch Day (Sept 1) | 5 |
| WO-331 | Signal Parity S2 — Sealed-sender + Sender Keys | 5 |
| WO-332 | Signal Parity S4 — View-once, NSE, GIF | 5 |
| WO-334 | Signal Parity S3 — Group signals + calls polish | 5 |
| WO-336 | Telegram Parity T0 — Durable channels + admin | 5 |

## Blueprints → requirements → WOs

Requirement Titles in the CSV export are empty; each blueprint title is the requirements-document section of the same name.

| Blueprint / requirement | WOs | completed | in_progress | backlog | blocked |
|-------------------------|-----|-----------|-------------|---------|---------|
| (unlinked) | 46 | 0 | 8 | 27 | 0 |
| Decentralized Identity and Authentication | 23 | 5 | 4 | 8 | 4 |
| ECHO Comply — Enterprise Compliance Messaging | 21 | 0 | 0 | 21 | 0 |
| Backend | 16 | 1 | 2 | 12 | 0 |
| Broadcast Channels and Community Features | 14 | 0 | 0 | 14 | 0 |
| User Rewards Tracker on Profile | 14 | 0 | 0 | 14 | 0 |
| ECHO Token Reward System and Incentive Economy | 13 | 0 | 0 | 13 | 0 |
| Decentralized Bot Framework and Automation | 12 | 0 | 0 | 12 | 0 |
| Disappearing Messages with Cryptographic Verification | 11 | 0 | 0 | 11 | 0 |
| In-App High-Assurance Identity Verification and Reward | 11 | 0 | 0 | 10 | 0 |
| Dynamic Trust Network and Social Verification | 10 | 0 | 0 | 9 | 1 |
| Public and Private Groups with Verified Status Display | 10 | 0 | 0 | 10 | 0 |
| Verified Financial Institution Integration | 10 | 0 | 0 | 10 | 0 |
| Advanced Message Search and Archive System | 9 | 0 | 0 | 9 | 0 |
| Data Layer | 9 | 1 | 3 | 4 | 1 |
| Frontend | 9 | 0 | 0 | 8 | 0 |
| Large File Sharing and Cloud Storage Integration | 9 | 0 | 0 | 8 | 0 |
| ECHO Token Economics and Founder Allocation | 8 | 0 | 0 | 8 | 0 |
| Hidden Folders with Biometric Protection | 8 | 0 | 0 | 8 | 0 |
| Blockchain-Anchored Messaging with Provable Integrity | 7 | 0 | 0 | 7 | 0 |
| Infrastructure | 7 | 1 | 1 | 5 | 0 |
| Production Launch | 7 | 1 | 1 | 5 | 0 |
| Streamlined Onboarding with Verifiable Credentials and Passkeys | 7 | 0 | 0 | 6 | 0 |
| Universal Onboarding and Identity Creation | 7 | 0 | 0 | 7 | 0 |
| Voice and Video Calls with Screen Sharing | 7 | 0 | 1 | 6 | 0 |
| and Deployment | 7 | 1 | 1 | 5 | 0 |
| Multiple Personas with Selective Visibility | 6 | 0 | 0 | 6 | 0 |
| Enterprise Organization Profiles with Verified Status | 5 | 0 | 0 | 5 | 0 |
| Message Reactions | 5 | 0 | 0 | 5 | 0 |
| Polls | 5 | 0 | 0 | 5 | 0 |
| Privacy Architecture and Secure Data Handling | 5 | 0 | 0 | 5 | 0 |
| Silent and Scheduled Private Chats | 5 | 0 | 0 | 5 | 0 |
| and Interactive Elements | 5 | 0 | 0 | 5 | 0 |
| ECHO Comply — Law Firms (Chain-of-Custody) | 4 | 0 | 0 | 4 | 0 |
| ECHO Tokenomics | 4 | 0 | 0 | 4 | 0 |
| Founder Allocation | 4 | 0 | 0 | 4 | 0 |
| Secure Enclave Key Management | 4 | 0 | 0 | 4 | 0 |
| and Token Launch | 4 | 0 | 0 | 4 | 0 |
| Post-Quantum Cryptography Mode | 3 | 0 | 1 | 2 | 0 |
| Privacy-Preserving Contact Discovery | 3 | 0 | 0 | 3 | 0 |
| Zero-Knowledge Proofs and Midnight Integration | 3 | 0 | 1 | 2 | 0 |
| Data Sovereignty Layer | 2 | 0 | 0 | 2 | 0 |
| ECHO Comply — Local Government (FOIA) | 2 | 0 | 0 | 2 | 0 |
| End-to-End Message Encryption and Commitment | 2 | 0 | 0 | 2 | 0 |
| Portable Social Graph and Protocol Layer | 2 | 0 | 0 | 2 | 0 |
| Privacy Commons Treasury | 2 | 0 | 0 | 2 | 0 |
| Privacy-Preserving Blockchain Data Model | 2 | 0 | 0 | 2 | 0 |
| ECHO Comply — Healthcare (HIPAA) | 1 | 0 | 0 | 1 | 0 |

## Product features (code reality)

| Feature | Status | Sept 1 | E2E | Factory blueprints |
|---------|--------|--------|-----|--------------------|
| did:key identity + passkey REST | shipped | yes | Both devices complete first-run; REST calls carry X-Sender-DID + X-Signature | Decentralized Identity and Authentication, Secure Enclave Key Management |
| Frozen onboarding + Glacial login | shipped | yes | Do not redesign. Validate both devices authenticate. Login is frozen. | Streamlined Onboarding with Verifiable Credentials and Passkeys, Universal Onboarding and Identity Creation |
| Content-blind WebSocket relay | shipped | yes | D2: live DM while both apps foregrounded | End-to-End Message Encryption and Commitment, Backend |
| Durable offline queue + APNs wake | partial | yes | D3: B force-quit → A sends → queue + optional APNs → B opens → flush | End-to-End Message Encryption and Commitment, Backend |
| Typing, read receipts, reactions | shipped | yes | D4: typing / receipts / reactions on both devices | Message Reactions, Polls, and Interactive Elements |
| @username public invite | shipped | yes | D1: A shares echo://invite?u=handle → B accepts → both in contacts | Privacy-Preserving Contact Discovery |
| QR identity exchange | shipped | yes | D1b: Profile QR → other device scans → POST /v3/contacts/add | Privacy-Preserving Contact Discovery |
| Phrase-encrypted backup + restore-did | partial | yes | D5: A backups → B restores phrase → chats return. Wallet must be linked. | Data Sovereignty Layer, Universal Onboarding and Identity Creation |
| Encrypted groups | shipped | yes | D6: create group, add B, bidirectional messages, offline fan-out | Public and Private Groups with Verified Status Display |
| Broadcast channels | partial | no | Optional: create/subscribe/post. Do not market as durable without Postgres soak. | Broadcast Channels and Community Features |
| Telegram-class messaging UX | partial | no | T1 two-device: Saved Messages, drafts, forward, search. T0 channels soak. T2 folder + schedule HTTP. | Telegram-class Messaging UX, Broadcast Channels and Community Features |
| 1:1 voice/video calls | partial | no | Optional if WebRTC linked. Missed-call push needs APNs. | Voice and Video Calls with Screen Sharing |
| Rewards hub (earn, not buy) | partial | yes | Smoke: Rewards tab loads; no withdraw/VIP purchase CTA | ECHO Token Reward System and Incentive Economy, User Rewards Tracker on Profile |
| Hidden folders (night palette) | partial | no | Per-device vault. Night tokens only here. | Hidden Folders with Biometric Protection |
| Identity L1 username + trust anchors | shipped | yes | Needed for validate-phase1 step 3 and username anchors. make start-identity. | Decentralized Identity and Authentication, Dynamic Trust Network and Social Verification |

## Sept 1 critical WOs (open in Factory)

Use the URL on each node in `graph.json`. Highest-signal numbers:

| WO | Why |
|----|-----|
| [WO-230](https://factory.8090.ai/project/0f477a55-4e26-4500-98f3-a69e4f3381b3) | Phase 1 go/no-go (`make validate-phase1`) |
| WO-4 / messaging core | Encryption + delivery |
| WO-221 / WO-222 | PSI (optional) + invite deep links |
| WO-288 | Link new device (Week B B6) |
| WO-233 | App Store go-live |

Full node list: [`graph.json`](graph.json).

