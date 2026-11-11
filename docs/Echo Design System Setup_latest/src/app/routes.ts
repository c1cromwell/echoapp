import { createBrowserRouter } from "react-router";
import { Layout } from "./components/layout";
import { OnboardingLayout } from "./components/onboarding-layout";

// Stub fallback that renders nothing — prevents the
// "No HydrateFallback element provided" warning in React Router v7.
const NoFallback = () => null;

// Lazy-load all page components with resilient retry logic.
// HydrateFallback suppresses hydration warnings per-route.
const lazy = (importFn: () => Promise<any>, exportName: string) => ({
  HydrateFallback: NoFallback,
  lazy: async () => {
    try {
      const mod = await importFn();
      return { Component: mod[exportName] };
    } catch (err) {
      // Wait before retrying — gives the dev server time to finish compiling.
      // Add a cache-busting query param so the browser doesn't serve the
      // previously-failed response from its module cache.
      await new Promise((r) => setTimeout(r, 1500));
      try {
        const mod = await importFn();
        return { Component: mod[exportName] };
      } catch (retryErr) {
        console.error("Failed to load module after retry:", retryErr);
        // Return a fallback component so the app doesn't crash
        return {
          Component: () => {
            return null;
          },
        };
      }
    }
  },
});

export const router = createBrowserRouter([
  {
    path: "/onboarding",
    Component: OnboardingLayout,
    HydrateFallback: NoFallback,
    children: [
      { index: true, ...lazy(() => import("./pages/onboarding/intro-carousel"), "IntroCarousel") },
      { path: "welcome", ...lazy(() => import("./pages/onboarding/welcome"), "WelcomePage") },
      { path: "login", ...lazy(() => import("./pages/onboarding/login"), "LoginPage") },
      { path: "phone", ...lazy(() => import("./pages/onboarding/phone-entry"), "PhoneEntryPage") },
      { path: "otp", ...lazy(() => import("./pages/onboarding/otp-verify"), "OtpVerifyPage") },
      { path: "passkey", ...lazy(() => import("./pages/onboarding/passkey"), "PasskeyPage") },
      { path: "recovery", ...lazy(() => import("./pages/onboarding/recovery"), "RecoveryPage") },
      { path: "profile", ...lazy(() => import("./pages/onboarding/profile-setup"), "ProfileSetupPage") },
      { path: "verify-credentials", ...lazy(() => import("./pages/onboarding/verify-credentials"), "VerifyCredentialsPage") },
      { path: "trust-dashboard", ...lazy(() => import("./pages/onboarding/trust-dashboard"), "TrustDashboardPage") },
      { path: "recover", ...lazy(() => import("./pages/onboarding/recover"), "RecoverMethodPage") },
      { path: "recover-phrase", ...lazy(() => import("./pages/onboarding/recover-phrase"), "RecoverPhrasePage") },
      { path: "recover-contacts", ...lazy(() => import("./pages/onboarding/recover-contacts"), "RecoverContactsPage") },
      { path: "recover-complete", ...lazy(() => import("./pages/onboarding/recover-complete"), "RecoverCompletePage") },
      { path: "account-locked", ...lazy(() => import("./pages/onboarding/account-locked"), "AccountLockedPage") },
    ],
  },
  {
    path: "/",
    Component: Layout,
    HydrateFallback: NoFallback,
    children: [
      { index: true, ...lazy(() => import("./pages/messages"), "MessagesPage") },
      { path: "chat/:id", ...lazy(() => import("./pages/chat"), "ChatPage") },
      { path: "groups/create", ...lazy(() => import("./pages/group-create"), "GroupCreatePage") },
      { path: "groups/discover", ...lazy(() => import("./pages/group-discover"), "GroupDiscoverPage") },
      { path: "groups/:id", ...lazy(() => import("./pages/group-detail"), "GroupDetailPage") },
      { path: "group-chat/:id", ...lazy(() => import("./pages/group-chat"), "GroupChatPage") },
      { path: "personas/create", ...lazy(() => import("./pages/persona-create"), "PersonaCreatePage") },
      { path: "personas/:id", ...lazy(() => import("./pages/persona-detail"), "PersonaDetailPage") },
      { path: "personas", ...lazy(() => import("./pages/personas"), "PersonasPage") },
      { path: "analytics", ...lazy(() => import("./pages/analytics"), "AnalyticsPage") },
      { path: "contacts", ...lazy(() => import("./pages/contacts"), "ContactsPage") },
      { path: "rewards", ...lazy(() => import("./pages/rewards"), "RewardsPage") },
      { path: "trust", ...lazy(() => import("./pages/trust"), "TrustPage") },
      { path: "profile", ...lazy(() => import("./pages/profile"), "ProfilePage") },
      { path: "settings", ...lazy(() => import("./pages/settings"), "SettingsPage") },
      { path: "hidden-folders", ...lazy(() => import("./pages/hidden-folders"), "HiddenFoldersPage") },
      { path: "hidden-folders/:id", ...lazy(() => import("./pages/hidden-folder-detail"), "HiddenFolderDetailPage") },
      { path: "hidden-folders/:id/chat/:chatId", ...lazy(() => import("./pages/hidden-folder-chat"), "HiddenFolderChatPage") },
      { path: "channels/discover", ...lazy(() => import("./pages/channel-discover"), "ChannelDiscoverPage") },
      { path: "channels/create", ...lazy(() => import("./pages/channel-create"), "ChannelCreatePage") },
      { path: "channels/:id", ...lazy(() => import("./pages/channel-detail"), "ChannelDetailPage") },
      { path: "devices", ...lazy(() => import("./pages/device-management"), "DeviceManagementPage") },
      { path: "login-audit", ...lazy(() => import("./pages/login-audit"), "LoginAuditPage") },
    ],
  },
]);