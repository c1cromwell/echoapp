import React, { useState } from "react";
import {
  X, Clock, Zap, Eye, MessageSquare, Timer, Send, Sparkles,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

type ConditionalTrigger = "time" | "when_online" | "after_read" | "no_response" | "conversation_ended";

interface SmartSuggestion {
  time: string;
  label: string;
  confidence: number; // 0-100
  reason: string;
}

interface ScheduleMessageModalProps {
  open: boolean;
  onClose: () => void;
  onSchedule: (config: {
    trigger: ConditionalTrigger;
    scheduledTime?: string;
    text: string;
  }) => void;
  messageText: string;
  contactName: string;
}

const triggers: { id: ConditionalTrigger; icon: React.ElementType; label: string; description: string }[] = [
  { id: "time", icon: Clock, label: "At a specific time", description: "Schedule delivery for a set date & time" },
  { id: "when_online", icon: Zap, label: "When they're online", description: "Send as soon as they become active" },
  { id: "after_read", icon: Eye, label: "After they read last", description: "Send after they've read your last message" },
  { id: "no_response", icon: Timer, label: "If no response", description: "Send a follow-up if no reply within timeframe" },
  { id: "conversation_ended", icon: MessageSquare, label: "When conversation ends", description: "Send when the current topic wraps up" },
];

// Mock smart suggestions
const mockSuggestions: SmartSuggestion[] = [
  { time: "9:00 AM", label: "Morning peak", confidence: 92, reason: "Typically active and responsive at this time" },
  { time: "12:30 PM", label: "Lunch break", confidence: 78, reason: "Often checks messages during lunch" },
  { time: "6:00 PM", label: "Evening", confidence: 85, reason: "High response rate in the evenings" },
];

const quickTimes = [
  { label: "In 30 min", value: "30min" },
  { label: "In 1 hour", value: "1h" },
  { label: "In 3 hours", value: "3h" },
  { label: "Tomorrow 9 AM", value: "tomorrow9" },
  { label: "Tomorrow noon", value: "tomorrow12" },
];

export function ScheduleMessageModal({
  open, onClose, onSchedule, messageText, contactName,
}: ScheduleMessageModalProps) {
  const [selectedTrigger, setSelectedTrigger] = useState<ConditionalTrigger>("time");
  const [selectedTime, setSelectedTime] = useState<string | null>(null);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const handleSchedule = () => {
    onSchedule({
      trigger: selectedTrigger,
      scheduledTime: selectedTime || undefined,
      text: messageText,
    });
    onClose();
  };

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-50"
            style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
          />
          <motion.div
            initial={{ y: "100%" }}
            animate={{ y: 0 }}
            exit={{ y: "100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
            style={{
              backgroundColor: "var(--echo-bg-elevated)",
              maxHeight: "85vh",
            }}
          >
            {/* Handle */}
            <div className="flex justify-center pt-3 pb-1">
              <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
            </div>

            <div className="flex items-center justify-between px-5 pb-2">
              <h3 style={{ fontSize: 18, fontWeight: 700, color: "var(--echo-text-primary)", margin: 0 }}>
                Schedule Message
              </h3>
              <button
                onClick={onClose}
                className="flex items-center justify-center rounded-full bg-transparent border-none cursor-pointer"
                style={{ width: 32, height: 32, backgroundColor: "var(--echo-bg-surface)" }}
              >
                <X size={16} style={{ color: "var(--echo-text-secondary)" }} />
              </button>
            </div>

            <div className="overflow-y-auto px-5 pb-8" style={{ maxHeight: "70vh" }}>
              {/* Message preview */}
              <div
                className="rounded-xl p-3 mb-4"
                style={{ backgroundColor: "var(--echo-bg-surface)", border: "1px solid var(--echo-border)" }}
              >
                <div style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", marginBottom: 4 }}>
                  Message to {contactName}
                </div>
                <div style={{ fontSize: 14, color: "var(--echo-text-primary)", lineHeight: "20px" }}>
                  {messageText || "Type your message first..."}
                </div>
              </div>

              {/* Conditional Triggers */}
              <div className="mb-4">
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", marginBottom: 8 }}>
                  Delivery Condition
                </div>
                <div className="flex flex-col gap-2">
                  {triggers.map((t) => {
                    const Icon = t.icon;
                    const isActive = selectedTrigger === t.id;
                    return (
                      <button
                        key={t.id}
                        onClick={() => setSelectedTrigger(t.id)}
                        className="flex items-center gap-3 w-full rounded-xl px-3 py-2.5 cursor-pointer text-left border-none transition-all"
                        style={{
                          backgroundColor: isActive ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                          border: `1.5px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                        }}
                      >
                        <div
                          className="flex items-center justify-center rounded-lg shrink-0"
                          style={{
                            width: 36,
                            height: 36,
                            backgroundColor: isActive ? "var(--echo-primary)" : "var(--echo-bg-elevated)",
                          }}
                        >
                          <Icon size={16} style={{ color: isActive ? "white" : "var(--echo-text-tertiary)" }} />
                        </div>
                        <div className="flex-1">
                          <div style={{ fontSize: 14, fontWeight: isActive ? 600 : 400, color: isActive ? "var(--echo-primary)" : "var(--echo-text-primary)" }}>
                            {t.label}
                          </div>
                          <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                            {t.description}
                          </div>
                        </div>
                        <div
                          className="flex items-center justify-center rounded-full shrink-0"
                          style={{
                            width: 20,
                            height: 20,
                            border: `2px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            backgroundColor: isActive ? "var(--echo-primary)" : "transparent",
                          }}
                        >
                          {isActive && <div className="rounded-full bg-white" style={{ width: 8, height: 8 }} />}
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Time picker for time-based trigger */}
              {selectedTrigger === "time" && (
                <div className="mb-4">
                  <div className="flex items-center justify-between mb-2">
                    <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                      Quick Schedule
                    </span>
                    <button
                      onClick={() => setShowSuggestions(!showSuggestions)}
                      className="flex items-center gap-1 bg-transparent border-none cursor-pointer"
                    >
                      <Sparkles size={13} style={{ color: "var(--echo-accent)" }} />
                      <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-accent)" }}>
                        {showSuggestions ? "Quick times" : "Smart suggest"}
                      </span>
                    </button>
                  </div>

                  {!showSuggestions ? (
                    <div className="flex flex-wrap gap-2">
                      {quickTimes.map((qt) => {
                        const isActive = selectedTime === qt.value;
                        return (
                          <button
                            key={qt.value}
                            onClick={() => setSelectedTime(qt.value)}
                            className="rounded-full px-3.5 py-2 border-none cursor-pointer transition-all"
                            style={{
                              backgroundColor: isActive ? "var(--echo-primary)" : "var(--echo-bg-surface)",
                              border: `1.5px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                              fontSize: 13,
                              fontWeight: isActive ? 600 : 400,
                              color: isActive ? "white" : "var(--echo-text-secondary)",
                            }}
                          >
                            {qt.label}
                          </button>
                        );
                      })}
                    </div>
                  ) : (
                    /* Smart Suggestions */
                    <div className="flex flex-col gap-2">
                      {mockSuggestions.map((s) => {
                        const isActive = selectedTime === s.time;
                        return (
                          <button
                            key={s.time}
                            onClick={() => setSelectedTime(s.time)}
                            className="flex items-center gap-3 w-full rounded-xl px-3 py-2.5 cursor-pointer text-left border-none transition-all"
                            style={{
                              backgroundColor: isActive ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                              border: `1.5px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            }}
                          >
                            <div className="flex-1">
                              <div className="flex items-center gap-2">
                                <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                                  {s.time}
                                </span>
                                <span
                                  className="rounded-full px-1.5 py-0.5"
                                  style={{
                                    fontSize: 10,
                                    fontWeight: 700,
                                    color: s.confidence >= 85 ? "var(--echo-success)" : "var(--echo-warning)",
                                    backgroundColor: s.confidence >= 85 ? "var(--echo-success-light)" : "var(--echo-warning-light)",
                                  }}
                                >
                                  {s.confidence}% match
                                </span>
                              </div>
                              <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                                {s.reason}
                              </div>
                            </div>
                            <Sparkles size={14} style={{ color: "var(--echo-accent)" }} />
                          </button>
                        );
                      })}
                      <div
                        className="rounded-lg p-2 mt-1"
                        style={{ backgroundColor: "var(--echo-info-light)" }}
                      >
                        <span style={{ fontSize: 11, color: "var(--echo-info)", lineHeight: "16px" }}>
                          Suggestions based on {contactName}'s typical activity patterns. Waking hours (7 AM - 11 PM) are enforced.
                        </span>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* No-response timeout config */}
              {selectedTrigger === "no_response" && (
                <div className="mb-4">
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px", marginBottom: 8 }}>
                    Send after no response for
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {["30 min", "1 hour", "3 hours", "6 hours", "24 hours"].map((t) => {
                      const isActive = selectedTime === t;
                      return (
                        <button
                          key={t}
                          onClick={() => setSelectedTime(t)}
                          className="rounded-full px-3.5 py-2 border-none cursor-pointer transition-all"
                          style={{
                            backgroundColor: isActive ? "var(--echo-primary)" : "var(--echo-bg-surface)",
                            border: `1.5px solid ${isActive ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            fontSize: 13,
                            fontWeight: isActive ? 600 : 400,
                            color: isActive ? "white" : "var(--echo-text-secondary)",
                          }}
                        >
                          {t}
                        </button>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Schedule button */}
              <button
                onClick={handleSchedule}
                disabled={!messageText.trim() || (selectedTrigger === "time" && !selectedTime)}
                className="w-full flex items-center justify-center gap-2 border-none cursor-pointer rounded-xl mt-2 active:scale-[0.98] transition-transform"
                style={{
                  height: 52,
                  background: messageText.trim()
                    ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
                    : "var(--echo-bg-surface)",
                  fontSize: 15,
                  fontWeight: 600,
                  color: messageText.trim() ? "white" : "var(--echo-text-tertiary)",
                  boxShadow: messageText.trim() ? "var(--echo-shadow-primary)" : "none",
                }}
              >
                <Send size={16} />
                Schedule Message
              </button>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
