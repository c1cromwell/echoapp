import React, { useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ArrowLeft, MoreHorizontal, Users, ShieldCheck, Lock, Eye, EyeOff,
  Crown, Shield, Star, UserCheck, User, Ban, TriangleAlert,
  ChevronRight, Settings, Bell, LogOut, Activity, TrendingUp,
} from "lucide-react";
import { EchoAvatar, GroupAvatar } from "../components/echo-avatar";
import { TrustRing } from "../components/trust-ring";
import {
  mockGroups,
  roleLabels,
  roleColors,
  categoryIcons,
  type Group,
  type GroupMember,
  type GroupRole,
} from "../components/group-data";
import { motion } from "motion/react";

const roleIcons: Record<GroupRole, React.ElementType> = {
  owner: Crown,
  admin: Shield,
  moderator: Star,
  verified_member: UserCheck,
  member: User,
  restricted: Ban,
  pending: TriangleAlert,
};

const tabs = ["Info", "Members", "Stats"];

function SecurityBadge({ rating }: { rating: string }) {
  const colors: Record<string, string> = {
    A: "var(--echo-success)",
    B: "#3B82F6",
    C: "var(--echo-warning)",
    D: "var(--echo-error)",
    F: "var(--echo-error)",
  };
  return (
    <span
      className="flex items-center justify-center rounded-lg"
      style={{
        width: 28,
        height: 28,
        backgroundColor: `${colors[rating]}20`,
        fontSize: 14,
        fontWeight: 700,
        color: colors[rating],
      }}
    >
      {rating}
    </span>
  );
}

