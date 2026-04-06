# Echo iOS App - Swift Implementation Specification

## Overview

This document provides a complete specification for implementing the Echo app frontend using Swift and SwiftUI for iOS. It covers architecture, design system integration, component implementation, and best practices.

---

## Table of Contents

1. [Project Setup](#1-project-setup)
2. [Architecture](#2-architecture)
3. [Design System Implementation](#3-design-system-implementation)
4. [Theme System](#4-theme-system)
5. [Core Components](#5-core-components)
6. [Screen Implementations](#6-screen-implementations)
7. [Navigation](#7-navigation)
8. [Data Layer](#8-data-layer)
9. [Security](#9-security)
10. [Testing](#10-testing)

---

## 1. Project Setup

### 1.1 Requirements

- **iOS Version**: 16.0+
- **Xcode**: 15.0+
- **Swift**: 5.9+
- **Architecture**: MVVM + Clean Architecture

### 1.2 Project Structure

```
EchoApp/
├── App/
│   ├── EchoApp.swift                 # App entry point
│   ├── AppDelegate.swift             # App lifecycle
│   └── SceneDelegate.swift           # Scene management
│
├── Core/
│   ├── DesignSystem/
│   │   ├── Colors/
│   │   │   ├── ColorPalette.swift    # Color definitions
│   │   │   ├── Theme.swift           # Theme protocol & implementations
│   │   │   └── ThemeManager.swift    # Theme switching logic
│   │   ├── Typography/
│   │   │   ├── FontStyles.swift      # Font definitions
│   │   │   └── TextStyles.swift      # Text style modifiers
│   │   ├── Spacing/
│   │   │   └── Spacing.swift         # Spacing constants
│   │   └── Components/
│   │       ├── Buttons/
│   │       ├── Inputs/
│   │       ├── Cards/
│   │       └── ...
│   │
│   ├── Extensions/
│   │   ├── View+Extensions.swift
│   │   ├── Color+Extensions.swift
│   │   └── ...
│   │
│   ├── Utilities/
│   │   ├── Haptics.swift
│   │   ├── KeychainManager.swift
│   │   └── BiometricAuth.swift
│   │
│   └── Networking/
│       ├── APIClient.swift
│       ├── Endpoints.swift
│       └── WebSocketManager.swift
│
├── Features/
│   ├── Auth/
│   │   ├── Views/
│   │   ├── ViewModels/
│   │   └── Models/
│   │
│   ├── Messages/
│   │   ├── Views/
│   │   │   ├── ConversationsListView.swift
│   │   │   ├── ChatView.swift
│   │   │   └── PinnedSectionView.swift
│   │   ├── ViewModels/
│   │   └── Models/
│   │
│   ├── Contacts/
│   ├── Trust/
│   └── Profile/
│
├── Resources/
│   ├── Assets.xcassets/
│   ├── Localizable.strings
│   └── Info.plist
│
└── Tests/
    ├── UnitTests/
    └── UITests/
```

### 1.3 Dependencies (Swift Package Manager)

```swift
// Package.swift dependencies
dependencies: [
    .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
    .package(url: "https://github.com/kean/Nuke.git", from: "12.0.0"),
    .package(url: "https://github.com/evgenyneu/keychain-swift.git", from: "20.0.0"),
    .package(url: "https://github.com/airbnb/lottie-ios.git", from: "4.3.0"),
    .package(url: "https://github.com/daltoniam/Starscream.git", from: "4.0.0"),
]
```

---

## 2. Architecture

### 2.1 MVVM + Clean Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PRESENTATION LAYER                        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │    Views    │◄───│  ViewModels │◄───│   States    │         │
│  │  (SwiftUI)  │    │ (Observable)│    │ (Published) │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DOMAIN LAYER                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │  Use Cases  │    │  Entities   │    │ Repositories │         │
│  │ (Interactors)│   │  (Models)   │    │ (Protocols) │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          DATA LAYER                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │ API Client  │    │  Database   │    │   Keychain  │         │
│  │  (Network)  │    │ (SwiftData) │    │  (Security) │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Design System Implementation

### 3.1 Color Extension

```swift
// Core/Extensions/Color+Extensions.swift

import SwiftUI

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3:
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6:
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8:
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (1, 1, 1, 0)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}
```

### 3.2 Color Palette Protocol

```swift
// Core/DesignSystem/Colors/ColorPalette.swift

import SwiftUI

protocol ColorPalette {
    // Primary
    var primary: Color { get }
    var primaryDark: Color { get }
    var primaryLight: Color { get }
    var primarySubtle: Color { get }
    
    // Accent
    var accent: Color { get }
    var accentDark: Color { get }
    var accentLight: Color { get }
    
    // Secondary
    var secondary: Color { get }
    var secondaryDark: Color { get }
    
    // Semantic
    var success: Color { get }
    var successLight: Color { get }
    var warning: Color { get }
    var warningLight: Color { get }
    var error: Color { get }
    var errorLight: Color { get }
    
    // Background
    var background: Color { get }
    var backgroundSecondary: Color { get }
    var surface: Color { get }
    var surfaceElevated: Color { get }
    
    // Text
    var textPrimary: Color { get }
    var textSecondary: Color { get }
    var textTertiary: Color { get }
    var textInverse: Color { get }
    var textOnPrimary: Color { get }
    
    // Border
    var border: Color { get }
    var borderLight: Color { get }
    var borderFocus: Color { get }
    
    // Gradients
    var primaryGradient: LinearGradient { get }
    var secondaryGradient: LinearGradient { get }
    var warmGradient: LinearGradient { get }
}
```

### 3.3 Theme Implementations

```swift
// Core/DesignSystem/Colors/Themes.swift

import SwiftUI

// MARK: - Electric Violet Theme (Default)
struct ElectricVioletPalette: ColorPalette {
    let primary = Color(hex: "7C3AED")
    let primaryDark = Color(hex: "6D28D9")
    let primaryLight = Color(hex: "A78BFA")
    let primarySubtle = Color(hex: "EDE9FE")
    
    let accent = Color(hex: "06B6D4")
    let accentDark = Color(hex: "0891B2")
    let accentLight = Color(hex: "67E8F9")
    
    let secondary = Color(hex: "F43F5E")
    let secondaryDark = Color(hex: "E11D48")
    
    let success = Color(hex: "10B981")
    let successLight = Color(hex: "D1FAE5")
    let warning = Color(hex: "F59E0B")
    let warningLight = Color(hex: "FEF3C7")
    let error = Color(hex: "EF4444")
    let errorLight = Color(hex: "FEE2E2")
    
    let background = Color(hex: "FFFFFF")
    let backgroundSecondary = Color(hex: "FAFAFA")
    let surface = Color(hex: "F4F4F5")
    let surfaceElevated = Color(hex: "FFFFFF")
    
    let textPrimary = Color(hex: "09090B")
    let textSecondary = Color(hex: "52525B")
    let textTertiary = Color(hex: "A1A1AA")
    let textInverse = Color(hex: "FAFAFA")
    let textOnPrimary = Color(hex: "FFFFFF")
    
    let border = Color(hex: "E4E4E7")
    let borderLight = Color(hex: "F4F4F5")
    let borderFocus = Color(hex: "7C3AED")
    
    var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "7C3AED"), Color(hex: "06B6D4")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "06B6D4"), Color(hex: "10B981")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "F43F5E"), Color(hex: "F59E0B")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}

// MARK: - Teal + Warm Neutral Theme
struct TealWarmPalette: ColorPalette {
    let primary = Color(hex: "0D9488")
    let primaryDark = Color(hex: "0F766E")
    let primaryLight = Color(hex: "2DD4BF")
    let primarySubtle = Color(hex: "CCFBF1")
    
    let accent = Color(hex: "F59E0B")
    let accentDark = Color(hex: "D97706")
    let accentLight = Color(hex: "FBBF24")
    
    let secondary = Color(hex: "C4A77D")
    let secondaryDark = Color(hex: "A68B5B")
    
    let success = Color(hex: "059669")
    let successLight = Color(hex: "D1FAE5")
    let warning = Color(hex: "EA580C")
    let warningLight = Color(hex: "FFEDD5")
    let error = Color(hex: "DC2626")
    let errorLight = Color(hex: "FEE2E2")
    
    // Warm backgrounds (avoids pure white)
    let background = Color(hex: "FFFCF7")
    let backgroundSecondary = Color(hex: "FBF8F3")
    let surface = Color(hex: "F5F0E8")
    let surfaceElevated = Color(hex: "FFFFFF")
    
    // Warm charcoal text (avoids pure black)
    let textPrimary = Color(hex: "1C1917")
    let textSecondary = Color(hex: "57534E")
    let textTertiary = Color(hex: "78716C")
    let textInverse = Color(hex: "FFFCF7")
    let textOnPrimary = Color(hex: "FFFFFF")
    
    let border = Color(hex: "E7E5E4")
    let borderLight = Color(hex: "F5F5F4")
    let borderFocus = Color(hex: "0D9488")
    
    var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "0D9488"), Color(hex: "2DD4BF")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "C4A77D"), Color(hex: "D4C8B8")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "F59E0B"), Color(hex: "EA580C")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}

// MARK: - Icy Minimal Theme
struct IcyMinimalPalette: ColorPalette {
    let primary = Color(hex: "64748B")
    let primaryDark = Color(hex: "475569")
    let primaryLight = Color(hex: "94A3B8")
    let primarySubtle = Color(hex: "F1F5F9")
    
    let accent = Color(hex: "0EA5E9")
    let accentDark = Color(hex: "0284C7")
    let accentLight = Color(hex: "7DD3FC")
    
    let secondary = Color(hex: "78716C")
    let secondaryDark = Color(hex: "57534E")
    
    let success = Color(hex: "10B981")
    let successLight = Color(hex: "D1FAE5")
    let warning = Color(hex: "F59E0B")
    let warningLight = Color(hex: "FEF3C7")
    let error = Color(hex: "EF4444")
    let errorLight = Color(hex: "FEE2E2")
    
    // Cool backgrounds
    let background = Color(hex: "F8FAFC")
    let backgroundSecondary = Color(hex: "F1F5F9")
    let surface = Color(hex: "E2E8F0")
    let surfaceElevated = Color(hex: "FFFFFF")
    
    // Cool charcoal text
    let textPrimary = Color(hex: "0F172A")
    let textSecondary = Color(hex: "475569")
    let textTertiary = Color(hex: "64748B")
    let textInverse = Color(hex: "F8FAFC")
    let textOnPrimary = Color(hex: "FFFFFF")
    
    let border = Color(hex: "E2E8F0")
    let borderLight = Color(hex: "F1F5F9")
    let borderFocus = Color(hex: "0EA5E9")
    
    var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "0EA5E9"), Color(hex: "7DD3FC")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "78716C"), Color(hex: "A8A29E")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
    
    var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: "F59E0B"), Color(hex: "FBBF24")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}
```

---

## 4. Theme System

### 4.1 Theme Manager

```swift
// Core/DesignSystem/Colors/ThemeManager.swift

import SwiftUI
import Combine

enum ThemeType: String, CaseIterable, Identifiable {
    case electricViolet = "electric_violet"
    case tealWarm = "teal_warm"
    case icyMinimal = "icy_minimal"
    
    var id: String { rawValue }
    
    var name: String {
        switch self {
        case .electricViolet: return "Electric Violet"
        case .tealWarm: return "Teal + Warm Neutral"
        case .icyMinimal: return "Icy Minimal"
        }
    }
    
    var description: String {
        switch self {
        case .electricViolet: return "Bold & vibrant"
        case .tealWarm: return "Modern & professional"
        case .icyMinimal: return "Fresh & techy"
        }
    }
    
    var palette: ColorPalette {
        switch self {
        case .electricViolet: return ElectricVioletPalette()
        case .tealWarm: return TealWarmPalette()
        case .icyMinimal: return IcyMinimalPalette()
        }
    }
}

struct Theme {
    let type: ThemeType
    let colors: ColorPalette
    
    init(type: ThemeType) {
        self.type = type
        self.colors = type.palette
    }
}

class ThemeManager: ObservableObject {
    @Published var currentTheme: Theme
    
    private let userDefaults = UserDefaults.standard
    private let themeKey = "selected_theme"
    
    init() {
        let savedTheme = userDefaults.string(forKey: themeKey) ?? ThemeType.electricViolet.rawValue
        let themeType = ThemeType(rawValue: savedTheme) ?? .electricViolet
        self.currentTheme = Theme(type: themeType)
    }
    
    func setTheme(_ type: ThemeType) {
        currentTheme = Theme(type: type)
        userDefaults.set(type.rawValue, forKey: themeKey)
    }
}

// MARK: - Environment Key
struct ThemeKey: EnvironmentKey {
    static let defaultValue: Theme = Theme(type: .electricViolet)
}

extension EnvironmentValues {
    var theme: Theme {
        get { self[ThemeKey.self] }
        set { self[ThemeKey.self] = newValue }
    }
}

// MARK: - View Modifier
extension View {
    func themed(_ themeManager: ThemeManager) -> some View {
        self.environment(\.theme, themeManager.currentTheme)
    }
}
```

### 4.2 Theme Usage in Views

```swift
// Example view using theme
struct ThemedButton: View {
    @Environment(\.theme) var theme
    let title: String
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(theme.colors.textOnPrimary)
                .frame(maxWidth: .infinity)
                .frame(height: 56)
                .background(theme.colors.primaryGradient)
                .cornerRadius(12)
        }
    }
}
```

---

## 5. Core Components

### 5.1 Pinned Section Component

The Pinned Section displays horizontally scrollable pinned contacts and groups below the persona header.

```swift
// Features/Messages/Views/PinnedSectionView.swift

import SwiftUI

// MARK: - Pinned Item Model
struct PinnedItem: Identifiable {
    let id: String
    let type: PinnedItemType
    let name: String
    let avatar: String?
    let initials: String
    let gradientIndex: Int
    var isOnline: Bool = false
    var unreadCount: Int = 0
    var members: [GroupMember]? = nil
    
    enum PinnedItemType {
        case contact
        case group
    }
}

struct GroupMember {
    let initials: String
    let color: Color
}

// MARK: - Pinned Avatar View
struct PinnedAvatarView: View {
    @Environment(\.theme) var theme
    let item: PinnedItem
    let size: CGFloat
    
    init(item: PinnedItem, size: CGFloat = 56) {
        self.item = item
        self.size = size
    }
    
    private var gradients: [[Color]] {
        [
            [Color(hex: "7C3AED"), Color(hex: "06B6D4")],
            [Color(hex: "06B6D4"), Color(hex: "10B981")],
            [Color(hex: "F43F5E"), Color(hex: "F59E0B")]
        ]
    }
    
    var body: some View {
        ZStack {
            if item.type == .group, let members = item.members {
                GroupAvatarView(members: members, size: size)
            } else {
                ContactAvatarView(
                    initials: item.initials,
                    gradient: gradients[item.gradientIndex % gradients.count],
                    size: size
                )
            }
            
            // Online indicator
            if item.isOnline && item.type == .contact {
                OnlineIndicator()
                    .position(x: size - 7, y: size - 7)
            }
            
            // Unread badge
            if item.unreadCount > 0 {
                UnreadBadge(count: item.unreadCount)
                    .position(x: size - 4, y: 4)
            }
        }
        .frame(width: size, height: size)
    }
}

// MARK: - Contact Avatar
struct ContactAvatarView: View {
    let initials: String
    let gradient: [Color]
    let size: CGFloat
    
    var body: some View {
        ZStack {
            Circle()
                .fill(
                    LinearGradient(
                        colors: gradient,
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .frame(width: size, height: size)
                .shadow(color: .black.opacity(0.15), radius: 8, x: 0, y: 4)
            
            Text(initials)
                .font(.system(size: size * 0.4, weight: .semibold))
                .foregroundColor(.white)
        }
    }
}

// MARK: - Group Avatar (2x2 Grid)
struct GroupAvatarView: View {
    @Environment(\.theme) var theme
    let members: [GroupMember]
    let size: CGFloat
    
    private let spacing: CGFloat = 2
    private let padding: CGFloat = 4
    
    private var faceSize: CGFloat {
        (size - padding * 2 - spacing) / 2
    }
    
    var body: some View {
        ZStack {
            Circle()
                .fill(theme.colors.surface)
                .frame(width: size, height: size)
                .overlay(
                    Circle()
                        .stroke(theme.colors.border, lineWidth: 2)
                )
            
            LazyVGrid(
                columns: [
                    GridItem(.fixed(faceSize), spacing: spacing),
                    GridItem(.fixed(faceSize), spacing: spacing)
                ],
                spacing: spacing
            ) {
                ForEach(0..<4, id: \.self) { index in
                    if index < members.count {
                        let member = members[index]
                        let displayText = index == 3 && members.count > 4
                            ? "+\(members.count - 3)"
                            : member.initials
                        
                        RoundedRectangle(cornerRadius: faceSize * 0.5)
                            .fill(member.color)
                            .frame(width: faceSize, height: faceSize)
                            .overlay(
                                Text(displayText)
                                    .font(.system(size: 10, weight: .semibold))
                                    .foregroundColor(.white)
                            )
                    }
                }
            }
            .padding(padding)
        }
    }
}

// MARK: - Online Indicator
struct OnlineIndicator: View {
    var body: some View {
        Circle()
            .fill(Color(hex: "10B981"))
            .frame(width: 14, height: 14)
            .overlay(
                Circle()
                    .stroke(Color.white, lineWidth: 2.5)
            )
    }
}

// MARK: - Unread Badge
struct UnreadBadge: View {
    let count: Int
    
    private var displayText: String {
        count > 99 ? "99+" : "\(count)"
    }
    
    var body: some View {
        Text(displayText)
            .font(.system(size: 10, weight: .bold))
            .foregroundColor(.white)
            .padding(.horizontal, 6)
            .frame(minWidth: 20, minHeight: 20)
            .background(Color(hex: "F43F5E"))
            .clipShape(Capsule())
            .overlay(
                Capsule()
                    .stroke(Color.white, lineWidth: 2)
            )
    }
}

// MARK: - Pinned Item Card
struct PinnedItemCard: View {
    @Environment(\.theme) var theme
    let item: PinnedItem
    let onTap: () -> Void
    
    var body: some View {
        Button(action: onTap) {
            VStack(spacing: 8) {
                PinnedAvatarView(item: item, size: 56)
                
                Text(item.name)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(theme.colors.textSecondary)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .frame(width: 72)
            }
        }
        .buttonStyle(ScaleButtonStyle())
    }
}

// MARK: - Scale Button Style
struct ScaleButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
            .animation(.spring(response: 0.3, dampingFraction: 0.6), value: configuration.isPressed)
    }
}

// MARK: - Pinned Section View
struct PinnedSectionView: View {
    @Environment(\.theme) var theme
    let items: [PinnedItem]
    let onItemTap: (PinnedItem) -> Void
    let onEditTap: () -> Void
    let maxItems: Int
    
    init(
        items: [PinnedItem],
        maxItems: Int = 9,
        onItemTap: @escaping (PinnedItem) -> Void,
        onEditTap: @escaping () -> Void
    ) {
        self.items = items
        self.maxItems = maxItems
        self.onItemTap = onItemTap
        self.onEditTap = onEditTap
    }
    
    var body: some View {
        if items.isEmpty {
            EmptyView()
        } else {
            VStack(spacing: 12) {
                // Header
                HStack {
                    HStack(spacing: 6) {
                        Text("📌")
                            .font(.system(size: 12))
                        
                        Text("PINNED")
                            .font(.system(size: 13, weight: .bold))
                            .foregroundColor(theme.colors.textTertiary)
                            .tracking(1)
                    }
                    
                    Spacer()
                    
                    Button(action: onEditTap) {
                        Text("Edit")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundColor(theme.colors.primary)
                    }
                }
                .padding(.horizontal, 16)
                
                // Horizontal scroll of pinned items
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(items.prefix(maxItems)) { item in
                            PinnedItemCard(item: item) {
                                onItemTap(item)
                            }
                        }
                    }
                    .padding(.horizontal, 16)
                }
            }
            .padding(.vertical, 12)
        }
    }
}
```

### 5.2 Usage in Messages Screen

```swift
// Features/Messages/Views/ConversationsListView.swift

import SwiftUI

struct ConversationsListView: View {
    @Environment(\.theme) var theme
    @StateObject private var viewModel = ConversationsViewModel()
    
    var body: some View {
        VStack(spacing: 0) {
            // Persona Header
            PersonaHeaderView()
            
            // Search Bar
            SearchBarView(text: $viewModel.searchText)
                .padding(.horizontal, 16)
                .padding(.bottom, 16)
            
            // Pinned Section
            PinnedSectionView(
                items: viewModel.pinnedItems,
                onItemTap: { item in
                    viewModel.openConversation(for: item)
                },
                onEditTap: {
                    viewModel.showEditPinned = true
                }
            )
            
            // Section Divider
            SectionDivider(title: "ALL MESSAGES")
            
            // Conversation List
            ScrollView {
                LazyVStack(spacing: 0) {
                    ForEach(viewModel.conversations) { conversation in
                        ConversationRowView(conversation: conversation)
                    }
                }
            }
        }
        .background(theme.colors.background)
    }
}

// MARK: - Section Divider
struct SectionDivider: View {
    @Environment(\.theme) var theme
    let title: String
    
    var body: some View {
        HStack(spacing: 12) {
            Rectangle()
                .fill(theme.colors.border)
                .frame(height: 1)
            
            Text(title)
                .font(.system(size: 12, weight: .semibold))
                .foregroundColor(theme.colors.textTertiary)
                .tracking(1)
            
            Rectangle()
                .fill(theme.colors.border)
                .frame(height: 1)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
    }
}
```

### 5.3 Typography

```swift
// Core/DesignSystem/Typography/FontStyles.swift

import SwiftUI

enum EchoFonts {
    static let display = Font.system(size: 36, weight: .bold)
    static let h1 = Font.system(size: 28, weight: .bold)
    static let h2 = Font.system(size: 24, weight: .semibold)
    static let h3 = Font.system(size: 20, weight: .semibold)
    static let h4 = Font.system(size: 17, weight: .semibold)
    
    static let bodyLarge = Font.system(size: 17, weight: .regular)
    static let body = Font.system(size: 16, weight: .regular)
    static let bodySmall = Font.system(size: 14, weight: .regular)
    
    static let caption = Font.system(size: 13, weight: .regular)
    static let tiny = Font.system(size: 12, weight: .medium)
    static let overline = Font.system(size: 11, weight: .semibold)
}
```

### 5.4 Spacing

```swift
// Core/DesignSystem/Spacing/Spacing.swift

import SwiftUI

enum Spacing {
    static let none: CGFloat = 0
    static let xxs: CGFloat = 4
    static let xs: CGFloat = 8
    static let sm: CGFloat = 12
    static let md: CGFloat = 16
    static let lg: CGFloat = 20
    static let xl: CGFloat = 24
    static let xxl: CGFloat = 32
    static let xxxl: CGFloat = 40
    
    static let screenHorizontal: CGFloat = 24
    static let screenVertical: CGFloat = 24
    static let cardPadding: CGFloat = 16
}

enum CornerRadius {
    static let sm: CGFloat = 4
    static let md: CGFloat = 8
    static let lg: CGFloat = 12
    static let xl: CGFloat = 16
    static let xxl: CGFloat = 24
    static let full: CGFloat = 9999
}
```

---

## 6. Screen Implementations

### 6.1 Feature Checklist

| Screen | Status | Priority |
|--------|--------|----------|
| Welcome | 🟢 Spec Ready | P0 |
| Phone Entry | 🟢 Spec Ready | P0 |
| OTP Verification | 🟢 Spec Ready | P0 |
| Passkey Setup | 🟢 Spec Ready | P0 |
| Persona Creation | 🟢 Spec Ready | P0 |
| Conversations List | 🟢 Spec Ready | P0 |
| Chat View | 🟢 Spec Ready | P0 |
| Contact Profile | 🟢 Spec Ready | P1 |
| Trust Dashboard | 🟢 Spec Ready | P1 |
| Profile Settings | 🟢 Spec Ready | P1 |
| Theme Picker | 🟢 Spec Ready | P2 |

---

## 7. Navigation

```swift
// App/Navigation/AppRouter.swift

import SwiftUI

enum Route: Hashable {
    case welcome
    case phoneEntry
    case otpVerification(phoneNumber: String)
    case passkeySetup
    case personaCreation
    case home
    case chat(conversationId: String)
    case contactProfile(contactId: String)
    case settings
    case trustDashboard
}

class AppRouter: ObservableObject {
    @Published var path = NavigationPath()
    
    func navigate(to route: Route) {
        path.append(route)
    }
    
    func pop() {
        path.removeLast()
    }
    
    func popToRoot() {
        path.removeLast(path.count)
    }
}
```

---

## 8. Data Layer

### 8.1 API Client

```swift
// Core/Networking/APIClient.swift

import Foundation

protocol APIClientProtocol {
    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T
}

class APIClient: APIClientProtocol {
    private let baseURL: URL
    private let session: URLSession
    
    init(baseURL: URL = URL(string: "https://api.echo.app")!) {
        self.baseURL = baseURL
        self.session = URLSession.shared
    }
    
    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T {
        let request = try endpoint.urlRequest(baseURL: baseURL)
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              200...299 ~= httpResponse.statusCode else {
            throw APIError.invalidResponse
        }
        
        return try JSONDecoder().decode(T.self, from: data)
    }
}

enum APIError: Error {
    case invalidResponse
    case decodingError
    case networkError
}
```

---

## 9. Security

### 9.1 Keychain Manager

```swift
// Core/Utilities/KeychainManager.swift

import Security
import Foundation

protocol KeychainManagerProtocol {
    func save(_ data: Data, for key: String) throws
    func load(for key: String) throws -> Data?
    func delete(for key: String) throws
}

class KeychainManager: KeychainManagerProtocol {
    func save(_ data: Data, for key: String) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)
        
        guard status == errSecSuccess else {
            throw KeychainError.saveFailed
        }
    }
    
    func load(for key: String) throws -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        guard status == errSecSuccess else {
            return nil
        }
        
        return result as? Data
    }
    
    func delete(for key: String) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key
        ]
        
        SecItemDelete(query as CFDictionary)
    }
}

enum KeychainError: Error {
    case saveFailed
    case loadFailed
}
```

### 9.2 Biometric Authentication

```swift
// Core/Utilities/BiometricAuth.swift

import LocalAuthentication

protocol BiometricAuthProtocol {
    func authenticate() async throws -> Bool
    var biometryType: LABiometryType { get }
}

class BiometricAuth: BiometricAuthProtocol {
    private let context = LAContext()
    
    var biometryType: LABiometryType {
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil)
        return context.biometryType
    }
    
    func authenticate() async throws -> Bool {
        try await withCheckedThrowingContinuation { continuation in
            context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: "Authenticate to access Echo"
            ) { success, error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume(returning: success)
                }
            }
        }
    }
}
```

---

## 10. Testing

### 10.1 Unit Test Example

```swift
// Tests/UnitTests/ThemeManagerTests.swift

import XCTest
@testable import EchoApp

final class ThemeManagerTests: XCTestCase {
    var sut: ThemeManager!
    
    override func setUp() {
        super.setUp()
        sut = ThemeManager()
    }
    
    override func tearDown() {
        sut = nil
        super.tearDown()
    }
    
    func testDefaultThemeIsElectricViolet() {
        XCTAssertEqual(sut.currentTheme.type, .electricViolet)
    }
    
    func testSetThemeUpdatesCurrent() {
        sut.setTheme(.tealWarm)
        XCTAssertEqual(sut.currentTheme.type, .tealWarm)
    }
    
    func testThemePersistsAcrossInstances() {
        sut.setTheme(.icyMinimal)
        
        let newManager = ThemeManager()
        XCTAssertEqual(newManager.currentTheme.type, .icyMinimal)
    }
}
```

### 10.2 UI Test Example

```swift
// Tests/UITests/PinnedSectionUITests.swift

import XCTest

final class PinnedSectionUITests: XCTestCase {
    var app: XCUIApplication!
    
    override func setUp() {
        super.setUp()
        continueAfterFailure = false
        app = XCUIApplication()
        app.launch()
    }
    
    func testPinnedSectionDisplaysItems() {
        let pinnedSection = app.scrollViews["PinnedSection"]
        XCTAssertTrue(pinnedSection.exists)
        
        let firstPinnedItem = app.buttons["PinnedItem_0"]
        XCTAssertTrue(firstPinnedItem.exists)
    }
    
    func testTappingEditShowsEditSheet() {
        let editButton = app.buttons["EditPinned"]
        editButton.tap()
        
        let editSheet = app.sheets["EditPinnedSheet"]
        XCTAssertTrue(editSheet.exists)
    }
}
```

---

## Summary

This specification covers the complete iOS Swift implementation for the Echo app, including:

- ✅ Project structure and architecture (MVVM + Clean)
- ✅ Three color themes (Electric Violet, Teal Warm, Icy Minimal)
- ✅ Theme switching system with persistence
- ✅ Core components including Pinned Section
- ✅ Typography and spacing systems
- ✅ Navigation architecture
- ✅ Security implementations (Keychain, Biometrics)
- ✅ Testing examples

---

*Last Updated: February 2025*
*iOS Target: 16.0+*
*Swift Version: 5.9+*
