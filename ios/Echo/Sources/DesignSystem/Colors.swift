import SwiftUI

// MARK: - Echo Refined Color System
// Based on Claude Design review: "Make the security visible. Make the rest disappear."
//
// Three surfaces:
//   paper  — warm off-white default; reads as real, not synthetic
//   ink    — near-black text; one accent (signal blue), one affirming hue (trust green)
//   night  — dark surface used ONLY for private moments (hidden persona, vault, recovery)
//            so privacy *feels* different, not just looks it

// MARK: - Primitives

public extension Color {

    // Surface
    static let echoPaper     = Color(hex: 0xFBFBF7) // warm paper — default background
    static let echoPaperDim  = Color(hex: 0xF2F1EA) // inset areas, cards
    static let echoPaperEdge = Color(hex: 0xE6E4DA) // subtle borders

    // Ink (primary text + UI)
    static let echoInk       = Color(hex: 0x0B0F10)
    static let echoInk70     = Color(rgba: 0x0B0F10, opacity: 0.70)
    static let echoInk55     = Color(rgba: 0x0B0F10, opacity: 0.55)
    static let echoInk40     = Color(rgba: 0x0B0F10, opacity: 0.40)
    static let echoInk20     = Color(rgba: 0x0B0F10, opacity: 0.18)
    static let echoInk10     = Color(rgba: 0x0B0F10, opacity: 0.09)
    static let echoInk05     = Color(rgba: 0x0B0F10, opacity: 0.045)
    static let echoHair      = Color(rgba: 0x0B0F10, opacity: 0.10)

    // Night — reserved for hidden / locked / private surfaces
    static let echoNight     = Color(hex: 0x0E1418)
    static let echoNightHi   = Color(hex: 0x161D22)
    static let echoNightInk  = Color(hex: 0xF4F1E8)
    static let echoNightInk70 = Color(rgba: 0xF4F1E8, opacity: 0.70)
    static let echoNightInk40 = Color(rgba: 0xF4F1E8, opacity: 0.40)
    static let echoNightHair  = Color(rgba: 0xF4F1E8, opacity: 0.10)

    // One accent — Echo signal blue (kept from brand)
    static let echoSignal    = Color(hex: 0x0E7AB8)
    static let echoSignalDim = Color(hex: 0x1A4E70)

    // One affirming hue — trust green (replaces 6-color trust rainbow)
    static let echoTrustGreen    = Color(hex: 0x1F7A4C)
    static let echoTrustGreenDim = Color(hex: 0x0F4D2E)

    // Alert — less saturated than pure red
    static let echoAlert     = Color(hex: 0xB5341B)

    // MARK: - Backward-compat aliases (referenced by existing components)
    static let echoWarning      = Color(hex: 0xF59E0B) // amber — kept for existing use sites
    static let echoInfo         = Color(hex: 0x3B82F6) // blue info state
    static let echoSecondary    = Color(hex: 0x64748B) // slate — kept for backward compat
    static let echoSurface      = echoPaperDim          // maps to warm inset surface
    static let echoLightSurface = echoPaperDim          // backward compat
    static let echoDarkBg       = echoNight             // backward compat
    static let echoDarkSurface  = echoNightHi           // backward compat
    static let echoLightBg      = echoPaper             // backward compat

    // MARK: - Trust — One hue, five opacities
    // Replaces the loyalty-program feel of grey/blue/green/purple/amber/pink.
    // T0 = none · T1 = seen · T2 = verified · T3 = trusted · T4 = primary
    static let echoTrustUnverified = Color(rgba: 0x1F7A4C, opacity: 0.10)
    static let echoTrustBasic      = Color(rgba: 0x1F7A4C, opacity: 0.25)
    static let echoTrustVerified   = Color(rgba: 0x1F7A4C, opacity: 0.45)
    static let echoTrustTrusted    = Color(rgba: 0x1F7A4C, opacity: 0.70)
    static let echoTrustPremium    = Color(rgba: 0x1F7A4C, opacity: 0.85)
    static let echoTrustElite      = Color(hex: 0x1F7A4C) // full

    // MARK: - Semantic aliases (consumed by components)
    static let echoPrimaryText    = echoInk
    static let echoSecondaryText  = echoInk55
    static let echoPrimary        = echoSignal        // backward compat
    static let echoBackground     = echoPaper
    static let echoCardBackground = echoPaperDim
    static let echoError          = echoAlert
    static let echoSuccess        = echoTrustGreen

    // Legacy trust palette — mapped to new single-hue system
    // Kept so existing call sites don't break.
    static let echoTrustNewcomer     = echoTrustUnverified
    static let echoTrustHighlyTrusted = echoTrustTrusted

    // Gray scale — kept for backward compatibility
    static let echoGray50  = Color(hex: 0xF9FAFB)
    static let echoGray100 = Color(hex: 0xF3F4F6)
    static let echoGray200 = Color(hex: 0xE5E7EB)
    static let echoGray300 = Color(hex: 0xD1D5DB)
    static let echoGray400 = Color(hex: 0x9CA3AF)
    static let echoGray500 = Color(hex: 0x6B7280)
    static let echoGray600 = Color(hex: 0x4B5563)
    static let echoGray700 = Color(hex: 0x374151)
    static let echoGray900 = Color(hex: 0x111827)

    // Trust-level function (uses new single-hue system)
    static func trustColor(for level: String) -> Color {
        switch level.lowercased() {
        case "newcomer", "t0", "unverified":  return .echoTrustUnverified
        case "basic", "t1", "seen", "member": return .echoTrustBasic
        case "verified", "t2":                return .echoTrustVerified
        case "trusted", "t3":                 return .echoTrustTrusted
        case "premium", "t4", "highlytrusted": return .echoTrustPremium
        case "elite", "t5", "primary":        return .echoTrustElite
        default:                              return .echoTrustUnverified
        }
    }
}

// MARK: - Hex + RGBA initialiser

extension Color {
    /// Initialise from a 6-digit hex value (e.g. `0x0E7AB8`).
    init(hex: UInt32) {
        let r = Double((hex >> 16) & 0xFF) / 255.0
        let g = Double((hex >> 8) & 0xFF) / 255.0
        let b = Double(hex & 0xFF) / 255.0
        self.init(red: r, green: g, blue: b)
    }

    /// Initialise from a hex value with an explicit opacity.
    /// `opacity` is in [0, 1]. The hex encodes the fully-opaque colour.
    init(rgba hex: UInt32, opacity: Double) {
        let r = Double((hex >> 16) & 0xFF) / 255.0
        let g = Double((hex >> 8) & 0xFF) / 255.0
        let b = Double(hex & 0xFF) / 255.0
        self.init(red: r, green: g, blue: b, opacity: opacity)
    }
}
