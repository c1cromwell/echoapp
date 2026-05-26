#if os(iOS)
import Foundation
import Contacts

struct DiscoveredContact: Identifiable, Sendable, Equatable {
    let id: String
    let did: String
    let displayName: String
    let phoneE164: String
}

/// Private contact discovery via OPRF-PSI (WO-221).
actor ContactDiscoveryService {
    private let oprf: any OPRFClient
    private let api: ContactDiscoveryAPIClient
    private let batchSize = 500

    init(oprf: any OPRFClient, api: ContactDiscoveryAPIClient) {
        self.oprf = oprf
        self.api = api
    }

    func discoverFromDeviceContacts() async throws -> [DiscoveredContact] {
        let phones = try await fetchNormalizedPhoneNumbers()
        guard !phones.isEmpty else { throw ContactDiscoveryError.noMatches }

        var matches: [DiscoveredContact] = []
        for chunk in phones.chunked(into: batchSize) {
            let batch = chunk.map(\.e164)
            let blind = try await oprf.blind(phones: batch)
            let response: ContactDiscoveryAPIClient.PSIResponse
            do {
                response = try await api.evaluate(blinded: blind.blinded)
            } catch {
                throw ContactDiscoveryError.discoveryUnavailable
            }
            let keys = try await oprf.finalize(sessionID: blind.sessionID, evaluated: response.evaluated)
            for (idx, key) in keys.enumerated() {
                guard let did = response.index[key], !did.isEmpty else { continue }
                let phone = batch[idx]
                let label = chunk.first(where: { $0.e164 == phone })?.label ?? phone
                matches.append(DiscoveredContact(
                    id: did,
                    did: did,
                    displayName: label,
                    phoneE164: phone
                ))
            }
        }
        return matches
    }

    private struct LabeledPhone: Sendable {
        let e164: String
        let label: String
    }

    private func fetchNormalizedPhoneNumbers() async throws -> [LabeledPhone] {
        let store = CNContactStore()
        let granted = try await requestContactsAccess(store: store)
        guard granted else { throw ContactDiscoveryError.permissionDenied }

        let keys: [CNKeyDescriptor] = [
            CNContactPhoneNumbersKey as CNKeyDescriptor,
            CNContactFormatter.descriptorForRequiredKeys(for: .fullName),
        ]
        var results: [LabeledPhone] = []
        let request = CNContactFetchRequest(keysToFetch: keys)
        try store.enumerateContacts(with: request) { contact, _ in
            let label = CNContactFormatter.string(from: contact, style: .fullName) ?? "Contact"
            for phone in contact.phoneNumbers {
                let e164 = PhoneNormalizer.e164(from: phone.value.stringValue)
                guard e164.count >= 8 else { continue }
                results.append(LabeledPhone(e164: PhoneNormalizer.normalize(e164), label: label))
            }
        }
        // De-dupe by E.164
        var seen = Set<String>()
        return results.filter { seen.insert($0.e164).inserted }
    }

    private func requestContactsAccess(store: CNContactStore) async throws -> Bool {
        try await withCheckedThrowingContinuation { continuation in
            store.requestAccess(for: .contacts) { granted, error in
                if let error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume(returning: granted)
                }
            }
        }
    }
}

private extension Array {
    func chunked(into size: Int) -> [[Element]] {
        guard size > 0 else { return [self] }
        var out: [[Element]] = []
        var idx = 0
        while idx < count {
            out.append(Array(self[idx ..< Swift.min(idx + size, count)]))
            idx += size
        }
        return out
    }
}
#endif
