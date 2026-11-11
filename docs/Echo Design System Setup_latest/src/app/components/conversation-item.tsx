import React from "react";
import { EchoAvatar, GroupAvatar } from "./echo-avatar";
import { VolumeX, Timer, Clock } from "lucide-react";
import { VerificationBadge, type VerificationStatus } from "./verification-badge";

export interface Conversation {
  id: string;
  name: string;
  lastMessage: string;
  time: string;
  unread?: number;
  online?: boolean;
  muted?: boolean;
  type?: "individual" | "group";
  gradient?: "primary" | "secondary" | "warm" | "cool";
  members?: { name: string; color?: string }[];
  image?: string;
  // Enhanced fields
  isSilentConversation?: boolean;
  isDisappearing?: boolean;
  hasScheduledMessages?: boolean;
  hasDeliveryFailure?: boolean;
  // Verification badge (Section 13)
  verificationStatus?: VerificationStatus;
}

interface ConversationItemProps {
  conversation: Conversation;
  onClick?: () => void;
}

export function ConversationItem({ conversation, onClick }: ConversationItemProps) {
  const {
    name, lastMessage, time, unread, online, muted, type, gradient, members, image,
    isSilentConversation, isDisappearing, hasScheduledMessages, hasDeliveryFailure,
    verificationStatus,
  } = conversation;
  const hasUnread = unread && unread > 0;

  return (
    <button
      onClick={onClick}
      className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer transition-colors hover:bg-[var(--echo-bg-surface)] active:bg-[var(--echo-bg-surface)] text-left"
    >
      {/* Avatar */}
      <div className="relative">
        {type === "group" && members ? (
          <GroupAvatar members={members} size={48} />
        ) : (
          <EchoAvatar name={name} size="md" online={online} gradient={gradient} image={image} />
        )}
        {/* Silent badge on avatar */}
        {isSilentConversation && (
          <div
            className="absolute -bottom-0.5 -right-0.5 flex items-center justify-center rounded-full"
            style={{
              width: 18,
              height: 18,
              backgroundColor: "var(--echo-bg-elevated)",
              border: "2px solid var(--echo-bg-primary)",
            }}
          >
            <VolumeX size={10} style={{ color: "var(--echo-primary)" }} />
          </div>
        )}
      </div>

      <div className="flex-1 min-w-0">
        {/* Top row: name + badge + other badges + time */}
        <div className="flex items-center justify-between mb-0.5">
          <div className="flex items-center gap-1 min-w-0">
            <span
              className="truncate"
              style={{
                fontSize: 16,
                fontWeight: hasUnread ? 600 : 400,
                color: "var(--echo-text-primary)",
                lineHeight: "22px",
              }}
            >
              {name}
            </span>
            {/* Verification badge — only for individual conversations */}
            {type !== "group" && verificationStatus && verificationStatus !== "none" && (
              <VerificationBadge status={verificationStatus} size="sm" />
            )}
            {/* Inline feature badges */}
            {isDisappearing && (
              <Timer size={12} style={{ color: "var(--echo-warning)", flexShrink: 0 }} />
            )}
            {hasScheduledMessages && (
              <Clock size={12} style={{ color: "var(--echo-accent)", flexShrink: 0 }} />
            )}
          </div>
          <span
            className="shrink-0 ml-2"
            style={{
              fontSize: 13,
              fontWeight: 400,
              color: hasUnread ? "var(--echo-accent)" : hasDeliveryFailure ? "var(--echo-error)" : "var(--echo-text-tertiary)",
              lineHeight: "18px",
            }}
          >
            {time}
          </span>
        </div>

        {/* Bottom row: last message + unread/muted */}
        <div className="flex items-center justify-between">
          <span
            className="truncate"
            style={{
              fontSize: 14,
              fontWeight: hasUnread ? 500 : 400,
              color: hasDeliveryFailure
                ? "var(--echo-error)"
                : hasUnread
                ? "var(--echo-text-secondary)"
                : "var(--echo-text-tertiary)",
              lineHeight: "20px",
              maxWidth: hasUnread ? "calc(100% - 30px)" : "100%",
            }}
          >
            {hasDeliveryFailure ? "Failed to deliver" : lastMessage}
          </span>
          <div className="flex items-center gap-1.5 shrink-0 ml-1">
            {hasUnread && !muted && (
              <span
                className="flex items-center justify-center rounded-full shrink-0"
                style={{
                  minWidth: 20,
                  height: 20,
                  padding: "0 6px",
                  backgroundColor: "var(--echo-accent)",
                  fontSize: 11,
                  fontWeight: 700,
                  color: "white",
                  lineHeight: 1,
                }}
              >
                {unread}
              </span>
            )}
            {muted && (
              <span style={{ fontSize: 14, color: "var(--echo-text-tertiary)" }}>🔇</span>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}