# Contributing to Echo

Welcome. This guide takes a new developer from "fresh laptop" to a passing
`make validate-phase1`. Keep it accurate — if something here is wrong, fix it in
the same PR; onboarding docs only stay correct when we treat them as code.

> **Audience:** engineers on the Phase-1 Echo stack — Go backend, Scala metagraph
> modules, the iOS app, and the Constellation Euclid testnet that ties them together.
>
> **Working with an AI agent / Cursor?** See [`AGENTS.md`](AGENTS.md) — architecture
> map, source-of-truth hierarchy, common commands, and agent constraints.
>
> **Shipping (TestFlight / end-to-end)?** This file is developer setup;
> [`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md) covers testing
> tiers and release.

## Table of contents

1. [Prerequisites](#1-prerequisites)
2. [One-time setup](#2-one-time-setup)
3. [First boot](#3-first-boot)
4. [Daily workflow](#4-daily-workflow)
5. [Repository layout](#5-repository-layout)
6. [Running tests](#6-running-tests)
7. [Submitting Scala L1 validation changes](#7-submitting-scala-l1-validation-changes)
8. [Common gotchas](#8-common-gotchas)
9. [Where to ask for help](#9-where-to-ask-for-help)

---

## 1. Prerequisites

| Tool             | Version           | Why                                                |
| ---------------- | ----------------- | -------------------------------------------------- |
| macOS or Linux   | macOS 14+ / Ubuntu 22.04 Server | Phase-1 is built and tested on both (Server image is lighter; no GUI needed). |
| JDK              | **21** (Temurin)  | Tessellation 4.0.0-rc.0 requires JDK 21.           |
| sbt              | 1.9+              | Builds the metagraph Scala modules.                |
| Scala            | 2.13.10           | Pinned by `metagraph/build.sbt`.                   |
| Go               | 1.21+             | Builds the backend.                                |
| Docker + Compose v2 | **8 GB+ RAM, 4+ CPUs** | Runs Postgres, Redis, NATS, MinIO, and the Euclid cluster. |
| Xcode            | 15+ (macOS only)  | iOS app (`ios/Echo/`).                             |
| `jq`, `yq`, `argc`, Ansible 2.16+, g8 (giter8) | latest | Required by `metagraph/scripts/setup-euclid.sh` and Euclid `hydra`. |

`metagraph/scripts/setup-euclid.sh` (run in step 2d) is the source of truth — it
checks every prerequisite and tells you exactly what's missing.

---

## 2. One-time setup

### 2a. macOS (Apple Silicon or Intel)

```bash
# Homebrew (skip if installed) — follow its PATH hint on Apple Silicon.
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

brew install --cask temurin@21          # JDK 21 first so it wins on PATH
brew install sbt coursier/formulas/coursier go jq yq argc ansible
cs install scala:2.13.10 scalafmt giter8
brew install --cask docker && open -a Docker   # launch once to create the VM
xcode-select --install
```

Then in Docker Desktop → **Settings → Resources**, set **Memory ≥ 8 GB** and
**CPUs ≥ 4** (Euclid warns, then fails, below this).

### 2b. Ubuntu 22.04+ / Debian (Server image recommended)

```bash
# Base packages
sudo apt update && sudo apt install -y curl git jq build-essential python3-pip
sudo pip3 install --user 'ansible>=2.16' && echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc

# JDK 21 + sbt + Scala via SDKMAN (handles x86_64 and ARM64; no GPG/apt-source dance)
curl -s "https://get.sdkman.io" | bash && source "$HOME/.sdkman/bin/sdkman-init.sh"
sdk install java 21.0.5-tem && sdk install sbt 1.9.9 && sdk install scala 2.13.10

# Docker Engine + Compose v2
curl -fsSL https://get.docker.com | sudo sh && sudo usermod -aG docker "$USER"   # re-login to apply

# Go (matches go.mod toolchain)
GO_VER=$(curl -fsSL "https://go.dev/VERSION?m=text" | head -1)
GO_ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -fsSL "https://go.dev/dl/${GO_VER}.linux-${GO_ARCH}.tar.gz" | sudo tar -C /usr/local -xz
echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
```

For the remaining tools, grab the binary for your arch from upstream releases:
**yq** (github.com/mikefarah/yq), **argc** (github.com/sigoden/argc), and **g8**
(`cs install giter8` once [Coursier](https://get-coursier.io) is installed).

### 2c. Verify the toolchain

```bash
java -version   # 21.x Temurin
sbt --version   # 1.9.x / 1.10.x
scala -version  # 2.13.10
go version       # 1.21+
docker version && docker compose version
```

If `java` resolves to an older JDK on macOS, pin it:
`echo 'export JAVA_HOME=$(/usr/libexec/java_home -v 21)' >> ~/.zshrc`

### 2d. Clone and bootstrap

```bash
git clone git@github.com:c1cromwell/echoapp.git && cd echoapp
cp .env.example .env          # edit if you need to override ports / DIDs
go mod download
cd metagraph && ./scripts/setup-euclid.sh && cd ..   # prereq check + Euclid clone (sibling dir)
```

`setup-euclid.sh` prints `✓ All prerequisites found` when step 2c is complete.

---

## 3. First boot

```bash
# 1. Build the metagraph fat JARs (~5–10 min first run; cached after).
#    `sbt assembly` is required — `sbt compile` alone does not produce runnable JARs.
cd metagraph && sbt assembly && cd ..

# 2. Bring up the Phase-1 cluster: Euclid metagraph + Postgres/Redis/NATS/MinIO + echoapp:8000.
make dev

# 3. (Optional) Identity L0/L1 for VC / trust-tier features — a SEPARATE metagraph
#    run from the assembly JARs (not hydra-managed; see docs/adr/0002…). Requires
#    step 2's cluster up: Identity L0 genesis-boots then peers with the Global L0.
make start-identity

# 4. Run the WO-230 go/no-go validation.
make validate-phase1
```

`validate-phase1` runs 7 steps: **0** prerequisites · **1** derive did:key ·
**2** register DID · **3** anchor a trust-tier commitment (Identity Metagraph) ·
**4** relay a test message · **5** commit a Merkle root to Data L1 + finality ·
**6** Global L0 height increments. Steps **3** and **5** report `skip` when the
Identity Metagraph / Data L1 read API aren't up — expected for backend-only work;
the run still prints **GO (with N pending steps)**.

Stop things when done:

```bash
make dev-stop                                              # backend stack (metagraph stays up)
make stop-identity                                         # Identity L0/L1 nodes
cd ../euclid-development-environment && scripts/hydra stop  # the metagraph itself
```

---

## 4. Daily workflow

```bash
make dev && make dev-status         # start of day (idempotent; reuses containers)

# Go backend
make run                            # foreground server with live env
make test                           # full Go suite
make lint && make vet && make fmt   # static checks + `go fmt`

# Scala metagraph (run from metagraph/)
sbt 'sharedData/test'               # fastest feedback (pure validators)
sbt 'identityL1/test'               # wired-validator integration
sbt test                            # everything · sbt scalafmtAll to format
# or from repo root: make metagraph-test

# iOS — open in Xcode, use the "Debug-Local" scheme (→ http://localhost:8000)
open ios/Echo/EchoApp.xcodeproj
```

Before opening a PR:

```bash
make fmt && make lint && make vet && make test
cd metagraph && sbt scalafmtCheckAll && sbt test && cd ..
make validate-phase1                # go/no-go must still pass
```

`make release-check` is the full pre-tag gate (build + race tests + vet + gofmt).

---

## 5. Repository layout

```
echoapp/
├── cmd/        # Go entry points (server, didkey CLI)
├── internal/   # Go backend: api/ database/ services/ auth/ metagraph/ infra/ logging/ …
├── pkg/        # Reusable Go: didkey/ did/ credentials/ (+ credentials/oidc4vc) …
├── metagraph/  # Scala / Tessellation 4.0.0-rc.0 (shared_data + 5 L0/L1 layer modules)
├── ios/Echo/   # iOS app
├── docs/       # PRD, ADRs, phase work orders, audits, launch & testing guides
├── scripts/    # validate-phase1.sh and friends
└── Makefile    # `make help` for the full menu
```

For the architecture map, per-module responsibilities, and the source-of-truth
hierarchy, see [`AGENTS.md`](AGENTS.md).

---

## 6. Running tests

| Surface                | Command                                                          |
| ---------------------- | ---------------------------------------------------------------- |
| Go (unit + integration)| `make test`                                                      |
| Go endpoint smoke      | `make test-endpoints` (against a running `make run`)             |
| Scala pure validators  | `cd metagraph && sbt 'sharedData/test'`                          |
| Scala wired validators | `cd metagraph && sbt 'identityL1/test' 'l1/test' 'data_l1/test'` |
| Phase-1 go/no-go       | `make validate-phase1` (needs `make dev` up)                     |
| iOS                    | Xcode → `Cmd+U`, "Debug-Local" scheme                            |

Full test matrix, regression tiers, and CI gates:
[`docs/TESTING.md`](docs/TESTING.md) and
[`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md).

