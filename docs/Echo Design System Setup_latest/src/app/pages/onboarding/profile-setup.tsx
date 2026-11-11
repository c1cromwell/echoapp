import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { ArrowRight, Camera, Check, Loader2, Info } from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import echoBrandLogo from "../../../imports/echo-ripple-brand-1.svg";

const colorPresets = [
  "linear-gradient(135deg, #38BDF8, #6366F1)",
  "linear-gradient(135deg, #34D399, #0F766E)",
  "linear-gradient(135deg, #FB923C, #E11D48)",
  "linear-gradient(135deg, #D946EF, #581C87)",
  "linear-gradient(135deg, #334155, #0F172A)",
];

export function ProfileSetupPage() {
  const navigate = useNavigate();
  const { state, update } = useOnboarding();
  const [name, setName] = useState(state.displayName);
  const [username, setUsername] = useState(state.username);
  const [usernameStatus, setUsernameStatus] = useState<"idle" | "checking" | "available" | "taken">("idle");

  useEffect(() => {
    if (!username || username.length < 3) { setUsernameStatus("idle"); return; }
    setUsernameStatus("checking");
    const timer = setTimeout(() => {
      setUsernameStatus(username.toLowerCase() === "taken_user" ? "taken" : "available");
    }, 300);
    return () => clearTimeout(timer);
  }, [username]);

  const canContinue = name.trim().length > 0;

  const handleContinue = () => {
    update({ displayName: name.trim(), username: username.trim() });
    navigate("/onboarding/verify-credentials");
  };

  return (
    <div className="flex flex-col h-screen" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
      <div className="fixed top-0 left-0 right-0 z-50" style={{ height: 2, backgroundColor: "var(--echo-secure-color)", boxShadow: "var(--echo-secure-glow)", opacity: 0.8 }} />

      <header className="flex items-center justify-between px-5 pt-14 pb-3" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
        <div className="flex items-center gap-2">
          <img src={echoBrandLogo} alt="Echo" width={24} height={24} />
          <span style={{ fontSize: 18, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--echo-text-primary)" }}>Secure Thread</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="h-1 rounded-full overflow-hidden" style={{ width: 80, backgroundColor: "var(--echo-border)" }}>
            <div className="h-full rounded-full" style={{ width: "100%", backgroundColor: "var(--echo-accent)" }} />
          </div>
          <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: "0.1em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Step 6 of 6</span>
        </div>
      </header>

      <main className="flex-1 overflow-y-auto px-6 pb-16">
        <div className="mb-10 mt-6">
          <h1 style={{ fontSize: 44, fontWeight: 800, letterSpacing: "-0.03em", lineHeight: 0.95, color: "var(--echo-text-primary)", margin: "0 0 12px" }}>
            Create Your Persona.
          </h1>
          <p style={{ fontSize: 17, color: "var(--echo-text-secondary)", lineHeight: "1.5", maxWidth: 340 }}>
            Set your public identity for the ECHO network. This is how others will recognize and connect with you.
          </p>
        </div>

        {/* Avatar Section */}
        <div className="flex flex-col items-center mb-8">
          <div className="relative">
            <div className="rounded-full overflow-hidden" style={{ width: 160, height: 160, padding: 3, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, boxShadow: `0 8px 32px var(--echo-sky-glow)` }}>
              <div className="w-full h-full rounded-full flex items-center justify-center overflow-hidden" style={{ backgroundColor: "var(--echo-bg-elevated)", border: "4px solid var(--echo-bg-secondary)" }}>
                {name.trim() ? (
                  <span style={{ fontSize: 56, fontWeight: 600, color: "var(--echo-text-tertiary)" }}>{name.trim()[0].toUpperCase()}</span>
                ) : (
                  <Camera size={48} style={{ color: "var(--echo-text-tertiary)" }} />
                )}
              </div>
            </div>
            <button className="absolute flex items-center justify-center rounded-full border-none cursor-pointer active:scale-95 transition-transform" style={{ bottom: 8, right: 8, width: 44, height: 44, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, boxShadow: `0 4px 12px var(--echo-sky-glow)` }}>
              <Camera size={18} color="white" />
            </button>
          </div>

          <div className="mt-6 w-full max-w-xs">
            <label style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)", marginBottom: 12, display: "block" }}>Quick Presets</label>
            <div className="flex justify-between gap-2">
              {colorPresets.map((bg, i) => (
                <div key={i} className="rounded-xl cursor-pointer hover:scale-110 transition-transform" style={{ width: 40, height: 40, background: bg }} />
              ))}
            </div>
          </div>
        </div>

        {/* Form Fields */}
        <div className="space-y-5 mb-8">
          <div>
            <label style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)", display: "block", marginBottom: 8, marginLeft: 4 }}>Display Name</label>
            <input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Alex Rivera" className="w-full rounded-xl px-4 outline-none transition-all" style={{ height: 56, backgroundColor: "var(--echo-bg-surface)", border: "none", fontSize: 17, color: "var(--echo-text-primary)" }} />
          </div>

          <div>
            <label style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)", display: "block", marginBottom: 8, marginLeft: 4 }}>Unique @Handle</label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2" style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-accent)" }}>@</span>
              <input value={username} onChange={(e) => setUsername(e.target.value.replace(/[^a-zA-Z0-9_]/g, ""))} placeholder="alex_echo" className="w-full rounded-xl pl-10 pr-12 outline-none transition-all" style={{ height: 56, backgroundColor: "var(--echo-bg-surface)", border: "none", fontSize: 17, color: "var(--echo-text-primary)" }} />
              {usernameStatus === "available" && (
                <div className="absolute right-4 top-1/2 -translate-y-1/2"><Check size={20} style={{ color: "var(--echo-success)" }} /></div>
              )}
            </div>
            {usernameStatus === "checking" && (
              <div className="flex items-center gap-1.5 mt-1.5 ml-1">
                <Loader2 size={14} className="animate-spin" style={{ color: "var(--echo-text-tertiary)" }} />
                <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Checking...</span>
              </div>
            )}
            {usernameStatus === "available" && (
              <p style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-success)", letterSpacing: "0.05em", textTransform: "uppercase" as const, marginTop: 6, marginLeft: 4 }}>This handle is available</p>
            )}
            {usernameStatus === "taken" && (
              <p style={{ fontSize: 12, color: "var(--echo-error)", marginTop: 4, marginLeft: 4 }}>Username is taken</p>
            )}
          </div>

          <div className="flex items-start gap-4 rounded-2xl p-5" style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}>
            <Info size={18} style={{ color: "var(--echo-accent)", flexShrink: 0, marginTop: 2 }} />
            <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "1.5", margin: 0 }}>
              Your identity is stored on the ECHO ledger. Once set, your handle cannot be changed for 30 days to ensure network stability.
            </p>
          </div>
        </div>

        <button onClick={handleContinue} disabled={!canContinue} className="w-full border-none rounded-full flex items-center justify-center gap-3 active:scale-[0.98] transition-all mb-4" style={{ height: 56, background: canContinue ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "var(--echo-bg-surface)", color: canContinue ? "white" : "var(--echo-text-tertiary)", fontSize: 17, fontWeight: 700, cursor: canContinue ? "pointer" : "not-allowed", boxShadow: canContinue ? "var(--echo-shadow-primary)" : "none" }}>
          Enter ECHO <ArrowRight size={18} />
        </button>
        <button onClick={() => { update({ displayName: name.trim(), username: username.trim() }); navigate("/onboarding/verify-credentials"); }} className="w-full bg-transparent border-none cursor-pointer py-3" style={{ fontSize: 12, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>
          Skip for now
        </button>
      </main>
    </div>
  );
}
