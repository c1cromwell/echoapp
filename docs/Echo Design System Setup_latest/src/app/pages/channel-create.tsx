import React, { useState } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, Radio, Globe, Lock, Eye, Check,
  MessageCircle, BarChart2, Image, X,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import {
  type ChannelType, type VisibilityMode,
  channelTypeLabels, channelTypeIcons, visibilityLabels,
  coverGradients,
} from "../components/broadcast-channel-data";

const channelTypes: { type: ChannelType; desc: string }[] = [
  { type: "news", desc: "Breaking news, media coverage, current events" },
  { type: "announcement", desc: "One-way broadcasts, updates, releases" },
  { type: "educational", desc: "Courses, tutorials, structured learning" },
  { type: "community", desc: "Two-way discussions, shared interests" },
];

const visibilityModes: { mode: VisibilityMode; icon: React.ReactNode; desc: string }[] = [
  { mode: "public", icon: <Globe size={18} />, desc: "Anyone can discover and subscribe" },
  { mode: "semi_private", icon: <Eye size={18} />, desc: "Discoverable, but requires approval to join" },
  { mode: "private", icon: <Lock size={18} />, desc: "Invite-only, hidden from search" },
];

const gradientOptions = ["primary", "secondary", "warm", "cool", "accent"] as const;

