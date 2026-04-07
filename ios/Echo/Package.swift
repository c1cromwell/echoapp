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
            path: "Sources"
        ),
        .testTarget(
            name: "EchoTests",
            dependencies: ["Echo"],
            path: "Tests"
        ),
    ]
)
