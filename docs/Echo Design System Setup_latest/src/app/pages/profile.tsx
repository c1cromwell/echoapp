import React from "react";
import { useNavigate } from "react-router";
import { PersonaSwitcher } from "../components/persona-switcher";
import { Settings, ChevronRight, Monitor, Shield, Bell, LogOut } from "lucide-react";
import { ImageWithFallback } from "../components/figma/ImageWithFallback";

const trustTiers = [
  { label: "Basic", active: false },
  { label: "Verified", active: false },
  { label: "Peer", active: false },
  { label: "Trusted", active: true },
  { label: "Inner", active: false },
];

const menuItems = [
  { icon: Monitor, label: "Device Management", subtitle: "3 active encrypted sessions" },
  { icon: Shield, label: "Security", subtitle: "Quantum-resistant keys active", path: "/settings" },
  { icon: Bell, label: "Notifications", subtitle: "Muted for focused work" },
];

export function ProfilePage() {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: "var(--echo-bg-primary)" }}>
      <div className="flex-1 overflow-y-auto">
        {/* Header bar */}
        <div className="flex items-center justify-between px-5 pt-14 pb-4">
          <span style={{ fontSize: 24, fontWeight: 800, color: "var(--echo-text-primary)", letterSpacing: "-0.5px" }}>ECHO</span>
          <button onClick={() => navigate("/settings")} className="flex items-center justify-center rounded-full border-none cursor-pointer" style={{ width: 40, height: 40, backgroundColor: "transparent" }}>
            <Settings size={22} style={{ color: "var(--echo-text-tertiary)" }} />
          </button>
        </div>

        {/* Profile Header */}
        <section className="flex flex-col items-center text-center px-6 pt-4 pb-8">
          <div className="relative mb-6">
            <div className="rounded-full overflow-hidden" style={{ width: 140, height: 140, padding: 3, background: `linear-gradient(135deg, var(--echo-sky), var(--echo-sky-light))`, boxShadow: `0 8px 32px var(--echo-sky-glow)` }}>
              <div className="w-full h-full rounded-full overflow-hidden" style={{ border: "4px solid var(--echo-bg-primary)" }}>
                <ImageWithFallback src="https://images.unsplash.com/photo-1769636929388-99eff95d3bf1?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx3b21hbiUyMHByb2Zlc3Npb25hbCUyMGhlYWRzaG90JTIwc3R1ZGlvfGVufDF8fHx8MTc3NTEzOTg0MHww&ixlib=rb-4.1.0&q=80&w=1080" alt="Profile" className="w-full h-full object-cover" />
              </div>
            </div>
            <div className="absolute flex items-center justify-center rounded-full" style={{ bottom: 4, right: 8, width: 36, height: 36, backgroundColor: "var(--echo-sky)", border: "4px solid var(--echo-bg-primary)", boxShadow: `0 2px 8px var(--echo-sky-glow)` }}>
              <Shield size={16} color="white" />
            </div>
          </div>
          <h1 style={{ fontSize: 28, fontWeight: 800, letterSpacing: "-0.5px", color: "var(--echo-text-primary)", margin: 0 }}>Alice Eth</h1>
          <p className="mt-2 inline-block rounded-full px-3 py-1" style={{ fontSize: 13, fontWeight: 600, letterSpacing: "0.1em", textTransform: "uppercase" as const, color: "var(--echo-tier-inner-text)", backgroundColor: "var(--echo-tier-inner-bg)", border: "1px solid var(--echo-tier-inner-border)" }}>
            echo:alice.eth
          </p>
        </section>

        {/* Trust Ring Visualization */}
        <section className="mx-5 rounded-[32px] p-6 mb-6" style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}>
          <div className="flex justify-between items-end mb-5">
            <div>
              <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.2em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Reputation Level</span>
              <h2 style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)", margin: "4px 0 0" }}>Trusted Node</h2>
            </div>
            <span style={{ fontSize: 24, fontWeight: 800, color: "var(--echo-trust-active)", fontStyle: "italic" }}>4/5</span>
          </div>

          <div className="flex items-center justify-between gap-2 h-12 mb-3">
            {[
              { h: 8, varName: "--echo-trust-bar-1" },
              { h: 12, varName: "--echo-trust-bar-2" },
              { h: 16, varName: "--echo-trust-bar-3" },
              { h: 24, varName: "--echo-trust-active", active: true },
              { h: 12, varName: "--echo-trust-bar-5" },
            ].map((bar, i) => (
              <div
                key={i}
                className="flex-1 rounded-full"
                style={{
                  height: bar.h,
                  backgroundColor: `var(${bar.varName})`,
                  boxShadow: bar.active ? `0 0 12px var(--echo-sky-glow), 0 0 0 4px var(--echo-sky-ring)` : "none",
                }}
              />
            ))}
          </div>
          <div className="grid grid-cols-5 px-1">
            {trustTiers.map((t) => (
              <span key={t.label} className="text-center" style={{ fontSize: 9, fontWeight: 700, textTransform: "uppercase" as const, letterSpacing: "-0.02em", color: t.active ? "var(--echo-trust-active)" : "var(--echo-text-tertiary)" }}>
                {t.label}
              </span>
            ))}
          </div>
        </section>

        <div className="px-5 mb-4"><PersonaSwitcher /></div>

        {/* Account Configuration */}
        <section className="px-5 mb-6">
          <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.2em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)", marginLeft: 8, marginBottom: 12, display: "block" }}>Account Configuration</span>
          <div className="space-y-2">
            {menuItems.map((item) => {
              const Icon = item.icon;
              return (
                <button key={item.label} onClick={() => item.path ? navigate(item.path) : undefined} className="flex items-center justify-between w-full p-4 rounded-2xl border-none cursor-pointer transition-all text-left" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center justify-center rounded-xl" style={{ width: 40, height: 40, backgroundColor: "var(--echo-bg-elevated)" }}>
                      <Icon size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                    </div>
                    <div>
                      <p style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-text-primary)", margin: 0 }}>{item.label}</p>
                      <p style={{ fontSize: 12, color: "var(--echo-text-tertiary)", margin: "2px 0 0" }}>{item.subtitle}</p>
                    </div>
                  </div>
                  <ChevronRight size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                </button>
              );
            })}
          </div>
        </section>

        <div className="flex justify-center pb-10 pt-4">
          <button className="flex items-center gap-2 rounded-full px-6 py-3 cursor-pointer transition-colors" style={{ fontSize: 12, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-error)", backgroundColor: "transparent", border: "1px solid var(--echo-error-light)" }}>
            <LogOut size={14} />
            Disconnect Identity
          </button>
        </div>
      </div>
    </div>
  );
}
