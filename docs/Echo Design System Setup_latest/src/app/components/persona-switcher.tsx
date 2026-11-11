import React, { useState } from "react";
import { useNavigate } from "react-router";
import { Check, ChevronDown, Plus, Settings, Shield } from "lucide-react";
import { EchoAvatar } from "./echo-avatar";
import { mockPersonas, categoryInfo, mockMasterIdentity, type Persona } from "./persona-data";
import { motion, AnimatePresence } from "motion/react";
import { VerificationBadge, type VerificationStatus } from "./verification-badge";

// Map master identity trust level to a verification badge status
function getUserBadgeStatus(master: typeof mockMasterIdentity): VerificationStatus {
  if (master.kycVerified) return "identity_full";
  if (master.emailVerified) return "identity_lite";
  if (master.phoneVerified) return "phone_verified";
  return "none";
}

const badgeLabels: Record<VerificationStatus, string> = {
  none: "Unverified",
  phone_verified: "Phone Verified",
  identity_lite: "Verified",
  identity_full: "Fully Verified",
  vip_pro: "VIP Pro",
};

const badgeColors: Record<VerificationStatus, string> = {
  none: "transparent",
  phone_verified: "#9CA3AF",
  identity_lite: "#6366F1",
  identity_full: "#10B981",
  vip_pro: "#F59E0B",
};

interface PersonaSwitcherProps {
  compact?: boolean;
}

