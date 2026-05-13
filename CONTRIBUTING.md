# Contributing to Echo

Welcome. This guide gets a new developer from "fresh laptop" to "passing
`make validate-phase1`" with the minimum amount of yak shaving. If
anything here is wrong or missing, please update it as part of the same
PR — onboarding docs only stay accurate when we treat them as code.

> **Audience:** engineers working on the Phase-1 Echo stack — Go backend,
> Scala metagraph modules, iOS app prototype, and the Constellation
> Euclid SDK testnet that ties them together.

> **Targeting a TestFlight launch?** See [`docs/PHASE1_LAUNCH.md`](docs/PHASE1_LAUNCH.md)
> for the end-to-end test checklist, code signing walkthrough, and the
> June 1 countdown. This file covers developer setup; that file covers
> shipping.

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
| macOS or Linux   | recent            | Phase-1 is built and tested on macOS 14+ and **Ubuntu 22.04 Server** (no GUI needed; Server image is lighter). |
| Homebrew         | latest            | Package manager (macOS install path below).        |
| JDK              | **21** (Temurin)  | Tessellation 4.0.0-rc.0 requires JDK 21.           |
| sbt              | 1.9+              | Builds the metagraph Scala modules.                |
| Scala            | 2.13.10           | Pinned by `metagraph/build.sbt`.                   |
| Coursier (`cs`)  | latest            | Installs Scala + scalafmt cleanly.                 |
| Docker Desktop   | latest, **8 GB+ RAM**, **4+ CPUs** | Runs Postgres, Redis, NATS, MinIO, and the Euclid metagraph cluster. |
| Docker Compose   | v2 (`docker compose`) | Used by `docker-compose.testnet.yml`.          |
| Go               | 1.21+             | Builds the backend.                                |
| Xcode            | 15+ (macOS only)  | iOS prototype (`ios/Echo/`).                       |
| `jq`, `yq`, `argc` | latest          | Required by `metagraph/scripts/setup-euclid.sh`. `argc` is in Homebrew core — `brew install argc`. |
| Git LFS          | optional          | Needed only if you touch large binary fixtures.    |

---

## 2. One-time setup

### 2a. macOS (Apple Silicon or Intel)

```bash
# Homebrew (skip if already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
# Apple Silicon: follow the post-install hint to add brew to PATH, e.g.
# echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc

# JDK 21 first so it wins on PATH
brew install --cask temurin@21

# sbt + Scala + scalafmt via Coursier
brew install sbt coursier/formulas/coursier
cs install scala:2.13.10 scalafmt

# Euclid prerequisites
brew install jq yq argc              # argc moved to Homebrew core; no tap needed

# Docker Desktop (then launch it once so the VM is created)
brew install --cask docker
open -a Docker

# Go (matches go.mod toolchain)
brew install go

# Xcode from the App Store, plus command-line tools:
xcode-select --install
```

In Docker Desktop → **Settings → Resources** bump **Memory ≥ 8 GB** and
**CPUs ≥ 4** before continuing. Euclid will warn (and eventually fail)
below this.

### 2b. Ubuntu 22.04+ / Debian (Server image recommended)

