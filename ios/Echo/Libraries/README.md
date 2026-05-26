# Echo iOS native libraries

## EchoOPRF.xcframework (WO-221)

Generated locally — not committed. From repo root:

```bash
./scripts/build-echooprf-ios.sh
```

Then add `EchoOPRF.xcframework` to the EchoApp target (Frameworks, Libraries, and Embedded Content → Embed & Sign). `OPRFClientFactory` uses the live client when the `Echooprf` module is linked; otherwise it falls back to `MockOPRFClient` for simulator/dev builds without the framework.
