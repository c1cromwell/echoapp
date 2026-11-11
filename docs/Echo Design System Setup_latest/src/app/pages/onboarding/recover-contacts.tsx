import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { ArrowLeft, CircleCheck, Clock, CircleX } from "lucide-react";
import { motion } from "motion/react";

interface TrustedContact {
  id: string;
  name: string;
  status: "confirmed" | "pending" | "declined";
}

const initialContacts: TrustedContact[] = [
  { id: "tc1", name: "Sarah M.", status: "confirmed" },
  { id: "tc2", name: "James K.", status: "pending" },
  { id: "tc3", name: "Maria L.", status: "pending" },
];

export function RecoverContactsPage() {
  const navigate = useNavigate();
  const [contacts, setContacts] = useState(initialContacts);
  const [expiry, setExpiry] = useState(85500); // ~23h 45m in seconds

  const confirmed = contacts.filter((c) => c.status === "confirmed").length;
  const required = 2;
  const progress = confirmed / required;

  // Simulate James confirming after 5 seconds
  useEffect(() => {
    const timer = setTimeout(() => {
      setContacts((prev) =>
        prev.map((c) => (c.id === "tc2" ? { ...c, status: "confirmed" } : c))
      );
    }, 5000);
    return () => clearTimeout(timer);
  }, []);

  // Countdown timer
  useEffect(() => {
    const timer = setInterval(() => {
      setExpiry((prev) => Math.max(0, prev - 1));
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // Navigate when threshold met
  useEffect(() => {
    if (confirmed >= required) {
      const timer = setTimeout(() => {
        navigate("/onboarding/recover-complete");
      }, 1500);
      return () => clearTimeout(timer);
    }
  }, [confirmed, required, navigate]);

  const hours = Math.floor(expiry / 3600);
  const minutes = Math.floor((expiry % 3600) / 60);

  const statusIcon = (status: string) => {
    if (status === "confirmed") return <CircleCheck size={20} style={{ color: "var(--echo-success)" }} />;
    if (status === "declined") return <CircleX size={20} style={{ color: "var(--echo-error)" }} />;
    return <Clock size={20} style={{ color: "var(--echo-text-tertiary)" }} />;
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <div className="flex items-center px-4 pt-14 pb-2">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center justify-center bg-transparent border-none cursor-pointer"
          style={{ width: 44, height: 44 }}
          aria-label="Go back"
        >
          <ArrowLeft size={24} style={{ color: "var(--echo-text-primary)" }} />
        </button>
      </div>

      <div className="flex-1 px-6">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <h1
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: "var(--echo-text-primary)",
              lineHeight: "36px",
              letterSpacing: "-0.5px",
              margin: "0 0 8px",
            }}
          >
            Waiting for friends
          </h1>
          <p
            style={{
              fontSize: 16,
              color: "var(--echo-text-secondary)",
              lineHeight: "24px",
              margin: "0 0 32px",
            }}
          >
            We've sent a request to your trusted contacts.{" "}
            <span style={{ fontWeight: 600 }}>{required} of {contacts.length}</span> need to confirm.
          </p>
        </motion.div>

        {/* Contact cards */}
        <div className="flex flex-col gap-3 mb-6">
          {contacts.map((contact, i) => (
            <motion.div
              key={contact.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 + i * 0.08, duration: 0.3 }}
              className="flex items-center gap-3 rounded-2xl px-4 py-3.5"
              style={{
                backgroundColor: "var(--echo-bg-elevated)",
                border: "1px solid var(--echo-border)",
                boxShadow: "var(--echo-shadow-sm)",
              }}
            >
              <div
                className="flex items-center justify-center rounded-full shrink-0"
                style={{
                  width: 40,
                  height: 40,
                  background: "linear-gradient(135deg, var(--echo-gradient-from), var(--echo-gradient-to))",
                }}
              >
                <span style={{ fontSize: 16, color: "white", fontWeight: 600 }}>
                  {contact.name.charAt(0)}
                </span>
              </div>
              <span
                className="flex-1"
                style={{ fontSize: 16, fontWeight: 500, color: "var(--echo-text-primary)" }}
              >
                {contact.name}
              </span>
              <motion.div
                key={contact.status}
                initial={{ scale: 0.8 }}
                animate={{ scale: 1 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                {statusIcon(contact.status)}
              </motion.div>
            </motion.div>
          ))}
        </div>

        {/* Progress bar */}
        <div className="mb-4">
          <div className="flex items-center justify-between mb-2">
            <span style={{ fontSize: 13, color: "var(--echo-text-tertiary)" }}>
              Progress: {confirmed} of {required} needed
            </span>
          </div>
          <div
            className="rounded-full overflow-hidden"
            style={{ height: 4, backgroundColor: "var(--echo-bg-surface)" }}
          >
            <motion.div
              className="h-full rounded-full"
              initial={{ width: 0 }}
              animate={{ width: `${Math.min(100, progress * 100)}%` }}
              transition={{ duration: 0.5, ease: "easeOut" }}
              style={{ backgroundColor: "var(--echo-primary)" }}
            />
          </div>
        </div>

        {/* Expiry */}
        <p
          style={{
            fontSize: 14,
            color: "var(--echo-text-tertiary)",
            textAlign: "center",
            lineHeight: "20px",
          }}
        >
          This request expires in{" "}
          <span style={{ fontWeight: 600 }}>
            {hours} hours, {minutes} minutes
          </span>
          .
        </p>
      </div>
    </div>
  );
}
