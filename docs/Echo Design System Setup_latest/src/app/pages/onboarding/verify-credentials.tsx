import React, { useState, useRef } from "react";
import { useNavigate } from "react-router";
import {
  ArrowLeft, ChevronRight, Check, Loader2,
  ShieldCheck, Shield, X,
} from "lucide-react";
import { useOnboarding } from "../../components/onboarding-context";
import { motion, AnimatePresence } from "motion/react";
import {
  credentialTypeInfo,
  mockTrustedIssuers,
  verificationStageInfo,
  trustLevelColors,
  simulateVerification,
  calculateTrustScore,
  type CredentialType,
  type VerificationStage,
  type VerificationResult,
  type TrustedIssuer,
} from "../../components/credential-data";

type ViewState = "select" | "verifying" | "result";

// Available credentials to verify during onboarding
const availableCredentials: { type: CredentialType; issuer: TrustedIssuer }[] = [
  { type: "passport", issuer: mockTrustedIssuers[0] },
  { type: "drivers_license", issuer: mockTrustedIssuers[0] },
  { type: "bank_account", issuer: mockTrustedIssuers[1] },
  { type: "education", issuer: mockTrustedIssuers[2] },
];

const stageOrder: VerificationStage[] = ["structure", "issuer_trust", "cryptographic", "revocation", "claims"];

