import React from "react";
import { BadgeCheck, Crown } from "lucide-react";

export type VerificationStatus =
  | "none"
  | "phone_verified"
  | "identity_lite"
  | "identity_full"
  | "vip_pro";

const badgeColors: Record<VerificationStatus, string> = {
  none: "transparent",
  phone_verified: "#9CA3AF",
  identity_lite: "#6366F1",
  identity_full: "#10B981",
  vip_pro: "#F59E0B",
};

interface VerificationBadgeProps {
  status: VerificationStatus;
  size?: "sm" | "md" | "lg";
}

const sizeMap = { sm: 14, md: 16, lg: 20 };

export function VerificationBadge({ status, size = "sm" }: VerificationBadgeProps) {
  if (status === "none") return null;

  const px = sizeMap[size];
  const color = badgeColors[status];

  if (status === "vip_pro") {
    return <Crown size={px} style={{ color, flexShrink: 0 }} fill={color} />;
  }

  return <BadgeCheck size={px} style={{ color, flexShrink: 0 }} />;
}

interface VerificationPillProps {
  status: VerificationStatus;
}

const pillLabels: Record<VerificationStatus, string> = {
  none: "",
  phone_verified: "Phone Verified",
  identity_lite: "Verified",
  identity_full: "Fully Verified",
  vip_pro: "VIP Pro",
};

const pillBgColors: Record<VerificationStatus, string> = {
  none: "transparent",
  phone_verified: "#F3F4F6",
  identity_lite: "#EEF2FF",
  identity_full: "#D1FAE5",
  vip_pro: "#FEF3C7",
};

export function VerificationPill({ status }: VerificationPillProps) {
  if (status === "none") return null;

  return (
    <div
      className="flex items-center gap-1.5 rounded-full px-2.5"
      style={{
        height: 24,
        backgroundColor: pillBgColors[status],
      }}
    >
      <VerificationBadge status={status} size="sm" />
      <span
        style={{
          fontSize: 12,
          fontWeight: 600,
          color: badgeColors[status],
          lineHeight: "16px",
        }}
      >
        {pillLabels[status]}
      </span>
    </div>
  );
}
