// Presentation/Screens/QRIdentity/QRIdentityView.swift
// QR code identity sharing and scanning for trusted contact attestation

import SwiftUI
import CoreImage.CIFilterBuiltins
#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

// MARK: - QR Identity View

struct QRIdentityView: View {
    @StateObject private var viewModel = QRIdentityViewModel()
    @State private var showScanner = false

    var body: some View {
        VStack(spacing: 32) {
            SecureThreadIndicator()

            Spacer()

            // QR Card — frosted glass
            VStack(spacing: 20) {
                // QR Code
                if let qrImage = viewModel.qrCodeImage {
                    #if canImport(UIKit)
                    Image(uiImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .frame(width: 200, height: 200)
                        .clipShape(RoundedRectangle(cornerRadius: 16))
                    #elseif canImport(AppKit)
                    Image(nsImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .frame(width: 200, height: 200)
                        .clipShape(RoundedRectangle(cornerRadius: 16))
                    #endif
                }

                // Identity info
                Text(viewModel.echoHandle)
                    .font(.custom("Inter", size: 16))
                    .fontWeight(.bold)
                    .foregroundStyle(Color.Echo.primaryContainer)

                Text(viewModel.didShort)
                    .font(.custom("Inter", size: 12))
                    .foregroundStyle(Color.Echo.outline)
                    .lineLimit(1)
                    .truncationMode(.middle)

                HStack(spacing: 6) {
                    Image(systemName: "lock.fill")
                        .font(.system(size: 10))
                    Text("Trust Score: \(viewModel.trustScore)/100")
                        .font(Font.Echo.labelMd)
                }
                .foregroundStyle(Color.Echo.outline)
            }
            .padding(32)
            .frame(maxWidth: .infinity)
            .background(
                RoundedRectangle(cornerRadius: 32)
                    .fill(.ultraThinMaterial)
                    .opacity(0.6)
            )
            .ghostBorder()
            .padding(.horizontal, 32)

            Spacer()

            // Action buttons
            HStack(spacing: 16) {
                Button {
                    showScanner = true
                } label: {
                    Label("Scan QR", systemImage: "camera.fill")
                        .font(.custom("Inter", size: 14))
                        .fontWeight(.bold)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 9999)
                                .fill(LinearGradient.signature)
                        )
                }
                .deepGlacialShadow()

                Button {
                    viewModel.shareLink()
                } label: {
                    Label("Share Link", systemImage: "square.and.arrow.up")
                        .font(.custom("Inter", size: 14))
                        .fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 9999)
                                .fill(Color.Echo.surfaceContainerLow)
                        )
                        .ghostBorder(opacity: 0.15)
                }
            }
            .padding(.horizontal, 32)

            Text("Scan another user's QR code to add them\nas a trusted contact with verified\nin-person attestation.")
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
                .multilineTextAlignment(.center)
                .padding(.bottom, 40)
        }
        .icyBackground()
        .navigationTitle("My Identity")
        #if os(iOS)
        .fullScreenCover(isPresented: $showScanner) {
            QRScannerView(onScan: { did in
                showScanner = false
                viewModel.handleScannedDID(did)
            })
        }
        #else
        .sheet(isPresented: $showScanner) {
            QRScannerView(onScan: { did in
                showScanner = false
                viewModel.handleScannedDID(did)
            })
        }
        #endif
    }
}

// MARK: - QR Identity ViewModel

@MainActor
class QRIdentityViewModel: ObservableObject {
    @Published var echoHandle: String = ""
    @Published var did: String = ""
    @Published var trustScore: Int = 0
    #if canImport(UIKit)
    @Published var qrCodeImage: UIImage?
    #elseif canImport(AppKit)
    @Published var qrCodeImage: NSImage?
    #endif

    var didShort: String {
        guard did.count > 20 else { return did }
        return "\(did.prefix(12))...\(did.suffix(6))"
    }

    init() {
        // Load identity info
        loadIdentity()
    }

    func loadIdentity() {
        // TODO: Load from identity service
        echoHandle = "echo:user"
        did = "did:cardano:addr1q8example"
        trustScore = 72
        generateQRCode()
    }

    func generateQRCode() {
        let context = CIContext()
        let filter = CIFilter.qrCodeGenerator()
        let data = Data(did.utf8)
        filter.setValue(data, forKey: "inputMessage")
        filter.setValue("M", forKey: "inputCorrectionLevel")

        guard let outputImage = filter.outputImage else { return }

        // Scale up for crisp rendering
        let transform = CGAffineTransform(scaleX: 10, y: 10)
        let scaledImage = outputImage.transformed(by: transform)

        guard let cgImage = context.createCGImage(scaledImage, from: scaledImage.extent) else { return }
        #if canImport(UIKit)
        qrCodeImage = UIImage(cgImage: cgImage)
        #elseif canImport(AppKit)
        qrCodeImage = NSImage(cgImage: cgImage, size: NSSize(width: scaledImage.extent.width, height: scaledImage.extent.height))
        #endif
    }

    func handleScannedDID(_ did: String) {
        // TODO: Validate DID and initiate trust handshake
    }

    func shareLink() {
        // TODO: Share via UIActivityViewController
    }
}

// MARK: - QR Scanner View

struct QRScannerView: View {
    let onScan: (String) -> Void
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ZStack {
            Color.Echo.deepNavy.ignoresSafeArea()

            VStack(spacing: 24) {
                // Camera preview placeholder
                RoundedRectangle(cornerRadius: 32)
                    .fill(Color.Echo.surfaceContainerHigh.opacity(0.3))
                    .frame(width: 280, height: 280)
                    .overlay(
                        RoundedRectangle(cornerRadius: 32)
                            .stroke(Color.Echo.primaryContainer, lineWidth: 2)
                    )
                    .overlay(
                        VStack(spacing: 12) {
                            Image(systemName: "camera.viewfinder")
                                .font(.system(size: 48))
                                .foregroundStyle(Color.Echo.primaryContainer)
                            Text("Point camera at QR code")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(.white.opacity(0.7))
                        }
                    )

                Button("Cancel") { dismiss() }
                    .font(.custom("Inter", size: 16))
                    .fontWeight(.bold)
                    .foregroundStyle(.white)
                    .padding(.horizontal, 32)
                    .padding(.vertical, 14)
                    .background(
                        Capsule()
                            .fill(.ultraThinMaterial)
                            .opacity(0.3)
                    )
            }
        }
    }
}
