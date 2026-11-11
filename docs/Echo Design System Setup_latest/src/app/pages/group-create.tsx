import React, { useState } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, Eye, Lock, EyeOff, Users, ChevronRight, Check,
  Globe, Briefcase, BookOpen, Home, ShoppingCart, Landmark, HeartHandshake, Pin,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import type { GroupType, GroupCategory, ApprovalMode } from "../components/group-data";

const steps = ["Type", "Details", "Rules", "Review"];

const groupTypes: { type: GroupType; icon: React.ElementType; title: string; description: string }[] = [
  { type: "public", icon: Eye, title: "Public", description: "Anyone can find and join this group" },
  { type: "private", icon: Lock, title: "Private", description: "Visible in search, but requires approval to join" },
  { type: "secret", icon: EyeOff, title: "Secret", description: "Invite-only — hidden from search and discovery" },
];

const categories: { id: GroupCategory; emoji: string; label: string; icon: React.ElementType }[] = [
  { id: "social", emoji: "🌐", label: "Social", icon: Globe },
  { id: "professional", emoji: "💼", label: "Professional", icon: Briefcase },
  { id: "education", emoji: "📚", label: "Education", icon: BookOpen },
  { id: "community", emoji: "🏘️", label: "Community", icon: Home },
  { id: "marketplace", emoji: "🛒", label: "Marketplace", icon: ShoppingCart },
  { id: "governance", emoji: "🏛️", label: "Governance", icon: Landmark },
  { id: "support", emoji: "🤝", label: "Support", icon: HeartHandshake },
  { id: "other", emoji: "📌", label: "Other", icon: Pin },
];

const approvalModes: { mode: ApprovalMode; title: string; desc: string }[] = [
  { mode: "auto", title: "Automatic", desc: "Anyone meeting requirements can join" },
  { mode: "manual", title: "Manual Approval", desc: "Admins review each request" },
  { mode: "vote", title: "Member Vote", desc: "Existing members vote on new joiners" },
];

