import React, { useState, useCallback, useRef } from "react";
import { useNavigate } from "react-router";
import { Fingerprint, Loader2, Check, ShieldAlert, ChevronDown, ChevronRight, MessageSquare, UserPlus } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { NewDeviceAlert } from "../../components/new-device-alert";
import { AuthError } from "../../components/auth-error";
import echoBrandLogo from "../../../imports/echo-ripple-brand-1.svg";
import echoWhiteMono from "../../../imports/echo-ripple-white-mono-1.svg";

type LoginState = "default" | "authenticating" | "success" | "failed";
const SIMULATE_NEW_DEVICE = true;

const countryCodes = [
  { code: "+1", flag: "\u{1F1FA}\u{1F1F8}", label: "US" },
  { code: "+44", flag: "\u{1F1EC}\u{1F1E7}", label: "UK" },
  { code: "+91", flag: "\u{1F1EE}\u{1F1F3}", label: "IN" },
  { code: "+49", flag: "\u{1F1E9}\u{1F1EA}", label: "DE" },
  { code: "+81", flag: "\u{1F1EF}\u{1F1F5}", label: "JP" },
];

export function LoginPage() {
  const navigate = useNavigate();
  const [state, setState] = useState<LoginState>("default");
  const [errorMsg, setErrorMsg] = useState("");
  const [attempts, setAttempts] = useState(0);
  const attemptsRef = useRef(0);
  const [showNewDevice, setShowNewDevice] = useState(false);
  const [showBlockedMsg, setShowBlockedMsg] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const [selectedCountry, setSelectedCountry] = useState(countryCodes[0]);
  const [showCountryPicker, setShowCountryPicker] = useState(false);
  const [showSmsSection, setShowSmsSection] = useState(false);
  const maxAttempts = 5;

  const clearError = useCallback(() => { if (errorMsg) setErrorMsg(""); }, [errorMsg]);

  const handlePasskeyLogin = () => {
    if (attemptsRef.current >= maxAttempts) { navigate("/onboarding/account-locked"); return; }
    setState("authenticating");
    attemptsRef.current += 1;
    setAttempts(attemptsRef.current);
    const currentAttempt = attemptsRef.current;
    setTimeout(() => {
      if (currentAttempt === 2) { setState("failed"); setErrorMsg("Couldn't verify. Try again or use SMS code."); return; }
      if (currentAttempt === 1 && SIMULATE_NEW_DEVICE) { setState("default"); setShowNewDevice(true); return; }
      setState("success");
      setTimeout(() => { localStorage.setItem("echo-onboarded", "true"); window.location.href = "/"; }, 600);
    }, 1500);
  };

  const handleSmsLogin = () => { clearError(); if (!inputValue.trim()) { setErrorMsg("Enter a valid phone number."); return; } navigate("/onboarding/otp"); };

  const handleNewDeviceVerify = () => {
    setShowNewDevice(false); setState("authenticating");
    setTimeout(() => { setState("success"); setTimeout(() => { localStorage.setItem("echo-onboarded", "true"); window.location.href = "/"; }, 600); }, 800);
  };

  const handleNewDeviceReject = () => { setShowNewDevice(false); setShowBlockedMsg(true); setTimeout(() => setShowBlockedMsg(false), 3500); };

  const remainingAttempts = maxAttempts - attempts;
  const showAttemptsWarning = attempts >= 3 && attempts < maxAttempts;

  return (
    <div className="relative flex flex-col min-h-screen overflow-hidden" style={{ backgroundColor: "var(--echo-bg-primary)" }}>
      <motion.div className="fixed top-0 left-0 w-full z-[60]" style={{ height: 2, backgroundColor: "var(--echo-secure-color)", filter: "blur(0.5px)" }} animate={{ opacity: [0.6, 1, 0.6] }} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }} />

      <div className="absolute pointer-events-none" style={{ top: "-10%", left: "-10%", width: "40%", height: "40%", background: "var(--echo-sky)", opacity: 0.06, borderRadius: "50%", filter: "blur(120px)" }} />
      <div className="absolute pointer-events-none" style={{ bottom: "-10%", right: "-10%", width: "30%", height: "30%", background: "var(--echo-sky-light)", opacity: 0.06, borderRadius: "50%", filter: "blur(100px)" }} />

      <header className="fixed top-0 w-full z-50 flex items-center justify-between px-6" style={{ height: 64, backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(24px)", WebkitBackdropFilter: "blur(24px)", boxShadow: "var(--echo-shadow-sm)", borderBottom: "1px solid var(--echo-ghost-border)" }}>
        <div className="flex items-center gap-2">
          <img src={echoBrandLogo} alt="Echo" width={32} height={32} />
          <span style={{ fontSize: 22, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--echo-accent)" }}>ECHO</span>
        </div>
      </header>

      <main className="flex-1 flex flex-col items-center justify-center px-4 pt-24 pb-28 relative z-10">
        <div className="w-full" style={{ maxWidth: 400 }}>
          <div className="flex flex-col gap-6">
            <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.4 }} className="text-center mb-2">
              <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-tertiary)", letterSpacing: "0.02em" }}>Private Messaging, Always</p>
            </motion.div>

            {/* Passkey Login */}
            <motion.button initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5, type: "spring", damping: 25 }} whileTap={{ scale: 0.95 }} onClick={handlePasskeyLogin} disabled={state === "authenticating" || state === "success"} className="relative w-full p-[1px] border-none cursor-pointer" style={{ borderRadius: 32, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, boxShadow: "var(--echo-shadow-primary)" }}>
              <div className="flex items-center justify-between p-6" style={{ borderRadius: 31, backgroundColor: "rgba(255,255,255,0.05)", backdropFilter: "blur(8px)" }}>
                <div className="flex items-center gap-4">
                  <div className="flex items-center justify-center rounded-full" style={{ width: 56, height: 56, backgroundColor: "rgba(255,255,255,0.1)", border: "1px solid rgba(255,255,255,0.2)" }}>
                    <AnimatePresence mode="wait">
                      {state === "authenticating" ? (<motion.div key="loading" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}><Loader2 size={28} className="animate-spin" color="white" /></motion.div>)
                      : state === "success" ? (<motion.div key="success" initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: "spring", stiffness: 300 }}><Check size={28} color="white" strokeWidth={3} /></motion.div>)
                      : (<motion.div key="fp" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}><Fingerprint size={28} color="white" /></motion.div>)}
                    </AnimatePresence>
                  </div>
                  <div className="text-left">
                    <span className="block" style={{ fontSize: 20, fontWeight: 700, color: "white" }}>{state === "success" ? "Welcome Back" : state === "failed" ? "Try Again" : "Login with Passkey"}</span>
                    <span className="block" style={{ fontSize: 14, color: "rgba(186,230,253,0.7)" }}>FaceID, TouchID, or PIN</span>
                  </div>
                </div>
                <ChevronRight size={20} style={{ color: "rgba(255,255,255,0.5)" }} />
              </div>
            </motion.button>

            <div className="w-full"><AuthError message={errorMsg} onDismiss={() => setErrorMsg("")} autoDismissMs={5000} /></div>
            <AnimatePresence>
              {showAttemptsWarning && !errorMsg && (
                <motion.div initial={{ opacity: 0, y: -4 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="flex items-center justify-center gap-1.5">
                  <ShieldAlert size={14} style={{ color: "var(--echo-warning)" }} />
                  <span style={{ fontSize: 13, color: "var(--echo-warning)" }}>{remainingAttempts} attempt{remainingAttempts !== 1 ? "s" : ""} remaining</span>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Divider */}
            <motion.button initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2, duration: 0.4 }} onClick={() => setShowSmsSection(!showSmsSection)} className="flex items-center gap-4 py-2 w-full cursor-pointer border-none bg-transparent">
              <div className="h-[1px] flex-1" style={{ backgroundColor: "var(--echo-ghost-border)" }} />
              <div className="flex items-center gap-2 px-2">
                <span style={{ fontSize: 10, fontWeight: 500, letterSpacing: "0.1em", textTransform: "uppercase" as const, color: "var(--echo-text-tertiary)" }}>Secure Alternative</span>
                <motion.div animate={{ rotate: showSmsSection ? 180 : 0 }} transition={{ duration: 0.2 }}><ChevronDown size={14} style={{ color: "var(--echo-text-tertiary)" }} /></motion.div>
              </div>
              <div className="h-[1px] flex-1" style={{ backgroundColor: "var(--echo-ghost-border)" }} />
            </motion.button>

            {/* SMS Section */}
            <AnimatePresence>
              {showSmsSection && (
                <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: "auto" }} exit={{ opacity: 0, height: 0 }} transition={{ duration: 0.35, type: "spring", damping: 28 }} className="overflow-hidden">
                  <div className="p-8" style={{ backgroundColor: "var(--echo-bg-surface)", borderRadius: 32, backdropFilter: "blur(24px)", border: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-sm)" }}>
                    <label className="block mb-4 px-1" style={{ fontSize: 11, fontWeight: 700, color: "var(--echo-text-tertiary)", textTransform: "uppercase" as const, letterSpacing: "0.08em" }}>SMS Verification</label>
                    <div className="flex flex-col gap-4">
                      <div className="relative">
                        <div className="flex items-center w-full overflow-hidden transition-all" style={{ borderRadius: 16, backgroundColor: "var(--echo-bg-elevated)", border: isFocused ? "1px solid var(--echo-sky-ring)" : "1px solid var(--echo-ghost-border)" }}>
                          <button onClick={(e) => { e.stopPropagation(); setShowCountryPicker(!showCountryPicker); }} className="flex items-center gap-1 shrink-0 px-4 bg-transparent border-none cursor-pointer" style={{ height: 52, borderRight: "1px solid var(--echo-ghost-border)" }}>
                            <span style={{ fontSize: 16 }}>{selectedCountry.flag}</span>
                            <span style={{ fontSize: 14, color: "var(--echo-text-primary)", fontWeight: 500 }}>{selectedCountry.code}</span>
                            <ChevronDown size={12} style={{ color: "var(--echo-text-tertiary)" }} />
                          </button>
                          <input type="tel" value={inputValue} onChange={(e) => { setInputValue(e.target.value); clearError(); }} onFocus={() => setIsFocused(true)} onBlur={() => setIsFocused(false)} placeholder="+1 (555) 000-0000" className="flex-1 px-4 bg-transparent border-none outline-none" style={{ height: 52, fontSize: 16, color: "var(--echo-text-primary)" }} />
                        </div>
                        <AnimatePresence>
                          {showCountryPicker && (
                            <motion.div initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }} className="absolute w-full mt-1 z-20 overflow-hidden" style={{ borderRadius: 16, backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-lg)", border: "1px solid var(--echo-ghost-border)" }}>
                              {countryCodes.map((cc) => (
                                <button key={cc.code} onClick={(e) => { e.stopPropagation(); setSelectedCountry(cc); setShowCountryPicker(false); }} className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer transition-colors">
                                  <span style={{ fontSize: 18 }}>{cc.flag}</span>
                                  <span style={{ fontSize: 14, color: "var(--echo-text-primary)" }}>{cc.label} ({cc.code})</span>
                                </button>
                              ))}
                            </motion.div>
                          )}
                        </AnimatePresence>
                      </div>
                      <button onClick={handleSmsLogin} className="w-full flex items-center justify-center gap-2 border-none cursor-pointer active:scale-[0.98] transition-transform" style={{ height: 52, borderRadius: 16, backgroundColor: "var(--echo-bg-surface)", color: "var(--echo-text-secondary)", fontSize: 16, fontWeight: 700, border: "1px solid var(--echo-ghost-border)" }}>
                        <span>Send Code</span><MessageSquare size={18} />
                      </button>
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.4, duration: 0.4 }} className="text-center pt-2">
              <p style={{ fontSize: 15, color: "var(--echo-text-secondary)", fontWeight: 500 }}>
                New to ECHO?{" "}
                <button onClick={() => navigate("/onboarding/phone")} className="bg-transparent border-none cursor-pointer p-0 ml-1" style={{ fontSize: 15, fontWeight: 700, color: "var(--echo-accent)" }}>Get Started</button>
              </p>
            </motion.div>
          </div>
        </div>
      </main>

      <nav className="fixed bottom-0 left-0 w-full flex justify-around items-center px-8 pb-8 pt-4 z-50" style={{ backgroundColor: "var(--echo-glass-bg)", backdropFilter: "blur(32px)", WebkitBackdropFilter: "blur(32px)", borderTop: "1px solid var(--echo-ghost-border)", boxShadow: "var(--echo-shadow-md)", borderTopLeftRadius: 32, borderTopRightRadius: 32 }}>
        <button onClick={handlePasskeyLogin} className="flex flex-col items-center justify-center border-none cursor-pointer px-6 py-2 active:scale-90 transition-transform" style={{ borderRadius: 9999, backgroundColor: "var(--echo-accent)", color: "white" }}>
          <Fingerprint size={20} /><span style={{ fontSize: 12, fontWeight: 500, letterSpacing: "0.02em" }}>Secure Login</span>
        </button>
        <button onClick={() => navigate("/onboarding/phone")} className="flex flex-col items-center justify-center bg-transparent border-none cursor-pointer px-6 py-2 transition-colors" style={{ color: "var(--echo-text-tertiary)" }}>
          <UserPlus size={20} /><span style={{ fontSize: 12, fontWeight: 500, letterSpacing: "0.02em" }}>Get Started</span>
        </button>
      </nav>

      <NewDeviceAlert isOpen={showNewDevice} onVerify={handleNewDeviceVerify} onReject={handleNewDeviceReject} deviceName="iPhone 16 Pro" deviceOS="iOS 18.0" location="Chicago, IL" />

      <AnimatePresence>
        {showBlockedMsg && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="fixed inset-0 z-50 flex items-center justify-center px-10" style={{ backgroundColor: "rgba(0,0,0,0.5)" }}>
            <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.9, opacity: 0 }} className="w-full p-6 text-center" style={{ maxWidth: 320, borderRadius: 24, backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-xl)" }}>
              <div className="flex justify-center mb-4">
                <div className="flex items-center justify-center rounded-full" style={{ width: 48, height: 48, backgroundColor: "var(--echo-success-light)" }}><Check size={24} style={{ color: "var(--echo-success)" }} /></div>
              </div>
              <h3 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 8px" }}>Device blocked</h3>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: 0, lineHeight: "20px" }}>We've blocked this device. Consider reviewing your trusted contacts and recovery settings.</p>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