export function VerifyCredentialsPage() {
  const navigate = useNavigate();
  const { state, update } = useOnboarding();
  const [view, setView] = useState<ViewState>("select");
  const [selectedTypes, setSelectedTypes] = useState<Set<CredentialType>>(new Set());
  const [currentStageIdx, setCurrentStageIdx] = useState(-1);
  const [currentCredIdx, setCurrentCredIdx] = useState(0);
  const [completedResults, setCompletedResults] = useState<VerificationResult[]>([]);
  const [stageStatuses, setStageStatuses] = useState<("pending" | "running" | "passed" | "failed")[]>(
    stageOrder.map(() => "pending")
  );
  const verifyingRef = useRef(false);

  const selectedList = availableCredentials.filter((c) => selectedTypes.has(c.type));

  const toggleCredential = (type: CredentialType) => {
    const next = new Set(selectedTypes);
    if (next.has(type)) next.delete(type);
    else next.add(type);
    setSelectedTypes(next);
  };

  // Run the 5-stage verification pipeline with delays
  const runVerification = async () => {
    if (verifyingRef.current) return;
    verifyingRef.current = true;
    setView("verifying");
    const results: VerificationResult[] = [];

    for (let ci = 0; ci < selectedList.length; ci++) {
      setCurrentCredIdx(ci);
      setStageStatuses(stageOrder.map(() => "pending"));
      setCurrentStageIdx(-1);

      for (let si = 0; si < stageOrder.length; si++) {
        setCurrentStageIdx(si);
        setStageStatuses((prev) => prev.map((s, i) => (i === si ? "running" : i < si ? "passed" : s)));
        await new Promise((r) => setTimeout(r, 600 + Math.random() * 400));
        setStageStatuses((prev) => prev.map((s, i) => (i === si ? "passed" : s)));
      }

      const result = simulateVerification(selectedList[ci].type, selectedList[ci].issuer);
      results.push(result);
      await new Promise((r) => setTimeout(r, 300));
    }

    setCompletedResults(results);
    const breakdown = calculateTrustScore(results, state.passkeyEnabled ? 15 : 10);
    update({ verifiedCredentials: results, trustScoreBreakdown: breakdown });
    setView("result");
    verifyingRef.current = false;
  };

  const handleContinue = () => {
    navigate("/onboarding/trust-dashboard");
  };

  const handleSkip = () => {
    update({ credentialSkipped: true });
    navigate("/onboarding/trust-dashboard");
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <div className="flex items-center justify-between px-4 pt-14 pb-2">
        <button
          onClick={() => (view === "select" ? navigate(-1) : setView("select"))}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
        {view === "select" && (
          <button
            onClick={handleSkip}
            className="bg-transparent border-none cursor-pointer"
            style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-tertiary)" }}
          >
            Skip
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        <AnimatePresence mode="wait">
          {/* SELECT VIEW */}
          {view === "select" && (
            <motion.div
              key="select"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
              className="px-6"
            >
              {/* Icon */}
              <motion.div
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ duration: 0.4, type: "spring" }}
                className="flex items-center justify-center mx-auto mb-6"
              >
                <div
                  className="flex items-center justify-center rounded-2xl"
                  style={{
                    width: 80,
                    height: 80,
                    background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                  }}
                >
                  <ShieldCheck size={40} color="white" strokeWidth={1.5} />
                </div>
              </motion.div>

              <div className="text-center mb-6">
                <h1
                  style={{
                    fontSize: 28,
                    fontWeight: 700,
                    color: "var(--echo-text-primary)",
                    lineHeight: "36px",
                    letterSpacing: "-0.5px",
                    margin: "0 0 12px",
                  }}
                >
                  Verify your identity
                </h1>
                <p
                  style={{
                    fontSize: 16,
                    color: "var(--echo-text-secondary)",
                    lineHeight: "24px",
                    maxWidth: 320,
                    margin: "0 auto",
                  }}
                >
                  Present credentials from trusted issuers to boost your trust score. Your data stays private with zero-knowledge proofs.
                </p>
              </div>

              {/* OIDC4VP badge */}
              <div
                className="flex items-center gap-2 rounded-xl p-3 mx-auto mb-6"
                style={{
                  backgroundColor: "var(--echo-info-light)",
                  maxWidth: 300,
                }}
              >
                <Shield size={16} style={{ color: "var(--echo-info)" }} />
                <span style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-info)" }}>
                  OIDC4VP + W3C Verifiable Credentials standard
                </span>
              </div>

              {/* Credential options */}
              <div className="flex flex-col gap-3 mb-6">
                {availableCredentials.map(({ type, issuer }) => {
                  const info = credentialTypeInfo[type];
                  const isSelected = selectedTypes.has(type);
                  const tl = trustLevelColors[issuer.trustLevel];
                  const points =
                    issuer.trustLevel === "high" ? info.pointsHigh :
                    issuer.trustLevel === "medium" ? info.pointsMedium :
                    info.pointsBasic;

                  return (
                    <motion.button
                      key={type}
                      whileTap={{ scale: 0.98 }}
                      onClick={() => toggleCredential(type)}
                      className="relative flex items-center gap-3.5 w-full rounded-2xl px-4 cursor-pointer text-left transition-all"
                      style={{
                        minHeight: 76,
                        backgroundColor: isSelected ? "var(--echo-primary-subtle)" : "var(--echo-bg-elevated)",
                        border: `2px solid ${isSelected ? "var(--echo-primary)" : "var(--echo-border)"}`,
                        boxShadow: isSelected ? "var(--echo-shadow-md)" : "var(--echo-shadow-sm)",
                      }}
                    >
                      <span style={{ fontSize: 28, lineHeight: 1 }}>{info.emoji}</span>
                      <div className="flex-1 py-2.5">
                        <div className="flex items-center gap-2">
                          <span style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            {info.label}
                          </span>
                          <span
                            className="rounded-full px-2 py-0.5"
                            style={{ fontSize: 10, fontWeight: 600, color: tl.color, backgroundColor: tl.bg }}
                          >
                            {tl.label}
                          </span>
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                          {issuer.name}
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-1">
                        <span style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-primary)" }}>
                          +{points}
                        </span>
                        <span style={{ fontSize: 10, color: "var(--echo-text-tertiary)" }}>pts</span>
                      </div>
                      {/* Checkbox */}
                      <div
                        className="flex items-center justify-center rounded-lg shrink-0"
                        style={{
                          width: 24,
                          height: 24,
                          border: `2px solid ${isSelected ? "var(--echo-primary)" : "var(--echo-border)"}`,
                          backgroundColor: isSelected ? "var(--echo-primary)" : "transparent",
                        }}
                      >
                        {isSelected && <Check size={14} color="white" strokeWidth={3} />}
                      </div>
                    </motion.button>
                  );
                })}
              </div>

              {/* Start verification */}
              <div className="flex flex-col items-center gap-3 mb-10">
                {selectedTypes.size > 0 && (
                  <motion.button
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    onClick={runVerification}
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
                    Verify {selectedTypes.size} credential{selectedTypes.size > 1 ? "s" : ""}
                    <ChevronRight size={18} />
                  </motion.button>
                )}
              </div>
            </motion.div>
          )}

          {/* VERIFYING VIEW */}
          {view === "verifying" && (
            <motion.div
              key="verifying"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="px-6"
            >
              <div className="text-center mb-8">
                <h2
                  style={{
                    fontSize: 22,
                    fontWeight: 700,
                    color: "var(--echo-text-primary)",
                    margin: "0 0 8px",
                  }}
                >
                  Verifying credentials
                </h2>
                <p style={{ fontSize: 14, color: "var(--echo-text-secondary)", margin: 0 }}>
                  {currentCredIdx + 1} of {selectedList.length} ·{" "}
                  {credentialTypeInfo[selectedList[currentCredIdx]?.type]?.label || ""}
                </p>
              </div>

              {/* Current credential being verified */}
              <div
                className="rounded-2xl p-4 mb-6"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  border: "1px solid var(--echo-border)",
                  boxShadow: "var(--echo-shadow-md)",
                }}
              >
                <div className="flex items-center gap-3 mb-4">
                  <span style={{ fontSize: 32 }}>
                    {credentialTypeInfo[selectedList[currentCredIdx]?.type]?.emoji}
                  </span>
                  <div>
                    <div style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                      {credentialTypeInfo[selectedList[currentCredIdx]?.type]?.label}
                    </div>
                    <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
                      {selectedList[currentCredIdx]?.issuer.name}
                    </div>
                  </div>
                </div>

                {/* 5-stage pipeline */}
                <div className="flex flex-col gap-1">
                  {stageOrder.map((stage, i) => {
                    const info = verificationStageInfo[stage];
                    const status = stageStatuses[i];
                    return (
                      <motion.div
                        key={stage}
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.05 }}
                        className="flex items-center gap-3 rounded-xl px-3"
                        style={{
                          height: 48,
                          backgroundColor: status === "running" ? "var(--echo-primary-subtle)" : "transparent",
                          transition: "background-color 0.2s",
                        }}
                      >
                        <div className="flex items-center justify-center shrink-0" style={{ width: 24, height: 24 }}>
                          {status === "pending" && (
                            <div
                              className="rounded-full"
                              style={{ width: 8, height: 8, backgroundColor: "var(--echo-border)" }}
                            />
                          )}
                          {status === "running" && (
                            <Loader2 size={18} className="animate-spin" style={{ color: "var(--echo-primary)" }} />
                          )}
                          {status === "passed" && (
                            <motion.div
                              initial={{ scale: 0 }}
                              animate={{ scale: 1 }}
                              transition={{ type: "spring", stiffness: 300 }}
                            >
                              <Check size={18} style={{ color: "var(--echo-success)" }} strokeWidth={2.5} />
                            </motion.div>
                          )}
                          {status === "failed" && (
                            <X size={18} style={{ color: "var(--echo-error)" }} strokeWidth={2.5} />
                          )}
                        </div>
                        <div className="flex-1">
                          <div
                            style={{
                              fontSize: 13,
                              fontWeight: status === "running" ? 600 : 400,
                              color:
                                status === "running"
                                  ? "var(--echo-primary)"
                                  : status === "passed"
                                  ? "var(--echo-text-primary)"
                                  : "var(--echo-text-tertiary)",
                            }}
                          >
                            {info.label}
                          </div>
                          {status === "running" && (
                            <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                              {info.description}
                            </div>
                          )}
                        </div>
                        <span style={{ fontSize: 12 }}>{info.emoji}</span>
                      </motion.div>
                    );
                  })}
                </div>
              </div>

              {/* Progress indicator */}
              <div className="flex gap-2 justify-center">
                {selectedList.map((_, i) => (
                  <div
                    key={i}
                    className="rounded-full transition-all"
                    style={{
                      width: i === currentCredIdx ? 24 : 8,
                      height: 8,
                      backgroundColor: i <= currentCredIdx ? "var(--echo-primary)" : "var(--echo-border)",
                    }}
                  />
                ))}
              </div>
            </motion.div>
          )}

          {/* RESULT VIEW */}
          {view === "result" && (
            <motion.div
              key="result"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="px-6"
            >
              {/* Success icon */}
              <div className="flex flex-col items-center mb-6">
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ type: "spring", stiffness: 200, delay: 0.2 }}
                  className="flex items-center justify-center rounded-full mb-4"
                  style={{
                    width: 80,
                    height: 80,
                    background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                  }}
                >
                  <ShieldCheck size={40} color="white" strokeWidth={1.5} />
                </motion.div>
                <h2
                  style={{
                    fontSize: 24,
                    fontWeight: 700,
                    color: "var(--echo-text-primary)",
                    margin: "0 0 4px",
                  }}
                >
                  Verification complete!
                </h2>
                <p style={{ fontSize: 14, color: "var(--echo-text-secondary)" }}>
                  {completedResults.length} credential{completedResults.length > 1 ? "s" : ""} verified successfully
                </p>
              </div>

              {/* Earned badges */}
              <div className="mb-6">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: "0 0 10px",
                  }}
                >
                  Earned Badges
                </h3>
                <div className="flex flex-wrap gap-2">
                  {completedResults.map((r, i) => (
                    <motion.div
                      key={r.credentialId}
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{ opacity: 1, scale: 1 }}
                      transition={{ delay: 0.3 + i * 0.1 }}
                      className="flex items-center gap-2 rounded-full px-3 py-1.5"
                      style={{
                        backgroundColor: "var(--echo-primary-subtle)",
                        border: "1px solid var(--echo-primary-lighter)",
                      }}
                    >
                      <span style={{ fontSize: 14 }}>
                        {credentialTypeInfo[r.credentialType].emoji}
                      </span>
                      <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-primary)" }}>
                        {credentialTypeInfo[r.credentialType].label} Verified
                      </span>
                    </motion.div>
                  ))}
                </div>
              </div>

              {/* Trust score breakdown */}
              {state.trustScoreBreakdown && (
                <div
                  className="rounded-2xl p-4 mb-6"
                  style={{
                    backgroundColor: "var(--echo-bg-elevated)",
                    border: "1px solid var(--echo-border)",
                    boxShadow: "var(--echo-shadow-sm)",
                  }}
                >
                  <div className="flex items-center justify-between mb-3">
                    <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                      Trust Score Breakdown
                    </span>
                    <span style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-primary)" }}>
                      {state.trustScoreBreakdown.totalScore}
                    </span>
                  </div>

                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between">
                      <span style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>Base (onboarding)</span>
                      <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                        +{state.trustScoreBreakdown.baseScore}
                      </span>
                    </div>
                    {state.trustScoreBreakdown.credentialScores.map((cs) => (
                      <div key={cs.type} className="flex items-center justify-between">
                        <span style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>
                          {credentialTypeInfo[cs.type].emoji} {credentialTypeInfo[cs.type].label}
                        </span>
                        <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-success)" }}>
                          +{cs.points}
                        </span>
                      </div>
                    ))}
                    {state.trustScoreBreakdown.multiCredentialBonus > 0 && (
                      <div className="flex items-center justify-between">
                        <span style={{ fontSize: 13, color: "var(--echo-text-secondary)" }}>
                          Multi-credential bonus
                        </span>
                        <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-accent)" }}>
                          +{state.trustScoreBreakdown.multiCredentialBonus}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Progress bar */}
                  <div className="mt-3 rounded-full overflow-hidden" style={{ height: 6, backgroundColor: "var(--echo-border)" }}>
                    <motion.div
                      initial={{ width: 0 }}
                      animate={{ width: `${state.trustScoreBreakdown.totalScore}%` }}
                      transition={{ duration: 1, delay: 0.5 }}
                      className="h-full rounded-full"
                      style={{ background: "linear-gradient(90deg, var(--echo-gradient-from), var(--echo-gradient-to))" }}
                    />
                  </div>
                </div>
              )}

              {/* Verified issuers */}
              <div className="mb-8">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: "0 0 10px",
                  }}
                >
                  Verified Through
                </h3>
                <div className="flex flex-col gap-2">
                  {[...new Map(completedResults.map((r) => [r.issuerInfo.id, r.issuerInfo])).values()].map((issuer) => {
                    const tl = trustLevelColors[issuer.trustLevel];
                    return (
                      <div
                        key={issuer.id}
                        className="flex items-center gap-3 rounded-xl px-3 py-2"
                        style={{ backgroundColor: "var(--echo-bg-surface)" }}
                      >
                        <Shield size={16} style={{ color: tl.color }} />
                        <div className="flex-1">
                          <div style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            {issuer.name}
                          </div>
                          <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                            DID: {issuer.did.slice(0, 25)}...
                          </div>
                        </div>
                        <span
                          className="rounded-full px-2 py-0.5"
                          style={{ fontSize: 10, fontWeight: 600, color: tl.color, backgroundColor: tl.bg }}
                        >
                          {tl.label}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* CTA for result view */}
      {view === "result" && (
        <div className="px-6 pb-10 pt-4 shrink-0">
          <button
            onClick={handleContinue}
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
            Continue to Trust Score
            <ChevronRight size={18} />
          </button>
        </div>
      )}
    </div>
  );
}