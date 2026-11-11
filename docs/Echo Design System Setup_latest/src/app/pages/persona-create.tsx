import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, ChevronRight, Check, Eye, EyeOff, Users } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { EchoAvatar } from "../components/echo-avatar";
import {
  categoryInfo,
  mockMasterIdentity,
  trustLevelLimits,
  type PersonaCategory,
  type VisibilityLevel,
} from "../components/persona-data";

const steps = ["Category", "Profile", "Privacy", "Review"];

const visibilityOptions: { level: VisibilityLevel; icon: React.ElementType; label: string; desc: string }[] = [
  { level: "everyone", icon: Eye, label: "Everyone", desc: "Anyone can discover this persona" },
  { level: "contacts", icon: Users, label: "Contacts Only", desc: "Only granted contacts can see" },
  { level: "nobody", icon: EyeOff, label: "Nobody", desc: "Completely hidden from discovery" },
];

export function PersonaCreatePage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const limits = trustLevelLimits[mockMasterIdentity.trustLevel];

  // Form state
  const [category, setCategory] = useState<PersonaCategory | null>(null);
  const [customCategory, setCustomCategory] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [username, setUsername] = useState("");
  const [bio, setBio] = useState("");
  const [visibility, setVisibility] = useState<VisibilityLevel>("contacts");
  const [readReceipts, setReadReceipts] = useState(true);
  const [typingIndicator, setTypingIndicator] = useState(true);
  const [searchable, setSearchable] = useState(true);
  const [crossPersonaLinking, setCrossPersonaLinking] = useState(false);

  const canNext = () => {
    if (step === 0) return !!category;
    if (step === 1) return displayName.trim().length >= 2 && username.trim().length >= 3;
    return true;
  };

  const handleCreate = () => {
    // Mock creation — navigate back to personas list
    navigate("/personas", { replace: true });
  };

  const gradients: Array<"primary" | "secondary" | "warm" | "cool"> = ["primary", "secondary", "warm", "cool"];
  const selectedGradient = category
    ? gradients[Object.keys(categoryInfo).indexOf(category) % 4]
    : "primary";

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
        <h1 style={{ fontSize: 18, fontWeight: 600, color: "var(--echo-text-primary)", margin: 0, flex: 1 }}>
          Create Persona
        </h1>
        <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)", marginRight: 8 }}>
          {step + 1}/{steps.length}
        </span>
      </div>

      {/* Progress */}
      <div className="px-4 pt-3 pb-1">
        <div className="flex rounded-full overflow-hidden" style={{ height: 3, backgroundColor: "var(--echo-border)" }}>
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
          {/* STEP 0: Category */}
          {step === 0 && (
            <motion.div key="s0" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} transition={{ duration: 0.2 }}>
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px" }}>
                Choose a category
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: "0 0 20px" }}>
                This helps organize your identities
              </p>

              <div className="flex flex-col gap-2.5">
                {(Object.entries(categoryInfo) as [PersonaCategory, typeof categoryInfo[PersonaCategory]][]).map(
                  ([key, info]) => {
                    if (key === "custom" && !limits.customCategories) return null;
                    const isSelected = category === key;
                    return (
                      <button
                        key={key}
                        onClick={() => setCategory(key)}
                        className="flex items-center gap-3 w-full rounded-2xl px-4 text-left cursor-pointer transition-all"
                        style={{
                          height: 64,
                          backgroundColor: isSelected ? "var(--echo-primary-subtle)" : "var(--echo-bg-elevated)",
                          border: `2px solid ${isSelected ? "var(--echo-primary)" : "var(--echo-border)"}`,
                        }}
                      >
                        <span style={{ fontSize: 24 }}>{info.emoji}</span>
                        <div className="flex-1">
                          <div style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            {info.label}
                          </div>
                          <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                            {info.description}
                          </div>
                        </div>
                        {isSelected && (
                          <div
                            className="flex items-center justify-center rounded-full"
                            style={{ width: 22, height: 22, backgroundColor: "var(--echo-primary)" }}
                          >
                            <Check size={14} color="white" strokeWidth={3} />
                          </div>
                        )}
                      </button>
                    );
                  }
                )}
              </div>

              {category === "custom" && (
                <div className="mt-4">
                  <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                    Custom Category Name
                  </label>
                  <input
                    value={customCategory}
                    onChange={(e) => setCustomCategory(e.target.value)}
                    placeholder="e.g., Activism, Side Hustle"
                    className="w-full rounded-xl px-4 outline-none"
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
              )}
            </motion.div>
          )}

          {/* STEP 1: Profile */}
          {step === 1 && (
            <motion.div key="s1" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} transition={{ duration: 0.2 }}>
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 20px" }}>
                Persona profile
              </h2>

              {/* Avatar preview */}
              <div className="flex justify-center mb-6">
                <div className="relative">
                  <EchoAvatar
                    name={displayName || "??"}
                    size="xl"
                    gradient={selectedGradient}
                  />
                  {category && (
                    <div
                      className="absolute -bottom-1 -right-1 flex items-center justify-center rounded-full"
                      style={{
                        width: 28,
                        height: 28,
                        backgroundColor: categoryInfo[category].color,
                        border: "3px solid var(--echo-bg-primary)",
                        fontSize: 14,
                      }}
                    >
                      {categoryInfo[category].emoji}
                    </div>
                  )}
                </div>
              </div>

              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Display Name *
              </label>
              <input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="How others will see you"
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

              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Username *
              </label>
              <div className="relative mb-5">
                <span
                  className="absolute left-4 top-1/2 -translate-y-1/2"
                  style={{ fontSize: 16, color: "var(--echo-text-tertiary)" }}
                >
                  @
                </span>
                <input
                  value={username}
                  onChange={(e) => setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ""))}
                  placeholder="unique_handle"
                  className="w-full rounded-xl pr-4 outline-none"
                  style={{
                    height: 56,
                    paddingLeft: 36,
                    backgroundColor: "var(--echo-bg-surface)",
                    border: "2px solid var(--echo-border)",
                    fontSize: 16,
                    color: "var(--echo-text-primary)",
                  }}
                  onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                  onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
                />
              </div>

              <label style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)", display: "block", marginBottom: 6 }}>
                Bio
              </label>
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value.slice(0, 200))}
                placeholder="Tell people about this persona..."
                rows={3}
                className="w-full rounded-xl px-4 py-3 outline-none resize-none mb-1"
                style={{
                  backgroundColor: "var(--echo-bg-surface)",
                  border: "2px solid var(--echo-border)",
                  fontSize: 15,
                  color: "var(--echo-text-primary)",
                  fontFamily: "inherit",
                  lineHeight: "22px",
                }}
                onFocus={(e) => (e.target.style.borderColor = "var(--echo-border-focus)")}
                onBlur={(e) => (e.target.style.borderColor = "var(--echo-border)")}
              />
              <div className="text-right">
                <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>{bio.length}/200</span>
              </div>
            </motion.div>
          )}

          {/* STEP 2: Privacy */}
          {step === 2 && (
            <motion.div key="s2" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} transition={{ duration: 0.2 }}>
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 4px" }}>
                Privacy settings
              </h2>
              <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: "0 0 20px" }}>
                Control who can see and interact with this persona
              </p>

              {/* Visibility */}
              <div className="mb-6">
                <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                  Profile Visibility
                </span>
                <div className="flex flex-col gap-2 mt-2">
                  {visibilityOptions.map((opt) => {
                    const Icon = opt.icon;
                    const isSelected = visibility === opt.level;
                    return (
                      <button
                        key={opt.level}
                        onClick={() => setVisibility(opt.level)}
                        className="flex items-center gap-3 w-full rounded-xl px-4 py-3 text-left cursor-pointer transition-all"
                        style={{
                          backgroundColor: isSelected ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
                          border: `2px solid ${isSelected ? "var(--echo-primary)" : "transparent"}`,
                        }}
                      >
                        <Icon size={18} style={{ color: isSelected ? "var(--echo-primary)" : "var(--echo-text-tertiary)" }} />
                        <div className="flex-1">
                          <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>{opt.label}</div>
                          <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>{opt.desc}</div>
                        </div>
                        <div
                          className="flex items-center justify-center rounded-full"
                          style={{
                            width: 20,
                            height: 20,
                            border: `2px solid ${isSelected ? "var(--echo-primary)" : "var(--echo-border)"}`,
                            backgroundColor: isSelected ? "var(--echo-primary)" : "transparent",
                          }}
                        >
                          {isSelected && <div className="rounded-full" style={{ width: 6, height: 6, backgroundColor: "white" }} />}
                        </div>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Toggle settings */}
              <div className="flex flex-col gap-1">
                {[
                  { label: "Read Receipts", desc: "Let others know you've read their messages", value: readReceipts, setter: setReadReceipts },
                  { label: "Typing Indicator", desc: "Show when you're typing", value: typingIndicator, setter: setTypingIndicator },
                  { label: "Searchable", desc: "Appear in search results", value: searchable, setter: setSearchable },
                  { label: "Cross-Persona Linking", desc: "Allow contacts to discover your other personas", value: crossPersonaLinking, setter: setCrossPersonaLinking },
                ].map((toggle) => (
                  <div
                    key={toggle.label}
                    className="flex items-center justify-between py-3"
                  >
                    <div>
                      <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                        {toggle.label}
                      </div>
                      <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                        {toggle.desc}
                      </div>
                    </div>
                    <button
                      onClick={() => toggle.setter(!toggle.value)}
                      className="relative rounded-full border-none cursor-pointer shrink-0 transition-all"
                      style={{
                        width: 48,
                        height: 28,
                        backgroundColor: toggle.value ? "var(--echo-primary)" : "var(--echo-border)",
                      }}
                    >
                      <div
                        className="absolute top-1 rounded-full transition-all"
                        style={{
                          width: 20,
                          height: 20,
                          backgroundColor: "white",
                          left: toggle.value ? 26 : 2,
                          boxShadow: "var(--echo-shadow-sm)",
                        }}
                      />
                    </button>
                  </div>
                ))}
              </div>
            </motion.div>
          )}

          {/* STEP 3: Review */}
          {step === 3 && (
            <motion.div key="s3" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} transition={{ duration: 0.2 }}>
              <h2 style={{ fontSize: 20, fontWeight: 600, color: "var(--echo-text-primary)", margin: "0 0 20px" }}>
                Review your persona
              </h2>

              <div
                className="rounded-2xl overflow-hidden mb-4"
                style={{ border: "1px solid var(--echo-border)", boxShadow: "var(--echo-shadow-sm)" }}
              >
                {/* Preview card */}
                <div
                  className="p-5 flex flex-col items-center"
                  style={{
                    background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                  }}
                >
                  <div className="relative mb-3">
                    <EchoAvatar name={displayName || "??"} size="xl" gradient={selectedGradient} />
                    {category && (
                      <div
                        className="absolute -bottom-1 -right-1 flex items-center justify-center rounded-full"
                        style={{
                          width: 28,
                          height: 28,
                          backgroundColor: categoryInfo[category].color,
                          border: "3px solid white",
                          fontSize: 14,
                        }}
                      >
                        {categoryInfo[category].emoji}
                      </div>
                    )}
                  </div>
                  <div style={{ fontSize: 20, fontWeight: 700, color: "white" }}>{displayName}</div>
                  <div style={{ fontSize: 14, color: "rgba(255,255,255,0.7)", marginTop: 2 }}>@{username}</div>
                </div>

                <div className="p-4" style={{ backgroundColor: "var(--echo-bg-elevated)" }}>
                  {bio && (
                    <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", lineHeight: "20px", margin: "0 0 12px" }}>
                      {bio}
                    </p>
                  )}
                  <div className="flex flex-col gap-2">
                    {[
                      ["Category", category ? `${categoryInfo[category].emoji} ${categoryInfo[category].label}` : "—"],
                      ["Visibility", visibilityOptions.find((v) => v.level === visibility)?.label || ""],
                      ["Read Receipts", readReceipts ? "On" : "Off"],
                      ["Typing Indicator", typingIndicator ? "On" : "Off"],
                      ["Searchable", searchable ? "Yes" : "No"],
                      ["Cross-Persona Linking", crossPersonaLinking ? "Enabled" : "Disabled"],
                    ].map(([label, value]) => (
                      <div key={label} className="flex items-center justify-between">
                        <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>{label}</span>
                        <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div
                className="rounded-xl p-3 flex items-start gap-3"
                style={{ backgroundColor: "var(--echo-info-light)" }}
              >
                <span style={{ fontSize: 16, marginTop: 1 }}>🔑</span>
                <div>
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-info)" }}>
                    Cryptographic Separation
                  </div>
                  <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "16px", marginTop: 2 }}>
                    This persona will receive unique signing and encryption keys derived from your master identity. Messages sent from different personas are cryptographically unlinkable.
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* CTA */}
      <div className="px-6 pb-10 pt-4 shrink-0">
        <button
          onClick={() => (step < 3 ? setStep(step + 1) : handleCreate())}
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
          {step === 3 ? "Create Persona" : "Continue"}
          {step < 3 && <ChevronRight size={18} />}
        </button>
      </div>
    </div>
  );
}
