import React from "react";
import { useNavigate } from "react-router";
import { GroupAvatar } from "./echo-avatar";
import { Lock, Eye, EyeOff, Users, ShieldCheck } from "lucide-react";
import type { Group } from "./group-data";

const typeIcons: Record<string, React.ReactNode> = {
  public: <Eye size={12} style={{ color: "var(--echo-text-tertiary)" }} />,
  private: <Lock size={12} style={{ color: "var(--echo-warning)" }} />,
  secret: <EyeOff size={12} style={{ color: "var(--echo-error)" }} />,
};

const typeLabels: Record<string, string> = {
  public: "Public",
  private: "Private",
  secret: "Secret",
};

interface GroupsListProps {
  groups: Group[];
}

export function GroupsList({ groups }: GroupsListProps) {
  const navigate = useNavigate();

  if (groups.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 px-6">
        <div
          className="flex items-center justify-center rounded-2xl mb-4"
          style={{
            width: 64,
            height: 64,
            backgroundColor: "var(--echo-primary-subtle)",
          }}
        >
          <Users size={28} style={{ color: "var(--echo-primary)" }} />
        </div>
        <p style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px" }}>
          No groups yet
        </p>
        <p style={{ fontSize: 14, color: "var(--echo-text-tertiary)", textAlign: "center" }}>
          Create or join a group to get started
        </p>
      </div>
    );
  }

  return (
    <div>
      {groups.map((group) => (
        <button
          key={group.id}
          onClick={() => navigate(`/groups/${group.id}`)}
          className="flex items-center gap-3 w-full px-4 py-3 cursor-pointer bg-transparent border-none text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
        >
          <GroupAvatar
            members={group.members.length > 0
              ? group.members.slice(0, 4).map((m) => ({ name: m.name }))
              : [{ name: group.name[0] }, { name: group.name.split(" ")[1]?.[0] || "G" }]
            }
            size={56}
          />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-1.5 mb-0.5">
              <span
                className="truncate"
                style={{
                  fontSize: 16,
                  fontWeight: 500,
                  color: "var(--echo-text-primary)",
                  lineHeight: "22px",
                }}
              >
                {group.name}
              </span>
              {typeIcons[group.type]}
              {group.stats.securityRating === "A" && (
                <ShieldCheck size={14} style={{ color: "var(--echo-success)" }} />
              )}
            </div>
            <div className="flex items-center gap-1">
              <span
                className="truncate"
                style={{
                  fontSize: 14,
                  color: "var(--echo-text-tertiary)",
                  lineHeight: "20px",
                }}
              >
                {group.lastMessage || `${group.memberCount} members`}
              </span>
            </div>
          </div>
          <div className="flex flex-col items-end gap-1 shrink-0">
            {group.lastMessageTime && (
              <span
                style={{
                  fontSize: 12,
                  color: group.unread ? "var(--echo-primary)" : "var(--echo-text-tertiary)",
                  fontWeight: group.unread ? 600 : 400,
                }}
              >
                {group.lastMessageTime}
              </span>
            )}
            {group.unread > 0 && (
              <span
                className="flex items-center justify-center rounded-full"
                style={{
                  minWidth: 20,
                  height: 20,
                  padding: "0 6px",
                  background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                  fontSize: 11,
                  fontWeight: 700,
                  color: "white",
                }}
              >
                {group.unread}
              </span>
            )}
          </div>
        </button>
      ))}
    </div>
  );
}
