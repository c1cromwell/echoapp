import React, { useState } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, Search, ShieldCheck, Lock, Crown, Users, TrendingUp,
  Star, Radio, Eye, Sparkles, Globe,
} from "lucide-react";
import {
  discoverableChannels, mockChannels,
  channelTypeIcons, channelTypeLabels, visibilityLabels,
  coverGradients, formatCount, timeAgo,
  type BroadcastChannel,
} from "../components/broadcast-channel-data";

const allDiscoverable = [
  ...discoverableChannels,
  ...mockChannels.filter((c) => c.visibilityMode === "public" && !c.isSubscribed),
];

const categories = ["All", "News", "Educational", "Community", "Announcement"];
const sortOptions = ["Trending", "Most Subscribers", "Newest", "Highest Trust"];

export function ChannelDiscoverPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [activeCategory, setActiveCategory] = useState("All");
  const [joinedIds, setJoinedIds] = useState<Set<string>>(new Set());
  const [requestedIds, setRequestedIds] = useState<Set<string>>(new Set());

  const filtered = allDiscoverable.filter((c) => {
    const matchesSearch =
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.tags.some((t) => t.toLowerCase().includes(search.toLowerCase())) ||
      c.topic.toLowerCase().includes(search.toLowerCase());
    const matchesCategory =
      activeCategory === "All" ||
      c.channelType === activeCategory.toLowerCase();
    return matchesSearch && matchesCategory;
  });

  const handleJoin = (channel: BroadcastChannel) => {
    if (channel.visibilityMode === "semi_private") {
      setRequestedIds((prev) => new Set(prev).add(channel.id));
    } else if (channel.visibilityMode === "public") {
      setJoinedIds((prev) => new Set(prev).add(channel.id));
    }
  };

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
          Discover Channels
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
            placeholder="Search channels or topics..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent border-none outline-none"
            style={{ fontSize: 16, color: "var(--echo-text-primary)" }}
          />
        </div>

        {/* Category chips */}
        <div className="flex gap-2 overflow-x-auto pb-3" style={{ scrollbarWidth: "none" }}>
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

      {/* Featured Section */}
      {activeCategory === "All" && search === "" && (
        <div className="px-4 pb-3">
          <div className="flex items-center gap-1.5 mb-2">
            <Sparkles size={14} style={{ color: "var(--echo-warning)" }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)" }}>
              Featured
            </span>
          </div>
          <div className="flex gap-3 overflow-x-auto pb-1" style={{ scrollbarWidth: "none" }}>
            {allDiscoverable
              .filter((c) => c.trustScore >= 85)
              .slice(0, 3)
              .map((channel) => (
                <button
                  key={`feat-${channel.id}`}
                  onClick={() => navigate(`/channels/${channel.id}`)}
                  className="shrink-0 rounded-2xl overflow-hidden border-none cursor-pointer text-left p-0"
                  style={{
                    width: 200,
                    backgroundColor: "var(--echo-bg-elevated)",
                    border: "1px solid var(--echo-border)",
                    boxShadow: "var(--echo-shadow-sm)",
                  }}
                >
                  <div
                    className="flex items-center justify-center"
                    style={{
                      height: 60,
                      background: coverGradients[channel.coverGradient],
                    }}
                  >
                    <span style={{ fontSize: 28 }}>
                      {channelTypeIcons[channel.channelType]}
                    </span>
                  </div>
                  <div className="p-3">
                    <div className="flex items-center gap-1 mb-0.5">
                      <span
                        className="truncate"
                        style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}
                      >
                        {channel.name}
                      </span>
                      {channel.verificationStatus !== "unverified" && (
                        <ShieldCheck size={12} style={{ color: "var(--echo-success)" }} />
                      )}
                    </div>
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                      {formatCount(channel.subscriberCount)} subscribers
                    </span>
                  </div>
                </button>
              ))}
          </div>
        </div>
      )}

      {/* Divider */}
      <div
        className="mx-4"
        style={{ height: 1, backgroundColor: "var(--echo-border-light)" }}
      />

      {/* Results */}
      <div className="flex-1 overflow-y-auto px-4 pb-6 pt-2">
        <div className="flex flex-col gap-3">
          {filtered.map((channel) => {
            const isJoined = joinedIds.has(channel.id) || mockChannels.some((c) => c.id === channel.id && c.isSubscribed);
            const isRequested = requestedIds.has(channel.id);

            return (
              <div
                key={channel.id}
                className="rounded-2xl overflow-hidden"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  border: "1px solid var(--echo-border)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                {/* Cover strip */}
                <div
                  style={{
                    height: 6,
                    background: coverGradients[channel.coverGradient],
                  }}
                />
                <div className="p-4">
                  <div className="flex items-start gap-3 mb-2">
                    <div
                      className="rounded-xl flex items-center justify-center shrink-0"
                      style={{
                        width: 48,
                        height: 48,
                        background: coverGradients[channel.coverGradient],
                      }}
                    >
                      <span style={{ fontSize: 24 }}>
                        {channelTypeIcons[channel.channelType]}
                      </span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <span
                          className="truncate"
                          style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}
                        >
                          {channel.name}
                        </span>
                        {channel.verificationStatus === "verified" && (
                          <ShieldCheck size={14} style={{ color: "var(--echo-success)" }} />
                        )}
                        {channel.verificationStatus === "premium" && (
                          <Crown size={14} style={{ color: "var(--echo-warning)" }} />
                        )}
                        {channel.visibilityMode === "private" && (
                          <Lock size={13} style={{ color: "var(--echo-warning)" }} />
                        )}
                      </div>
                      <div className="flex items-center gap-2 mt-0.5">
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {channelTypeIcons[channel.channelType]} {channelTypeLabels[channel.channelType]}
                        </span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>·</span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {formatCount(channel.subscriberCount)} subs
                        </span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>·</span>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {visibilityLabels[channel.visibilityMode]}
                        </span>
                      </div>
                    </div>
                  </div>

                  <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "20px", margin: "0 0 10px" }}>
                    {channel.description}
                  </p>

                  {/* Tags */}
                  {channel.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 mb-3">
                      {channel.tags.slice(0, 4).map((tag) => (
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
                        Trust: {channel.trustScore}
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Radio size={13} style={{ color: "var(--echo-accent)" }} />
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        {channel.totalPostCount} posts
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Globe size={13} style={{ color: "var(--echo-text-tertiary)" }} />
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        {channel.language.toUpperCase()}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Action */}
                <div
                  className="flex items-center justify-between px-4 py-2.5"
                  style={{ borderTop: "1px solid var(--echo-border-light)", backgroundColor: "var(--echo-bg-secondary)" }}
                >
                  <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                    by {channel.creatorName}
                  </span>
                  <div>
                    {isJoined ? (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          navigate(`/channels/${channel.id}`);
                        }}
                        className="rounded-full px-4 py-1.5 border-none cursor-pointer"
                        style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-success)", backgroundColor: "var(--echo-success-light)" }}
                      >
                        Subscribed · Open
                      </button>
                    ) : isRequested ? (
                      <button
                        className="rounded-full px-4 py-1.5 border-none cursor-not-allowed"
                        style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-warning)", backgroundColor: "var(--echo-warning-light)" }}
                      >
                        Pending Approval
                      </button>
                    ) : (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleJoin(channel);
                        }}
                        className="rounded-full px-4 py-1.5 border-none cursor-pointer"
                        style={{
                          fontSize: 13,
                          fontWeight: 600,
                          color: "white",
                          background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                        }}
                      >
                        {channel.visibilityMode === "private"
                          ? "Invite Only"
                          : channel.visibilityMode === "semi_private"
                          ? "Request to Join"
                          : "Subscribe"}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}

          {filtered.length === 0 && (
            <div className="flex flex-col items-center py-16">
              <Radio size={32} style={{ color: "var(--echo-text-tertiary)", marginBottom: 8 }} />
              <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
                No channels found
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