export function GroupDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [activeTab, setActiveTab] = useState(0);
  const [showModMenu, setShowModMenu] = useState<string | null>(null);

  const group = mockGroups.find((g) => g.id === id);

  if (!group) {
    return (
      <div className="flex flex-col items-center justify-center h-full px-6">
        <p style={{ fontSize: 16, color: "var(--echo-text-tertiary)" }}>Group not found</p>
        <button
          onClick={() => navigate("/")}
          className="mt-4 bg-transparent border-none cursor-pointer"
          style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}
        >
          Go back
        </button>
      </div>
    );
  }

  const typeConfig = {
    public: { icon: Eye, label: "Public", color: "var(--echo-success)" },
    private: { icon: Lock, label: "Private", color: "var(--echo-warning)" },
    secret: { icon: EyeOff, label: "Secret", color: "var(--echo-error)" },
  };

  const TypeIcon = typeConfig[group.type].icon;

  // Sort members by role priority
  const rolePriority: GroupRole[] = ["owner", "admin", "moderator", "verified_member", "member", "restricted", "pending"];
  const sortedMembers = [...group.members].sort(
    (a, b) => rolePriority.indexOf(a.role) - rolePriority.indexOf(b.role)
  );

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="shrink-0"
        style={{
          background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
          paddingTop: 44,
        }}
      >
        <div className="flex items-center justify-between px-4 py-2">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center justify-center bg-transparent border-none cursor-pointer rounded-full"
            style={{ width: 44, height: 44 }}
          >
            <ArrowLeft size={24} color="white" />
          </button>
          <div className="flex items-center gap-2">
            <button
              onClick={() => navigate(`/group-chat/${group.id}`)}
              className="rounded-full border-none cursor-pointer px-4 py-1.5"
              style={{
                backgroundColor: "rgba(255,255,255,0.2)",
                fontSize: 13,
                fontWeight: 600,
                color: "white",
              }}
            >
              Open Chat
            </button>
            <button
              className="flex items-center justify-center bg-transparent border-none cursor-pointer rounded-full"
              style={{ width: 44, height: 44 }}
            >
              <MoreHorizontal size={24} color="white" />
            </button>
          </div>
        </div>

        {/* Group info */}
        <div className="flex flex-col items-center pb-6 px-6">
          <div className="mb-3">
            <GroupAvatar
              members={group.members.slice(0, 4).map((m) => ({ name: m.name }))}
              size={72}
            />
          </div>
          <h1
            style={{
              fontSize: 22,
              fontWeight: 700,
              color: "white",
              margin: "0 0 4px",
              textAlign: "center",
            }}
          >
            {group.name}
          </h1>
          <div className="flex items-center gap-2 mb-2">
            <TypeIcon size={14} color="rgba(255,255,255,0.8)" />
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.8)" }}>
              {typeConfig[group.type].label}
            </span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.5)" }}>•</span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.8)" }}>
              {categoryIcons[group.category]} {group.category}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <Users size={14} color="rgba(255,255,255,0.7)" />
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.7)" }}>
              {group.memberCount}/{group.maxMembers} members
            </span>
          </div>
        </div>
      </div>

      {/* Tab bar */}
      <div
        className="flex shrink-0"
        style={{
          borderBottom: "1px solid var(--echo-border-light)",
          backgroundColor: "var(--echo-bg-elevated)",
        }}
      >
        {tabs.map((tab, i) => (
          <button
            key={tab}
            onClick={() => setActiveTab(i)}
            className="flex-1 flex items-center justify-center bg-transparent border-none cursor-pointer py-3 transition-all"
            style={{
              fontSize: 14,
              fontWeight: activeTab === i ? 600 : 400,
              color: activeTab === i ? "var(--echo-primary)" : "var(--echo-text-tertiary)",
              borderBottom: activeTab === i ? "2px solid var(--echo-primary)" : "2px solid transparent",
            }}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-y-auto">
        {/* INFO TAB */}
        {activeTab === 0 && (
          <div className="px-6 py-4">
            {/* Description */}
            <div className="mb-6">
              <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", letterSpacing: "0.5px", textTransform: "uppercase", margin: "0 0 8px" }}>
                About
              </h3>
              <p style={{ fontSize: 15, color: "var(--echo-text-secondary)", lineHeight: "22px", margin: 0 }}>
                {group.description}
              </p>
            </div>

            {/* Tags */}
            {group.tags.length > 0 && (
              <div className="mb-6">
                <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", letterSpacing: "0.5px", textTransform: "uppercase", margin: "0 0 8px" }}>
                  Tags
                </h3>
                <div className="flex flex-wrap gap-2">
                  {group.tags.map((tag) => (
                    <span
                      key={tag}
                      className="rounded-full px-3 py-1"
                      style={{
                        fontSize: 13,
                        fontWeight: 500,
                        color: "var(--echo-primary)",
                        backgroundColor: "var(--echo-primary-subtle)",
                      }}
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Requirements */}
            <div className="mb-6">
              <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", letterSpacing: "0.5px", textTransform: "uppercase", margin: "0 0 8px" }}>
                Entry Requirements
              </h3>
              <div
                className="rounded-2xl p-4 flex flex-col gap-3"
                style={{
                  backgroundColor: "var(--echo-bg-surface)",
                }}
              >
                <div className="flex items-center justify-between">
                  <span style={{ fontSize: 14, color: "var(--echo-text-secondary)" }}>Min Trust Score</span>
                  <span
                    className="flex items-center gap-1.5 rounded-full px-2.5 py-0.5"
                    style={{
                      backgroundColor: "var(--echo-primary-subtle)",
                      fontSize: 13,
                      fontWeight: 600,
                      color: "var(--echo-primary)",
                    }}
                  >
                    <ShieldCheck size={13} />
                    {group.minTrustScore}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span style={{ fontSize: 14, color: "var(--echo-text-secondary)" }}>Approval</span>
                  <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                    {group.approvalMode === "auto" ? "Automatic" : group.approvalMode === "manual" ? "Manual" : "Vote"}
                  </span>
                </div>
                {group.requiredBadges.length > 0 && (
                  <div className="flex items-center justify-between">
                    <span style={{ fontSize: 14, color: "var(--echo-text-secondary)" }}>Required Badges</span>
                    <div className="flex gap-1">
                      {group.requiredBadges.map((b) => (
                        <span
                          key={b}
                          className="rounded-full px-2 py-0.5"
                          style={{ fontSize: 11, fontWeight: 600, color: "white", backgroundColor: "var(--echo-accent)" }}
                        >
                          {b}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Rules */}
            {group.rules.length > 0 && (
              <div className="mb-6">
                <h3 style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", letterSpacing: "0.5px", textTransform: "uppercase", margin: "0 0 8px" }}>
                  Rules
                </h3>
                <div className="flex flex-col gap-2">
                  {group.rules.map((rule, i) => (
                    <div key={i} className="flex items-start gap-3">
                      <span
                        className="flex items-center justify-center rounded-full shrink-0 mt-0.5"
                        style={{
                          width: 22,
                          height: 22,
                          backgroundColor: "var(--echo-primary-subtle)",
                          fontSize: 11,
                          fontWeight: 700,
                          color: "var(--echo-primary)",
                        }}
                      >
                        {i + 1}
                      </span>
                      <span style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "20px" }}>
                        {rule}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Quick actions */}
            <div className="flex flex-col gap-1 mb-6">
              {[
                { icon: Settings, label: "Group Settings", color: "var(--echo-text-primary)" },
                { icon: Bell, label: "Notifications", color: "var(--echo-text-primary)" },
                { icon: LogOut, label: "Leave Group", color: "var(--echo-error)" },
              ].map((action) => (
                <button
                  key={action.label}
                  className="flex items-center gap-3 w-full bg-transparent border-none cursor-pointer rounded-xl px-3 py-3 hover:bg-[var(--echo-bg-surface)] transition-colors text-left"
                >
                  <action.icon size={20} style={{ color: action.color }} />
                  <span style={{ fontSize: 15, fontWeight: 500, color: action.color, flex: 1 }}>
                    {action.label}
                  </span>
                  <ChevronRight size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                </button>
              ))}
            </div>
          </div>
        )}

        {/* MEMBERS TAB */}
        {activeTab === 1 && (
          <div className="py-2">
            {/* Role groups */}
            {(["owner", "admin", "moderator", "verified_member", "member", "pending"] as GroupRole[]).map((role) => {
              const membersInRole = sortedMembers.filter((m) => m.role === role);
              if (membersInRole.length === 0) return null;
              const RoleIcon = roleIcons[role];

              return (
                <div key={role}>
                  <div className="flex items-center gap-2 px-6 py-2">
                    <RoleIcon size={14} style={{ color: roleColors[role] }} />
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        color: "var(--echo-text-tertiary)",
                        letterSpacing: "0.5px",
                        textTransform: "uppercase",
                      }}
                    >
                      {roleLabels[role]}s — {membersInRole.length}
                    </span>
                  </div>
                  {membersInRole.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center gap-3 px-6 py-2.5 hover:bg-[var(--echo-bg-surface)] transition-colors relative"
                    >
                      <EchoAvatar
                        name={member.name}
                        size="md"
                        online={member.online}
                        gradient={member.gradient}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5">
                          <span
                            className="truncate"
                            style={{
                              fontSize: 15,
                              fontWeight: 500,
                              color: "var(--echo-text-primary)",
                            }}
                          >
                            {member.name}
                          </span>
                          {member.isVerified && (
                            <ShieldCheck size={14} style={{ color: "var(--echo-success)" }} />
                          )}
                        </div>
                        <div className="flex items-center gap-2 mt-0.5">
                          <span
                            className="rounded-full px-1.5 py-0.5"
                            style={{
                              fontSize: 10,
                              fontWeight: 600,
                              color: roleColors[member.role],
                              backgroundColor: `${roleColors[member.role]}15`,
                            }}
                          >
                            {roleLabels[member.role]}
                          </span>
                          <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                            Trust: {member.trustScore}
                          </span>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <div
                          className="flex items-center gap-1 rounded-full px-2 py-0.5"
                          style={{
                            backgroundColor:
                              member.trustScore >= 80
                                ? "var(--echo-success-light)"
                                : member.trustScore >= 60
                                ? "var(--echo-warning-light)"
                                : "var(--echo-error-light)",
                          }}
                        >
                          <span
                            style={{
                              fontSize: 12,
                              fontWeight: 600,
                              color:
                                member.trustScore >= 80
                                  ? "var(--echo-success)"
                                  : member.trustScore >= 60
                                  ? "var(--echo-warning)"
                                  : "var(--echo-error)",
                            }}
                          >
                            {member.trustScore}
                          </span>
                        </div>
                        <button
                          onClick={() => setShowModMenu(showModMenu === member.id ? null : member.id)}
                          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
                          style={{ width: 32, height: 32 }}
                        >
                          <MoreHorizontal size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                        </button>
                      </div>

                      {/* Mod menu */}
                      {showModMenu === member.id && member.role !== "owner" && (
                        <motion.div
                          initial={{ opacity: 0, scale: 0.95 }}
                          animate={{ opacity: 1, scale: 1 }}
                          className="absolute right-6 top-12 z-20 rounded-xl overflow-hidden"
                          style={{
                            backgroundColor: "var(--echo-bg-elevated)",
                            boxShadow: "var(--echo-shadow-lg)",
                            border: "1px solid var(--echo-border)",
                            width: 180,
                          }}
                        >
                          {[
                            { label: "Change Role", color: "var(--echo-text-primary)" },
                            { label: "Mute Member", color: "var(--echo-warning)" },
                            { label: "Warn Member", color: "var(--echo-warning)" },
                            { label: "Ban Member", color: "var(--echo-error)" },
                          ].map((action) => (
                            <button
                              key={action.label}
                              onClick={() => setShowModMenu(null)}
                              className="w-full px-4 py-2.5 bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                              style={{ fontSize: 14, color: action.color }}
                            >
                              {action.label}
                            </button>
                          ))}
                        </motion.div>
                      )}
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        )}

        {/* STATS TAB */}
        {activeTab === 2 && (
          <div className="px-6 py-4">
            {/* Trust ring */}
            <div className="flex flex-col items-center mb-6">
              <TrustRing
                score={group.stats.averageTrustScore}
                size={120}
                strokeWidth={8}
                label="Avg Trust Score"
              />
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-2 gap-3 mb-6">
              {[
                {
                  icon: ShieldCheck,
                  label: "Verified Members",
                  value: `${group.stats.verifiedPercentage}%`,
                  color: "var(--echo-success)",
                },
                {
                  icon: Activity,
                  label: "Activity Level",
                  value: group.stats.activityLevel.replace("_", " "),
                  color: "var(--echo-accent)",
                },
                {
                  icon: TrendingUp,
                  label: "Total Messages",
                  value: group.stats.totalMessages.toLocaleString(),
                  color: "var(--echo-primary)",
                },
                {
                  icon: Users,
                  label: "Active Members",
                  value: `${group.stats.activeMembers}`,
                  color: "var(--echo-info)",
                },
              ].map((stat) => {
                const Icon = stat.icon;
                return (
                  <div
                    key={stat.label}
                    className="rounded-2xl p-4"
                    style={{
                      backgroundColor: "var(--echo-bg-surface)",
                    }}
                  >
                    <Icon size={20} style={{ color: stat.color, marginBottom: 8 }} />
                    <div style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)", lineHeight: 1 }}>
                      {stat.value}
                    </div>
                    <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 4 }}>
                      {stat.label}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Security Rating */}
            <div
              className="rounded-2xl p-4 flex items-center justify-between"
              style={{
                backgroundColor: "var(--echo-bg-surface)",
              }}
            >
              <div>
                <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                  Security Rating
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                  Based on verification and trust metrics
                </div>
              </div>
              <SecurityBadge rating={group.stats.securityRating} />
            </div>

            {/* Creation info */}
            <div className="mt-6 text-center">
              <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                Created {new Date(group.createdAt).toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" })}
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}