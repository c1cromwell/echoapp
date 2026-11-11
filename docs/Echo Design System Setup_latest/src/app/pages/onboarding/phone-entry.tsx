import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ArrowRight, ChevronDown, ShieldCheck, Lock, Fingerprint } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { useOnboarding } from "../../components/onboarding-context";
import { AuthError } from "../../components/auth-error";
import echoBrandLogo from "../../../imports/echo-ripple-brand-1.svg";

const countryCodes = [
  { code: "+1", flag: "\u{1F1FA}\u{1F1F8}", name: "US" },
  { code: "+44", flag: "\u{1F1EC}\u{1F1E7}", name: "UK" },
  { code: "+91", flag: "\u{1F1EE}\u{1F1F3}", name: "IN" },
  { code: "+81", flag: "\u{1F1EF}\u{1F1F5}", name: "JP" },
  { code: "+49", flag: "\u{1F1E9}\u{1F1EA}", name: "DE" },
];

function formatPhone(value: string): string {
  const digits = value.replace(/\D/g, "").slice(0, 10);
  if (digits.length === 0) return "";
  if (digits.length <= 3) return digits;
  if (digits.length <= 6) return `${digits.slice(0, 3)}-${digits.slice(3)}`;
  return `${digits.slice(0, 3)}-${digits.slice(3, 6)}-${digits.slice(6)}`;
}

