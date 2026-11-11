import React, { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { Lock, LifeBuoy } from "lucide-react";
import { motion } from "motion/react";

type LockReason = "attempts" | "suspicious" | "suspended";

const subtitleMap: Record<LockReason, string> = {
  attempts: "Too many failed attempts. Try again in:",
  suspicious: "Unusual activity detected. Try again in:",
  suspended: "Your account has been suspended. Contact support for help.",
};

export function AccountLockedPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const reason = (searchParams.get("reason") as LockReason) || "attempts";
  const initialSeconds = parseInt(searchParams.get("duration") || "2852", 10);
  const [remaining, setRemaining] = useState(initialSeconds);
  const [unlocked, setUnlocked] = useState(false);

  useEffect(() => {
    if (reason === "suspended") return;
    const timer = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          setUnlocked(true);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(timer);
  }, [reason]);

  const minutes = Math.floor(remaining / 60);
  const seconds = remaining % 60;

  const handleTryAgain = () => {
    navigate("/onboarding/login");
  };

  return (
    <div className="flex flex-col h-screen px-6">
      <div className="flex-1 flex flex-col items-center justify-center">
        {/* Lock icon */}
        <motion.div
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.4, type: "spring" }}
          className="flex items-center justify-center rounded-full mb-8"
          style={{
            width: 80,
            height: 80,
            backgroundColor: "var(--echo-bg-surface)",
          }}
        >
          <Lock size={40} style={{ color: "var(--echo-text-tertiary)" }} />
        </motion.div>

        {/* Title */}
        <motion.h1
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15, duration: 0.3 }}
          style={{
            fontSize: 28,
            fontWeight: 700,
            color: "var(--echo-text-primary)",
            lineHeight: "36px",
            letterSpacing: "-0.5px",
            margin: "0 0 12px",
            textAlign: "center",
          }}
        >
          Account locked
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.25, duration: 0.3 }}
          style={{
            fontSize: 16,
            color: "var(--echo-text-secondary)",
            lineHeight: "24px",
            textAlign: "center",
            margin: "0 0 24px",
          }}
        >
          {subtitleMap[reason]}
        </motion.p>

        {/* Countdown */}
        {reason !== "suspended" && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.35, duration: 0.3 }}
          >
            {unlocked ? (
              <motion.button
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                onClick={handleTryAgain}
                className="border-none cursor-pointer rounded-xl px-8 active:scale-[0.98] transition-transform"
                style={{
                  height: 48,
                  background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                  fontSize: 16,
                  fontWeight: 600,
                  color: "white",
                  boxShadow: "var(--echo-shadow-primary)",
                }}
              >
                Try Again
              </motion.button>
            ) : (
              <span
                style={{
                  fontSize: 48,
                  fontWeight: 700,
                  color: "var(--echo-primary)",
                  fontVariantNumeric: "tabular-nums",
                }}
              >
                {String(minutes).padStart(2, "0")}:{String(seconds).padStart(2, "0")}
              </span>
            )}
          </motion.div>
        )}
      </div>

      {/* Bottom actions */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5, duration: 0.3 }}
        className="flex flex-col items-center gap-5 pb-12"
      >
        <div style={{ width: "100%", height: 1, backgroundColor: "var(--echo-border)" }} />

        <p
          style={{
            fontSize: 14,
            color: "var(--echo-text-secondary)",
            textAlign: "center",
            lineHeight: "20px",
            margin: 0,
          }}
        >
          If this wasn't you,{" "}
          <button
            onClick={() => navigate("/onboarding/recover")}
            className="bg-transparent border-none cursor-pointer p-0"
            style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}
          >
            recover your account
          </button>{" "}
          to secure it.
        </p>

        <button
          className="w-full cursor-pointer rounded-xl flex items-center justify-center gap-2 active:scale-[0.98] transition-transform"
          style={{
            height: 48,
            backgroundColor: "transparent",
            border: "2px solid var(--echo-border)",
            fontSize: 16,
            fontWeight: 600,
            color: "var(--echo-text-secondary)",
          }}
        >
          <LifeBuoy size={18} />
          Contact Support
        </button>
      </motion.div>
    </div>
  );
}
