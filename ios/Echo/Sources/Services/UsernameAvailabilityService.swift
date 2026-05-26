#if os(iOS)
import Foundation

struct UsernameAvailabilityResult: Sendable, Equatable {
    let username: String
    let available: Bool
    let reason: String?
}

/// Pre-auth username poll for first-run onboarding (WO-14). Public route — no passkey signing.
protocol UsernameAvailabilityClient: Sendable {
    func checkAvailability(username: String) async throws -> UsernameAvailabilityResult
}

struct UsernameAvailabilityService: UsernameAvailabilityClient {
    private let session: URLSession
    private let baseURL: URL

    init(session: URLSession = .shared, baseURL: URL? = nil) {
        self.session = session
        if let baseURL {
            self.baseURL = baseURL
        } else if let raw = ProcessInfo.processInfo.environment["ECHO_API_URL"],
                  let url = URL(string: raw) {
            self.baseURL = url
        } else {
            self.baseURL = APIConfiguration.default.baseURL
        }
    }

    func checkAvailability(username: String) async throws -> UsernameAvailabilityResult {
        let trimmed = username.trimmingCharacters(in: .whitespacesAndNewlines)
        guard UsernameValidator.isValid(trimmed) else {
            return UsernameAvailabilityResult(username: trimmed, available: false, reason: "invalid_format")
        }

        var components = URLComponents(
            url: baseURL.appendingPathComponent("v1/users/check-username"),
            resolvingAgainstBaseURL: false
        )
        components?.queryItems = [URLQueryItem(name: "username", value: trimmed)]
        guard let url = components?.url else {
            throw URLError(.badURL)
        }

        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }
        guard (200 ... 299).contains(http.statusCode) else {
            throw URLError(.badServerResponse)
        }

        let decoded = try JSONDecoder().decode(WireResponse.self, from: data)
        return UsernameAvailabilityResult(
            username: decoded.username,
            available: decoded.available,
            reason: decoded.reason
        )
    }

    private struct WireResponse: Decodable {
        let username: String
        let available: Bool
        let reason: String?
    }
}
#endif
