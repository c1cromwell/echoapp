import React, { useState } from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, TrendingUp, Users, Clock, BarChart3, ShieldCheck, AlertCircle } from "lucide-react";
import {
  mockFunnelMetrics,
  credentialTypeInfo,
  type FunnelMetrics,
} from "../components/credential-data";
import { motion } from "motion/react";

function formatDuration(ms: number): string {
  const secs = Math.round(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const rem = secs % 60;
  return `${mins}m ${rem}s`;
}

function ProgressBar({ value, max, color }: { value: number; max: number; color: string }) {
  const pct = max > 0 ? (value / max) * 100 : 0;
  return (
    <div className="rounded-full overflow-hidden" style={{ height: 6, backgroundColor: "var(--echo-border)", flex: 1 }}>
      <div className="h-full rounded-full transition-all" style={{ width: `${pct}%`, backgroundColor: color }} />
    </div>
  );
}

export function AnalyticsPage() {
  const navigate = useNavigate();
  const data = mockFunnelMetrics;
  const [activeView, setActiveView] = useState<"funnel" | "credentials" | "issuers">("funnel");
  const maxStepCount = data.steps[0]?.count || 1;

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
          Onboarding Analytics
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto">
        {/* Summary Cards */}
        <div className="grid grid-cols-2 gap-3 px-4 pt-4 pb-2">
          {[
            { icon: Users, label: "Total Sessions", value: data.sessions.toLocaleString(), color: "var(--echo-primary)" },
            { icon: TrendingUp, label: "Conversion Rate", value: `${data.conversionRate}%`, color: "var(--echo-success)" },
            { icon: Clock, label: "Avg Duration", value: formatDuration(data.avgSessionDuration), color: "var(--echo-info)" },
            { icon: ShieldCheck, label: "Completed", value: data.steps[data.steps.length - 1].count.toLocaleString(), color: "var(--echo-accent)" },
          ].map((card) => {
            const Icon = card.icon;
            return (
              <div
                key={card.label}
                className="rounded-2xl p-3.5"
                style={{
                  backgroundColor: "var(--echo-bg-elevated)",
                  boxShadow: "var(--echo-shadow-sm)",
                }}
              >
                <Icon size={18} style={{ color: card.color, marginBottom: 6 }} />
                <div style={{ fontSize: 20, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                  {card.value}
                </div>
                <div style={{ fontSize: 11, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                  {card.label}
                </div>
              </div>
            );
          })}
        </div>

        {/* View toggle */}
        <div className="px-4 pt-2 pb-3">
          <div
            className="flex rounded-xl p-1"
            style={{ backgroundColor: "var(--echo-bg-surface)" }}
          >
            {([
              { key: "funnel" as const, label: "Funnel" },
              { key: "credentials" as const, label: "Credentials" },
              { key: "issuers" as const, label: "Issuers" },
            ]).map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveView(tab.key)}
                className="flex-1 rounded-lg border-none cursor-pointer py-2 transition-all"
                style={{
                  backgroundColor: activeView === tab.key ? "var(--echo-bg-elevated)" : "transparent",
                  boxShadow: activeView === tab.key ? "var(--echo-shadow-sm)" : "none",
                  fontSize: 13,
                  fontWeight: activeView === tab.key ? 600 : 400,
                  color: activeView === tab.key ? "var(--echo-text-primary)" : "var(--echo-text-tertiary)",
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* FUNNEL VIEW */}
        {activeView === "funnel" && (
          <div className="px-4 pb-6">
            <div className="flex flex-col gap-1">
              {data.steps.map((step, i) => {
                const prevCount = i > 0 ? data.steps[i - 1].count : step.count;
                const dropoff = prevCount > 0 ? Math.round(((prevCount - step.count) / prevCount) * 100) : 0;
                return (
                  <motion.div
                    key={step.name}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.04 }}
                  >
                    <div
                      className="rounded-xl p-3 mb-1"
                      style={{ backgroundColor: "var(--echo-bg-elevated)" }}
                    >
                      <div className="flex items-center justify-between mb-1.5">
                        <div className="flex items-center gap-2">
                          <span
                            className="flex items-center justify-center rounded-full"
                            style={{
                              width: 20,
                              height: 20,
                              fontSize: 10,
                              fontWeight: 700,
                              color: "white",
                              backgroundColor: "var(--echo-primary)",
                            }}
                          >
                            {i + 1}
                          </span>
                          <span style={{ fontSize: 13, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                            {step.name}
                          </span>
                        </div>
                        <span style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                          {step.count.toLocaleString()}
                        </span>
                      </div>

                      <div className="flex items-center gap-2 mb-1">
                        <ProgressBar value={step.count} max={maxStepCount} color="var(--echo-primary)" />
                        <span style={{ fontSize: 11, fontWeight: 600, color: "var(--echo-text-tertiary)", minWidth: 32, textAlign: "right" }}>
                          {Math.round((step.count / maxStepCount) * 100)}%
                        </span>
                      </div>

                      <div className="flex items-center gap-3">
                        {dropoff > 0 && (
                          <span style={{ fontSize: 10, color: "var(--echo-error)" }}>
                            -{dropoff}% dropoff
                          </span>
                        )}
                        {step.skipped > 0 && (
                          <span style={{ fontSize: 10, color: "var(--echo-warning)" }}>
                            {step.skipped.toLocaleString()} skipped
                          </span>
                        )}
                        {step.avgDuration > 0 && (
                          <span style={{ fontSize: 10, color: "var(--echo-text-tertiary)" }}>
                            avg {formatDuration(step.avgDuration)}
                          </span>
                        )}
                      </div>
                    </div>

                    {/* Connector */}
                    {i < data.steps.length - 1 && (
                      <div className="flex justify-center" style={{ height: 8 }}>
                        <div style={{ width: 2, height: "100%", backgroundColor: "var(--echo-border)" }} />
                      </div>
                    )}
                  </motion.div>
                );
              })}
            </div>
          </div>
        )}

        {/* CREDENTIALS VIEW */}
        {activeView === "credentials" && (
          <div className="px-4 pb-6">
            <div className="flex flex-col gap-3">
              {data.credentialStats.map((cs) => {
                const info = credentialTypeInfo[cs.type];
                const successRate = cs.attempts > 0 ? Math.round((cs.successes / cs.attempts) * 100) : 0;
                return (
                  <div
                    key={cs.type}
                    className="rounded-2xl p-4"
                    style={{
                      backgroundColor: "var(--echo-bg-elevated)",
                      boxShadow: "var(--echo-shadow-sm)",
                    }}
                  >
                    <div className="flex items-center gap-3 mb-3">
                      <span style={{ fontSize: 24 }}>{info.emoji}</span>
                      <div className="flex-1">
                        <div style={{ fontSize: 15, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                          {info.label}
                        </div>
                        <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>
                          {cs.attempts.toLocaleString()} attempts
                        </div>
                      </div>
                      <div className="text-right">
                        <div style={{ fontSize: 18, fontWeight: 700, color: successRate >= 95 ? "var(--echo-success)" : "var(--echo-warning)" }}>
                          {successRate}%
                        </div>
                        <div style={{ fontSize: 10, color: "var(--echo-text-tertiary)" }}>success</div>
                      </div>
                    </div>

                    <div className="flex gap-3">
                      <div className="flex-1 rounded-lg p-2 text-center" style={{ backgroundColor: "var(--echo-success-light)" }}>
                        <div style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-success)" }}>
                          {cs.successes.toLocaleString()}
                        </div>
                        <div style={{ fontSize: 10, color: "var(--echo-success)" }}>Verified</div>
                      </div>
                      <div className="flex-1 rounded-lg p-2 text-center" style={{ backgroundColor: "var(--echo-error-light)" }}>
                        <div style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-error)" }}>
                          {cs.failures.toLocaleString()}
                        </div>
                        <div style={{ fontSize: 10, color: "var(--echo-error)" }}>Failed</div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* ISSUERS VIEW */}
        {activeView === "issuers" && (
          <div className="px-4 pb-6">
            <div className="flex flex-col gap-2">
              {data.issuerDistribution
                .sort((a, b) => b.count - a.count)
                .map((issuer, i) => {
                  const maxCount = data.issuerDistribution[0]?.count || 1;
                  return (
                    <motion.div
                      key={issuer.issuer}
                      initial={{ opacity: 0, y: 8 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.05 }}
                      className="rounded-xl p-3"
                      style={{ backgroundColor: "var(--echo-bg-elevated)" }}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span style={{ fontSize: 14, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                          {issuer.issuer}
                        </span>
                        <span style={{ fontSize: 14, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                          {issuer.count.toLocaleString()}
                        </span>
                      </div>
                      <ProgressBar value={issuer.count} max={maxCount} color="var(--echo-primary)" />
                      <div className="flex items-center justify-between mt-1">
                        <span style={{ fontSize: 10, color: "var(--echo-text-tertiary)" }}>
                          #{i + 1} most used
                        </span>
                        <span style={{ fontSize: 10, color: "var(--echo-text-tertiary)" }}>
                          {Math.round((issuer.count / data.sessions) * 100)}% of sessions
                        </span>
                      </div>
                    </motion.div>
                  );
                })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
