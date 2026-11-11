import React, { useState } from "react";
import { useNavigate } from "react-router";
import { Shield, Eye, EyeOff, Printer, AlertTriangle } from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import { motion } from "motion/react";

const seedWords = [
  "quantum", "vortex", "glacier", "echo", "pillar", "cipher",
  "nebula", "stable", "harbor", "zenith", "obsidian", "tundra",
  "quartz", "mantle", "legacy", "beacon", "static", "horizon",
  "kernel", "plasma", "anchor", "vector", "summit", "drift",
];

export function RecoveryPage() {
  const navigate = useNavigate();
  const { state, update } = useOnboarding();
  const [confirmed, setConfirmed] = useState(false);
  const [wordsHidden, setWordsHidden] = useState(false);

  const handleContinue = () => {
    update({ recoveryMethod: "wallet" });
    navigate("/onboarding/profile");
  };

  const handleSkip = () => {
    update({ recoverySkipped: true });
    navigate("/onboarding/profile");
  };

  return (
    <div className="flex flex-col h-screen" style={{ backgroundColor: "var(--echo-bg-secondary)" }}>
      {/* Secure Thread Indicator */}
      <div className="fixed top-0 left-0 right-0 z-50" style={{ height: 2, backgroundColor: "var(--echo-secure-color)", boxShadow: "var(--echo-secure-glow)", opacity: 0.9 }} />

      <div className="flex-1 overflow-y-auto pt-8 pb-12 px-5">
        {/* Progress Indicator */}
        <div className="flex items-center gap-2 mb-10">
          {[1, 2, 3, 4, 5].map((step) => (
            <div key={step} className="h-1 flex-1 rounded-full" style={{ backgroundColor: "var(--echo-accent)" }} />
          ))}
          <div className="h-1 w-24 rounded-full" style={{ backgroundColor: "var(--echo-border)" }} />
        </div>

        <div className="mb-6">
          <span className="inline-flex items-center rounded-full px-3 py-1 mb-4" style={{ fontSize: 11, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-accent)", backgroundColor: "var(--echo-accent-subtle)" }}>
            Step 05 &bull; Security
          </span>

          <motion.h1 initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5, type: "spring" }} style={{ fontSize: 44, fontWeight: 800, letterSpacing: "-0.03em", lineHeight: 0.9, color: "var(--echo-text-primary)", margin: "0 0 16px" }}>
            Your Master Key.
          </motion.h1>
          <p style={{ fontSize: 17, color: "var(--echo-text-secondary)", lineHeight: "1.6", maxWidth: 380 }}>
            This 24-word phrase is the only way to recover your account. Write it down and store it offline.
          </p>
        </div>

        {/* Warning Box */}
        <div className="flex items-start gap-4 rounded-3xl p-5 mb-6" style={{ backgroundColor: "var(--echo-error-light)", border: "1px solid var(--echo-ghost-border)" }}>
          <AlertTriangle size={20} style={{ color: "var(--echo-error)", flexShrink: 0, marginTop: 2 }} />
          <div>
            <p style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-text-primary)", margin: 0, lineHeight: "1.2" }}>Never share this phrase.</p>
            <p style={{ fontSize: 13, color: "var(--echo-text-secondary)", margin: "4px 0 0", lineHeight: "1.5" }}>Anyone with these words can access your assets and encrypted messages permanently.</p>
          </div>
        </div>

        {/* Recovery Phrase Card */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5, delay: 0.15 }} className="rounded-[2rem] p-6 relative overflow-hidden mb-8" style={{ backgroundColor: "var(--echo-bg-elevated)", border: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-lg)" }}>
          <div className="absolute -top-20 -right-20 rounded-full pointer-events-none" style={{ width: 200, height: 200, background: `radial-gradient(circle, var(--echo-sky-subtle), transparent)`, filter: "blur(60px)" }} />

          <div className="relative z-10">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center rounded-xl" style={{ width: 40, height: 40, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` }}>
                  <Shield size={18} color="white" />
                </div>
                <span style={{ fontSize: 13, fontWeight: 700, color: "var(--echo-text-primary)" }}>Encryption Seed v2.4</span>
              </div>
              <button onClick={() => setWordsHidden(!wordsHidden)} className="flex items-center gap-1 bg-transparent border-none cursor-pointer" style={{ fontSize: 11, fontWeight: 700, color: "var(--echo-accent)" }}>
                {wordsHidden ? <Eye size={14} /> : <EyeOff size={14} />}
                {wordsHidden ? "SHOW" : "HIDE"} WORDS
              </button>
            </div>

            <div className="grid grid-cols-2 gap-2">
              {seedWords.map((word, i) => (
                <div key={i} className="flex items-center gap-3 rounded-2xl px-3 py-2.5" style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-ghost-border)" }}>
                  <span style={{ fontSize: 10, fontWeight: 700, color: "var(--echo-text-tertiary)", width: 16, textAlign: "right" as const }}>{String(i + 1).padStart(2, "0")}</span>
                  <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>{wordsHidden ? "••••••" : word}</span>
                </div>
              ))}
            </div>

            <div className="flex items-center justify-between mt-6">
              <button className="flex items-center gap-2 rounded-2xl px-4 py-2.5 border-none cursor-pointer" style={{ fontSize: 13, fontWeight: 700, color: "var(--echo-accent)", backgroundColor: "var(--echo-bg-surface)" }}>
                <Printer size={14} /> PRINT TO PDF
              </button>
              <div className="flex items-center gap-2">
                <div className="rounded-full animate-pulse" style={{ width: 8, height: 8, backgroundColor: "var(--echo-accent)" }} />
                <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Offline Generation Active</span>
              </div>
            </div>
          </div>
        </motion.div>

        <label className="flex items-center gap-3 cursor-pointer mb-6 px-1">
          <input type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} className="rounded-lg" style={{ width: 24, height: 24, accentColor: "var(--echo-accent)", flexShrink: 0 }} />
          <span style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "1.4" }}>I understand that if I lose this phrase, ECHO cannot recover my account.</span>
        </label>

        <button onClick={handleContinue} disabled={!confirmed} className="w-full border-none rounded-full flex items-center justify-center active:scale-[0.98] transition-all mb-4" style={{ height: 56, background: confirmed ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "var(--echo-bg-surface)", color: confirmed ? "white" : "var(--echo-text-tertiary)", fontSize: 17, fontWeight: 700, cursor: confirmed ? "pointer" : "not-allowed", boxShadow: confirmed ? "var(--echo-shadow-primary)" : "none" }}>
          I've Secured My Phrase
        </button>
        <button onClick={handleSkip} className="w-full bg-transparent border-none cursor-pointer py-3" style={{ fontSize: 12, fontWeight: 700, letterSpacing: "0.15em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>
          Back to step 4
        </button>
      </div>
    </div>
  );
}
