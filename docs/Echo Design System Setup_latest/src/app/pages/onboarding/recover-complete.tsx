import React from "react";
import { useNavigate } from "react-router";
import { CircleCheck, ScanFace } from "lucide-react";
import { motion } from "motion/react";

export function RecoverCompletePage() {
  const navigate = useNavigate();

  const handleSetup = () => {
    navigate("/onboarding/passkey");
  };

  return (
    <div className="flex flex-col h-screen px-6">
      <div className="flex-1 flex flex-col items-center justify-center">
        {/* Success icon */}
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: "spring", stiffness: 200, damping: 15 }}
          className="flex items-center justify-center rounded-full mb-8"
          style={{
            width: 80,
            height: 80,
            backgroundColor: "var(--echo-success-light)",
          }}
        >
          <CircleCheck size={48} style={{ color: "var(--echo-success)" }} />
        </motion.div>

        {/* Title */}
        <motion.h1
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.3 }}
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
          Account recovered!
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3, duration: 0.3 }}
          style={{
            fontSize: 16,
            color: "var(--echo-text-secondary)",
            lineHeight: "24px",
            textAlign: "center",
            maxWidth: 280,
            margin: 0,
          }}
        >
          Set up a new passkey to secure your account.
        </motion.p>
      </div>

      {/* CTA */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4, duration: 0.3 }}
        className="pb-12"
      >
        <button
          onClick={handleSetup}
          className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center gap-3 active:scale-[0.98] transition-transform"
          style={{
            height: 56,
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            fontSize: 16,
            fontWeight: 600,
            color: "white",
            boxShadow: "var(--echo-shadow-primary)",
          }}
        >
          <ScanFace size={20} color="white" />
          Set Up Now
        </button>
      </motion.div>
    </div>
  );
}
