import React, { useState } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, Search, ShieldCheck, Lock, Eye, Users, TrendingUp, Star,
} from "lucide-react";
import { GroupAvatar } from "../components/echo-avatar";
import { discoverableGroups, mockGroups, categoryIcons, type Group } from "../components/group-data";

const allDiscoverable = [...discoverableGroups, ...mockGroups.filter((g) => g.type === "public")];
const categories = ["All", "Professional", "Social", "Community", "Education"];

export function GroupDiscoverPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [activeCategory, setActiveCategory] = useState("All");

  const filtered = allDiscoverable.filter((g) => {
    const matchesSearch = g.name.toLowerCase().includes(search.toLowerCase()) || g.tags.some((t) => t.toLowerCase().includes(search.toLowerCase()));
    const matchesCategory = activeCategory === "All" || g.category === activeCategory.toLowerCase();
    return matchesSearch && matchesCategory;
  });

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="flex items-center gap-3 px-3 shrink-0"
        style={{
          height: 96,
          paddingTop: 44,
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          style={{ width: 44, height: 44 }}
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <h1 style={{ fontSize: 18, fontWeight: 600, color: "var(--echo-text-primary)", margin: 0, flex: 1 }}>
          Discover Groups
        </h1>
      </div>

      <div className="px-4 pt-3">
        {/* Search */}
        <div
          className="flex items-center gap-2 rounded-xl px-4 mb-3"
          style={{ height: 44, backgroundColor: "var(--echo-bg-surface)" }}
        >
          <Search size={18} style={{ color: "var(--echo-text-tertiary)" }} />
          <input
            type="text"
            placeholder="Search groups or tags..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent border-none outline-none"
            style={{ fontSize: 16, color: "var(--echo-text-primary)" }}
          />
        </div>

        {/* Category chips */}
        <div className="flex gap-2 overflow-x-auto pb-3">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className="rounded-full px-4 py-1.5 border-none cursor-pointer shrink-0 transition-all"
              style={{
                fontSize: 13,
                fontWeight: activeCategory === cat ? 600 : 400,
                color: activeCategory === cat ? "white" : "var(--echo-text-secondary)",
                backgroundColor: activeCategory === cat ? "var(--echo-primary)" : "var(--echo-bg-surface)",
              }}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-y-auto px-4 pb-6">
        <div className="flex flex-col gap-3">
          {filtered.map((group) => {
            const isMember = mockGroups.some((g) => g.id === group.id);
            return (
              <div
                key={group.id}
                className="rounded-2xl overflow-hidden"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  border: "1px solid var(--echo-border)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                <div className="p-4">
                  <div className="flex items-start gap-3 mb-3">
                    <GroupAvatar
                      members={group.members.length > 0
                        ? group.members.slice(0, 4).map((m) => ({ name: m.name }))
                        : [{ name: group.name[0] }, { name: group.name.split(" ")[1]?.[0] || "G" }]
                      }
                      size={48}
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <span className="truncate" style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                          {group.name}
                        </span>
                        {group.type === "private" && <Lock size={13} style={{ color: "var(--echo-warning)" }} />}
                        {group.stats.securityRating === "A" && <ShieldCheck size={14} style={{ color: "var(--echo-success)" }} />}
                      </div>
                      <div className="flex items-center gap-2 mt-0.5">
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {categoryIcons[group.category]} {group.category}
                        </span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>·</span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {group.memberCount} members
                        </span>
                      </div>
                    </div>
                  </div>

                  <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "20px", margin: "0 0 10px" }}>
                    {group.description}
                  </p>

                  {/* Tags */}
                  {group.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 mb-3">
                      {group.tags.slice(0, 4).map((tag) => (
                        <span
                          key={tag}
                          className="rounded-full px-2 py-0.5"
                          style={{ fontSize: 11, fontWeight: 500, color: "var(--echo-primary)", backgroundColor: "var(--echo-primary-subtle)" }}
                        >
                          #{tag}
                        </span>
                      ))}
                    </div>
                  )}

                  {/* Stats row */}
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-1">
                      <ShieldCheck size={13} style={{ color: "var(--echo-success)" }} />
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        {group.stats.verifiedPercentage}% verified
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <TrendingUp size={13} style={{ color: "var(--echo-accent)" }} />
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        {group.stats.activityLevel.replace("_", " ")} activity
                      </span>
                    </div>
                    {group.minTrustScore > 0 && (
                      <div className="flex items-center gap-1">
                        <Star size={13} style={{ color: "var(--echo-warning)" }} />
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          Min trust: {group.minTrustScore}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Action */}
                <div
                  className="flex items-center justify-between px-4 py-2.5"
                  style={{ borderTop: "1px solid var(--echo-border-light)", backgroundColor: "var(--echo-bg-secondary)" }}
                >
                  {group.minTrustScore > 0 && (
                    <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                      Requires trust score {group.minTrustScore}+
                    </span>
                  )}
                  <div className="ml-auto">
                    {isMember ? (
                      <button
                        onClick={() => navigate(`/group-chat/${group.id}`)}
                        className="rounded-full px-4 py-1.5 border-none cursor-pointer"
                        style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-success)", backgroundColor: "var(--echo-success-light)" }}
                      >
                        Joined · Open
                      </button>
                    ) : (
                      <button
                        className="rounded-full px-4 py-1.5 border-none cursor-pointer"
                        style={{
                          fontSize: 13,
                          fontWeight: 600,
                          color: "white",
                          background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                        }}
                      >
                        {group.type === "private" ? "Request to Join" : "Join Group"}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}

          {filtered.length === 0 && (
            <div className="flex flex-col items-center py-16">
              <Users size={32} style={{ color: "var(--echo-text-tertiary)", marginBottom: 8 }} />
              <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
                No groups found
              </p>
              <p style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
                Try a different search or category
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
