---
name: echo-work-order-sync
description: >-
  Reconciles Echo Software Factory work orders with codebase reality and updates
  WO status via MCP. Use when syncing work orders, updating phase doc headers,
  after large merges, or when the user asks to align tickets with shipped code.
---

# Echo work order sync

## Source-of-truth hierarchy

1. **Code + tests + CHANGELOG** — what is actually shipped  
2. **Gap audits** — `docs/PHASE2_GAP_AUDIT.md`, feature specs (`PHASE3_IOS_UI_SPEC.md`)  
3. **Software Factory** (`software-factory-echo` MCP) — canonical WO **status**  
4. **`docs/phase-*-work-orders.md`** — human export; update headers after SF sync  

Never mark `in_review` unless the user explicitly asks (MCP server rule).

## MCP workflow

```
1. whoami
2. list_work_orders (filters: phase_number, assignee_name, status)
3. read_work_order for candidates
4. Grep/codebase search for evidence (paths in WO description, CHANGELOG, commits)
5. edit_work_orders — status only unless user asked for title/description
6. edit_work_order_description — full replacement markdown when scope changed
7. Update docs/phase-N-work-orders.md status summary + "Last synced: YYYY-MM-DD"
```

## Status decision rubric

| Evidence | Status |
|----------|--------|
| No code, WO in scope | `backlog` or `ready` |
| Partial / stubs / missing acceptance criteria | `in_progress` |
| Code + tests match acceptance; user confirmed or obvious ship | `completed` |
| Cardano/PRISM spec superseded by did:key | `blocked` or `completed` with rewritten description |
| Active development | `in_progress` |

## Obsolete WOs (do not implement as written)

| WO | Replace with |
|----|--------------|
| 180, 182 | WO-273, WO-274 (Constellation + did:key) |
| 132 (Cardano VC) | WO-274 + WO-100 iOS; mark completed when Constellation path ships |

Always read `docs/PHASE2_GAP_AUDIT.md` before closing Phase 2 WOs.

## Code evidence shortcuts

| WO area | Look for |
|---------|----------|
| Passkey auth | `internal/api/passkey_auth.go`, `PasskeySigningInterceptor` |
| did:key | `pkg/didkey/`, `POST /identity/register` |
| VC / OIDC4VC | `pkg/credentials/oidc4vc/`, WO-274 |
| Metagraph L1 | `metagraph/.../IdentityValidations.scala`, `make metagraph-test` |
| T0–T7 CI | `.semgrep/t0_t7_rules.yaml`, `.github/workflows/go-ci.yml` |
| Phase 3 iOS signals | `ConversationSignalService`, `EchoPhase3Tests` |
| Phase 3 UI | `TypingIndicatorView`, `ChatDetailViewModel` wiring in `MessagingScreens` |

Search recent commits: `git log --oneline -20 --grep=WO-`

## Phase doc header template

After SF sync, update top of `docs/phase-N-work-orders.md`:

```markdown
**Total Work Orders:** N
**Status Summary:** X Completed, Y In Progress, Z Ready, W Backlog, B Blocked
**Last synced with Software Factory:** YYYY-MM-DD
```

Do not rewrite entire WO bodies in markdown — SF descriptions are canonical.

## Batch edit example

Use `edit_work_orders` with multiple `{ work_order_id, status }` entries in one call.

Document changes in commit message or tell the user what moved.

## Related

- [`docs/AGENT_TOOLING_RECOMMENDATIONS.md`](../../docs/AGENT_TOOLING_RECOMMENDATIONS.md)
- [`AGENTS.md`](../../AGENTS.md)
