import SwiftUI

/// Contacts - Contacts List Screen
public struct ContactsListView: View {
    @State private var searchText = ""
    @State private var selectedFilter = "All"
    @State private var contacts: [ContactModel] = [
        ContactModel(id: "1", name: "John Doe", username: "johndoe", trustLevel: "Verified"),
        ContactModel(id: "2", name: "Jane Smith", username: "janesmith", trustLevel: "Trusted"),
        ContactModel(id: "3", name: "Alice Johnson", username: "alice_j", trustLevel: "Basic")
    ]
    
    let onSelectContact: (String) -> Void
    
    public init(onSelectContact: @escaping (String) -> Void = { _ in }) {
        self.onSelectContact = onSelectContact
    }
    
    var filteredContacts: [ContactModel] {
        var filtered = contacts
        
        if !searchText.isEmpty {
            filtered = filtered.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        
        if selectedFilter != "All" {
            filtered = filtered.filter { $0.trustLevel == selectedFilter }
        }
        
        return filtered.sorted { $0.name < $1.name }
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Contacts",
                    showBackButton: false,
                    trailingAction: {},
                    trailingIcon: Image(systemName: "person.badge.plus")
                )
                
                VStack(spacing: Spacing.lg.rawValue) {
                    // Search Bar
                    HStack {
                        Image(systemName: "magnifyingglass")
                            .font(.system(size: 14))
                            .foregroundColor(.echoGray500)
                        
                        TextField("Search contacts", text: $searchText)
                            .textFieldStyle(.roundedBorder)
                        
                        if !searchText.isEmpty {
                            Button(action: { searchText = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .font(.system(size: 14))
                                    .foregroundColor(.echoGray400)
                            }
                        }
                    }
                    .padding(Spacing.md.rawValue)
                    .background(Color.echoSurface)
                    .cornerRadius(12)
                    
                    // Filter Picker
                    Picker("Filter", selection: $selectedFilter) {
                        Text("All").tag("All")
                        Text("Inner Circle").tag("Inner Circle")
                        Text("Trusted").tag("Trusted")
                        Text("Acquaintance").tag("Acquaintance")
                    }
                    .pickerStyle(.segmented)
                    .padding(.horizontal, Spacing.md.rawValue)
                    
                    if filteredContacts.isEmpty {
                        VStack(spacing: Spacing.md.rawValue) {
                            Image(systemName: "person.slash")
                                .font(.system(size: 48))
                                .foregroundColor(.echoGray400)
                            
                            Text("No contacts found")
                                .typographyStyle(.h4, color: .echoGray600)
                            
                            Text("Add contacts to get started")
                                .typographyStyle(.body, color: .echoSecondaryText)
                        }
                        .frame(maxHeight: .infinity, alignment: .center)
                    } else {
                        List {
                            ForEach(filteredContacts) { contact in
                                ContactListItem(
                                    name: contact.name,
                                    username: contact.username,
                                    trustLevel: contact.trustLevel,
                                    onTap: { onSelectContact(contact.id) }
                                )
                                .listRowSeparator(.hidden)
                                .listRowInsets(.init())
                                .listRowBackground(Color.clear)
                                .padding(.vertical, Spacing.xs.rawValue)
                            }
                        }
                        .listStyle(.plain)
                        .scrollContentBackground(.hidden)
                    }
                }
                .echoSpacing(.lg)
            }
        }
    }
}

struct ContactModel: Identifiable {
    let id: String
    let name: String
    let username: String
    let trustLevel: String
}

// MARK: - Trust Dashboard Screen

public struct TrustDashboardView: View {
    @State private var trustScore = 65
    @State private var trustLevel = "Verified"
    @State private var showVerificationFlow = false
    
    public init() {}
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Trust Score",
                    showBackButton: false
                )
                
                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Trust Score Circle
                        TrustScoreView(score: trustScore, level: trustLevel)
                        
                        // Score Breakdown
                        VStack(spacing: Spacing.md.rawValue) {
                            Text("Score Breakdown")
                                .typographyStyle(.h4, color: .echoPrimaryText)
                                .frame(maxWidth: .infinity, alignment: .leading)
                            
                            ScoreBreakdownCard(label: "Identity", score: 25, maxScore: 30)
                            ScoreBreakdownCard(label: "Behavior", score: 18, maxScore: 25)
                            ScoreBreakdownCard(label: "Network", score: 15, maxScore: 25)
                            ScoreBreakdownCard(label: "Activity", score: 7, maxScore: 20)
                        }
                        
                        // Verification Status
                        VStack(spacing: Spacing.md.rawValue) {
                            Text("Verification Status")
                                .typographyStyle(.h4, color: .echoPrimaryText)
                                .frame(maxWidth: .infinity, alignment: .leading)
                            
                            SettingsListItem(
                                icon: Image(systemName: "checkmark.circle.fill"),
                                title: "Phone Number",
                                value: "Verified"
                            )
                            
                            SettingsListItem(
                                icon: Image(systemName: "exclamationmark.circle"),
                                title: "Identity Document",
                                value: "Not Verified"
                            )
                        }
                        
                        // Improve Score CTA
                        EchoButton(
                            "Verify Your Identity",
                            style: .primary,
                            size: .large,
                            icon: Image(systemName: "checkmark.shield"),
                            action: { showVerificationFlow = true }
                        )
                    }
                    .echoSpacing(.lg)
                }
            }
        }
    }
}

struct ScoreBreakdownCard: View {
    let label: String
    let score: Int
    let maxScore: Int
    
    var percentage: Double {
        Double(score) / Double(maxScore)
    }
    
    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
            HStack {
                Text(label)
                    .typographyStyle(.body, color: .echoPrimaryText)
                
                Spacer()
                
                Text("\(score)/\(maxScore)")
                    .typographyStyle(.caption, color: .echoGray500)
            }
            
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.echoGray200)
                    
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.echoPrimary)
                        .frame(width: geometry.size.width * percentage)
                }
            }
            .frame(height: 8)
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
    }
}

// MARK: - Preview

#if DEBUG
struct ContactsAndTrustScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            ContactsListView()
        }
    }
}
#endif
