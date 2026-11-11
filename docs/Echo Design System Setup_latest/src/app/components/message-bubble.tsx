import React, { useState } from "react";
import {
  Check, CheckCheck, Clock, Timer, Pencil, Trash2, AlertCircle,
  VolumeX, RotateCcw,
} from "lucide-react";

export interface Message {
  id: string;
  text: string;
  time: string;
  sent: boolean;
  status?: "sending" | "sent" | "delivered" | "read" | "failed";
  // Enhanced fields
  isScheduled?: boolean;
  scheduledTime?: string;
  conditionalTrigger?: "time" | "when_online" | "after_read" | "no_response" | "conversation_ended";
  isSilent?: boolean;
  isDisappearing?: boolean;
  disappearTimer?: number; // seconds remaining
  isEdited?: boolean;
  editCount?: number;
  isRecalled?: boolean;
  recalledBy?: string;
  deliveryFailure?: {
    type: "offline" | "blocked" | "network" | "key_release_failed" | "account_deleted";
    retryCount: number;
    canRetry: boolean;
  };
}

interface MessageBubbleProps {
  message: Message;
  onLongPress?: (msg: Message) => void;
  onRetry?: (msg: Message) => void;
}

const triggerLabels: Record<string, string> = {
  time: "Scheduled",
  when_online: "When online",
  after_read: "After read",
  no_response: "If no response",
  conversation_ended: "On end",
};

const failureLabels: Record<string, string> = {
  offline: "Recipient offline",
  blocked: "Blocked",
  network: "Network error",
  key_release_failed: "Encryption error",
  account_deleted: "Account deleted",
};

