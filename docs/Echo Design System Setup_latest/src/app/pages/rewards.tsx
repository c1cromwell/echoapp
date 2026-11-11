import React from "react";
import { useNavigate } from "react-router";
import {
  ArrowUpRight,
  TrendingUp,
  RefreshCw,
  Shield,
  Flame,
  Calendar,
  BarChart3,
  Star,
  Trophy,
  Gem,
  Users,
  ScrollText,
  MessageCircle,
  BadgeCheck,
  ChevronRight,
} from "lucide-react";
import { motion } from "motion/react";

const quickActions = [
  { icon: Trophy, label: "Achievements", sublabel: "12 unlocked", color: "var(--echo-warning)" },
  { icon: Gem, label: "Staking", sublabel: "12% APY", color: "var(--echo-accent)" },
  { icon: Users, label: "Referrals", sublabel: "50 ECHO each", color: "var(--echo-primary)" },
  { icon: ScrollText, label: "History", sublabel: "View all", color: "var(--echo-text-secondary)" },
];

const recentEarnings = [
  {
    id: "e1",
    icon: MessageCircle,
    title: "Messages Sent",
    subtitle: "24 messages today",
    amount: "+12.0",
    time: "2m",
    bgColor: "var(--echo-primary-subtle)",
    iconColor: "var(--echo-primary)",
  },
  {
    id: "e2",
    icon: BadgeCheck,
    title: "Identity Verified",
    subtitle: "Government ID",
    amount: "+100",
    time: "1h",
    bgColor: "var(--echo-success-light)",
    iconColor: "var(--echo-success)",
  },
  {
    id: "e3",
    icon: Users,
    title: "Referral Bonus",
    subtitle: "Sarah joined",
    amount: "+50.0",
    time: "Yesterday",
    bgColor: "var(--echo-accent-subtle)",
    iconColor: "var(--echo-accent)",
  },
];

