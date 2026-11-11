import React, { useState, useRef } from "react";
import { useNavigate } from "react-router";
import { Shield, Lock, Fingerprint, Key, Coins, ArrowRight, ShieldCheck, Users } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import echoWhiteMono from "../../../imports/echo-ripple-white-mono-1.svg";
import echoBrandLogo from "../../../imports/echo-ripple-brand-1.svg";

const slides = [
  {
    id: 0,
    icon: "shield",
    title: "Always Private. You Own Your Data.",
    body: "End-to-end encrypted by default. Zero metadata. No tracking, no profiling — not even ECHO can read your messages.",
  },
  {
    id: 1,
    icon: "trust",
    title: "Community-Owned. Earned Rewards.",
    body: "No central servers or Big Tech middlemen. Build your reputation through trusted connections and earn rewards along the way.",
  },
];

function HeroVisual({ slideIndex, active }: { slideIndex: number; active: boolean }) {
  const icons = [ShieldCheck, Users];
  const Icon = icons[slideIndex % icons.length];

  return (
    <motion.div
      initial={{ scale: 0.9, opacity: 0 }}
      animate={active ? { scale: 1, opacity: 1 } : { scale: 0.9, opacity: 0 }}
      transition={{ duration: 0.5, type: "spring", damping: 22 }}
      className="relative flex items-center justify-center"
      style={{ width: "100%", maxWidth: 340, aspectRatio: "4/3" }}
    >
      <div
        className="w-full h-full flex items-center justify-center overflow-hidden relative"
        style={{ borderRadius: 40, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, boxShadow: "var(--echo-shadow-xl)" }}
      >
        <div className="absolute top-0 left-0 w-full" style={{ height: 3, backgroundColor: "var(--echo-secure-color)", boxShadow: "var(--echo-secure-glow)", opacity: 0.8 }} />
        <div className="relative flex items-center justify-center">
          <motion.div className="absolute rounded-full" style={{ width: 160, height: 160, border: "2px solid rgba(255,255,255,0.15)" }} animate={active ? { scale: [1, 1.02, 1] } : {}} transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }} />
          <motion.div className="absolute rounded-full" style={{ width: 110, height: 110, border: "2px solid rgba(255,255,255,0.3)" }} animate={active ? { scale: [1, 1.03, 1] } : {}} transition={{ duration: 2.5, repeat: Infinity, ease: "easeInOut", delay: 0.3 }} />
          <Icon size={56} color="white" strokeWidth={1.5} />
        </div>
        <div className="absolute rounded-full" style={{ top: "15%", right: "15%", width: 28, height: 28, backgroundColor: "rgba(125,211,252,0.5)", filter: "blur(4px)" }} />
        <div className="absolute rounded-full" style={{ bottom: "20%", left: "18%", width: 20, height: 20, backgroundColor: "rgba(255,255,255,0.4)", filter: "blur(3px)" }} />
      </div>
    </motion.div>
  );
}

export function IntroCarousel() {
  const navigate = useNavigate();
  const [current, setCurrent] = useState(0);
  const touchStartX = useRef(0);
  const touchEndX = useRef(0);

  const handleSwipe = () => {
    const diff = touchStartX.current - touchEndX.current;
    if (Math.abs(diff) > 50) {
      if (diff > 0 && current < slides.length - 1) setCurrent(current + 1);
      if (diff < 0 && current > 0) setCurrent(current - 1);
    }
  };

  const goToWelcome = () => navigate("/onboarding/login");
  const isLast = current === slides.length - 1;

  return (
    <div
      className="relative flex flex-col min-h-screen overflow-hidden"
      style={{ backgroundColor: "var(--echo-bg-primary)" }}
      onTouchStart={(e) => (touchStartX.current = e.touches[0].clientX)}
      onTouchEnd={(e) => { touchEndX.current = e.changedTouches[0].clientX; handleSwipe(); }}
    >
      <div className="absolute pointer-events-none" style={{ top: "-10%", left: "50%", transform: "translateX(-50%)", width: "80%", height: "40%", background: "var(--echo-sky)", opacity: 0.06, borderRadius: "50%", filter: "blur(120px)" }} />

      <div className="flex justify-center px-6 pt-14 pb-4 relative z-10">
        <div className="flex items-center gap-2.5">
          <img src={echoBrandLogo} alt="Echo" width={28} height={28} />
          <span style={{ fontSize: 22, fontWeight: 800, color: "var(--echo-text-primary)", letterSpacing: "-0.02em" }}>ECHO</span>
        </div>
      </div>

      <main className="flex-1 flex flex-col items-center justify-center px-6 relative z-10">
        <div className="w-full flex justify-center mb-10">
          <AnimatePresence mode="wait">
            <HeroVisual key={current} slideIndex={current} active />
          </AnimatePresence>
        </div>
        <AnimatePresence mode="wait">
          <motion.div key={current} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -20 }} transition={{ duration: 0.35, ease: "easeOut" }} className="flex flex-col items-center text-center">
            <h2 style={{ fontSize: 36, fontWeight: 700, color: "var(--echo-text-primary)", lineHeight: 1.1, letterSpacing: "-0.02em", margin: "0 0 12px", maxWidth: 360 }}>{slides[current].title}</h2>
            <p style={{ fontSize: 15, color: "var(--echo-text-secondary)", lineHeight: 1.6, maxWidth: 320, margin: 0, opacity: 0.8 }}>{slides[current].body}</p>
          </motion.div>
        </AnimatePresence>
      </main>

      <div className="flex flex-col items-center gap-6 px-6 pb-10 relative z-10">
        <button onClick={isLast ? goToWelcome : () => setCurrent(current + 1)} className="w-full border-none cursor-pointer flex items-center justify-center gap-2 active:scale-95 transition-transform" style={{ height: 56, maxWidth: 320, borderRadius: 9999, background: `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))`, fontSize: 17, fontWeight: 700, color: "white", boxShadow: "var(--echo-shadow-primary)" }}>
          <span>Get Started</span>
          <ArrowRight size={20} />
        </button>

        <div className="flex items-center gap-2.5">
          {slides.map((_, i) => (
            <button key={i} onClick={() => setCurrent(i)} className="border-none cursor-pointer p-0 rounded-full transition-all duration-300" style={{ width: i === current ? 28 : 8, height: 8, borderRadius: 4, background: i === current ? `linear-gradient(135deg, var(--echo-sig-from), var(--echo-sig-to))` : "var(--echo-bg-surface)" }} aria-label={`Slide ${i + 1}`} />
          ))}
        </div>

        <button onClick={goToWelcome} className="bg-transparent border-none cursor-pointer" style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-accent)", textTransform: "uppercase" as const, letterSpacing: "0.1em" }}>
          Already have an account?
        </button>
      </div>
    </div>
  );
}