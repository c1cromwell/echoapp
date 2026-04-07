#if canImport(UIKit)
import UIKit
#endif
import LocalAuthentication
import CryptoKit

// MARK: - Device Info

struct AuthDeviceInfo: Codable {
    let deviceId: String
    let platform: String
    let osVersion: String
    let appVersion: String
    let model: String
    let locale: String
    let timezone: String
    let secureEnclave: Bool
    let biometricType: String
    let jailbreakStatus: Bool

    enum CodingKeys: String, CodingKey {
        case deviceId = "device_id"
        case platform
        case osVersion = "os_version"
        case appVersion = "app_version"
        case model
        case locale
        case timezone
        case secureEnclave = "secure_enclave"
        case biometricType = "biometric_type"
        case jailbreakStatus = "jailbreak_status"
    }
}

// MARK: - Device Fingerprint Service

final class DeviceFingerprintService {

    func collectDeviceInfo() -> AuthDeviceInfo {
        #if os(iOS)
        let device = UIDevice.current
        #endif
        let context = LAContext()
        _ = context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        )

        #if os(iOS)
        let osVersion = device.systemVersion
        #else
        let osVersion = "unknown"
        #endif

        return AuthDeviceInfo(
            deviceId: computeDeviceId(),
            platform: "ios",
            osVersion: osVersion,
            appVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"]
                as? String ?? "unknown",
            model: deviceModelIdentifier(),
            locale: Locale.current.identifier,
            timezone: TimeZone.current.identifier,
            secureEnclave: hasSecureEnclave(),
            biometricType: biometricTypeString(context.biometryType),
            jailbreakStatus: isJailbroken()
        )
    }

    /// JSON-encoded device info for the X-Device-Info header
    func deviceInfoHeader() -> String {
        let info = collectDeviceInfo()
        guard let data = try? JSONEncoder().encode(info),
              let string = String(data: data, encoding: .utf8)
        else { return "{}" }
        return string
    }

    // MARK: - Private

    private func computeDeviceId() -> String {
        #if os(iOS)
        let vendorId = UIDevice.current.identifierForVendor?.uuidString ?? "unknown"
        #else
        let vendorId = "unknown"
        #endif
        let model = deviceModelIdentifier()
        let data = "\(vendorId):\(model)".data(using: .utf8)!
        let hash = SHA256.hash(data: data)
        return hash.compactMap { String(format: "%02x", $0) }.joined()
    }

    private func deviceModelIdentifier() -> String {
        var systemInfo = utsname()
        uname(&systemInfo)
        return withUnsafePointer(to: &systemInfo.machine) {
            $0.withMemoryRebound(to: CChar.self, capacity: 1) {
                String(cString: $0)
            }
        }
    }

    private func hasSecureEnclave() -> Bool {
        let context = LAContext()
        return context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        )
    }

    private func biometricTypeString(_ type: LABiometryType) -> String {
        switch type {
        case .faceID: return "face_id"
        case .touchID: return "touch_id"
        case .opticID: return "optic_id"
        default: return "none"
        }
    }

    private func isJailbroken() -> Bool {
        #if targetEnvironment(simulator)
        return false
        #else
        let suspiciousPaths = [
            "/Applications/Cydia.app",
            "/Library/MobileSubstrate/MobileSubstrate.dylib",
            "/bin/bash",
            "/usr/sbin/sshd",
            "/etc/apt",
            "/private/var/lib/apt/"
        ]
        for path in suspiciousPaths {
            if FileManager.default.fileExists(atPath: path) {
                return true
            }
        }
        let testPath = "/private/jailbreak_test_\(UUID().uuidString)"
        do {
            try "test".write(toFile: testPath, atomically: true, encoding: .utf8)
            try FileManager.default.removeItem(atPath: testPath)
            return true
        } catch {
            return false
        }
        #endif
    }
}