export function RewardsPage() {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 pt-14 pb-2">
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
          Rewards
        </h1>
        <button
          className="flex items-center justify-center rounded-full border-none cursor-pointer"
          style={{
            width: 36,
            height: 36,
            backgroundColor: "var(--echo-bg-surface)",
          }}
        >
          <ScrollText size={18} style={{ color: "var(--echo-text-tertiary)" }} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-6">
        {/* Balance Card */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
          className="rounded-3xl p-6 mb-4"
          style={{
            background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
            boxShadow: "var(--echo-shadow-primary)",
          }}
        >
          <p style={{ fontSize: 14, color: "rgba(255,255,255,0.9)", margin: "0 0 4px" }}>
            Available Balance
          </p>
          <div className="flex items-baseline gap-2 mb-1">
            <span style={{ fontSize: 42, fontWeight: 700, color: "white", lineHeight: 1.1 }}>
              1,247.50
            </span>
            <span style={{ fontSize: 20, fontWeight: 500, color: "rgba(255,255,255,0.9)" }}>
              ECHO
            </span>
          </div>
          <p style={{ fontSize: 16, color: "rgba(255,255,255,0.7)", margin: "0 0 20px" }}>
            ≈ $124.75 USD
          </p>

          <div className="flex gap-3">
            {[
              { icon: ArrowUpRight, label: "Transfer" },
              { icon: TrendingUp, label: "Stake" },
              { icon: RefreshCw, label: "Convert" },
            ].map((action) => (
              <button
                key={action.label}
                className="flex-1 flex items-center justify-center gap-1.5 rounded-xl border-none cursor-pointer py-2.5 active:scale-[0.97] transition-transform"
                style={{
                  backgroundColor: "rgba(255,255,255,0.2)",
                  backdropFilter: "blur(10px)",
                }}
              >
                <action.icon size={14} color="white" />
                <span style={{ fontSize: 13, fontWeight: 600, color: "white" }}>
                  {action.label}
                </span>
              </button>
            ))}
          </div>
        </motion.div>

        {/* Multiplier Badges */}
        <div className="flex gap-3 mb-4">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.08, duration: 0.3 }}
            className="flex-1 flex items-center gap-3 rounded-2xl px-4 py-3"
            style={{
              backgroundColor: "var(--echo-bg-surface)",
              border: "1px solid var(--echo-border-light)",
            }}
          >
            <Shield size={18} style={{ color: "var(--echo-accent)" }} />
            <div>
              <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>Trust</div>
              <div style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-accent)" }}>1.5x</div>
            </div>
          </motion.div>
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.12, duration: 0.3 }}
            className="flex-1 flex items-center gap-3 rounded-2xl px-4 py-3"
            style={{
              backgroundColor: "var(--echo-bg-surface)",
              border: "1px solid var(--echo-border-light)",
            }}
          >
            <Flame size={18} style={{ color: "var(--echo-warning)" }} />
            <div>
              <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)" }}>7-Day</div>
              <div style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-accent)" }}>1.2x</div>
            </div>
          </motion.div>
        </div>

        {/* Stat Cards */}
        <div className="flex gap-3 mb-5">
          {[
            { icon: Calendar, value: "42.5", label: "Today", color: "var(--echo-accent)" },
            { icon: BarChart3, value: "892", label: "Month", color: "var(--echo-primary)" },
            { icon: Star, value: "58/100", label: "Cap", color: "var(--echo-warning)" },
          ].map((stat, i) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.15 + i * 0.05, duration: 0.3 }}
              className="flex-1 flex flex-col items-center rounded-2xl py-3"
              style={{
                backgroundColor: "var(--echo-bg-surface)",
                border: "1px solid var(--echo-border-light)",
              }}
            >
              <stat.icon size={20} style={{ color: stat.color, marginBottom: 6 }} />
              <span style={{ fontSize: 18, fontWeight: 700, color: "var(--echo-text-primary)" }}>
                {stat.value}
              </span>
              <span style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 2 }}>
                {stat.label}
              </span>
            </motion.div>
          ))}
        </div>

        {/* Quick Actions */}
        <h3
          style={{
            fontSize: 13,
            fontWeight: 600,
            color: "var(--echo-text-tertiary)",
            letterSpacing: "0.5px",
            textTransform: "uppercase",
            margin: "0 0 10px 4px",
          }}
        >
          Quick Actions
        </h3>
        <div className="grid grid-cols-2 gap-3 mb-5">
          {quickActions.map((action, i) => (
            <motion.button
              key={action.label}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.25 + i * 0.05, duration: 0.3 }}
              whileTap={{ scale: 0.97 }}
              className="flex flex-col items-center gap-2 rounded-2xl py-5 border-none cursor-pointer"
              style={{
                backgroundColor: "var(--echo-bg-surface)",
                border: "1px solid var(--echo-border-light)",
              }}
            >
              <action.icon size={28} style={{ color: action.color }} />
              <div className="text-center">
                <div style={{ fontSize: 14, fontWeight: 600, color: "var(--echo-text-primary)" }}>
                  {action.label}
                </div>
                <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                  {action.sublabel}
                </div>
              </div>
            </motion.button>
          ))}
        </div>

        {/* Recent Earnings */}
        <div className="flex items-center justify-between mb-3">
          <h3
            style={{
              fontSize: 13,
              fontWeight: 600,
              color: "var(--echo-text-tertiary)",
              letterSpacing: "0.5px",
              textTransform: "uppercase",
              margin: 0,
              marginLeft: 4,
            }}
          >
            Recent Earnings
          </h3>
          <button
            className="flex items-center gap-0.5 bg-transparent border-none cursor-pointer p-0"
            style={{ fontSize: 13, fontWeight: 600, color: "var(--echo-accent)" }}
          >
            See All
            <ChevronRight size={14} style={{ color: "var(--echo-accent)" }} />
          </button>
        </div>

        <div
          className="rounded-2xl overflow-hidden"
          style={{
            backgroundColor: "var(--echo-bg-elevated)",
            boxShadow: "var(--echo-shadow-sm)",
          }}
        >
          {recentEarnings.map((earning, i) => {
            const Icon = earning.icon;
            return (
              <motion.div
                key={earning.id}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.35 + i * 0.05, duration: 0.2 }}
                className="flex items-center gap-3 px-4 py-3.5"
                style={{
                  borderBottom: i < recentEarnings.length - 1 ? "1px solid var(--echo-border-light)" : "none",
                }}
              >
                <div
                  className="flex items-center justify-center rounded-xl shrink-0"
                  style={{
                    width: 44,
                    height: 44,
                    backgroundColor: earning.bgColor,
                  }}
                >
                  <Icon size={22} style={{ color: earning.iconColor }} />
                </div>
                <div className="flex-1 min-w-0">
                  <div style={{ fontSize: 15, fontWeight: 500, color: "var(--echo-text-primary)" }}>
                    {earning.title}
                  </div>
                  <div style={{ fontSize: 13, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                    {earning.subtitle}
                  </div>
                </div>
                <div className="text-right shrink-0">
                  <div style={{ fontSize: 16, fontWeight: 700, color: "var(--echo-success)" }}>
                    {earning.amount}
                  </div>
                  <div style={{ fontSize: 12, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                    {earning.time}
                  </div>
                </div>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
}