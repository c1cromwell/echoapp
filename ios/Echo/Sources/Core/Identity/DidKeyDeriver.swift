import Foundation

/// Canonical `did:key` derivation for P-256 keys (matches Go `pkg/didkey`).
enum DidKeyDeriver {

    static let prefix = "did:key:z"

    private static let p256MulticodecVarint: [UInt8] = [0x80, 0x24]
    private static let compressedLen = 33

    enum Error: LocalizedError {
        case invalidPublicKey(String)

        var errorDescription: String? {
            switch self {
            case .invalidPublicKey(let reason):
                return "Invalid P-256 public key: \(reason)"
            }
        }
    }

    /// Derives `did:key:z…` from lowercase hex (uncompressed `04||X||Y` or compressed `02/03||X`).
    static func deriveFromPublicKeyHex(_ hexStr: String) throws -> String {
        var hex = hexStr.trimmingCharacters(in: .whitespacesAndNewlines)
        if hex.hasPrefix("0x") || hex.hasPrefix("0X") {
            hex = String(hex.dropFirst(2))
        }
        guard hex.count % 2 == 0,
              let raw = Data(hexString: hex) else {
            throw Error.invalidPublicKey("hex decode failed")
        }
        return try deriveFromPublicKeyBytes(raw)
    }

    static func deriveFromPublicKeyBytes(_ raw: Data) throws -> String {
        let compressed = try compressP256(raw)
        var body = Data(p256MulticodecVarint)
        body.append(contentsOf: compressed)
        return prefix + Base58BTC.encode(body)
    }

    private static func compressP256(_ raw: Data) throws -> Data {
        switch raw.count {
        case compressedLen:
            guard raw[0] == 0x02 || raw[0] == 0x03 else {
                throw Error.invalidPublicKey("invalid compressed prefix")
            }
            return raw
        case 65:
            guard raw[0] == 0x04 else {
                throw Error.invalidPublicKey("expected uncompressed SEC1 prefix 0x04")
            }
            let yByte = raw[64]
            var out = Data(count: compressedLen)
            out[0] = (yByte & 1) == 1 ? 0x03 : 0x02
            out.replaceSubrange(1 ..< compressedLen, with: raw[1 ..< 33])
            return out
        default:
            throw Error.invalidPublicKey("expected 33 or 65 bytes, got \(raw.count)")
        }
    }
}

// MARK: - Base58btc (multibase `z`)

private enum Base58BTC {
    private static let alphabet = Array("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz".utf8)

    static func encode(_ input: Data) -> String {
        guard !input.isEmpty else { return "" }

        var zeros = 0
        while zeros < input.count && input[zeros] == 0 {
            zeros += 1
        }

        var value = [UInt8](input)
        var encoded: [UInt8] = []

        var start = zeros
        while start < value.count {
            var carry = 0
            for i in start ..< value.count {
                let combined = carry * 256 + Int(value[i])
                value[i] = UInt8(combined / 58)
                carry = combined % 58
            }
            encoded.append(alphabet[carry])
            while start < value.count && value[start] == 0 {
                start += 1
            }
        }

        encoded.append(contentsOf: repeatElement(alphabet[0], count: zeros))
        encoded.reverse()
        return String(bytes: encoded, encoding: .utf8) ?? ""
    }
}

private extension Data {
    init?(hexString: String) {
        guard hexString.count % 2 == 0 else { return nil }
        var data = Data(capacity: hexString.count / 2)
        var index = hexString.startIndex
        while index < hexString.endIndex {
            let next = hexString.index(index, offsetBy: 2)
            guard next <= hexString.endIndex else { return nil }
            let byte = hexString[index ..< next]
            guard let value = UInt8(byte, radix: 16) else { return nil }
            data.append(value)
            index = next
        }
        self = data
    }
}