export function GroupCreatePage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const [type, setType] = useState<GroupType | null>(null);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState<GroupCategory | null>(null);
  const [maxMembers, setMaxMembers] = useState("100");
  const [minTrust, setMinTrust] = useState("30");
  const [approval, setApproval] = useState<ApprovalMode>("auto");
  const [rules, setRules] = useState<string[]>([""]);
  const [tags, setTags] = useState("");

  const canNext = () => {
    if (step === 0) return !!type && !!category;
    if (step === 1) return name.trim().length >= 3;
    return true;
  };

  const handleNext = () => {
    if (step < 3) setStep(step + 1);
    else {
      // "Create" the group — navigate to groups list
      navigate("/", { replace: true });
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
          onClick={() => (step > 0 ? setStep(step - 1) : navigate(-1))}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer p-1"
          style={{ width: 44, height: 44 }}
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-primary)" }} />
        </button>
        <h1
          style={{
            fontSize: 18,
            fontWeight: 600,
            color: "var(--echo-text-primary)",
            margin: 0,
            flex: 1,
          }}
        >
          Create Group
        </h1>
        <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)", marginRight: 8 }}>
          {step + 1} of {steps.length}
        </span>
      </div>

      {/* Progress bar */}
      <div className="px-4 pt-3 pb-1">
        <div
          className="flex rounded-full overflow-hidden"
          style={{ height: 3, backgroundColor: "var(--echo-border)" }}
        >
          <div
            className="rounded-full transition-all duration-300"
            style={{
              width: `${((step + 1) / steps.length) * 100}%`,
              background: "linear-gradient(90deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            }}
          />
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <AnimatePresence mode="wait">
          {step === 0 && (
            <motion.div
              key="step0"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              {/* Group Type */}
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px", lineHeight: "28px" }}>
                Choose group type
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: "0 0 20px" }}>
                This determines who can find and join your group
              </p>

              <div className="flex flex-col gap-3 mb-8">
                {groupTypes.map((gt) => {
                  const Icon = gt.icon;
                  const isSelected = type === gt.type;
                  return (
                    <button
                      key={gt.type}
                      onClick={() => setType(gt.type)}
                      className="flex items-center gap-4 w-full rounded-2xl px-4 text-left cursor-pointer transition-all"
                      style={{
                        height: 72,
                        backgroundColor: isSelected ? "var(--echo-primary-subtle)" : "var(--echo-bg-elevated)",
                        border: `2px solid ${isSelected ? "var(--echo-primary)" : "var(--echo-border)"}`,
                      }}
                    >
                      <div
                        className="flex items-center justify-center rounded-xl shrink-0"
                        style={{
                          width: 40,
                          height: 40,
                          backgroundColor: isSelected ? "var(--echo-primary)" : "var(--echo-bg-surface)",
                        }}
                      >
                        <Icon size={20} color={isSelected ? "white" : "var(--echo-text-secondary)"} />
                      </div>
                      <div className="flex-1">
                        <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>{gt.title}</div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 2 }}>{gt.description}</div>
                      </div>
                      {isSelected && (
                        <div
                          className="flex items-center justify-center rounded-full shrink-0"
                          style={{ width: 22, height: 22, backgroundColor: "var(--echo-primary)" }}
                        >
                          <Check size={14} color="white" strokeWidth={3} />
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>

              {/* Category */}
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px", lineHeight: "28px" }}>
                Category
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: "0 0 16px" }}>
                Helps people discover your group
              </p>

              <div className="grid grid-cols-4 gap-2">
                {categories.map((cat) => {
                  const isSelected = category === cat.id;
                  return (
                    <button
                      key={cat.id}
                      onClick={() => setCategory(cat.id)}
                      className="flex flex-col items-center gap-1.5 py-3 rounded-xl cursor-pointer transition-all"
                      style={{
                        backgroundColor: isSelected ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                        border: `2px solid ${isSelected ? "var(--echo-primary)" : "transparent"}`,
                      }}
                    >
                      <span style={{ fontSize: 22 }}>{cat.emoji}</span>
                      <span style={{ fontSize: 11, fontWeight: 500, color: isSelected ? "var(--echo-primary)" : "var(--echo-text-secondary)" }}>
                        {cat.label}
                      </span>
                    </button>
                  );
                })}
              </div>
            </motion.div>
          )}

          {step === 1 && (
            <motion.div
              key="step1"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 20px", lineHeight: "28px" }}>
                Group details
              </h2>

              {/* Name */}
              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Group Name *
              </label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Tech Enthusiasts"
                className="w-full rounded-xl px-4 outline-none mb-5"
                style={{
                  height: 56,
                  backgroundColor: "var(--echo-bg-surface)",
                  border: "2px solid var(--echo-border)",
                  fontSize: 16,
                  color: "var(--echo-text-primary)",
                }}
                onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
              />

              {/* Description */}
              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value.slice(0, 300))}
                placeholder="What's this group about?"
                rows={3}
                className="w-full rounded-xl px-4 py-3 outline-none resize-none mb-1"
                style={{
                  backgroundColor: "var(--echo-bg-surface)",
                  border: "2px solid var(--echo-border)",
                  fontSize: 16,
                  color: "var(--echo-text-primary)",
                  fontFamily: "inherit",
                  lineHeight: "24px",
                }}
                onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
              />
              <div className="text-right mb-5">
                <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>{description.length}/300</span>
              </div>

              {/* Tags */}
              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Tags (comma-separated)
              </label>
              <input
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder="e.g., tech, programming, AI"
                className="w-full rounded-xl px-4 outline-none mb-5"
                style={{
                  height: 56,
                  backgroundColor: "var(--echo-bg-surface)",
                  border: "2px solid var(--echo-border)",
                  fontSize: 16,
                  color: "var(--echo-text-primary)",
                }}
                onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
              />

              {/* Settings row */}
              <div className="flex flex-col gap-4">
                {/* Max Members */}
                <div className="flex items-center justify-between">
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>Max Members</div>
                    <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Limit based on trust level</div>
                  </div>
                  <input
                    type="number"
                    value={maxMembers}
                    onChange={(e) => setMaxMembers(e.target.value)}
                    className="rounded-lg px-3 text-center outline-none"
                    style={{
                      width: 72,
                      height: 40,
                      backgroundColor: "var(--echo-bg-surface)",
                      border: "2px solid var(--echo-border)",
                      fontSize: 15,
                      fontWeight: 600,
                      color: "var(--echo-text-primary)",
                    }}
                  />
                </div>

                {/* Min Trust Score */}
                <div className="flex items-center justify-between">
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>Min Trust Score</div>
                    <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Required to join (0-100)</div>
                  </div>
                  <input
                    type="number"
                    value={minTrust}
                    onChange={(e) => setMinTrust(e.target.value)}
                    min={0}
                    max={100}
                    className="rounded-lg px-3 text-center outline-none"
                    style={{
                      width: 72,
                      height: 40,
                      backgroundColor: "var(--echo-bg-surface)",
                      border: "2px solid var(--echo-border)",
                      fontSize: 15,
                      fontWeight: 600,
                      color: "var(--echo-text-primary)",
                    }}
                  />
                </div>

                {/* Approval Mode */}
                <div>
                  <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)", marginBottom: 8 }}>
                    Approval Mode
                  </div>
                  <div className="flex flex-col gap-2">
                    {approvalModes.map((am) => (
                      <button
                        key={am.mode}
                        onClick={() => setApproval(am.mode)}
                        className="flex items-center gap-3 w-full rounded-xl px-4 py-3 text-left cursor-pointer transition-all"
                        style={{
                          backgroundColor: approval === am.mode ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                          border: `2px solid ${approval === am.mode ? "var(--echo-primary)" : "transparent"}`,
                        }}
                      >
                        <div
                          className="flex items-center justify-center rounded-full shrink-0"
                          style={{
                            width: 20,
                            height: 20,
                            border: `2px solid ${approval === am.mode ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            backgroundColor: approval === am.mode ? "var(--echo-primary)" : "transparent",
                          }}
                        >
                          {approval === am.mode && (
                            <div className="rounded-full" style={{ width: 6, height: 6, backgroundColor: "white" }} />
                          )}
                        </div>
                        <div>
                          <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>{am.title}</div>
                          <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>{am.desc}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </motion.div>
          )}

          {step === 2 && (
            <motion.div
              key="step2"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px", lineHeight: "28px" }}>
                Group rules
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: "0 0 20px" }}>
                Set guidelines for your members (optional)
              </p>

              <div className="flex flex-col gap-3">
                {rules.map((rule, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <span
                      className="flex items-center justify-center rounded-full shrink-0"
                      style={{
                        width: 24,
                        height: 24,
                        backgroundColor: "var(--echo-primary-subtle)",
                        fontSize: 12,
                        fontWeight: 700,
                        color: "var(--echo-primary)",
                      }}
                    >
                      {i + 1}
                    </span>
                    <input
                      value={rule}
                      onChange={(e) => {
                        const updated = [...rules];
                        updated[i] = e.target.value;
                        setRules(updated);
                      }}
                      placeholder={`Rule ${i + 1}`}
                      className="flex-1 rounded-xl px-4 outline-none"
                      style={{
                        height: 48,
                        backgroundColor: "var(--echo-bg-surface)",
                        border: "2px solid var(--echo-border)",
                        fontSize: 15,
                        color: "var(--echo-text-primary)",
                      }}
                      onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                      onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
                    />
                  </div>
                ))}
              </div>

              {rules.length < 10 && (
                <button
                  onClick={() => setRules([...rules, ""])}
                  className="mt-3 bg-transparent border-none cursor-pointer"
                  style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-primary)" }}
                >
                  + Add rule
                </button>
              )}
            </motion.div>
          )}

          {step === 3 && (
            <motion.div
              key="step3"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 20px", lineHeight: "28px" }}>
                Review your group
              </h2>

              {/* Summary card */}
              <div
                className="rounded-2xl p-5 mb-4"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  border: "1px solid var(--echo-border)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className="flex items-center justify-center rounded-xl"
                    style={{
                      width: 48,
                      height: 48,
                      background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                    }}
                  >
                    <Users size={24} color="white" />
                  </div>
                  <div>
                    <div style={{ fontSize: 18, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                      {name || "Untitled Group"}
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <span
                        className="rounded-full px-2 py-0.5"
                        style={{
                          fontSize: 11,
                          fontWeight: 600,
                          color: "white",
                          backgroundColor: type === "public" ? "var(--echo-success)" : type === "private" ? "var(--echo-warning)" : "var(--echo-error)",
                        }}
                      >
                        {type?.toUpperCase()}
                      </span>
                      {category && (
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {categories.find((c) => c.id === category)?.emoji} {categories.find((c) => c.id === category)?.label}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {description && (
                  <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "20px", margin: "0 0 16px" }}>
                    {description}
                  </p>
                )}

                <div className="flex flex-col gap-2">
                  {[
                    ["Max Members", maxMembers],
                    ["Min Trust Score", minTrust],
                    ["Approval", approvalModes.find((a) => a.mode === approval)?.title || ""],
                    ["Rules", `${rules.filter((r) => r.trim()).length} rules`],
                    ["Tags", tags || "None"],
                  ].map(([label, value]) => (
                    <div key={label} className="flex items-center justify-between">
                      <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>{label}</span>
                      <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>{value}</span>
                    </div>
                  ))}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Bottom CTA */}
      <div className="px-6 pb-10 pt-4 shrink-0">
        <button
          onClick={handleNext}
          disabled={!canNext()}
          className="w-full border-none rounded-xl flex items-center justify-center gap-2 active:scale-[0.98] transition-all"
          style={{
            height: 56,
            background: canNext()
              ? "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))"
              : "var(--echo-bg-surface)",
            color: canNext() ? "white" : "var(--echo-text-tertiary)",
            fontSize: 16,
            fontWeight: 600,
            cursor: canNext() ? "pointer" : "not-allowed",
            boxShadow: canNext() ? "var(--echo-shadow-primary)" : "none",
          }}
        >
          {step === 3 ? "Create Group" : "Continue"}
          {step < 3 && <ChevronRight size={18} />}
        </button>
      </div>
    </div>
  );
}
