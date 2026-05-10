#if os(iOS)
import Foundation
import os

// WO-6: Privacy-safe structured logger.
//
// # T0–T7 contract
//
// | Pattern               | Action   | Rationale |
// |-----------------------|----------|-----------|
// | 64-char hex (32 byte) | REDACTED | Likely T0/T1 private key |
// | 32-char hex (16 byte) | REDACTED | Intermediate key material |
// | email address         | REDACTED | PII |
// | E.164 phone number    | REDACTED | PII |
// | DID (did:…)           | KEPT     | T7 public chain data |
//
// Usage:
//   let log = EchoLogger(subsystem: "identity")
//   log.info("registered DID", ["did": did, "method": "key"])
//   log.warn("slow relay", ["ms": 340])

/// Log severity levels.
enum LogLevel: Int, Comparable {
    case debug = 0, info, warn, error

    static func < (lhs: LogLevel, rhs: LogLevel) -> Bool { lhs.rawValue < rhs.rawValue }

    var label: String {
        switch self {
        case .debug: return "debug"
        case .info:  return "info"
        case .warn:  return "warn"
        case .error: return "error"
        }
    }

    var osLogType: OSLogType {
        switch self {
        case .debug: return .debug
        case .info:  return .info
        case .warn:  return .default
        case .error: return .error
        }
    }
}

/// Privacy-safe structured logger.
///
/// All string values — message and fields — are passed through `sanitize(_:)`
/// before being handed to `os.Logger`, ensuring that PII never reaches system
/// logs or crash reporters.
struct EchoLogger {
    let subsystem: String
    var minimumLevel: LogLevel = .info

    private static let bundleID = Bundle.main.bundleIdentifier ?? "com.echo"

    private var osLogger: os.Logger {
        os.Logger(subsystem: Self.bundleID, category: subsystem)
    }

    // MARK: - Logging API

    func debug(_ message: String, _ fields: [String: Any] = [:]) {
        guard minimumLevel <= .debug else { return }
        emit(.debug, message, fields)
    }

    func info(_ message: String, _ fields: [String: Any] = [:]) {
        guard minimumLevel <= .info else { return }
        emit(.info, message, fields)
    }

    func warn(_ message: String, _ fields: [String: Any] = [:]) {
        guard minimumLevel <= .warn else { return }
        emit(.warn, message, fields)
    }

    func error(_ message: String, _ fields: [String: Any] = [:]) {
        guard minimumLevel <= .error else { return }
        emit(.error, message, fields)
    }

    // MARK: - Sanitisation

    /// Removes known PII patterns, replacing them with "[REDACTED]".
    static func sanitize(_ input: String) -> String {
        var s = input

        // 32-byte private key (64 lowercase hex chars)
        s = s.replacingOccurrences(
            of: #"\b[0-9a-f]{64}\b"#,
            with: "[REDACTED]",
            options: .regularExpression
        )
        // 16-byte intermediate key (32 lowercase hex chars)
        s = s.replacingOccurrences(
            of: #"\b[0-9a-f]{32}\b"#,
            with: "[REDACTED]",
            options: .regularExpression
        )
        // Email address
        s = s.replacingOccurrences(
            of: #"[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}"#,
            with: "[REDACTED]",
            options: .regularExpression
        )
        // E.164 phone number
        s = s.replacingOccurrences(
            of: #"\+[1-9]\d{9,14}\b"#,
            with: "[REDACTED]",
            options: .regularExpression
        )
        return s
    }

    // MARK: - Internals

    private func emit(_ level: LogLevel, _ message: String, _ fields: [String: Any]) {
        let safeMsg = Self.sanitize(message)
        let safeFields = fields.mapValues { value -> Any in
            if let s = value as? String { return Self.sanitize(s) }
            return value
        }
        let fieldStr = safeFields.isEmpty ? "" :
            " " + safeFields.map { "\($0.key)=\($0.value)" }.sorted().joined(separator: " ")
        let line = "[\(subsystem)] \(safeMsg)\(fieldStr)"

        switch level {
        case .debug: osLogger.debug("\(line, privacy: .public)")
        case .info:  osLogger.info("\(line, privacy: .public)")
        case .warn:  osLogger.warning("\(line, privacy: .public)")
        case .error: osLogger.error("\(line, privacy: .public)")
        }
    }
}

// MARK: - Convenience singleton

extension EchoLogger {
    /// Shared default logger; configure minimumLevel from settings.
    static var shared = EchoLogger(subsystem: "echo")
}
#endif
