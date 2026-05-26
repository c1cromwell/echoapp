# Echo OPRF (WO-221)

Go package exposing RFC 9497 OPRF client operations compatible with `internal/services/contacts/oprf.go`.

## iOS integration

Build an XCFramework (requires full Xcode, not Command Line Tools only):

```bash
./scripts/build-echooprf-ios.sh
```

Or manually:

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
gomobile bind -target=ios -o ios/Echo/Libraries/EchoOPRF.xcframework ./mobile/echooprf
```

Add `ios/Echo/Libraries/EchoOPRF.xcframework` to EchoApp (Embed & Sign). `OPRFClientFactory.makeDefault()` selects `LiveOPRFClient` when the `Echooprf` module is linked.

## Verify

```bash
go test ./mobile/echooprf/...
```
