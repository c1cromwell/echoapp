# Echo Bar

**Status: steps 1–4 started; iPhone window uses the shared thread (8 Nov 2026).** Allow and revoke are passkey-signed on `POST /v3/contacts/bar` and `POST /v3/contacts/bar/revoke`. The window is Settings → Echo Bar on iPhone, the same SwiftUI column on Mac, and `/bar` on the web. On iPhone the transcript is `ChatDetailViewModel.messages` on the existing sealed send path. Email lookup is still step 5. Both simulators launch Echo without a crash. A sealed message between two signed-in accounts has not been confirmed, and `localhost:8000` was down for that launch. The web page does not sign with the device passkey yet, so Allow there stays in the browser until that signature exists.

Echo Bar is a second door into the same threads. It is one window and a composer fixed to the bottom, in the shape of the Claude and ChatGPT message bar. The people in that window are friends you have allowed. On a phone those friends come from the device contact list. On a Mac or in the browser you reach them by `@username` or by an email they have published as a reach address.

The full app stays as it is: `MainTabView` (Messages, Contacts, Rewards, Profile), frozen first-run (`FirstRunCoordinator`, `GlacialLoginScreen`), and `did:key` from the device passkey (ADR-0001). The bar does not add a tab and does not add an email account.

Visual layout: open the Echo Bar canvas beside this chat.

---

## The window

One column. On Mac and web it is a window about 420 points wide. On the phone the same column fills the screen. There is no tab bar.

| Region | What it shows |
|--------|----------------|
| Header | The friend’s name, the word Allowed, and a control that switches people. |
| Transcript | Their lines on the leading side, yours on the trailing side. Text only in the first cut. |
| Composer bar | A single rounded field pinned to the bottom. Placeholder is `Message {name}`. Enter sends on Mac and web. The send control sits on the trailing edge of the field. |

When nobody is selected, the transcript area is the allow-list and the same bar reads `Name, @username, or email`.

The first cut leaves out attachments, GIFs, polls, payments, calls, scheduling, channels, and groups. Those stay in the full app. The bar reuses the existing sealed send path (`ChatDetailViewModel` / `sealed_text`) once a thread id exists. It does not invent a second relay.

## Who you can message

An allow-list sits on the contact you already have. Discovery finds a candidate. Allow is a separate action. A match is not enough.

| Surface | How you add someone | What already exists |
|---------|---------------------|---------------------|
| iPhone | Add from contacts. Echo runs the current phone discovery (`ContactDiscoveryService`, OPRF-PSI). You then tap Allow on a match. | Device contacts are read locally. The server sees blinded phone numbers, not the address book. |
| Mac and web | Type `@username`, or an email that person published. | `@username` already resolves to a `did:key`. Email lookup is the same blinded shape as phone discovery, and it is not built yet. |

Rules:

- You can open a bar thread only with someone you allowed.
- They always receive the message in the full Echo app.
- They see it inside their own bar only when they have allowed you.
- Revoke removes them from the bar. The thread remains in the full app.
- Sign-in on every surface is the device passkey. Email and the phone number are reach addresses, the same way `@username` is. They are not accounts, and they do not appear in logs (T0–T7).

The Comply portal under `web/` is a different product. The bar gets its own page and does not use `PortalShell`.

## Build order

Ship in this order. Each step is done when its test or human check passes. Do not start the next until the previous one is true.

1. **Allow-list.** Add `bar_allowed` on the contact record. Allow and revoke are passkey-signed. A unit test covers allow, revoke, and “discover does not allow.”
2. **iPhone window.** A bar scene, not a `MainTabView` tab. Header, transcript, composer. Add from contacts uses `ContactDiscoveryService`, then Allow. Human check: two simulators, one allowed friend, one sealed text arrives.
3. **Mac window.** The package already lists macOS 14. Same column. The bar accepts `@username` first. Email field is present and explains that email lookup is not live until step 5.
4. **Web window.** A standalone `/bar` page: passkey sign-in, allow-list, composer. Same sealed send as the phone.
5. **Email reach address.** Optional blinded email, same privacy as phone PSI. Until this lands, Mac and web add friends by `@username`.

## What stays put

- Frozen onboarding and login.
- `did:key` only. No email signup, no vendor account, no Cardano identity.
- Full Echo tabs, rewards, and governance.
- Agent grants (`docs/AGENT_BUS.md`) are a different allow-list: agents, not friends. The bar does not reuse `Authorization: Grant`.
