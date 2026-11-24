# Vendored Tessellation

Echo runs its own Tessellation network (see `docs/OWN_L0_PLAN.md`). This directory makes the chain
buildable with no access to Constellation's GitHub.

| File | Purpose |
|---|---|
| `SHA256SUMS-<version>` | Committed checksums of the node binaries Echo trusts for that version |
| `fetch.sh` | Downloads the binaries from Echo's mirrors and verifies them; called by `scripts/setup-euclid.sh` |
| `build-from-source.sh` | Rebuilds the binaries from Echo's Tessellation fork when adopting a version, or if every mirror is gone |
| `tessellation/<version>/` | The binaries themselves (not in git) |

Binaries: `cl-node.jar` (Global L0), `cl-dag-l1.jar`, `cl-wallet.jar`, `cl-keytool.jar`, and
`tessellation-sdk-maven.tar.gz` — a Maven-layout copy of the `io.constellationnetwork` artefacts the
metagraph compiles against. The build normally resolves those from Maven Central; to build without
it, unpack the archive into `~/.m2/repository` (`build.sbt` already lists `Resolver.mavenLocal`).

Mirrors, in order: release `echo-<version>` on Echo's Tessellation fork, then any extra base URLs in
`ECHO_VENDOR_MIRRORS` (space-separated), for example the object-storage copy.

`make chain-supply-check` fails if a Constellation-hosted URL returns to the build path or the
binaries stop matching the checksums.
