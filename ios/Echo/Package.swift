// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "Echo",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(
            name: "Echo",
            targets: ["Echo"]
        ),
    ],
    dependencies: [],
    targets: [
        .target(
            name: "Echo",
            dependencies: [],
            path: "Sources",
            exclude: ["App/EchoApp.swift"]  // entry point lives in EchoApp.xcodeproj, not the package
        ),
        .testTarget(
            name: "EchoTests",
            dependencies: ["Echo"],
            path: "Tests"
        ),
        // Isolated security tests (WO-208/211/223/224) that compile cleanly.
        // Kept separate from EchoTests because EchoTests has pre-existing
        // compilation issues unrelated to these work orders.
        .testTarget(
            name: "EchoSecurityTests",
            dependencies: ["Echo"],
            path: "SecurityTests"
        ),
    ]
)
