import React from "react";
import { useNavigate } from "react-router";
import { ShieldCheck, Lock, EyeOff, Fingerprint } from "lucide-react";
import { motion } from "motion/react";
import echoBrandLogo from "../../../imports/echo-ripple-brand-1.svg";
import echoWhiteMono from "../../../imports/echo-ripple-white-mono-1.svg";

const features = [
  { icon: Lock, title: "No Central Server", body: "Peer-to-peer architecture ensures your messages never touch a corporate database." },
  { icon: EyeOff, title: "Zero Metadata", body: "We don't know who you are, who you talk to, or when. Truly invisible communication." },
  { icon: Fingerprint, title: "Biometric Vault", body: "Local-only key storage protected by your device's most advanced security." },
];

export function WelcomePage() {
  const navigate = useNavigate();

  return (
    <div className="relative flex flex-col min-h-screen overflow-hidden" style={{ backgroundColor: "var(--echo-bg-primary)" }}>
      {/* Secure Thread Indicator */}
      <motion.div
        className="fixed top-0 left-0 w-full z-[60]"
        style={{ height: 2, backgroundColor: "var(--echo-secure-color)", filter: "blur(0.5px)" }}
        animate={{ opacity: [0.6, 1, 0.6] }}
        transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
      />

      {/* Icy Background Orbs */}
      <div className="absolute pointer-events-none" style={{ top: "-10%", left: "-10%", width: "50%", height: "50%", background: "var(--echo-sky)", opacity: 0.08, borderRadius: "50%", filter: "blur(120px)" }} />
      <div className="absolute pointer-events-none" style={{ bottom: "-10%", right: "-10%", width: "40%", height: "40%", background: "var(--echo-sky-light)", opacity: 0.06, borderRadius: "50%", filter: "blur(100px)" }} />

      {/* Top Nav */}
      <nav className="fixed top-0 w-full z-50 flex items-center justify-between px-6" style={{ height: 64, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(24px)", WebkitBackdropFilter: "blur(24px)", borderBottom: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-sm)" }}>
        <div className="flex items-center gap-2">
          <img src={echoBrandLogo} alt="Echo" width={24} height={24} />
          <span style={{ fontSize: 18, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--echo-text-primary)" }}>Secure Thread</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="rounded-full" style={{ width: 8, height: 8, backgroundColor: "var(--echo-secure-color)", boxShadow: `0 0 15px var(--echo-sky-glow)` }} />
          <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.1em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Encrypted</span>
        </div>
      </nav>

      <main className="flex-1 flex flex-col items-center px-6 pt-28 pb-16 relative z-10 overflow-y-auto">
        <div className="w-full flex flex-col items-center text-center" style={{ maxWidth: 480 }}>
          {/* Logo */}
          <motion.div initial={{ scale: 0.8, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} transition={{ duration: 0.6, type: "spring", damping: 20 }} className="relative mb-10">
            <div className="flex items-center justify-center" style={{ width: 96, height: 96, borderRadius: 32, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, boxShadow: "var(--echo-shadow-xl)", transform: "rotate(3deg)" }}>
              <img src={echoWhiteMono} alt="Echo" width={56} height={56} />
            </div>
            <div className="absolute flex items-center justify-center" style={{ bottom: -8, right: -8, width: 32, height: 32, borderRadius: "50%", backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-md)", border: "1px solid var(--echo-ghost-border)" }}>
              <ShieldCheck size={18} style={{ color: "var(--echo-sky)" }} />
            </div>
          </motion.div>

          {/* Headline */}
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15, duration: 0.5 }} className="mb-6">
            <h1 style={{ fontSize: 40, fontWeight: 800, color: "var(--echo-text-primary)", lineHeight: 1.1, letterSpacing: "-0.02em", margin: 0 }}>
              The Future of Messaging is{" "}
              <span style={{ background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>
                Decentralized.
              </span>
            </h1>
          </motion.div>

          <motion.p initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.25, duration: 0.4 }} style={{ fontSize: 16, color: "var(--echo-text-secondary)", lineHeight: 1.6, maxWidth: 360, margin: "0 0 32px", opacity: 0.8 }}>
            ECHO connects you directly, without middlemen or metadata logging. Your keys, your conversations, your absolute privacy.
          </motion.p>

          {/* CTA Buttons */}
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.35, duration: 0.4 }} className="flex flex-col items-center gap-4 w-full mb-12" style={{ maxWidth: 320 }}>
            <button onClick={() => navigate("/onboarding/phone")} className="w-full border-none cursor-pointer flex items-center justify-center active:scale-95 transition-transform" style={{ height: 56, borderRadius: 9999, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, fontSize: 17, fontWeight: 700, color: "white", boxShadow: "var(--echo-shadow-primary)" }}>
              Get Started
            </button>
            <button onClick={() => navigate("/onboarding/login")} className="w-full cursor-pointer flex items-center justify-center active:scale-95 transition-transform" style={{ height: 56, borderRadius: 9999, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(20px)", border: "1px solid var(--echo-ghost-border)", fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
              I already have a Vault
            </button>
          </motion.div>

          {/* Feature Cards */}
          <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.5, duration: 0.5 }} className="flex flex-col gap-5 w-full">
            {features.map((f, i) => (
              <div key={i} className="text-left" style={{ padding: "24px 28px", borderRadius: 24, backgroundColor: "var(--echo-bg-surface)", transition: "background-color 0.2s" }}>
                <div className="flex items-center justify-center mb-3" style={{ width: 44, height: 44, borderRadius: 14, backgroundColor: "var(--echo-sky-subtle)" }}>
                  <f.icon size={22} style={{ color: "var(--echo-accent)" }} />
                </div>
                <h3 style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-text-primary)", margin: "0 0 4px" }}>{f.title}</h3>
                <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: 1.5, margin: 0 }}>{f.body}</p>
              </div>
            ))}
          </motion.div>
        </div>
      </main>

      <div className="flex flex-col items-center gap-4 py-8 px-6 relative z-10" style={{ opacity: 0.6 }}>
        <p style={{ fontSize: 11, textTransform: "uppercase" as const, letterSpacing: "0.15em", fontWeight: 700, color: "var(--echo-text-tertiary)" }}>
          Powered by ECHO Decentralized Protocol
        </p>
      </div>
    </div>
  );
}
