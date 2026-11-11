import React from "react";

interface AvatarProps {
  name: string;
  size?: "xs" | "sm" | "md" | "lg" | "xl";
  online?: boolean;
  gradient?: "primary" | "secondary" | "warm" | "cool";
  image?: string;
}

const sizeMap = {
  xs: 24,
  sm: 32,
  md: 48,
  lg: 64,
  xl: 80,
};

const fontSizeMap = {
  xs: "9px",
  sm: "11px",
  md: "18px",
  lg: "24px",
  xl: "30px",
};

const gradients = {
  primary: "from-[var(--echo-gradient-from)] to-[var(--echo-gradient-to)]",
  secondary: "from-[#06B6D4] to-[#10B981]",
  warm: "from-[#F43F5E] to-[#F59E0B]",
  cool: "from-[#3B82F6] to-[#8B5CF6]",
};

const onlineDotSize = {
  xs: 8,
  sm: 10,
  md: 14,
  lg: 16,
  xl: 20,
};

export function EchoAvatar({ name, size = "md", online, gradient = "primary", image }: AvatarProps) {
  const px = sizeMap[size];
  const initials = name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
  const dotPx = onlineDotSize[size];

  return (
    <div className="relative inline-flex shrink-0" style={{ width: px, height: px }}>
      {image ? (
        <img
          src={image}
          alt={name}
          className="rounded-full object-cover"
          style={{ width: px, height: px }}
        />
      ) : (
        <div
          className={`rounded-full bg-gradient-to-br ${gradients[gradient]} flex items-center justify-center`}
          style={{
            width: px,
            height: px,
            boxShadow: "0 4px 8px rgba(0,0,0,0.15)",
          }}
        >
          <span
            className="text-white"
            style={{ fontSize: fontSizeMap[size], fontWeight: 600, lineHeight: 1 }}
          >
            {initials}
          </span>
        </div>
      )}
      {online !== undefined && (
        <span
          className="absolute rounded-full"
          style={{
            width: dotPx,
            height: dotPx,
            bottom: size === "xs" ? -1 : 2,
            right: size === "xs" ? -1 : 2,
            backgroundColor: online ? "#10B981" : "#A1A1AA",
            border: `${size === "xs" ? 1.5 : 2.5}px solid white`,
          }}
        />
      )}
    </div>
  );
}

interface GroupAvatarProps {
  members: { name: string; color?: string }[];
  size?: number;
}

const memberColors = ["#7C3AED", "#06B6D4", "#F43F5E", "#10B981"];

export function GroupAvatar({ members, size = 56 }: GroupAvatarProps) {
  const displayed = members.slice(0, 4);
  const overflow = members.length - 3;
  const faceSize = (size - 12) / 2;

  return (
    <div
      className="rounded-full flex flex-wrap items-center justify-center gap-[2px] shrink-0"
      style={{
        width: size,
        height: size,
        backgroundColor: "var(--echo-bg-surface)",
        border: "2px solid var(--echo-border)",
        padding: 4,
        overflow: "hidden",
      }}
    >
      {displayed.map((m, i) => {
        const isOverflow = i === 3 && overflow > 1;
        return (
          <div
            key={i}
            className="rounded-[10px] flex items-center justify-center"
            style={{
              width: faceSize,
              height: faceSize,
              backgroundColor: m.color || memberColors[i % memberColors.length],
            }}
          >
            <span style={{ fontSize: 10, fontWeight: 600, color: "white", lineHeight: 1 }}>
              {isOverflow ? `+${overflow}` : m.name[0].toUpperCase()}
            </span>
          </div>
        );
      })}
    </div>
  );
}
