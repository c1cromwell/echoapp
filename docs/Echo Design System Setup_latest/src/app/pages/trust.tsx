import React, { useState } from "react";
import { TrustRing } from "../components/trust-ring";
import {
  Shield, CircleCheck, TriangleAlert, Lock, Eye, Fingerprint,
  ChevronRight, ChevronDown, Globe, ShieldCheck, Users, Heart,
  TrendingUp, Award, Zap, UserCheck, ArrowUpRight, ArrowDownRight,
  Clock, Star, MessageCircle, Phone, CircleDot, X,
} from "lucide-react";
import {
  credentialTypeInfo,
  mockTrustedIssuers,
  issuerTypeInfo,
  trustLevelColors,
  type CredentialType,
} from "../components/credential-data";
import {
  trustLevelConfig,
  endorsementCategoryInfo,
  circleLevelConfig,
  getTrustLevel,
  mockTrustComponents,
  mockTrustHistory,
  mockTrustContacts,
  mockTrustCircles,
  mockEndorsements,
  mockVerificationMethods,
  mockFeatureGates,
  mockDisputes,
  mockRateLimits,
  type TrustLevel as TrustLevelType,
} from "../components/trust-network-data";
import { motion, AnimatePresence } from "motion/react";
import {
  AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, ReferenceLine,
} from "recharts";

const verifiedCredentials: { type: CredentialType; issuer: string; date: string; points: number }[] = [
  { type: "passport", issuer: "US DMV", date: "2024-11-15", points: 90 },
  { type: "bank_account", issuer: "Wells Fargo", date: "2024-12-01", points: 75 },
  { type: "education", issuer: "Stanford University", date: "2025-01-10", points: 50 },
];

const tabs = ["Overview", "Network", "Credentials", "Registry"];

