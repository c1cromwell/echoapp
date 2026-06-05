#if os(iOS)
import SwiftUI
import Contacts

struct NewContactSheet: View {
    @Environment(\.dismiss) private var dismiss

    @State private var firstName = ""
    @State private var lastName = ""
    @State private var echoUsername = ""
    @State private var phoneNumber = ""
    @State private var showQRScanner = false
    @State private var qrCoordinator = QRContactAddCoordinator()
    @State private var selectedCountry = CountryDialCode.unitedStates
    @State private var showCountryPicker = false
    @State private var syncToPhone = true
    @State private var isSaving = false
    @State private var saveError: String?
    @State private var saveSuccess = false

    let onSaved: () -> Void

    private var trimmedUsername: String {
        echoUsername
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .replacingOccurrences(of: "@", with: "")
    }

    private var canSave: Bool {
        let hasName = !firstName.trimmingCharacters(in: .whitespaces).isEmpty
            || !lastName.trimmingCharacters(in: .whitespaces).isEmpty
        return hasName || trimmedUsername.count >= 2
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    nameFields
                    echoUsernameField
                    phoneFields
                    syncToggle
                    qrCodeButton

                    if let saveError {
                        Text(saveError)
                            .font(.system(size: 13))
                            .foregroundColor(.echoAlert)
                            .padding(.horizontal, Spacing.lg.rawValue)
                    }

                    if saveSuccess {
                        HStack(spacing: 6) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(.echoTrustGreen)
                            Text("Contact saved")
                                .font(.system(size: 14, weight: .medium))
                                .foregroundColor(.echoTrustGreen)
                        }
                        .padding(.top, 4)
                    }
                }
                .padding(.top, 20)
            }
            .background(Color.echoPaper)
            .navigationTitle("New Contact")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button { dismiss() } label: {
                        Image(systemName: "xmark")
                            .font(.system(size: 13, weight: .bold))
                            .foregroundColor(.echoInk55)
                            .frame(width: 30, height: 30)
                            .background(Color.echoPaperDim)
                            .clipShape(Circle())
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button { saveContact() } label: {
                        if isSaving {
                            ProgressView()
                        } else {
                            Image(systemName: "checkmark")
                                .font(.system(size: 13, weight: .bold))
                                .foregroundColor(canSave ? .white : .echoInk40)
                                .frame(width: 30, height: 30)
                                .background(canSave ? Color.echoSignal : Color.echoPaperDim)
                                .clipShape(Circle())
                        }
                    }
                    .disabled(!canSave || isSaving)
                }
            }
            .sheet(isPresented: $showCountryPicker) {
                CountryPickerSheet(selected: $selectedCountry)
            }
            .sheet(isPresented: $showQRScanner) {
                NavigationStack {
                    LiveQRCodeScannerView { raw in
                        showQRScanner = false
                        Task {
                            await qrCoordinator.handleScan(raw)
                            if qrCoordinator.resultIsError {
                                saveError = qrCoordinator.resultMessage
                            } else if qrCoordinator.resultMessage != nil {
                                saveSuccess = true
                                onSaved()
                                try? await Task.sleep(nanoseconds: 800_000_000)
                                dismiss()
                            }
                        }
                    }
                    .navigationTitle("Scan profile QR")
                    .navigationBarTitleDisplayMode(.inline)
                }
            }
        }
    }

    private var echoUsernameField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("ECHO username (optional)")
                .font(.system(size: 12, weight: .medium))
                .foregroundColor(.echoInk55)
                .padding(.horizontal, Spacing.lg.rawValue)
            HStack(spacing: 8) {
                Text("@")
                    .font(.system(size: 16, weight: .medium))
                    .foregroundColor(.echoInk55)
                TextField("username", text: $echoUsername)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .font(.system(size: 16))
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 14)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 14))
            .padding(.horizontal, Spacing.lg.rawValue)
        }
    }

    // MARK: - Name Fields

    private var nameFields: some View {
        VStack(spacing: 0) {
            TextField("First Name", text: $firstName)
                .font(.system(size: 16))
                .foregroundColor(.echoInk)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, 14)
            Divider().padding(.leading, Spacing.lg.rawValue)
            TextField("Last Name", text: $lastName)
                .font(.system(size: 16))
                .foregroundColor(.echoInk)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, 14)
        }
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, Spacing.lg.rawValue)
    }

    // MARK: - Phone Fields

    private var phoneFields: some View {
        VStack(spacing: 0) {
            Button { showCountryPicker = true } label: {
                HStack {
                    Text(selectedCountry.flag)
                        .font(.system(size: 20))
                    Text(selectedCountry.name)
                        .font(.system(size: 16))
                        .foregroundColor(.echoInk)
                    Spacer()
                    Image(systemName: "chevron.right")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundColor(.echoInk40)
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, 14)
            }
            .buttonStyle(.plain)

            Divider().padding(.leading, Spacing.lg.rawValue)

            HStack(spacing: 12) {
                Text(selectedCountry.dialCode)
                    .font(.system(size: 16, weight: .medium))
                    .foregroundColor(.echoInk)
                    .frame(minWidth: 32)
                TextField("000 000 0000", text: $phoneNumber)
                    .font(.system(size: 16))
                    .foregroundColor(.echoInk)
                    .keyboardType(.phonePad)
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 14)
        }
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, Spacing.lg.rawValue)
    }

    // MARK: - Sync Toggle

    private var syncToggle: some View {
        Toggle(isOn: $syncToPhone) {
            Text("Sync Contact to Phone")
                .font(.system(size: 16))
                .foregroundColor(.echoInk)
        }
        .tint(.echoSignal)
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, 12)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, Spacing.lg.rawValue)
    }

    // MARK: - QR Code

    private var qrCodeButton: some View {
        Button { showQRScanner = true } label: {
            HStack(spacing: 12) {
                Image(systemName: "qrcode")
                    .font(.system(size: 18))
                    .foregroundColor(.echoSignal)
                Text("Add via QR Code")
                    .font(.system(size: 16, weight: .medium))
                    .foregroundColor(.echoSignal)
                Spacer()
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 14)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 14))
        }
        .buttonStyle(.plain)
        .padding(.horizontal, Spacing.lg.rawValue)
    }

    // MARK: - Save

    private func saveContact() {
        let name = [firstName, lastName]
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
        guard !name.isEmpty || trimmedUsername.count >= 2 else { return }

        isSaving = true
        saveError = nil

        Task { @MainActor in
            defer { isSaving = false }

            if syncToPhone, !name.isEmpty {
                await saveToPhoneContacts(firstName: firstName, lastName: lastName, phone: fullPhoneNumber)
            }

            if trimmedUsername.count >= 2, await addEchoContactByUsername() {
                saveSuccess = true
            } else if saveError == nil {
                saveSuccess = true
            }

            guard saveError == nil else { return }
            onSaved()
            try? await Task.sleep(nanoseconds: 800_000_000)
            dismiss()
        }
    }

    private var fullPhoneNumber: String {
        let digits = phoneNumber.filter(\.isNumber)
        return digits.isEmpty ? "" : "\(selectedCountry.dialCode)\(digits)"
    }

    private func saveToPhoneContacts(firstName: String, lastName: String, phone: String) async {
        let store = CNContactStore()
        do {
            let granted = try await store.requestAccess(for: .contacts)
            guard granted else {
                saveError = "Contacts access denied. Enable in Settings."
                return
            }
            let contact = CNMutableContact()
            contact.givenName = firstName
            contact.familyName = lastName
            if !phone.isEmpty {
                contact.phoneNumbers = [CNLabeledValue(
                    label: CNLabelPhoneNumberMobile,
                    value: CNPhoneNumber(stringValue: phone)
                )]
            }
            let request = CNSaveRequest()
            request.add(contact, toContainerWithIdentifier: nil)
            try store.execute(request)
        } catch {
            saveError = "Could not save to phone contacts."
        }
    }

    /// Returns true when an ECHO contact was added via @username.
    private func addEchoContactByUsername() async -> Bool {
        let handle = trimmedUsername
        guard handle.count >= 2 else { return false }

        guard let client = DIContainer.shared.resolveAPIClient() else {
            saveError = "Sign in required to add ECHO contacts."
            return false
        }

        let social = ContactSocialAPIClient(apiClient: client)
        do {
            let results = try await social.searchUsername(handle)
            guard let first = results.first else {
                saveError = "No ECHO user found for @\(handle)."
                return false
            }
            _ = try await social.addContact(did: first.did, addedVia: "manual_add")
            _ = await ContactThreadHelper.upsertDirectThread(
                peerDID: first.did,
                displayName: "@\(first.username)"
            )
            return true
        } catch {
            saveError = error.localizedDescription
            return false
        }
    }
}

