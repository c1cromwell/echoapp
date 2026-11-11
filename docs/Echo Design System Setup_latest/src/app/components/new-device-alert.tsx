import React, { useState } from "react";
import { TriangleAlert, Loader2, Check, Smartphone } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

interface NewDeviceAlertProps {
  isOpen: boolean;
  onVerify: () => void;
  onReject: () => void;
  deviceName?: string;
  deviceOS?: string;
  location?: string;
}

export function NewDeviceAlert({
  isOpen,
  onVerify,
  onReject,
  deviceName = "iPhone 16 Pro",
  deviceOS = "iOS 18.0",
  location = "Chicago, IL",
}: NewDeviceAlertProps) {
  const [verifying, setVerifying] = useState(false);

  const handleVerify = () => {
    setVerifying(true);
    setTimeout(() => {
      setVerifying(false);
      onVerify();
    }, 1500);
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center px-10"
        style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
      >
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          transition={{ type: "spring", damping: 20, stiffness: 300 }}
          className="w-full rounded-[20px] p-6"
          style={{
            maxWidth: 320,
            backgroundColor: "var(--echo-bg-elevated)",
            boxShadow: "0 8px 32px rgba(0,0,0,0.15)",
          }}
        >
          {/* Warning icon */}
          <div className="flex justify-center mb-4">
            <div
              className="flex items-center justify-center rounded-full"
              style={{
                width: 48,
                height: 48,
                backgroundColor: "var(--echo-warning-light)",
              }}
            >
              <TriangleAlert size={24} style={{ color: "var(--echo-warning)" }} />
            </div>
          </div>

          {/* Title */}
          <h3
            style={{
              fontSize: 20,
              fontWeight: 600,
              color: "var(--echo-text-primary)",
              margin: "0 0 8px",
              textAlign: "center",
            }}
          >
            New device
          </h3>

          {/* Body */}
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              margin: "0 0 16px",
              textAlign: "center",
              lineHeight: "24px",
            }}
          >
            We noticed you're signing in from a new device.
          </p>

          {/* Device info card */}
          <div
            className="flex items-center gap-3 rounded-xl px-4 py-3 mb-5"
            style={{ backgroundColor: "var(--echo-bg-surface)" }}
          >
            <Smartphone size={24} style={{ color: "var(--echo-text-secondary)", flexShrink: 0 }} />
            <div>
              <div style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                {deviceName}
              </div>
              <div style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>{deviceOS}</div>
              <div style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>{location}</div>
            </div>
          </div>

          {/* Verify button */}
          <button
            onClick={handleVerify}
            disabled={verifying}
            className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center gap-2 mb-3 transition-all"
            style={{
              height: 48,
              background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              fontSize: 16,
              fontWeight: 600,
              color: "white",
            }}
          >
            {verifying ? (
              <Loader2 size={20} className="animate-spin" color="white" />
            ) : (
              "Verify & Add"
            )}
          </button>

          {/* Reject */}
          <button
            onClick={onReject}
            className="w-full bg-transparent border-none cursor-pointer py-2"
            style={{
              fontSize: 14,
              fontWeight: 600,
              color: "var(--echo-error)",
            }}
          >
            That's not me
          </button>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}