export function MessageBubble({ message, onLongPress, onRetry }: MessageBubbleProps) {
  const { text, time, sent, status, isScheduled, scheduledTime, conditionalTrigger, isSilent, isDisappearing, disappearTimer, isEdited, isRecalled, deliveryFailure } = message;
  const [pressTimer, setPressTimer] = useState<ReturnType<typeof setTimeout> | null>(null);

  const handleTouchStart = () => {
    if (onLongPress) {
      const timer = setTimeout(() => onLongPress(message), 500);
      setPressTimer(timer);
    }
  };
  const handleTouchEnd = () => {
    if (pressTimer) clearTimeout(pressTimer);
    setPressTimer(null);
  };

  // Recalled message
  if (isRecalled) {
    return (
      <div className={`flex ${sent ? "justify-end" : "justify-start"} px-4 mb-2`}>
        <div
          className="flex items-center gap-2 rounded-2xl px-4 py-2.5"
          style={{
            backgroundColor: "var(--echo-bg-surface)",
            border: "1px dashed var(--echo-border)",
          }}
        >
          <Trash2 size={14} style={{ color: "var(--echo-text-tertiary)" }} />
          <span style={{ fontSize: 14, fontStyle: "italic", color: "var(--echo-text-tertiary)" }}>
            {sent ? "You deleted this message" : "This message was deleted"}
          </span>
        </div>
      </div>
    );
  }

  return (
    <div
      className={`flex ${sent ? "justify-end" : "justify-start"} px-4 mb-2`}
      onMouseDown={handleTouchStart}
      onMouseUp={handleTouchEnd}
      onMouseLeave={handleTouchEnd}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      <div className="max-w-[280px] flex flex-col">
        {/* Scheduled indicator above bubble */}
        {isScheduled && sent && (
          <div className={`flex items-center gap-1 mb-1 ${sent ? "self-end" : "self-start"}`}>
            <Clock size={11} style={{ color: "var(--echo-accent)" }} />
            <span style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-accent)" }}>
              {conditionalTrigger ? triggerLabels[conditionalTrigger] : "Scheduled"}{scheduledTime ? ` · ${scheduledTime}` : ""}
            </span>
          </div>
        )}

        <div
          style={{
            padding: "10px 14px",
            borderRadius: sent ? "20px 20px 4px 20px" : "20px 20px 20px 4px",
            backgroundColor: sent
              ? (isSilent ? "var(--echo-primary-darker)" : "var(--echo-primary)")
              : "var(--echo-bg-surface)",
            boxShadow: "var(--echo-shadow-sm)",
            opacity: isScheduled ? 0.85 : 1,
            border: isDisappearing
              ? `1px solid ${sent ? "rgba(255,255,255,0.3)" : "var(--echo-warning)"}`
              : deliveryFailure
              ? `1px solid var(--echo-error)`
              : "none",
          }}
        >
          {/* Message text */}
          <p
            className="m-0 mb-1"
            style={{
              fontSize: 16,
              fontWeight: 400,
              lineHeight: "24px",
              color: sent ? "white" : "var(--echo-text-primary)",
              wordBreak: "break-word",
            }}
          >
            {text}
          </p>

          {/* Footer row */}
          <div className={`flex items-center gap-1.5 flex-wrap ${sent ? "justify-end" : "justify-start"}`}>
            {/* Silent badge */}
            {isSilent && (
              <span className="flex items-center gap-0.5">
                <VolumeX size={10} style={{ color: sent ? "rgba(255,255,255,0.6)" : "var(--echo-text-tertiary)" }} />
                <span style={{ fontSize: 10, color: sent ? "rgba(255,255,255,0.5)" : "var(--echo-text-tertiary)" }}>Silent</span>
              </span>
            )}

            {/* Disappearing badge */}
            {isDisappearing && (
              <span className="flex items-center gap-0.5">
                <Timer size={10} style={{ color: sent ? "rgba(255,255,255,0.6)" : "var(--echo-warning)" }} />
                {disappearTimer !== undefined && (
                  <span style={{ fontSize: 10, color: sent ? "rgba(255,255,255,0.5)" : "var(--echo-warning)" }}>
                    {disappearTimer > 3600 ? `${Math.floor(disappearTimer / 3600)}h` : disappearTimer > 60 ? `${Math.floor(disappearTimer / 60)}m` : `${disappearTimer}s`}
                  </span>
                )}
              </span>
            )}

            {/* Edited badge */}
            {isEdited && (
              <span className="flex items-center gap-0.5">
                <Pencil size={9} style={{ color: sent ? "rgba(255,255,255,0.5)" : "var(--echo-text-tertiary)" }} />
                <span style={{ fontSize: 10, color: sent ? "rgba(255,255,255,0.5)" : "var(--echo-text-tertiary)" }}>edited</span>
              </span>
            )}

            {/* Time */}
            <span
              style={{
                fontSize: 12,
                fontWeight: 500,
                color: sent ? "rgba(255,255,255,0.7)" : "var(--echo-text-tertiary)",
                lineHeight: "16px",
              }}
            >
              {time}
            </span>

            {/* Status */}
            {sent && status && (
              <span className="flex items-center">
                {status === "sending" && (
                  <Clock size={12} style={{ color: "rgba(255,255,255,0.5)" }} />
                )}
                {status === "sent" && (
                  <Check size={13} style={{ color: "rgba(255,255,255,0.7)" }} />
                )}
                {status === "delivered" && (
                  <CheckCheck size={13} style={{ color: "rgba(255,255,255,0.5)" }} />
                )}
                {status === "read" && (
                  <CheckCheck size={13} style={{ color: "#60E0FF" }} />
                )}
                {status === "failed" && (
                  <AlertCircle size={13} style={{ color: "var(--echo-error)" }} />
                )}
              </span>
            )}
          </div>
        </div>

        {/* Delivery failure info */}
        {deliveryFailure && (
          <div className={`flex items-center gap-1.5 mt-1 ${sent ? "self-end" : "self-start"}`}>
            <AlertCircle size={11} style={{ color: "var(--echo-error)" }} />
            <span style={{ fontSize: 10, color: "var(--echo-error)" }}>
              {failureLabels[deliveryFailure.type]}
              {deliveryFailure.retryCount > 0 && ` (retry ${deliveryFailure.retryCount})`}
            </span>
            {deliveryFailure.canRetry && onRetry && (
              <button
                onClick={() => onRetry(message)}
                className="flex items-center gap-0.5 bg-transparent border-none cursor-pointer p-0"
              >
                <RotateCcw size={10} style={{ color: "var(--echo-primary)" }} />
                <span style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-primary)" }}>Retry</span>
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}