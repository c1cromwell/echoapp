import Foundation

/// Base protocol for all services
public protocol EchoService {
    var name: String { get }
    var version: String { get }
    
    func health() async throws -> [String: Any]
    func start() async throws
    func stop() async throws
}

/// Base service implementation
public class BaseEchoService: EchoService {
    public let name: String
    public let version: String
    private(set) public var status: String = "initialized"
    
    public init(name: String, version: String) {
        self.name = name
        self.version = version
    }
    
    public func health() async throws -> [String: Any] {
        return [
            "status": status,
            "service": name,
            "version": version,
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]
    }
    
    public func start() async throws {
        status = "running"
    }
    
    public func stop() async throws {
        status = "stopped"
    }
    
    // MARK: - Helper Methods
    
    public func setStatus(_ newStatus: String) {
        self.status = newStatus
    }
    
    public func getStatus() -> String {
        return status
    }
}

/// Service registry definition
public struct ServiceDefinition {
    public let name: String
    public let version: String
    public let port: Int
    public let healthEndpoint: String
    public let dependencies: [String]
    
    public init(name: String, version: String, port: Int, healthEndpoint: String = "/health", dependencies: [String] = []) {
        self.name = name
        self.version = version
        self.port = port
        self.healthEndpoint = healthEndpoint
        self.dependencies = dependencies
    }
}

/// Service registry
public enum ServiceRegistry {
    public static let services: [String: ServiceDefinition] = [
        "identity": ServiceDefinition(
            name: "identity-service",
            version: "v1",
            port: 8001,
            dependencies: ["postgres", "redis"]
        ),
        "messaging": ServiceDefinition(
            name: "messaging-service",
            version: "v1",
            port: 8002,
            dependencies: ["postgres", "redis", "metagraph"]
        ),
        "trust": ServiceDefinition(
            name: "trust-service",
            version: "v1",
            port: 8003,
            dependencies: ["postgres", "redis", "metagraph"]
        ),
        "rewards": ServiceDefinition(
            name: "rewards-service",
            version: "v1",
            port: 8004,
            dependencies: ["postgres", "redis", "metagraph"]
        ),
        "contacts": ServiceDefinition(
            name: "contacts-service",
            version: "v1",
            port: 8005,
            dependencies: ["postgres", "redis"]
        ),
        "metagraph-gateway": ServiceDefinition(
            name: "metagraph-gateway",
            version: "v1",
            port: 8006,
            dependencies: ["metagraph-l0", "metagraph-l1", "redis"]
        ),
        "notification": ServiceDefinition(
            name: "notification-service",
            version: "v1",
            port: 8007,
            dependencies: ["redis", "apns"]
        ),
        "media": ServiceDefinition(
            name: "media-service",
            version: "v1",
            port: 8008,
            dependencies: ["postgres", "s3", "cdn"]
        ),
    ]
    
    public static func getService(_ name: String) -> ServiceDefinition? {
        return services[name]
    }
}
