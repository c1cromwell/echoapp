import React, { useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ArrowLeft, ShieldCheck, Lock, Crown, Users, Radio, Bell, BellOff,
  Heart, MessageCircle, Share2, Pin, Sparkles, MoreHorizontal, BarChart3,
  Send, Settings, Eye, ChevronDown, X, TriangleAlert, Flag,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import {
  mockChannels, discoverableChannels, mockChannelPosts,
  channelTypeIcons, channelTypeLabels, visibilityLabels,
  coverGradients, formatCount, timeAgo, tierPricing,
  type BroadcastChannel, type ChannelPost, type SubscriptionTier,
} from "../components/broadcast-channel-data";

const allChannels = [...mockChannels, ...discoverableChannels];

export function ChannelDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const channel = allChannels.find((c) => c.id === id);
  const posts = mockChannelPosts[id || ""] || [];

  const [isSubscribed, setIsSubscribed] = useState(channel?.isSubscribed ?? false);
  const [showInfo, setShowInfo] = useState(false);
  const [showTierSheet, setShowTierSheet] = useState(false);
  const [likedPosts, setLikedPosts] = useState<Set<string>>(
    new Set(posts.filter((p) => p.isLikedByUser).map((p) => p.id))
  );
  const [votedPolls, setVotedPolls] = useState<Record<string, number>>(() => {
    const initial: Record<string, number> = {};
    posts.forEach((p) => {
      if (p.poll?.userVoted !== undefined) initial[p.id] = p.poll.userVoted;
    });
    return initial;
  });
  const [notifMode, setNotifMode] = useState(channel?.notificationMode || "all");

  if (!channel) {
    return (
      <div className="flex flex-col h-full items-center justify-center">
        <Radio size={48} style={{ color: "var(--echo-text-tertiary)", marginBottom: 12 }} />
        <p style={{ fontSize: 16, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
          Channel not found
        </p>
        <button
          onClick={() => navigate(-1)}
          className="mt-4 rounded-full px-6 py-2 border-none cursor-pointer"
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: "white",
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
          }}
        >
          Go Back
        </button>
      </div>
    );
  }

  const isAdmin = channel.userRole === "admin";
  const isMod = channel.userRole === "moderator";

  const toggleLike = (postId: string) => {
    setLikedPosts((prev) => {
      const next = new Set(prev);
      if (next.has(postId)) next.delete(postId);
      else next.add(postId);
      return next;
    });
  };

  const handleVote = (postId: string, optionIndex: number) => {
    setVotedPolls((prev) => ({ ...prev, [postId]: optionIndex }));
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div
        className="shrink-0 relative"
        style={{
          background: coverGradients[channel.coverGradient],
        }}
      >
        {/* Nav bar overlay */}
        <div
          className="flex items-center justify-between px-3"
          style={{ height: 96, paddingTop: 44 }}
        >
          <button
            onClick={() => navigate(-1)}
            className="flex items-center justify-center rounded-full border-none cursor-pointer"
            style={{ width: 36, height: 36, backgroundColor: "rgba(0,0,0,0.2)" }}
          >
            <ArrowLeft size={20} color="white" />
          </button>
          <div className="flex items-center gap-2">
            {isAdmin && (
              <button
                onClick={() => navigate(`/channels/${channel.id}/analytics`)}
                className="flex items-center justify-center rounded-full border-none cursor-pointer"
                style={{ width: 36, height: 36, backgroundColor: "rgba(0,0,0,0.2)" }}
              >
                <BarChart3 size={18} color="white" />
              </button>
            )}
            <button
              onClick={() => setShowInfo(true)}
              className="flex items-center justify-center rounded-full border-none cursor-pointer"
              style={{ width: 36, height: 36, backgroundColor: "rgba(0,0,0,0.2)" }}
            >
              <MoreHorizontal size={18} color="white" />
            </button>
          </div>
        </div>

        {/* Channel info overlay */}
        <div className="px-4 pb-4">
          <div className="flex items-center gap-2 mb-1">
            <span style={{ fontSize: 24 }}>{channelTypeIcons[channel.channelType]}</span>
            <h1 style={{ fontSize: 22, fontWeight: 700, color: "white", margin: 0, lineHeight: "28px" }}>
              {channel.name}
            </h1>
            {channel.verificationStatus === "verified" && (
              <ShieldCheck size={18} color="white" style={{ opacity: 0.9 }} />
            )}
            {channel.verificationStatus === "premium" && (
              <Crown size={18} color="#FFD700" />
            )}
          </div>
          <div className="flex items-center gap-3">
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.85)" }}>
              {formatCount(channel.subscriberCount)} subscribers
            </span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.85)" }}>·</span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.85)" }}>
              {channel.totalPostCount} posts
            </span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.85)" }}>·</span>
            <span style={{ fontSize: 13, color: "rgba(255,255,255,0.85)" }}>
              {visibilityLabels[channel.visibilityMode]}
            </span>
          </div>
        </div>
      </div>

      {/* Action bar */}
      <div
        className="flex items-center gap-2 px-4 py-2.5 shrink-0"
        style={{
          backgroundColor: "var(--echo-bg-elevated)",
          borderBottom: "1px solid var(--echo-border-light)",
        }}
      >
        {isSubscribed ? (
          <>
            <button
              onClick={() => {
                const modes: ("all" | "important" | "none")[] = ["all", "important", "none"];
                const idx = modes.indexOf(notifMode);
                setNotifMode(modes[(idx + 1) % modes.length]);
              }}
              className="flex items-center gap-1.5 rounded-full px-3 py-1.5 border-none cursor-pointer"
              style={{
                fontSize: 13,
                fontWeight: 500,
                color: notifMode === "none" ? "var(--echo-text-tertiary)" : "var(--echo-primary)",
                backgroundColor: notifMode === "none" ? "var(--echo-bg-surface)" : "var(--echo-primary-subtle)",
              }}
            >
              {notifMode === "none" ? <BellOff size={14} /> : <Bell size={14} />}
              {notifMode === "all" ? "All" : notifMode === "important" ? "Important" : "Muted"}
            </button>
            <button
              onClick={() => setShowTierSheet(true)}
              className="flex items-center gap-1.5 rounded-full px-3 py-1.5 border-none cursor-pointer"
              style={{
                fontSize: 13,
                fontWeight: 500,
                color: "var(--echo-text-secondary)",
                backgroundColor: "var(--echo-bg-surface)",
              }}
            >
              <Users size={14} />
              {channel.userTier === "free" ? "Free" : channel.userTier?.charAt(0).toUpperCase() + channel.userTier?.slice(1)}
            </button>
            <div className="flex-1" />
            {isAdmin && (
              <button
                onClick={() => navigate(`/channels/create?edit=${channel.id}`)}
                className="flex items-center gap-1.5 rounded-full px-3 py-1.5 border-none cursor-pointer"
                style={{
                  fontSize: 13,
                  fontWeight: 500,
                  color: "var(--echo-error)",
                  backgroundColor: "var(--echo-error-light)",
                }}
              >
                <Settings size={14} />
                Manage
              </button>
            )}
          </>
        ) : (
          <>
            <button
              onClick={() => setIsSubscribed(true)}
              className="flex-1 flex items-center justify-center gap-2 rounded-full py-2 border-none cursor-pointer"
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: "white",
                background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              }}
            >
              <Radio size={16} />
              {channel.visibilityMode === "semi_private" ? "Request to Join" : "Subscribe"}
            </button>
          </>
        )}
      </div>

      {/* Posts feed */}
      <div className="flex-1 overflow-y-auto">
        {posts.length === 0 ? (
          <div className="flex flex-col items-center py-16">
            <Radio size={40} style={{ color: "var(--echo-text-tertiary)", marginBottom: 12 }} />
            <p style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
              No posts yet
            </p>
            <p style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
              {isAdmin ? "Create your first post!" : "Check back later for updates"}
            </p>
          </div>
        ) : (
          posts.map((post) => (
            <PostCard
              key={post.id}
              post={post}
              channel={channel}
              isLiked={likedPosts.has(post.id)}
              onToggleLike={() => toggleLike(post.id)}
              votedOption={votedPolls[post.id]}
              onVote={(idx) => handleVote(post.id, idx)}
              isAdmin={isAdmin}
              isMod={isMod}
            />
          ))
        )}
      </div>

      {/* Channel Info Sheet */}
      <AnimatePresence>
        {showInfo && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
              onClick={() => setShowInfo(false)}
            />
            <motion.div
              initial={{ y: "100%", opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: "100%", opacity: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 350 }}
              className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
              style={{ backgroundColor: "var(--echo-bg-elevated)", maxHeight: "70vh" }}
            >
              <div className="flex justify-center pt-3 pb-1">
                <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
              </div>
              <div className="px-5 pb-8 overflow-y-auto" style={{ maxHeight: "calc(70vh - 30px)" }}>
                <div className="flex items-center justify-between mb-4">
                  <span style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                    Channel Info
                  </span>
                  <button onClick={() => setShowInfo(false)} className="bg-transparent border-none cursor-pointer p-1">
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "22px", marginBottom: 16 }}>
                  {channel.description}
                </p>

                <div className="flex flex-col gap-3 mb-4">
                  <InfoRow label="Type" value={`${channelTypeIcons[channel.channelType]} ${channelTypeLabels[channel.channelType]}`} />
                  <InfoRow label="Visibility" value={visibilityLabels[channel.visibilityMode]} />
                  <InfoRow label="Created by" value={channel.creatorName} />
                  <InfoRow label="Created" value={channel.createdAt} />
                  <InfoRow label="Trust Score" value={`${channel.trustScore}/100`} />
                  <InfoRow label="Language" value={channel.language.toUpperCase()} />
                  <InfoRow label="Comments" value={channel.allowComments ? "Enabled" : "Disabled"} />
                  <InfoRow label="Polls" value={channel.allowPolls ? "Enabled" : "Disabled"} />
                </div>

                {channel.tags.length > 0 && (
                  <div className="mb-4">
                    <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", marginBottom: 6, display: "block" }}>
                      Tags
                    </span>
                    <div className="flex flex-wrap gap-1.5">
                      {channel.tags.map((tag) => (
                        <span
                          key={tag}
                          className="rounded-full px-2.5 py-1"
                          style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-primary)", backgroundColor: "var(--echo-primary-subtle)" }}
                        >
                          #{tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {isSubscribed && (
                  <button
                    onClick={() => {
                      setIsSubscribed(false);
                      setShowInfo(false);
                    }}
                    className="w-full rounded-xl py-3 border-none cursor-pointer mt-2"
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      color: "var(--echo-error)",
                      backgroundColor: "var(--echo-error-light)",
                    }}
                  >
                    Unsubscribe
                  </button>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* Tier Sheet */}
      <AnimatePresence>
        {showTierSheet && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
              onClick={() => setShowTierSheet(false)}
            />
            <motion.div
              initial={{ y: "100%", opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: "100%", opacity: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 350 }}
              className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
              style={{ backgroundColor: "var(--echo-bg-elevated)" }}
            >
              <div className="flex justify-center pt-3 pb-1">
                <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
              </div>
              <div className="px-5 pb-8">
                <div className="flex items-center justify-between mb-4">
                  <span style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                    Subscription Tier
                  </span>
                  <button onClick={() => setShowTierSheet(false)} className="bg-transparent border-none cursor-pointer p-1">
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                <div className="flex flex-col gap-2">
                  {(Object.keys(tierPricing) as SubscriptionTier[]).map((tier) => {
                    const info = tierPricing[tier];
                    const isActive = (channel.userTier || "free") === tier;
                    return (
                      <button
                        key={tier}
                        className="flex items-start gap-3 rounded-xl px-4 py-3 border-none cursor-pointer text-left transition-all"
                        style={{
                          backgroundColor: isActive ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                          border: isActive ? "2px solid var(--echo-primary)" : "2px solid transparent",
                        }}
                      >
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                              {info.label}
                            </span>
                            <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-primary)" }}>
                              {info.price}
                            </span>
                            {isActive && (
                              <span
                                className="rounded-full px-2 py-0.5"
                                style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-success)", backgroundColor: "var(--echo-success-light)" }}
                              >
                                Current
                              </span>
                            )}
                          </div>
                          <div className="flex flex-wrap gap-1">
                            {info.features.map((f, i) => (
                              <span key={i} style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                                {i > 0 ? " · " : ""}{f}
                              </span>
                            ))}
                          </div>
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}

// ===== Sub-components =====

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>{label}</span>
      <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>{value}</span>
    </div>
  );
}

interface PostCardProps {
  post: ChannelPost;
  channel: BroadcastChannel;
  isLiked: boolean;
  onToggleLike: () => void;
  votedOption?: number;
  onVote: (idx: number) => void;
  isAdmin: boolean;
  isMod: boolean;
}

function PostCard({ post, channel, isLiked, onToggleLike, votedOption, onVote, isAdmin, isMod }: PostCardProps) {
  const [expanded, setExpanded] = useState(false);
  const isLong = post.content.length > 200;

  return (
    <div
      className="px-4 py-3"
      style={{ borderBottom: "1px solid var(--echo-border-light)" }}
    >
      {/* Post header */}
      <div className="flex items-center gap-2 mb-2">
        <div
          className="rounded-full flex items-center justify-center"
          style={{
            width: 28,
            height: 28,
            background: coverGradients[channel.coverGradient],
          }}
        >
          <span style={{ fontSize: 13 }}>{channelTypeIcons[channel.channelType]}</span>
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5">
            <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
              {post.creatorName}
            </span>
            {post.isPinned && (
              <Pin size={11} style={{ color: "var(--echo-primary)", transform: "rotate(45deg)" }} />
            )}
            {post.isSponsored && (
              <span
                className="rounded-full px-1.5"
                style={{ fontSize: 9, fontWeight: 600, color: "var(--echo-warning)", backgroundColor: "var(--echo-warning-light)", lineHeight: "14px" }}
              >
                Sponsored
              </span>
            )}
            {post.isFeatured && (
              <Sparkles size={11} style={{ color: "var(--echo-warning)" }} />
            )}
          </div>
        </div>
        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
          {timeAgo(post.publishedAt)}
        </span>
      </div>

      {/* Content */}
      <div className="mb-2">
        <p style={{
          fontSize: 15,
          color: "var(--echo-text-primary)",
          lineHeight: "22px",
          margin: 0,
          whiteSpace: "pre-wrap",
        }}>
          {isLong && !expanded ? post.content.slice(0, 200) + "..." : post.content}
        </p>
        {isLong && (
          <button
            onClick={() => setExpanded(!expanded)}
            className="bg-transparent border-none cursor-pointer p-0 mt-1"
            style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-primary)" }}
          >
            {expanded ? "Show less" : "Read more"}
          </button>
        )}
      </div>

      {/* Poll */}
      {post.poll && (
        <div
          className="rounded-xl p-3 mb-2"
          style={{ backgroundColor: "var(--echo-bg-surface)" }}
        >
          <p style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 8px" }}>
            {post.poll.question}
          </p>
          <div className="flex flex-col gap-1.5">
            {post.poll.options.map((opt, i) => {
              const pct = post.poll!.totalVotes > 0
                ? Math.round((opt.votes / post.poll!.totalVotes) * 100)
                : 0;
              const isVoted = votedOption === i;
              const hasVoted = votedOption !== undefined;

              return (
                <button
                  key={i}
                  onClick={() => !hasVoted && onVote(i)}
                  className="relative rounded-lg overflow-hidden border-none cursor-pointer text-left"
                  style={{
                    height: 36,
                    backgroundColor: "var(--echo-bg-elevated)",
                    border: isVoted ? "2px solid var(--echo-primary)" : "1px solid var(--echo-border)",
                  }}
                  disabled={hasVoted}
                >
                  {hasVoted && (
                    <div
                      className="absolute top-0 left-0 h-full rounded-lg"
                      style={{
                        width: `${pct}%`,
                        backgroundColor: isVoted ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                        transition: "width 0.5s ease",
                      }}
                    />
                  )}
                  <div className="relative flex items-center justify-between px-3 h-full">
                    <span style={{ fontSize: 13, fontWeight: isVoted ? 600 : 400, color: "var(--echo-text-primary)" }}>
                      {opt.label}
                    </span>
                    {hasVoted && (
                      <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-text-secondary)" }}>
                        {pct}%
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 6, display: "block" }}>
            {formatCount(post.poll.totalVotes)} votes
          </span>
        </div>
      )}

      {/* Engagement row */}
      <div className="flex items-center gap-4">
        <button
          onClick={onToggleLike}
          className="flex items-center gap-1 bg-transparent border-none cursor-pointer p-0"
        >
          <Heart
            size={16}
            fill={isLiked ? "var(--echo-error)" : "none"}
            style={{ color: isLiked ? "var(--echo-error)" : "var(--echo-text-tertiary)" }}
          />
          <span style={{ fontSize: 12, fontWeight: 500, color: isLiked ? "var(--echo-error)" : "var(--echo-text-tertiary)" }}>
            {formatCount(post.likeCount + (isLiked && !post.isLikedByUser ? 1 : 0) - (!isLiked && post.isLikedByUser ? 1 : 0))}
          </span>
        </button>
        {post.allowReplies && (
          <button className="flex items-center gap-1 bg-transparent border-none cursor-pointer p-0">
            <MessageCircle size={16} style={{ color: "var(--echo-text-tertiary)" }} />
            <span style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-text-tertiary)" }}>
              {formatCount(post.commentCount)}
            </span>
          </button>
        )}
        <button className="flex items-center gap-1 bg-transparent border-none cursor-pointer p-0">
          <Share2 size={16} style={{ color: "var(--echo-text-tertiary)" }} />
          <span style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-text-tertiary)" }}>
            {formatCount(post.shareCount)}
          </span>
        </button>
        <div className="flex-1" />
        {(isAdmin || isMod) && (
          <button className="flex items-center gap-1 bg-transparent border-none cursor-pointer p-0">
            <Flag size={14} style={{ color: "var(--echo-text-tertiary)" }} />
          </button>
        )}
      </div>
    </div>
  );
}