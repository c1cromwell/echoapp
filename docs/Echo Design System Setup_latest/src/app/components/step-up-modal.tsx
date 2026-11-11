import React, { useState } from "react";
import { ScanFace, Loader2, Check } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { AuthError } from "./auth-error";

interface StepUpModalProps {
  isOpen: boolean;
  onClose: () => void;
  onVerified: () => void;
  onFallbackOtp?: () => void;
  subtitle?: string;
}

type VerifyState = "default" | "authenticating" | "success" | "failed";

export function StepUpModal({ isOpen, onClose, onVerified, onFallbackOtp, subtitle }: StepUpModalProps) {
  const [state, setState] = useState<VerifyState>("default");
  const [error, setError] = useState("");
  const [attempts, setAttempts] = useState(0);

  const handleVerify = () => {
    const newAttempts = attempts + 1;
    setAttempts(newAttempts);
    setError("");
    setState("authenticating");

    setTimeout(() => {
      // Simulate failure on 2nd attempt for demo
      if (newAttempts === 2) {
        setState("failed");
        setError("Couldn't verify your identity. Try again.");
        return;
      }

      setState("success");
      setTimeout(() => {
        setState("default");
        setAttempts(0);
        setError("");
        onVerified();
      }, 500);
    }, 1200);
  };

  const handleFallback = () => {
    if (onFallbackOtp) {
      onFallbackOtp();
    }
  };

  const handleClose = () => {
    setState("default");
    setError("");
    setAttempts(0);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-end justify-center"
        style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
        onClick={handleClose}
      >
        <motion.div
          initial={{ y: "100%" }}
          animate={{ y: 0 }}
          exit={{ y: "100%" }}
          transition={{ type: "spring", damping: 25, stiffness: 300 }}
          onClick={(e) => e.stopPropagation()}
          className="w-full rounded-t-3xl px-6 pb-10 pt-3"
          style={{
            backgroundColor: "var(--echo-bg-primary)",
            maxHeight: "45%",
          }}
        >
          {/* Drag indicator */}
          <div className="flex justify-center mb-6">
            <div
              className="rounded-full"
              style={{
                width: 36,
                height: 5,
                backgroundColor: "var(--echo-border)",
              }}
            />
          </div>

          {/* Title */}
          <h2
            style={{
              fontSize: 24,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              margin: "0 0 8px",
              textAlign: "center",
            }}
          >
            Verify it's you
          </h2>

          {/* Subtitle */}
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              margin: "0 0 20px",
              textAlign: "center",
              lineHeight: "24px",
            }}
          >
            {subtitle || "This action requires additional verification."}
          </p>

          {/* Error message */}
          <div className="flex justify-center mb-3">
            <AuthError message={error} onDismiss={() => setError("")} autoDismissMs={5000} />
          </div>

          {/* Verify button */}
          <button
            onClick={handleVerify}
            disabled={state === "authenticating" || state === "success"}
            className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center gap-3 mb-4 transition-all"
            style={{
              height: 48,
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              fontSize: 16,
              fontWeight: 600,
              color: "white",
              opacity: state === "authenticating" ? 0.8 : 1,
            }}
          >
            {state === "authenticating" && <Loader2 size={22} className="animate-spin" color="white" />}
            {state === "success" && <Check size={24} color="white" strokeWidth={3} />}
            {(state === "default" || state === "failed") && (
              <>
                <ScanFace size={20} color="white" />
                {state === "failed" ? "Try again" : "Verify now"}
              </>
            )}
          </button>

          {/* Fallback */}
          <button
            onClick={handleFallback}
            className="w-full bg-transparent border-none cursor-pointer py-2"
            style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}
          >
            Use verification code
          </button>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
