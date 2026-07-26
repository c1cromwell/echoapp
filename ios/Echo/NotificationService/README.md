# Echo Notification Service Extension (Wave S4)

Signal ships `SignalNSE` to decrypt push payloads and show rich previews.
Echo’s equivalent is this extension target (add to `EchoApp.xcodeproj` in Xcode).

## Status

- Skeleton: [`NotificationService.swift`](NotificationService.swift)
- **Xcode required:** create an App extension target `EchoNotificationService`, set App Group if sharing Keychain, enable Push Notifications
- Decrypt path must reuse Echo Keychain + Kinnami / Double Ratchet session stores — do not log plaintext

## Acceptance

1. Background push arrives while app suspended  
2. NSE decrypts envelope (when keys available)  
3. Notification shows contact name + short preview (or “New message” if locked / missing keys)  
4. No plaintext in extension logs  

## Related

[`docs/SIGNAL_ECHO_PARITY.md`](../../../docs/SIGNAL_ECHO_PARITY.md) Wave S4 · Signal-iOS `SignalNSE/`