export function TrustPage() {
  const [activeTab, setActiveTab] = useState(0);
  const [expandedIssuer, setExpandedIssuer] = useState<string | null>(null);
  const [expandedCircle, setExpandedCircle] = useState<string | null>("circle-inner");
  const [showEndorsementDetail, setShowEndorsementDetail] = useState(false);

  const currentScore = 87;
  const currentLevel = getTrustLevel(currentScore);
  const levelConfig = trustLevelConfig[currentLevel];
  const previousScore = mockTrustHistory[mockTrustHistory.length - 2]?.score || 0;
  const scoreDelta = currentScore - previousScore;

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 pt-14 pb-2">
        <div className="flex items-center gap-2 mb-2">
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
            Trust Score
          </h1>
        </div>
        <p
          style={{
            fontSize: 14,
            color: "var(--echo-text-tertiary)",
            margin: "0 0 8px",
          }}
        >
          Your decentralized trust profile & network
        </p>

        {/* Tabs */}
        <div
          className="flex rounded-xl p-1"
          style={{ backgroundColor: "var(--echo-bg-surface)" }}
        >
          {tabs.map((tab, i) => (
            <button
              key={tab}
              onClick={() => setActiveTab(i)}
              className="flex-1 flex items-center justify-center rounded-lg border-none cursor-pointer py-2 transition-all"
              style={{
                backgroundColor: activeTab === i ? "var(--echo-bg-elevated)" : "transparent",
                boxShadow: activeTab === i ? "var(--echo-shadow-sm)" : "none",
                fontSize: 13,
                fontWeight: activeTab === i ? 600 : 400,
                color: activeTab === i ? "var(--echo-text-primary)" : "var(--echo-text-tertiary)",
              }}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-4">
        {/* ========== OVERVIEW TAB ========== */}
        {activeTab === 0 && (
          <>
            {/* Trust Ring + Level Badge */}
            <div
              className="flex flex-col items-center py-6 rounded-2xl mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-md)",
              }}
            >
              <TrustRing score={currentScore} label={levelConfig.label} />
              <div className="flex items-center gap-2 mt-3">
                <span
                  className="rounded-full px-3 py-1"
                  style={{
                    fontSize: 12,
                    fontWeight: 600,
                    color: levelConfig.color,
                    backgroundColor: levelConfig.bgColor,
                  }}
                >
                  {levelConfig.label}
                </span>
                <div className="flex items-center gap-1">
                  {scoreDelta >= 0 ? (
                    <ArrowUpRight size={14} style={{ color: "var(--echo-success)" }} />
                  ) : (
                    <ArrowDownRight size={14} style={{ color: "var(--echo-error)" }} />
                  )}
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: scoreDelta >= 0 ? "var(--echo-success)" : "var(--echo-error)",
                    }}
                  >
                    {scoreDelta >= 0 ? "+" : ""}{scoreDelta} this week
                  </span>
                </div>
              </div>
              {/* Percentile */}
              <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 6 }}>
                Top 15% of network
              </span>
              {/* No-decay reassurance */}
              <div className="flex items-center gap-1.5 mt-3" style={{ opacity: 0.8 }}>
                <Lock size={12} style={{ color: "var(--echo-success)" }} />
                <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                  Your trust score never decreases from inactivity
                </span>
              </div>
            </div>

            {/* Trust Score History (90 days) */}
            <div
              className="rounded-2xl p-4 mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <div className="flex items-center justify-between mb-3">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: 0,
                  }}
                >
                  Score History (90 days)
                </h3>
                <TrendingUp size={16} style={{ color: "var(--echo-success)" }} />
              </div>
              <div style={{ width: "100%", height: 140 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={mockTrustHistory} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="trustGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="var(--echo-primary)" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="var(--echo-primary)" stopOpacity={0.02} />
                      </linearGradient>
                    </defs>
                    <XAxis
                      dataKey="date"
                      tick={{ fontSize: 10, fill: "var(--echo-text-tertiary)" }}
                      axisLine={false}
                      tickLine={false}
                      interval={2}
                    />
                    <YAxis
                      domain={[30, 100]}
                      tick={{ fontSize: 10, fill: "var(--echo-text-tertiary)" }}
                      axisLine={false}
                      tickLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: "var(--echo-bg-elevated)",
                        border: "1px solid var(--echo-border)",
                        borderRadius: 12,
                        fontSize: 12,
                        boxShadow: "var(--echo-shadow-md)",
                      }}
                      formatter={(value: number) => [`${value}`, "Trust Score"]}
                      labelFormatter={(label: string) => label}
                    />
                    {/* Trust level boundary lines */}
                    <ReferenceLine key="ref-80" y={80} strokeDasharray="3 3" stroke="var(--echo-success)" strokeOpacity={0.4} />
                    <ReferenceLine key="ref-60" y={60} strokeDasharray="3 3" stroke="var(--echo-primary)" strokeOpacity={0.3} />
                    <ReferenceLine key="ref-40" y={40} strokeDasharray="3 3" stroke="var(--echo-warning)" strokeOpacity={0.3} />
                    <Area
                      type="monotone"
                      dataKey="score"
                      stroke="var(--echo-primary)"
                      strokeWidth={2.5}
                      fill="url(#trustGradient)"
                      dot={false}
                      activeDot={{ r: 5, fill: "var(--echo-primary)", strokeWidth: 2, stroke: "var(--echo-bg-elevated)" }}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
              {/* History events */}
              <div className="flex flex-wrap gap-1.5 mt-2">
                {mockTrustHistory.filter((p) => p.event).slice(-3).map((p, i) => (
                  <span
                    key={i}
                    className="rounded-full px-2 py-0.5"
                    style={{
                      fontSize: 10,
                      fontWeight: 500,
                      color: "var(--echo-primary)",
                      backgroundColor: "var(--echo-primary-subtle)",
                    }}
                  >
                    {p.event}
                  </span>
                ))}
              </div>
            </div>

            {/* 4-Component Breakdown */}
            <div
              className="rounded-2xl p-4 mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <h3
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  textTransform: "uppercase",
                  letterSpacing: "0.5px",
                  margin: "0 0 12px",
                }}
              >
                Score Components
              </h3>
              <div className="flex flex-col gap-3">
                {[
                  { key: "verification", label: "Verification", max: 30, value: mockTrustComponents.verification, icon: ShieldCheck, color: "#10B981", desc: "Email, phone, KYC, social" },
                  { key: "network", label: "Network", max: 25, value: mockTrustComponents.network, icon: Users, color: "#8B5CF6", desc: "Endorsements, connections" },
                  { key: "behavior", label: "Behavior", max: 25, value: mockTrustComponents.behavior, icon: Star, color: "#3B82F6", desc: "Account age, response rate" },
                  { key: "transactions", label: "Transactions", max: 20, value: mockTrustComponents.transactions, icon: Zap, color: "#F59E0B", desc: "P2P completions, disputes" },
                ].map((comp) => {
                  const pct = (comp.value / comp.max) * 100;
                  const Icon = comp.icon;
                  return (
                    <div key={comp.key}>
                      <div className="flex items-center justify-between mb-1">
                        <div className="flex items-center gap-2">
                          <Icon size={14} style={{ color: comp.color }} />
                          <span style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                            {comp.label}
                          </span>
                          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                            ({comp.max}% weight)
                          </span>
                        </div>
                        <span style={{ fontSize: 13, fontWeight: 700, color: comp.color }}>
                          {comp.value}/{comp.max}
                        </span>
                      </div>
                      <div
                        className="rounded-full overflow-hidden"
                        style={{ height: 6, backgroundColor: "var(--echo-bg-surface)" }}
                      >
                        <motion.div
                          initial={{ width: 0 }}
                          animate={{ width: `${pct}%` }}
                          transition={{ duration: 0.8, ease: "easeOut" }}
                          className="h-full rounded-full"
                          style={{ backgroundColor: comp.color }}
                        />
                      </div>
                      <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 2, display: "block" }}>
                        {comp.desc}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Trust Level Ladder */}
            <div
              className="rounded-2xl p-4 mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <h3
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  textTransform: "uppercase",
                  letterSpacing: "0.5px",
                  margin: "0 0 12px",
                }}
              >
                Trust Levels
              </h3>
              <div className="flex flex-col gap-2">
                {(Object.keys(trustLevelConfig) as TrustLevelType[]).reverse().map((level) => {
                  const cfg = trustLevelConfig[level];
                  const isActive = level === currentLevel;
                  const isPast = cfg.range[1] < currentScore;
                  return (
                    <div
                      key={level}
                      className="flex items-center gap-3 rounded-xl px-3 py-2.5 transition-all"
                      style={{
                        backgroundColor: isActive ? cfg.bgColor : "var(--echo-bg-surface)",
                        border: isActive ? `2px solid ${cfg.color}` : "2px solid transparent",
                        opacity: isPast || isActive ? 1 : 0.5,
                      }}
                    >
                      <div
                        className="flex items-center justify-center rounded-full shrink-0"
                        style={{
                          width: 32,
                          height: 32,
                          backgroundColor: isActive ? cfg.color : isPast ? cfg.color + "30" : "var(--echo-border)",
                        }}
                      >
                        {isActive ? (
                          <CircleDot size={16} color="white" />
                        ) : isPast ? (
                          <CircleCheck size={16} style={{ color: cfg.color }} />
                        ) : (
                          <Lock size={14} style={{ color: "var(--echo-text-tertiary)" }} />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span style={{ fontSize: 14, fontWeight: isActive ? 700 : 500, color: isActive ? cfg.color : "var(--echo-text-primary)" }}>
                            {cfg.label}
                          </span>
                          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                            {cfg.range[0]}-{cfg.range[1]}
                          </span>
                        </div>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {cfg.description}
                        </span>
                      </div>
                      {isActive && (
                        <span
                          className="rounded-full px-2 py-0.5 shrink-0"
                          style={{ fontSize: 10, fontWeight: 600, color: "white", backgroundColor: cfg.color }}
                        >
                          Current
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Feature Access Gates */}
            <div
              className="rounded-2xl p-4 mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
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
                Feature Access
              </h3>
              <div className="flex flex-col gap-1.5">
                {mockFeatureGates.map((gate) => {
                  const gateLevel = trustLevelConfig[gate.requiredLevel];
                  return (
                    <div key={gate.feature} className="flex items-center gap-2.5 py-1.5">
                      <CheckCircle
                        size={15}
                        style={{
                          color: gate.currentAccess ? "var(--echo-success)" : "var(--echo-text-tertiary)",
                          flexShrink: 0,
                        }}
                      />
                      <span
                        style={{
                          fontSize: 13,
                          fontWeight: 500,
                          color: gate.currentAccess ? "var(--echo-text-primary)" : "var(--echo-text-tertiary)",
                          flex: 1,
                        }}
                      >
                        {gate.feature}
                      </span>
                      <span
                        className="rounded-full px-1.5 py-0.5"
                        style={{
                          fontSize: 9,
                          fontWeight: 600,
                          color: gateLevel.color,
                          backgroundColor: gateLevel.bgColor,
                        }}
                      >
                        {gateLevel.label}+
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* No Decay Reassurance — v3.1.1 policy change */}
            <div
              className="rounded-xl p-3 flex items-start gap-3 mb-4"
              style={{ backgroundColor: "var(--echo-success-light, rgba(16,185,129,0.08))" }}
            >
              <Lock size={16} style={{ color: "var(--echo-success)", marginTop: 2, flexShrink: 0 }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-success)" }}>
                  No Score Decay
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "18px", marginTop: 2 }}>
                  Your trust score never decreases from inactivity. Only reports or blocks from other users can reduce your score.
                </div>
              </div>
            </div>
          </>
        )}

        {/* ========== NETWORK TAB ========== */}
        {activeTab === 1 && (
          <>
            {/* Trust Circles */}
            <div className="mb-4 mt-2">
              <div className="flex items-center justify-between mb-3">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: 0,
                  }}
                >
                  Trust Circles
                </h3>
              </div>

              {/* Concentric visualization */}
              <div
                className="flex items-center justify-center rounded-2xl py-6 mb-4"
                style={{ backgroundColor: "var(--echo-bg-elevated)", boxShadow: "var(--echo-shadow-sm)" }}
              >
                <div className="relative" style={{ width: 220, height: 220 }}>
                  {/* Public ring */}
                  <div
                    className="absolute rounded-full flex items-center justify-center"
                    style={{
                      width: 220, height: 220, top: 0, left: 0,
                      border: `2px dashed ${circleLevelConfig.public.color}40`,
                      backgroundColor: circleLevelConfig.public.bgColor + "40",
                    }}
                  >
                    <span
                      className="absolute"
                      style={{ top: 6, fontSize: 9, fontWeight: 600, color: circleLevelConfig.public.color }}
                    >
                      Public
                    </span>
                  </div>
                  {/* Known ring */}
                  <div
                    className="absolute rounded-full flex items-center justify-center"
                    style={{
                      width: 166, height: 166, top: 27, left: 27,
                      border: `2px solid ${circleLevelConfig.known.color}40`,
                      backgroundColor: circleLevelConfig.known.bgColor + "60",
                    }}
                  >
                    <span
                      className="absolute"
                      style={{ top: 6, fontSize: 9, fontWeight: 600, color: circleLevelConfig.known.color }}
                    >
                      Known (4)
                    </span>
                  </div>
                  {/* Trusted ring */}
                  <div
                    className="absolute rounded-full flex items-center justify-center"
                    style={{
                      width: 112, height: 112, top: 54, left: 54,
                      border: `2px solid ${circleLevelConfig.trusted.color}60`,
                      backgroundColor: circleLevelConfig.trusted.bgColor + "80",
                    }}
                  >
                    <span
                      className="absolute"
                      style={{ top: 6, fontSize: 9, fontWeight: 600, color: circleLevelConfig.trusted.color }}
                    >
                      Trusted (3)
                    </span>
                  </div>
                  {/* Inner ring */}
                  <div
                    className="absolute rounded-full flex items-center justify-center"
                    style={{
                      width: 58, height: 58, top: 81, left: 81,
                      background: `linear-gradient(135deg, ${circleLevelConfig.inner.color}20, ${circleLevelConfig.inner.color}40)`,
                      border: `2px solid ${circleLevelConfig.inner.color}`,
                    }}
                  >
                    <span style={{ fontSize: 9, fontWeight: 700, color: circleLevelConfig.inner.color }}>
                      Inner (3)
                    </span>
                  </div>
                </div>
              </div>

              {/* Circle list with contacts */}
              <div className="flex flex-col gap-3">
                {mockTrustCircles.map((circle) => {
                  const cfg = circleLevelConfig[circle.level];
                  const isExpanded = expandedCircle === circle.id;
                  return (
                    <div
                      key={circle.id}
                      className="rounded-2xl overflow-hidden"
                      style={{
                        backgroundColor: "var(--echo-bg-elevated)",
                        border: "1px solid var(--echo-border)",
                        boxShadow: "var(--echo-shadow-sm)",
                      }}
                    >
                      <button
                        onClick={() => setExpandedCircle(isExpanded ? null : circle.id)}
                        className="flex items-center gap-3 w-full px-4 py-3 bg-transparent border-none cursor-pointer text-left"
                      >
                        <div
                          className="flex items-center justify-center rounded-full shrink-0"
                          style={{ width: 36, height: 36, backgroundColor: cfg.bgColor }}
                        >
                          <Users size={16} style={{ color: cfg.color }} />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                              {cfg.label}
                            </span>
                            <span
                              className="rounded-full px-1.5 py-0.5"
                              style={{ fontSize: 10, fontWeight: 600, color: cfg.color, backgroundColor: cfg.bgColor }}
                            >
                              {circle.currentSize}/{circle.maxSize}
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-1 mt-0.5">
                            {cfg.permissions.slice(0, 2).map((p) => (
                              <span key={p} style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                                {p}{cfg.permissions.indexOf(p) < 1 ? " · " : ""}
                              </span>
                            ))}
                          </div>
                        </div>
                        <ChevronDown
                          size={16}
                          style={{
                            color: "var(--echo-text-tertiary)",
                            transform: isExpanded ? "rotate(180deg)" : "rotate(0deg)",
                            transition: "transform 0.2s",
                          }}
                        />
                      </button>

                      <AnimatePresence>
                        {isExpanded && (
                          <motion.div
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: "auto", opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            transition={{ duration: 0.2 }}
                            className="overflow-hidden"
                          >
                            <div
                              className="px-4 pb-3 pt-1"
                              style={{ borderTop: "1px solid var(--echo-border-light)" }}
                            >
                              {circle.contacts.map((contact) => {
                                const contactLevel = trustLevelConfig[contact.trustLevel];
                                return (
                                  <div
                                    key={contact.id}
                                    className="flex items-center gap-3 py-2.5"
                                    style={{ borderBottom: "1px solid var(--echo-border-light)" }}
                                  >
                                    <div
                                      className="rounded-full flex items-center justify-center shrink-0"
                                      style={{
                                        width: 32,
                                        height: 32,
                                        background: `linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))`,
                                      }}
                                    >
                                      <span style={{ fontSize: 13, color: "white", fontWeight: 600 }}>
                                        {contact.name.charAt(0)}
                                      </span>
                                    </div>
                                    <div className="flex-1 min-w-0">
                                      <div className="flex items-center gap-1.5">
                                        <span className="truncate" style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                                          {contact.name}
                                        </span>
                                        {contact.endorsedYou && (
                                          <Award size={12} style={{ color: "var(--echo-warning)", flexShrink: 0 }} />
                                        )}
                                      </div>
                                      <span
                                        style={{ fontSize: 11, color: contactLevel.color }}
                                      >
                                        {contactLevel.label} · {contact.trustScore}
                                      </span>
                                    </div>
                                    <ChevronRight size={14} style={{ color: "var(--echo-text-tertiary)" }} />
                                  </div>
                                );
                              })}
                              {/* Auto-promotion info */}
                              <div className="flex items-start gap-2 mt-2 pt-2">
                                <Zap size={13} style={{ color: "var(--echo-accent)", marginTop: 1, flexShrink: 0 }} />
                                <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)", lineHeight: "16px" }}>
                                  Contacts auto-promote based on messages exchanged, call count, and time known
                                </span>
                              </div>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Endorsements Received */}
            <div className="mb-4">
              <div className="flex items-center justify-between mb-3">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: 0,
                  }}
                >
                  Endorsements Received ({mockEndorsements.length})
                </h3>
                <button
                  onClick={() => setShowEndorsementDetail(true)}
                  className="bg-transparent border-none cursor-pointer p-0"
                  style={{ fontSize: 12, fontWeight: 500, color: "var(--echo-primary)" }}
                >
                  View All
                </button>
              </div>

              {/* Category badges */}
              <div className="flex flex-wrap gap-2 mb-3">
                {(Object.keys(endorsementCategoryInfo) as (keyof typeof endorsementCategoryInfo)[]).map((cat) => {
                  const info = endorsementCategoryInfo[cat];
                  const count = mockEndorsements.filter((e) => e.category === cat).length;
                  if (count === 0) return null;
                  return (
                    <div
                      key={cat}
                      className="flex items-center gap-1.5 rounded-full px-2.5 py-1"
                      style={{
                        backgroundColor: info.color + "15",
                        border: `1px solid ${info.color}30`,
                      }}
                    >
                      <span style={{ fontSize: 13 }}>{info.emoji}</span>
                      <span style={{ fontSize: 12, fontWeight: 600, color: info.color }}>
                        {info.label}
                      </span>
                      <span
                        className="rounded-full flex items-center justify-center"
                        style={{
                          width: 18,
                          height: 18,
                          fontSize: 10,
                          fontWeight: 700,
                          color: "white",
                          backgroundColor: info.color,
                        }}
                      >
                        {count}
                      </span>
                    </div>
                  );
                })}
              </div>

              {/* Recent endorsements */}
              <div
                className="rounded-2xl overflow-hidden"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                {mockEndorsements.slice(0, 3).map((endorsement, i) => {
                  const catInfo = endorsementCategoryInfo[endorsement.category];
                  return (
                    <div
                      key={endorsement.id}
                      className="flex items-center gap-3 px-4 py-3"
                      style={{
                        borderBottom: i < 2 ? "1px solid var(--echo-border-light)" : "none",
                      }}
                    >
                      <div
                        className="rounded-full flex items-center justify-center shrink-0"
                        style={{
                          width: 36,
                          height: 36,
                          background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                        }}
                      >
                        <span style={{ fontSize: 14, color: "white", fontWeight: 600 }}>
                          {endorsement.fromUserName.charAt(0)}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5">
                          <span className="truncate" style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            {endorsement.fromUserName}
                          </span>
                          {endorsement.weight >= 1.5 && (
                            <ShieldCheck size={12} style={{ color: "var(--echo-success)", flexShrink: 0 }} />
                          )}
                        </div>
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          Endorsed you as {catInfo.emoji} {catInfo.label}
                        </span>
                      </div>
                      <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                        {new Date(endorsement.timestamp).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Rate Limits */}
            <div
              className="rounded-2xl p-4 mb-4"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
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
                Daily Limits
              </h3>
              <div className="flex flex-col gap-2">
                {[
                  { label: "Endorsements Given", ...mockRateLimits.endorsementsGiven },
                  { label: "Circle Promotions", ...mockRateLimits.circlePromotions },
                  { label: "Reports Filed", ...mockRateLimits.reportsFiled },
                ].map((item) => (
                  <div key={item.label} className="flex items-center gap-3">
                    <span style={{ fontSize: 13, color: "var(--echo-text-secondary)", flex: 1 }}>
                      {item.label}
                    </span>
                    <div className="flex items-center gap-2">
                      <div
                        className="rounded-full overflow-hidden"
                        style={{ width: 60, height: 5, backgroundColor: "var(--echo-bg-surface)" }}
                      >
                        <div
                          className="h-full rounded-full"
                          style={{
                            width: `${(item.used / item.limit) * 100}%`,
                            backgroundColor: item.used >= item.limit ? "var(--echo-error)" : "var(--echo-primary)",
                          }}
                        />
                      </div>
                      <span style={{ fontSize: 12, fontWeight: 600, color: "var(--echo-text-tertiary)", width: 30, textAlign: "right" }}>
                        {item.used}/{item.limit}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Trust-Messaging Integration */}
            <div
              className="rounded-xl p-3 flex items-start gap-3 mb-4"
              style={{ backgroundColor: "var(--echo-primary-subtle)" }}
            >
              <MessageCircle size={16} style={{ color: "var(--echo-primary)", marginTop: 2, flexShrink: 0 }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-primary)" }}>
                  Trust-Based Messaging
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "18px", marginTop: 2 }}>
                  As a <span style={{ fontWeight: 600 }}>{levelConfig.label}</span> user: no message filters, can call anyone, create groups up to 200 members. Messages include your trust level badge.
                </div>
              </div>
            </div>

            {/* Disputes */}
            {mockDisputes.length > 0 && (
              <div className="mb-4">
                <h3
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: "var(--echo-text-tertiary)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                    margin: "0 0 8px 4px",
                  }}
                >
                  Dispute History
                </h3>
                {mockDisputes.map((dispute) => (
                  <div
                    key={dispute.id}
                    className="rounded-2xl p-4"
                    style={{
                      backgroundColor: "var(--echo-bg-elevated)",
                      border: "1px solid var(--echo-border)",
                      boxShadow: "var(--echo-shadow-sm)",
                    }}
                  >
                    <div className="flex items-center gap-2 mb-2">
                      <TriangleAlert size={14} style={{ color: "var(--echo-warning)" }} />
                      <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                        {dispute.type.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase())}
                      </span>
                      <span
                        className="rounded-full px-2 py-0.5 ml-auto"
                        style={{
                          fontSize: 10,
                          fontWeight: 600,
                          color: dispute.status === "resolved" ? "var(--echo-success)" : "var(--echo-warning)",
                          backgroundColor: dispute.status === "resolved" ? "var(--echo-success-light)" : "var(--echo-warning-light)",
                        }}
                      >
                        {dispute.status.charAt(0).toUpperCase() + dispute.status.slice(1)}
                      </span>
                    </div>
                    <p style={{ fontSize: 13, color: "var(--echo-text-secondary)", lineHeight: "18px", margin: "0 0 8px" }}>
                      {dispute.description}
                    </p>
                    <div className="flex items-center gap-4">
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        {dispute.jurorCount} jurors
                      </span>
                      <span style={{ fontSize: 12, color: "var(--echo-success)" }}>
                        {dispute.votesFor} for
                      </span>
                      <span style={{ fontSize: 12, color: "var(--echo-error)" }}>
                        {dispute.votesAgainst} against
                      </span>
                      <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                        Resolved {dispute.resolvedAt ? new Date(dispute.resolvedAt).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : ""}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Anti-Sybil Status */}
            <div
              className="rounded-xl p-3 flex items-start gap-3"
              style={{ backgroundColor: "var(--echo-success-light)" }}
            >
              <Shield size={16} style={{ color: "var(--echo-success)", marginTop: 2, flexShrink: 0 }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-success)" }}>
                  Sybil Protection Active
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "18px", marginTop: 2 }}>
                  Your account passes all anti-sybil checks: device fingerprint, behavioral analysis, and endorsement graph verification.
                </div>
              </div>
            </div>
          </>
        )}

        {/* ========== CREDENTIALS TAB ========== */}
        {activeTab === 2 && (
          <>
            {/* Verification Methods */}
            <div className="mb-4 mt-2">
              <h3
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  textTransform: "uppercase",
                  letterSpacing: "0.5px",
                  margin: "0 0 10px 4px",
                }}
              >
                Verification Methods
              </h3>

              {/* Verification badge status */}
              <div className="flex gap-2 mb-3">
                {[
                  { label: "Basic", emoji: "✓", req: "Email + Phone", achieved: true },
                  { label: "Standard", emoji: "✓✓", req: "+ 2 Social", achieved: true },
                  { label: "Full", emoji: "✓✓✓", req: "+ Gov ID", achieved: true },
                ].map((badge) => (
                  <div
                    key={badge.label}
                    className="flex-1 rounded-xl p-2.5 flex flex-col items-center gap-1"
                    style={{
                      backgroundColor: badge.achieved ? "var(--echo-success-light)" : "var(--echo-bg-surface)",
                      border: badge.achieved ? "1px solid var(--echo-success)" : "1px solid var(--echo-border)",
                    }}
                  >
                    <span style={{ fontSize: 16 }}>{badge.emoji}</span>
                    <span style={{
                      fontSize: 11,
                      fontWeight: 600,
                      color: badge.achieved ? "var(--echo-success)" : "var(--echo-text-tertiary)",
                    }}>
                      {badge.label}
                    </span>
                    <span style={{ fontSize: 9, color: "var(--echo-text-tertiary)", textAlign: "center" }}>
                      {badge.req}
                    </span>
                  </div>
                ))}
              </div>

              <div
                className="rounded-2xl overflow-hidden"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                {mockVerificationMethods.map((method, i) => (
                  <div
                    key={method.id}
                    className="flex items-center gap-3 px-4 py-3"
                    style={{
                      borderBottom: i < mockVerificationMethods.length - 1 ? "1px solid var(--echo-border-light)" : "none",
                    }}
                  >
                    <span style={{ fontSize: 20 }}>{method.icon}</span>
                    <div className="flex-1">
                      <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                        {method.method}
                      </div>
                      {method.completed && method.completedAt && (
                        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                          Verified {new Date(method.completedAt).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <span style={{
                        fontSize: 13,
                        fontWeight: 600,
                        color: method.completed ? "var(--echo-success)" : "var(--echo-text-tertiary)",
                      }}>
                        +{method.points}
                      </span>
                      {method.completed ? (
                        <CheckCircle size={16} style={{ color: "var(--echo-success)" }} />
                      ) : (
                        <ChevronRight size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Verified Credentials */}
            <div className="mb-4">
              <h3
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  textTransform: "uppercase",
                  letterSpacing: "0.5px",
                  margin: "0 0 10px 4px",
                }}
              >
                Verified Credentials ({verifiedCredentials.length})
              </h3>
              <div className="flex flex-col gap-3">
                {verifiedCredentials.map((vc) => {
                  const info = credentialTypeInfo[vc.type];
                  return (
                    <div
                      key={vc.type}
                      className="rounded-2xl p-4"
                      style={{
                        backgroundColor: "var(--echo-bg-elevated)",
                        border: "1px solid var(--echo-border)",
                        boxShadow: "var(--echo-shadow-sm)",
                      }}
                    >
                      <div className="flex items-start gap-3 mb-3">
                        <div
                          className="flex items-center justify-center rounded-xl shrink-0"
                          style={{
                            width: 48,
                            height: 48,
                            backgroundColor: "var(--echo-primary-subtle)",
                          }}
                        >
                          <span style={{ fontSize: 24 }}>{info.emoji}</span>
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <span style={{ fontSize: 16, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                              {info.label}
                            </span>
                            <ShieldCheck size={16} style={{ color: "var(--echo-success)" }} />
                          </div>
                          <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                            Issued by {vc.issuer}
                          </div>
                        </div>
                        <span style={{ fontSize: 18, fontWeight: 700, color: "var(--echo-primary)" }}>
                          +{vc.points}
                        </span>
                      </div>

                      {/* Verification stages */}
                      <div className="flex items-center gap-1.5 mb-3">
                        {["Structure", "Issuer", "Crypto", "Status", "Claims"].map((s) => (
                          <div
                            key={s}
                            className="flex items-center gap-1 rounded-full px-2 py-0.5"
                            style={{ backgroundColor: "var(--echo-success-light)" }}
                          >
                            <CheckCircle size={10} style={{ color: "var(--echo-success)" }} />
                            <span style={{ fontSize: 9, fontWeight: 600, color: "var(--echo-success)" }}>{s}</span>
                          </div>
                        ))}
                      </div>

                      <div className="flex items-center justify-between">
                        <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          Verified {new Date(vc.date).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
                        </span>
                        <span
                          className="rounded-full px-2 py-0.5"
                          style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-success)", backgroundColor: "var(--echo-success-light)" }}
                        >
                          Active
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Available to verify */}
            <div>
              <h3
                style={{
                  fontSize: 13,
                  fontWeight: 600,
                  color: "var(--echo-text-tertiary)",
                  textTransform: "uppercase",
                  letterSpacing: "0.5px",
                  margin: "0 0 10px 4px",
                }}
              >
                Available to Verify
              </h3>
              <div className="flex flex-col gap-2">
                {(["drivers_license", "employment", "phone", "email"] as CredentialType[]).map((type) => {
                  const info = credentialTypeInfo[type];
                  return (
                    <button
                      key={type}
                      className="flex items-center gap-3 w-full rounded-xl px-4 py-3 bg-transparent border-none cursor-pointer text-left hover:bg-[var(--echo-bg-surface)] transition-colors"
                      style={{ border: "1px dashed var(--echo-border)" }}
                    >
                      <span style={{ fontSize: 20 }}>{info.emoji}</span>
                      <div className="flex-1">
                        <div style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-secondary)" }}>
                          {info.label}
                        </div>
                        <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                          Up to +{info.pointsHigh} points
                        </div>
                      </div>
                      <ChevronRight size={16} style={{ color: "var(--echo-text-tertiary)" }} />
                    </button>
                  );
                })}
              </div>
            </div>

            {/* ZK Privacy Note */}
            <div
              className="rounded-xl p-3 flex items-start gap-3 mt-4"
              style={{ backgroundColor: "var(--echo-info-light)" }}
            >
              <Lock size={16} style={{ color: "var(--echo-info)", marginTop: 2, flexShrink: 0 }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-info)" }}>
                  Zero-Knowledge Privacy
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "18px", marginTop: 2 }}>
                  Verification proofs are stored as ZK commitments. The blockchain proves <em>that</em> verification occurred, not <em>what</em> was verified.
                </div>
              </div>
            </div>
          </>
        )}

        {/* ========== REGISTRY TAB ========== */}
        {activeTab === 3 && (
          <>
            <div className="mb-3 mt-2">
              <div
                className="rounded-xl p-3 flex items-start gap-3"
                style={{ backgroundColor: "var(--echo-info-light)" }}
              >
                <Globe size={16} style={{ color: "var(--echo-info)", marginTop: 2 }} />
                <div>
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-info)" }}>
                    Trust Registry
                  </div>
                  <div style={{ fontSize: 12, color: "var(--echo-text-secondary)", lineHeight: "16px", marginTop: 2 }}>
                    Echo maintains a curated registry of trusted credential issuers verified through OIDC4VP standards with Constellation Network anchoring.
                  </div>
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-3">
              {mockTrustedIssuers.map((issuer) => {
                const it = issuerTypeInfo[issuer.type];
                const tl = trustLevelColors[issuer.trustLevel];
                const isExpanded = expandedIssuer === issuer.id;

                return (
                  <div
                    key={issuer.id}
                    className="rounded-2xl overflow-hidden"
                    style={{
                      backgroundColor: "var(--echo-bg-elevated)",
                      border: "1px solid var(--echo-border)",
                      boxShadow: "var(--echo-shadow-sm)",
                    }}
                  >
                    <button
                      onClick={() => setExpandedIssuer(isExpanded ? null : issuer.id)}
                      className="flex items-center gap-3 w-full px-4 py-3.5 bg-transparent border-none cursor-pointer text-left"
                    >
                      <div
                        className="flex items-center justify-center rounded-xl shrink-0"
                        style={{ width: 44, height: 44, backgroundColor: it.color + "18" }}
                      >
                        <span style={{ fontSize: 22 }}>{it.emoji}</span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span
                            className="truncate"
                            style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}
                          >
                            {issuer.name}
                          </span>
                          {issuer.status === "active" && (
                            <div className="rounded-full" style={{ width: 6, height: 6, backgroundColor: "var(--echo-success)" }} />
                          )}
                        </div>
                        <div className="flex items-center gap-2 mt-0.5">
                          <span
                            className="rounded-full px-1.5 py-0.5"
                            style={{ fontSize: 10, fontWeight: 600, color: tl.color, backgroundColor: tl.bg }}
                          >
                            {tl.label}
                          </span>
                          <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                            {it.label} · {issuer.jurisdiction}
                          </span>
                        </div>
                      </div>
                      <ChevronDown
                        size={16}
                        style={{
                          color: "var(--echo-text-tertiary)",
                          transform: isExpanded ? "rotate(180deg)" : "rotate(0deg)",
                          transition: "transform 0.2s",
                        }}
                      />
                    </button>

                    <AnimatePresence>
                      {isExpanded && (
                        <motion.div
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: "auto", opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.2 }}
                          className="overflow-hidden"
                        >
                          <div
                            className="px-4 pb-4 pt-1"
                            style={{ borderTop: "1px solid var(--echo-border-light)" }}
                          >
                            {/* DID */}
                            <div className="mb-3">
                              <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                DID
                              </span>
                              <div style={{ fontSize: 12, fontFamily: "monospace", color: "var(--echo-text-secondary)", marginTop: 2 }}>
                                {issuer.did}
                              </div>
                            </div>

                            {/* Risk & Weight */}
                            <div className="flex gap-3 mb-3">
                              <div className="flex-1 rounded-lg p-2" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                                <div style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                  Risk Score
                                </div>
                                <div style={{ fontSize: 16, fontWeight: 700, color: issuer.riskScore < 15 ? "var(--echo-success)" : "var(--echo-warning)", marginTop: 2 }}>
                                  {issuer.riskScore}/100
                                </div>
                              </div>
                              <div className="flex-1 rounded-lg p-2" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                                <div style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                  Onboarding Weight
                                </div>
                                <div style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-primary)", marginTop: 2 }}>
                                  {issuer.onboardingWeight}%
                                </div>
                              </div>
                            </div>

                            {/* Supported Credentials */}
                            <div className="mb-3">
                              <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                Supported Credentials
                              </span>
                              <div className="flex flex-wrap gap-1.5 mt-2">
                                {issuer.supportedCredentials.map((ct) => (
                                  <span
                                    key={ct}
                                    className="flex items-center gap-1 rounded-full px-2 py-0.5"
                                    style={{ fontSize: 11, fontWeight: 500, color: "var(--echo-primary)", backgroundColor: "var(--echo-primary-subtle)" }}
                                  >
                                    {credentialTypeInfo[ct].emoji} {credentialTypeInfo[ct].label}
                                  </span>
                                ))}
                              </div>
                            </div>

                            {/* Qualifications */}
                            {issuer.qualifications.length > 0 && (
                              <div>
                                <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                  Qualifications
                                </span>
                                <div className="flex flex-col gap-1 mt-2">
                                  {issuer.qualifications.map((q) => (
                                    <div key={q.name} className="flex items-center justify-between">
                                      <span style={{ fontSize: 12, color: "var(--echo-text-secondary)" }}>{q.name}</span>
                                      <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                                        Expires {new Date(q.expiresAt).toLocaleDateString("en-US", { month: "short", year: "numeric" })}
                                      </span>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}

                            {/* Constellation anchor */}
                            {issuer.anchorHash && (
                              <div className="mt-3 rounded-lg p-2" style={{ backgroundColor: "var(--echo-bg-surface)" }}>
                                <div style={{ fontSize: 10, fontWeight: 600, color: "var(--echo-text-tertiary)", textTransform: "uppercase" }}>
                                  Constellation Anchor
                                </div>
                                <div style={{ fontSize: 11, fontFamily: "monospace", color: "var(--echo-accent)", marginTop: 2 }}>
                                  {issuer.anchorHash}
                                </div>
                              </div>
                            )}
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                );
              })}
            </div>
          </>
        )}
      </div>

      {/* Endorsement Detail Sheet */}
      <AnimatePresence>
        {showEndorsementDetail && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50"
              style={{ backgroundColor: "rgba(0,0,0,0.35)" }}
              onClick={() => setShowEndorsementDetail(false)}
            />
            <motion.div
              initial={{ y: "100%", opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: "100%", opacity: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 350 }}
              className="fixed bottom-0 left-0 right-0 z-50 rounded-t-3xl max-w-[430px] mx-auto"
              style={{ backgroundColor: "var(--echo-bg-elevated)", maxHeight: "75vh" }}
            >
              <div className="flex justify-center pt-3 pb-1">
                <div className="rounded-full" style={{ width: 36, height: 4, backgroundColor: "var(--echo-border)" }} />
              </div>
              <div className="px-5 pb-8 overflow-y-auto" style={{ maxHeight: "calc(75vh - 30px)" }}>
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <Award size={20} style={{ color: "var(--echo-warning)" }} />
                    <span style={{ fontSize: 17, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                      All Endorsements
                    </span>
                  </div>
                  <button
                    onClick={() => setShowEndorsementDetail(false)}
                    className="bg-transparent border-none cursor-pointer p-1"
                  >
                    <X size={20} style={{ color: "var(--echo-text-tertiary)" }} />
                  </button>
                </div>

                <div className="flex flex-col gap-1">
                  {mockEndorsements.map((endorsement) => {
                    const catInfo = endorsementCategoryInfo[endorsement.category];
                    return (
                      <div
                        key={endorsement.id}
                        className="flex items-center gap-3 py-3"
                        style={{ borderBottom: "1px solid var(--echo-border-light)" }}
                      >
                        <div
                          className="rounded-full flex items-center justify-center shrink-0"
                          style={{
                            width: 40,
                            height: 40,
                            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                          }}
                        >
                          <span style={{ fontSize: 16, color: "white", fontWeight: 600 }}>
                            {endorsement.fromUserName.charAt(0)}
                          </span>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-1.5">
                            <span style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                              {endorsement.fromUserName}
                            </span>
                            {endorsement.weight >= 1.5 && (
                              <ShieldCheck size={12} style={{ color: "var(--echo-success)" }} />
                            )}
                          </div>
                          <div className="flex items-center gap-1.5 mt-0.5">
                            <span style={{ fontSize: 13 }}>{catInfo.emoji}</span>
                            <span style={{ fontSize: 13, fontWeight: 500, color: catInfo.color }}>
                              {catInfo.label}
                            </span>
                            <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                              · Weight {endorsement.weight}x
                            </span>
                          </div>
                        </div>
                        <span style={{ fontSize: 11, color: "var(--echo-text-tertiary)" }}>
                          {new Date(endorsement.timestamp).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                        </span>
                      </div>
                    );
                  })}
                </div>

                <div
                  className="rounded-xl p-3 mt-4 flex items-start gap-2"
                  style={{ backgroundColor: "var(--echo-bg-surface)" }}
                >
                  <Zap size={14} style={{ color: "var(--echo-accent)", marginTop: 1, flexShrink: 0 }} />
                  <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)", lineHeight: "18px" }}>
                    Endorsements from verified users (score 80+) carry 1.5x weight. Each user can endorse once per category. Rate limit: 5 endorsements per day.
                  </span>
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}
