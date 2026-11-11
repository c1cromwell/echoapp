import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ScanFace, Fingerprint, Check, Loader2 } from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import { motion } from "motion/react";

// Simulate device biometric detection
const biometricType: "faceId" | "touchId" = "faceId";

export function PasskeyPage() {
  const navigate = useNavigate();
  const { update } = useOnboarding();
  const [enabling, setEnabling] = useState(false);

  const handleEnable = () => {
    setEnabling(true);
    setTimeout(() => {
      update({ passkeyEnabled: true });
      navigate("/onboarding/recovery");
    }, 2000);
  };

  const handleSkip = () => {
    update({ passkeySkipped: true });
    navigate("/onboarding/recovery");
  };

  const BiometricIcon = biometricType === "faceId" ? ScanFace : Fingerprint;
  const biometricLabel = biometricType === "faceId" ? "Face ID" : "Touch ID";

  return (
    <div className="flex flex-col h-screen px-6">
      <div className="flex-1 flex flex-col items-center justify-center">
        {/* Icon */}
        <motion.div
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.4, type: "spring" }}
          className="flex items-center justify-center rounded-[28px] mb-8"
          style={{
            width: 100,
            height: 100,
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
          }}
        >
          <motion.div
            animate={{ scale: [1, 1.05, 1] }}
            transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
          >
            <BiometricIcon size={52} color="white" strokeWidth={1.5} />
          </motion.div>
        </motion.div>

        {/* Text */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.4 }}
          className="text-center"
        >
          <h1
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "36px",
              letterSpacing: "-0.5px",
              margin: "0 0 12px",
            }}
          >
            Secure your account
          </h1>
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              lineHeight: "24px",
              maxWidth: 300,
              margin: "0 auto",
            }}
          >
            Sign in with your face — no passwords, no codes. Your biometric data never leaves your device.
          </p>
        </motion.div>
      </div>

      {/* Actions */}
      <div className="flex flex-col items-center gap-3 pb-6">
        <button
          onClick={handleEnable}
          disabled={enabling}
          className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center gap-2 active:scale-[0.98] transition-transform"
          style={{
            height: 56,
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            fontSize: 16,
            fontWeight: 600,
            color: "white",
            boxShadow: "var(--echo-shadow-primary)",
          }}
        >
          {enabling ? (
            <>
              <Loader2 size={20} className="animate-spin" />
              Securing your identity...
            </>
          ) : (
            <>
              <BiometricIcon size={20} color="white" />
              Enable {biometricLabel}
            </>
          )}
        </button>

        <button
          onClick={handleSkip}
          className="bg-transparent border-none cursor-pointer py-3"
          style={{
            fontSize: 14,
            fontWeight: 500,
            color: "var(--echo-text-tertiary)",
          }}
        >
          Skip for now
        </button>
      </div>

      {/* Benefits */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5, duration: 0.4 }}
        className="flex flex-col gap-3 pb-10 pt-4"
        style={{ borderTop: "1px solid var(--echo-border-light)" }}
      >
        {[
          "No passwords to remember",
          "Your face stays on-device",
          "Move to a new phone easily",
        ].map((text) => (
          <div key={text} className="flex items-center gap-3">
            <div
              className="flex items-center justify-center rounded-full shrink-0"
              style={{
                width: 24,
                height: 24,
                backgroundColor: "var(--echo-success-light)",
              }}
            >
              <Check size={14} style={{ color: "var(--echo-success)" }} strokeWidth={2.5} />
            </div>
            <span
              style={{
                fontSize: 14,
                color: "var(--echo-text-secondary)",
                lineHeight: "20px",
              }}
            >
              {text}
            </span>
          </div>
        ))}
      </motion.div>
    </div>
  );
}