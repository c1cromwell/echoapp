# Backend and Metagraph end-to-end testing

This guide walks through verifying the Echo **Go backend** together with the **Constellation / Euclid metagraph** stack used in Phase 1 (WO-230, ADR-0001). It complements `scripts/validate-phase1.sh` with manual steps and URLs.

## Prerequisites

1. **Docker** (for Postgres, Redis, NATS, MinIO, and the `echoapp` API container—or run the API on the host with the same env vars).
2. **Euclid SDK** sibling repo — cloned by `metagraph/scripts/setup-euclid.sh` (see `metagraph/euclid.json` for Git ref).
3. **Host tools:** `curl`, `jq`, `openssl`, `go`, optional `make`, optional `sbt` + JDK 21 for Scala tests.
4. **Network:** metagraph nodes listen on localhost (or `host.docker.internal` from Docker). Ports are fixed in `metagraph/euclid.json`.

## 1. One-time SDK setup

```bash
cd metagraph && ./scripts/setup-euclid.sh
```

Note the printed `EUCLID_DIR` (usually `../euclid-development-environment` next to `echoapp`).

## 2. Start the metagraph cluster (Hydra)

From **Euclid**:

```bash
cd "$EUCLID_DIR"   # e.g. ../euclid-development-environment
scripts/hydra build          # first time / after SDK changes
scripts/hydra start-genesis
scripts/hydra status
```

Wait until **Global L0**, **Metagraph L0**, **Currency L1**, **Data L1**, **Identity L0**, and **Identity L1** are up.

### Quick health checks (all should return JSON or HTTP 200)

| Layer        | URL |
|-------------|-----|
| Global L0   | `http://localhost:9000/node/info` |
| Metagraph L0 | `http://localhost:9200/node/info` |
| Currency L1 | `http://localhost:9300/node/info` |
| Data L1     | `http://localhost:9400/node/info` |
| Identity L0 | `http://localhost:9100/node/info` |
| Identity L1 | `http://localhost:9500/node/info` |

```bash
for u in \
  http://localhost:9000/node/info \
  http://localhost:9200/node/info \
  http://localhost:9300/node/info \
  http://localhost:9400/node/info \
  http://localhost:9100/node/info \
  http://localhost:9500/node/info
do
  echo "== $u"; curl -fsS --max-time 5 "$u" | head -c 200; echo; echo
done
```

## 3. Start the backend stack

From the **echoapp** repo root:

```bash
make dev
```

This runs `metagraph-verify-skeleton` (static checks), starts/reuses Hydra, then `docker compose -f docker-compose.testnet.yml up -d --build`.

Confirm API health:

```bash
curl -fsS http://localhost:8000/health
```

The compose file sets metagraph URLs via `host.docker.internal` (`docker-compose.testnet.yml`). On Linux you may need extra host mapping; on Docker Desktop (macOS/Windows) this works out of the box.

### Environment variables (reference)

| Variable | Typical value (host metagraph) |
|----------|---------------------------------|
| `GLOBAL_L0_URL` | `http://localhost:9000` |
| `METAGRAPH_L0_URL` | `http://localhost:9200` |
| `CURRENCY_L1_URL` | `http://localhost:9300` |
| `DATA_L1_URL` | `http://localhost:9400` |
| `IDENTITY_L0_URL` | `http://localhost:9100` |
| `IDENTITY_L1_URL` | `http://localhost:9500` |
| `IDENTITY_SERVICE_DID` | Issuer `did:key:z…` (must match Identity L1 `IDENTITY_SERVICE_DID` for validates submissions) |

## 4. Automated Phase-1 script (WO-230)

```bash
make validate-phase1
# or
./scripts/validate-phase1.sh
```

What it covers:

- **Step 1–2:** Local `did:key` derivation (`cmd/didkey`) and `POST /identity/register`, `GET /identity/{did}` on the backend.
- **Step 3:** Identity L0/L1 reachability; VC anchor assertion is tied to WO-274 / issuer wiring when enabled.
- **Step 4:** Relay E2E (`go test` integration).
- **Step 5–6:** Data L1 Merkle submission (when routed) and Global L0 ordinal advancement.

Env overrides: `BACKEND_URL`, `GLOBAL_L0_URL`, `IDENTITY_L1_URL`, `FINALITY_TIMEOUT_SECS`, etc. (see script header).

## 5. Identity HTTP API (backend)

### Register primary device (did:key)

```bash
# Produce PEM then derive (or use go run ./cmd/didkey -pem pub.pem)
openssl ecparam -name prime256v1 -genkey -noout -out key.pem
openssl ec -in key.pem -pubout -out pub.pem
eval "$(go run ./cmd/didkey -pem pub.pem | awk -F= '/^did=/{print \"export DID=\"$2} /^public_key_hex=/{print \"export PUB=\"$2}')"

curl -fsS -X POST "http://localhost:8000/identity/register" \
  -H "Content-Type: application/json" \
  -d "{\"did\":\"$DID\",\"public_key_hex\":\"$PUB\"}" | jq .
```

