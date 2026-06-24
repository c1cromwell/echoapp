# WebRTC Xcode Setup (M4c)

Echo uses [stasel/WebRTC](https://github.com/stasel/WebRTC) for real `RTCPeerConnection` on iOS. Without this package, the app falls back to the **stub** engine (signaling-only, no real audio/video).

## Add the package (one-time, in Xcode)

1. Open `ios/Echo/EchoApp.xcodeproj`.
2. **File → Add Package Dependencies…**
3. URL: `https://github.com/stasel/WebRTC.git`
4. Version: **Exact** `131.0.0`
5. Add product **WebRTC** to the **EchoApp** app target (not test targets only).

## Verify live engine

Build and run on a physical device or simulator. In a voice call between two clients:

- `CallViewModel.usesLiveWebRTC` should be `true` (debugger breakpoint optional).
- You should hear audio when ICE connects (not just UI timer).

Headless `swift build` on macOS uses the stub engine (`#if canImport(WebRTC)` is false). CI stays green without linking WebRTC on macOS.

## Permissions

Ensure `Info.plist` includes:

- `NSMicrophoneUsageDescription`
- `NSCameraUsageDescription` (video calls)

## TURN (optional, NAT traversal)

On the backend, set:

```bash
ECHO_TURN_URL=turn:your-turn.example.com:3478
ECHO_TURN_USERNAME=echo
ECHO_TURN_CREDENTIAL=secret
```

`GET /v3/calls/ice-servers` will include the TURN entry; iOS fetches this before creating the peer connection.

**Local dev:** start optional coturn with `docker compose --profile webrtc up`, then set `ECHO_TURN_URL` to your Mac's LAN IP (simulators/devices cannot reach the Docker service name):

```bash
# In docker-compose.yml echoapp environment (uncomment and set your LAN IP):
ECHO_TURN_URL=turn:192.168.1.100:3478
ECHO_TURN_USERNAME=echo
ECHO_TURN_CREDENTIAL=echo_dev_turn
```

## E2E gate

Run §6.12 in [`E2E_MESSAGING_SIGNOFF_CHECKLIST.md`](../docs/E2E_MESSAGING_SIGNOFF_CHECKLIST.md).
