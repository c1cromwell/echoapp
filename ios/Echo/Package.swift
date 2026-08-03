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
    dependencies: [
        .package(url: "https://github.com/MarlonJD/MLKEMNativeSwift.git", from: "0.2.0"),
    ],
    targets: [
        .target(
            name: "Echo",
            dependencies: [
                .product(name: "MLKEMNativeSwift", package: "MLKEMNativeSwift"),
            ],
            path: "Sources",
            exclude: ["App/EchoApp.swift"]  // entry point lives in EchoApp.xcodeproj, not the package
        ),
        // EchoTests omitted from SPM host runs: pre-existing iOS-only compile failures
        // block EchoPackageTests linking. Prefer EchoSecurityTests / EchoPhase3Tests.
        // .testTarget(
        //     name: "EchoTests",
        //     dependencies: ["Echo"],
        //     path: "Tests"
        // ),
        // Isolated security tests (WO-208/211/223/224) that compile cleanly.
        // Kept separate from EchoTests because EchoTests has pre-existing
        // compilation issues unrelated to these work orders.
        .testTarget(
            name: "EchoSecurityTests",
            dependencies: ["Echo"],
            path: "SecurityTests"
        ),
        .testTarget(
            name: "EchoPhase3Tests",
            dependencies: ["Echo"],
            path: "Phase3Tests"
        ),
        .testTarget(
            name: "EchoPhase2Tests",
            dependencies: ["Echo"],
            path: "Phase2Tests"
        ),
        .testTarget(
            name: "EchoScreenCatalogTests",
            dependencies: ["Echo"],
            path: "ScreenCatalog"
        ),
    ]
)