// MARK: - Country Dial Code

struct CountryDialCode: Identifiable, Hashable {
    let id: String
    let flag: String
    let name: String
    let dialCode: String

    static let unitedStates = CountryDialCode(id: "US", flag: "\u{1F1FA}\u{1F1F8}", name: "United States", code: "+1")
    static let all: [CountryDialCode] = [
        unitedStates,
        CountryDialCode(id: "CA", flag: "\u{1F1E8}\u{1F1E6}", name: "Canada", code: "+1"),
        CountryDialCode(id: "GB", flag: "\u{1F1EC}\u{1F1E7}", name: "United Kingdom", code: "+44"),
        CountryDialCode(id: "AU", flag: "\u{1F1E6}\u{1F1FA}", name: "Australia", code: "+61"),
        CountryDialCode(id: "DE", flag: "\u{1F1E9}\u{1F1EA}", name: "Germany", code: "+49"),
        CountryDialCode(id: "FR", flag: "\u{1F1EB}\u{1F1F7}", name: "France", code: "+33"),
        CountryDialCode(id: "JP", flag: "\u{1F1EF}\u{1F1F5}", name: "Japan", code: "+81"),
        CountryDialCode(id: "KR", flag: "\u{1F1F0}\u{1F1F7}", name: "South Korea", code: "+82"),
        CountryDialCode(id: "IN", flag: "\u{1F1EE}\u{1F1F3}", name: "India", code: "+91"),
        CountryDialCode(id: "BR", flag: "\u{1F1E7}\u{1F1F7}", name: "Brazil", code: "+55"),
        CountryDialCode(id: "MX", flag: "\u{1F1F2}\u{1F1FD}", name: "Mexico", code: "+52"),
        CountryDialCode(id: "NG", flag: "\u{1F1F3}\u{1F1EC}", name: "Nigeria", code: "+234"),
        CountryDialCode(id: "ZA", flag: "\u{1F1FF}\u{1F1E6}", name: "South Africa", code: "+27"),
        CountryDialCode(id: "AE", flag: "\u{1F1E6}\u{1F1EA}", name: "UAE", code: "+971"),
        CountryDialCode(id: "SG", flag: "\u{1F1F8}\u{1F1EC}", name: "Singapore", code: "+65"),
    ]