### Resolve all device keys

```bash
curl -fsS "http://localhost:8000/identity/$DID" | jq .
curl -fsS "http://localhost:8000/identity/resolve/$DID" | jq .
```

### Add a second device (signed request)

The body must be signed with a key already registered for `subject_did` (`X-Identity-Signature`, ECDSA over SHA-256 of raw JSON body). See `openapi.yaml` → `/identity/devices` and `internal/api/identity_register_handlers.go`.

## 6. Identity Metagraph L1 — JSON transaction shapes (dev)

The Go client posts to **`POST {IDENTITY_L1_URL}/transactions`**. Bodies are flat JSON objects that map to Scala `IdentityUpdate` variants (`metagraph/modules/shared_data/types/IdentityTypes.scala`).

### 6.1 VC issuance metadata (`VCIssuanceUpdate`)

Fields: `credentialId`, `subjectDID`, `issuerDID`, `credentialType`, `issuedAt` (epoch **millis**), `schemaVersion`.

Example (issuer DID must match `IDENTITY_SERVICE_DID` on the L1 node):

```bash
NOW_MS=$(($(date +%s) * 1000))
curl -sS -X POST "http://localhost:9500/transactions" \
  -H "Content-Type: application/json" \
  -d "{
    \"credentialId\": \"urn:uuid:$(uuidgen | tr '[:upper:]' '[:lower:]')\",
    \"subjectDID\": \"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK\",
    \"issuerDID\": \"did:key:z6MkjvBkt8ETnxXGBFPSGgYKb43q7oNHLX8BiYSPcXVG6gY6\",
    \"credentialType\": \"TrustTierCredential\",
    \"issuedAt\": $NOW_MS,
    \"schemaVersion\": \"v1.0.0\"
  }"
```

**Note:** Tessellation may wrap this in an outer transaction envelope depending on hydra version; if `404`/`400`, check Euclid docs for the exact path.

### 6.2 Device key registration (`DeviceKeyRegistrationUpdate`)

Fields: `subjectDID`, `publicKeyHex`, `deviceLabel`, `addedAt` (epoch millis). See `internal/metagraph/identity_l1.go` and Scala `DeviceKeyRegistrationUpdate`.

### 6.3 StatusList2021, trust tier, org role

See `IdentityValidations.scala` and WO-272 descriptions for field shapes (`bitVector` length, `commitment` hex, etc.).

## 7. Data L1 — Merkle / governance-style payloads

The backend’s `internal/metagraph` client uses `POST {DATA_L1_URL}/transactions`. `scripts/validate-phase1.sh` step 5 attempts `POST /v1/data-l1/merkle-roots` on the API when implemented.

For raw experimentation, mirror the JSON shapes used in `internal/governance/service.go` (`SubmitDataL1`).

## 8. Currency L1 — token operations

Use `internal/metagraph/client.go` (`SubmitCurrencyL1`) and reward/batch code as references. Public base: `http://localhost:9300`.

## 9. Scala tests (validators, codecs)

From `metagraph/`:

```bash
sbt "sharedData/test" "identityL0/test" "identityL1/test"
```

CI-friendly static check (no cluster):

```bash
make metagraph-verify-skeleton
```

## 10. Credentials service (WO-274) — VC issuance + Identity L1 anchor

The `pkg/credentials` issuer can emit **W3C VC 2.0**-shaped JSON-LD (when `UseW3CVC2` is enabled) and publish **VC issuance metadata** to Identity L1 when `EnableAnchor` and `IDENTITY_L1_URL` are set. Run the standalone credentials service (`cmd/credentials`) with valid issuer key material and env from `pkg/credentials/config.go`.

## 11. Tear down

- Backend: `make testnet-down` or `make dev-stop`
- Metagraph: `cd "$EUCLID_DIR" && scripts/hydra stop`

## Troubleshooting

| Symptom | Check |
|--------|--------|
| Identity L1 rejects submissions | `IDENTITY_SERVICE_DID` on L1 matches `issuerDID` / `ISSUER_DID` in Go; see `identity_l1/Main.scala`. |
| Docker cannot reach metagraph | `extra_hosts: host.docker.internal:host-gateway` in compose; URLs use `host.docker.internal`. |
| Step 3 skipped in validate-phase1 | Identity nodes not running or VC/anchor path not yet asserted by script. |
| `hydra` not found | Run `setup-euclid.sh` and use the **absolute** path to `scripts/hydra` inside Euclid. |
