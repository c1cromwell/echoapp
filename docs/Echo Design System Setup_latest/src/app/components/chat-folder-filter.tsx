import React, { useState, useRef } from "react";
import { motion } from "motion/react";
import { ShieldOff, ShieldCheck, UserCheck, Users } from "lucide-react";
import type { VerificationStatus } from "./verification-badge";

export type FolderFilter = "all" | "verified" | "trusted" | "unverified";

interface FilterDef {
  key: FolderFilter;
  label: string;
  dotColor: string | null;
}

const filters: FilterDef[] = [
  { key: "all", label: "All", dotColor: null },
  { key: "verified", label: "Verified", dotColor: "#10B981" },
  { key: "trusted", label: "Trusted", dotColor: "#6366F1" },
  { key: "unverified", label: "Unverified", dotColor: "#9CA3AF" },
];

interface ChatFolderFilterProps {
  active: FolderFilter;
  onChange: (filter: FolderFilter) => void;
  unreadCounts: Record<FolderFilter, number>;
}

export function ChatFolderFilter({ active, onChange, unreadCounts }: ChatFolderFilterProps) {
  return (
    <div
      className="flex gap-2 overflow-x-auto px-4 py-2"
      style={{ scrollbarWidth: "none" }}
    >
      {filters.map((f) => {
        const isActive = active === f.key;
        const unread = unreadCounts[f.key] || 0;

        return (
          <button
            key={f.key}
            onClick={() => onChange(f.key)}
            className="flex items-center gap-1.5 shrink-0 border-none cursor-pointer transition-all"
            style={{
              height: 32,
              borderRadius: 16,
              paddingLeft: 16,
              paddingRight: 16,
              backgroundColor: isActive
                ? "var(--echo-accent)"
                : "var(--echo-bg-secondary)",
              color: isActive
                ? "var(--echo-text-on-accent)"
                : "var(--echo-text-secondary)",
              fontSize: 13,
              fontWeight: isActive ? 600 : 500,
            }}
          >
            {/* Dot indicator */}
            {f.dotColor && (
              <span
                className="shrink-0 rounded-full"
                style={{
                  width: 6,
                  height: 6,
                  backgroundColor: isActive
                    ? "var(--echo-text-on-accent)"
                    : f.dotColor,
                }}
              />
            )}
            <span>{f.label}</span>
            {/* Unread count */}
            {f.key !== "all" && unread > 0 && (
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 700,
                  color: isActive
                    ? "var(--echo-text-on-accent)"
                    : "var(--echo-accent)",
                  marginLeft: 2,
                }}
              >
                {unread}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}

/** Empty state for a filtered view */
const emptyMessages: Record<Exclude<FolderFilter, "all">, { title: string; subtitle: string }> = {
  verified: {
    title: "No verified contacts yet",
    subtitle: "Contacts who verify their identity will appear here.",
  },
  trusted: {
    title: "No trusted contacts yet",
    subtitle: "Contacts who verify their phone will appear here.",
  },
  unverified: {
    title: "No unverified contacts",
    subtitle: "All your contacts have some level of verification!",
  },
};

interface FilterEmptyStateProps {
  filter: Exclude<FolderFilter, "all">;
}

export function FilterEmptyState({ filter }: FilterEmptyStateProps) {
  const msg = emptyMessages[filter];
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6">
      <div
        className="flex items-center justify-center rounded-full mb-4"
        style={{
          width: 56,
          height: 56,
          backgroundColor: "var(--echo-bg-surface)",
        }}
      >
        {filter === "verified" && (
          <ShieldCheck size={28} style={{ color: "var(--echo-text-tertiary)" }} />
        )}
        {filter === "trusted" && (
          <UserCheck size={28} style={{ color: "var(--echo-text-tertiary)" }} />
        )}
        {filter === "unverified" && (
          <ShieldOff size={28} style={{ color: "var(--echo-text-tertiary)" }} />
        )}
      </div>
      <span
        style={{
          fontSize: 16,
          fontWeight: 600,
          color: "var(--echo-text-secondary)",
          marginBottom: 4,
        }}
      >
        {msg.title}
      </span>
      <span
        style={{
          fontSize: 14,
          color: "var(--echo-text-tertiary)",
          textAlign: "center",
          maxWidth: 260,
          lineHeight: "20px",
        }}
      >
        {msg.subtitle}
      </span>
    </div>
  );
}

/** Utility: check if a conversation matches a folder filter based on verification status */
export function matchesFilter(
  filter: FolderFilter,
  verificationStatus?: VerificationStatus
): boolean {
  if (filter === "all") return true;
  const status = verificationStatus || "none";

  switch (filter) {
    case "verified":
      return status === "identity_full" || status === "vip_pro";
    case "trusted":
      return status === "identity_lite" || status === "phone_verified";
    case "unverified":
      return status === "none";
    default:
      return true;
  }
}

/** For group conversations, return the highest-tier filter any member could match */
export function getGroupHighestFilter(
  memberStatuses: (VerificationStatus | undefined)[]
): FolderFilter {
  const hasVerified = memberStatuses.some(
    (s) => s === "identity_full" || s === "vip_pro"
  );
  if (hasVerified) return "verified";

  const hasTrusted = memberStatuses.some(
    (s) => s === "identity_lite" || s === "phone_verified"
  );
  if (hasTrusted) return "trusted";

  return "unverified";
}