    init(id: String, flag: String, name: String, code: String) {
        self.id = id
        self.flag = flag
        self.name = name
        self.dialCode = code
    }
}

// MARK: - Country Picker

struct CountryPickerSheet: View {
    @Binding var selected: CountryDialCode
    @Environment(\.dismiss) private var dismiss
    @State private var search = ""

    private var filtered: [CountryDialCode] {
        if search.isEmpty { return CountryDialCode.all }
        return CountryDialCode.all.filter {
            $0.name.localizedCaseInsensitiveContains(search)
                || $0.dialCode.contains(search)
        }
    }

    var body: some View {
        NavigationStack {
            List(filtered) { country in
                Button {
                    selected = country
                    dismiss()
                } label: {
                    HStack(spacing: 12) {
                        Text(country.flag).font(.system(size: 22))
                        Text(country.name)
                            .font(.system(size: 16))
                            .foregroundColor(.echoInk)
                        Spacer()
                        Text(country.dialCode)
                            .font(.system(size: 14))
                            .foregroundColor(.echoInk55)
                        if country.id == selected.id {
                            Image(systemName: "checkmark")
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundColor(.echoSignal)
                        }
                    }
                }
                .buttonStyle(.plain)
            }
            .searchable(text: $search, prompt: "Search countries")
            .navigationTitle("Select Country")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }
}
#endif