Use **Ubuntu Server** (not Desktop) — no GUI is needed and Server is
significantly lighter (2 GB image vs. 4+ GB Desktop). Download from
[ubuntu.com/download/server](https://ubuntu.com/download/server).
The commands below work identically on both images and on Debian 12.

```bash
sudo apt update && sudo apt install -y curl gnupg jq yq git build-essential

# JDK 21 (Temurin)
curl -fsSL https://packages.adoptium.net/artifactory/api/gpg/key/public \
  | sudo gpg --dearmor -o /usr/share/keyrings/adoptium.gpg
echo "deb [signed-by=/usr/share/keyrings/adoptium.gpg] https://packages.adoptium.net/artifactory/deb $(lsb_release -cs) main" \
  | sudo tee /etc/apt/sources.list.d/adoptium.list
sudo apt update && sudo apt install -y temurin-21-jdk

# Coursier + Scala + sbt + scalafmt  (no apt/GPG needed, works on x86_64 + ARM64)
CS_ARCH=$(uname -m | sed 's/aarch64/aarch64/;s/arm64/aarch64/;s/x86_64/x86_64/')
curl -fL "https://github.com/coursier/coursier/releases/latest/download/cs-${CS_ARCH}-pc-linux.gz" \
  | gunzip > /tmp/cs && chmod +x /tmp/cs && sudo mv /tmp/cs /usr/local/bin/cs
cs install sbt scala:2.13.10 scalafmt
echo 'export PATH="$PATH:$HOME/.local/share/coursier/bin"' >> ~/.bashrc
source ~/.bashrc

# argc
cargo install argc   # requires Rust toolchain; or download release binary

# Docker Engine + Compose v2
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"   # log out / back in for this to take effect

# Go (matches go.mod toolchain — pin to 1.21+)
curl -fsSL https://go.dev/dl/go1.22.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
echo 'export PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
```

### 2c. Verify the toolchain

Run each of these and confirm the expected output:

```bash
java -version          # openjdk version "21.x.x" ... Temurin
javac -version         # javac 21.x.x
sbt --version          # sbt script version: 1.9.x or 1.10.x
scala -version         # 2.13.10
go version             # go1.21+ (1.22 recommended)
docker version         # client + server both report
docker compose version # v2.x.y
jq --version
yq --version
argc --version
```

If `java -version` reports anything older than 21, set `JAVA_HOME`
explicitly (macOS):

```bash
echo 'export JAVA_HOME=$(/usr/libexec/java_home -v 21)' >> ~/.zshrc
echo 'export PATH="$JAVA_HOME/bin:$PATH"' >> ~/.zshrc
```

### 2d. Clone and bootstrap the repo

```bash
git clone git@github.com:thechadcromwell/echoapp.git
cd echoapp

# Local environment file
cp .env.example .env
# Edit .env if you need to override ports, identity service DID, etc.

# Pull Go dependencies
go mod download

# One-shot Euclid prereq check + clone of euclid-development-environment
# into ../euclid-development-environment (sibling to this repo).
cd metagraph && ./scripts/setup-euclid.sh && cd ..
```

`setup-euclid.sh` will print `✓ All prerequisites found` if step 2c
worked. If something is missing, fix it and re-run.

---

## 3. First boot

```bash
# 1. Compile + test the metagraph modules (~5 min first run; jars cached after)
cd metagraph && sbt compile && sbt test && cd ..

# 2. Bring up the full Phase-1 cluster
#    - Euclid: Global L0 (9000), Metagraph L0 (9200), Currency L1 (9300),
#              Data L1 (9400), Identity L0 (9100), Identity L1 (9500)
#    - Backend: Postgres, Redis, NATS, MinIO, echoapp (8000)
make dev

# 3. Run the WO-230 6-step go/no-go validation
make validate-phase1
```

Expected: the script reports `ok` for steps 0, 1, 2, 3 (Identity L0/L1
reachability), 4, and 6, and `skip` for steps 3 (VC issuance assertion —
WO-273) and 5 (Data L1 anchor — separate WO).

Other useful targets (`make help` shows the full list):

```bash
make dev-status       # show status of all testnet components
make dev-logs         # tail backend stack logs
make dev-restart      # restart backend only (metagraph keeps running)
make dev-stop         # tear down backend stack
```

To stop the metagraph itself:

```bash
cd ../euclid-development-environment && scripts/hydra stop
```

---

## 4. Daily workflow

```bash
# Start of day
make dev                    # idempotent — re-uses running containers
make dev-status             # sanity check

# Iterate on Go backend
make run                    # foreground server with live env vars
make test                   # full Go test suite
make lint && make vet       # static checks
make fmt                    # gofmt + goimports

# Iterate on Scala metagraph
cd metagraph
sbt 'sharedData/test'              # fastest feedback (pure validators)
sbt 'identityL1/test'              # wired-validator integration spec
sbt test                            # everything
sbt scalafmtAll                     # formatter
# Or from repo root:  make metagraph-test   # sharedData + identityL0 + identityL1 (WO-272/277)

# Iterate on iOS
open ios/Echo/EchoApp.xcodeproj
# Use the "Debug-Local" scheme — points to http://localhost:8000

# End of day
make dev-stop                       # backend down (metagraph stays up)
# OR fully clean:
cd ../euclid-development-environment && scripts/hydra stop
```

Before opening a PR:

```bash
make fmt && make lint && make vet && make test
cd metagraph && sbt scalafmtCheckAll && sbt test && cd ..
make validate-phase1         # full go/no-go must still pass
```

---

## 5. Repository layout

```
echoapp/
├── cmd/                   # Go entry points (server, didkey CLI, etc.)
├── internal/              # Go backend internals
│   ├── api/               # HTTP routers + handlers (incl. /identity/register)
│   ├── database/          # Postgres + in-memory adapters
│   └── ...
├── pkg/                   # Reusable Go packages
│   ├── didkey/            # Canonical W3C did:key derivation (P-256)
│   ├── did/               # DID HTTP glue + errors (Phase 1: did:key; see ADR-0001)
│   └── identity/          # Identity service (legacy gin handlers)
├── metagraph/             # Scala / Tessellation 4.0.0-rc.0 modules
│   ├── build.sbt          # 6 sub-projects (sharedData + 5 layers)
│   ├── euclid.json        # Container topology + port allocation
│   ├── modules/
│   │   ├── shared_data/   # Pure validators, types, ClusterIds
│   │   ├── l0/            # Currency Metagraph L0
│   │   ├── l1/            # Currency L1 (token ops)
│   │   ├── data_l1/       # Data L1 (Merkle roots, trust commitments)
│   │   ├── identity_l0/   # Identity Metagraph L0 (consensus)
│   │   └── identity_l1/   # Identity Metagraph L1 (VC, StatusList2021, OrgRole)
│   └── scripts/setup-euclid.sh
├── ios/Echo/              # iOS app prototype
├── docs/
│   ├── PRD.md             # Product requirements
│   ├── adr/               # Architecture Decision Records
│   └── api/               # API reference
├── scripts/
│   ├── validate-phase1.sh # 6-step Phase-1 go/no-go
│   └── ...
├── docker-compose.yml         # Local Postgres/Redis/etc. (no metagraph)
├── docker-compose.testnet.yml # Backend stack pointing at host metagraph
├── Makefile               # `make help` for the full menu
└── .env.example           # Copy to .env
```

---

## 6. Running tests

| Surface                | Command                                              | Notes                                  |
| ---------------------- | ---------------------------------------------------- | -------------------------------------- |
| Go unit + integration  | `make test`                                          | Uses `go test ./... -count=1`.         |
| Go endpoint smoke      | `make test-endpoints`                                | Hits a running `make run` instance.    |
| Go did:key derivation  | `go test ./pkg/didkey/...`                           | 11 tests, < 1s.                        |
| Identity register API  | `go test ./internal/api/...`                         | 9 handler tests, < 1s.                 |
| Scala pure validators  | `cd metagraph && sbt 'sharedData/test'`              | ~70 ScalaTest assertions, < 30s after warm-up. |
| Scala wired validators | `cd metagraph && sbt 'identityL1/test' 'l1/test' 'data_l1/test'` | Demonstrates `Main.dispatch` wiring.   |
| Phase-1 go/no-go       | `make validate-phase1`                               | Requires `make dev` to be running.     |
| iOS                    | Open in Xcode → `Cmd+U` against the iOS simulator    | "Debug-Local" scheme points at localhost backend. |

---

## 7. Submitting Scala L1 validation changes

The Scala L1 modules use a layered pattern that keeps the rules unit-testable:

1. **Add the rule** to a pure function in
   `metagraph/modules/shared_data/src/main/scala/com/echo/shared_data/validations/`.
   Functions return `Either[String, Unit]` and take all dependencies as
   parameters (no I/O, no clocks — inject `now: Long` if you need time).
2. **Wire it** into the relevant `Main.dispatch` (in `identity_l1`,
   `l1`, or `data_l1`). Don't put rule logic in `Main.dispatch` itself —
   it's just a router.
3. **Add a pure-function test** in
   `shared_data/src/test/scala/.../validations/` covering happy path +
   every failure branch.
4. **Add a wired-validator test** in the L1 module's `MainSpec.scala`
   showing the rule fires through `Main.dispatch` end-to-end. Per
   WO-277 acceptance criterion #3, every wired validator must have at
   least one such test.
5. Run `cd metagraph && sbt scalafmtAll && sbt test` and make sure it's
   green before pushing.

If you're adding a new on-chain field, also update:

- `IdentityTypes.scala` (or `Types.scala` for currency/data)
- The corresponding `IdentityCalculatedState` / `EchoCalculatedState`
  combiner if it should be queryable.
- `docs/adr/` if the change is architecturally meaningful.

---

## 8. Common gotchas

- **`sbt: command not found` after install.** Open a fresh terminal so
  the new PATH is picked up.
- **`docker: Cannot connect to the Docker daemon`.** Docker Desktop
  isn't running yet — launch it from Spotlight and wait for the whale
  icon to stop animating.
- **`hydra install` fails.** Usually a wrong tag on the Euclid clone.
  `git -C ../euclid-development-environment checkout v0.19.0` fixes it.
- **Port already in use.** Anything sitting on 9000-9002, 9100-9102,
  9200-9202, 9300-9302, 9400-9402, 9500-9502 will conflict. Find
  squatters with:
  ```bash
  lsof -nP -iTCP -sTCP:LISTEN | awk '$9 ~ /:(9[0-5][0-9][0-9])$/'
  ```
- **`sbt compile` is super slow on the first run.** Expected — it pulls
  the Tessellation 4.0.0-rc.0 SDK from `mavenLocal` + Constellation's
  repo. Subsequent runs hit `~/.cache/coursier/` and finish in seconds.
- **JVM picks an old JDK.** macOS sometimes resolves `java` to
  whatever's first on PATH. Force JDK 21 with
  `export JAVA_HOME=$(/usr/libexec/java_home -v 21)`.
- **Apple Silicon + Tessellation jars.** They're JVM bytecode so the
  chip doesn't matter, but Docker images may pull `linux/amd64` and run
  under Rosetta. That's fine for dev; expect ~2× CPU.
- **`make dev` reports backend healthy but Identity L0/L1 unreachable.**
  Hydra is still starting the JVMs — give it 30-60s, then re-run
  `make dev-status`. If it stays down, `cd ../euclid-development-environment
  && scripts/hydra logs identity-l0`.
- **`IDENTITY_SERVICE_DID` not set.** Identity L1 will reject every
  submission because the authorized-sender check fails. Generate a DID
  with `go run ./cmd/didkey -in /path/to/key.pem` and put it in `.env`.

---

## 9. Where to ask for help

- **Architecture / blueprint questions.** Read `docs/PRD.md` and
  `docs/adr/` first; if the answer isn't there, open a discussion or
  ping the relevant blueprint owner in `software-factory-echo`.
- **Tessellation SDK questions.** Constellation's docs:
  <https://docs.constellationnetwork.io/sdk> — and the Euclid repo
  itself: <https://github.com/Constellation-Labs/euclid-development-environment>.
- **Stuck on a Phase-1 work order.** The work order in
  `software-factory-echo` has hand-off notes; check those before pinging
  the WO assignee.