export function PhoneEntryPage() {
  const navigate = useNavigate();
  const { update } = useOnboarding();
  const [phone, setPhone] = useState("");
  const [selectedCode, setSelectedCode] = useState(countryCodes[0]);
  const [showCodes, setShowCodes] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [isFocused, setIsFocused] = useState(false);

  const rawDigits = phone.replace(/\D/g, "");
  const isValid = rawDigits.length >= 7;

  const handleSubmit = () => {
    if (submitting) return;
    if (rawDigits.length > 0 && rawDigits.length < 7) { setError("Enter a valid phone number to continue."); return; }
    if (!isValid) return;
    setError("");
    setSubmitting(true);
    setTimeout(() => {
      setSubmitting(false);
      update({ phoneNumber: rawDigits, countryCode: selectedCode.code });
      navigate("/onboarding/otp");
    }, 600);
  };

  const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (error) setError("");
    setPhone(formatPhone(e.target.value));
  };

  return (
    <div className="relative flex flex-col min-h-screen overflow-hidden" style={{ backgroundColor: "var(--echo-bg-primary)" }}>
      {/* Progress bar */}
      <div className="fixed top-0 left-0 w-full z-[60]" style={{ height: 4, backgroundColor: "var(--echo-bg-surface)" }}>
        <motion.div initial={{ width: "0%" }} animate={{ width: "40%" }} transition={{ duration: 0.6, ease: "easeOut" }} style={{ height: "100%", backgroundColor: "var(--echo-accent)" }} />
      </div>

      {/* Secure Thread */}
      <div className="fixed top-0 left-0 w-full z-[70] pointer-events-none">
        <motion.div style={{ height: 2, backgroundColor: "var(--echo-secure-color)" }} animate={{ opacity: [0.6, 1, 0.6] }} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }} />
      </div>

      {/* Header */}
      <header className="fixed top-0 w-full z-50 flex items-center justify-between px-6" style={{ height: 64, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(24px)", WebkitBackdropFilter: "blur(24px)", borderBottom: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-sm)" }}>
        <div className="flex items-center gap-2">
          <img src={echoBrandLogo} alt="Echo" width={24} height={24} />
          <span style={{ fontSize: 18, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--echo-text-primary)" }}>Secure Thread</span>
        </div>
        <div className="px-3 py-1 rounded-full" style={{ backgroundColor: "var(--echo-accent-subtle)" }}>
          <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: "0.1em", textTransform: "uppercase" as const, color: "var(--echo-accent)" }}>Step 2 of 4</span>
        </div>
      </header>

      <main className="flex-1 w-full px-6 pt-28 pb-24 flex flex-col items-start relative z-10" style={{ maxWidth: 560, margin: "0 auto" }}>
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }} className="mb-8 w-full">
          <h1 style={{ fontSize: 44, fontWeight: 800, color: "var(--echo-text-primary)", lineHeight: 1.1, letterSpacing: "-0.02em", margin: "0 0 16px" }}>
            Secure Your<br />Connection.
          </h1>
          <p style={{ fontSize: 17, color: "var(--echo-text-secondary)", lineHeight: 1.6, maxWidth: 400, margin: 0 }}>
            ECHO uses your number only as a <span style={{ color: "var(--echo-accent)", fontWeight: 600 }}>unique anchor</span>. It is hashed and never leaves your device.
          </p>
        </motion.div>

        {/* Phone Input Card */}
        <motion.section initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15, duration: 0.5 }} className="w-full mb-8">
          <div className="p-8" style={{ borderRadius: 32, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(24px)", WebkitBackdropFilter: "blur(24px)", border: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-sm)" }}>
            <div className="grid gap-3" style={{ gridTemplateColumns: "1fr 2fr" }}>
              <div className="relative">
                <label className="absolute px-1 z-10" style={{ top: -10, left: 14, fontSize: 10, fontWeight: 700, textTransform: "uppercase" as const, letterSpacing: "0.1em", color: "var(--echo-accent)", backgroundColor: "var(--echo-bg-primary)" }}>Country</label>
                <button onClick={() => setShowCodes(!showCodes)} className="flex items-center justify-between w-full bg-transparent cursor-pointer" style={{ height: 52, paddingLeft: 14, paddingRight: 10, borderRadius: 12, backgroundColor: "var(--echo-bg-surface)", border: "none" }}>
                  <span className="flex items-center gap-2">
                    <span style={{ fontSize: 18 }}>{selectedCode.flag}</span>
                    <span style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>{selectedCode.code}</span>
                  </span>
                  <ChevronDown size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                </button>
                <AnimatePresence>
                  {showCodes && (
                    <motion.div initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }} className="absolute top-14 left-0 w-40 z-20 overflow-hidden" style={{ borderRadius: 16, backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-lg)", border: "1px solid var(--echo-ghost-border)" }}>
                      {countryCodes.map((c) => (
                        <button key={c.code} onClick={() => { setSelectedCode(c); setShowCodes(false); }} className="flex items-center gap-2 w-full px-3 py-2.5 bg-transparent border-none cursor-pointer transition-colors">
                          <span style={{ fontSize: 16 }}>{c.flag}</span>
                          <span style={{ fontSize: 14, color: "var(--echo-text-primary)" }}>{c.code}</span>
                          <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>{c.name}</span>
                        </button>
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
              <div className="relative">
                <label className="absolute px-1 z-10" style={{ top: -10, left: 14, fontSize: 10, fontWeight: 700, textTransform: "uppercase" as const, letterSpacing: "0.1em", color: "var(--echo-accent)", backgroundColor: "var(--echo-bg-primary)" }}>Phone Number</label>
                <input type="tel" placeholder="555-0123" value={phone} onChange={handlePhoneChange} onFocus={() => setIsFocused(true)} onBlur={() => setIsFocused(false)} className="w-full outline-none" style={{ height: 52, paddingLeft: 14, paddingRight: 14, borderRadius: 12, backgroundColor: "var(--echo-bg-surface)", border: "none", fontSize: 17, fontWeight: 500, color: "var(--echo-text-primary)", boxShadow: isFocused ? `inset 0 0 0 1px var(--echo-sky-ring)` : "none", transition: "box-shadow 0.2s" }} autoFocus />
              </div>
            </div>
            <div className="flex items-start gap-3 mt-6 p-4" style={{ borderRadius: 16, backgroundColor: "var(--echo-accent-subtle)" }}>
              <ShieldCheck size={20} style={{ color: "var(--echo-accent)", marginTop: 1, flexShrink: 0 }} />
              <p style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: 1.5, margin: 0 }}>
                Encryption keys are locally generated using this identifier as a seed. No personal data is stored on ECHO servers.
              </p>
            </div>
          </div>
        </motion.section>

        <div className="w-full mb-4"><AuthError message={error} onDismiss={() => setError("")} /></div>

        <motion.button initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3, duration: 0.4 }} onClick={handleSubmit} disabled={!isValid || submitting} className="w-full border-none flex items-center justify-center gap-3 active:scale-95 transition-transform" style={{ height: 56, borderRadius: 9999, background: isValid ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "var(--echo-bg-surface)", color: isValid ? "white" : "var(--echo-text-tertiary)", fontSize: 17, fontWeight: 700, cursor: isValid ? "pointer" : "not-allowed", boxShadow: isValid ? "var(--echo-shadow-primary)" : "none" }}>
          <span>{submitting ? "Sending..." : "Send Verification Code"}</span>
          {!submitting && <ArrowRight size={20} />}
        </motion.button>

        <div className="flex justify-center w-full mt-4">
          <button className="bg-transparent border-none cursor-pointer" style={{ fontSize: 13, fontWeight: 700, color: "var(--echo-text-tertiary)", textTransform: "uppercase" as const, letterSpacing: "0.1em" }}>Why do we need this?</button>
        </div>

        <div className="grid grid-cols-2 gap-4 w-full mt-auto pt-10">
          {[{ icon: Lock, title: "End-to-End", body: "Standardized zero-knowledge proof verification." }, { icon: Fingerprint, title: "Privacy First", body: "Your number is transformed into a cryptographic hash." }].map((f, i) => (
            <div key={i} className="flex flex-col gap-2 p-5" style={{ borderRadius: 24, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(24px)", border: "1px solid var(--echo-ghost-border)" }}>
              <f.icon size={20} style={{ color: "var(--echo-accent)" }} />
              <h3 style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-text-primary)", margin: 0 }}>{f.title}</h3>
              <p style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: 1.4, margin: 0 }}>{f.body}</p>
            </div>
          ))}
        </div>
      </main>

      <footer className="fixed bottom-0 left-0 w-full flex justify-around items-center px-8 pb-4 z-50" style={{ height: 72, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(32px)", WebkitBackdropFilter: "blur(32px)", borderTop: "1px solid var(--echo-ghost-border)" }}>
        <div className="flex items-center justify-center p-3" style={{ color: "var(--echo-text-tertiary)" }}><ShieldCheck size={22} /></div>
        <div className="flex items-center justify-center p-3 rounded-full" style={{ backgroundColor: "var(--echo-accent-subtle)", color: "var(--echo-accent)" }}><Fingerprint size={22} /></div>
        <div className="flex items-center justify-center p-3" style={{ color: "var(--echo-text-tertiary)" }}><Lock size={22} /></div>
      </footer>
    </div>
  );
}
