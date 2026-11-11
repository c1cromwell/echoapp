import Foundation

struct BarPerson: Identifiable, Equatable, Sendable {
    var did: String
    var name: String
    var allowed: Bool

    var id: String { did }
}

struct BarLine: Identifiable, Equatable, Sendable {
    var id: String
    var mine: Bool
    var text: String
}

enum BarQuery: Equatable {
    case empty
    case emailNotLive
    case username(String)
}

enum BarThread {
    static func id(a: String, b: String) -> String {
        let pair = [a, b].sorted()
        return "dm:\(pair[0]):\(pair[1])"
    }
}

/// Who can be opened in Echo Bar, and the lines typed in this window.
struct BarSession: Equatable, Sendable {
    var people: [BarPerson] = []
    var selectedDID: String?
    var lines: [String: [BarLine]] = [:]
    var notice: String?

    var selected: BarPerson? {
        guard let selectedDID else { return nil }
        return people.first { $0.did == selectedDID && $0.allowed }
    }

    var allowed: [BarPerson] { people.filter(\.allowed) }
    var candidates: [BarPerson] { people.filter { !$0.allowed } }

    func canOpen(_ did: String) -> Bool {
        people.contains { $0.did == did && $0.allowed }
    }

    /// A discovery match is a candidate. It does not allow them.
    mutating func noteDiscovery(_ found: [BarPerson]) {
        for person in found {
            if people.contains(where: { $0.did == person.did }) { continue }
            people.append(BarPerson(did: person.did, name: person.name, allowed: false))
        }
    }

    mutating func allow(_ person: BarPerson) {
        if let index = people.firstIndex(where: { $0.did == person.did }) {
            people[index].allowed = true
            if people[index].name.isEmpty {
                people[index].name = person.name
            }
        } else {
            var copy = person
            copy.allowed = true
            people.append(copy)
        }
        notice = nil
    }

    mutating func revoke(did: String) {
        people.removeAll { $0.did == did }
        if selectedDID == did {
            selectedDID = nil
        }
    }

    mutating func select(did: String) {
        guard canOpen(did) else { return }
        selectedDID = did
        notice = nil
    }

    mutating func showList() {
        selectedDID = nil
    }

    static func interpret(_ raw: String) -> BarQuery {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty { return .empty }
        if trimmed.contains("@") && !trimmed.hasPrefix("@") {
            return .emailNotLive
        }
        let handle = trimmed.hasPrefix("@") ? String(trimmed.dropFirst()) : trimmed
        let name = handle.trimmingCharacters(in: .whitespacesAndNewlines)
        if name.isEmpty || name.contains("@") { return .empty }
        return .username(name)
    }

    /// Appends a line when the selected person is allowed and the text is non-empty.
    mutating func send(_ text: String) -> BarLine? {
        guard let person = selected else { return nil }
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        let line = BarLine(id: UUID().uuidString, mine: true, text: trimmed)
        lines[person.did, default: []].append(line)
        return line
    }
}
