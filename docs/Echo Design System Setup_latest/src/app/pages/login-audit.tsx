import React from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, CheckCircle, XCircle, TriangleAlert, RefreshCw, Key, Shield } from "lucide-react";
import { motion } from "motion/react";

type EventType = "success" | "failed" | "new_device" | "device_removed" | "passkey_changed" | "recovery";

interface AuditEvent {
  id: string;
  type: EventType;
  title: string;
  device: string;
  method?: string;
  location: string;
  time: string;
  suspicious?: boolean;
}

const eventConfig: Record<EventType, { icon: React.ElementType; color: string; bgColor: string }> = {
  success: { icon: CheckCircle, color: "var(--echo-success)", bgColor: "var(--echo-success-light)" },
  failed: { icon: XCircle, color: "var(--echo-error)", bgColor: "var(--echo-error-light)" },
  new_device: { icon: TriangleAlert, color: "var(--echo-warning)", bgColor: "var(--echo-warning-light)" },
  device_removed: { icon: RefreshCw, color: "var(--echo-text-tertiary)", bgColor: "var(--echo-bg-surface)" },
  passkey_changed: { icon: Key, color: "var(--echo-primary)", bgColor: "var(--echo-primary-subtle)" },
  recovery: { icon: Shield, color: "var(--echo-accent)", bgColor: "var(--echo-primary-subtle)" },
};

interface AuditSection {
  label: string;
  events: AuditEvent[];
}

const mockAuditLog: AuditSection[] = [
  {
    label: "Today",
    events: [
      {
        id: "a1",
        type: "success",
        title: "Successful login",
        device: "iPhone 15 Pro",
        method: "Face ID",
        location: "New York, NY",
        time: "2:34 PM",
      },
    ],
  },
  {
    label: "Yesterday",
    events: [
      {
        id: "a2",
        type: "success",
        title: "Successful login",
        device: "iPhone 15 Pro",
        method: "Face ID",
        location: "New York, NY",
        time: "11:20 AM",
      },
      {
        id: "a3",
        type: "failed",
        title: "Failed login",
        device: "Unknown device",
        location: "Chicago, IL",
        time: "8:15 AM",
        suspicious: true,
      },
    ],
  },
  {
    label: "March 8",
    events: [
      {
        id: "a4",
        type: "new_device",
        title: "New device added",
        device: "iPad Air registered",
        location: "New York, NY",
        time: "3:45 PM",
      },
    ],
  },
  {
    label: "March 5",
    events: [
      {
        id: "a5",
        type: "passkey_changed",
        title: "Passkey updated",
        device: "iPhone 15 Pro",
        location: "New York, NY",
        time: "9:10 AM",
      },
      {
        id: "a6",
        type: "success",
        title: "Successful login",
        device: "iPhone 15 Pro",
        method: "Face ID",
        location: "New York, NY",
        time: "9:08 AM",
      },
    ],
  },
  {
    label: "February 28",
    events: [
      {
        id: "a7",
        type: "device_removed",
        title: "Device removed",
        device: "MacBook Air revoked",
        location: "New York, NY",
        time: "4:20 PM",
      },
    ],
  },
];

export function LoginAuditPage() {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-3 px-4 pt-14 pb-3">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
        <h1
          style={{
            fontSize: 20,
            fontWeight: 600,
            color: "var(--echo-text-primary)",
            margin: 0,
          }}
        >
          Login History
        </h1>
      </div>

      <div className="flex-1 overflow-y-auto px-4 pb-6">
        {mockAuditLog.map((section) => (
          <div key={section.label} className="mb-5">
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
              {section.label}
            </h3>

            <div
              className="rounded-2xl overflow-hidden"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              {section.events.map((event, i) => {
                const config = eventConfig[event.type];
                const Icon = config.icon;
                return (
                  <motion.div
                    key={event.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.05, duration: 0.2 }}
                    className="flex items-start gap-3 px-4 py-3.5"
                    style={{
                      borderBottom:
                        i < section.events.length - 1
                          ? "1px solid var(--echo-border-light)"
                          : "none",
                      backgroundColor: event.suspicious
                        ? "var(--echo-warning-light)"
                        : "transparent",
                      borderLeft: event.suspicious
                        ? "3px solid var(--echo-warning)"
                        : "3px solid transparent",
                    }}
                  >
                    <div
                      className="flex items-center justify-center rounded-full shrink-0 mt-0.5"
                      style={{
                        width: 32,
                        height: 32,
                        backgroundColor: config.bgColor,
                      }}
                    >
                      <Icon size={16} style={{ color: config.color }} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div
                        style={{
                          fontSize: 16,
                          fontWeight: 600,
                          color: "var(--echo-text-primary)",
                        }}
                      >
                        {event.title}
                      </div>
                      <div style={{ fontSize: 14, color: "var(--echo-text-secondary)", marginTop: 2 }}>
                        {event.device}
                        {event.method && ` \u00B7 ${event.method}`}
                      </div>
                      <div style={{ fontSize: 14, color: "var(--echo-text-tertiary)", marginTop: 1 }}>
                        {event.location} {" \u00B7 "} {event.time}
                      </div>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}