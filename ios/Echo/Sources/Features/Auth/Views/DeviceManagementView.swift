import SwiftUI

struct DeviceManagementView: View {
    @StateObject var viewModel: DeviceManagementViewModel
    let onStepUpRequired: (StepUpAction, @escaping (String) -> Void) -> Void
    let onLoggedOut: () -> Void

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(title: "Devices", showBackButton: true)

                if viewModel.isLoading {
                    Spacer()
                    ProgressView()
                    Spacer()
                } else {
                    ScrollView {
                        VStack(spacing: Spacing.lg.rawValue) {
                            // Current device
                            if let current = viewModel.currentDevice {
                                VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                    Text("This Device")
                                        .typographyStyle(.h4, color: .echoPrimaryText)

                                    DeviceRowView(
                                        device: current,
                                        isCurrent: true,
                                        onRevoke: nil
                                    )
                                }
                            }

                            // Other devices
                            if !viewModel.otherDevices.isEmpty {
                                VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                    Text("Other Devices")
                                        .typographyStyle(.h4, color: .echoPrimaryText)

                                    ForEach(viewModel.otherDevices) { device in
                                        DeviceRowView(
                                            device: device,
                                            isCurrent: false,
                                            onRevoke: { revokeDevice(id: device.id) }
                                        )
                                    }
                                }
                            }

                            // Log out all devices
                            if !viewModel.otherDevices.isEmpty {
                                EchoButton(
                                    "Log out all devices",
                                    style: .secondary,
                                    size: .medium,
                                    icon: Image(systemName: "rectangle.portrait.and.arrow.right"),
                                    action: {
                                        Task {
                                            await viewModel.logoutAllDevices()
                                            onLoggedOut()
                                        }
                                    }
                                )
                            }
                        }
                        .padding(.horizontal, Spacing.lg.rawValue)
                        .padding(.top, Spacing.md.rawValue)
                    }
                }

                // Error
                if let error = viewModel.errorMessage {
                    Text(error)
                        .typographyStyle(.caption, color: .red)
                        .padding()
                }
            }
        }
        .task { await viewModel.loadDevices() }
    }

    private func revokeDevice(id: String) {
        onStepUpRequired(.revokeDevice) { elevatedToken in
            Task {
                _ = await viewModel.revokeDevice(id: id, elevatedToken: elevatedToken)
            }
        }
    }
}

// MARK: - Device Row

struct DeviceRowView: View {
    let device: DeviceSession
    let isCurrent: Bool
    let onRevoke: (() -> Void)?

    var body: some View {
        HStack(spacing: Spacing.md.rawValue) {
            Image(systemName: device.platform == "ios" ? "iphone" : "desktopcomputer")
                .font(.system(size: 24))
                .foregroundColor(.echoPrimary)
                .frame(width: 40, height: 40)

            VStack(alignment: .leading, spacing: 2) {
                HStack {
                    Text(device.friendlyName)
                        .typographyStyle(.bodyLarge, color: .echoPrimaryText)

                    if isCurrent {
                        Text("Current")
                            .typographyStyle(.caption, color: .white)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 2)
                            .background(Color.echoPrimary)
                            .cornerRadius(8)
                    }
                }

                Text("\(device.platform.uppercased()) \(device.osVersion)")
                    .typographyStyle(.caption, color: .echoSecondaryText)

                if let location = device.lastLocation {
                    Text("Last active: \(location)")
                        .typographyStyle(.caption, color: .echoGray500)
                }
            }

            Spacer()

            if !isCurrent, let onRevoke {
                Button(action: onRevoke) {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 20))
                        .foregroundColor(.echoGray400)
                }
            }
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
    }
}