export function PersonaSwitcher({ compact = false }: PersonaSwitcherProps) {
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const [personas, setPersonas] = useState(mockPersonas);
  const [showBadgeTooltip, setShowBadgeTooltip] = useState(false);

  const activePersona = personas.find((p) => p.isActive) || personas[0];
  const userBadge = getUserBadgeStatus(mockMasterIdentity);

  const handleSwitch = (personaId: string) => {
    setPersonas(
      personas.map((p) => ({
        ...p,
        isActive: p.personaId === personaId,
      }))
    );
    setIsOpen(false);
  };

  if (compact) {
    return (
      <div className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 bg-transparent border-none cursor-pointer px-2 py-1.5 rounded-xl hover:bg-[var(--echo-bg-surface)] transition-colors"
        >
          <EchoAvatar name={activePersona.displayName} size="xs" gradient={activePersona.gradient} />
          <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-primary)" }}>
            {activePersona.displayName}
          </span>
          {/* User's own verification badge */}
          {userBadge !== "none" && (
            <span
              className="relative"
              onClick={(e) => {
                e.stopPropagation();
                setShowBadgeTooltip(!showBadgeTooltip);
              }}
            >
              <VerificationBadge status={userBadge} size="md" />
            </span>
          )}
          <ChevronDown
            size={14}
            style={{
              color: "var(--echo-text-tertiary)",
              transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
              transition: "transform 0.2s ease",
            }}
          />
        </button>

        {/* Badge tooltip */}
        <AnimatePresence>
          {showBadgeTooltip && (
            <>
              <div className="fixed inset-0 z-30" onClick={() => setShowBadgeTooltip(false)} />
              <motion.div
                initial={{ opacity: 0, y: -4, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -4, scale: 0.95 }}
                transition={{ duration: 0.15 }}
                className="absolute top-full left-8 mt-1 z-40 rounded-xl"
                style={{
                  maxWidth: 220,
                  padding: 12,
                  backgroundColor: "var(--echo-bg-elevated)",
                  boxShadow: "var(--echo-shadow-lg)",
                  border: "1px solid var(--echo-border)",
                }}
              >
                <div style={{ fontSize: 13, color: "var(--echo-text-secondary)", marginBottom: 6 }}>
                  Your trust status:
                </div>
                <div className="flex items-center gap-1.5 mb-3">
                  <VerificationBadge status={userBadge} size="sm" />
                  <span
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      color: badgeColors[userBadge],
                    }}
                  >
                    {badgeLabels[userBadge]}
                  </span>
                </div>
                <button
                  onClick={() => {
                    setShowBadgeTooltip(false);
                    navigate("/trust");
                  }}
                  className="bg-transparent border-none cursor-pointer p-0"
                  style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-accent)" }}
                >
                  Learn more &rarr;
                </button>
                {/* Tooltip arrow */}
                <div
                  className="absolute"
                  style={{
                    top: -6,
                    left: 24,
                    width: 0,
                    height: 0,
                    borderLeft: "6px solid transparent",
                    borderRight: "6px solid transparent",
                    borderBottom: "6px solid var(--echo-bg-elevated)",
                  }}
                />
              </motion.div>
            </>
          )}
        </AnimatePresence>

        <AnimatePresence>
          {isOpen && (
            <>
              <div className="fixed inset-0 z-30" onClick={() => setIsOpen(false)} />
              <motion.div
                initial={{ opacity: 0, y: -8, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -8, scale: 0.95 }}
                transition={{ duration: 0.15 }}
                className="absolute top-full left-0 mt-1 z-40 rounded-2xl overflow-hidden"
                style={{
                  width: 240,
                  backgroundColor: "var(--echo-bg-elevated)",
                  boxShadow: "var(--echo-shadow-xl)",
                  border: "1px solid var(--echo-border)",
                }}
              >
                <div className="p-2">
                  {personas.map((persona) => {
                    const cat = categoryInfo[persona.category];
                    return (
                      <button
                        key={persona.personaId}
                        onClick={() => handleSwitch(persona.personaId)}
                        className="flex items-center gap-3 w-full px-3 py-2.5 rounded-xl bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                      >
                        <EchoAvatar name={persona.displayName} size="sm" gradient={persona.gradient} />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-1.5">
                            <span
                              className="truncate"
                              style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}
                            >
                              {persona.displayName}
                            </span>
                            <span style={{ fontSize: 12 }}>{cat.emoji}</span>
                          </div>
                          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                            @{persona.username}
                          </span>
                        </div>
                        {persona.isActive && (
                          <Check size={16} style={{ color: "var(--echo-primary)" }} strokeWidth={2.5} />
                        )}
                      </button>
                    );
                  })}
                </div>
                <div
                  style={{ height: 1, backgroundColor: "var(--echo-border-light)" }}
                />
                <div className="p-2">
                  <button
                    onClick={() => {
                      setIsOpen(false);
                      navigate("/personas/create");
                    }}
                    className="flex items-center gap-3 w-full px-3 py-2 rounded-xl bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                  >
                    <div
                      className="flex items-center justify-center rounded-full"
                      style={{ width: 32, height: 32, backgroundColor: "var(--echo-primary-subtle)" }}
                    >
                      <Plus size={16} style={{ color: "var(--echo-primary)" }} />
                    </div>
                    <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-primary)" }}>
                      New Persona
                    </span>
                    <span
                      className="ml-auto rounded-full px-1.5"
                      style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", backgroundColor: "var(--echo-bg-surface)" }}
                    >
                      {mockMasterIdentity.personaCount}/{mockMasterIdentity.maxPersonas}
                    </span>
                  </button>
                  <button
                    onClick={() => {
                      setIsOpen(false);
                      navigate("/personas");
                    }}
                    className="flex items-center gap-3 w-full px-3 py-2 rounded-xl bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                  >
                    <div
                      className="flex items-center justify-center rounded-full"
                      style={{ width: 32, height: 32, backgroundColor: "var(--echo-bg-surface)" }}
                    >
                      <Settings size={14} style={{ color: "var(--echo-text-secondary)" }} />
                    </div>
                    <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
                      Manage Personas
                    </span>
                  </button>
                </div>
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </div>
    );
  }

  // Full card version for profile page
  return (
    <div
      className="rounded-2xl p-4"
      style={{
        backgroundColor: "var(--echo-bg-elevated)",
        boxShadow: "var(--echo-shadow-sm)",
      }}
    >
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Shield size={16} style={{ color: "var(--echo-primary)" }} />
          <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase", letterSpacing: "0.5px" }}>
            Active Persona
          </span>
        </div>
        <button
          onClick={() => navigate("/personas")}
          className="bg-transparent border-none cursor-pointer"
          style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-primary)" }}
        >
          Manage
        </button>
      </div>

      <div className="flex items-center gap-3 mb-3">
        <EchoAvatar name={activePersona.displayName} size="md" gradient={activePersona.gradient} />
        <div className="flex-1">
          <div className="flex items-center gap-1.5">
            <span style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
              {activePersona.displayName}
            </span>
            <span style={{ fontSize: 14 }}>{categoryInfo[activePersona.category].emoji}</span>
          </div>
          <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
            @{activePersona.username} · {categoryInfo[activePersona.category].label}
          </span>
        </div>
      </div>

      {/* Quick switch pills */}
      <div className="flex gap-2 overflow-x-auto pb-1">
        {personas.map((p) => (
          <button
            key={p.personaId}
            onClick={() => handleSwitch(p.personaId)}
            className="flex items-center gap-1.5 rounded-full px-3 py-1.5 border-none cursor-pointer shrink-0 transition-all"
            style={{
              backgroundColor: p.isActive ? "var(--echo-primary-subtle)" : "var(--echo-bg-surface)",
              border: p.isActive ? "2px solid var(--echo-primary)" : "2px solid transparent",
            }}
          >
            <EchoAvatar name={p.displayName} size="xs" gradient={p.gradient} />
            <span
              style={{
                fontSize: 12,
                fontWeight: p.isActive ? 600 : 400,
                color: p.isActive ? "var(--echo-primary)" : "var(--echo-text-secondary)",
              }}
            >
              {p.displayName}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}