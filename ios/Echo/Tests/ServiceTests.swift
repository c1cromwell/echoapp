import XCTest
@testable import Echo

final class ServiceTests: XCTestCase {
    
    func testBaseEchoServiceInit() {
        let service = BaseEchoService(name: "test-service", version: "v1.0")
        
        XCTAssertEqual(service.name, "test-service")
        XCTAssertEqual(service.version, "v1.0")
        XCTAssertEqual(service.status, "initialized")
    }
    
    func testServiceStatusUpdate() {
        let service = BaseEchoService(name: "test-service", version: "v1.0")
        
        XCTAssertEqual(service.getStatus(), "initialized")
        
        service.setStatus("running")
        XCTAssertEqual(service.getStatus(), "running")
    }
    
    @MainActor
    func testServiceHealth() async throws {
        let service = BaseEchoService(name: "test-service", version: "v1.0")
        
        let health = try await service.health()
        
        XCTAssertEqual(health["status"] as? String, "initialized")
        XCTAssertEqual(health["service"] as? String, "test-service")
        XCTAssertEqual(health["version"] as? String, "v1.0")
    }
    
    @MainActor
    func testServiceStartStop() async throws {
        let service = BaseEchoService(name: "test-service", version: "v1.0")
        
        try await service.start()
        XCTAssertEqual(service.getStatus(), "running")
        
        try await service.stop()
        XCTAssertEqual(service.getStatus(), "stopped")
    }
    
    func testServiceRegistry() {
        let services = ["identity", "messaging", "trust", "rewards", "contacts", "metagraph-gateway", "notification", "media"]
        
        for serviceName in services {
            let service = ServiceRegistry.getService(serviceName)
            XCTAssertNotNil(service, "Service \(serviceName) should exist")
        }
    }
    
    func testServiceRegistryPorts() {
        let expectedPorts: [String: Int] = [
            "identity": 8001,
            "messaging": 8002,
            "trust": 8003,
            "rewards": 8004,
            "contacts": 8005,
            "metagraph-gateway": 8006,
            "notification": 8007,
            "media": 8008,
        ]
        
        for (name, expectedPort) in expectedPorts {
            let service = ServiceRegistry.getService(name)
            XCTAssertNotNil(service, "Service \(name) should exist")
            if let service = service {
                XCTAssertEqual(service.port, expectedPort, "Port for \(name) should be \(expectedPort)")
            }
        }
    }
    
    func testServiceDependencies() {
        let identity = ServiceRegistry.getService("identity")
        XCTAssertNotNil(identity)
        XCTAssertTrue(identity?.dependencies.contains("postgres") ?? false)
        XCTAssertTrue(identity?.dependencies.contains("redis") ?? false)
    }
}