---

## 7. Submitting Scala L1 validation changes

The Scala L1 modules keep rules unit-testable with a layered pattern:

1. **Add the rule** as a pure function in
   `metagraph/modules/shared_data/.../validations/` — returns `Either[String, Unit]`,
   takes all dependencies as parameters (no I/O or clocks; inject `now: Long`).
2. **Wire it** into the relevant `Main.dispatch` (`identity_l1`, `l1`, or `data_l1`).
   `Main.dispatch` is just a router — no rule logic there.
3. **Add a pure-function test** covering happy path + every failure branch.
4. **Add a wired-validator test** in the L1 module's `MainSpec.scala` (WO-277 #3:
   every wired validator needs at least one).
5. `cd metagraph && sbt scalafmtAll && sbt test` must be green before pushing.

If you add a new on-chain field or `IdentityUpdate`/`EchoUpdate` **variant**, also
update its `Encoder` in `IdentityTypes.scala` / `Types.scala` **and every
`Main.dispatch`** — the build runs under **`-Werror`**, so a non-exhaustive `match`
fails compilation. Run `sbt compile` (all modules), not just `sbt 'sharedData/test'`.
Update the calculated-state combiner if the field should be queryable, and `docs/adr/`
if the change is architecturally meaningful.

---

## 8. Common gotchas

