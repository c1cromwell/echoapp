import React from "react";
import { useNavigate } from "react-router";
import { useTheme, type AppearanceMode } from "../components/echo-theme";
import {
  Bell,
  Lock,
  Globe,
  HelpCircle,
  LogOut,
  ChevronRight,
  Moon,
  Shield,
  Database,
  Smartphone,
  Users,
  RotateCcw,
  UserCircle,
  BarChart3,
  FolderLock,
  Radio,
  History,
  ArrowLeft,
  IdCard,
  Sun,
  Monitor,
} from "lucide-react";

const settingsGroups = [
  {
    title: "Trust & Identity",
    items: [
      { icon: Shield, label: "Trust Score", subtitle: "72/100 — Trusted", path: "/trust" },
      { icon: IdCard, label: "Verify Identity", subtitle: "Government ID, Apple ID" },
      { icon: Smartphone, label: "Devices", subtitle: "2 devices registered", path: "/devices" },
      { icon: History, label: "Login History", subtitle: "Last login: 2 min ago", path: "/login-audit" },
    ],
  },
  {
    title: "Preferences",
    items: [
      { icon: Bell, label: "Notifications", subtitle: "Push, email, sounds" },
      { icon: Globe, label: "Language", subtitle: "English (US)" },
    ],
  },
  {
    title: "Identity",
    items: [
      { icon: UserCircle, label: "Personas", subtitle: "Manage your identities", path: "/personas" },
      { icon: Users, label: "Groups", subtitle: "Your groups & settings", path: "/" },
      { icon: Radio, label: "Channels", subtitle: "Broadcast channels & subscriptions", path: "/channels/discover" },
    ],
  },
  {
    title: "Privacy & Security",
    items: [
      { icon: Lock, label: "Privacy", subtitle: "Last seen, profile photo" },
      { icon: FolderLock, label: "Hidden Folders", subtitle: "Biometric-protected vaults", path: "/hidden-folders", requireAuth: true },
      { icon: Database, label: "Data & Storage", subtitle: "Network usage, storage" },
    ],
  },
  {
    title: "Support",
    items: [
      { icon: HelpCircle, label: "Help Center", subtitle: "FAQ, contact us" },
      { icon: BarChart3, label: "Analytics", subtitle: "Onboarding funnel metrics", path: "/analytics" },
    ],
  },
];

const appearanceModes: { value: AppearanceMode; label: string; icon: React.ElementType }[] = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor },
];

