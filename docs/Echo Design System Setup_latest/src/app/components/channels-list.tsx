import React from "react";
import { useNavigate } from "react-router";
import {
  ShieldCheck, Lock, Eye, Users, Radio, Crown, Bell, BellOff,
} from "lucide-react";
import {
  type BroadcastChannel,
  channelTypeIcons,
  coverGradients,
  formatCount,
  timeAgo,
} from "./broadcast-channel-data";

interface ChannelsListProps {
  channels: BroadcastChannel[];
}

export function ChannelsList({ channels }: ChannelsListProps) {
  const navigate = useNavigate();

  if (channels.length === 0) {
    return (
      <div className="flex flex-col items-center py-16">
        <Radio size={32} style={{ color: "var(--echo-text-tertiary)", marginBottom: 8 }} />
        <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
          No channels found
        </p>
        <p style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
          Try a different search
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {channels.map((channel) => (
        <button
          key={channel.id}
          onClick={() => navigate(`/channels/${channel.id}`)}
          className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer transition-colors hover:bg-[var(--echo-bg-surface)] active:bg-[var(--echo-bg-surface)] text-left"
        >
          {/* Channel avatar */}
          <div
            className="rounded-2xl flex items-center justify-center shrink-0"
            style={{
              width: 48,
              height: 48,
              background: coverGradients[channel.coverGradient],
              boxShadow: "0 4px 8px rgba(0,0,0,0.15)",
            }}
          >
            <span style={{ fontSize: 22 }}>{channelTypeIcons[channel.channelType]}</span>
          </div>

          <div className="flex-1 min-w-0">
            {/* Top row: name + badges + time */}
            <div className="flex items-center justify-between mb-0.5">
              <div className="flex items-center gap-1.5 min-w-0">
                <span
                  className="truncate"
                  style={{
                    fontSize: 16,
                    fontWeight: 600,
                    color: "var(--echo-text-primary)",
                    lineHeight: "22px",
                  }}
                >
                  {channel.name}
                </span>
                {channel.verificationStatus === "verified" && (
                  <ShieldCheck size={14} style={{ color: "var(--echo-success)", flexShrink: 0 }} />
                )}
                {channel.verificationStatus === "premium" && (
                  <Crown size={14} style={{ color: "var(--echo-warning)", flexShrink: 0 }} />
                )}
                {channel.visibilityMode === "private" && (
                  <Lock size={12} style={{ color: "var(--echo-warning)", flexShrink: 0 }} />
                )}
              </div>
              <span
                className="shrink-0 ml-2"
                style={{
                  fontSize: 13,
                  fontWeight: 400,
                  color: "var(--echo-text-tertiary)",
                  lineHeight: "18px",
                }}
              >
                {channel.lastPostAt ? timeAgo(channel.lastPostAt) : ""}
              </span>
            </div>

            {/* Bottom row: subscriber count + notification */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span
                  className="truncate"
                  style={{
                    fontSize: 14,
                    fontWeight: 400,
                    color: "var(--echo-text-tertiary)",
                    lineHeight: "20px",
                  }}
                >
                  {formatCount(channel.subscriberCount)} subscribers
                </span>
                <span
                  className="rounded-full px-1.5"
                  style={{
                    fontSize: 10,
                    fontWeight: 600,
                    color: "var(--echo-primary)",
                    backgroundColor: "var(--echo-primary-subtle)",
                    lineHeight: "16px",
                  }}
                >
                  {channelTypeIcons[channel.channelType]} {channel.channelType}
                </span>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                {channel.isMutedByUser && (
                  <BellOff size={13} style={{ color: "var(--echo-text-tertiary)" }} />
                )}
                {channel.userRole === "admin" && (
                  <span
                    className="rounded-full px-1.5"
                    style={{
                      fontSize: 10,
                      fontWeight: 600,
                      color: "var(--echo-error)",
                      backgroundColor: "var(--echo-error-light)",
                      lineHeight: "16px",
                    }}
                  >
                    Admin
                  </span>
                )}
                {channel.userRole === "moderator" && (
                  <span
                    className="rounded-full px-1.5"
                    style={{
                      fontSize: 10,
                      fontWeight: 600,
                      color: "var(--echo-info)",
                      backgroundColor: "var(--echo-info-light)",
                      lineHeight: "16px",
                    }}
                  >
                    Mod
                  </span>
                )}
              </div>
            </div>
          </div>
        </button>
      ))}
    </div>
  );
}