export function ChannelCreatePage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [channelType, setChannelType] = useState<ChannelType>("community");
  const [visibility, setVisibility] = useState<VisibilityMode>("public");
  const [coverGradient, setCoverGradient] = useState<typeof gradientOptions[number]>("primary");
  const [tags, setTags] = useState("");
  const [allowComments, setAllowComments] = useState(true);
  const [allowPolls, setAllowPolls] = useState(true);
  const [allowMedia, setAllowMedia] = useState(true);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleCreate = () => {
    setShowSuccess(true);
    setTimeout(() => {
      navigate("/", { replace: true });
    }, 2000);
  };

  const canProceed = () => {
    if (step === 0) return name.trim().length >= 3;
    if (step === 1) return true;
    if (step === 2) return true;
    return true;
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
          onClick={() => step > 0 ? setStep(step - 1) : navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          style={{ width: 44, height: 44 }}
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <h1 style={{ fontSize: 18, fontWeight: 600, color: "var(--echo-text-primary)", margin: 0, flex: 1 }}>
          Create Channel
        </h1>
        <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
          {step + 1} of 3
        </span>
      </div>

      {/* Progress bar */}
      <div className="shrink-0" style={{ height: 3, backgroundColor: "var(--echo-bg-surface)" }}>
        <motion.div
          animate={{ width: `${((step + 1) / 3) * 100}%` }}
          transition={{ type: "spring", damping: 25, stiffness: 200 }}
          style={{
            height: "100%",
            background: "linear-gradient(90deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            borderRadius: 2,
          }}
        />
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-4 pt-4 pb-6">
        <AnimatePresence mode="wait">
          {step === 0 && (
            <motion.div
              key="step0"
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -30 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)", marginBottom: 4, marginTop: 0 }}>
                Basic Info
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-tertiary)", marginBottom: 20, marginTop: 0 }}>
                Name your channel and add a description
              </p>

              {/* Name */}
              <label style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", marginBottom: 6, display: "block" }}>
                Channel Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value.slice(0, 100))}
                placeholder="e.g., Tech Pulse Daily"
                className="w-full rounded-xl px-4 border-none outline-none mb-1"
                style={{
                  height: 48,
                  fontSize: 16,
                  color: "var(--echo-text-primary)",
                  backgroundColor: "var(--echo-bg-surface)",
                }}
              />
              <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginBottom: 16, display: "block" }}>
                {name.length}/100 characters
              </span>

              {/* Description */}
              <label style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", marginBottom: 6, display: "block" }}>
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value.slice(0, 500))}
                placeholder="What's your channel about?"
                rows={4}
                className="w-full rounded-xl px-4 py-3 border-none outline-none resize-none"
                style={{
                  fontSize: 15,
                  color: "var(--echo-text-primary)",
                  backgroundColor: "var(--echo-bg-surface)",
                  fontFamily: "inherit",
                }}
              />
              <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginBottom: 16, display: "block" }}>
                {description.length}/500 characters
              </span>

              {/* Tags */}
              <label style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", marginBottom: 6, display: "block" }}>
                Tags (comma separated, max 10)
              </label>
              <input
                type="text"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder="tech, AI, news"
                className="w-full rounded-xl px-4 border-none outline-none"
                style={{
                  height: 48,
                  fontSize: 16,
                  color: "var(--echo-text-primary)",
                  backgroundColor: "var(--echo-bg-surface)",
                }}
              />

              {/* Cover gradient */}
              <label style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-secondary)", marginBottom: 8, display: "block", marginTop: 16 }}>
                Cover Style
              </label>
              <div className="flex gap-3 mb-2">
                {gradientOptions.map((g) => (
                  <button
                    key={g}
                    onClick={() => setCoverGradient(g)}
                    className="rounded-xl border-none cursor-pointer"
                    style={{
                      width: 48,
                      height: 48,
                      background: coverGradients[g],
                      border: coverGradient === g ? "3px solid var(--echo-text-primary)" : "3px solid transparent",
                      boxShadow: coverGradient === g ? "var(--echo-shadow-md)" : "none",
                    }}
                  />
                ))}
              </div>
            </motion.div>
          )}

          {step === 1 && (
            <motion.div
              key="step1"
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -30 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)", marginBottom: 4, marginTop: 0 }}>
                Channel Type
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-tertiary)", marginBottom: 20, marginTop: 0 }}>
                Choose the type that best fits your content
              </p>

              <div className="flex flex-col gap-3 mb-6">
                {channelTypes.map((ct) => (
                  <button
                    key={ct.type}
                    onClick={() => setChannelType(ct.type)}
                    className="flex items-center gap-3 rounded-xl px-4 py-3 border-none cursor-pointer text-left transition-all"
                    style={{
                      backgroundColor: channelType === ct.type ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                      border: channelType === ct.type ? "2px solid var(--echo-primary)" : "2px solid transparent",
                    }}
                  >
                    <span style={{ fontSize: 28 }}>{channelTypeIcons[ct.type]}</span>
                    <div className="flex-1">
                      <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                        {channelTypeLabels[ct.type]}
                      </div>
                      <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)", lineHeight: "18px" }}>
                        {ct.desc}
                      </div>
                    </div>
                    {channelType === ct.type && (
                      <Check size={18} style={{ color: "var(--echo-primary)" }} />
                    )}
                  </button>
                ))}
              </div>

              <h3 style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-text-primary)", marginBottom: 12, marginTop: 0 }}>
                Visibility
              </h3>
              <div className="flex flex-col gap-3">
                {visibilityModes.map((vm) => (
                  <button
                    key={vm.mode}
                    onClick={() => setVisibility(vm.mode)}
                    className="flex items-center gap-3 rounded-xl px-4 py-3 border-none cursor-pointer text-left transition-all"
                    style={{
                      backgroundColor: visibility === vm.mode ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                      border: visibility === vm.mode ? "2px solid var(--echo-primary)" : "2px solid transparent",
                    }}
                  >
                    <div
                      className="flex items-center justify-center rounded-xl"
                      style={{
                        width: 40,
                        height: 40,
                        backgroundColor: visibility === vm.mode ? "var(--echo-primary)" : "var(--echo-bg-elevated)",
                        color: visibility === vm.mode ? "white" : "var(--echo-text-tertiary)",
                      }}
                    >
                      {vm.icon}
                    </div>
                    <div className="flex-1">
                      <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                        {visibilityLabels[vm.mode]}
                      </div>
                      <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)", lineHeight: "18px" }}>
                        {vm.desc}
                      </div>
                    </div>
                    {visibility === vm.mode && (
                      <Check size={18} style={{ color: "var(--echo-primary)" }} />
                    )}
                  </button>
                ))}
              </div>
            </motion.div>
          )}

          {step === 2 && (
            <motion.div
              key="step2"
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -30 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)", marginBottom: 4, marginTop: 0 }}>
                Content Settings
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-tertiary)", marginBottom: 20, marginTop: 0 }}>
                Configure what content is allowed
              </p>

              <div className="flex flex-col gap-3 mb-6">
                <ToggleRow
                  icon={<MessageCircle size={18} />}
                  label="Allow Comments"
                  description="Subscribers can reply to posts"
                  enabled={allowComments}
                  onToggle={() => setAllowComments(!allowComments)}
                />
                <ToggleRow
                  icon={<BarChart2 size={18} />}
                  label="Allow Polls"
                  description="Create polls for subscriber engagement"
                  enabled={allowPolls}
                  onToggle={() => setAllowPolls(!allowPolls)}
                />
                <ToggleRow
                  icon={<Image size={18} />}
                  label="Allow Media"
                  description="Post images, videos, and files"
                  enabled={allowMedia}
                  onToggle={() => setAllowMedia(!allowMedia)}
                />
              </div>

              {/* Preview */}
              <h3 style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-text-primary)", marginBottom: 12, marginTop: 0 }}>
                Preview
              </h3>
              <div
                className="rounded-2xl overflow-hidden"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  border: "1px solid var(--echo-border)",
                  boxShadow: "var(--echo-shadow-md)",
                }}
              >
                <div
                  className="flex items-center justify-center"
                  style={{ height: 60, background: coverGradients[coverGradient] }}
                >
                  <span style={{ fontSize: 32 }}>{channelTypeIcons[channelType]}</span>
                </div>
                <div className="p-4">
                  <div className="flex items-center gap-2 mb-1">
                    <span style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                      {name || "Channel Name"}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 mb-2">
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                      {channelTypeLabels[channelType]}
                    </span>
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>·</span>
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                      {visibilityLabels[visibility]}
                    </span>
                  </div>
                  {description && (
                    <p style={{ fontSize: 13, color: "var(--echo-text-secondary)", lineHeight: "20px", margin: 0 }}>
                      {description.slice(0, 100)}{description.length > 100 ? "..." : ""}
                    </p>
                  )}
                  {tags && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {tags.split(",").slice(0, 4).map((tag, i) => (
                        <span
                          key={i}
                          className="rounded-full px-2 py-0.5"
                          style={{ fontSize: 11, fontWeight: 500, color: "var(--echo-primary)", backgroundColor: "var(--echo-primary-subtle)" }}
                        >
                          #{tag.trim()}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Bottom action */}
      <div
        className="px-4 pb-8 pt-3 shrink-0"
        style={{ borderTop: "1px solid var(--echo-border-light)" }}
      >
        <button
          onClick={() => {
            if (step < 2) setStep(step + 1);
            else handleCreate();
          }}
          disabled={!canProceed()}
          className="w-full rounded-2xl py-3.5 border-none cursor-pointer transition-opacity"
          style={{
            fontSize: 16,
            fontWeight: 600,
            color: "white",
            background: canProceed()
              ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
              : "var(--echo-text-tertiary)",
            opacity: canProceed() ? 1 : 0.5,
            boxShadow: canProceed() ? "var(--echo-shadow-primary)" : "none",
          }}
        >
          {step < 2 ? "Continue" : "Create Channel"}
        </button>
      </div>

      {/* Success overlay */}
      <AnimatePresence>
        {showSuccess && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex flex-col items-center justify-center"
            style={{ backgroundColor: "var(--echo-bg-primary)" }}
          >
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ type: "spring", damping: 15, stiffness: 200, delay: 0.2 }}
              className="flex items-center justify-center rounded-full mb-6"
              style={{
                width: 80,
                height: 80,
                background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
              }}
            >
              <Radio size={40} color="white" />
            </motion.div>
            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              style={{ fontSize: 24, fontWeight: 700, color: "var(--echo-text-primary)", textAlign: "center", margin: 0 }}
            >
              Channel Created!
            </motion.h2>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.6 }}
              style={{ fontSize: 15, color: "var(--echo-text-tertiary)", textAlign: "center", marginTop: 8 }}
            >
              Start posting content to your subscribers
            </motion.p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

function ToggleRow({
  icon,
  label,
  description,
  enabled,
  onToggle,
}: {
  icon: React.ReactNode;
  label: string;
  description: string;
  enabled: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      onClick={onToggle}
      className="flex items-center gap-3 rounded-xl px-4 py-3 border-none cursor-pointer text-left"
      style={{ backgroundColor: "var(--echo-bg-surface)" }}
    >
      <div
        className="flex items-center justify-center rounded-xl shrink-0"
        style={{
          width: 40,
          height: 40,
          backgroundColor: enabled ? "var(--echo-primary-subtle)" : "var(--echo-bg-elevated)",
          color: enabled ? "var(--echo-primary)" : "var(--echo-text-tertiary)",
        }}
      >
        {icon}
      </div>
      <div className="flex-1">
        <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>{label}</div>
        <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>{description}</div>
      </div>
      <div
        className="flex items-center justify-center rounded-full transition-all"
        style={{
          width: 48,
          height: 28,
          backgroundColor: enabled ? "var(--echo-primary)" : "var(--echo-border)",
          padding: 2,
        }}
      >
        <motion.div
          animate={{ x: enabled ? 10 : -10 }}
          className="rounded-full"
          style={{ width: 24, height: 24, backgroundColor: "white", boxShadow: "var(--echo-shadow-sm)" }}
        />
      </div>
    </button>
  );
}
