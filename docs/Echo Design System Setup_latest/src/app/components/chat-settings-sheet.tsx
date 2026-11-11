import React from "react";
import {
  X, VolumeX, Timer, Shield, Globe,
  Bell, BellOff, EyeOff, Lock,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

interface ChatSettingsSheetProps {
  open: boolean;
  onClose: () => void;
  contactName: string;
  // State
  silentMode: boolean;
  onToggleSilent: () => void;
  disappearingEnabled: boolean;
  disappearingDuration: number; // seconds
  onSetDisappearing: (seconds: number) => void;
  timezone?: string;
}

const disappearOptions = [
  { label: "Off", value: 0 },
  { label: "30s", value: 30 },
  { label: "5m", value: 300 },
  { label: "1h", value: 3600 },
  { label: "24h", value: 86400 },
  { label: "7d", value: 604800 },
];

export function ChatSettingsSheet({
  open, onClose, contactName,
  silentMode, onToggleSilent,
  disappearingEnabled, disappearingDuration, onSetDisappearing,
  timezone,
}: ChatSettingsSheetProps) {
  return (
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-50"
            style={{ backgroundColor: "rgba(0,0,0,0.4)" }}
          />
          {/* Sheet */}
          <motion.div
            initial={{ y: "100%" }}
            animate={{ y: 0 }}
            exit={{ y: "100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
            style={{
              backgroundColor: "var(--echo-bg-elevated)",
              maxHeight: "80vh",
            }}
          >
            {/* Handle */}
            <div className="flex justify-center pt-3 pb-1">
              <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
            </div>

            {/* Header */}
            <div className="flex items-center justify-between px-5 pb-3">
              <h3 style={{ fontSize: 18, fontWeight: 700, color: "var(--echo-text-primary)", margin: 0 }}>
                Chat Settings
              </h3>
              <button
                onClick={onClose}
                className="flex items-center justify-center rounded-full bg-transparent border-none cursor-pointer"
                style={{ width: 32, height: 32, backgroundColor: "var(--echo-bg-surface)" }}
              >
                <X size={16} style={{ color: "var(--echo-text-secondary)" }} />
              </button>
            </div>

            <div className="overflow-y-auto px-5 pb-8" style={{ maxHeight: "65vh" }}>
              {/* Silent Mode */}
              <div className="mb-5">
                <div className="flex items-center gap-2 mb-2">
                  <VolumeX size={16} style={{ color: "var(--echo-primary)" }} />
                  <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                    Silent Mode
                  </span>
                </div>
                <button
                  onClick={onToggleSilent}
                  className="flex items-center justify-between w-full rounded-2xl px-4 py-3 bg-transparent cursor-pointer"
                  style={{
                    backgroundColor: silentMode ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                    border: `1.5px solid ${silentMode ? "var(--echo-primary)" : "var(--echo-border)"}`,
                  }}
                >
                  <div className="flex items-center gap-3 text-left">
                    <div
                      className="flex items-center justify-center rounded-xl"
                      style={{
                        width: 40,
                        height: 40,
                        backgroundColor: silentMode ? "var(--echo-primary)" : "var(--echo-bg-elevated)",
                      }}
                    >
                      {silentMode ? (
                        <BellOff size={18} color="white" />
                      ) : (
                        <Bell size={18} style={{ color: "var(--echo-text-tertiary)" }} />
                      )}
                    </div>
                    <div>
                      <div style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                        {silentMode ? "Silent mode active" : "Normal notifications"}
                      </div>
                      <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                        {silentMode
                          ? "Messages delivered silently — no sound or banner"
                          : "Tap to send messages without notifications"}
                      </div>
                    </div>
                  </div>
                  <div
                    className="rounded-full shrink-0 transition-all"
                    style={{
                      width: 44,
                      height: 24,
                      backgroundColor: silentMode ? "var(--echo-primary)" : "var(--echo-border)",
                      padding: 2,
                    }}
                  >
                    <div
                      className="rounded-full bg-white transition-all"
                      style={{
                        width: 20,
                        height: 20,
                        transform: silentMode ? "translateX(20px)" : "translateX(0)",
                      }}
                    />
                  </div>
                </button>
              </div>

              {/* Disappearing Messages */}
              <div className="mb-5">
                <div className="flex items-center gap-2 mb-2">
                  <Timer size={16} style={{ color: "var(--echo-warning)" }} />
                  <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                    Disappearing Messages
                  </span>
                </div>
                <div
                  className="rounded-2xl p-3"
                  style={{ backgroundColor: "var(--echo-bg-surface)", border: "1.5px solid var(--echo-border)" }}
                >
                  <div style={{ fontSize: 13, color: "var(--echo-text-secondary)", marginBottom: 10, lineHeight: "18px" }}>
                    New messages will auto-delete after the set duration. Both participants will see the timer.
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {disappearOptions.map((opt) => {
                      const isActive = disappearingDuration === opt.value;
                      return (
                        <button
                          key={opt.value}
                          onClick={() => onSetDisappearing(opt.value)}
                          className="rounded-full px-3.5 py-1.5 border-none cursor-pointer transition-all"
                          style={{
                            backgroundColor: isActive ? "var(--echo-primary)" : "var(--echo-bg-elevated)",
                            border: `1.5px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            fontSize: 13,
                            fontWeight: isActive ? 600 : 400,
                            color: isActive ? "white" : "var(--echo-text-secondary)",
                          }}
                        >
                          {opt.label}
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* Security Info */}
              <div className="mb-5">
                <div className="flex items-center gap-2 mb-2">
                  <Shield size={16} style={{ color: "var(--echo-success)" }} />
                  <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                    Security
                  </span>
                </div>
                <div className="flex flex-col gap-2">
                  {[
                    { icon: Lock, label: "End-to-End Encrypted", value: "Active", color: "var(--echo-success)" },
                    { icon: EyeOff, label: "Screenshot Protection", value: "Enabled", color: "var(--echo-success)" },
                    { icon: Globe, label: "Timezone", value: timezone || "Auto-detected", color: "var(--echo-text-secondary)" },
                  ].map((item) => {
                    const Icon = item.icon;
                    return (
                      <div
                        key={item.label}
                        className="flex items-center gap-3 rounded-xl px-3 py-2.5"
                        style={{ backgroundColor: "var(--echo-bg-surface)" }}
                      >
                        <Icon size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                        <span className="flex-1" style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>
                          {item.label}
                        </span>
                        <span style={{ fontSize: 12, fontWeight: 600, color: item.color }}>
                          {item.value}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Recipient Controls */}
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <Shield size={16} style={{ color: "var(--echo-info)" }} />
                  <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                    {contactName}'s Preferences
                  </span>
                </div>
                <div
                  className="rounded-2xl p-3"
                  style={{ backgroundColor: "var(--echo-bg-surface)" }}
                >
                  <div className="flex flex-col gap-2">
                    {[
                      { label: "Accepts silent messages", value: true },
                      { label: "Accepts scheduled messages", value: true },
                      { label: "Requires approval for non-trusted", value: false },
                    ].map((pref) => (
                      <div key={pref.label} className="flex items-center justify-between">
                        <span style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>
                          {pref.label}
                        </span>
                        <span style={{ fontSize: 11, fontWeight: 600, color: pref.value ? "var(--echo-success)" : "var(--echo-text-tertiary)" }}>
                          {pref.value ? "Yes" : "No"}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}