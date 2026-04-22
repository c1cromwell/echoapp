// Features/Profile/TrustBadge/EchoColorPalette+FirstRun.swift
// Glacial Interface color additions for first-run flow and trust-tier badges (v2.5.3)

import SwiftUI

extension Color.Echo {
    /// Tier 1–2 trust badge fill. Amber — "standard trust" without reading as a warning.
    static let trustYellow = Color(red: 0xF5/255, green: 0x9E/255, blue: 0x0B/255)

    /// Tier 5 IAL2+ outer ring. Teal distinguishes "identity verified" from "credential verified".
    static let trustPremium = Color(red: 0x0D/255, green: 0x94/255, blue: 0x88/255)

    /// Primary compose FAB color. Green = affirmative action, distinct from brand sky-blue.
    static let compose     = Color(red: 0x22/255, green: 0xC5/255, blue: 0x5E/255)
    static let composeDeep = Color(red: 0x16/255, green: 0xA3/255, blue: 0x4A/255)
}
