import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { Shield, CircleCheck, Circle, ArrowRight, Lock, Users, Award, Landmark } from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import { TrustRing } from "../../components/trust-ring";
import { credentialTypeInfo } from "../../components/credential-data";
import { motion } from "motion/react";

const unlockBenefits = [
  { icon: Users, text: "Create groups up to 30 people" },
  { icon: Award, text: "Earn 2\u00d7 message rewards" },
  { icon: Shield, text: "Verified badge on your profile" },
  { icon: Landmark, text: "Connect bank for instant transfers" },
];

export function TrustDashboardPage() {
  const navigate = useNavigate();
  const { state } = useOnboarding();
  const [animatedScore, setAnimatedScore] = useState(0);

  // Calculate score from credential verification or fallback
  let score = 5; // Base for completing onboarding
  if (state.phoneNumber || true) score += 5;
  if (state.passkeyEnabled) score += 5;

  // If credentials were verified, use that score instead
  if (state.trustScoreBreakdown) {
    score = state.trustScoreBreakdown.totalScore;
  }

  const trustTier =
    score >= 80 ? "Verified" :
    score >= 60 ? "Trusted" :
    score >= 40 ? "Member" :
    score >= 20 ? "Newcomer" :
    "Unverified";

  useEffect(() => {
    const timer = setTimeout(() => setAnimatedScore(score), 300);
    return () => clearTimeout(timer);
  }, [score]);

  const securityItems = [
    {
      done: true,
      label: "Phone verified",
      action: null,
      points: "+5",
    },
    {
      done: state.passkeyEnabled,
      label: "Passkey enabled",
      action: state.passkeyEnabled ? null : "Enable",
      points: "+5",
    },
    {
      done: !!state.recoveryMethod,
      label: "Recovery method",
      action: state.recoveryMethod ? null : "Add",
      points: "+5",
    },
    {
      done: state.verifiedCredentials.length > 0,
      label: "Credentials verified",
      action: state.verifiedCredentials.length > 0 ? null : "Verify",
      points: `+${state.verifiedCredentials.length > 0 ? state.trustScoreBreakdown?.credentialScores.reduce((s, c) => s + c.points, 0) || 0 : "50"}`,
    },
  ];

  const handleStart = () => {
    localStorage.setItem("echo-onboarded", "true");
    navigate("/");
  };

  return (
    <div className="flex flex-col h-screen">
      <div className="flex-1 overflow-y-auto">
        {/* Header */}
        <div className="flex items-center gap-2 px-6 pt-14 mb-2">
          <Shield size={28} style={{ color: "var(--echo-primary)" }} />
          <h1
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "36px",
              letterSpacing: "-0.5px",
              margin: 0,
            }}
          >
            Your Trust Score
          </h1>
        </div>

        {/* Trust Ring */}
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
          className="flex flex-col items-center py-8"
        >
          <TrustRing score={animatedScore} size={140} strokeWidth={10} />
          <span
            className="mt-3"
            style={{
              fontSize: 16,
              fontWeight: 600,
              color: "var(--echo-text-secondary)",
            }}
          >
            {trustTier}
          </span>
        </motion.div>

        {/* Verified credential badges */}
        {state.verifiedCredentials.length > 0 && (
          <div className="px-6 mb-4">
            <div className="flex flex-wrap gap-2 justify-center">
              {state.verifiedCredentials.map((vc, i) => (
                <motion.div
                  key={vc.credentialId}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.5 + i * 0.1 }}
                  className="flex items-center gap-1.5 rounded-full px-3 py-1.5"
                  style={{
                    backgroundColor: "var(--echo-primary-subtle)",
                    border: "1px solid var(--echo-primary-lighter)",
                  }}
                >
                  <span style={{ fontSize: 13 }}>{credentialTypeInfo[vc.credentialType].emoji}</span>
                  <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-primary)" }}>
                    {credentialTypeInfo[vc.credentialType].label}
                  </span>
                  <CircleCheck size={12} style={{ color: "var(--echo-success)" }} />
                </motion.div>
              ))}
            </div>
          </div>
        )}

        {/* Score breakdown (if credentials verified) */}
        {state.trustScoreBreakdown && state.trustScoreBreakdown.credentialScores.length > 0 && (
          <div className="px-6 mb-4">
            <div
              className="rounded-2xl p-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between">
                  <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Base (onboarding)</span>
                  <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                    +{state.trustScoreBreakdown.baseScore}
                  </span>
                </div>
                {state.trustScoreBreakdown.credentialScores.map((cs) => (
                  <div key={cs.type} className="flex items-center justify-between">
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                      {credentialTypeInfo[cs.type].emoji} {credentialTypeInfo[cs.type].label}
                    </span>
                    <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-success)" }}>
                      +{cs.points}
                    </span>
                  </div>
                ))}
                {state.trustScoreBreakdown.multiCredentialBonus > 0 && (
                  <div className="flex items-center justify-between">
                    <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Multi-credential bonus</span>
                    <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-accent)" }}>
                      +{state.trustScoreBreakdown.multiCredentialBonus}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Security Setup Section */}
        <div className="px-6 mb-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="flex-1 h-px" style={{ backgroundColor: "var(--echo-border)" }} />
            <span
              style={{
                fontSize: 11,
                fontWeight: 600,
                color: "var(--echo-text-tertiary)",
                letterSpacing: "1px",
                textTransform: "uppercase" as const,
              }}
            >
              Your Security Setup
            </span>
            <div className="flex-1 h-px" style={{ backgroundColor: "var(--echo-border)" }} />
          </div>

          <div className="flex flex-col gap-1">
            {securityItems.map((item, i) => (
              <motion.div
                key={item.label}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3 + i * 0.1, duration: 0.3 }}
                className="flex items-center justify-between rounded-xl px-4"
                style={{ height: 44 }}
              >
                <div className="flex items-center gap-3">
                  {item.done ? (
                    <CircleCheck size={20} style={{ color: "var(--echo-success)" }} />
                  ) : (
                    <Circle size={20} style={{ color: "var(--echo-border)" }} />
                  )}
                  <span
                    style={{
                      fontSize: 14,
                      fontWeight: 400,
                      color: item.done ? "var(--echo-text-primary)" : "var(--echo-text-secondary)",
                    }}
                  >
                    {item.label}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  {item.done && (
                    <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-success)" }}>
                      {item.points}
                    </span>
                  )}
                  {item.action && (
                    <button
                      className="flex items-center gap-1 bg-transparent border-none cursor-pointer"
                      style={{
                        fontSize: 14,
                        fontWeight: 500,
                        color: "var(--echo-primary)",
                      }}
                    >
                      {item.action}
                      <ArrowRight size={14} style={{ color: "var(--echo-primary)" }} />
                    </button>
                  )}
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Unlock Benefits Section */}
        <div className="px-6 mb-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="flex-1 h-px" style={{ backgroundColor: "var(--echo-border)" }} />
            <span
              style={{
                fontSize: 11,
                fontWeight: 600,
                color: "var(--echo-text-tertiary)",
                letterSpacing: "1px",
                textTransform: "uppercase" as const,
              }}
            >
              Unlock with Trust
            </span>
            <div className="flex-1 h-px" style={{ backgroundColor: "var(--echo-border)" }} />
          </div>

          <div className="flex flex-col gap-3">
            {unlockBenefits.map((benefit, i) => {
              const Icon = benefit.icon;
              return (
                <motion.div
                  key={benefit.text}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.6 + i * 0.1, duration: 0.3 }}
                  className="flex items-center gap-3"
                >
                  <div
                    className="flex items-center justify-center rounded-lg shrink-0"
                    style={{
                      width: 32,
                      height: 32,
                      backgroundColor: "var(--echo-primary-subtle)",
                    }}
                  >
                    <Icon size={16} style={{ color: "var(--echo-primary)" }} />
                  </div>
                  <span
                    style={{
                      fontSize: 14,
                      color: "var(--echo-text-secondary)",
                      lineHeight: "20px",
                    }}
                  >
                    {benefit.text}
                  </span>
                </motion.div>
              );
            })}
          </div>
        </div>
      </div>

      {/* CTA */}
      <div className="px-6 pb-10 pt-4">
        <button
          onClick={handleStart}
          className="w-full border-none cursor-pointer rounded-xl flex items-center justify-center gap-2 active:scale-[0.98] transition-transform"
          style={{
            height: 56,
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            fontSize: 16,
            fontWeight: 600,
            color: "white",
            boxShadow: "var(--echo-shadow-primary)",
          }}
        >
          Start messaging
          <ArrowRight size={20} />
        </button>
      </div>
    </div>
  );
}