- **`sbt`/`go` not found after install.** Open a fresh terminal so the new PATH loads.
- **`docker: Cannot connect to the Docker daemon`.** Docker Desktop isn't running — launch it and wait for the whale icon to settle.
- **`sbt assembly` slow on first run.** Expected — it pulls the Tessellation SDK; later runs hit the Coursier cache and finish in seconds.
- **JVM picks an old JDK.** Force 21: `export JAVA_HOME=$(/usr/libexec/java_home -v 21)` (macOS).
- **Port already in use.** Euclid/Identity use the 9000–9602 range. Find squatters: `lsof -nP -iTCP -sTCP:LISTEN | awk '$9 ~ /:(9[0-6][0-9][0-9])$/'`.
- **Identity L0/L1 unreachable.** Identity is a **separate** metagraph — `make dev` does **not** start it (it's not hydra-managed; see [ADR-0002](docs/adr/0002-identity-metagraph-deployment.md)). Run `make start-identity` after the core cluster is up (it preflights the JARs, `metagraph-base-image`, and Global L0). L0 genesis takes ~30s to reach `Ready`. Logs via `docker logs identity-l0` / `docker logs identity-l1`.
- **`IDENTITY_SERVICE_DID` not set.** Identity L1 rejects every submission (authorized-sender check). Generate one: `go run ./cmd/didkey -in /path/to/key.pem`, then set it in `.env`.
- **Env vars.** Required values live in `.env.example`. Notable: `IDENTITY_SERVICE_DID`, `JWT_SIGNING_KEY` (**required in production**), `CONTACT_OPRF_KEY` (contact discovery), `LOG_MASTER_KEY` (audit log). Generation steps are in `docs/E2E_LAUNCH_AND_TESTING.md §3`.

---

## 9. Where to ask for help

- **Architecture / blueprint.** Read `AGENTS.md`, `docs/PRD.md`, and `docs/adr/` first; then open a discussion or ping the blueprint owner in `software-factory-echo`.
- **Tessellation SDK.** Constellation docs (<https://docs.constellationnetwork.io/sdk>) and the [Euclid repo](https://github.com/Constellation-Labs/euclid-development-environment).
- **Stuck on a Phase-1 work order.** Check the work order's hand-off notes in `software-factory-echo` before pinging the assignee.
