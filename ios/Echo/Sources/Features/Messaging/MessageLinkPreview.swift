import Foundation

/// Privacy-safe link helpers: extract URLs locally; optional client-side title fetch (no Echo server unfurl).
enum MessageLinkPreview {
    struct Preview: Equatable, Sendable, Identifiable {
        var id: String { url.absoluteString }
        let url: URL
        let domain: String
        var title: String?
    }

    static func firstURL(in text: String) -> URL? {
        guard let detector = try? NSDataDetector(types: NSTextCheckingResult.CheckingType.link.rawValue) else {
            return nil
        }
        let range = NSRange(text.startIndex..<text.endIndex, in: text)
        guard let match = detector.firstMatch(in: text, options: [], range: range),
              let url = match.url else {
            return nil
        }
        return url
    }

    static func domain(for url: URL) -> String {
        url.host?.replacingOccurrences(of: "www.", with: "") ?? url.absoluteString
    }

    static func previewStub(from text: String) -> Preview? {
        guard let url = firstURL(in: text) else { return nil }
        return Preview(url: url, domain: domain(for: url), title: nil)
    }

    private static let receivedKey = "echo.link_previews.received"

    /// Whether to fetch previews for links other people send. Off unless the user turns it on.
    static var fetchForReceivedMessages: Bool {
        get { UserDefaults.standard.bool(forKey: receivedKey) }
        set { UserDefaults.standard.set(newValue, forKey: receivedKey) }
    }

    /// Fetching a preview contacts the linked site from this device, which tells that site the
    /// device's network address and that the message was seen. For a link the user typed, the
    /// user chose the site. For a link someone sent, the sender chose it, so a sender could use a
    /// link to learn where a recipient is. Received links are not fetched unless the user opts in.
    static func shouldFetch(isOwnMessage: Bool, receivedOptIn: Bool = fetchForReceivedMessages) -> Bool {
        isOwnMessage || receivedOptIn
    }

    private static func makeSession() -> URLSession {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.httpCookieStorage = nil
        configuration.httpShouldSetCookies = false
        #if os(iOS)
        TransportProxySettings.apply(to: configuration)
        #endif
        return URLSession(configuration: configuration)
    }

    /// Best-effort HTML `<title>` fetch on-device. Call only when `shouldFetch` allows it.
    static func fetchTitle(for url: URL) async -> String? {
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.timeoutInterval = 8
        request.setValue("text/html", forHTTPHeaderField: "Accept")
        do {
            // Its own session: no cookies kept between sites, and the proxy setting applies.
            let session = makeSession()
            defer { session.finishTasksAndInvalidate() }
            let (data, response) = try await session.data(for: request)
            guard let http = response as? HTTPURLResponse,
                  (200...299).contains(http.statusCode),
                  let html = String(data: data.prefix(64_000), encoding: .utf8)
                    ?? String(data: data.prefix(64_000), encoding: .isoLatin1) else {
                return nil
            }
            return extractTitle(from: html)
        } catch {
            return nil
        }
    }

    static func extractTitle(from html: String) -> String? {
        guard let regex = try? NSRegularExpression(
            pattern: #"<title[^>]*>(.*?)</title>"#,
            options: [.caseInsensitive, .dotMatchesLineSeparators]
        ) else { return nil }
        let range = NSRange(html.startIndex..<html.endIndex, in: html)
        guard let match = regex.firstMatch(in: html, options: [], range: range),
              match.numberOfRanges > 1,
              let titleRange = Range(match.range(at: 1), in: html) else {
            return nil
        }
        let raw = String(html[titleRange])
            .replacingOccurrences(of: "\n", with: " ")
            .trimmingCharacters(in: .whitespacesAndNewlines)
        guard !raw.isEmpty else { return nil }
        return String(raw.prefix(120))
    }
}

#if os(iOS)
import SwiftUI

/// Compact domain chip under a message; expands title after a one-shot client fetch.
struct ChatInlineLinkPreview: View {
    let text: String
    let alignTrailing: Bool
    /// True for a message this user sent. A received link shows its domain only (see `shouldFetch`).
    var isOwnMessage: Bool = false

    @State private var preview: MessageLinkPreview.Preview?
    @State private var loading = false

    var body: some View {
        Group {
            if let preview {
                Button {
                    UIApplication.shared.open(preview.url)
                } label: {
                    HStack(spacing: 8) {
                        Image(systemName: "link")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(Color.echoSignal)
                        VStack(alignment: .leading, spacing: 2) {
                            if let title = preview.title, !title.isEmpty {
                                Text(title)
                                    .font(.system(size: 12, weight: .semibold))
                                    .foregroundStyle(Color.echoInk)
                                    .lineLimit(2)
                            } else if loading {
                                Text("Loading preview…")
                                    .font(.system(size: 11))
                                    .foregroundStyle(Color.echoInk40)
                            }
                            Text(preview.domain)
                                .font(.system(size: 11, weight: .medium))
                                .foregroundStyle(Color.echoInk55)
                                .lineLimit(1)
                        }
                        Spacer(minLength: 0)
                    }
                    .padding(.horizontal, 10)
                    .padding(.vertical, 8)
                    .background(Color.echoPaperDim)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color.echoHair, lineWidth: 1)
                    )
                }
                .buttonStyle(.plain)
                .frame(maxWidth: 280, alignment: alignTrailing ? .trailing : .leading)
                .task(id: preview.url.absoluteString) {
                    guard MessageLinkPreview.shouldFetch(isOwnMessage: isOwnMessage) else { return }
                    guard preview.title == nil, !loading else { return }
                    loading = true
                    let title = await MessageLinkPreview.fetchTitle(for: preview.url)
                    loading = false
                    if let title {
                        self.preview = MessageLinkPreview.Preview(
                            url: preview.url,
                            domain: preview.domain,
                            title: title
                        )
                    }
                }
            }
        }
        .onAppear {
            if preview == nil {
                preview = MessageLinkPreview.previewStub(from: text)
            }
        }
    }
}
#endif
