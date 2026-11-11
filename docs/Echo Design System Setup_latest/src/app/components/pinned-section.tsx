import React from "react";
import { EchoAvatar, GroupAvatar } from "./echo-avatar";

interface PinnedContact {
  id: string;
  name: string;
  type: "contact" | "group";
  online?: boolean;
  unread?: number;
  gradient?: "primary" | "secondary" | "warm" | "cool";
  members?: { name: string; color?: string }[];
}

const pinnedData: PinnedContact[] = [
  { id: "1", name: "Sarah", type: "contact", online: true, unread: 3, gradient: "primary" },
  { id: "2", name: "Marcus", type: "contact", online: true, gradient: "secondary" },
  { id: "3", name: "Product Team", type: "group", unread: 12, members: [
    { name: "Alex" }, { name: "Beth" }, { name: "Chris" }, { name: "Dana" }, { name: "Eve" }, { name: "Frank" },
  ]},
  { id: "4", name: "Elena", type: "contact", online: false, gradient: "warm" },
  { id: "5", name: "Chase", type: "contact", online: true, unread: 1, gradient: "cool" },
];

export function PinnedSection() {
  return (
    <div className="py-3">
      {/* Header */}
      <div className="flex items-center justify-between px-4 mb-3">
        <div className="flex items-center gap-1.5">
          <span style={{ fontSize: 12 }}>📌</span>
          <span
            className="tracking-[1px]"
            style={{
              fontSize: 13,
              fontWeight: 700,
              color: "var(--echo-text-tertiary)",
              letterSpacing: "1px",
            }}
          >
            PINNED
          </span>
        </div>
        <button
          style={{
            fontSize: 13,
            fontWeight: 600,
            color: "var(--echo-primary)",
            background: "none",
            border: "none",
            cursor: "pointer",
          }}
        >
          Edit
        </button>
      </div>

      {/* Scroll area */}
      <div className="flex gap-3 px-4 overflow-x-auto pt-1 pb-1" style={{ scrollbarWidth: "none", overflowY: "visible" }}>
        {pinnedData.map((item) => (
          <PinnedItem key={item.id} item={item} />
        ))}
      </div>
    </div>
  );
}

function PinnedItem({ item }: { item: PinnedContact }) {
  return (
    <button
      className="flex flex-col items-center gap-1.5 shrink-0 bg-transparent border-none cursor-pointer active:scale-95 transition-transform"
      style={{ width: 64 }}
    >
      <div className="relative">
        {item.type === "group" && item.members ? (
          <GroupAvatar members={item.members} size={46} />
        ) : (
          <EchoAvatar
            name={item.name}
            size="md"
            online={item.online}
            gradient={item.gradient}
          />
        )}
        {item.unread && item.unread > 0 && (
          <span
            className="absolute flex items-center justify-center rounded-full"
            style={{
              top: -3,
              right: -3,
              minWidth: 18,
              height: 18,
              padding: "0 5px",
              backgroundColor: "#F43F5E",
              border: "2px solid white",
              fontSize: 9,
              fontWeight: 700,
              color: "white",
              lineHeight: 1,
            }}
          >
            {item.unread}
          </span>
        )}
      </div>
      <span
        className="w-full text-center truncate"
        style={{
          fontSize: 11,
          fontWeight: 500,
          color: "var(--echo-text-secondary)",
          lineHeight: "16px",
        }}
      >
        {item.name}
      </span>
    </button>
  );
}