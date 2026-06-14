#if os(iOS)
import Foundation

struct CallICEServer: Decodable, Sendable {
    let urls: [String]
    let username: String?
    let credential: String?
}

struct CallICEConfigResponse: Decodable, Sendable {
    let iceServers: [CallICEServer]

    enum CodingKeys: String, CodingKey {
        case iceServers = "ice_servers"
    }
}

enum CallICEEndpoint: APIEndpoint {
    case iceServers

    var path: String { "/v3/calls/ice-servers" }
}

actor CallICEAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchICEServers() async throws -> [CallICEServer] {
        let resp: CallICEConfigResponse = try await apiClient.get(endpoint: CallICEEndpoint.iceServers)
        return resp.iceServers
    }
}
#endif
