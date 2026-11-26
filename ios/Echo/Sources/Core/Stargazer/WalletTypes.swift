// Core/Stargazer/WalletTypes.swift
// Names used by the account recovery key. (The folder and type names are historical: this code
// once also served a token wallet, which has been removed. What remains derives the recovery key
// from the recovery phrase and proves ownership of it when an account is restored.)

import Foundation

enum WalletKeychain {
    /// Where the recovery key's address is kept on this device.
    static let dagAddressKey = "echo.wallet.dagAddress"
}

enum StargazerError: Error, LocalizedError {
    case walletCreationFailed

    var errorDescription: String? {
        switch self {
        case .walletCreationFailed: return "Could not create the recovery key"
        }
    }
}