export function SettingsPage() {
  const { appearance, setAppearance, isDark } = useTheme();
  const navigate = useNavigate();

  const handleResetOnboarding = () => {
    localStorage.removeItem("echo-onboarded");
    localStorage.removeItem("echo-nudge-dismissed");
    window.location.href = "/onboarding";
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-3 px-4 pt-14 pb-2">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
        <h1
          style={{
            fontSize: 28,
            fontWeight: 700,
            color: "var(--echo-text-primary)",
            lineHeight: "36px",
            letterSpacing: "-0.5px",
            margin: 0,
          }}
        >
          Settings
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-6">
        {/* Appearance Mode */}
        <div className="mb-6">
          <h3
            style={{
              fontSize: 13,
              fontWeight: 600,
              color: "var(--echo-text-tertiary)",
              letterSpacing: "0.5px",
              textTransform: "uppercase",
              margin: "0 0 10px 4px",
            }}
          >
            Appearance
          </h3>

          {/* Segmented Control */}
          <div
            className="flex rounded-2xl p-1.5"
            style={{
              backgroundColor: "var(--echo-bg-surface)",
            }}
          >
            {appearanceModes.map((mode) => {
              const isActive = appearance === mode.value;
              const Icon = mode.icon;
              return (
                <button
                  key={mode.value}
                  onClick={() => setAppearance(mode.value)}
                  className="flex-1 flex items-center justify-center gap-2 rounded-xl border-none cursor-pointer py-3 transition-all"
                  style={{
                    backgroundColor: isActive ? "var(--echo-bg-elevated)" : "transparent",
                    boxShadow: isActive ? "var(--echo-shadow-sm)" : "none",
                  }}
                >
                  <Icon
                    size={16}
                    style={{
                      color: isActive ? "var(--echo-accent)" : "var(--echo-text-tertiary)",
                    }}
                  />
                  <span
                    style={{
                      fontSize: 14,
                      fontWeight: isActive ? 600 : 500,
                      color: isActive ? "var(--echo-text-primary)" : "var(--echo-text-tertiary)",
                    }}
                  >
                    {mode.label}
                  </span>
                </button>
              );
            })}
          </div>

          {/* Mini Preview */}
          <div
            className="mt-3 rounded-2xl overflow-hidden p-3"
            style={{
              backgroundColor: isDark ? "#0F172A" : "#F8FAFC",
              border: `1px solid ${isDark ? "#334155" : "#E2E8F0"}`,
            }}
          >
            <div className="flex flex-col gap-2">
              {["Sarah Chen", "Marcus J."].map((name, i) => (
                <div key={name} className="flex items-center gap-2.5">
                  <div
                    className="rounded-full shrink-0"
                    style={{
                      width: 32,
                      height: 32,
                      background: i === 0
                        ? "linear-gradient(135deg, #0EA5E9, #7DD3FC)"
                        : "linear-gradient(135deg, #64748B, #94A3B8)",
                    }}
                  />
                  <div className="flex-1 min-w-0">
                    <div
                      style={{
                        fontSize: 13,
                        fontWeight: 600,
                        color: isDark ? "#F1F5F9" : "#0F172A",
                      }}
                    >
                      {name}
                    </div>
                    <div
                      style={{
                        fontSize: 11,
                        color: isDark ? "#94A3B8" : "#64748B",
                      }}
                    >
                      {i === 0 ? "That sounds great!" : "The presentation is ready"}
                    </div>
                  </div>
                  <span style={{ fontSize: 10, color: isDark ? "#64748B" : "#94A3B8" }}>
                    {i === 0 ? "2m" : "15m"}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Settings Groups */}
        {settingsGroups.map((group) => (
          <div key={group.title} className="mb-5">
            <h3
              style={{
                fontSize: 13,
                fontWeight: 600,
                color: "var(--echo-text-tertiary)",
                letterSpacing: "0.5px",
                textTransform: "uppercase",
                margin: "0 0 8px 4px",
              }}
            >
              {group.title}
            </h3>
            <div
              className="rounded-2xl overflow-hidden"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              {group.items.map((item, i) => {
                const Icon = item.icon;
                return (
                  <div
                    key={item.label}
                    className="flex items-center gap-3 px-4 py-3.5 cursor-pointer hover:bg-[var(--echo-bg-surface)] transition-colors"
                    style={{
                      borderBottom:
                        i < group.items.length - 1
                          ? "1px solid var(--echo-border-light)"
                          : "none",
                    }}
                    onClick={() => (item as any).path ? navigate((item as any).path) : undefined}
                  >
                    <div
                      className="flex items-center justify-center rounded-xl"
                      style={{
                        width: 36,
                        height: 36,
                        backgroundColor: "var(--echo-accent-subtle)",
                      }}
                    >
                      <Icon size={18} style={{ color: "var(--echo-accent)" }} />
                    </div>
                    <div className="flex-1">
                      <div
                        style={{
                          fontSize: 15,
                          fontWeight: 500,
                          color: "var(--echo-text-primary)",
                        }}
                      >
                        {item.label}
                      </div>
                      <div
                        style={{
                          fontSize: 13,
                          color: "var(--echo-text-tertiary)",
                        }}
                      >
                        {item.subtitle}
                      </div>
                    </div>
                    <ChevronRight size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                  </div>
                );
              })}
            </div>
          </div>
        ))}

        {/* Reset Onboarding */}
        <button
          onClick={handleResetOnboarding}
          className="w-full flex items-center justify-center gap-2 rounded-2xl border-none cursor-pointer py-3.5 mb-3"
          style={{
            backgroundColor: "var(--echo-warning-light)",
          }}
        >
          <RotateCcw size={18} style={{ color: "var(--echo-warning)" }} />
          <span
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: "var(--echo-warning)",
            }}
          >
            Reset Onboarding
          </span>
        </button>

        {/* Logout */}
        <button
          onClick={() => {
            localStorage.removeItem("echo-onboarded");
            window.location.href = "/onboarding/login";
          }}
          className="w-full flex items-center justify-center gap-2 rounded-2xl border-none cursor-pointer py-3.5"
          style={{
            backgroundColor: "var(--echo-error-light)",
          }}
        >
          <LogOut size={18} style={{ color: "var(--echo-error)" }} />
          <span
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: "var(--echo-error)",
            }}
          >
            Log Out
          </span>
        </button>
      </div>
    </div>
  );
}
