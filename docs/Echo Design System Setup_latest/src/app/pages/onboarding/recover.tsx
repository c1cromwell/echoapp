import React from "react";
import { useNavigate } from "react-router";
import { ArrowLeft } from "lucide-react";
import { motion } from "motion/react";

const methods = [
  {
    id: "phrase",
    emoji: "\uD83D\uDD11",
    title: "Recovery Phrase",
    subtitle: "Enter your 12-word backup phrase",
  },
  {
    id: "contacts",
    emoji: "\uD83D\uDC65",
    title: "Trusted Contacts",
    subtitle: "Ask friends to help verify your identity",
  },
  {
    id: "phone",
    emoji: "\uD83D\uDCF1",
    title: "Phone Verification",
    subtitle: "Verify via your registered phone",
  },
];

export function RecoverMethodPage() {
  const navigate = useNavigate();

  const handleSelect = (id: string) => {
    if (id === "contacts") {
      navigate("/onboarding/recover-contacts");
    } else if (id === "phone") {
      navigate("/onboarding/otp?context=recovery");
    } else {
      navigate("/onboarding/recover-phrase");
    }
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <div className="flex items-center px-4 pt-14 pb-2">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
      </div>

      <div className="flex-1 px-6">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <h1
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "36px",
              letterSpacing: "-0.5px",
              margin: "0 0 8px",
            }}
          >
            Recover your account
          </h1>
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              lineHeight: "24px",
              margin: "0 0 32px",
            }}
          >
            Choose how to verify your identity.
          </p>
        </motion.div>

        <div className="flex flex-col gap-3">
          {methods.map((method, i) => (
            <motion.button
              key={method.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 + i * 0.08, duration: 0.3 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => handleSelect(method.id)}
              className="flex items-center gap-4 w-full rounded-2xl px-4 py-4 cursor-pointer text-left transition-all"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                border: "1px solid var(--echo-border)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <span style={{ fontSize: 32, lineHeight: 1 }}>{method.emoji}</span>
              <div className="flex-1">
                <div style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                  {method.title}
                </div>
                <div style={{ fontSize: 14, color: "var(--echo-text-secondary)", marginTop: 2 }}>
                  {method.subtitle}
                </div>
              </div>
            </motion.button>
          ))}
        </div>
      </div>
    </div>
  );
}