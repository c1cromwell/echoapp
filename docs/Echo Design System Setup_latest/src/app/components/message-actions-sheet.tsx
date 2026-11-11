import React from "react";
import {
  Reply, Pencil, Trash2, Clock, Copy, Forward, Pin,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import type { Message } from "./message-bubble";

interface MessageActionsSheetProps {
  open: boolean;
  onClose: () => void;
  message: Message | null;
  onAction: (action: "reply" | "edit" | "recall" | "schedule" | "copy" | "forward" | "pin") => void;
}

export function MessageActionsSheet({ open, onClose, message, onAction }: MessageActionsSheetProps) {
  if (!message) return null;

  const actions = [
    { id: "reply" as const, icon: Reply, label: "Reply", color: "var(--echo-text-primary)", always: true },
    { id: "copy" as const, icon: Copy, label: "Copy", color: "var(--echo-text-primary)", always: true },
    { id: "forward" as const, icon: Forward, label: "Forward", color: "var(--echo-text-primary)", always: true },
    { id: "pin" as const, icon: Pin, label: "Pin Message", color: "var(--echo-text-primary)", always: true },
    { id: "edit" as const, icon: Pencil, label: "Edit", color: "var(--echo-primary)", always: false, sentOnly: true },
    { id: "schedule" as const, icon: Clock, label: "Reschedule", color: "var(--echo-accent)", always: false, scheduledOnly: true },
    { id: "recall" as const, icon: Trash2, label: "Delete for Everyone", color: "var(--echo-error)", always: false, sentOnly: true },
  ];

  const filtered = actions.filter((a) => {
    if (a.always) return true;
    if ((a as any).sentOnly && !message.sent) return false;
    if ((a as any).scheduledOnly && !message.isScheduled) return false;
    return true;
  });

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
            style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
          />
          <motion.div
            initial={{ y: "100%", opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: "100%", opacity: 0 }}
            transition={{ type: "spring", damping: 28, stiffness: 350 }}
            className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
            style={{ backgroundColor: "var(--echo-bg-elevated)" }}
          >
            <div className="flex justify-center pt-3 pb-1">
              <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
            </div>

            {/* Preview */}
            <div className="px-5 pb-3">
              <div
                className="rounded-xl px-3 py-2"
                style={{ backgroundColor: "var(--echo-bg-surface)", maxHeight: 60, overflow: "hidden" }}
              >
                <span
                  className="line-clamp-2"
                  style={{ fontSize: 13, color: "var(--echo-text-secondary)", lineHeight: "18px" }}
                >
                  {message.text}
                </span>
              </div>
            </div>

            {/* Actions grid */}
            <div className="px-5 pb-8">
              <div className="grid grid-cols-4 gap-2">
                {filtered.map((a) => {
                  const Icon = a.icon;
                  return (
                    <button
                      key={a.id}
                      onClick={() => {
                        onAction(a.id);
                        onClose();
                      }}
                      className="flex flex-col items-center gap-1.5 rounded-xl py-3 bg-transparent border-none cursor-pointer hover:bg-[var(--echo-bg-surface)] transition-colors"
                    >
                      <div
                        className="flex items-center justify-center rounded-xl"
                        style={{
                          width: 44,
                          height: 44,
                          backgroundColor: a.id === "recall" ? "var(--echo-error-light)" : "var(--echo-bg-surface)",
                        }}
                      >
                        <Icon size={20} style={{ color: a.color }} />
                      </div>
                      <span style={{ fontSize: 11, fontWeight: 500, color: a.color }}>
                        {a.label}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
