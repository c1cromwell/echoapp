# Besu spike

Throwaway code that tested whether Hyperledger Besu can replace Tessellation as Echo's own chain.
Results and the plan that follows from them: [`docs/BESU_MIGRATION_PLAN.md`](../../docs/BESU_MIGRATION_PLAN.md).
Keys made here are for this spike only.

```bash
./setup.sh                 # four validator keys + genesis into .network/, compiles the contract
docker compose up -d       # four validators, JSON-RPC on localhost:8545-8548
(cd cmd/echobesu && GOWORK=off go build -o ../../build/echobesu .)
./build/echobesu deploy    # contract + the gateway's signing key as its only writer
./build/echobesu anchor    # anchor a random Merkle root
./build/echobesu reject    # a root from an unlisted key must be refused
./build/echobesu prove <root> build/proof.json
./build/echobesu verify build/proof.json "$(cat .network/node*/address | paste -sd, -)"   # no network
./failure-tests.sh         # one down, below quorum and back, wiped node, rolling restarts
docker compose down -v     # remove everything
```

The Go client is its own module so that `go-ethereum` does not enter the main `go.mod` before the
migration is approved stage by stage.
